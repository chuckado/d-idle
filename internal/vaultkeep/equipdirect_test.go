package vaultkeep

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/bits"
	"strings"
	"testing"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/vault"
)

// dhand is the hero's body for the direct path: the worn set is the DirectFake's, the names are by item bytes.
type dhand struct {
	*fakeHand
	df    *d2dctl.DirectFake
	names map[string]string // item bytes -> worn name
	class string
}

var dslots = map[string][]string{"cap": {"head"}, "arm": {"torso"}, "glv": {"gloves"}, "boo": {"feet"}, "blt": {"belt"}, "amu": {"neck"},
	"shd": {"lhand"}, "wsp": {"rhand"}, "two": {"rhand"}, "rin": {"rring", "lring"}}

func (h dhand) Preview(rec vault.Rec) Preview {
	sl := dslots[rec.Code]
	if sl == nil {
		return Preview{Why: "charms and potions are not worn"}
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
		if rec.Ilvl > 60 && h.class == "sorceress" {
			v.Can, v.Why = false, "needs level 61"
		}
		if x, ok := h.df.Worn[s]; ok {
			v.Worn = &Worn{Name: h.names[x]}
		}
		p.Slots = append(p.Slots, v)
	}
	return p
}

type drig struct {
	*rig
	df *d2dctl.DirectFake
	h  dhand
}

func directRig(t *testing.T, class string) *drig {
	r := newRig(t, class)
	df := &d2dctl.DirectFake{Worn: map[string]string{}}
	df.Rule = func(hex, slot string) *d2dctl.DirectRefused {
		if strings.Contains(hex, hexOf("lvl61")) && class == "sorceress" {
			return &d2dctl.DirectRefused{Code: "level", Why: "needs level 61"}
		}
		if strings.Contains(hex, hexOf("strfail")) {
			return &d2dctl.DirectRefused{Code: "str", Why: "needs 90 strength"}
		}
		return nil
	}
	df.Pushes = func(hex, slot string) []string {
		if strings.Contains(hex, hexOf("two")) {
			return []string{"lhand"}
		}
		return nil
	}
	h := dhand{fakeHand: newHand(r.api.VaultFake, class), df: df, names: map[string]string{}, class: class}
	r.k.Hand, r.k.EquipPath = h, r.dir+"/equip.json"
	r.k.Direct, r.k.DirectOn = df, func() bool { return true }
	return &drig{rig: r, df: df, h: h}
}

// wearOld puts an item of the hero's own on the body, decodable by the fake game.
func (d *drig) wearOld(slot, code, name string) string {
	rec := d2dctl.FakeRecord(0, code, 6, 20, "old"+slot)
	d.df.Worn[slot] = rec.Norm
	d.h.names[rec.Norm] = name
	if d.api.VaultFake.Bytes == nil {
		d.api.VaultFake.Bytes = map[string]d2dctl.VaultRecord{}
	}
	d.api.VaultFake.Bytes[rec.Norm] = rec
	return rec.Norm
}

func (d *drig) add(i int, code, extra string) string {
	rec := d2dctl.FakeRecord(i, code, 6, 30+i, extra)
	it, err := d.st.Add(vault.Item{Raw: rec.Norm, Code: code, Name: fmt.Sprintf("%s %d", code, i), Quality: 6, Ilvl: 30 + i, Identified: true})
	if err != nil {
		d.t.Fatal(err)
	}
	return it.ID
}

func (d *drig) live(zone string) int {
	n := 0
	for _, r := range d.st.Dump() {
		if r.State == vault.Recorded && (zone == "" || r.Found.Zone == zone) {
			n++
		}
	}
	return n
}

func TestDirectEquipSwap(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			d := directRig(t, class)
			d.wearOld("head", "cap", "Old Cap")
			id := d.add(1, "cap", "new")
			res := d.k.EquipEnqueue(context.Background(), id, "", false)
			if !res.OK || !res.Direct || res.Slot != "head" || len(res.Removed) != 1 || res.Where != "vault" {
				t.Fatalf("%+v", res)
			}
			if d.df.Worn["head"] != createHex(mustRec(t, d, "new")) && !strings.Contains(d.df.Worn["head"], hexOf("new")) {
				t.Fatalf("head holds %q", d.df.Worn["head"])
			}
			if got := d.live("previously worn"); got != 1 {
				t.Fatalf("previously worn records: %d", got)
			}
			if rec, _ := d.st.Get(id); rec.State == vault.Recorded {
				t.Fatalf("the equipped record is still live: %+v", rec)
			}
			if o := d.k.EquipOrderOf(id); o == nil || o.State != EqDone || !o.Direct {
				t.Fatalf("order %+v", o)
			}
			if d.saves == 0 {
				t.Fatal("the character was not saved")
			}
		})
	}
}

