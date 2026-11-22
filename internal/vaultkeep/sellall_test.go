package vaultkeep

import (
	"context"
	"errors"
	"strings"
	"testing"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/vault"
)

// Sell all over the instant sale: the bot lists the records itself, under the sale lock, when the order runs. A Paladin and a
// Sorceress hero run every case.

// sellAllRig has plain records, pinned ones and ones held by an order (a keep, an equip order, a reveal, an older sale).
type sellAllRig struct {
	*saleRig
	plain, pinned, held []string
}

func newSellAllRig(t *testing.T, class string, stash int) *sellAllRig {
	r := &sellAllRig{saleRig: newSaleRig(t, class, stash)}
	for i := 1; i <= 4; i++ {
		r.plain = append(r.plain, r.item(i, 3, 1, 10*i)) // identified price: 10, 20, 30, 40
	}
	for i := 5; i <= 6; i++ {
		id := r.item(i, 3, 1, 100)
		if err := r.st.Pin(id); err != nil {
			t.Fatal(err)
		}
		r.pinned = append(r.pinned, id)
	}
	keep := r.item(7, 3, 1, 100)
	if err := r.st.Keep(keep); err != nil {
		t.Fatal(err)
	}
	equip := r.item(8, 3, 1, 100)
	r.k.mu.Lock()
	r.k.bookLocked().orders[equip] = &EquipOrder{ID: equip, State: EqQueued}
	r.k.mu.Unlock()
	r.held = []string{keep, equip}
	return r
}

func (r *sellAllRig) sellAll(id string) SellAllOutcome { return r.k.SellAll(context.Background(), id) }

func TestSellAllSellsExactlyTheUnmanagedAndSaysWhatItSkipped(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newSellAllRig(t, class, 1000)
			if n, left := r.k.SellAllCount(); n != 4 || len(left) != 2 {
				t.Fatalf("the button count: %d %+v", n, left)
			}
			out := r.sellAll("tap1")
			v := out.View
			if out.Code != "" || v == nil || v.State != SaleDone || v.Done != 4 || v.Gold != 100 || !v.All {
				t.Fatalf("out %+v view %+v", out, v)
			}
			if r.pay.Stash != 1100 || len(r.pay.Calls) != 1 || r.pay.Calls[0] != "credit 100 sa-tap1" {
				t.Fatalf("one credit of the sum, once: stash %d %v", r.pay.Stash, r.pay.Calls)
			}
			if want := "Sold 4 for 100 gold; 4 skipped: 2 pinned, 2 in an order"; v.Line != want {
				t.Fatalf("line %q want %q", v.Line, want)
			}
			for _, id := range r.plain {
				if r.state(id).State != vault.Sold {
					t.Fatalf("%s is not sold", id)
				}
			}
			for _, id := range append(append([]string{}, r.pinned...), r.held...) {
				if r.state(id).State == vault.Sold {
					t.Fatalf("%s was sold but is managed", id)
				}
			}
			if r.state(r.held[0]).State != vault.KeepQ || r.k.EquipOrderOf(r.held[1]) == nil {
				t.Fatal("the orders are intact")
			}
		})
	}
}

func TestSellAllSkipsARecordThatBecameQueuedBetweenTheTaps(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newSellAllRig(t, class, 1000)
			q := r.k.SellAllPlan(context.Background()) // what the first tap showed
			if len(q.Quote.Lines) != 4 {
				t.Fatalf("estimate %+v", q.Quote)
			}
			late := r.plain[1] // an equip order arrives before the second tap
			r.k.mu.Lock()
			r.k.bookLocked().orders[late] = &EquipOrder{ID: late, State: EqWithdraw}
			r.k.mu.Unlock()
			out := r.sellAll("tap2")
			if out.View == nil || out.View.Done != 3 || out.View.Gold != 80 || r.pay.Stash != 1080 {
				t.Fatalf("the late record is excluded: %+v", out.View)
			}
			if r.state(late).State != vault.Recorded || !strings.Contains(out.View.Line, "3 in an order") {
				t.Fatalf("late record %+v line %q", r.state(late), out.View.Line)
			}
		})
	}
}

func TestSellAllIsIdempotent(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newSellAllRig(t, class, 1000)
			first := r.sellAll("dbl")
			r.item(20, 3, 1, 500) // a new plain record appears; the SAME order id must not sell it
			again := r.sellAll("dbl")
			if first.View == nil || again.View == nil || again.View.ID != first.View.ID || again.View.Done != 4 || r.pay.Stash != 1100 || len(r.pay.Calls) != 1 {
				t.Fatalf("a resend of the same id sells nothing more: %+v / %+v stash %d %v", first.View, again.View, r.pay.Stash, r.pay.Calls)
			}
			// a lost reply: the credit may have happened, the order is kept, the same id settles it with the SAME fork order id
			r2 := newSellAllRig(t, class, 1000)
			r2.pay.Fail = errors.New("socket closed")
			lost := r2.sellAll("lost")
			if lost.View == nil || lost.View.State != SaleQueued || r2.state(r2.plain[0]).State != vault.Recorded {
				t.Fatalf("queued while unconfirmed: %+v", lost.View)
			}
			r2.pay.Fail = nil
			if done := r2.sellAll("lost"); done.View == nil || done.View.State != SaleDone || r2.pay.Stash != 1100 {
				t.Fatalf("settled: %+v stash %d", done.View, r2.pay.Stash)
			}
			if r2.pay.Calls[0] != r2.pay.Calls[len(r2.pay.Calls)-1] {
				t.Fatalf("same order id: %v", r2.pay.Calls)
			}
		})
	}
}

func TestSellAllPartialOnTheStashCapAndRefusedWhenFull(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newSellAllRig(t, class, d2dctl.StashLimit-60) // room for 60 of the 100
			out := r.sellAll("cap")
			if out.View == nil || out.View.Gold > 60 || out.View.Done == 0 || out.View.Done == 4 || r.pay.Stash > d2dctl.StashLimit {
				t.Fatalf("partial: %+v stash %d", out.View, r.pay.Stash)
			}
			if !strings.Contains(out.View.Line, "no room in the stash") {
				t.Fatalf("line %q", out.View.Line)
			}
			r2 := newSellAllRig(t, class, d2dctl.StashLimit)
			if out := r2.sellAll("full"); out.Code != "state" || out.View != nil || len(r2.pay.Calls) != 0 {
				t.Fatalf("a full stash sells nothing: %+v", out)
			}
		})
	}
}

func TestSellAllRefusedInAMeasuredRunAndWithoutTheForkVerbs(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newSellAllRig(t, class, 1000)
			r.measured = true
			if out := r.sellAll("m"); out.Code != "measured" || out.View != nil || len(r.pay.Calls) != 0 || r.state(r.plain[0]).State != vault.Recorded {
				t.Fatalf("measured: %+v", out)
			}
			r.measured = false
			r.api.FlagOff = true // an older fork: no creditgold verb
			r.now = r.now.Add(2 * 60 * 1e9)
			if out := r.sellAll("old"); out.Code == "" || out.View != nil || len(r.pay.Calls) != 0 {
				t.Fatalf("no verbs: %+v", out)
			}
			if ids := r.k.SellAllIDs(); len(ids) != 4 { // the old town-visit queue gets the same list
				t.Fatalf("fallback list %v", ids)
			}
		})
	}
}
