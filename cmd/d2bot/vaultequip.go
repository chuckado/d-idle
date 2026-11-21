package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/invdesk"
	"d2lod-bot/internal/itemreq"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/vault"
	"d2lod-bot/internal/vaultkeep"
	"d2lod-bot/internal/watch"
)

// One-tap Equip of a vault item: the hero side (vaultkeep.Hand) over the inventory desk, the leveler's gear rule and the
// snapshot, and the page face (watch.VaultEquipper). Docs: docs/vault.md, section "Equip".

// equipOn is the knob loot.vaultEquip (never in a measured run).
func equipOn(e *env.Env) func() bool {
	return func() bool {
		return e != nil && e.Tune != nil && os.Getenv("D2BOT_RUN") == "" && e.Tune.Loot.VaultEquip != 0
	}
}

type equipHand struct {
	ref  *runnerRef
	e    *env.Env
	snap func() *d2.Snapshot
}

var slotKeyOf = func() map[int]string {
	m := map[int]string{}
	for k, b := range invdesk.SlotBody {
		m[b] = k
	}
	return m
}()

var slotLabel = map[string]string{"head": "helm", "neck": "amulet", "torso": "armour", "rhand": "right hand", "lhand": "left hand",
	"rring": "right ring", "lring": "left ring", "belt": "belt", "feet": "boots", "gloves": "gloves"}

// ringSide names the two ring slots with the words of the Inventory page ("left ring" is the game's lring).
func ringSide(slot string) string {
	switch slot {
	case "lring":
		return vaultkeep.SlotLeft
	case "rring":
		return vaultkeep.SlotRight
	}
	return ""
}

func (h equipHand) fresh() *d2.Snapshot {
	if h.e != nil {
		if s, err := h.e.Snapshot(); err == nil && s != nil {
			return s
		}
	}
	return h.snap()
}

// wornUnits maps slot key to the item unit worn there.
func wornUnits(s *d2.Snapshot) map[string]d2.Unit {
	out := map[string]d2.Unit{}
	if s == nil || s.Inventory == nil {
		return out
	}
	for _, u := range s.Inventory.Items {
		if u.Item != nil && u.Item.Where == d2.ContEquipment {
			if k, ok := slotKeyOf[u.Item.Body]; ok {
				out[k] = u
			}
		}
	}
	return out
}

func (h equipHand) baseIndex(code string) int {
	db := h.e.Data
	if db == nil {
		return -1
	}
	b := db.ItemByCode(code)
	if b == nil {
		return -1
	}
	for i, x := range db.Items {
		if x == b {
			return i
		}
	}
	return -1
}

func (h equipHand) score(s *d2.Snapshot, u d2.Unit) float64 {
	var st d2.Stats
	if u.Item != nil && u.Item.Identified() && h.e.Read != nil {
		st, _ = h.e.Read.UnitStats(u.Addr)
	}
	return leveler.Candidate(h.e.Data, s.Player, u, st).Score
}

