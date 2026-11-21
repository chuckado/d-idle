package vaultkeep

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"d2lod-bot/internal/vault"
)

// One-tap Equip of a Vault item (docs/vault.md, section "Equip"). An order is Keep to the backpack plus an equip in the SAME
// town visit: Enqueue queues the Keep (target inv) and records the order; Run (after runKeep) finds the created unit, wears it
// through Hand (the desk's verified equip path), vaults the displaced item back, saves, and marks the order done.
//
//	queued/withdrawing -> equipping -> done | failed(reason) | cancelled
//
// Safety: only in town (Keeper.Ready), never in a measured run (Keeper.Enabled), one active order per slot, the displaced
// item is recorded in the vault BEFORE it leaves the hero (and stays in the backpack, protected, when the vault is full),
// and the equip result is read back from the equipment, never assumed.

// Order states.
const (
	EqQueued     = "queued"      // accepted, the Keep has not moved yet
	EqWithdraw   = "withdrawing" // the item is being created in the backpack
	EqEquipping  = "equipping"   // it is in the backpack and is being worn
	EqDone       = "done"
	EqFailed     = "failed"
	EqCancelled  = "cancelled"
	ReasonEqOff  = "one-tap Equip is switched off (loot.vaultEquip 0, loot.vault 0 or a measured run)"
	ReasonEqNone = "this bot cannot equip from the vault"
)

// Slot words of an order. Left and right are the game's left and right ring (the Inventory page's words); auto lets the bot
// choose (an empty slot, else the weaker one).
const (
	SlotAuto  = "auto"
	SlotLeft  = "left"
	SlotRight = "right"
)

// Worn is one equipped item as the page and the executor see it.
type Worn struct {
	Unit  uint32   `json:"unit,omitempty"`
	Name  string   `json:"name"`
	Lines []string `json:"lines,omitempty"` // key stats, a few short lines
	Score float64  `json:"score,omitempty"`
}

// SlotView is one body slot an item could go to: what is worn there and whether the candidate may take it.
type SlotView struct {
	Slot    string  `json:"slot"`           // desk slot key: head, rring, lring, rhand ...
	Label   string  `json:"label"`          // "left ring", "helm" ...
	Side    string  `json:"side,omitempty"` // left or right for the ring slots, else ""
	Worn    *Worn   `json:"worn,omitempty"` // nil: empty
	Can     bool    `json:"can"`
	Why     string  `json:"why,omitempty"`  // when not Can
	Note    string  `json:"note,omitempty"` // when Can: what else happens (a shield comes off for a two-hander)
	Upgrade *bool   `json:"upgrade,omitempty"`
	Gain    float64 `json:"gain,omitempty"`
}

// Preview is the answer for the item detail sheet.
type Preview struct {
	Wearable bool       `json:"wearable"`
	Why      string     `json:"why,omitempty"` // not wearable (a charm, a potion), or unidentified: Reveal first
	Reveal   bool       `json:"reveal,omitempty"`
	Req      []string   `json:"req,omitempty"` // "level 24", "strength 61" ...
	Slots    []SlotView `json:"slots"`
}

// Hand is the hero side of Equip. The bot implements it over the inventory desk (verified equip, pins) and the snapshot.
type Hand interface {
	// Preview judges a vault record for this hero: slots, worn items, requirements. Never writes.
	Preview(rec vault.Rec) Preview
	// Pick resolves an order slot (auto, left, right) for the created pack unit to a desk slot key; auto prefers an
	// empty slot, else the weaker worn one. It errors when the unit does not fit.
	Pick(unit uint32, slot string) (string, error)
	// Wear puts pack unit on the slot, verified against the equipment afterwards, and pins that slot. Items that left the
	// body (the displaced one, a shield a two-hander pushed off) are returned with their unit ids; they are in the backpack.
	Wear(ctx context.Context, unit uint32, slot string) ([]Worn, error)
	// Worn reads the slot from a fresh snapshot (0: empty).
	WornUnit(slot string) uint32
	// Protect keeps the bot's chores from selling or dropping the pack unit (the keep mark).
	Protect(unit uint32)
}

