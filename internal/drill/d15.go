package drill

import (
	"context"
	"time"
)

// D15Lab is the instance as the DRILLS15 skill-smoke drills use it: every drill is one method call that returns the
// finished Outcome (the work needs the live verbs, so there is no offline fake; the verdict rules are in the lab).
type D15Lab interface {
	FieldLab
	D15(ctx context.Context, r *Run, id string) Outcome
}

// D15 returns the DRILLS15 group (letter R): first live skill-smoke drills of the traces sk-teleport (T), sk-mana (M),
// sk-missile (K) and sk-armor (E). PASS means d2d matches the trace's game.exe expectation (a read of game.exe, V);
// FAIL means d2d differs from it (a deviation, not a harness error); RECORD has no expectation to compare.
func D15(l D15Lab) []Drill {
	lim := 10 * time.Minute
	mk := func(id, prim, title string) Drill {
		return Drill{ID: id, Primitive: prim, Title: title, Limit: lim, Do: func(ctx context.Context, r *Run) Outcome { return l.D15(ctx, r, id) }}
	}
	return []Drill{
		mk("D15-T1", "P26", "T1 Teleport in town: refused, no mana taken (sk-teleport-trace D, section 8)"),
		mk("D15-T2", "P26", "T2 Teleport hops at 3..24 cells on open ground: which land, what mana"),
		mk("D15-T3", "P26", "T3 Teleport at a blocked target: game.exe snaps to the nearest free spot; d2d?"),
		mk("D15-T4", "P26", "T4 Teleport at a monster (point and unit): game.exe lands next to it"),
		mk("D15-T5", "P26", "T5 does the merc move with a Teleport hop on d2d"),
		mk("D15-T7", "P26", "T7 mana paid on a hop game.exe refuses (24 points of mana, far target)"),
		mk("D15-M2", "P26", "M2 a real monster blow during a Fire Bolt cast: missile and mana"),
		mk("D15-M3", "P26", "M3 Teleport level 25: the minmana floor (castable at half a point and at 0 mana)"),
		mk("D15-M4", "P26", "M4 mana regeneration over 30 s, without and with Warmth"),
		mk("D15-M6", "P26", "M6 Blizzard cast twice: the skill delay (45 frames after the action frame)"),
		mk("D15-K1", "P26", "K1 missile reach and flight time: Fire Bolt 46.9, Ice Bolt 28.1, Lightning 35.2 subtiles"),
		mk("D15-K2", "P26", "K2 Charged Bolt count at levels 1, 5, 20: min(24, 3 + level - 1)"),
		mk("D15-E1", "P26", "E1 Energy Shield at 0 mana: absorb and state against game.exe"),
	}
}
