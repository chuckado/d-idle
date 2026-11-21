package vaultkeep

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/vault"
)

// chaos kills the "process" after limit game calls: every later call fails with a transport error until reset (a restart).
// With lost the call that hits the limit still takes effect in the game and only its reply is lost.
type chaos struct {
	n, limit int
	lost     bool
	dead     bool
}

var errLost = errors.New("connection lost")

func (c *chaos) reset() { c.n, c.dead = 0, false }

// gate reports whether the call may run (before) and whether to fail it after it ran (lostReply).
func (c *chaos) gate() (run, lostReply bool) {
	if c.dead {
		return false, false
	}
	c.n++
	if c.n > c.limit {
		c.dead = true
		return c.lost, c.lost
	}
	return true, false
}

type chaosAPI struct {
	f *flaky
	c *chaos
}

func (a chaosAPI) GroundFull(ctx context.Context, id int) ([]d2dctl.VaultRecord, error) {
	return a.f.GroundFull(ctx, id)
}
func (a chaosAPI) GroundTake(ctx context.Context, id int) (d2dctl.VaultRecord, error) {
	return a.f.GroundTake(ctx, id)
}
func (a chaosAPI) ItemFull(ctx context.Context, id int) ([]d2dctl.VaultRecord, error) {
	run, lost := a.c.gate()
	if !run {
		return nil, errLost
	}
	r, err := a.f.ItemFull(ctx, id)
	if lost {
		return nil, errLost
	}
	return r, err
}
func (a chaosAPI) ItemVault(ctx context.Context, id int) (d2dctl.VaultRecord, error) {
	run, lost := a.c.gate()
	if !run {
		return d2dctl.VaultRecord{}, errLost
	}
	r, err := a.f.ItemVault(ctx, id)
	if lost {
		return d2dctl.VaultRecord{}, errLost
	}
	return r, err
}
func (a chaosAPI) ItemCreate(ctx context.Context, h string, o d2dctl.CreateOpts) (d2dctl.CreateResult, error) {
	run, lost := a.c.gate()
	if !run {
		return d2dctl.CreateResult{}, errLost
	}
	r, err := a.f.ItemCreate(ctx, h, o)
	if lost {
		return d2dctl.CreateResult{}, errLost
	}
	return r, err
}
func (a chaosAPI) VaultStats(ctx context.Context) (d2dctl.VaultStats, error) {
	return a.f.VaultStats(ctx)
}
func (a chaosAPI) VaultCaps(ctx context.Context) (d2dctl.VaultCaps, error) { return a.f.VaultCaps(ctx) }

// fakeTools is the identify side: scrolls or tome charges, or Cain.
type fakeTools struct {
	api     *flaky
	charges int
	cain    bool
	err     error // the identify itself fails
	used    int
}

type chaosTools struct {
	t *fakeTools
	c *chaos
}

func (t chaosTools) Charges(context.Context) (int, error) { return t.t.charges, nil }
func (t chaosTools) Free() bool                           { return t.t.cain }
func (t chaosTools) Identify(_ context.Context, unit int, free bool) error {
	run, lost := t.c.gate()
	if !run {
		return errLost
	}
	if t.t.err != nil {
		return t.t.err
	}
	if !free {
		if t.t.charges <= 0 {
			return ErrNoTool
		}
		t.t.charges--
		t.t.used++
	}
	if err := t.t.api.VaultFake.Identify(unit); err != nil {
		return err
	}
	if lost {
		return errLost
	}
	return nil
}

// fakeVendor is the sell chore: it buys at price unless told otherwise.
type fakeVendor struct {
	api        *flaky
	gold       int
	price      int
	refuse     bool
	noGoldRise bool
	postpone   bool
	sold       int
}

type chaosVendor struct {
	v *fakeVendor
	c *chaos
}

