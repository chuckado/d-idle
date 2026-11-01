package d2dsrc

import "d2lod-bot/internal/d2"

// Stat ids the melee model reads (internal/combat/melee.go); gear sums are in whole points, not 8.8.
const (
	statToHit  = 19
	statMinDmg = 21
	statMaxDmg = 22
	statDmgPct = 25
)

// gearStats turns the lab2-stats additions of dump stats ("gear" sums per stat and param, "weapons" base damage per
// worn weapon, docs/forkdbg.md) into the player stat ids the melee model reads. It returns nothing for a dump
// without the "gear" key (an older fork), so the stat stays absent and the model keeps its defaults.
// 21 / 22 are the right-hand weapon's base range plus gear flats, and only when a weapon sits in the right hand
// (slot 4); the model applies stat 25 itself. The dump has no weapon speed, so none is modelled (U: no source).
func gearStats(st obj) map[int]int {
	if _, ok := st["gear"].([]any); !ok {
		return nil
	}
	sum := map[int]int{}
	for _, g := range st.list("gear") {
		id, ok := g.int("stat")
		if !ok {
			continue
		}
		if p, _ := g.int("param"); p != 0 { // only the unparameterised sums feed the melee stats
			continue
		}
		n, _ := g.int("sum")
		sum[id] += n
	}
	out := map[int]int{statToHit: sum[statToHit], statDmgPct: sum[statDmgPct]}
	for _, w := range st.list("weapons") {
		if slot, _ := w.int("slot"); slot != d2.BodyRHand {
			continue
		}
		lo, lok := w.int("mindam")
		hi, hok := w.int("maxdam")
		if lok && hok {
			out[statMinDmg], out[statMaxDmg] = lo+sum[statMinDmg], hi+sum[statMaxDmg]
		}
	}
	return out
}
