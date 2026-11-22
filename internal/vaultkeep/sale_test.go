package vaultkeep

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/vault"
)

// The instant sale over d2dctl.PayFake (the fork's creditgold) and VaultFake.Prices (the fork's item price), a Paladin and a
// Sorceress hero. The prices are made up per record; the point is that the credit is exactly their sum.

type saleRig struct {
	*rig
	pay     *d2dctl.PayFake
	instant bool
	confirm int
}

func newSaleRig(t *testing.T, class string, stash int) *saleRig {
	r := &saleRig{rig: newRig(t, class), pay: &d2dctl.PayFake{Stash: stash, Purse: 300}, instant: true}
	r.api.Prices = map[string]d2dctl.ItemPrice{}
	r.wireSale()
	return r
}

func (r *saleRig) wireSale() {
	r.k.Credit, r.k.Pricer = r.pay, r.api.VaultFake
	r.k.StashGold = func(context.Context) (int, error) { return r.pay.Stash, nil }
	r.k.SellInstant = func() bool { return r.instant }
	r.k.SellConfirm = func() int { return r.confirm }
	r.k.SaleFile = func() string { return filepath.Join(r.dir, "sale.json") }
	r.k.Measured = func() bool { return r.measured }
}

// item records an item the engine prices at sell (as recorded) and sellID (identified).
func (r *saleRig) item(i int, q, sell, sellID int) string {
	id := r.rig.add(i, "wsp", q)
	rec := r.state(id)
	r.api.Prices[rec.Raw] = d2dctl.ItemPrice{Sell: sell, SellID: sellID, Npc: "Akara", Was: rec.Identified}
	return id
}

func (r *saleRig) sellNow(confirm bool, ids ...string) ([]Result, *SaleView) {
	return r.k.SellNow(context.Background(), ids, confirm)
}

func okCount(res []Result) int {
	n := 0
	for _, x := range res {
		if x.OK {
			n++
		}
	}
	return n
}

func TestSellNowCreditsExactlyTheEnginePriceAndTheRecordsLeave(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newSaleRig(t, class, 1000)
			a, b, c := r.item(1, 2, 7, 100), r.item(2, 3, 9, 200), r.item(3, 3, 11, 300) // identified: the identified price counts
			res, v := r.sellNow(false, a, b, c)
			if okCount(res) != 3 || v == nil || v.State != SaleDone || v.Gold != 600 || v.Done != 3 || v.StashAfter != 1600 {
				t.Fatalf("res %+v view %+v", res, v)
			}
			if r.pay.Stash != 1600 || r.pay.Purse != 300 {
				t.Fatalf("stash %d purse %d", r.pay.Stash, r.pay.Purse)
			}
			if len(r.pay.Calls) != 1 || r.pay.Calls[0] != "credit 600 "+v.ID {
				t.Fatalf("one credit of the whole order: %v", r.pay.Calls)
			}
			for i, id := range []string{a, b, c} {
				rec := r.state(id)
				if rec.State != vault.Sold || rec.Gold != []int{100, 200, 300}[i] {
					t.Fatalf("record %s: %+v", id, rec)
				}
			}
			if !r.heroEmpty() || r.k.Info().Sold != 3 || r.k.Info().GoldEarned != 600 {
				t.Fatal("nothing is created in the hero; the counters follow the sale")
			}
			if !strings.Contains(v.Line, "Sold 3 items for 600 gold. Stash gold: 1,600") {
				t.Fatalf("line %q", v.Line)
			}
			if arch, _ := r.st.Archived(); len(arch) != 3 {
				t.Fatalf("every record is archived before it leaves: %d", len(arch))
			}
		})
	}
}

func TestSellNowPricesUnidentifiedAsRecordedAndRevealedAsIdentified(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newSaleRig(t, class, 0)
			un := r.item(1, 4, 20, 900) // magic, unidentified: the engine pays the base
			rv := r.item(2, 4, 20, 900)
			if err := r.st.Seal(rv, "stat:1:2"); err != nil {
				t.Fatal(err)
			}
			if err := r.st.Unseal(rv); err != nil {
				t.Fatal(err)
			}
			q := r.k.SaleQuote(context.Background(), []string{un, rv})
			if q.Total != 920 || len(q.Lines) != 2 || q.Lines[0].Gold != 20 || q.Lines[1].Gold != 900 {
				t.Fatalf("quote %+v", q)
			}
			if r.pay.Stash != 0 || len(r.pay.Calls) != 0 {
				t.Fatal("a quote changes nothing")
			}
		})
	}
}

