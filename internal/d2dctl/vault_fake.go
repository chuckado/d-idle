package d2dctl

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

// VaultFake is an in-memory stand-in for a fork running with --vault-verbs, in the mem.Fake style: no socket, the same
// refusal codes and the same record bytes in and out, so bot code and tests run against VaultAPI offline. It models
// the floor, the stash and pack grids, the worn set, a dead hero, a joined game, the flag, and the key set. U: it
// follows the fork's documented behaviour (docs/control_channel.md in the fork), not the game's real grids.
type VaultFake struct {
	mu sync.Mutex

	Class        string // informational: the hero class the fixture stands for
	FlagOff      bool   // d2d runs without --vault-verbs: caps lists nothing, every verb refuses off
	Dead, Joined bool
	NoGame       bool
	StashCols    int // default 6 x 8 (classic)
	StashRows    int
	PackCols     int // default 10 x 4
	PackRows     int
	Sizes        map[string][2]int // item code -> width,height; default 1x1
	Floor        []VaultRecord     // non-gold floor items, UnitID is the floor unit
	Gold         map[int]bool      // floor unit ids that are gold
	Carried      []FakeCarried
	Stats        VaultStats
	keys         map[string]bool
	nextID       int
}

// Where an item sits on the hero.
const (
	FakeStash = "stash"
	FakePack  = "inv"
	FakeWorn  = "worn"
	FakeBelt  = "belt"
	FakeHand  = "hand"
)

// FakeCarried is an item the fake hero carries.
type FakeCarried struct {
	Rec      VaultRecord
	Where    string
	Col, Row int
}

// NewVaultFake gives a fake hero of a class with a classic stash and pack.
func NewVaultFake(class string) *VaultFake {
	return &VaultFake{Class: class, StashCols: 6, StashRows: 8, PackCols: 10, PackRows: 4, nextID: 100, keys: map[string]bool{}, Gold: map[int]bool{}}
}

// FakeRecord builds a record whose norm hex is derived from its content, so equal content has equal bytes (and a
// different roll differs). The hex is not a real item bitstream: only the fork round-trips real ones.
func FakeRecord(unit int, code string, quality, ilvl int, extra string) VaultRecord {
	norm := hex.EncodeToString([]byte(fmt.Sprintf("N|%s|%d|%d|%s", code, quality, ilvl, extra)))
	return VaultRecord{UnitID: unit, Raw: "00" + norm, Norm: norm, Code: code, Quality: quality, ILvl: ilvl, Identified: quality <= 3}
}

// FormatVaultLine writes a record as the fork writes it (the inverse of ParseVaultLine).
func FormatVaultLine(r VaultRecord) string {
	var b strings.Builder
	fmt.Fprintf(&b, "#%d raw=%s norm=%s ", r.UnitID, r.Raw, r.Norm)
	if r.HasPos {
		fmt.Fprintf(&b, "at=%.1f,%.1f ", r.X, r.Y)
	}
	b2i := func(v bool) int {
		if v {
			return 1
		}
		return 0
	}
	fmt.Fprintf(&b, "code=%s q=%d ilvl=%d uid=%d id'd=%d eth=%d sock=%d def=%d dur=%d/%d qty=%d pre=%d suf=%d rare=%d,%d aff=",
		r.Code, r.Quality, r.ILvl, r.UID, b2i(r.Identified), b2i(r.Ethereal), r.Sockets, r.Defense, r.Dur, r.MaxDur, r.Qty, r.Prefix, r.Suffix, r.Rare1, r.Rare2)
	for _, a := range r.Affixes {
		fmt.Fprintf(&b, "%d.", a)
	}
	fmt.Fprintf(&b, " set=%d uniq=%d rw=%d nprops=%d props=%s children=%d", r.SetID, r.UniqueID, r.RunewordID, r.NProps, r.Props, r.Children)
	return b.String()
}

func (f *VaultFake) gate() error {
	if f.FlagOff {
		return &VaultRefused{Code: VaultErrOff, Why: "vault verbs off"}
	}
	if f.NoGame {
		return &VaultRefused{Code: VaultErrNoGame, Why: "not in a game"}
	}
	return nil
}

func (f *VaultFake) size(code string) (int, int) {
	if s, ok := f.Sizes[code]; ok {
		return s[0], s[1]
	}
	return 1, 1
}

// spot finds the first free cell for a w x h item (column-major scan top to bottom, like the player's search order is
// not modelled: row-major is enough for a fake).
func (f *VaultFake) spot(where string, w, h int) (int, int, bool) {
	cols, rows := f.StashCols, f.StashRows
	if where == FakePack {
		cols, rows = f.PackCols, f.PackRows
	}
	for r := 0; r+h <= rows; r++ {
		for c := 0; c+w <= cols; c++ {
			if f.free(where, c, r, w, h) {
				return c, r, true
			}
		}
	}
	return -1, -1, false
}

func (f *VaultFake) free(where string, col, row, w, h int) bool {
	for _, it := range f.Carried {
		if it.Where != where {
			continue
		}
		iw, ih := f.size(it.Rec.Code)
		if col < it.Col+iw && it.Col < col+w && row < it.Row+ih && it.Row < row+h {
			return false
		}
	}
	return true
}

// GroundFull lists floor items (id < 0: all).
func (f *VaultFake) GroundFull(_ context.Context, id int) ([]VaultRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.gate(); err != nil {
		return nil, err
	}
	var out []VaultRecord
	for _, r := range f.Floor {
		if f.Gold[r.UnitID] || (id >= 0 && r.UnitID != id) {
			continue
		}
		r.HasPos = true
		out = append(out, r)
	}
	return out, nil
}

