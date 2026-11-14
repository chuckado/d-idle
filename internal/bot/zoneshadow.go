package bot

import "d2lod-bot/internal/d2"

// zoneShadow logs the survival margin of the chosen zone next to the `planner` line (slice S-B, shadow only: nothing
// reads the result).
func (r *Runner) zoneShadow(s *d2.Snapshot, area int, diff string) {
	if r.d.ZoneShadow == nil || s == nil || s.Player == nil {
		return
	}
	potions := -1
	if healing, _, known := r.beltAndInv(s); known {
		potions = healing
	}
	if line := r.d.ZoneShadow(s, area, diff, potions); line != "" {
		r.e.Log.Info(line)
	}
}
