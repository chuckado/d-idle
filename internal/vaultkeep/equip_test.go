package vaultkeep

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/vault"
)

// fakeHand is a hero's body over the VaultFake: worn items live in api.Carried with Where worn and a body slot name.
type fakeHand struct {
	api       *d2dctl.VaultFake
	body      map[int]string // unit -> slot
	names     map[int]string
	protected map[uint32]bool
	failWear  error
	noWear    bool  // Wear returns ok but nothing is worn (verify must catch it)
	noStash   error // the stash has no room (nil: it has)
	class     string
}

func newHand(api *d2dctl.VaultFake, class string) *fakeHand {
	return &fakeHand{api: api, body: map[int]string{}, names: map[int]string{}, protected: map[uint32]bool{}, class: class}
}

func (h *fakeHand) wornIn(slot string) int {
	for u, s := range h.body {
		if s == slot {
			return u
		}
	}
	return 0
}

// slotsOf: rings go to both ring slots, a "cap" to head; anything else is not wearable.
func slotsOf(code string) []string {
	switch code {
	case "rin":
		return []string{"rring", "lring"}
	case "cap":
		return []string{"head"}
	case "wsp":
		return []string{"rhand"}
	}
	return nil
}

func (h *fakeHand) Preview(rec vault.Rec) Preview {
	sl := slotsOf(rec.Code)
	if sl == nil {
		return Preview{Why: "charms and potions are not worn (belt and backpack are out of scope)"}
	}
	p := Preview{Wearable: true}
	for _, s := range sl {
		v := SlotView{Slot: s, Label: s, Can: true}
		switch s {
		case "rring":
			v.Side = SlotRight
		case "lring":
			v.Side = SlotLeft
		}
		if rec.Ilvl > 60 && h.class == "sorceress" { // a stand-in for a requirement
			v.Can, v.Why = false, "needs level 61"
		}
		if u := h.wornIn(s); u != 0 {
			v.Worn = &Worn{Unit: uint32(u), Name: h.names[u]}
		}
		p.Slots = append(p.Slots, v)
	}
	return p
}

func (h *fakeHand) Pick(unit uint32, slot string) (string, error) {
	switch slot {
	case SlotLeft:
		return "lring", nil
	case SlotRight:
		return "rring", nil
	}
	for _, c := range h.api.Carried {
		if uint32(c.Rec.UnitID) != unit {
			continue
		}
		sl := slotsOf(c.Rec.Code)
		for _, s := range sl {
			if h.wornIn(s) == 0 {
				return s, nil
			}
		}
		return sl[len(sl)-1], nil // both worn: the weaker one (the fake says the last)
	}
	return "", errors.New("not in the backpack")
}

func (h *fakeHand) Wear(_ context.Context, unit uint32, slot string) ([]Worn, error) {
	if h.failWear != nil {
		return nil, h.failWear
	}
	var off []Worn
	if u := h.wornIn(slot); u != 0 {
		delete(h.body, u)
		for i := range h.api.Carried {
			if h.api.Carried[i].Rec.UnitID == u {
				h.api.Carried[i].Where = d2dctl.FakePack
			}
		}
		off = append(off, Worn{Unit: uint32(u), Name: h.names[u]})
	}
	if h.noWear {
		return off, nil
	}
	for i := range h.api.Carried {
		if uint32(h.api.Carried[i].Rec.UnitID) == unit {
			h.api.Carried[i].Where = d2dctl.FakeWorn
			h.body[int(unit)] = slot
			h.names[int(unit)] = "item" + fmt.Sprint(unit)
		}
	}
	return off, nil
}

func (h *fakeHand) StashRoom(string) error { return h.noStash }

func (h *fakeHand) Stash(_ context.Context, unit uint32) error {
	if h.noStash != nil {
		return h.noStash
	}
	for i := range h.api.Carried {
		if uint32(h.api.Carried[i].Rec.UnitID) == unit && h.api.Carried[i].Where == d2dctl.FakePack {
			h.api.Carried[i].Where = d2dctl.FakeStash
			return nil
		}
	}
	return errors.New("not in the backpack")
}