func mustRec(t *testing.T, d *drig, extra string) vault.Rec {
	for _, r := range d.st.Dump() {
		if strings.Contains(r.Raw, hexOf(extra)) {
			return r
		}
	}
	t.Fatal("no record")
	return vault.Rec{}
}

func hexOf(s string) string { return fmt.Sprintf("%x", s) }

func TestDirectEquipTwoHanderRemovesShield(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			d := directRig(t, class)
			d.wearOld("rhand", "wsp", "Old Sword")
			d.wearOld("lhand", "shd", "Old Shield")
			id := d.add(2, "two", "two")
			res := d.k.EquipEnqueue(context.Background(), id, "", false)
			if !res.OK || len(res.Removed) != 2 || d.live("previously worn") != 2 {
				t.Fatalf("%+v live=%d", res, d.live("previously worn"))
			}
			if _, still := d.df.Worn["lhand"]; still {
				t.Fatal("the shield stayed on")
			}
		})
	}
}

func TestDirectEquipRefusals(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			d := directRig(t, class)
			// unrevealed record: refused before the engine, "reveal first"
			rec := d2dctl.FakeRecord(3, "cap", 6, 30, "sealed")
			un, _ := d.st.Add(vault.Item{Raw: rec.Norm, Code: "cap", Name: "Sealed Cap", Quality: 6, Ilvl: 30})
			res := d.k.EquipEnqueue(context.Background(), un.ID, "", false)
			if res.OK || res.Code != "reveal" || !strings.Contains(res.Why, "Reveal it first") || len(d.df.Calls) != 0 {
				t.Fatalf("unrevealed: %+v calls %v", res, d.df.Calls)
			}
			// requirements: the engine's own refusal comes back as the reason
			id := d.add(61, "cap", "lvl61")
			res = d.k.EquipEnqueue(context.Background(), id, "", false)
			if class == "sorceress" {
				if res.OK || res.Code != "requirements" || !strings.Contains(res.Why, "61") {
					t.Fatalf("level: %+v", res)
				}
				if r, _ := d.st.Get(id); r.State != vault.Recorded {
					t.Fatal("a refused piece left the vault")
				}
			} else if !res.OK {
				t.Fatalf("%+v", res)
			}
			// the engine's own refusal (strength) is the reason, the record stays
			sf := d.add(5, "arm", "strfail")
			if res = d.k.EquipEnqueue(context.Background(), sf, "", false); res.OK || res.Code != "requirements" || res.Why != "needs 90 strength" {
				t.Fatalf("str: %+v", res)
			}
			// a charm is not worn
			ch := d.add(4, "cm1", "charm")
			if res = d.k.EquipEnqueue(context.Background(), ch, "", false); res.OK || res.Code != "unwearable" {
				t.Fatalf("charm: %+v", res)
			}
		})
	}
}

func TestDirectEquipMeasuredRunRefused(t *testing.T) {
	d := directRig(t, "paladin")
	d.k.Measured = func() bool { return true }
	d.k.EquipOn = func() bool { return false }
	id := d.add(1, "cap", "m")
	res := d.k.EquipEnqueue(context.Background(), id, "", false)
	if res.Direct || len(d.df.Calls) != 0 {
		t.Fatalf("a measured run must never reach the verbs: %+v %v", res, d.df.Calls)
	}
	if ok, _, _ := d.k.EquipGate(context.Background()); ok {
		t.Fatal("the gate is open in a measured run")
	}
	// the fork refuses too: the answer is a refusal, nothing leaves the vault
	d.k.Measured = nil
	d.df.Measured = true
	id = d.add(2, "cap", "m2")
	res = d.k.EquipEnqueue(context.Background(), id, "", false)
	if res.OK || res.Code != "off" {
		t.Fatalf("%+v", res)
	}
	if r, _ := d.st.Get(id); r.State != vault.Recorded {
		t.Fatal("the record left the vault")
	}
}

func TestDirectWearSetTenInOneCall(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			d := directRig(t, class)
			d.wearOld("head", "cap", "Old Cap")
			d.wearOld("rring", "rin", "Old Ring")
			codes := []string{"cap", "arm", "glv", "boo", "blt", "amu", "shd", "wsp", "rin", "rin"}
			var ids []string
			for i, c := range codes {
				ids = append(ids, d.add(10+i, c, fmt.Sprintf("set%d", i)))
			}
			out := d.k.EquipSet(context.Background(), ids, true)
			for i, r := range out {
				if !r.OK || !r.Direct {
					t.Fatalf("piece %d: %+v", i, r)
				}
			}
			n := 0
			for _, c := range d.df.Calls {
				if strings.HasPrefix(c, "wearset 10 ") {
					n++
				}
			}
			if n != 1 || len(d.df.Calls) != 1 {
				t.Fatalf("calls %v", d.df.Calls)
			}
			if len(d.df.Worn) != 10 {
				t.Fatalf("worn %d", len(d.df.Worn))
			}
			if d.live("previously worn") != 2 || d.live("") != 2 {
				t.Fatalf("the two old pieces must be the only live records: %d %d", d.live("previously worn"), d.live(""))
			}
		})
	}
}

