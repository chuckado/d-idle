package drill

import (
	"context"
	"time"
)

// D22Lab is the instance as the DRILLS22 fixtelekey live proof uses it: every drill is one method call that returns
// the finished Outcome (the work needs the live verbs, so there is no offline fake).
type D22Lab interface {
	FieldLab
	D22(ctx context.Context, r *Run, id string) Outcome
}

// D22 returns the DRILLS22 group (letter T): the Teleport capability of card fixtelekey on a hidden d2d slot. Order
// matters: the control (no Teleport) runs before the drill that learns it.
func D22(l D22Lab) []Drill {
	lim := 6 * time.Minute
	mk := func(id, title string) Drill {
		return Drill{ID: id, Primitive: "P26", Title: title, Limit: lim, Do: func(ctx context.Context, r *Run) Outcome { return l.D22(ctx, r, id) }}
	}
	return []Drill{
		mk("D22-C", "control: Act 2 leg at clvl 18 without Teleport learned: walks, no hops, no error"),
		mk("D22-PC", "pocket start (Rocky Waste 165,220, a walk trap) without Teleport: the walking control for D22-P"),
		mk("D22-A", "Act 2 leg at clvl 18 with Teleport learned: hops, refused, mana, ticks against the control"),
		mk("D22-P", "pocket start with Teleport learned: what Teleport is for (walls, pockets), ticks against D22-PC"),
		mk("D22-B", "the Teleport bind: Fighter.TeleportKey, the player's skill list and the hop through it"),
		mk("D22-T", "town: the wired navigator never hops; Duriel's Lair and the Summoner level by rule"),
		mk("D22-R", "refused hops bounded: the unwired navigator in a town (the game refuses every cast)"),
	}
}