func (v chaosVendor) Gold(context.Context) (int, error) { return v.v.gold, nil }
func (v chaosVendor) Sell(_ context.Context, units []int) error {
	run, lost := v.c.gate()
	if !run {
		return errLost
	}
	switch {
	case v.v.postpone:
		return fmt.Errorf("fight: %w", ErrPostpone)
	case v.v.refuse:
		return errors.New("the vendor would not buy it")
	}
	for _, u := range units {
		v.v.api.VaultFake.Remove(u)
		v.v.sold++
		if !v.v.noGoldRise {
			v.v.gold += v.v.price
		}
	}
	if lost {
		return errLost
	}
	return nil
}

func (r *rig) withTools(charges int) {
	r.tools = &fakeTools{api: r.api, charges: charges}
	r.k.Tools = chaosTools{r.tools, r.ch}
}

func (r *rig) withVendor(gold, price int) {
	r.vend = &fakeVendor{api: r.api, gold: gold, price: price}
	r.k.Vendor = chaosVendor{r.vend, r.ch}
}

// addUnid adds a magic item that was recorded unidentified.
func (r *rig) addUnid(i int, code string) string {
	r.t.Helper()
	rec := d2dctl.FakeRecord(i, code, 4, 30+i, fmt.Sprintf("roll%d", i))
	it, err := r.st.Add(vault.Item{Raw: rec.Norm, Code: code, Name: fmt.Sprintf("Unidentified %s %d", code, i), Quality: 4, Ilvl: 30 + i, Value: 40 + i})
	if err != nil {
		r.t.Fatal(err)
	}
	return it.ID
}

func (r *rig) heroEmpty() bool {
	recs, _ := r.api.VaultFake.ItemFull(context.Background(), -1)
	return len(recs) == 0
}

func (r *rig) reveal(ids ...string) []Result {
	return r.k.EnqueueReveal(context.Background(), ids, false)
}

func TestRevealFlow(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(3)
			id := r.addUnid(1, "wsp")
			old := r.state(id)
			if res := r.reveal(id); !res[0].OK {
				t.Fatalf("enqueue: %+v", res[0])
			}
			if again := r.reveal(id); !again[0].OK || !again[0].Already {
				t.Fatalf("second press must be idempotent: %+v", again[0])
			}
			if got := r.k.Pending(); got != 1 {
				t.Fatalf("pending %d", got)
			}
			lines := r.run()
			rec := r.state(id)
			if !rec.Identified || rec.Raw != d2dctl.IdentifiedForm(old.Raw) || rec.Hash == old.Hash || rec.ID != id {
				t.Fatalf("not superseded in place: %+v lines %v", rec, lines)
			}
			if rec.RevQ || rec.Stage != "" || rec.State != vault.Recorded {
				t.Fatalf("flags left: %+v", rec)
			}
			if !r.heroEmpty() || r.tools.charges != 2 {
				t.Fatalf("hero empty %v charges %d", r.heroEmpty(), r.tools.charges)
			}
			if s, _ := r.api.VaultStats(context.Background()); s.Creates != 1 || s.Vaults != 1 {
				t.Fatalf("stats %+v", s)
			}
			if st := r.st.Status(); st.Count != 1 {
				t.Fatalf("a superseded record must not duplicate: %+v", st)
			}
		})
	}
}

func TestRevealFreeAtCain(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(0)
			r.tools.cain = true
			id := r.addUnid(1, "rin")
			if res := r.reveal(id); !res[0].OK {
				t.Fatalf("free at Cain must queue with no scroll: %+v", res[0])
			}
			if p := r.k.RevealPrice(context.Background()); !p.Free || p.PerItem != 1 {
				t.Fatalf("price %+v", p)
			}
			r.run()
			if !r.state(id).Identified || r.tools.used != 0 || !r.heroEmpty() {
				t.Fatalf("identified %v used %d", r.state(id).Identified, r.tools.used)
			}
		})
	}
}

