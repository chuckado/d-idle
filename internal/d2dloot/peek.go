package d2dloot

import (
	"context"
	"errors"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dctl"
)

// Peeker is the part of the fork's vault verbs the classifier uses (d2dctl.VaultAPI satisfies it).
type Peeker interface {
	GroundFull(ctx context.Context, id int) ([]d2dctl.VaultRecord, error)
	VaultCaps(ctx context.Context) (d2dctl.VaultCaps, error)
}

// peeker is the per-Source state of the real-quality read. All fields are guarded by Source.mu.
type peeker struct {
	api     Peeker
	checked bool // caps were read
	off     bool // the build has no `ground full`: label guess only
	real    map[uint32]peeked
}

// peeked is the real data of one floor unit (wire shape: d2dctl.ParseVaultLine; V in the fork's own test, U in a bot run).
// quality 0 = the verb did not answer for this unit.
type peeked struct {
	quality    int
	ilvl       int
	identified bool
	ethereal   bool
}

// peekTypes are the item types worth a peek: gear, jewelry, charms. Potions, scrolls, keys, gems, runes and gold are
// never asked about (their label is exact), which bounds the cost of a loot pass to one call, and only when a new unit
// of these types is on the floor.
var peekTypes = []string{"armo", "weap", "ring", "amul", "jewl", "char"}

// neverNormal are the types that never drop normal: a label equal to the base name is an unidentified magic or rare
// item there, so without real data their quality is unknown (0), not normal.
var neverNormal = []string{"ring", "amul", "jewl", "char"}

// UsePeek turns on the real-quality read through p (nil keeps the label guess). It returns s for chaining.
func (s *Source) UsePeek(p Peeker) *Source {
	s.mu.Lock()
	s.pk.api = p
	s.mu.Unlock()
	return s
}

func (s *Source) base(class int) bool {
	return s.DB != nil && class >= 0 && class < len(s.DB.Items) && s.DB.Items[class] != nil
}

func (s *Source) typeIn(class int, set []string) bool {
	if !s.base(class) {
		return false
	}
	b := s.DB.Items[class]
	for _, want := range set {
		if s.DB.IsA(b.Type, want) || b.Type2 != "" && s.DB.IsA(b.Type2, want) {
			return true
		}
	}
	return false
}

// wantsPeek: gear, jewelry and charms, plus a base the table does not know (it cannot be judged otherwise).
func (s *Source) wantsPeek(class int) bool {
	if !s.base(class) {
		return true
	}
	if b := s.DB.Items[class]; b.Stackable || b.Type == "gold" {
		return false
	}
	return s.typeIn(class, peekTypes)
}

// peekGround writes the real quality, item level and identified and ethereal flags into the floor units that want them.
// One `ground full` reads the whole floor, only when a unit not seen before is there; results are cached per unit id and
// dropped when the unit leaves the floor. Without the verb (feature-detected once) a never-normal base keeps quality 0.
func (s *Source) peekGround(snap *d2.Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pk := &s.pk
	if pk.real == nil {
		pk.real = map[uint32]peeked{}
	}
	here := map[uint32]bool{}
	var want []uint32
	for i := range snap.Ground {
		u := &snap.Ground[i]
		if u.Item == nil || !s.wantsPeek(u.Class) {
			continue
		}
		here[u.ID] = true
		if _, ok := pk.real[u.ID]; !ok {
			want = append(want, u.ID)
		}
	}
	for id := range pk.real {
		if !here[id] {
			delete(pk.real, id)
		}
	}
	if len(want) > 0 && pk.api != nil && !pk.off {
		s.fetchLocked(want)
	}
	for i := range snap.Ground {
		u := &snap.Ground[i]
		if u.Item == nil || !here[u.ID] {
			continue
		}
		if r := pk.real[u.ID]; r.quality > 0 {
			u.Item.Quality = r.quality
			if r.ilvl > 0 {
				u.Item.ILvl = r.ilvl
			}
			if r.identified {
				u.Item.Flags |= d2.FlagIdentified
			}
			if r.ethereal {
				u.Item.Flags |= d2.FlagEthereal
			}
		} else if u.Item.Quality == d2.QualityNormal && s.typeIn(u.Class, neverNormal) {
			u.Item.Quality = 0
		}
	}
}

// fetchLocked does the caps check once and one `ground full` for the floor; s.mu is held.
func (s *Source) fetchLocked(want []uint32) {
	pk := &s.pk
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if !pk.checked {
		caps, err := pk.api.VaultCaps(ctx)
		if err != nil {
			return // not final: ask again on the next pass
		}
		pk.checked = true
		if !caps.GroundFull {
			pk.off = true
			return
		}
	}
	recs, err := pk.api.GroundFull(ctx, -1)
	switch {
	case errors.Is(err, d2dctl.ErrVaultUnsupported), errors.Is(err, d2dctl.ErrVaultOff):
		pk.off = true
		return
	case err != nil:
		return // transient: the label guess stands for this pass
	}
	for _, id := range want { // a unit the verb did not return is cached as unknown, so it is not asked for again
		pk.real[id] = peeked{}
	}
	for _, r := range recs {
		pk.real[uint32(r.UnitID)] = peeked{quality: r.Quality, ilvl: r.ILvl, identified: r.Identified, ethereal: r.Ethereal}
	}
}