func (h *fakeHand) WornUnit(slot string) uint32 { return uint32(h.wornIn(slot)) }
func (h *fakeHand) Protect(unit uint32)         { h.protected[unit] = true }

// wear puts a hero's own item on the body (a worn start state).
func (h *fakeHand) wear(slot, code, extra string) int {
	rec := d2dctl.FakeRecord(0, code, 6, 20, extra)
	res, _ := h.api.ItemCreate(context.Background(), rec.Norm, d2dctl.CreateOpts{Pack: true})
	for i := range h.api.Carried {
		if h.api.Carried[i].Rec.UnitID == res.ID {
			h.api.Carried[i].Where = d2dctl.FakeWorn
		}
	}
	h.body[res.ID] = slot
	h.names[res.ID] = "Old " + extra
	return res.ID
}

func equipRig(t *testing.T, class string) (*rig, *fakeHand) {
	r := newRig(t, class)
	h := newHand(r.api.VaultFake, class)
	r.k.Hand = h
	r.k.EquipPath = r.dir + "/equip.json"
	return r, h
}

func (r *rig) addID(i int, code string) string {
	r.t.Helper()
	rec := d2dctl.FakeRecord(i, code, 6, 30+i, fmt.Sprintf("roll%d", i))
	it, err := r.st.Add(vault.Item{Raw: rec.Norm, Code: code, Name: fmt.Sprintf("%s %d", code, i), Quality: 6, Ilvl: 30 + i, Identified: true})
	if err != nil {
		r.t.Fatal(err)
	}
	return it.ID
}

func (r *rig) order(id, slot string) Result {
	return r.k.EquipEnqueue(context.Background(), id, slot, false)
}

func (r *rig) carried(where string) int {
	n := 0
	for _, c := range r.api.VaultFake.Carried {
		if c.Where == where {
			n++
		}
	}
	return n
}

func TestEquipOneTapEmptySlot(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r, h := equipRig(t, class)
			id := r.addID(1, "cap")
			if res := r.order(id, ""); !res.OK {
				t.Fatalf("order: %+v", res)
			}
			if res := r.order(id, ""); !res.OK || !res.Already {
				t.Fatalf("a second press must be idempotent: %+v", res)
			}
			if o := r.k.EquipOrderOf(id); o == nil || o.State != EqQueued {
				t.Fatalf("order %+v", o)
			}
			lines := r.run()
			o := r.k.EquipOrderOf(id)
			if o.State != EqDone || o.Worn != "head" {
				t.Fatalf("order %+v lines %v", o, lines)
			}
			if rec := r.state(id); rec.State != vault.Kept {
				t.Fatalf("record %s", rec.State)
			}
			if h.wornIn("head") == 0 || r.carried(d2dctl.FakePack) != 0 || r.carried(d2dctl.FakeStash) != 0 {
				t.Fatalf("carried: worn %d pack %d stash %d", r.carried(d2dctl.FakeWorn), r.carried(d2dctl.FakePack), r.carried(d2dctl.FakeStash))
			}
			if r.saves < 1 || r.k.Pending() != 0 {
				t.Fatalf("saves %d pending %d", r.saves, r.k.Pending())
			}
		})
	}
}

// MUST FAIL if the displaced item is dropped, sold or lost: it lands in the vault as a new live record and leaves the hero once.
func TestEquipSwapKeepsTheOldItem(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r, h := equipRig(t, class)
			old := h.wear("head", "cap", "oldhelm")
			id := r.addID(1, "cap")
			r.order(id, "")
			before := len(r.st.Dump())
			r.run()
			o := r.k.EquipOrderOf(id)
			if o.State != EqDone || o.Where != "vault" || len(o.Displaced) != 1 {
				t.Fatalf("order %+v", o)
			}
			if h.wornIn("head") == old || h.wornIn("head") == 0 {
				t.Fatal("new item must be worn")
			}
			live := 0
			for _, rec := range r.st.Dump() {
				if rec.State == vault.Recorded && rec.Name == "Old oldhelm" {
					live++
				}
			}
			if live != 1 || len(r.st.Dump()) != before+1 {
				t.Fatalf("the old helm must be one live vault record: %d", live)
			}
			for _, c := range r.api.VaultFake.Carried {
				if c.Rec.UnitID == old {
					t.Fatal("the old item must have left the hero for the vault")
				}
			}
		})
	}
}

