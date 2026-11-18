package d2denv

import (
	"context"
	"errors"
	"fmt"
	"time"

	"d2lod-bot/internal/d2dctl"
)

// The player's Pause on d2d (docs/pause.md). Mechanism: the fork's runtime switch "lockstep on" with no "step" behind it
// stops the game clock while the main loop keeps running, so the control socket still answers, the screenshot verb
// still draws the frozen frame and the process never looks dead; "lockstep off" resumes in the same process. Tags: V
// the verbs exist and switch at runtime (docs/d2d-fork.md section 3, used by the lab for menus); U that a long freeze
// resumes cleanly in a realtime tester d2d (live checklist in docs/pause.md). SIGSTOP of the process was rejected: the
// socket would go silent (the channel monitors would restart a healthy game) and the wall clock would jump on SIGCONT.

// freezeClient is what the freezer needs of d2dctl.Client (a fake in the tests).
type freezeClient interface {
	Do(ctx context.Context, line string) (*d2dctl.Reply, error)
	Clock(ctx context.Context) (d2dctl.ClockInfo, error)
}

// Freezer is bot.Freezer for d2d.
type Freezer struct {
	C   freezeClient
	Clk *LateClock // a bot that already steps a lockstep d2d (the lab) cannot freeze this way: CanFreeze is false
}

// NewFreezer returns the freezer of one backend.
func NewFreezer(c *d2dctl.Client, clk *LateClock) *Freezer { return &Freezer{C: c, Clk: clk} }

// CanFreeze: the "clock" verb answers, the game is free running, the bot does not drive it with steps, and the
// runtime switch exists (probed with the harmless "lockstep off").
func (f *Freezer) CanFreeze(ctx context.Context) bool {
	if f.Clk != nil && f.Clk.Lockstep() {
		return false
	}
	ci, err := f.C.Clock(ctx)
	if err != nil || ci.Lockstep {
		return false
	}
	// "lockstep off" on a free-running d2d changes nothing; an unknown verb (stock or gated build) means no button
	_, err = f.C.Do(ctx, "lockstep off")
	return err == nil
}

// Freeze switches the lockstep clock on (frozen) or off (running) and checks the answer of "clock".
func (f *Freezer) Freeze(ctx context.Context, on bool) error {
	word := "off"
	if on {
		word = "on"
	}
	if _, err := f.C.Do(ctx, "lockstep "+word); err != nil {
		return fmt.Errorf("lockstep %s: %w", word, err)
	}
	ci, err := f.C.Clock(ctx)
	if err != nil {
		return fmt.Errorf("clock after lockstep %s: %w", word, err)
	}
	if ci.Lockstep != on {
		return errors.New("the game did not switch (lockstep " + word + " had no effect)")
	}
	return nil
}

// SetFrozen stands the wall clock still (on) or lets it run again without the paused stretch (off). A lockstep clock
// needs no offset: it only moves when stepped.
func (l *LateClock) SetFrozen(on bool) {
	l.fmu.Lock()
	defer l.fmu.Unlock()
	switch {
	case on && l.frozenAt.IsZero():
		l.frozenAt = time.Now()
	case !on && !l.frozenAt.IsZero():
		l.offset += time.Since(l.frozenAt)
		l.frozenAt = time.Time{}
	}
}

// wall is time.Now without the paused time.
func (l *LateClock) wall() time.Time {
	l.fmu.Lock()
	defer l.fmu.Unlock()
	if !l.frozenAt.IsZero() {
		return l.frozenAt.Add(-l.offset)
	}
	return time.Now().Add(-l.offset)
}
