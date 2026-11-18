package invdesk

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/itemreq"
	"d2lod-bot/internal/status"
)

var mercCaps = append(append([]string{}, allCaps...), CapItemMercEquip, CapItemMercUnequip)

// mercFake is the backend with the merc verbs. It applies the change to the snapshot the lane reads back, unless lie is
// set (the verb answers ok but the game shows nothing).
type mercFake struct {
	*fakeFast
	snap *d2.Snapshot
	lie  bool
	err  error
}

func (m *mercFake) MercEquip(_ context.Context, id uint32) (bool, error) {
	if err := m.rec("merc", "mequip:"+itoa(id)); err != nil {
		return false, err
	}
	if m.err != nil {
		return false, m.err
	}
	if !m.lie {
		m.snap.Merc = &d2.MercSheet{Hired: true, Alive: true, GearKnown: true, Gear: []d2.Unit{{ID: id}}}
	}
	return false, nil
}

func (m *mercFake) MercUnequip(_ context.Context, slot string) (bool, error) {
	if err := m.rec("merc", "munequip:"+slot); err != nil {
		return false, err
	}
	return false, m.err
}

// mercEq judges by merc class: item 1 (a bow) only the Rogue (271) may wear, item 2 (a spear) only the Desert Guard
// (338), whatever the hero's class.
type mercEq struct{ *fakeEq }

func (mercEq) Judge(_ *d2.Snapshot, _ *d2.Unit, it itemreq.Item, m *itemreq.Merc) itemreq.Judgement {
	j := itemreq.Judgement{Wearable: it.Code != "potion"}
	if m == nil {
		return j
	}
	want := map[string]int{"10": 271, "11": 338}[it.Code]
	if want != 0 && m.Class != want {
		j.Merc = []itemreq.Reason{{Code: itemreq.MercType, Text: "this mercenary does not carry that kind of item"}}
	}
	return j
}

type mercEnv struct {
	*fastEnv
	mf *mercFake
}

func mercRig(t *testing.T, class d2.Class, area, mercClass int) *mercEnv {
	t.Helper()
	e := fastRig(t, class, area)
	mf := &mercFake{fakeFast: &fakeFast{caps: mercCaps, err: map[string]error{}}, snap: e.snap}
	e.fl.Verbs = Serial{Ch: NewChannel(), V: mf}
	e.fl.Fresh = func() (*d2.Snapshot, error) { return e.snap, nil }
	e.d.SetEquipper(mercEq{e.eq})
	base := e.d.c.Inventory
	e.d.c.Inventory = func() status.InventoryData {
		d := base()
		d.Merc = &status.MercInv{Hired: true, Alive: true, Class: mercClass, Level: 20, GearKnown: true,
			Equipment: map[string]*status.Item{"mhead": {ID: 7, Code: "cap", Name: "Cap"}}}
		return d
	}
	return &mercEnv{e, mf}
}

func (e *mercEnv) merc(alive bool) {
	base := e.d.c.Inventory
	e.d.c.Inventory = func() status.InventoryData {
		d := base()
		d.Merc.Alive = alive
		return d
	}
}

func TestMercEquipTwoHirelings(t *testing.T) {
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		// Rogue (271) takes the bow (item 1), refuses the spear (item 2); the Desert Guard the other way round.
		for _, c := range []struct {
			merc   int
			id     uint32
			wantOK bool
		}{{271, 1, true}, {271, 2, false}, {338, 2, true}, {338, 1, false}} {
			e := mercRig(t, class, town1, c.merc)
			r, err := e.d.RequestEquipFor(c.id, "", TargetMerc)
			if c.wantOK {
				if err != nil || r.State != StateDone || r.How != HowFast || r.Target != TargetMerc {
					t.Fatalf("%v merc %d item %d: %+v %v", class, c.merc, c.id, r, err)
				}
				if got := e.mf.log(); len(got) != 1 || got[0] != "mequip:"+itoa(c.id) {
					t.Errorf("calls %v", got)
				}
				continue
			}
			if code(err) != itemreq.MercType || len(e.mf.log()) != 0 {
				t.Errorf("%v merc %d item %d: want merctype and no verb, got %v %v", class, c.merc, c.id, err, e.mf.log())
			}
		}
	}
}

