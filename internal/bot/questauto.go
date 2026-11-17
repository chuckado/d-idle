package bot

import "d2lod-bot/internal/questboard"

// The knobs of the opt-in auto-start (slice S8, docs/questboard.md "Auto-start and away"): the limits live in
// internal/env QuestsTuning, the rule in internal/questboard/auto.go. Nothing here runs with quests.board = 0.

func (r *Runner) questAutoOptions() questboard.AutoOptions {
	q := r.e.Tune.Quests
	return questboard.AutoOptions{DelaySec: q.AutoDelaySec, AwayMin: q.AwayMin, HoldMin: q.AutoHoldMin, GapMin: q.AutoGapMin, DayCap: q.AutoDayCap}
}

// noteQuestHold tells the board the player paused, resumed or reset: no auto-start for quests.autoHoldMin minutes.
func (r *Runner) noteQuestHold() {
	if rt := r.qbRT(); rt != nil {
		rt.exec.NoteHold()
	}
}

// QuestLevel is the loaded hero's level (0 before a snapshot), for "levels to go" on an offer.
func (r *Runner) QuestLevel() int {
	if s := r.latest(); s != nil && s.Player != nil {
		return s.Player.Level
	}
	return 0
}