func TestDirectSetOnePieceRefusedOthersWorn(t *testing.T) {
	d := directRig(t, "sorceress")
	a := d.add(1, "cap", "ok")
	b := d.add(61, "arm", "lvl61")
	out := d.k.EquipSet(context.Background(), []string{a, b}, true)
	if !out[0].OK || out[1].OK || out[1].Code != "requirements" {
		t.Fatalf("%+v", out)
	}
	if r, _ := d.st.Get(b); r.State != vault.Recorded {
		t.Fatal("the refused piece left the vault")
	}
}

func TestDirectOlderImageFallsBack(t *testing.T) {
	d := directRig(t, "paladin")
	d.df.FlagOff = true
	d.k.EquipOn = func() bool { return true }
	id := d.add(1, "cap", "old")
	res := d.k.EquipEnqueue(context.Background(), id, "", false)
	if !res.OK || res.Direct || len(d.df.Calls) != 0 {
		t.Fatalf("the order flow must take over: %+v", res)
	}
	if o := d.k.EquipOrderOf(id); o == nil || o.State != EqQueued {
		t.Fatalf("order %+v", o)
	}
}

func TestDirectFullVaultKeepsTheOldPieceOnTheHero(t *testing.T) {
	d := directRig(t, "paladin")
	d.wearOld("head", "cap", "Old Cap")
	d.cap = 1
	d.open()
	d.k.Hand, d.k.Direct, d.k.DirectOn = d.h, d.df, func() bool { return true }
	d.k.EquipPath = d.dir + "/equip.json"
	id := d.add(1, "cap", "only")
	// the vault is at capacity: the equipped record frees one place, so the swap fits exactly
	res := d.k.EquipEnqueue(context.Background(), id, "", false)
	if !res.OK || res.Where != "vault" {
		t.Fatalf("%+v", res)
	}
}

func TestDirectRepairFinishesAnAppliedOrder(t *testing.T) {
	d := directRig(t, "paladin")
	old := d.wearOld("head", "cap", "Old Cap")
	id := d.add(1, "cap", "r")
	d.k.putEq(&EquipOrder{ID: id, Name: "cap 1", State: EqEquipping, Direct: true, Applied: true, Removed: []string{old}, Worn: "head", Unit: 9})
	d.k.repairDirect(context.Background())
	if o := d.k.EquipOrderOf(id); o.State != EqDone || d.live("previously worn") != 1 {
		t.Fatalf("%+v live %d", o, d.live("previously worn"))
	}
	d.k.putEq(&EquipOrder{ID: "x", State: EqEquipping, Direct: true})
	d.k.repairDirect(context.Background())
	if o := d.k.EquipOrderOf("x"); o.State != EqFailed || o.Code != "interrupted" {
		t.Fatalf("%+v", o)
	}
}

// createHex flips ONLY the identified bit (bit 4 of the third byte: the item flags' identified bit, V live devp30 probe 2026-11-25:
// created units kept durability, defence and properties for hax, skp, lea, rin, cm1, lbl at normal, magic and rare quality) and
// leaves everything else, durability fields included, byte for byte.
func TestCreateHexFlipsOnlyTheIdentifiedBit(t *testing.T) {
	for n, raw := range []string{"4a4d00a0e000000000", "4a4d0080000000", "4a4d00800000ffffffffffff00ff", "4a4d00000000", "4a4d00a0e0000000001122334455667788"} {
		b, _ := hex.DecodeString(raw)
		rec := vault.Rec{Item: vault.Item{Raw: raw, Identified: true, Unlocked: true}}
		got, _ := hex.DecodeString(createHex(rec))
		if len(got) != len(b) {
			t.Fatalf("%d: length changed", n)
		}
		diff := 0
		for i := range b {
			diff += bits.OnesCount8(b[i] ^ got[i])
		}
		already := b[2]&0x10 != 0
		if want := map[bool]int{true: 0, false: 1}[already]; diff != want || got[2]&0x10 == 0 {
			t.Fatalf("%d: %d bits differ, want %d", n, diff, want)
		}
		if h := createHex(vault.Rec{Item: vault.Item{Raw: raw, Identified: true}}); h != raw {
			t.Fatalf("%d: a record whose bytes are not unlocked must stay as dropped", n)
		}
	}
}