// EquipOrder is one one-tap order.
type EquipOrder struct {
	ID        string    `json:"id"` // vault record id
	Name      string    `json:"name"`
	Slot      string    `json:"slot"` // auto, left, right
	Group     []string  `json:"group,omitempty"`
	State     string    `json:"state"`
	Reason    string    `json:"reason,omitempty"`
	Code      string    `json:"code,omitempty"`
	Unit      int       `json:"unit,omitempty"`
	Worn      string    `json:"worn,omitempty"`      // desk slot it went to
	Displaced []string  `json:"displaced,omitempty"` // names of the items that came off
	Where     string    `json:"where,omitempty"`     // "vault", or "backpack" when the vault had no room
	At        time.Time `json:"at"`
	Updated   time.Time `json:"updated"`
}

func (o EquipOrder) active() bool {
	return o.State == EqQueued || o.State == EqWithdraw || o.State == EqEquipping
}

type equipFile struct {
	Version int          `json:"version"`
	Orders  []EquipOrder `json:"orders"`
}

// Equipment holds the orders of one hero. Zero value is unusable: set Path (may be empty for memory only).
type equipBook struct {
	orders map[string]*EquipOrder
	loaded bool
}

func (k *Keeper) bookLocked() *equipBook {
	if k.eq.orders == nil {
		k.eq.orders = map[string]*EquipOrder{}
	}
	if !k.eq.loaded {
		k.eq.loaded = true
		if k.EquipPath != "" {
			if b, err := os.ReadFile(k.EquipPath); err == nil {
				var f equipFile
				if json.Unmarshal(b, &f) == nil {
					for i := range f.Orders {
						o := f.Orders[i]
						k.eq.orders[o.ID] = &o
					}
				}
			}
		}
	}
	return &k.eq
}