// MUST FAIL if a full vault makes the swap drop the old item: it stays in the backpack, protected.
func TestEquipSwapVaultFullKeepsItInBackpack(t *testing.T) {
	r, h := equipRig(t, "paladin")
	h.noStash = errors.New("the stash is full")
	old := h.wear("head", "cap", "oldhelm")
	id := r.addID(1, "cap")
	extra := r.addID(2, "cap")
	if err := r.st.Pin(extra); err != nil {
		t.Fatal(err)
	}
	r.order(id, "")
	r.cap = 1 // full of pinned records: Add is refused (ErrFull)
	r.run()
	o := r.k.EquipOrderOf(id)
	if o.State != EqDone || o.Where != "backpack" {
		t.Fatalf("order %+v", o)
	}
	if !h.protected[uint32(old)] {
		t.Fatal("the displaced item must be protected from the chores")
	}
	found := false
	for _, c := range r.api.VaultFake.Carried {
		if c.Rec.UnitID == old && c.Where == d2dctl.FakePack {
			found = true
		}
	}
	if !found {
		t.Fatal("the old item must still be in the backpack")
	}
}

func TestEquipRingSidesAndAuto(t *testing.T) {
	for _, class := range classes {
		for _, side := range []string{SlotLeft, SlotRight} {
			t.Run(class+"/"+side, func(t *testing.T) {
				r, h := equipRig(t, class)
				l := h.wear("lring", "rin", "L")
				rr := h.wear("rring", "rin", "R")
				id := r.addID(1, "rin")
				if res := r.order(id, side); !res.OK {
					t.Fatalf("%+v", res)
				}
				r.run()
				o := r.k.EquipOrderOf(id)
				want := map[string]string{SlotLeft: "lring", SlotRight: "rring"}[side]
				other := map[string]string{SlotLeft: "rring", SlotRight: "lring"}[side]
				if o.State != EqDone || o.Worn != want {
					t.Fatalf("order %+v", o)
				}
				untouched := map[string]int{"lring": l, "rring": rr}[other]
				if h.wornIn(other) != untouched {
					t.Fatalf("the other ring must stay: %s", other)
				}
			})
		}
	}
	// auto fills the empty slot first
	r, h := equipRig(t, "paladin")
	h.wear("lring", "rin", "L")
	id := r.addID(1, "rin")
	r.order(id, "")
	r.run()
	if o := r.k.EquipOrderOf(id); o.Worn != "rring" {
		t.Fatalf("auto must take the empty ring slot: %+v", o)
	}
}

func TestEquipOneOrderPerSlotAndCancel(t *testing.T) {
	r, _ := equipRig(t, "paladin")
	a, b := r.addID(1, "rin"), r.addID(2, "rin")
	if res := r.order(a, SlotLeft); !res.OK {
		t.Fatal(res)
	}
	if res := r.order(b, SlotLeft); res.OK || res.Code != "slot" {
		t.Fatalf("a second order for the left ring must be refused: %+v", res)
	}
	if res := r.order(b, SlotRight); !res.OK {
		t.Fatalf("the right ring is free: %+v", res)
	}
	if res := r.k.EquipCancel(a); !res.OK {
		t.Fatal(res)
	}
	if rec := r.state(a); rec.State != vault.Recorded {
		t.Fatalf("a cancelled order puts the record back: %s", rec.State)
	}
	if res := r.order(b, SlotLeft); res.OK && !res.Already {
		t.Fatal("same record, other slot while active is refused")
	}
}

