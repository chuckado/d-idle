package loot

import (
	"errors"
	"fmt"
	"sort"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/lootrules"
)

// Preview is the dry run of the Loot Rules sheet: the candidate rule set (nil = the active one) decided for every item
// of the last pack and stash the bot looked at, next to the active set's answer. Nothing is sold, stashed or dropped,
// nothing is sent to the game; the shadow's own counters and recent list are not touched. The sell answer assumes the
// town's trip threshold for a full pack and no "not worth a scroll" exception, and the player's marks are not known here.
func (s *Shadow) Preview(cand, active *lootrules.RuleSet) (doc lootrules.PreviewDoc, err error) {
	s.mu.Lock()
	p := s.lastPass
	s.mu.Unlock()
	if p == nil || p.snap == nil || p.snap.Inventory == nil {
		return doc, errors.New("the bot has not looked at the pack yet; try again in town")
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("preview failed: %v", r)
		}
	}()
	if cand == nil {
		cand = active
	}
	cc, err := lootrules.Compile(cand, p.cx.Class, p.cx.Build)
	if err != nil {
		return doc, lootrules.Errors{{Path: "$", Msg: err.Error()}}
	}
	ac, err := lootrules.Compile(active, p.cx.Class, p.cx.Build)
	if err != nil {
		return doc, err
	}
	sc := SellCtx{Full: p.cx.PackFullPct >= p.l.e.Tune.Town.TripPackUsedPct, ValuableMinPr: p.l.e.Tune.Loot.ValuableMinPrio}
	doc = lootrules.PreviewDoc{Schema: lootrules.PreviewSchema, Name: cand.Name, Items: []lootrules.PreviewRow{}, Totals: map[string]int{},
		Note: "The dry run does not know your keep and sell marks; an item you marked keeps its mark."}
	for _, place := range []struct {
		c     d2.Container
		where string
	}{{d2.ContInventory, lootrules.PlacePack}, {d2.ContStash, lootrules.PlaceStash}} {
		for _, u := range p.snap.Inventory.Container(place.c) {
			if u.Item == nil || u.Item.Flags&d2.FlagStartItem != 0 {
				continue
			}
			it, ierr := p.item(ShadowTidy, u, "", false, nil)
			if ierr != nil {
				continue
			}
			doc.AsOfItems++
			row := lootrules.PreviewRow{ID: u.ID, Name: nameOf(&it), Place: place.where}
			it.Place = place.where
			d, wd := cc.Decide(&it, p.cx), ac.Decide(&it, p.cx)
			row.Tidy, row.WasTidy = lootrules.PreviewAct{Action: d.Action, Rule: d.Rule, Why: d.Why}, wd.Action
			outcome := d.Action
			if place.where == lootrules.PlacePack {
				cs, csd := decideSell(cc, &it, p.cx, sc)
				as, _ := decideSell(ac, &it, p.cx, sc)
				row.Sell = &lootrules.PreviewAct{Action: sellWord(cs), Rule: csd.Rule, Why: csd.Why}
				row.WasSell = sellWord(as)
				if cs {
					outcome = lootrules.Sell
				}
			}
			row.Changed = row.Tidy.Action != row.WasTidy || row.Sell != nil && row.Sell.Action != row.WasSell
			switch outcome {
			case lootrules.Sell:
				doc.Totals["sold"]++
			case lootrules.Stash:
				doc.Totals["stashed"]++
			case lootrules.Drop:
				doc.Totals["dropped"]++
			default:
				doc.Totals["kept"]++
			}
			if row.Changed {
				doc.Changed++
			}
			doc.Items = append(doc.Items, row)
		}
	}
	// changed rows first, then the rest, so the cap never hides a difference
	sort.SliceStable(doc.Items, func(i, j int) bool { return doc.Items[i].Changed && !doc.Items[j].Changed })
	if len(doc.Items) > lootrules.PreviewCap {
		doc.Items, doc.Capped = doc.Items[:lootrules.PreviewCap], true
	}
	return doc, nil
}

func sellWord(sold bool) string {
	if sold {
		return lootrules.Sell
	}
	return lootrules.Keep
}

// PreviewRules is Preview on the process-wide shadow against the active rule set (the store's route).
func PreviewRules(cand *lootrules.RuleSet) (lootrules.PreviewDoc, error) {
	return DefaultShadow.Preview(cand, DefaultEngine.Rules())
}
