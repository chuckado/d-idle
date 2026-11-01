package main

import (
	"context"
	"errors"

	"d2lod-bot/internal/data"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/task"
	"d2lod-bot/internal/town"
)

// Act 2 to 5 pieces of the task adapter (docs/act2.md, docs/act3.md). Classes are V (monstats hcIdx, quest/act2.go and
// act3.go); every dialog row below is U until a live run opened the menu (docs/act2.md C2, docs/act3.md A3-4). A wrong
// row shows up as a failed step (needs_attention), never as a guessed click.

// mercRows is the hire-window row per act (1-based) for the wanted aura, filled after the live check that reads the
// window (docs/act2.md M1, docs/act3.md A3-4). Empty: the hire stops with a prereq fault and the step is routed around,
// because the aura of a row is only visible on screen and a guessed row costs gold.
var mercRows = map[int]int{}

// mercHirer is bot.TaskDeps.HireMerc over town.Hire.
func mercHirer(tw *town.Town, e *env.Env) func(ctx context.Context, act int, aura string) error {
	return func(ctx context.Context, act int, aura string) error {
		// Soft rule (plan-v2 10.1): never pay Kashya. The Rogue comes free after Blood Raven through the claim step
		// (task.MercClaimer); an Act 1 hire here would spend the gold the kit needs. No task step asks for it today.
		if act == 1 {
			return task.Errorf(task.KindPrereq, "merc: no paid Kashya hire; the Rogue is claimed free after Blood Raven")
		}
		s, err := e.Snapshot()
		if err != nil {
			return err
		}
		w := town.MercWant{Act: act, Diff: data.Difficulty(s.Player.Difficulty), Aura: aura, Row: -1}
		if r, ok := mercRows[act]; ok {
			w.Row = r
		}
		rep, err := tw.Hire(ctx, w)
		switch {
		case errors.Is(err, town.ErrMercRowUnknown), errors.Is(err, town.ErrMercUnverified), errors.Is(err, town.ErrNoCandidate), errors.Is(err, town.ErrNoMercNPC):
			return task.Wrap(task.KindPrereq, err)
		case err != nil:
			return err
		case !rep.Hired:
			return task.Errorf(task.KindResource, "mercenary not hired: %s", rep.Skipped)
		}
		return nil
	}
}