// Preview judges the vault record for the hero with the gear rule (leveler.Candidate): the same slots, requirements and
// refusals the Inventory page's Wear uses, on a stand-in unit built from the record (class row, quality, level, flags). The
// stat-driven score of an affixed candidate is not readable from the record, so "upgrade" is answered for plain and
// superior items only (U: shown as unknown otherwise).
func (h equipHand) Preview(rec vault.Rec) vaultkeep.Preview {
	s := h.snap()
	db := h.e.Data
	if s == nil || s.Player == nil || db == nil {
		return vaultkeep.Preview{Why: "the hero is not readable right now"}
	}
	idx := h.baseIndex(rec.Code)
	if idx < 0 {
		return vaultkeep.Preview{Why: "this item is not in the game tables"}
	}
	b := db.Items[idx]
	var flags uint32
	if rec.Identified {
		flags |= d2.FlagIdentified
	}
	if rec.Eth {
		flags |= d2.FlagEthereal
	}
	u := d2.Unit{Type: d2.TypeItem, Class: idx, Item: &d2.Item{Quality: rec.Quality, ILvl: rec.Ilvl, Flags: flags, Where: d2.ContInventory}}
	c := leveler.Candidate(db, s.Player, u, nil)
	if len(c.Slots) == 0 {
		why := "this item is not worn"
		if db.IsA(b.Type, "char") {
			why = "charms are not equipped: belt and backpack items are out of scope for Equip"
		}
		return vaultkeep.Preview{Why: why}
	}
	for _, r := range c.Reasons {
		if r.Code == itemreq.Identified {
			return vaultkeep.Preview{Why: "unidentified: reveal it first", Reveal: true}
		}
	}
	p := vaultkeep.Preview{Wearable: true}
	if c.Req.Lvl > 0 {
		p.Req = append(p.Req, fmt.Sprintf("level %d", c.Req.Lvl))
	}
	if c.Req.Str > 0 {
		p.Req = append(p.Req, fmt.Sprintf("strength %d", c.Req.Str))
	}
	if c.Req.Dex > 0 {
		p.Req = append(p.Req, fmt.Sprintf("dexterity %d", c.Req.Dex))
	}
	worn := wornUnits(s)
	var view map[string]*invdesk.ItemView
	if h.ref.inv != nil {
		view = h.ref.inv.View().Equipment
	}
	plain := rec.Quality <= d2.QualitySuperior
	for _, body := range c.Slots {
		key, ok := slotKeyOf[body]
		if !ok {
			continue
		}
		sv := vaultkeep.SlotView{Slot: key, Label: slotLabel[key], Side: ringSide(key)}
		sv.Can = c.Usable
		if !c.Usable {
			sv.Why = c.Why
		}
		if wu, has := worn[key]; has {
			w := &vaultkeep.Worn{Unit: wu.ID, Name: "item"}
			if tw := h.ref.tw.Load(); tw != nil {
				w.Name = tw.ItemName(wu)
			}
			if iv := view[key]; iv != nil {
				if iv.Name != "" {
					w.Name = iv.Name
				}
				w.Lines = iv.Props
				if len(w.Lines) > 4 {
					w.Lines = w.Lines[:4]
				}
			}
			w.Score = h.score(s, wu)
			sv.Worn = w
			if plain && sv.Can {
				up := c.Score > w.Score
				sv.Upgrade, sv.Gain = &up, c.Score-w.Score
			}
		}
		if sv.Can && c.TwoHanded && key == "rhand" {
			if _, shield := worn["lhand"]; shield {
				sv.Note = "takes the shield off (a two-hander uses both hands)"
			}
		}
		if sv.Can && (key == "lhand") {
			if o, two := worn["rhand"]; two && h.twoHanded(s, o) {
				sv.Can, sv.Why = false, "a two-handed weapon is worn: take it off first"
			}
		}
		p.Slots = append(p.Slots, sv)
	}
	sort.SliceStable(p.Slots, func(i, j int) bool { return p.Slots[i].Slot < p.Slots[j].Slot })
	return p
}

func (h equipHand) twoHanded(s *d2.Snapshot, u d2.Unit) bool {
	if h.e.Data == nil || u.Class < 0 || u.Class >= len(h.e.Data.Items) {
		return false
	}
	b := h.e.Data.Items[u.Class]
	return b.TwoHanded && b.Kind == data.KindWeapon
}

// Pick resolves the order's slot word for the created pack unit.
func (h equipHand) Pick(unit uint32, slot string) (string, error) {
	d := h.ref.inv
	if d == nil {
		return "", errors.New("equipping is not available on this bot")
	}
	defer h.ref.holdSnap(h.fresh())()
	fits, err := d.FitSlots(unit)
	if err != nil {
		return "", err
	}
	want := map[string]string{vaultkeep.SlotLeft: "lring", vaultkeep.SlotRight: "rring"}[slot]
	if want != "" {
		for _, f := range fits {
			if f == want {
				return want, nil
			}
		}
		return "", fmt.Errorf("it does not fit the %s ring slot", slot)
	}
	if len(fits) == 1 {
		return fits[0], nil
	}
	s := h.fresh()
	worn := wornUnits(s)
	best, bestScore := "", 0.0
	for _, f := range fits {
		wu, has := worn[f]
		if !has {
			return f, nil // an empty slot first
		}
		if sc := h.score(s, wu); best == "" || sc < bestScore {
			best, bestScore = f, sc
		}
	}
	return best, nil
}