func (k *Keeper) saveBookLocked() {
	if k.EquipPath == "" {
		return
	}
	var f equipFile
	f.Version = 1
	for _, o := range k.eq.orders {
		f.Orders = append(f.Orders, *o)
	}
	sort.Slice(f.Orders, func(i, j int) bool { return f.Orders[i].At.Before(f.Orders[j].At) })
	if len(f.Orders) > 200 { // keep the file small: drop the oldest finished ones
		var keep []EquipOrder
		for _, o := range f.Orders {
			if o.active() || len(f.Orders)-len(keep) <= 200 {
				keep = append(keep, o)
			}
		}
		f.Orders = keep[max(0, len(keep)-200):]
	}
	b, err := json.MarshalIndent(f, "", " ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(k.EquipPath), 0o755)
	tmp := k.EquipPath + ".tmp"
	if os.WriteFile(tmp, b, 0o644) == nil {
		_ = os.Rename(tmp, k.EquipPath)
	}
}

// EquipEnabled is the gate of the feature: knob and hand present on top of the Keep gate.
func (k *Keeper) equipEnabled() bool {
	return k.Hand != nil && (k.EquipOn == nil || k.EquipOn())
}

// EquipGate says whether an order may be taken now (the Keep gate plus the Equip knob).
func (k *Keeper) EquipGate(ctx context.Context) (ok bool, code, why string) {
	if k.Hand == nil {
		return false, "off", ReasonEqNone
	}
	if !k.equipEnabled() || !k.enabled() {
		return false, "off", ReasonEqOff
	}
	if g, w := k.Gate(ctx); !g {
		code = "stopped"
		switch w {
		case ReasonDead:
			code = "dead"
		case ReasonVerbs:
			code = "verbs"
		case ReasonOff:
			code = "off"
		}
		return false, code, w
	}
	return true, "", ""
}

// EquipPreview is the sheet's data; the order of the same record, if any, rides along.
func (k *Keeper) EquipPreview(id string) (Preview, *EquipOrder, bool) {
	st := k.store()
	if st == nil || k.Hand == nil {
		return Preview{}, nil, false
	}
	rec, ok := st.Get(id)
	if !ok {
		return Preview{}, nil, false
	}
	p := k.Hand.Preview(rec)
	k.mu.Lock()
	o := k.bookLocked().orders[id]
	var cp *EquipOrder
	if o != nil {
		c := *o
		cp = &c
	}
	k.mu.Unlock()
	return p, cp, true
}

// EquipOrderOf returns the order of a record (nil: none).
func (k *Keeper) EquipOrderOf(id string) *EquipOrder {
	k.mu.Lock()
	defer k.mu.Unlock()
	if o := k.bookLocked().orders[id]; o != nil {
		c := *o
		return &c
	}
	return nil
}

// EquipEnqueue takes one order. Idempotent: the same record and slot again changes nothing (Already); a second order that
// wants the same slot is refused (code "slot"). The Keep itself is queued here, so the bot's next Run withdraws it.
func (k *Keeper) EquipEnqueue(ctx context.Context, id, slot string, wantNow bool) Result {
	res := Result{ID: id}
	if slot == "" {
		slot = SlotAuto
	}
	if slot != SlotAuto && slot != SlotLeft && slot != SlotRight {
		res.Code, res.Why = "slot", "slot is auto, left or right"
		return res
	}
	st := k.store()
	if st == nil || k.Hand == nil {
		res.Code, res.Why = "missing", "there is no vault for this hero yet"
		return res
	}
	rec, found := st.Get(id)
	if !found {
		res.Code, res.Why = "missing", "no such item"
		return res
	}
	k.mu.Lock()
	b := k.bookLocked()
	if o := b.orders[id]; o != nil && o.active() {
		k.mu.Unlock()
		if o.Slot == slot {
			res.OK, res.Already = true, true
			return res
		}
		res.Code, res.Why = "busy", "this item already has an order for another slot: cancel it first"
		return res
	}
	k.mu.Unlock()
	if rec.State != vault.Recorded {
		res.Code, res.Why = "state", "the item is not in the vault any more or is busy ("+string(rec.State)+")"
		return res
	}
	pv := k.Hand.Preview(rec)
	group, why, code := orderGroup(pv, slot)
	if why != "" {
		res.Code, res.Why = code, why
		return res
	}
	k.mu.Lock()
	for _, o := range b.orders {
		if !o.active() {
			continue
		}
		for _, g := range o.Group {
			for _, h := range group {
				if g == h {
					k.mu.Unlock()
					res.Code, res.Why = "slot", "another item is already on its way to that slot ("+o.Name+"): wait or cancel it"
					return res
				}
			}
		}
	}
	k.mu.Unlock()
	if err := st.KeepTo(id, vault.WhereInv); err != nil {
		res.Code, res.Why = codeOf(err), err.Error()
		return res
	}
	now := k.now()
	k.mu.Lock()
	b.orders[id] = &EquipOrder{ID: id, Name: rec.Name, Slot: slot, Group: group, State: EqQueued, At: now, Updated: now}
	k.saveBookLocked()
	if wantNow {
		k.tripWant = true
	}
	k.mu.Unlock()
	res.OK = true
	return res
}

// orderGroup picks the slots the order claims and refuses with a specific reason when the item cannot go there.
func orderGroup(pv Preview, slot string) (group []string, why, code string) {
	if !pv.Wearable {
		if pv.Reveal {
			return nil, pv.Why, "reveal"
		}
		return nil, pv.Why, "unwearable"
	}
	var firstWhy string
	for _, s := range pv.Slots {
		if slot != SlotAuto && s.Side != slot {
			continue
		}
		if !s.Can {
			if firstWhy == "" {
				firstWhy = s.Why
			}
			continue
		}
		group = append(group, s.Slot)
	}
	if len(group) == 0 {
		if firstWhy == "" {
			firstWhy = "it does not fit that slot"
		}
		return nil, firstWhy, "requirements"
	}
	return group, "", ""
}

// EquipCancel withdraws an order that has not made the item leave the vault yet. After the withdraw the item is in the
// backpack: the order then only stops, nothing is put back.
func (k *Keeper) EquipCancel(id string) Result {
	res := Result{ID: id}
	k.mu.Lock()
	o := k.bookLocked().orders[id]
	if o == nil || !o.active() {
		k.mu.Unlock()
		res.OK, res.Already = true, true
		return res
	}
	k.mu.Unlock()
	if st := k.store(); st != nil {
		if r, ok := st.Get(id); ok && r.State == vault.KeepQ && !r.InFlight {
			if err := st.CancelKeep(id); err != nil {
				res.Code, res.Why = codeOf(err), err.Error()
				return res
			}
		} else if ok && r.InFlight {
			res.Code, res.Why = "busy", "the item is being taken out right now"
			return res
		}
	}
	k.mu.Lock()
	if o := k.bookLocked().orders[id]; o != nil {
		o.State, o.Reason, o.Updated = EqCancelled, "cancelled from the sheet", k.now()
		k.saveBookLocked()
	}
	k.mu.Unlock()
	res.OK = true
	return res
}

func (k *Keeper) setOrder(id string, f func(o *EquipOrder)) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if o := k.bookLocked().orders[id]; o != nil {
		f(o)
		o.Updated = k.now()
		k.saveBookLocked()
	}
}