func TestRevealNoScroll(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(1)
			a, b := r.addUnid(1, "wsp"), r.addUnid(2, "wsp")
			res := r.reveal(a, b)
			if !res[0].OK || res[1].OK || res[1].Code != "noscroll" || res[1].Why != ReasonNoScroll {
				t.Fatalf("one charge covers one item: %+v", res)
			}
			// the charge disappears while the request waits: the visit fails it with the reason and creates nothing
			r.tools.charges = 0
			r.run()
			rec := r.state(a)
			if rec.RevQ || RevealReason(rec.Reason) != ReasonNoScroll || rec.Identified {
				t.Fatalf("want a failed reveal with the reason: %+v", rec)
			}
			if s, _ := r.api.VaultStats(context.Background()); s.Creates != 0 {
				t.Fatalf("nothing may be created without a tool: %+v", s)
			}
		})
	}
}

func TestRevealRefusals(t *testing.T) {
	r := newRig(t, "sorceress")
	r.withTools(2)
	id := r.add(1, "wsp", 2) // normal, identified
	if res := r.reveal(id); res[0].OK || res[0].Code != "state" {
		t.Fatalf("an identified record cannot be revealed: %+v", res[0])
	}
	r.ready = ReasonDead
	if ok, why := r.k.GateReveal(context.Background()); ok || why != ReasonDead {
		t.Fatalf("dead hero: %v %q", ok, why)
	}
	r2 := newRig(t, "paladin")
	if ok, _ := r2.k.GateReveal(context.Background()); ok {
		t.Fatal("no identifier: the gate must refuse")
	}
	r2.on = false
	r2.withTools(1)
	if ok, why := r2.k.GateReveal(context.Background()); ok || why != ReasonOff {
		t.Fatalf("off: %v %q", ok, why)
	}
}

func TestRevealIdentifyFailsLeavesRecord(t *testing.T) {
	r := newRig(t, "paladin")
	r.withTools(2)
	r.tools.err = errors.New("panel closed")
	id := r.addUnid(1, "wsp")
	old := r.state(id)
	r.reveal(id)
	r.run()
	rec := r.state(id)
	if rec.Identified || rec.Raw != old.Raw || rec.RevQ || !strings.Contains(RevealReason(rec.Reason), "panel closed") {
		t.Fatalf("record must be unchanged with the reason: %+v", rec)
	}
	if !r.heroEmpty() {
		t.Fatal("the temporary copy must be removed")
	}
}

func TestRevealKeepIsOneFlow(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(2)
			id := r.addUnid(1, "wsp")
			r.enq(id)
			if res := r.reveal(id); !res[0].OK {
				t.Fatalf("%+v", res[0])
			}
			if r.k.Pending() != 1 {
				t.Fatalf("keep and reveal of one record count once, pending %d", r.k.Pending())
			}
			r.run()
			rec := r.state(id)
			if rec.State != vault.Kept || !rec.Identified || !strings.HasPrefix(rec.Reason, "backpack") {
				t.Fatalf("want kept and identified in the backpack: %+v", rec)
			}
			if s, _ := r.api.VaultStats(context.Background()); s.Creates != 1 || s.Vaults != 0 {
				t.Fatalf("one create, no removal: %+v", s)
			}
			recs, _ := r.api.VaultFake.ItemFull(context.Background(), -1)
			if len(recs) != 1 || !recs[0].Identified || recs[0].Norm != rec.Raw {
				t.Fatalf("hero carries the identified item: %+v", recs)
			}
		})
	}
}

