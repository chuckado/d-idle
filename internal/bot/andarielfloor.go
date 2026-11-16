package bot

import (
	"strings"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/plandata"
	"d2lod-bot/internal/task"
)

// fillAndariel sets the Andariel level inputs of w: the class row of plandata.BossLevels (recommended and floor, with
// and without a merc), each overridden by its knob when that is set (boss.andarielMinLevel*, 0 = the class row). A class
// without a row is logged once and takes the old single-number defaults, never another class's row.
func (r *Runner) fillAndariel(w *task.World, class d2.Class) {
	b := r.e.Tune.Boss
	w.AndarielMinLevel, w.AndarielMinLevelMerc = b.AndarielMinLevel, b.AndarielMinLevelMerc
	row, err := plandata.BossLevels(plandata.GateAndariel, strings.ToLower(class.String()))
	if err != nil {
		if s := err.Error(); s != r.loop.task.andyRowErr {
			r.loop.task.andyRowErr = s
			r.e.Log.Warn("andariel level row missing, using the default floors", "err", s)
		}
		if w.AndarielMinLevel <= 0 {
			w.AndarielMinLevel = env.DefaultAndarielMinLevel
		}
		if w.AndarielMinLevelMerc <= 0 {
			w.AndarielMinLevelMerc = env.DefaultAndarielMinLevelMerc
		}
		return
	}
	if w.AndarielMinLevel <= 0 {
		w.AndarielMinLevel = row.Floor
	}
	if w.AndarielMinLevelMerc <= 0 {
		w.AndarielMinLevelMerc = row.FloorMerc
	}
	w.AndarielRec, w.AndarielRecMerc = row.Recommended, row.RecommendedMerc
	r.zoneBossFloor(w, b)
}

// bossKeyAndariel is the danger table key of Andariel (internal/danger/bosses.json).
const bossKeyAndariel = "andariel"

// zoneBossFloor lets the survival margin replace the Andariel class floor (bot.zoneModel 1, docs/survival/zonemodel.md):
// only when margin data for her exists, and only to LOWER the wall: a "safe" margin makes the level floor met; "risky" or
// "too hard" leaves the legacy floor and turns on the recommended-level warning (warn, never wall). An explicit knob
// floor (boss.andarielMinLevel*) is the user's and stays.
func (r *Runner) zoneBossFloor(w *task.World, b env.BossTuning) {
	if !r.zoneModelOn() || r.d.BossMargin == nil || b.AndarielMinLevel > 0 || b.AndarielMinLevelMerc > 0 {
		return
	}
	s := r.latest()
	if s == nil || s.Player == nil {
		return
	}
	word, ok := r.d.BossMargin(s, bossKeyAndariel, difficultyName(s.Player.Difficulty), r.zoneDial())
	if !ok {
		return
	}
	switch word {
	case "safe":
		w.AndarielMinLevel = min(w.AndarielMinLevel, s.Player.Level)
		w.AndarielMinLevelMerc = min(w.AndarielMinLevelMerc, s.Player.Level)
	default:
		w.AndarielRec = max(w.AndarielRec, s.Player.Level+1)
		w.AndarielRecMerc = max(w.AndarielRecMerc, s.Player.Level+1)
	}
}

// andarielLow is the lowest Andariel floor of the class (the merc one), for the antidote shopping window.
func (r *Runner) andarielLow(class d2.Class) int {
	var w task.World
	r.fillAndariel(&w, class)
	return min(w.AndarielMinLevel, w.AndarielMinLevelMerc)
}
