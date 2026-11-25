package main

import (
	"context"

	"d2lod-bot/internal/d2dctl"
)

// d2dCorpse is the bot's corpse click on the d2d backend: "debug corpses" lists the player's corpses of the game and
// "cmd interact -3000-i" takes number i (V: docs/d2d-town.md). A corpse the hero left in the field is moved to town by the
// next game (V d2d, docs/d2d-spike.md), where the bot takes it first thing.
type d2dCorpse struct{ c *d2dctl.Client }

// TakeCorpse clicks the first corpse that carries items and lies in area (0: any level); found is false when none is listed.
func (a d2dCorpse) TakeCorpse(ctx context.Context, area int) (bool, error) {
	l, err := a.c.Corpses(ctx)
	if err != nil {
		return false, err
	}
	for i, c := range l.Items {
		if c.Items > 0 && (area <= 0 || c.Level == area) {
			return true, a.c.CorpseInteract(ctx, i)
		}
	}
	return false, nil
}
