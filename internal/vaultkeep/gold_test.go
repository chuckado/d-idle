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
)

// fakePayer is the payment side of a gold order over d2dctl.PayFake (the same refusals as the fork's `debug paygold`).
type fakePayer struct {
	*d2dctl.PayFake
	price   int
	readErr error
}

func (p *fakePayer) Price() int { return p.price }
func (p *fakePayer) Stash(context.Context) (int, error) {
	if p.readErr != nil {
		return 0, p.readErr
	}
	return p.Stash0(), nil
}

// Stash0 is the stash gold the fork holds right now.
func (p *fakePayer) Stash0() int { return p.PayFake.Stash }

func (r *rig) wirePay() {
	if r.pay == nil {
		return
	}
	r.k.Payer = r.pay
	r.k.GoldOn = func() bool { return r.goldOn }
	r.k.GoldConfirm = func() int { return r.confirm }
	r.k.OrderFile = func() string { return filepath.Join(r.dir, "revealgold.json") }
	r.k.Measured = func() bool { return r.measured }
}

func (r *rig) withPay(stash, price int) {
	r.goldOn = true
	r.pay = &fakePayer{PayFake: &d2dctl.PayFake{Stash: stash, Purse: 300}, price: price}
	r.wirePay()
}

// sealeds records n unidentified items whose affixes the vault holds.
func (r *rig) sealeds(n int) []string {
	var ids []string
	for i := 1; i <= n; i++ {
		ids = append(ids, r.addSealed(i, "wsp"))
	}
	return ids
}

func (r *rig) revealed(ids []string) int {
	n := 0
	for _, id := range ids {
		if r.state(id).Identified {
			n++
		}
	}
	return n
}

func (r *rig) gold(ids []string, maxGold int, confirm bool) ([]Result, *GoldView) {
	return r.k.EnqueueRevealGold(context.Background(), ids, maxGold, confirm)
}

// The order pays from the stash at once, reveals in the same call, buys no scroll and visits no town.
func TestGoldRevealPaysFromTheStashAndRevealsAtOnce(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(0) // no scroll, no tome charge, Cain not here: gold must not need any
			r.withPay(7200, 80)
			ids := r.sealeds(5)
			res, v := r.gold(ids, 0, false)
			for _, x := range res {
				if !x.OK {
					t.Fatalf("refused: %+v", x)
				}
			}
			if v == nil || v.State != OrderRevealed || v.Done != 5 || v.Paid != 400 || v.StashAfter != 6800 {
				t.Fatalf("view %+v", v)
			}
			if !strings.Contains(v.Line, "Stash gold: 6,800") {
				t.Fatalf("line %q", v.Line)
			}
			if r.revealed(ids) != 5 || r.pay.Stash0() != 6800 || r.pay.Purse != 300 {
				t.Fatalf("revealed %d stash %d purse %d", r.revealed(ids), r.pay.Stash0(), r.pay.Purse)
			}
			if len(r.pay.Calls) != 1 || r.pay.Calls[0] != fmt.Sprintf("400 %s", v.ID) {
				t.Fatalf("one payment of the whole order: %v", r.pay.Calls)
			}
			if r.k.GoldPending() || r.k.Pending() != 0 || r.tools.charges != 0 || !r.heroEmpty() {
				t.Fatal("nothing is queued, no scroll or tome charge is used, nothing is created in the hero")
			}
			for _, id := range ids {
				if rec := r.state(id); rec.Props == "" || rec.Sealed != "" || !rec.Unlocked {
					t.Fatalf("the affixes appear only now: %+v", rec)
				}
			}
			if lines := r.run(); len(lines) != 0 {
				t.Fatalf("a town visit finds nothing to do for a paid order: %v", lines)
			}
		})
	}
}