// GroundTake removes floor item id and returns its record.
func (f *VaultFake) GroundTake(_ context.Context, id int) (VaultRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.gate(); err != nil {
		return VaultRecord{}, err
	}
	if f.Joined {
		return VaultRecord{}, &VaultRefused{Code: VaultErrState, Why: "joined game: the floor is the host's"}
	}
	if f.Dead {
		return VaultRecord{}, &VaultRefused{Code: VaultErrState, Why: "hero is dead"}
	}
	for i, r := range f.Floor {
		if r.UnitID != id {
			continue
		}
		if f.Gold[id] {
			return VaultRecord{}, &VaultRefused{Code: VaultErrGold, Why: "gold is not an item record"}
		}
		f.Floor = append(f.Floor[:i], f.Floor[i+1:]...)
		f.Stats.Takes++
		r.HasPos = true
		return r, nil
	}
	return VaultRecord{}, &VaultRefused{Code: VaultErrItem, Why: "no such ground item"}
}

// ItemFull lists carried items (id < 0: all). Read-only, so a dead hero still answers.
func (f *VaultFake) ItemFull(_ context.Context, id int) ([]VaultRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.gate(); err != nil {
		return nil, err
	}
	var out []VaultRecord
	for _, it := range f.Carried {
		if id < 0 || it.Rec.UnitID == id {
			out = append(out, it.Rec)
		}
	}
	return out, nil
}

// ItemVault removes a stored or belt item and returns its record.
func (f *VaultFake) ItemVault(_ context.Context, id int) (VaultRecord, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.gate(); err != nil {
		return VaultRecord{}, err
	}
	if f.Dead {
		return VaultRecord{}, &VaultRefused{Code: VaultErrState, Why: "hero is dead"}
	}
	for i, it := range f.Carried {
		if it.Rec.UnitID != id {
			continue
		}
		switch it.Where {
		case FakeWorn:
			return VaultRecord{}, &VaultRefused{Code: VaultErrPlace, Why: "worn: take it off first"}
		case FakeHand:
			return VaultRecord{}, &VaultRefused{Code: VaultErrCursor, Why: "the item is in the hand"}
		}
		f.Carried = append(f.Carried[:i], f.Carried[i+1:]...)
		f.Stats.Vaults++
		return it.Rec, nil
	}
	return VaultRecord{}, &VaultRefused{Code: VaultErrItem, Why: "no such item"}
}

// ItemCreate puts the record into the stash or pack under a new unit id.
func (f *VaultFake) ItemCreate(_ context.Context, hexNorm string, o CreateOpts) (CreateResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.gate(); err != nil {
		return CreateResult{}, err
	}
	if f.Dead {
		return CreateResult{}, &VaultRefused{Code: VaultErrState, Why: "hero is dead"}
	}
	if o.Key != "" && f.keys[o.Key] {
		return CreateResult{}, &VaultRefused{Code: VaultErrDup, Why: "key already created in this game"}
	}
	if err := ValidHex(hexNorm); err != nil {
		return CreateResult{}, err
	}
	// The fake accepts exactly what FakeRecord (or a captured record of this fake) produced.
	if raw, err := hex.DecodeString(hexNorm); err != nil || !strings.HasPrefix(string(raw), "N|") {
		return CreateResult{}, &VaultRefused{Code: VaultErrParse, Why: "d2s items: read past end"}
	}
	where := FakeStash
	if o.Pack {
		where = FakePack
	}
	cols, rows := f.StashCols, f.StashRows
	if o.Pack {
		cols, rows = f.PackCols, f.PackRows
	}
	rec := recordFromNorm(hexNorm)
	w, h := f.size(rec.Code)
	col, row := o.Col, o.Row
	if !o.Fixed {
		var ok bool
		if col, row, ok = f.spot(where, w, h); !ok {
			return CreateResult{}, &VaultRefused{Code: VaultErrNoSpace, Why: "no room there"}
		}
	} else if col < 0 || row < 0 || col+w > cols || row+h > rows || !f.free(where, col, row, w, h) {
		return CreateResult{}, &VaultRefused{Code: VaultErrPlace, Why: "overlap or outside the grid"}
	}
	f.nextID++
	rec.UnitID = f.nextID
	f.Carried = append(f.Carried, FakeCarried{Rec: rec, Where: where, Col: col, Row: row})
	if o.Key != "" {
		f.keys[o.Key] = true
	}
	f.Stats.Creates++
	return CreateResult{ID: rec.UnitID, Col: col, Row: row}, nil
}

// recordFromNorm rebuilds a record from fake norm bytes ("N|code|q|ilvl|extra").
func recordFromNorm(norm string) VaultRecord {
	raw, _ := hex.DecodeString(norm)
	p := strings.SplitN(string(raw), "|", 5)
	var q, il int
	if len(p) >= 4 {
		fmt.Sscan(p[2], &q)
		fmt.Sscan(p[3], &il)
	}
	code := ""
	if len(p) >= 2 {
		code = p[1]
	}
	return VaultRecord{Raw: "00" + norm, Norm: norm, Code: code, Quality: q, ILvl: il, Identified: q <= 3}
}

// VaultStats returns the ledger.
func (f *VaultFake) VaultStats(context.Context) (VaultStats, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.gate(); err != nil {
		return VaultStats{}, err
	}
	return f.Stats, nil
}

// VaultCaps is complete unless the flag is off.
func (f *VaultFake) VaultCaps(context.Context) (VaultCaps, error) {
	if f.FlagOff {
		return VaultCaps{}, nil
	}
	return VaultCaps{GroundFull: true, GroundTake: true, ItemFull: true, ItemCreate: true, ItemVault: true}, nil
}

var (
	_ VaultAPI = (*Client)(nil)
	_ VaultAPI = (*VaultFake)(nil)
)
