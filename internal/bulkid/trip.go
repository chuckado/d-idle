package bulkid

import (
	"context"
	"errors"
	"fmt"

	"d2lod-bot/internal/env"
)

// Lister reads the hero's unidentified items in the pack and the stash (fresh each call).
type Lister interface {
	Unidentified(ctx context.Context) ([]Item, error)
}

// CainTalk is the game side of Cain: seen in this town, and the Identify line (town.CainIdentify).
type CainTalk interface {
	Present(ctx context.Context) bool
	IdentifyAll(ctx context.Context) (int, error)
}

// Effective returns the tuning a run must use: measured runs (gates, soaks) never identify in bulk.
func Effective(tu env.IdentifyTuning, measured bool) env.IdentifyTuning {
	if measured {
		tu.Policy = env.IdentifyManual
	}
	return tu
}

// Report is what one Cain visit did.
type Report struct {
	Reason      string // why nothing (or less) happened
	Planned     int
	Staged      int
	Identified  int    // planned items that read identified afterwards
	Returned    int    // moved back to the stash
	Kept        int    // stayed in the pack because the stash had no room
	NeedBuy     int    // scrolls the plan would buy first (the caller buys only if allowed)
	ScrollPicks []Pick // picks the policy wants done with scrolls (the town's identify chore does them)
}

// Trip runs the Cain visit of the design (section 4). It is called from a town visit that is already planned.
type Trip struct {
	F        *Filter
	Tu       env.IdentifyTuning
	Hero     func(ctx context.Context) Hero
	L        Lister
	Cain     CainTalk
	S        *Stager
	Measured bool
}

// Run plans, stages the stash picks, talks to Cain, verifies and returns the stash items. The stash items are returned
// whatever happens to Cain's dialog.
func (t *Trip) Run(ctx context.Context) (Report, error) {
	var rep Report
	tu := Effective(t.Tu, t.Measured)
	// A restart (or a failed return) may have left stash items in the pack: put them back whatever the policy says.
	if len(t.S.J.IDs) > 0 {
		rr, rerr := t.S.Return(ctx)
		rep.Returned, rep.Kept = len(rr.Back), len(rr.Kept)
		rep.Reason = rr.Reason
		if rerr != nil {
			return rep, rerr
		}
	}
	if tu.Policy == env.IdentifyManual || tu.Policy == env.IdentifyNever {
		rep.Reason = WhyManual
		if tu.Policy == env.IdentifyNever {
			rep.Reason = WhyNever
		}
		return rep, nil
	}
	present := t.Cain.Present(ctx)
	if tu.Policy == env.IdentifyCainFree && !present {
		rep.Reason = WhyNoCain // nothing is listed, opened or moved for an absent Cain
		return rep, nil
	}
	items, err := t.L.Unidentified(ctx)
	if err != nil {
		return rep, fmt.Errorf("bulkid: list: %w", err)
	}
	h := t.Hero(ctx)
	h.CainOpen = h.CainOpen && present // nothing moves for an absent Cain
	plan := Make(t.F, tu, h, items)
	rep.Reason = plan.Reason
	rep.NeedBuy = plan.NeedBuy
	for _, p := range plan.Picks {
		if p.Via == ViaScroll {
			rep.ScrollPicks = append(rep.ScrollPicks, p)
		}
	}
	var cain []Pick
	for _, p := range plan.Picks {
		if p.Via == ViaCain {
			cain = append(cain, p)
		}
	}
	rep.Planned = len(cain)
	if len(cain) == 0 {
		return rep, nil
	}
	var stashIDs []uint32
	for _, p := range cain {
		if p.Item.Place == PlaceStash {
			stashIDs = append(stashIDs, p.Item.ID)
		}
	}
	var runErr error
	if len(stashIDs) > 0 {
		st, err := t.S.StageIn(ctx, stashIDs)
		rep.Staged = len(st.Moved)
		if err != nil {
			runErr = err
		}
		if st.Reason != "" {
			rep.Reason = st.Reason
		}
	}
	if runErr == nil {
		if _, err := t.Cain.IdentifyAll(ctx); err != nil {
			runErr = err
			rep.Reason = err.Error()
		}
	}
	// Whatever happened, the stash items go home before the visit ends.
	rr, rerr := t.S.Return(ctx)
	rep.Returned, rep.Kept = len(rr.Back), len(rr.Kept)
	if rr.Reason != "" {
		rep.Reason = rr.Reason
	}
	left, lerr := t.L.Unidentified(ctx)
	if lerr == nil {
		still := map[uint32]bool{}
		for _, it := range left {
			still[it.ID] = true
		}
		for _, p := range cain {
			if !still[p.Item.ID] {
				rep.Identified++
			}
		}
	}
	return rep, errors.Join(runErr, rerr)
}