func TestSellNowIsIdempotentAndSurvivesALostReply(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newSaleRig(t, class, 1000)
			a, b := r.item(1, 3, 1, 50), r.item(2, 3, 1, 70)
			r.pay.Fail = errors.New("socket closed") // unanswered: the credit may or may not have happened
			res, v := r.sellNow(false, a, b)
			if okCount(res) != 0 || v == nil || v.State != SaleQueued {
				t.Fatalf("res %+v view %+v", res, v)
			}
			if r.state(a).State != vault.Recorded || r.state(b).State != vault.Recorded {
				t.Fatal("the records stay until the stash confirms")
			}
			r.pay.Fail = nil
			r.open() // a bot restart: the order file survives, the Keeper does not
			r.wireSale()
			lines := r.run()
			if len(lines) == 0 || r.pay.Stash != 1120 || r.state(a).State != vault.Sold || r.state(b).State != vault.Sold {
				t.Fatalf("settled by the next step: lines %v stash %d", lines, r.pay.Stash)
			}
			last := r.pay.Calls[len(r.pay.Calls)-1]
			first := r.pay.Calls[0]
			if first != last {
				t.Fatalf("the same order id both times: %v", r.pay.Calls)
			}
			res2, _ := r.sellNow(false, a, b) // pressing again changes nothing
			if okCount(res2) != 0 || r.pay.Stash != 1120 {
				t.Fatalf("second press: %+v stash %d", res2, r.pay.Stash)
			}
			for _, x := range res2 {
				if x.Code != "state" {
					t.Fatalf("a sold record answers state: %+v", x)
				}
			}
		})
	}
}

func TestSellNowCreditedButRestartedBeforeTheRecordsLeft(t *testing.T) {
	r := newSaleRig(t, "paladin", 1000)
	a := r.item(1, 3, 1, 80)
	o := SaleOrder{ID: "s-7", Lines: []SaleLine{{ID: a, Gold: 80}}, Total: 80, State: SaleCredited, StashAfter: 1080}
	r.k.saveSale(o)
	r.pay.Stash = 1080 // the gold is already in the stash
	r.open()
	r.wireSale()
	r.run()
	if r.pay.Stash != 1080 || len(r.pay.Calls) != 0 || r.state(a).State != vault.Sold || r.state(a).Reason != "s-7" {
		t.Fatalf("no second credit; the record leaves: stash %d calls %v %+v", r.pay.Stash, r.pay.Calls, r.state(a))
	}
}

func TestSellNowPartialWhenTheStashWouldOverflow(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newSaleRig(t, class, d2dctl.StashLimit-200) // room for 200 more
			a, b, c := r.item(1, 3, 1, 100), r.item(2, 3, 1, 100), r.item(3, 3, 1, 100)
			res, v := r.sellNow(false, a, b, c)
			if okCount(res) != 2 || v == nil || v.Gold != 200 || r.pay.Stash != d2dctl.StashLimit {
				t.Fatalf("res %+v view %+v stash %d", res, v, r.pay.Stash)
			}
			if res[2].Code != "room" || !strings.Contains(res[2].Why, "room for 0") {
				t.Fatalf("the third says why: %+v", res[2])
			}
			if r.state(c).State != vault.Recorded {
				t.Fatal("the record that did not fit stays in the vault")
			}
			q := r.k.SaleQuote(context.Background(), []string{c})
			if len(q.Lines) != 0 || q.Reasons[0].Code != "room" || q.Reasons[0].N != 1 {
				t.Fatalf("quote of a full stash: %+v", q)
			}
		})
	}
}

func TestSellNowRefusesPinnedAndSaysWhy(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newSaleRig(t, class, 0)
			a, b := r.item(1, 3, 1, 40), r.item(2, 3, 1, 60)
			if err := r.st.Pin(b); err != nil {
				t.Fatal(err)
			}
			res, v := r.sellNow(false, a, b)
			if !res[0].OK || res[1].OK || res[1].Code != "pinned" || !strings.Contains(res[1].Why, "never sold") {
				t.Fatalf("res %+v", res)
			}
			if v.Gold != 40 || r.pay.Stash != 40 || r.state(b).State != vault.Recorded {
				t.Fatalf("view %+v stash %d", v, r.pay.Stash)
			}
		})
	}
}

