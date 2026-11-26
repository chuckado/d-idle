package bot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
)

// reflexFighter is a fake Fighter that can send a belt drink from the guard.
type reflexFighter struct {
	*fakeFighter
	handled bool
	err     error
	sent    atomic.Int32
}

func (f *reflexFighter) ReflexDrink(ctx context.Context, a SafetyAction) (bool, error) {
	f.sent.Add(1)
	f.w.call("reflex %s", a)
	return f.handled, f.err
}

func reflexRig(t *testing.T, knob int, rf *reflexFighter) *rig {
	g := newRig(t, func(c *Config, d *Deps, tu *env.Tuning) {
		tu.Bot.ReactionMs, tu.Bot.ReflexDrink = 0, knob
		rf.fakeFighter = &fakeFighter{}
		d.Fighter = rf
	})
	rf.fakeFighter.w = g.w
	return g
}

// bot.reflexDrink: the guard sends the belt drink itself and the running action is not cancelled; with the knob off, or when
// only the action can drink (pack), it cancels as before. Two classes: the rule is about the drink, not the build.
func TestReflexDrinkDoesNotCancelTheAction(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Barbarian} {
		for _, c := range []struct {
			name    string
			knob    int
			handled bool
			cancels bool
		}{
			{"on, belt", 1, true, false},
			{"on, needs the pack", 1, false, true},
			{"off", 0, true, true},
		} {
			t.Run(class.String()+"/"+c.name, func(t *testing.T) {
				rf := &reflexFighter{handled: c.handled}
				if !c.handled {
					rf.err = errors.New("no belt column")
				}
				g := reflexRig(t, c.knob, rf)
				g.w.set(func(w *world) { w.state = d2.StateInGame; w.area = 2; w.life, w.maxL = 30, 100; w.class = class })
				rf.safety = func(*d2.Snapshot) SafetyAction { return SafetyDrinkLife }
				ctx, cancel := context.WithCancelCause(context.Background())
				defer cancel(nil)
				fired := g.r.guardOnce(ctx, cancel, StateClearing)
				var si safetyInterrupt
				cancelled := errors.As(context.Cause(ctx), &si)
				if fired != c.cancels || cancelled != c.cancels {
					t.Fatalf("fired %v cancelled %v, want %v", fired, cancelled, c.cancels)
				}
				if got := int(rf.sent.Load()); (c.knob == 1) != (got == 1) {
					t.Fatalf("reflex drinks sent %d with the knob at %d", got, c.knob)
				}
				if c.knob == 1 && c.handled {
					g.r.obMu.Lock()
					l := g.r.ob.sl
					g.r.obMu.Unlock()
					if l.reflex != 1 || l.n != 1 || l.lastMs > 100 {
						t.Fatalf("reflex stats %+v, want one reflex drink within 100 ms", l)
					}
				}
			})
		}
	}
}

// A chicken or portal cancel gives the action bot.reflexDeadlineMs; an action that ignores the cancel is then bypassed:
// the guard carries the safety action out itself, once, and the loop does not repeat it.
func TestHardCancelIsCarriedOutByTheGuardAfterTheDeadline(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		for _, deadline := range []int{200, 0} {
			t.Run(fmt.Sprintf("%v/deadline %d", class, deadline), func(t *testing.T) {
				g := newRig(t, func(c *Config, d *Deps, tu *env.Tuning) { tu.Bot.ReflexDeadlineMs, tu.Bot.ReactionMs = deadline, 0 })
				g.w.set(func(w *world) { w.state = d2.StateInGame; w.area = 2; w.life, w.maxL = 10, 100; w.class = class })
				g.fight.safety = func(*d2.Snapshot) SafetyAction { return SafetyChicken }
				release := make(chan struct{})
				stuck := func(ctx context.Context) error { // an action that never looks at its context
					<-release
					return ctx.Err()
				}
				res := make(chan bool, 1)
				go func() {
					res <- g.r.do(context.Background(), g.r.control(), StateClearing, "clearing", 0, "clear", stuck)
				}()
				acted := func() bool { return strings.Contains(g.w.log(), "act chicken") }
				wait := 5 * time.Second
				if deadline == 0 {
					wait = 150 * time.Millisecond // off: nothing to wait for; the loop carries it out after the action returns
				}
				deadlineAt := time.Now().Add(wait)
				for !acted() && time.Now().Before(deadlineAt) {
					time.Sleep(time.Millisecond)
				}
				if got := acted(); got != (deadline > 0) {
					t.Fatalf("guard carried the chicken out while the action was stuck = %v, deadline %d ms", got, deadline)
				}
				close(release)
				<-res
				if n := strings.Count(g.w.log(), "act chicken"); n != 1 {
					t.Fatalf("chicken carried out %d times, want exactly once", n)
				}
			})
		}
	}
}