// The stash pays for as many as it can, in the player's list order, and says so.
func TestGoldRevealPartialInListOrder(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(0)
			r.withPay(250, 80) // three of five
			ids := r.sealeds(5)
			want := []string{ids[3], ids[1], ids[4], ids[0], ids[2]} // the player's order, not the vault's
			res, v := r.gold(want, 0, false)
			if v == nil || v.State != OrderPartial || v.Done != 3 || v.Want != 5 || v.Paid != 240 || v.StashAfter != 10 {
				t.Fatalf("view %+v", v)
			}
			if !strings.Contains(v.Line, "3 of 5") || !strings.Contains(v.Note, "stash paid for those") {
				t.Fatalf("line %q note %q", v.Line, v.Note)
			}
			for i, x := range res {
				if (i < 3) != x.OK || (i >= 3 && x.Code != "gold") {
					t.Fatalf("result %d: %+v", i, x)
				}
			}
			if !r.state(ids[3]).Identified || !r.state(ids[1]).Identified || !r.state(ids[4]).Identified || r.state(ids[0]).Identified || r.state(ids[2]).Identified {
				t.Fatal("the first three of the player's list are revealed, the rest stay sealed")
			}
		})
	}
}

func TestGoldRevealOverdraftAndEmptyStashRefused(t *testing.T) {
	for _, class := range classes {
		r := newRig(t, class)
		r.withTools(0)
		r.withPay(79, 80)
		ids := r.sealeds(2)
		res, v := r.gold(ids, 0, false)
		if v != nil || res[0].OK || res[0].Code != "gold" || !strings.Contains(res[0].Why, "short of gold") || len(r.pay.Calls) != 0 || r.revealed(ids) != 0 || r.pay.Stash0() != 79 {
			t.Fatalf("%s: %+v %+v calls %v", class, res[0], v, r.pay.Calls)
		}
		// the fork refuses after the bot's read was stale: the unlock is taken back, the order ends failed with the reason
		r.pay.PayFake.Stash = 500
		r.pay.price = 80
		r.k.Payer = &staleRead{fakePayer: r.pay, shown: 500, real: 100}
		_, v = r.gold(ids, 0, false)
		if v == nil || v.State != OrderFailed || !strings.Contains(v.Note, "short of gold") || r.revealed(ids) != 0 {
			t.Fatalf("%s: refused payment: %+v", class, v)
		}
		for _, id := range ids {
			if rec := r.state(id); rec.Unlocked || rec.Sealed == "" {
				t.Fatalf("%s: a refused payment must leave the item sealed: %+v", class, rec)
			}
		}
		if r.k.GoldPending() {
			t.Fatal("a failed order must not keep the bot busy")
		}
	}
}

// staleRead shows the bot a bigger stash than the fork holds.
type staleRead struct {
	*fakePayer
	shown, real int
}

func (s *staleRead) Stash(context.Context) (int, error) { return s.shown, nil }
func (s *staleRead) PayGold(ctx context.Context, amount int, order string) (d2dctl.PayResult, error) {
	s.PayFake.Stash = s.real
	return s.fakePayer.PayGold(ctx, amount, order)
}

// A repeated press, or a repeated order id at the fork, is never charged twice.
func TestGoldRevealIdempotent(t *testing.T) {
	for _, class := range classes {
		r := newRig(t, class)
		r.withTools(0)
		r.withPay(1000, 100)
		ids := r.sealeds(3)
		_, v := r.gold(ids, 0, false)
		again, _ := r.gold(ids, 0, false)
		if !again[0].OK || !again[0].Already || !again[2].Already || r.pay.Stash0() != 700 || len(r.pay.Calls) != 1 {
			t.Fatalf("%s: second press %+v stash %d calls %v", class, again, r.pay.Stash0(), r.pay.Calls)
		}
		// the same order id at the fork: paid 0, already
		res, err := r.pay.PayGold(context.Background(), 300, v.ID)
		if err != nil || res.Paid != 0 || !res.Already || r.pay.Stash0() != 700 {
			t.Fatalf("%s: repeat of the order id: %+v %v stash %d", class, res, err, r.pay.Stash0())
		}
	}
}