// TestRevealRestartAtEveryStep kills the process after each game call (with and without the last call taking effect) and
// restarts: the record ends identified, the hero is empty, one charge is spent, never a second copy.
func TestRevealRestartAtEveryStep(t *testing.T) {
	for _, class := range classes {
		for _, keep := range []bool{false, true} {
			for _, lost := range []bool{false, true} {
				for limit := 0; limit < 14; limit++ {
					t.Run(fmt.Sprintf("%s/keep%v/lost%v/%d", class, keep, lost, limit), func(t *testing.T) {
						r := newRig(t, class)
						r.ch = &chaos{limit: limit, lost: lost}
						r.open()
						r.withTools(1)
						id := r.addUnid(1, "wsp")
						old := r.state(id)
						if keep {
							r.enq(id)
						}
						if res := r.reveal(id); !res[0].OK {
							t.Fatalf("%+v", res[0])
						}
						r.run() // dies somewhere
						r.ch.limit = 1 << 30
						r.open()
						for i := 0; i < 4; i++ {
							r.run()
						}
						rec := r.state(id)
						if !rec.Identified || rec.Raw != d2dctl.IdentifiedForm(old.Raw) || rec.Stage != "" || rec.RevQ {
							t.Fatalf("not revealed: %+v", rec)
						}
						recs, _ := r.api.VaultFake.ItemFull(context.Background(), -1)
						switch {
						case keep && (rec.State != vault.Kept || len(recs) != 1):
							t.Fatalf("keep: state %s hero %d items", rec.State, len(recs))
						case !keep && (rec.State != vault.Recorded || len(recs) != 0):
							t.Fatalf("reveal only: state %s hero %d items", rec.State, len(recs))
						}
						if r.tools.used != 1 {
							t.Fatalf("charges spent %d, want exactly 1", r.tools.used)
						}
					})
				}
			}
		}
	}
}

func (r *rig) sell(ids ...string) []Result {
	return r.k.EnqueueSell(context.Background(), ids, false)
}

func TestSellFlow(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withVendor(100, 37)
			a, b := r.add(1, "wsp", 4), r.add(2, "rin", 4)
			for _, x := range r.sell(a, b) {
				if !x.OK {
					t.Fatalf("%+v", x)
				}
			}
			if again := r.sell(a); !again[0].Already {
				t.Fatal("second press must be idempotent")
			}
			lines := r.run()
			for _, id := range []string{a, b} {
				if rec := r.state(id); rec.State != vault.Sold || rec.Gold != 37 || rec.Stage != "" {
					t.Fatalf("%s: %+v lines %v", id, rec, lines)
				}
			}
			if r.vend.gold != 174 || !r.heroEmpty() {
				t.Fatalf("gold %d empty %v", r.vend.gold, r.heroEmpty())
			}
			if in := r.k.Info(); in.Sold != 2 || in.GoldEarned != 74 {
				t.Fatalf("info %+v", in)
			}
		})
	}
}

func TestSellGoldReadbackMismatch(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withVendor(100, 37)
			r.vend.noGoldRise = true
			a, b := r.add(1, "wsp", 4), r.add(2, "rin", 4)
			r.sell(a, b)
			r.run()
			rec := r.state(a)
			if rec.State != vault.Sold || rec.Gold != 0 || !strings.Contains(SellReason(rec.Reason), "did not rise") {
				t.Fatalf("a vanished item with no gold is closed with a note: %+v", rec)
			}
			if r.state(b).State != vault.SellQ {
				t.Fatalf("sales stop after a mismatch: %+v", r.state(b))
			}
			if r.k.Info().Tripped == "" || r.k.NeedsText() == "" {
				t.Fatal("the mismatch must trip the keeper and show on Needs-you")
			}
			if ok, _ := r.k.GateSell(context.Background()); ok {
				t.Fatal("gate must refuse while tripped")
			}
		})
	}
}

func TestSellVendorRefusesAndPostpones(t *testing.T) {
	r := newRig(t, "paladin")
	r.withVendor(100, 37)
	a := r.add(1, "wsp", 4)
	r.sell(a)
	r.vend.postpone = true
	r.run()
	if rec := r.state(a); rec.State != vault.SellQ || rec.Stage != "" || !r.heroEmpty() {
		t.Fatalf("postponed: stays queued, nothing left in the hero: %+v", rec)
	}
	r.vend.postpone, r.vend.refuse = false, true
	r.run()
	rec := r.state(a)
	if rec.State != vault.Recorded || SellReason(rec.Reason) == "" || !r.heroEmpty() {
		t.Fatalf("refused: back in the vault with a reason: %+v", rec)
	}
}