func TestEquipRefusalsNameTheReason(t *testing.T) {
	r, _ := equipRig(t, "sorceress")
	charm := r.add(1, "cm1", 7)
	if res := r.order(charm, ""); res.OK || !strings.Contains(res.Why, "not worn") {
		t.Fatalf("charm: %+v", res)
	}
	unid := r.addID(2, "cap")
	rec := r.state(unid)
	_ = rec
	hi := fakeHi(r, 3)
	if res := r.order(hi, ""); res.OK || !strings.Contains(res.Why, "level") {
		t.Fatalf("requirement: %+v", res)
	}
}

func fakeHi(r *rig, i int) string {
	rec := d2dctl.FakeRecord(i, "cap", 6, 70, "hi")
	it, err := r.st.Add(vault.Item{Raw: rec.Norm, Code: "cap", Name: "High", Quality: 6, Ilvl: 70, Identified: true})
	if err != nil {
		r.t.Fatal(err)
	}
	return it.ID
}

func TestEquipVerifyAndFailureKeepsItem(t *testing.T) {
	r, h := equipRig(t, "paladin")
	id := r.addID(1, "cap")
	r.order(id, "")
	h.noWear = true
	r.run()
	if o := r.k.EquipOrderOf(id); o.State != EqFailed || o.Code != "verify" {
		t.Fatalf("equip not shown in the equipment must fail: %+v", o)
	}
	if r.carried(d2dctl.FakePack) != 1 {
		t.Fatal("the withdrawn item stays in the backpack")
	}
	r2, h2 := equipRig(t, "sorceress")
	id2 := r2.addID(1, "cap")
	r2.order(id2, "")
	h2.failWear = errors.New("needs 61 strength (you have 30)")
	r2.run()
	if o := r2.k.EquipOrderOf(id2); o.State != EqFailed || !strings.Contains(o.Reason, "strength") {
		t.Fatalf("%+v", o)
	}
}

func TestEquipOffAndRestartResumes(t *testing.T) {
	r, h := equipRig(t, "paladin")
	id := r.addID(1, "cap")
	r.order(id, "")
	r.on = false
	if ok, code, _ := r.k.EquipGate(context.Background()); ok || code != "off" {
		t.Fatalf("gate %v %s", ok, code)
	}
	r.run()
	if r.k.EquipOrderOf(id).State != EqQueued {
		t.Fatal("off: nothing moves")
	}
	r.on = true
	r.ready = "not in town"
	r.run()
	if r.k.EquipOrderOf(id).State != EqQueued {
		t.Fatal("outside town: nothing moves")
	}
	r.ready = ""
	r.run()
	if r.k.EquipOrderOf(id).State != EqDone {
		t.Fatalf("%+v", r.k.EquipOrderOf(id))
	}
	_ = h
	// a bot restart between the withdraw and the equip: the order file survives, the item is found by its bytes
	r2, h2 := equipRig(t, "sorceress")
	id2 := r2.addID(1, "cap")
	r2.order(id2, "")
	h2.failWear = context.Canceled
	r2.run()
	if o := r2.k.EquipOrderOf(id2); o.State != EqEquipping {
		t.Fatalf("%+v", o)
	}
	path := r2.k.EquipPath
	r2.open()
	r2.k.Hand, r2.k.EquipPath = h2, path
	h2.failWear = nil
	r2.run()
	if o := r2.k.EquipOrderOf(id2); o == nil || o.State != EqDone {
		t.Fatalf("restart: %+v", o)
	}
	if h2.wornIn("head") == 0 {
		t.Fatal("not worn after the restart")
	}
}

