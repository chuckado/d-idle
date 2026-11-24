package plandata

import (
	"fmt"
	"sort"
)

// BossLevel is the per-class level row of one boss gate. Recommended is the level a careful player has (a warning
// below it, never a wall: the bot logs "going on below the recommended level" and continues, still bound by the
// danger, chicken and gear-gate rules); Floor is the lowest level the campaign starts the gate at. Both come in a
// no-merc and a living-merc value.
type BossLevel struct {
	Recommended, RecommendedMerc int
	Floor, FloorMerc             int
}

// Gate names of BossLevels.
const GateAndariel = "andariel"

// andarielLevels: per class, Normal. Recalibrated 2026-11-24 (user: a human kills her at clvl 9 to 10; docs/probes/andariel-gate-recalibration-2026-11-24.md): floor 11 alone and 10 with a merc, recommended 14 and 13, the same for every class (the old per-class rows came from the SME ballpark below). Recommended = the class readiness row of docs/sme/class-<class>.md (the range
// "clvl a-b": the merc value is a, the no-merc value b; S: the SME's table, EXP: a human's read, nothing measured on the
// bot). Paladin 15 / 17 is the user's starting data (ADR 0011 15 and 17; project review 2026-11-10 section 4.4: a human
// Paladin with the Rogue kills her at clvl 14 to 18). Floor = recommended minus 1 (the SME's "not overcautious" Andariel
// floor for the Paladin is 14 with the merc and 16 without; the other classes copy the one-level rule, U). The Andariel
// kit (8 potions, antidotes) and the unattended hold are separate gates (task.BossGate).
var andarielLevels = map[string]BossLevel{
	"paladin":     {14, 13, 11, 10},
	"sorceress":   {14, 13, 11, 10},
	"necromancer": {14, 13, 11, 10},
	"amazon":      {14, 13, 11, 10},
	"druid":       {14, 13, 11, 10},
	"assassin":    {14, 13, 11, 10},
	"barbarian":   {14, 13, 11, 10},
}

var bossLevels = map[string]map[string]BossLevel{GateAndariel: andarielLevels}

// BossLevels returns the level row of gate for a class name as d2.Class.String lowercases it ("paladin"). An unknown
// gate or class is an error that names it; there is no silent fallback to another class.
func BossLevels(gate, class string) (BossLevel, error) {
	g, ok := bossLevels[gate]
	if !ok {
		return BossLevel{}, fmt.Errorf("plandata: no boss level table for gate %q", gate)
	}
	r, ok := g[class]
	if !ok {
		names := make([]string, 0, len(g))
		for n := range g {
			names = append(names, n)
		}
		sort.Strings(names)
		return BossLevel{}, fmt.Errorf("plandata: no %s level row for class %q (have %v)", gate, class, names)
	}
	return r, nil
}

// LowestFloor is the lowest FloorMerc of gate over all classes: the class-less script check of the boss runner takes it
// and the campaign gate (internal/task) decides per class.
func LowestFloor(gate string) int {
	low := 0
	for _, r := range bossLevels[gate] {
		if low == 0 || r.FloorMerc < low {
			low = r.FloorMerc
		}
	}
	return low
}
