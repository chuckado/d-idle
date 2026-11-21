package questboard

import (
	"strings"
	"testing"
	"time"
)

var smClasses = []string{"Paladin", "Sorceress"}

// parkRig parks den_of_evil/clear_den by three failed slices.
func parkRig(t *testing.T, class string, sm bool) *rig {
	r := newRigSM(t, class, 6, sm)
	if err := r.x.Accept(den, clr); err != nil {
		t.Fatal(err)
	}
	fail := StepResult{StepFailed, "no path to the Den", ""}
	r.f.Script = []StepResult{fail, fail, fail}
	r.tick(3)
	if d := r.next(); d.Kind != DecBlocked {
		t.Fatalf("not parked: %+v", d)
	}
	return r
}

// Q5: a park survives a restart with the machine, and (the contrast) is forgotten without it.
func TestSMParkSurvivesRestart(t *testing.T) {
	skipSM(t, "the contrast half runs with the machine off")
	for _, class := range smClasses {
		t.Run(class, func(t *testing.T) {
			on := parkRig(t, class, true)
			on.restart()
			if d := on.next(); d.Kind != DecBlocked || !strings.Contains(d.Why, "gave up after 3 tries") {
				t.Fatalf("machine on: park lost over a restart: %+v", d)
			}
			off := parkRig(t, class, false)
			off.restart()
			if d := off.next(); d.Kind == DecBlocked {
				t.Fatalf("machine off: the old behaviour (a restart clears the park) changed: %+v", d)
			}
		})
	}
}

// Retry clears the stored park and lifts the task executor's block of the group.
type unparkActions struct {
	*Fake
	got []string
}

func (u *unparkActions) Unpark(q Quest, c Chunk) { u.got = append(u.got, ChunkKey(q.ID, c.ID)) }

func TestSMRetryClearsParkAndExecutorGroup(t *testing.T) {
	for _, class := range smClasses {
		t.Run(class, func(t *testing.T) {
			r := parkRig(t, class, true)
			ua := &unparkActions{Fake: r.f}
			r.x.act = ua
			r.x.Unblock(den, clr)
			if len(ua.got) != 1 || ua.got[0] != ChunkKey(den, clr) {
				t.Fatalf("task group not unparked: %v", ua.got)
			}
			r.restart()
			if d := r.next(); d.Kind != DecRun {
				t.Fatalf("retried chunk does not run, also after a restart: %+v", d)
			}
		})
	}
}

// 3.6: a chunk the board ran and saw done stays done when the proxy goes away or the process restarts.
func TestSMLatchOutlivesTheProxy(t *testing.T) {
	const cain = "search_for_cain"
	for _, class := range smClasses {
		t.Run(class, func(t *testing.T) {
			for _, sm := range []bool{true, false} {
				r := newRigSM(t, class, 17, sm)
				if err := r.x.Accept(cain, "scroll"); err != nil {
					t.Fatal(err)
				}
				r.f.ChunkDone[ChunkKey(cain, "scroll")] = TierProxy
				if got := r.chunk(cain, "scroll"); got.State != Done {
					t.Fatalf("sm=%v: not done: %+v", sm, got)
				}
				delete(r.f.ChunkDone, ChunkKey(cain, "scroll")) // the key left the pack
				r.restart()
				got := r.chunk(cain, "scroll")
				if sm && got.State != Done {
					t.Fatalf("machine on: the latch did not hold: %+v", got)
				}
				if !sm && got.State == Done && !testSM {
					t.Fatalf("machine off: the proxy cannot still read done: %+v", got)
				}
			}
		})
	}
}

