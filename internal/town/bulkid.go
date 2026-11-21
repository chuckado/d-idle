package town

import (
	"context"
	"errors"
	"fmt"

	"d2lod-bot/internal/bulkid"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/loot"
)

// BulkID wires the bulk identify chore (docs/design/bulk-identify-2026-11-18.md): the policy, the filter table and the
// journal that survives a restart. The chore never starts a visit of its own.
type BulkID struct {
	Filter   *bulkid.Filter
	Tuning   func() env.IdentifyTuning // identify.* knobs
	Measured func() bool               // a measured run forces manual (bulkid.Effective)
	Journal  *bulkid.Journal
	// Rescued says whether the Search for Cain is done for the hero: quest record 4 bit 1 or 0 (bulkid.CainRescued, V
	// docs/probes/search-for-cain-l2-2026-11-18.md). Cain's town position is never assumed: he is found by class.
	Rescued func(*d2.Snapshot) bool
}

// SetBulkID installs the bulk identify chore; nil switches it off.
func (t *Town) SetBulkID(b *BulkID) { t.bulk = b; t.bulkIdle = -1 }

func (b *BulkID) tuning() env.IdentifyTuning {
	if b.Tuning == nil {
		return env.DefaultIdentifyTuning()
	}
	return b.Tuning()
}

func (b *BulkID) measured() bool { return b.Measured != nil && b.Measured() }

// cainHere reports whether an alive monster of Cain's classes stands in the hero's town (by class, not by position).
func cainHere(s *d2.Snapshot) bool {
	if s == nil || s.Player == nil || !s.Player.InTown() {
		return false
	}
	for i := range s.Monsters {
		if m := &s.Monsters[i]; m.Alive() && IsCainClass(m.Class) {
			return true
		}
	}
	return false
}

func packCount(s *d2.Snapshot) int { return len(s.Inventory.Container(d2.ContInventory)) }

// needBulk is the cheap check before the chore opens anything: a journal with items to bring back, or an active policy
// with Cain rescued and standing here (policy 2 also works without him), and not the same pack that planned nothing
// last time.
func (t *Town) needBulk(s *d2.Snapshot) bool {
	b := t.bulk
	if b == nil || s == nil || s.Inventory == nil {
		return false
	}
	if b.Journal != nil && len(b.Journal.IDs) > 0 {
		return true
	}
	tu := bulkid.Effective(b.tuning(), b.measured())
	switch tu.Policy {
	case env.IdentifyCainFree:
		if !cainHere(s) || b.Rescued == nil || !b.Rescued(s) {
			return false
		}
	case env.IdentifyScrolls:
	default:
		return false
	}
	return packCount(s) != t.bulkIdle
}

// bulkSide is the town's side of a trip: it implements bulkid.Mover, Lister, CainTalk and ScrollIO. The stash is opened
// whenever a call needs it, so the trip survives the panels Cain's dialog closes.
type bulkSide struct {
	t *Town
	a *Act
}

func (b bulkSide) stashOpen(ctx context.Context) (*d2.Snapshot, error) {
	s, err := b.t.current()
	if err != nil {
		return nil, err
	}
	if s.UI.Open(d2.UIStash) {
		return s, nil
	}
	if err := b.t.openStash(ctx, b.a); err != nil {
		return nil, err
	}
	return b.t.current()
}

// item builds the planner's view of a unit: type, quality, level and the base value, nothing a tooltip hides.
func (t *Town) bulkItem(u d2.Unit, place string, f *bulkid.Filter) bulkid.Item {
	it := bulkid.Item{ID: u.ID, Place: place}
	if u.Item == nil {
		return it
	}
	it.Quality, it.Ilvl = loot.QualityName(u.Item.Quality), u.Item.ILvl
	it.W, it.H = t.h.Size(u)
	if b := t.base(u); b != nil {
		it.Code = b.Code
		for _, code := range append([]string{b.Type, b.Type2}, f.TypeCodes()...) {
			if code != "" && (code == b.Type || code == b.Type2 || t.e.Data.IsA(b.Type, code)) && !has(it.Types, code) {
				it.Types = append(it.Types, code)
			}
		}
		it.Value = loot.SellValue(t.e.Data, b, u.Item, 0)
	}
	return it
}