// A lost reply: the order stays queued, the gold may already be gone; the settle repeats the same order id and ends it.
func TestGoldRevealLostReplyIsSettledOnce(t *testing.T) {
	r := newRig(t, "paladin")
	r.withTools(0)
	r.withPay(1000, 100)
	ids := r.sealeds(2)
	lossy := &lostReply{fakePayer: r.pay}
	r.k.Payer = lossy
	_, v := r.gold(ids, 0, false)
	if v == nil || v.State != OrderQueued || !r.k.GoldPending() {
		t.Fatalf("an unanswered payment keeps the order queued: %+v", v)
	}
	if r.pay.Stash0() != 800 { // the fork took the gold, the reply never came
		t.Fatalf("stash %d", r.pay.Stash0())
	}
	r.open() // the bot restarts; the order is read back from the file
	r.wirePay()
	r.k.Payer = lossy
	lossy.lose = false
	if lines := r.run(); len(lines) == 0 {
		t.Fatal("the visit settles the order")
	}
	o := r.k.GoldView()
	if o == nil || o.State != OrderRevealed || r.pay.Stash0() != 800 || r.revealed(ids) != 2 || o.Paid != 200 {
		t.Fatalf("%+v stash %d", o, r.pay.Stash0())
	}
	if len(r.pay.Calls) != 2 || r.pay.Calls[0] != r.pay.Calls[1] {
		t.Fatalf("the retry carries the same order id: %v", r.pay.Calls)
	}
}

// lostReply takes the gold at the fork but answers with a transport error once.
type lostReply struct {
	*fakePayer
	lose bool
	done bool
}

func (l *lostReply) PayGold(ctx context.Context, amount int, order string) (d2dctl.PayResult, error) {
	res, err := l.fakePayer.PayGold(ctx, amount, order)
	if !l.done {
		l.done, l.lose = true, true
	}
	if l.lose && err == nil {
		l.lose = false
		return d2dctl.PayResult{}, errors.New("read: connection reset")
	}
	return res, err
}

// A measured run never pays: the bot refuses before the fork is asked, and the fork's own refusal is handled too.
func TestGoldRevealMeasuredRunRefuses(t *testing.T) {
	for _, class := range classes {
		r := newRig(t, class)
		r.withTools(0)
		r.withPay(1000, 100)
		ids := r.sealeds(2)
		r.measured = true
		res, v := r.gold(ids, 0, false)
		if v != nil || res[0].OK || res[0].Code != "measured" || len(r.pay.Calls) != 0 || r.pay.Stash0() != 1000 || r.revealed(ids) != 0 {
			t.Fatalf("%s: %+v %+v", class, res[0], v)
		}
		if f := r.k.GoldFacts(context.Background()); f.On {
			t.Fatal("the bar must not offer gold in a measured run")
		}
		r.measured = false
		r.pay.PayFake.Measured = true // the fork says measured though the bot did not know
		res, v = r.gold(ids, 0, false)
		if v == nil || v.State != OrderFailed || res[0].Code != "measured" || r.revealed(ids) != 0 || r.pay.Stash0() != 1000 {
			t.Fatalf("%s: fork refusal %+v %+v", class, res[0], v)
		}
		if rec := r.state(ids[0]); rec.Unlocked {
			t.Fatal("a refused payment must not leave an unlocked item")
		}
	}
}

func TestGoldRevealFlagOffAtTheFork(t *testing.T) {
	r := newRig(t, "sorceress")
	r.withTools(0)
	r.withPay(1000, 100)
	r.pay.PayFake.FlagOff = true
	ids := r.sealeds(1)
	res, v := r.gold(ids, 0, false)
	if v == nil || v.State != OrderFailed || res[0].Code != "payoff" || !strings.Contains(v.Note, "--vault-pay") || r.revealed(ids) != 0 {
		t.Fatalf("%+v %+v", res[0], v)
	}
}