// R1 to R4 through the executor: only exact denial clears a latch, and it parks the dependants' head.
func TestSMRegressionParksAndAsks(t *testing.T) {
	const cain = "search_for_cain"
	r := newRigSM(t, "Paladin", 17, true)
	if err := r.x.Accept(cain, "scroll"); err != nil {
		t.Fatal(err)
	}
	r.f.ChunkDone[ChunkKey(cain, "scroll")] = TierMemory
	r.chunk(cain, "scroll") // latch written
	if err := r.x.Accept(cain, "read"); err != nil {
		t.Fatal(err)
	}
	delete(r.f.ChunkDone, ChunkKey(cain, "scroll"))
	if got := r.chunk(cain, "scroll"); got.State != Done {
		t.Fatalf("silence cleared the latch: %+v", got)
	}
	r.f.Denied = map[string]bool{ChunkKey(cain, "scroll"): true}
	d := r.next()
	if d.Kind != DecBlocked || !strings.Contains(d.Why, "no longer reads as done") {
		t.Fatalf("regression must park and say why: %+v", d)
	}
	if err := r.x.st.UpdateSM(nil, func(m *SM) error {
		if _, _, err := m.Retry(r.b, "normal", cain, "read", true, smT0); err != nil {
			return err
		}
		if m.find(cain, "scroll", "normal") == nil {
			t.Fatal("Run it again did not accept the regressed part")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSMAcceptChainBusyAndStop(t *testing.T) {
	const cain = "search_for_cain"
	for _, class := range smClasses {
		t.Run(class, func(t *testing.T) {
			r := newRigSM(t, class, 17, true)
			if err := r.x.Accept(cain, "stones"); err != nil {
				t.Fatal(err)
			}
			if err := r.x.Accept(cain, "stones"); err != nil {
				t.Fatalf("the same tap again must be a no-op: %v", err)
			}
			err := r.x.Accept(den, clr)
			if err == nil || !strings.Contains(err.Error(), "already accepted") || !strings.Contains(err.Error(), cain+"/scroll") {
				t.Fatalf("D5 refusal must name the running chunk: %v", err)
			}
			if err := r.x.Cancel(cain, "scroll"); err != nil {
				t.Fatal(err)
			}
			if st, _ := r.x.State(); st.Active != nil {
				t.Fatalf("D2: the dependants must stop with it: %+v", st.Active)
			}
			if err := r.x.Cancel(cain, "scroll"); err != nil {
				t.Fatalf("second stop: %v", err)
			}
		})
	}
}

// Q9, Q10: a refusal a town visit fixes is a hold, not a failure; it parks with hold_timeout when it outlives its
// budget, and the park survives a restart.
func TestSMHoldIsNoFailureAndTimesOut(t *testing.T) {
	for _, class := range smClasses {
		t.Run(class, func(t *testing.T) {
			r := newRigSM(t, class, 6, true)
			now := smT0
			r.x.o.Now = func() time.Time { return now }
			if err := r.x.Accept(den, clr); err != nil {
				t.Fatal(err)
			}
			d := r.next()
			held := StepResult{StepHeld, "potions 3 < 4", "needs_potions"}
			for i := 0; i < 6; i++ { // three failures would park an old-style refusal
				r.x.Report(d, held)
			}
			st, _ := r.x.State()
			if st.Active == nil || st.Active.Failures != 0 || st.Active.Hold != "needs_potions" || r.next().Kind != DecRun {
				t.Fatalf("a hold must count nothing and keep the chunk running: %+v", st.Active)
			}
			r.x.Report(d, StepResult{Outcome: StepRan})
			if st, _ = r.x.State(); st.Active.Hold != "" {
				t.Fatalf("a slice that ran must lift the hold: %+v", st.Active)
			}
			r.x.Report(d, held)
			now = now.Add(21 * time.Minute) // the chore budget is 20 minutes
			r.x.Report(d, held)
			if dd := r.next(); dd.Kind != DecBlocked || !strings.Contains(dd.Why, "Still waiting after 20 minutes") {
				t.Fatalf("budget exceeded must park with hold_timeout: %+v", dd)
			}
			r.restart()
			if dd := r.next(); dd.Kind != DecBlocked {
				t.Fatalf("hold_timeout park lost over a restart: %+v", dd)
			}
		})
	}
	// a code outside the closed list is a producer bug and stays the failure it used to be
	r := newRigSM(t, "Paladin", 6, true)
	r.x.Accept(den, clr)
	d := r.next()
	for i := 0; i < 3; i++ {
		r.x.Report(d, StepResult{StepHeld, "x", "made_up_code"})
	}
	if r.next().Kind != DecRun {
		t.Fatal("unknown hold code: see the failure counters")
	}
}
