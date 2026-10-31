package d2denv

import (
	"context"

	"d2lod-bot/internal/d2dctl"
)

// NewLateClock is a wall clock until Detect or Install looks at d2d. Backend builds its own; a lab tool that talks to d2d
// without a Backend (cmd/d2dshop) uses this and Install.
func NewLateClock(ctx context.Context, c *d2dctl.Client) *LateClock {
	return &LateClock{c: c, ctx: ctx}
}

// Install detects the clock and puts its Pause and Now into d2dctl, so the pauses and the deadlines of the town verbs
// (hover, click, NPC approach) are game time under lockstep and wall time otherwise (docs/probes/waits-audit.md). It is
// process-wide state: one d2d per process. Returns whether lockstep is on. Backend.Start already sets Pause; the d2d
// programs call Install after Start so the deadline clock follows too.
func (l *LateClock) Install(ctx context.Context) bool {
	ls := l.Detect(ctx)
	d2dctl.Pause = l.Pause
	d2dctl.Now = l.Now
	return ls
}
