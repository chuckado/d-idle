package task

import (
	"context"
	"testing"
)

// worlds of the two class families the guards serve: the caster and the melee hero differ in Caster only; the guard
// code must not depend on it (build-flexibility rule).
func holdWorlds() map[string]World {
	base := func(caster bool) World {
		return World{Level: 12, Gold: 3000, Potions: 0, Gated: true, GatedWhy: "no potions", GatedFix: "potions", SafeArea: 2, Caster: caster, BoardSM: true}
	}
	return map[string]World{"Paladin": base(false), "Sorceress": base(true)}
}

// bug 4: with the board machine a missing potion kit that the purse pays for sends the hero to town, nothing is refused.
func TestFieldReadyPotionsAreAChore(t *testing.T) {
	for name, w := range holdWorlds() {
		t.Run(name, func(t *testing.T) {
			f := fieldReady(4)(w)
			if f == nil || f.Kind != KindResource || f.Code != "needs_potions" || len(f.Insert) != 1 || f.Hold {
				t.Fatalf("potion kit must insert the chores step: %+v", f)
			}
			// the contrast: the old behaviour (knob off) is a plain danger refusal
			w.BoardSM = false
			if f := fieldReady(4)(w); f == nil || f.Kind != KindDanger || f.Code != "" || f.Hold || len(f.Insert) != 0 {
				t.Fatalf("knob off must keep the old fault: %+v", f)
			}
		})
	}
}

func TestFieldReadyHoldCodes(t *testing.T) {
	for name, base := range holdWorlds() {
		t.Run(name, func(t *testing.T) {
			cases := []struct {
				mod  func(*World)
				code string
			}{
				{func(w *World) { w.GatedFix = ""; w.Gold = 0 }, "needs_gold_farm"},
				{func(w *World) { w.GatedFix = ""; w.Gold = 3000; w.Potions = 2 }, "too_weak"},
				{func(w *World) { w.Gated = false; w.Banned = map[Area]bool{4: true} }, "area_banned"},
				{func(w *World) { w.Gated = false; w.RouteBlocked = "clvl 12 < 20" }, "route_closed"},
			}
			for _, c := range cases {
				w := base
				c.mod(&w)
				if c.code == "route_closed" && !RouteGuarded(4) {
					continue // area 4 is not a guarded route area in this data: the other cases cover the table
				}
				f := fieldReady(4)(w)
				if f == nil || f.Code != c.code || !f.Hold || f.Kind != KindDanger {
					t.Fatalf("%s: %+v", c.code, f)
				}
			}
		})
	}
}

type holdEnv struct{ w World }

func (h holdEnv) World(context.Context) (World, error) { return h.w, nil }

type holdPlanner struct{ p Plan }

func (h holdPlanner) Plan(World, Session) Plan { return h.p }

// A hold is no failure: however many ticks the guard refuses, the executor never counts it, escapes or blocks the group.
func TestExecutorNeverEscapesAHold(t *testing.T) {
	for _, hold := range []bool{true, false} {
		w := World{Level: 12, Gold: 0, Gated: true, GatedWhy: "no potions and no gold", SafeArea: 2, BoardSM: hold}
		step := Step{ID: "quest.x/go", Title: "Go", Ready: fieldReady(4), Need: func(World) (bool, Evidence) { return true, EvMemory }, Do: func(context.Context, Actions) error { return nil }}
		ex := NewExecutor(holdEnv{w}, nil)
		pl := holdPlanner{Plan{ID: "quest.x", Group: "x", Nodes: []Node{{Step: step}}}}
		var last Result
		for i := 0; i < 12; i++ {
			last, _ = ex.Tick(context.Background(), pl)
			if hold && (last.Outcome != OutRetry || !last.Hold || last.Code != "needs_gold_farm") {
				t.Fatalf("tick %d: %+v", i, last)
			}
		}
		_, blocked := ex.Blocked("x")
		if hold && blocked {
			t.Fatal("a hold blocked the group")
		}
		if !hold && !blocked {
			t.Fatalf("knob off: the old escape after repeated refusals is gone: %+v", last) // the contrast
		}
	}
}

// The trip a coded fault inserts reports the fault's code on the result of the step that ran, so the board can show the
// hold while the hero is in town (and no failure is counted when the trip is not enough).
func TestViaCarriesTheInsertingFaultCode(t *testing.T) {
	for name, w := range holdWorlds() {
		t.Run(name, func(t *testing.T) {
			trips := 0
			chores := Step{ID: "guard/chores", Title: "Chores", Need: func(World) (bool, Evidence) { return true, EvMemory },
				Do: func(context.Context, Actions) error { trips++; return nil }}
			step := Step{ID: "quest.x/go", Title: "Go", Need: func(World) (bool, Evidence) { return true, EvMemory },
				Ready: func(World) *Fault {
					if trips == 0 {
						return &Fault{Kind: KindResource, Code: "needs_potions", Why: "3 healing potions < 4", Insert: []Step{chores}}
					}
					return nil
				},
				Do: func(context.Context, Actions) error { return nil }}
			ex := NewExecutor(holdEnv{w}, nil)
			pl := holdPlanner{Plan{ID: "quest.x", Group: "x", Nodes: []Node{{Step: step}}}}
			res, _ := ex.Tick(context.Background(), pl)
			if res.StepID != "guard/chores" || res.Via != "needs_potions" || res.ViaWhy == "" || trips != 1 {
				t.Fatalf("the chores trip must carry the hold: %+v trips %d", res, trips)
			}
			if res, _ = ex.Tick(context.Background(), pl); res.Via != "" {
				t.Fatalf("a step nothing inserted carries no hold: %+v", res)
			}
		})
	}
}