// Wear equips through the desk (verified, pinned) and reports what left the body.
func (h equipHand) Wear(ctx context.Context, unit uint32, slot string) ([]vaultkeep.Worn, error) {
	d := h.ref.inv
	if d == nil {
		return nil, errors.New("equipping is not available on this bot")
	}
	fresh := h.fresh()
	defer h.ref.holdSnap(fresh)()
	before := wornUnits(fresh)
	names := map[uint32]string{}
	for _, u := range before {
		names[u.ID] = "item"
		if tw := h.ref.tw.Load(); tw != nil {
			names[u.ID] = tw.ItemName(u)
		}
	}
	if err := d.EquipNow(ctx, unit, slot); err != nil {
		return nil, err
	}
	after := wornUnits(h.fresh())
	still := map[uint32]bool{}
	for _, u := range after {
		still[u.ID] = true
	}
	var off []vaultkeep.Worn
	for _, u := range before {
		if !still[u.ID] {
			off = append(off, vaultkeep.Worn{Unit: u.ID, Name: names[u.ID]})
		}
	}
	return off, nil
}

func (h equipHand) WornUnit(slot string) uint32 {
	if u, ok := wornUnits(h.fresh())[slot]; ok {
		return u.ID
	}
	return 0
}

// Protect puts the keep mark on the pack unit so no chore sells or drops it. The snapshot may not list a created item yet:
// a few short retries.
func (h equipHand) Protect(unit uint32) {
	if h.ref.inv == nil {
		return
	}
	defer h.ref.holdSnap(h.fresh())()
	for try := 0; try < 6; try++ {
		err := h.ref.inv.SetMark(unit, invdesk.MarkKeep)
		var rf *invdesk.Refusal
		if err == nil || !errors.As(err, &rf) || rf.Code != "item" {
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
}

var _ vaultkeep.Hand = equipHand{}

// equipFace adapts the keeper of a vaultHome to watch.VaultEquipper.
type equipFace struct{ h *vaultHome }

func (f equipFace) k() *vaultkeep.Keeper {
	f.h.mu.Lock()
	defer f.h.mu.Unlock()
	return f.h.keep
}

func (f equipFace) Gate(ctx context.Context) (bool, string, string) {
	k := f.k()
	if k == nil {
		return false, "off", vaultkeep.ReasonEqNone
	}
	return k.EquipGate(ctx)
}

// equipSheet is the reply of GET /api/vault/equip/<id>.
type equipSheet struct {
	Enabled bool   `json:"enabled"`
	Off     string `json:"offWhy,omitempty"` // why Equip cannot be pressed now (knob off, dead hero ...)
	vaultkeep.Preview
	Order *vaultkeep.EquipOrder `json:"order,omitempty"`
	Line  string                `json:"line,omitempty"`
}

func (f equipFace) Preview(id string) (any, bool) {
	k := f.k()
	if k == nil {
		return nil, false
	}
	p, o, ok := k.EquipPreview(id)
	if !ok {
		return nil, false
	}
	v := equipSheet{Preview: p, Order: o}
	v.Enabled = true
	if g, _, why := k.EquipGate(context.Background()); !g {
		v.Enabled, v.Off = false, why
	}
	if o != nil && o.State != vaultkeep.EqDone && o.State != vaultkeep.EqFailed && o.State != vaultkeep.EqCancelled {
		_, v.Line = k.EquipInfo()
	}
	if v.Slots == nil {
		v.Slots = []vaultkeep.SlotView{}
	}
	return v, true
}

func (f equipFace) Enqueue(ctx context.Context, id, slot string, now bool) watch.VaultKeepResult {
	k := f.k()
	if k == nil {
		return watch.VaultKeepResult{ID: id, Code: "off", Why: vaultkeep.ReasonEqNone}
	}
	r := k.EquipEnqueue(ctx, id, slot, now)
	return watch.VaultKeepResult{ID: r.ID, OK: r.OK, Already: r.Already, Code: r.Code, Why: r.Why}
}

func (f equipFace) Cancel(id string) watch.VaultKeepResult {
	k := f.k()
	if k == nil {
		return watch.VaultKeepResult{ID: id, Code: "off", Why: vaultkeep.ReasonEqNone}
	}
	r := k.EquipCancel(id)
	return watch.VaultKeepResult{ID: r.ID, OK: r.OK, Already: r.Already, Code: r.Code, Why: r.Why}
}
