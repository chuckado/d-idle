package bot

// Points gate (docs/points.md): who spends stat and skill points. By default the player does (resource.pointsDecisions
// = 0): the bot banks them, the Points tile and the Needs-you row show them, and nothing is spent until the player acts or
// turns the hero's autopilot switch on. The gate is a function so the switch is read at the moment the loop would spend.
// Without a gate (library and tests) the bot spends as it always did.

// SetPointsGate installs the check; it returns true when the bot may spend by itself. nil removes it.
func (r *Runner) SetPointsGate(f func() bool) {
	if f == nil {
		r.pointsGate.Store(nil)
		return
	}
	r.pointsGate.Store(&f)
}

func (r *Runner) botMaySpendPoints() bool {
	f := r.pointsGate.Load()
	return f == nil || (*f)()
}

// FightOn reports whether hostiles are near the hero right now (the clear radius the loop uses): the Points desk spends
// only at a calm moment, like the leveler. The newest snapshot is used; unknown means calm.
func (r *Runner) FightOn() bool {
	s := r.LastSnapshot()
	if s == nil || s.Player == nil || s.Player.InTown() {
		return false
	}
	return r.hostileNear(s, r.e.Tune.Bot.ClearRadius) > 0
}