func TestSellNeverPinned(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withVendor(100, 37)
			a := r.add(1, "wsp", 4)
			if err := r.st.Pin(a); err != nil {
				t.Fatal(err)
			}
			if res := r.sell(a); res[0].OK || res[0].Code != "pinned" {
				t.Fatalf("pinned must be refused: %+v", res[0])
			}
			if err := r.st.MarkSell(a); !errors.Is(err, vault.ErrPinned) {
				t.Fatalf("the store refuses too: %v", err)
			}
			r.run()
			if s, _ := r.api.VaultStats(context.Background()); s.Creates != 0 || r.vend.sold != 0 {
				t.Fatalf("nothing may be created or sold: %+v", s)
			}
		})
	}
}

func TestSellPerVisitCapAndPackFull(t *testing.T) {
	r := newRig(t, "sorceress")
	r.withVendor(0, 10)
	r.k.SellPerVisit = func() int { return 2 }
	var ids []string
	for i := 1; i <= 5; i++ {
		ids = append(ids, r.add(i, "wsp", 4))
	}
	r.sell(ids...)
	r.run()
	if n := len(r.st.SellQueued()); n != 3 || r.vend.sold != 2 {
		t.Fatalf("cap 2: %d queued, %d sold", n, r.vend.sold)
	}
	r.api.PackCols, r.api.PackRows = 0, 0
	r.run()
	if len(r.st.SellQueued()) != 3 || !strings.Contains(r.k.Info().Waiting, "backpack is full") {
		t.Fatalf("a full pack keeps the sale queued: %+v", r.k.Info())
	}
}

// TestSellRestartAtEveryStep: a crash anywhere ends with the record sold exactly once and nothing left in the hero.
func TestSellRestartAtEveryStep(t *testing.T) {
	for _, class := range classes {
		for _, lost := range []bool{false, true} {
			for limit := 0; limit < 10; limit++ {
				t.Run(fmt.Sprintf("%s/lost%v/%d", class, lost, limit), func(t *testing.T) {
					r := newRig(t, class)
					r.ch = &chaos{limit: limit, lost: lost}
					r.open()
					r.withVendor(100, 37)
					a := r.add(1, "wsp", 4)
					r.sell(a)
					r.run()
					r.ch.limit = 1 << 30
					r.open()
					for i := 0; i < 4; i++ {
						r.run()
					}
					rec := r.state(a)
					if rec.State != vault.Sold || !r.heroEmpty() || r.vend.sold != 1 {
						t.Fatalf("state %s sold %d empty %v", rec.State, r.vend.sold, r.heroEmpty())
					}
					if r.vend.gold != 137 || rec.Gold > 37 {
						t.Fatalf("gold %d record %d", r.vend.gold, rec.Gold)
					}
				})
			}
		}
	}
}

func TestSellOffer(t *testing.T) {
	r := newRig(t, "paladin")
	r.withVendor(0, 10)
	short, prompt := 150, false
	r.k.ShortBy = func() int { return short }
	r.k.Prompt = func() bool { return prompt }
	low := r.addUnid(1, "wsp")
	rare := r.add(2, "rin", 6)
	pinned := r.add(3, "amu", 4)
	_ = r.st.Pin(pinned)
	if r.k.Offer() != nil {
		t.Fatal("the prompt is off by default")
	}
	prompt = true
	o := r.k.Offer()
	if o == nil || len(o.IDs) != 1 || o.IDs[0] != low || o.Short != 150 {
		t.Fatalf("offer %+v: only the unpinned magic item, never rare or pinned (%s %s)", o, rare, pinned)
	}
	short = 0
	if r.k.Offer() != nil {
		t.Fatal("no shortage, no prompt")
	}
	short = 150
	r.k.DismissOffer()
	if r.k.Offer() != nil {
		t.Fatal("dismissed prompt stays quiet")
	}
	r.now = r.now.Add(time.Hour)
	if r.k.Offer() == nil {
		t.Fatal("the mute ends")
	}
}