func (k *Keeper) failOrder(id, code, why string) {
	k.setOrder(id, func(o *EquipOrder) { o.State, o.Code, o.Reason = EqFailed, code, why })
}

// EquipPending counts the orders that have work Run can do now (not those still waiting on a Keep).
func (k *Keeper) equipPending() int {
	st := k.store()
	if st == nil || k.Hand == nil {
		return 0
	}
	n := 0
	k.mu.Lock()
	defer k.mu.Unlock()
	for _, o := range k.bookLocked().orders {
		if !o.active() {
			continue
		}
		if r, ok := st.Get(o.ID); ok && (r.State == vault.Kept || o.State == EqEquipping || r.State == vault.Recorded) {
			n++
		}
	}
	return n
}

// EquipInfo counts active orders for the progress line.
func (k *Keeper) EquipInfo() (active int, line string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	var first *EquipOrder
	for _, o := range k.bookLocked().orders {
		if o.active() {
			active++
			if first == nil || o.At.Before(first.At) {
				first = o
			}
		}
	}
	if first != nil {
		line = fmt.Sprintf("Equipping %s: %s", first.Name, orderWord(*first))
		if active > 1 {
			line = fmt.Sprintf("Equipping %d items: %s", active, orderWord(*first))
		}
	}
	return
}

func orderWord(o EquipOrder) string {
	switch o.State {
	case EqQueued:
		return "waiting for town"
	case EqWithdraw:
		return "taking it out of the vault"
	case EqEquipping:
		return "putting it on"
	}
	return o.State
}