// The price rule is the vendor price of one Scroll of Identify per item; an unknown price or an unsealed item refuses.
func TestGoldRevealRefusals(t *testing.T) {
	r := newRig(t, "sorceress")
	r.withTools(0)
	r.withPay(5000, 100)
	ctx := context.Background()
	ids := r.sealeds(3)
	plain := r.addUnid(9, "wsp")
	r.goldOn = false
	if res, _ := r.gold(ids, 0, false); res[0].Code != "off" {
		t.Fatalf("knob 0: %+v", res[0])
	}
	r.goldOn, r.confirm = true, 250
	if res, _ := r.gold(ids, 0, false); res[0].Code != "confirm" || len(r.pay.Calls) != 0 {
		t.Fatalf("a 300 gold order needs confirm: %+v", res[0])
	}
	if res, _ := r.gold([]string{plain}, 0, false); res[0].Code != "nosealed" {
		t.Fatalf("no held affixes: %+v", res[0])
	}
	if res, _ := r.gold(ids, 150, true); !res[0].OK || res[1].OK || res[1].Code != "gold" || res[2].OK {
		t.Fatalf("a cap of 150 pays one item: %+v", res)
	}
	if res, _ := r.gold(ids[2:], 50, true); res[0].Code != "gold" || !strings.Contains(res[0].Why, "cap") {
		t.Fatalf("a cap below one reveal: %+v", res[0])
	}
	r.pay.price = 0
	if res, _ := r.gold(ids[2:], 0, true); res[0].Code != "noshop" {
		t.Fatalf("unknown price: %+v", res[0])
	}
	r.pay.price, r.pay.readErr = 100, errors.New("no snapshot")
	if res, _ := r.gold(ids[2:], 0, true); res[0].Code != "state" {
		t.Fatalf("unreadable stash: %+v", res[0])
	}
	if res, _ := r.k.EnqueueRevealGold(ctx, nil, 0, false); len(res) != 0 {
		t.Fatal("no ids")
	}
}

func TestGoldRevealSurvivesARestartAndFadesWhenOld(t *testing.T) {
	r := newRig(t, "paladin")
	r.withTools(0)
	r.withPay(2000, 80)
	ids := r.sealeds(2)
	r.gold(ids, 0, false)
	r.open() // the bot restarts: the finished order is read back from the file
	r.wirePay()
	if o := r.k.GoldView(); o == nil || o.State != OrderRevealed || r.k.GoldPending() {
		t.Fatalf("%+v", o)
	}
	r.now = r.now.Add(time.Hour)
	if r.k.GoldView() != nil {
		t.Fatal("an old result must not stay on the page")
	}
}

func TestGoldCancelOnlyBeforePayment(t *testing.T) {
	r := newRig(t, "sorceress")
	r.withTools(0)
	r.withPay(1000, 100)
	ids := r.sealeds(2)
	if _, ok := r.k.CancelGold(); ok {
		t.Fatal("nothing to cancel")
	}
	r.k.Payer = &lostReply{fakePayer: r.pay}
	r.gold(ids, 0, false) // unanswered: queued, items unlocked
	v, ok := r.k.CancelGold()
	if !ok || v.State != OrderCancelled || r.revealed(ids) != 0 {
		t.Fatalf("cancel %+v %v", v, ok)
	}
	r.k.Payer = r.pay
	r.gold(ids, 0, false)
	if _, ok := r.k.CancelGold(); ok {
		t.Fatal("a paid order cannot be cancelled")
	}
}

// Cain standing here: the page offers the free reveal and no gold; gold orders need the payer wired.
func TestGoldNoPayerIsNoShop(t *testing.T) {
	r := newRig(t, "paladin")
	r.withTools(0)
	ids := r.sealeds(1)
	if res, _ := r.gold(ids, 0, false); res[0].Code != "noshop" {
		t.Fatalf("%+v", res[0])
	}
}

func (r *rig) goldAll(maxGold int, confirm bool) ([]Result, *GoldView) {
	r.now = r.now.Add(time.Second) // a new order id per press, as the real clock gives
	return r.k.EnqueueRevealGoldAll(context.Background(), maxGold, confirm)
}

// Identify all (scope all_unrevealed): the keeper lists the whole vault itself; identified records and the ones already
// unlocked are left alone, the exact gold leaves the stash once, for both classes.
func TestGoldRevealAllScopeTakesTheWholeVault(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(0)
			r.withPay(10000, 80)
			ids := r.sealeds(30)
			if _, v := r.gold(ids[:3], 0, false); v == nil || v.State != OrderRevealed { // three revealed earlier: not counted again
				t.Fatalf("setup %+v", v)
			}
			res, v := r.goldAll(0, false)
			if v == nil || v.State != OrderRevealed || v.Done != 27 || v.Want != 27 || v.Paid != 27*80 || v.StashAfter != 10000-30*80 {
				t.Fatalf("view %+v", v)
			}
			if len(res) != 27 {
				t.Fatalf("one result per paid record: %d", len(res))
			}
			if r.revealed(ids) != 30 || r.pay.Stash0() != 10000-30*80 || len(r.pay.Calls) != 2 {
				t.Fatalf("revealed %d stash %d calls %v", r.revealed(ids), r.pay.Stash0(), r.pay.Calls)
			}
			// nothing is left: a refusal, no payment
			res, v = r.goldAll(0, false)
			if v != nil || len(res) != 1 || res[0].Code != "state" || len(r.pay.Calls) != 2 {
				t.Fatalf("empty vault: %+v %+v calls %v", res, v, r.pay.Calls)
			}
		})
	}
}