func has(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

func (b bulkSide) list(s *d2.Snapshot, c d2.Container, place string, onlyUnid bool) []bulkid.Item {
	var out []bulkid.Item
	for _, u := range s.Inventory.Container(c) {
		if u.Item == nil || onlyUnid && u.Item.Identified() {
			continue
		}
		out = append(out, b.t.bulkItem(u, place, b.t.bulk.Filter))
	}
	return out
}

func (b bulkSide) Pack(ctx context.Context) ([]bulkid.Item, error) {
	s, err := b.t.current()
	if err != nil {
		return nil, err
	}
	return b.list(s, d2.ContInventory, bulkid.PlacePack, false), nil
}

func (b bulkSide) Stash(ctx context.Context) ([]bulkid.Item, error) {
	s, err := b.stashOpen(ctx)
	if err != nil {
		return nil, err
	}
	return b.list(s, d2.ContStash, bulkid.PlaceStash, false), nil
}

func (b bulkSide) Unidentified(ctx context.Context) ([]bulkid.Item, error) {
	s, err := b.stashOpen(ctx)
	if err != nil {
		return nil, err
	}
	return append(b.list(s, d2.ContInventory, bulkid.PlacePack, true), b.list(s, d2.ContStash, bulkid.PlaceStash, true)...), nil
}

// Move puts one item into the other place through the town's own stash move (the item verb when the mover has it, else
// the click path), refusing with bulkid.ErrNoRoom when no spot is free; into the pack the loot floor of free cells stays
// free, so a trip can never fill the pack.
func (b bulkSide) Move(ctx context.Context, id uint32, to string) error {
	s, err := b.stashOpen(ctx)
	if err != nil {
		return err
	}
	u, ok := b.t.findUnit(s, id)
	if !ok {
		return errors.New("the item is not in the inventory or stash")
	}
	if to == bulkid.PlacePack {
		lay, err := b.t.layout(s)
		if err != nil {
			return err
		}
		w, h := b.t.h.Size(u)
		o := b.t.h.Occupancy(s.Inventory, d2.ContInventory, lay.Inventory)
		if _, _, found := o.FindSpot(w, h); !found || o.FreeCells()-w*h < b.t.e.Tune.Loot.FreeCellsKeep {
			return bulkid.ErrNoRoom
		}
	}
	switch err := b.t.moveOne(ctx, id, to == bulkid.PlaceStash); {
	case errors.Is(err, ErrNoRoomPack), errors.Is(err, ErrNoRoomStash):
		return bulkid.ErrNoRoom
	default:
		return err
	}
}

func (b bulkSide) Present(ctx context.Context) bool {
	s, err := b.t.current()
	return err == nil && cainHere(s)
}

// IdentifyAll runs Cain's dialog (CainIdentify) and returns how many carried items turned identified.
func (b bulkSide) IdentifyAll(ctx context.Context) (int, error) {
	s, err := b.t.current()
	if err != nil {
		return 0, err
	}
	before := len(unidentified(s))
	if _, err := b.t.CainIdentify(ctx); err != nil {
		return 0, err
	}
	after := before
	if e, err := b.t.current(); err == nil {
		after = len(unidentified(e))
	}
	return before - after, nil
}

func (b bulkSide) Charges(ctx context.Context) int {
	s, err := b.t.current()
	if err != nil {
		return 0
	}
	tome, _, scrolls := b.t.IdentifyTools(s)
	return max(tome, 0) + scrolls
}

func (b bulkSide) Gold(ctx context.Context) int {
	if s, err := b.t.current(); err == nil {
		return s.Player.Gold
	}
	return 0
}

var errNoCharge = bulkid.ErrNoTool

func (b bulkSide) Identify(ctx context.Context, id uint32) error {
	if b.Charges(ctx) == 0 {
		return errNoCharge
	}
	s, err := b.t.current()
	if err != nil {
		return err
	}
	u, ok := b.t.findUnit(s, id)
	if !ok {
		return errors.New("the item is not in the inventory or stash")
	}
	if u.Item.Where == d2.ContStash {
		if _, err := b.stashOpen(ctx); err != nil {
			return err
		}
		return b.t.identifyInStash(ctx, id)
	}
	return b.t.identifyOne(ctx, u)
}

func (b bulkSide) Buy(ctx context.Context, n int) (int, error) {
	before := b.Charges(ctx)
	if _, err := b.t.buyIdentify(ctx, n, 0, false); err != nil {
		return 0, err
	}
	return max(b.Charges(ctx)-before, 0), nil
}

// bulkTrip is the chore: recover a journal, plan, stage, Cain, verify, return, then the scroll path for what Cain did
// not cover. Errors are the chore's errors; the items are home or in the pack, never gone.
func (t *Town) bulkTrip(ctx context.Context, a *Act, rep *ChoreReport) error {
	b := t.bulk
	side := bulkSide{t, a}
	tu := b.tuning()
	trip := &bulkid.Trip{F: b.Filter, Tu: tu, Measured: b.measured(), L: side, Cain: side,
		S: &bulkid.Stager{M: side, J: b.Journal, Chunk: tu.Chunk, Logf: func(f string, args ...any) { t.e.Log.Info(fmt.Sprintf(f, args...)) }},
		Hero: func(ctx context.Context) bulkid.Hero {
			h := bulkid.Hero{Charges: side.Charges(ctx), Gold: side.Gold(ctx)}
			if s, err := t.current(); err == nil {
				h.CainOpen = b.Rescued != nil && b.Rescued(s)
				if t.tune().IdentifyTome != 0 { // nothing is bought unless town.identifyTome says so
					if base := t.e.Data.ItemByCode("isc"); base != nil {
						h.ScrollPrice = base.Cost
					}
				}
			}
			return h
		}}
	r, err := trip.Run(ctx)
	rep.Identified += r.Identified
	t.e.Log.Info("bulkid trip", "planned", r.Planned, "staged", r.Staged, "identified", r.Identified, "returned", r.Returned,
		"kept", r.Kept, "reason", r.Reason)
	if len(r.ScrollPicks) > 0 {
		sr := bulkid.RunScrolls(ctx, bulkid.Effective(tu, b.measured()), r.ScrollPicks, r.NeedBuy, side.scrollPrice(),
			t.tune().IdentifyTome != 0, side)
		rep.Identified += sr.Identified
		t.e.Log.Info("bulkid scrolls", "identified", sr.Identified, "bought", sr.Bought, "reason", sr.Reason)
		_ = t.h.CloseAll(ctx)
	}
	if s, serr := t.current(); serr == nil {
		t.bulkIdle = -1
		if r.Planned == 0 && len(r.ScrollPicks) == 0 && err == nil {
			t.bulkIdle = packCount(s)
		}
	}
	return err
}

func (b bulkSide) scrollPrice() int {
	if base := b.t.e.Data.ItemByCode("isc"); base != nil {
		return base.Cost
	}
	return 0
}