// runEquip finishes the orders whose item is now in the backpack. It runs after runKeep in the same visit.
func (k *Keeper) runEquip(ctx context.Context) []string {
	if k.Hand == nil || !k.equipEnabled() || k.equipPending() == 0 && !k.anyActive() {
		return nil
	}
	st := k.store()
	if st == nil {
		return nil
	}
	if k.Ready != nil && k.Ready() != "" {
		return nil
	}
	var lines []string
	k.mu.Lock()
	var ids []string
	for id, o := range k.bookLocked().orders {
		if o.active() {
			ids = append(ids, id)
		}
	}
	k.mu.Unlock()
	sort.Strings(ids)
	for _, id := range ids {
		if ctx.Err() != nil {
			break
		}
		if l := k.equipOne(ctx, st, id); l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

func (k *Keeper) anyActive() bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	for _, o := range k.bookLocked().orders {
		if o.active() {
			return true
		}
	}
	return false
}

func (k *Keeper) equipOne(ctx context.Context, st *vault.Store, id string) string {
	o := k.EquipOrderOf(id)
	rec, ok := st.Get(id)
	if o == nil || !o.active() {
		return ""
	}
	if !ok {
		k.failOrder(id, "missing", "the vault record is gone")
		return ""
	}
	switch rec.State {
	case vault.KeepQ:
		if o.State == EqQueued && rec.InFlight {
			k.setOrder(id, func(o *EquipOrder) { o.State = EqWithdraw })
		}
		return "" // the Keep has not finished: next Run
	case vault.Recorded:
		// The Keep failed or was cancelled by the player: the order ends with the Keep's own reason.
		why := Reason(rec.Reason)
		if why == "" {
			why = "the item could not be taken out"
		}
		k.failOrder(id, "withdraw", why)
		return fmt.Sprintf("equip: %s not taken out: %s", o.Name, why)
	case vault.Kept:
	default:
		k.failOrder(id, "state", "the item is "+string(rec.State))
		return ""
	}
	if o.State != EqEquipping {
		k.setOrder(id, func(o *EquipOrder) { o.State = EqEquipping })
	}
	unit := uint32(o.Unit)
	if unit == 0 {
		unit = uint32(k.unitFor(ctx, rec))
	}
	if unit == 0 {
		k.failOrder(id, "unit", "it was taken out but is not in the backpack (maybe already worn: check the Inventory page)")
		return fmt.Sprintf("equip: %s: taken out, not found in the backpack", o.Name)
	}
	k.setOrder(id, func(o *EquipOrder) { o.Unit = int(unit) })
	k.Hand.Protect(unit) // the chores must not sell it between the withdraw and the equip
	slot, err := k.Hand.Pick(unit, o.Slot)
	if err != nil {
		k.failOrder(id, "requirements", err.Error())
		return fmt.Sprintf("equip: %s not worn: %v (it stays in your backpack)", o.Name, err)
	}
	off, err := k.Hand.Wear(ctx, unit, slot)
	if err != nil {
		if ctx.Err() != nil || errors.Is(err, context.Canceled) {
			return "" // try again next visit; the order stays
		}
		k.failOrder(id, "equip", err.Error())
		return fmt.Sprintf("equip: %s not worn: %v (it stays in your backpack)", o.Name, err)
	}
	if k.Hand.WornUnit(slot) != unit { // read the equipment, do not trust the verb
		k.failOrder(id, "verify", "the game did not show it worn after the equip")
		return fmt.Sprintf("equip: %s: not worn after the equip", o.Name)
	}
	where := "vault"
	var names []string
	for _, w := range off {
		names = append(names, w.Name)
		if !k.vaultBack(ctx, st, w) {
			where = "backpack"
		}
	}
	if err := k.Save(ctx); err != nil {
		k.note("save", "Equip: "+ReasonSave+" ("+err.Error()+")")
	}
	k.setOrder(id, func(o *EquipOrder) { o.State, o.Worn, o.Displaced, o.Where = EqDone, slot, names, where })
	k.mu.Lock()
	k.info.Done++
	k.mu.Unlock()
	if len(names) == 0 {
		return fmt.Sprintf("equip: %s worn in %s", o.Name, slot)
	}
	return fmt.Sprintf("equip: %s worn in %s, %s went to the %s", o.Name, slot, strings.Join(names, ", "), where)
}

// unitFor finds the unit the Keep created for rec in the backpack: the newest pack item with the recorded bytes.
func (k *Keeper) unitFor(ctx context.Context, rec vault.Rec) int {
	have, err := k.API.ItemFull(ctx, -1)
	if err != nil {
		return 0
	}
	best := 0
	for _, h := range have {
		if strings.EqualFold(h.Norm, rec.Raw) && h.UnitID > best {
			best = h.UnitID
		}
	}
	return best
}

// vaultBack records the item that came off in the vault, then removes it from the hero. A failed record leaves it in the
// backpack under the keep mark: never dropped, never sold. It reports whether the item is in the vault now.
func (k *Keeper) vaultBack(ctx context.Context, st *vault.Store, w Worn) bool {
	recs, err := k.API.ItemFull(ctx, int(w.Unit))
	if err != nil || len(recs) == 0 {
		k.Hand.Protect(w.Unit)
		k.note("equip-"+w.Name, "The "+w.Name+" you replaced is in your backpack: the game would not show it for the vault.")
		return false
	}
	r := recs[0]
	it := vault.Item{Raw: r.Norm, Code: r.Code, Name: w.Name, Quality: r.Quality, Ilvl: r.ILvl, Identified: r.Identified,
		Sockets: r.Sockets, Eth: r.Ethereal, Found: vault.Found{Time: k.now().UTC(), Zone: "worn until replaced"}}
	if r.Identified {
		it.Props = r.Props
	} else {
		it.Sealed = r.Props
	}
	it.Pic, it.Tint, it.FileIndex = r.Pic, r.Tint, r.FileIndex()
	if k.Chain != nil {
		it.Chain = k.Chain(r.Code)
	}
	rec, err := st.Add(it)
	if err != nil && !errors.Is(err, vault.ErrDuplicate) {
		k.Hand.Protect(w.Unit)
		k.note("equip-"+w.Name, "The "+w.Name+" you replaced stays in your backpack: the vault could not take it ("+err.Error()+").")
		return false
	}
	if _, err := k.API.ItemVault(ctx, int(w.Unit)); err != nil {
		// still carried: archive the new record again so the item exists once, and keep the carried one safe
		if rec != nil {
			_ = st.Discard(rec.ID, "equip: the replaced item could not leave the hero")
		}
		k.Hand.Protect(w.Unit)
		k.note("equip-"+w.Name, "The "+w.Name+" you replaced stays in your backpack: "+err.Error())
		return false
	}
	return true
}