func TestMercEquipRefusalsNeverLookLikeSuccess(t *testing.T) {
	// the verb's own refusals come back with their code, the request is recorded failed, nothing is queued
	for name, want := range map[string]error{
		"merctype":     &VerbRefusal{"merctype", "no"},
		"dead":         &VerbRefusal{"dead", "he died"},
		"requirements": &VerbRefusal{"requirements", "too weak"},
	} {
		e := mercRig(t, d2.Paladin, town1, 271)
		e.mf.err = want
		r, err := e.d.RequestEquipFor(1, "", TargetMerc)
		if code(err) != name || r.State != StateFailed || r.Code != name {
			t.Errorf("%s: %+v %v", name, r, err)
		}
		if n := len(e.d.st.Queue); n != 0 {
			t.Errorf("%s: %d requests queued", name, n)
		}
	}
	// the verb says ok but the game shows nothing: verify, failed
	e := mercRig(t, d2.Sorceress, town1, 271)
	e.mf.lie = true
	if _, err := e.d.RequestEquipFor(1, "", TargetMerc); code(err) != "verify" {
		t.Errorf("lie: %v", err)
	}
	// a transport failure after the send is a timeout (it may have happened), never a success, and nothing falls back to the hero's click path
	e = mercRig(t, d2.Sorceress, town1, 271)
	e.mf.err = errors.New("socket closed")
	if _, err := e.d.RequestEquipFor(1, "", TargetMerc); code(err) != "timeout" || len(e.eq.worn) != 0 || len(e.d.st.Queue) != 0 {
		t.Errorf("transport: %v worn %v queue %d", err, e.eq.worn, len(e.d.st.Queue))
	}
}

func TestMercPreconditions(t *testing.T) {
	e := mercRig(t, d2.Paladin, town1, 271)
	e.merc(false)
	if _, err := e.d.RequestEquipFor(1, "", TargetMerc); code(err) != "dead" {
		t.Errorf("dead: %v", err)
	}
	if _, err := e.d.RequestUnequipFor("mhead", TargetMerc); code(err) != "dead" {
		t.Errorf("dead unequip: %v", err)
	}
	e = mercRig(t, d2.Paladin, 2, 271) // the field (area 2 is not a town)
	if _, err := e.d.RequestEquipFor(1, "", TargetMerc); code(err) != "town" {
		t.Errorf("field: %v", err)
	}
	e = mercRig(t, d2.Paladin, town1, 271)
	if _, err := e.d.RequestEquipFor(1, "bogus", TargetMerc); code(err) != "slot" {
		t.Errorf("slot: %v", err)
	}
	if _, err := e.d.RequestEquipFor(99, "", TargetMerc); code(err) != "item" {
		t.Errorf("item: %v", err)
	}
	if _, err := e.d.RequestUnequipFor("mtorso", TargetMerc); code(err) != "slot" {
		t.Errorf("nothing worn: %v", err)
	}
}

func TestMercUnequip(t *testing.T) {
	e := mercRig(t, d2.Sorceress, town1, 338)
	// the item lands in the pack: make the read-back see it there
	e.snap.Inventory.Items = append(e.snap.Inventory.Items, d2.Unit{ID: 7, Item: &d2.Item{Where: d2.ContInventory}})
	r, err := e.d.RequestUnequipFor("mhead", TargetMerc)
	if err != nil || r.State != StateDone || r.Kind != KindUnequip || r.ID != 7 || r.Target != TargetMerc {
		t.Fatalf("%+v %v", r, err)
	}
	if got := e.mf.log(); len(got) != 1 || got[0] != "munequip:head" {
		t.Errorf("calls %v", got)
	}
	// no room: the fork's nospace refusal is final and says so
	e = mercRig(t, d2.Sorceress, town1, 338)
	e.mf.err = &VerbRefusal{"nospace", "no room in the backpack"}
	if _, err := e.d.RequestUnequipFor("mhead", TargetMerc); code(err) != "nospace" {
		t.Errorf("nospace: %v", err)
	}
}

// Without the capability, with the lane knob off, or without a lane the desk answers exactly as before: code merc.
func TestMercWithoutCapabilityKeepsRefusing(t *testing.T) {
	e := mercRig(t, d2.Paladin, town1, 271)
	e.mf.caps = allCaps
	if _, err := e.d.RequestEquipFor(1, "", TargetMerc); code(err) != "merc" || len(e.mf.log()) != 0 {
		t.Errorf("no cap: %v %v", err, e.mf.log())
	}
	e = mercRig(t, d2.Paladin, town1, 271)
	e.fl.Instant = func() bool { return false }
	if _, err := e.d.RequestEquipFor(1, "", TargetMerc); code(err) != "merc" || len(e.mf.log()) != 0 {
		t.Errorf("instant off: %v", err)
	}
	e = mercRig(t, d2.Paladin, town1, 271)
	e.fl.On = func() bool { return false }
	if _, err := e.d.RequestUnequipFor("mhead", TargetMerc); code(err) != "merc" {
		t.Errorf("verbs off: %v", err)
	}
	e = mercRig(t, d2.Paladin, town1, 271)
	e.d.SetFast(nil)
	if _, err := e.d.RequestEquipFor(1, "", TargetMerc); code(err) != "merc" {
		t.Errorf("no lane: %v", err)
	}
	// a client that does not implement MercVerbs behind Serial: the capability can never be listed, refusing stays
	e = mercRig(t, d2.Paladin, town1, 271)
	e.fl.Verbs = Serial{Ch: NewChannel(), V: e.fv}
	if _, err := e.d.RequestEquipFor(1, "", TargetMerc); code(err) != "merc" {
		t.Errorf("client without merc verbs: %v", err)
	}
}
