package main

import (
	"context"
	"errors"
	"os"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/invdesk"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/town"
)

// wireItemVerbs puts the verb-first executors in front of the click paths on the d2d backend
// (docs/design/item-verbs.md). Every executor asks the knob invdesk.verbs (default 0) at call time and the backend's
// caps, and falls back to the click path it wraps, so with the knob off nothing changes. Stash verbs are only asked by
// the town while the stash is open at the chest. Not used on Game.exe (no client).
//
// One Channel serialises every verb call: the bot's executors (equip, mover, trasher) and the desk's fast lane (instant
// actions, docs/inventory-actions.md) reach the client only through the same Serial, so there is one owner of the
// control channel at a time and no lock is held across desks.
func wireItemVerbs(ref *runnerRef, tw *town.Town, lv *leveler.Leveler, c *d2dctl.Client, e *env.Env) {
	if ref == nil || ref.inv == nil || c == nil || e == nil {
		return
	}
	on := func() bool { return e.Tune != nil && e.Tune.Invdesk.Verbs != 0 }
	verbs := invdesk.Serial{Ch: invdesk.NewChannel(), V: coder{d2dctl.ItemVerbs{C: c}}}
	mover := invdesk.NewVerbMover(verbs, on)
	mover.Fresh = e.Snapshot
	if tw != nil {
		tw.SetMover(mover)
	}
	ref.inv.SetHowSource(mover)
	if lv != nil {
		ve := invdesk.NewVerbEquipper(lv, verbs, on)
		ve.Verify = mover.Equipped
		ref.inv.SetEquipper(ve)
		ref.inv.SetTrasher(invdesk.VerbTrasher{Click: lv, Mover: mover})
	}
	ref.inv.SetFast(&invdesk.FastLane{
		Verbs: verbs,
		On:    on,
		// A named run (gate, soak, drill, D2BOT_RUN) keeps the queue path: whether a verb in a measured run satisfies the
		// crutch rule is the user's decision (item-verbs section 10), so the lane stays off there. Read once at wiring.
		Named: namedRun(os.Getenv("D2BOT_RUN") != ""),
		Hold:  inputHeld,
		Fresh: e.Snapshot,
	})
}

func namedRun(v bool) func() bool { return func() bool { return v } }

// inputHeld says why the lane must not send a verb now. A cursor item means a click path (or the player) holds the hand.
// U: a safety action in progress (chicken, flee, portal) has no signal outside internal/bot; a verb never touches the
// cursor or moves the hero, so the lane tolerates it, and the queue path keeps the ladder (docs/inventory-actions.md).
func inputHeld(s *d2.Snapshot) string {
	if s != nil && s.Inventory != nil && s.Inventory.Cursor != nil {
		return "the cursor holds an item"
	}
	return ""
}

// coder adds unequip and turns the client's refusals into invdesk.VerbRefusal (the code stays in the request).
type coder struct{ d2dctl.ItemVerbs }

func refusal(err error) error {
	var r *d2dctl.ItemRefused
	if errors.As(err, &r) {
		return &invdesk.VerbRefusal{Code: r.Code, Why: r.Why}
	}
	return err
}

func (a coder) ItemEquip(ctx context.Context, id uint32, slot int) (bool, error) {
	n, err := a.ItemVerbs.ItemEquip(ctx, id, slot)
	return n, refusal(err)
}

func (a coder) ItemUnequip(ctx context.Context, id uint32) (bool, error) {
	r, err := a.C.ItemUnequip(ctx, id)
	return r.Noop, refusal(err)
}

func (a coder) ItemMove(ctx context.Context, id uint32, panel string, col, row int) (bool, error) {
	n, err := a.ItemVerbs.ItemMove(ctx, id, panel, col, row)
	return n, refusal(err)
}

func (a coder) ItemIdentify(ctx context.Context, id uint32) (bool, error) {
	n, err := a.ItemVerbs.ItemIdentify(ctx, id)
	return n, refusal(err)
}

func (a coder) ItemDrop(ctx context.Context, id uint32) error {
	return refusal(a.ItemVerbs.ItemDrop(ctx, id))
}

func (a coder) GoldStash(ctx context.Context, amount int, deposit bool) (int, int, error) {
	b, c, err := a.ItemVerbs.GoldStash(ctx, amount, deposit)
	return b, c, refusal(err)
}