func TestSellNowRefusedInAMeasuredRunAndWithoutTheForkVerbs(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newSaleRig(t, class, 0)
			a := r.item(1, 3, 1, 40)
			r.measured = true
			res, v := r.sellNow(true, a)
			if res[0].OK || res[0].Code != "measured" || v != nil || len(r.pay.Calls) != 0 || r.state(a).State != vault.Recorded {
				t.Fatalf("measured: %+v %v", res, r.pay.Calls)
			}
			r.measured = false
			r.pay.Measured = true // the fork's own marker refuses too
			res, v = r.sellNow(true, a)
			if res[0].OK || res[0].Code != "measured" || r.state(a).State != vault.Recorded || v == nil || v.State != SaleFailed {
				t.Fatalf("fork measured: %+v %+v", res, v)
			}
			// an older fork: no creditgold verb: the old queue is the way
			r.pay.Measured = false
			r.api.FlagOff = true
			r.now = r.now.Add(2 * time.Minute) // past the caps cache
			res, _ = r.sellNow(true, a)
			if res[0].OK || res[0].Code != "busy" && res[0].Code != "off" {
				t.Fatalf("no verbs: %+v", res)
			}
			if on, why := r.k.SaleOn(context.Background()); on || why == "" {
				t.Fatalf("SaleOn %v %q", on, why)
			}
			r.api.FlagOff = false
			r.instant = false
			if on, why := r.k.SaleOn(context.Background()); on || why != ReasonSaleOff {
				t.Fatalf("knob 0: %v %q", on, why)
			}
			if got := r.k.EnqueueSell(context.Background(), []string{a}, false); !got[0].OK || r.state(a).State != vault.SellQ {
				t.Fatalf("the old queue still works: %+v", got)
			}
		})
	}
}

func TestSellNowCreditRefusedKeepsTheRecords(t *testing.T) {
	r := newSaleRig(t, "sorceress", 0)
	a := r.item(1, 3, 1, 40)
	r.pay.FlagOff = true
	res, v := r.sellNow(false, a)
	if res[0].OK || v == nil || v.State != SaleFailed || r.state(a).State != vault.Recorded || res[0].Code != "payoff" {
		t.Fatalf("%+v %+v", res, v)
	}
	r.pay.FlagOff = false
	res, v = r.sellNow(false, a) // a failed order does not block the next press
	if !res[0].OK || v.State != SaleDone {
		t.Fatalf("%+v %+v", res, v)
	}
}

func TestSellNowAsksForAConfirmFromTheThreshold(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newSaleRig(t, class, 0)
			r.confirm = 500
			a, b := r.item(1, 3, 1, 300), r.item(2, 3, 1, 250)
			res, v := r.sellNow(false, a, b)
			if okCount(res) != 0 || v != nil || res[0].Code != "confirm" || len(r.pay.Calls) != 0 {
				t.Fatalf("%+v", res)
			}
			if q := r.k.SaleQuote(context.Background(), []string{a, b}); !q.NeedConfirm || q.StashAfter != 550 {
				t.Fatalf("quote %+v", q)
			}
			if res, _ := r.sellNow(true, a, b); okCount(res) != 2 || r.pay.Stash != 550 {
				t.Fatalf("with the ack: %+v stash %d", res, r.pay.Stash)
			}
			c := r.item(3, 3, 1, 100) // below the threshold: no ack needed
			if res, _ := r.sellNow(false, c); okCount(res) != 1 {
				t.Fatalf("small: %+v", res)
			}
		})
	}
}

// The old queue's orders (sold at the next town visit) are paid at once when the instant sale is on; marks are left alone.
func TestOrderedQueueFlushesInstantlyMarksStay(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newSaleRig(t, class, 0)
			a, b := r.item(1, 3, 1, 45), r.item(2, 3, 1, 55)
			if err := r.st.OrderSell(a); err != nil {
				t.Fatal(err)
			}
			if err := r.st.MarkSell(b); err != nil {
				t.Fatal(err)
			}
			lines := r.run()
			if r.state(a).State != vault.Sold || r.state(b).State != vault.SellQ || r.state(b).Ordered || r.pay.Stash != 45 || len(lines) == 0 {
				t.Fatalf("ordered sold, mark waits: %v %+v stash %d", lines, r.state(b), r.pay.Stash)
			}
			if r.api.Stats.Creates != 0 {
				t.Fatal("no item was created in the hero")
			}
		})
	}
}

func TestSaleQuoteExplainsEverySkippedRecord(t *testing.T) {
	r := newSaleRig(t, "paladin", 0)
	ok1, pinned, sold := r.item(1, 3, 1, 10), r.item(2, 3, 1, 10), r.item(3, 3, 1, 10)
	if err := r.st.Pin(pinned); err != nil {
		t.Fatal(err)
	}
	if res, _ := r.sellNow(false, sold); okCount(res) != 1 {
		t.Fatal("setup sale")
	}
	q := r.k.SaleQuote(context.Background(), []string{ok1, pinned, sold, "v-nothere"})
	got := map[string]int{}
	for _, x := range q.Reasons {
		got[x.Code] = x.N
	}
	if q.Asked != 4 || len(q.Lines) != 1 || got["pinned"] != 1 || got["state"] != 1 || got["missing"] != 1 {
		t.Fatalf("%+v", q)
	}
	var sum int
	for _, x := range q.Reasons {
		sum += x.N
	}
	if sum+len(q.Lines) != q.Asked {
		t.Fatalf("the reasons add up to what was asked: %+v", q)
	}
	_ = fmt.Sprint
}