// An UNREVEALED record equips: the game lets a player wear an item the affixes of which are unseen, so Equip never asks for a
// reveal first (user decision 2026-11-21).
func TestEquipUnrevealedRecord(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r, h := equipRig(t, class)
			rec := d2dctl.FakeRecord(1, "cap", 4, 31, "unrevealed")
			it, err := r.st.Add(vault.Item{Raw: rec.Norm, Code: "cap", Name: "Unidentified cap", Quality: 4, Ilvl: 31})
			if err != nil {
				t.Fatal(err)
			}
			if res := r.order(it.ID, ""); !res.OK || res.Code == "reveal" {
				t.Fatalf("an unrevealed record is refused: %+v", res)
			}
			r.run()
			if o := r.k.EquipOrderOf(it.ID); o == nil || o.State != EqDone || h.wornIn("head") == 0 {
				t.Fatalf("order %+v worn %d", o, h.wornIn("head"))
			}
			if got := r.state(it.ID); got.Identified {
				t.Fatalf("wearing does not reveal it: %+v", got)
			}
		})
	}
}

// MUST FAIL if a full Vault makes the displaced item land in the backpack although the stash has room, or loses it: the
// order's own record left the Vault, a pinned record filled that place again, the stash takes the old helm.
func TestEquipSwapFullVaultGoesToTheStash(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r, h := equipRig(t, class)
			old := h.wear("head", "cap", "oldhelm")
			id := r.addID(1, "cap")
			extra := r.addID(2, "cap")
			if err := r.st.Pin(extra); err != nil {
				t.Fatal(err)
			}
			r.order(id, "")
			r.cap = 1
			r.run()
			o := r.k.EquipOrderOf(id)
			if o.State != EqDone || o.Where != "stash" || len(o.Displaced) != 1 {
				t.Fatalf("order %+v", o)
			}
			if h.wornIn("head") == 0 || h.wornIn("head") == old {
				t.Fatal("the new helm must be worn")
			}
			var stashed, packed int
			for _, c := range r.api.VaultFake.Carried {
				if c.Rec.UnitID == old {
					switch c.Where {
					case d2dctl.FakeStash:
						stashed++
					case d2dctl.FakePack:
						packed++
					}
				}
			}
			if stashed != 1 || packed != 0 {
				t.Fatalf("old helm: stash %d pack %d", stashed, packed)
			}
		})
	}
}

// MUST FAIL if an order that would leave the old item without a home is accepted: the Vault is over its capacity and the
// stash is full. The refusal names the reason and changes nothing (the record stays in the Vault, the body is untouched).
func TestEquipRefusedWhenNothingCanTakeTheOldItem(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r, h := equipRig(t, class)
			h.noStash = errors.New("the stash is full")
			old := h.wear("head", "cap", "oldhelm")
			id := r.addID(1, "cap")
			r.addID(2, "cap")
			r.cap = 1 // two records in a Vault of one: over capacity
			res := r.order(id, "")
			if res.OK || res.Code != "room" || res.Why != ReasonNoRoom {
				t.Fatalf("order must be refused for room: %+v", res)
			}
			if r.k.EquipOrderOf(id) != nil || r.state(id).State != vault.Recorded {
				t.Fatal("a refused order changes nothing")
			}
			if h.wornIn("head") != old {
				t.Fatal("the worn helm must stay")
			}
			// an empty slot needs no home for anything: the same Vault accepts the order
			h.body = map[int]string{}
			for i := range r.api.VaultFake.Carried {
				if r.api.VaultFake.Carried[i].Rec.UnitID == old {
					r.api.VaultFake.Carried[i].Where = d2dctl.FakePack
				}
			}
			if res := r.order(id, ""); !res.OK {
				t.Fatalf("an empty slot must be accepted: %+v", res)
			}
		})
	}
}

// With stash room an over-capacity Vault still takes the order (the old item goes to the stash).
func TestEquipOverCapacityWithStashRoomIsAccepted(t *testing.T) {
	r, h := equipRig(t, "sorceress")
	h.wear("head", "cap", "oldhelm")
	id := r.addID(1, "cap")
	r.addID(2, "cap")
	r.cap = 1
	if res := r.order(id, ""); !res.OK {
		t.Fatalf("%+v", res)
	}
}