// A stash that is short pays for the best K, says "partial" with the asked count, and the rest stays sealed for later.
func TestGoldRevealAllScopePartialWhenShort(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(0)
			r.withPay(80*12+30, 80)
			ids := r.sealeds(20)
			res, v := r.goldAll(0, false)
			if v == nil || v.State != OrderPartial || v.Done != 12 || v.Want != 20 || v.Paid != 960 || v.StashAfter != 30 {
				t.Fatalf("view %+v", v)
			}
			if len(res) != 12 || r.revealed(ids) != 12 {
				t.Fatalf("results %d revealed %d", len(res), r.revealed(ids))
			}
			// best value first: sealeds() values rise with the index, so the top twelve are the last twelve
			for i, id := range ids {
				if got := r.state(id).Identified; got != (i >= 8) {
					t.Fatalf("record %d revealed=%v", i, got)
				}
			}
			// an empty stash: refused, nothing paid
			res, v = r.goldAll(0, false)
			if v != nil || res[0].OK || res[0].Code != "gold" || len(r.pay.Calls) != 1 {
				t.Fatalf("short stash: %+v %+v", res, v)
			}
		})
	}
}

// The cap and the confirm threshold apply to the all scope; a repeated press takes only what is still unrevealed.
func TestGoldRevealAllScopeCapConfirmAndRepeat(t *testing.T) {
	r := newRig(t, classes[0])
	r.withTools(0)
	r.withPay(100000, 80)
	r.confirm = 1000
	ids := r.sealeds(30)
	if res, v := r.goldAll(0, false); v != nil || res[0].Code != "confirm" || len(r.pay.Calls) != 0 {
		t.Fatalf("confirm: %+v %+v", res, v)
	}
	if res, v := r.goldAll(400, true); v == nil || v.Done != 5 || len(res) != 5 || v.Paid != 400 {
		t.Fatalf("cap of five: %+v %+v", res, v)
	}
	before := r.pay.Stash0()
	if res, v := r.goldAll(0, true); v == nil || v.State != OrderRevealed || before-r.pay.Stash0() != 25*80 || len(res) != 25 {
		t.Fatalf("rest: %+v %+v", res, v)
	}
	if r.revealed(ids) != 30 {
		t.Fatal("all revealed")
	}
}

// A lost reply to Identify all, then the same press again: the running order is settled with its own order id, the stash
// is charged once and the paid items are revealed once.
func TestGoldRevealAllScopeResendIsIdempotent(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(0)
			r.withPay(5000, 100)
			ids := r.sealeds(6)
			lossy := &lostReply{fakePayer: r.pay}
			r.k.Payer = lossy
			if _, v := r.goldAll(0, false); v == nil || v.State != OrderQueued {
				t.Fatalf("queued: %+v", v)
			}
			if r.pay.Stash0() != 4400 {
				t.Fatalf("the fork took the gold: %d", r.pay.Stash0())
			}
			res, v := r.goldAll(0, false) // the page sends the same order again
			if v == nil || v.State != OrderRevealed || v.Paid != 600 || r.pay.Stash0() != 4400 || r.revealed(ids) != 6 {
				t.Fatalf("settled: %+v stash %d", v, r.pay.Stash0())
			}
			if len(res) != 6 || !res[0].Already {
				t.Fatalf("answer %+v", res)
			}
			if len(r.pay.Calls) != 2 || r.pay.Calls[0] != r.pay.Calls[1] {
				t.Fatalf("one order id for both sends: %v", r.pay.Calls)
			}
		})
	}
}
