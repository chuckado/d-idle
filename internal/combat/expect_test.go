package combat

import (
	"context"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

func expectMode(n int) func(*env.Tuning) { return func(t *env.Tuning) { t.Combat.DeliveryExpect = n } }

type simRun struct {
	keys        []string
	moves, cast int
	end         time.Time
}

func expectRun(t *testing.T, mode int, refuse bool) (simRun, *harness) {
	h := novaHarness(t, expectMode(mode))
	h.s.castCost = 20 // at least any skill's catalog cost: a cast that happened drops the pool by its price
	if refuse {
		h.s.castCost = 1e6 // every paid cast is refused: the pool never drops
	}
	for i := 0; i < 3; i++ {
		h.s.addMon(uint32(10+i), clsFallen, 101+float64(i), 101, 60, 1, 0)
	}
	h.f.Clear(context.Background(), ClearOpts{Timeout: 8 * time.Second})
	return simRun{append([]string(nil), h.s.castKeys...), len(h.s.moves), h.s.casts, h.f.now()}, h
}

// INVARIANT (golden parity, knob 0 against 1): recording and logging never change what the hero does.
func TestExpectLogOnlyChangesNoDecision(t *testing.T) {
	off, _ := expectRun(t, 0, false)
	on, h := expectRun(t, 1, false)
	if strings.Join(off.keys, ",") != strings.Join(on.keys, ",") || off.moves != on.moves || off.cast != on.cast || !off.end.Equal(on.end) {
		t.Fatalf("off %+v\non  %+v", off, on)
	}
	if len(off.keys) == 0 {
		t.Fatal("no casts")
	}
	ok := 0
	for _, c := range h.f.ExpectCounters() {
		ok += c.OK
	}
	if ok == 0 {
		t.Error("mode 1 must settle paid casts as met")
	}
}

func TestExpectCountersStayEmptyWhenOff(t *testing.T) {
	_, h := expectRun(t, 0, false)
	if n := len(h.f.ExpectCounters()); n != 0 {
		t.Errorf("mode 0 recorded %d skills", n)
	}
}

// A refusing simulator: mode 1 counts the failures and demotes in the books only; mode 2 acts on them.
func TestExpectRefusedCastsDemoteOnlyInModeTwo(t *testing.T) {
	_, h1 := expectRun(t, 1, true)
	_, h2 := expectRun(t, 2, true)
	demoted := func(h *harness) (n int) {
		for _, c := range h.f.ExpectCounters() {
			n += c.Demotions
		}
		return n
	}
	if demoted(h1) == 0 || demoted(h2) == 0 {
		t.Fatalf("refused casts must reach the demotion threshold: mode1 %v mode2 %v", h1.f.ExpectCounters(), h2.f.ExpectCounters())
	}
	for n := range h1.f.ExpectCounters() {
		if h1.f.expectDemoted(n) {
			t.Errorf("mode 1 acted on a demotion of %s", n)
		}
	}
	acted := false
	for n := range h2.f.ExpectCounters() {
		acted = acted || h2.f.expectDemoted(n)
	}
	if !acted {
		t.Error("mode 2 must report the demoted skill")
	}
}

// The ledger knows no class: skills of two classes strike out and recover the same way.
func TestExpectLedgerStrikesAndBackoff(t *testing.T) {
	for _, cl := range []d2.Class{d2.Sorceress, d2.Paladin} {
		pr := ProfileFor(cl, "")
		names := pr.AttackNames(60, func(string) bool { return true })
		if len(names) == 0 {
			t.Fatalf("%v: no attack name", cl)
		}
		name := names[0]
		h := novaHarness(t, expectMode(2))
		f := h.f
		k := expectKey{name, expectReasonMana}
		now := f.now()
		for i := 0; i < expectStrikes-1; i++ {
			f.expectSettle(k, false, false, now)
		}
		if f.expectDemoted(name) {
			t.Fatalf("%s: demoted before %d strikes", name, expectStrikes)
		}
		f.expectSettle(k, false, false, now)
		if !f.expectDemoted(name) {
			t.Fatalf("%s: not demoted after %d strikes", name, expectStrikes)
		}
		// the back-off ends and the skill plays again; a second demotion lasts longer.
		h.s.t = h.s.t.Add(expectBackoff + time.Second)
		if f.expectDemoted(name) {
			t.Fatalf("%s: still demoted after the back-off", name)
		}
		for i := 0; i < expectStrikes; i++ {
			f.expectSettle(k, false, false, f.now())
		}
		if b := f.expect.backoff[k]; b != 2*expectBackoff {
			t.Errorf("%s: second back-off %v, want %v", name, b, 2*expectBackoff)
		}
		// a met expectation clears the books; a cut says nothing about the skill.
		f.expectSettle(k, true, false, f.now())
		f.expectSettle(k, false, true, f.now())
		if f.expect.fails[k] != 0 || f.expect.backoff[k] != 0 {
			t.Errorf("%s: a met cast must reset strikes and back-off: %v %v", name, f.expect.fails[k], f.expect.backoff[k])
		}
	}
}
