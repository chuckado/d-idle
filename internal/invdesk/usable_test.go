package invdesk

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/itemreq"
	"d2lod-bot/internal/merc"
	"d2lod-bot/internal/status"
)

// uworld is the sets-test world plus a judge and an optional mercenary. The judge here is a table (item id to the
// reasons the hero and the mercenary get), because the rule itself is tested against the game tables in
// internal/leveler (judge_test.go); what these tests pin is how the view combines the answers.
type uworld struct {
	*gworld
	merc    *status.MercInv
	heroReq map[uint32][]itemreq.Reason // item-level reasons for the hero
	mercReq map[uint32][]itemreq.Reason
	broken  map[uint32]bool
}

func (u *uworld) Judge(s *d2.Snapshot, unit *d2.Unit, it itemreq.Item, m *itemreq.Merc) itemreq.Judgement {
	id := uint32(0)
	if unit != nil {
		id = unit.ID
	}
	j := itemreq.Judgement{Wearable: it.Code != "potion", Req: itemreq.Req{Lvl: 5, Str: 20, Class: []string{"Paladin"}}}
	if unit != nil {
		j.Hero = u.heroReq[id]
	}
	if m != nil {
		j.Merc = u.mercReq[id]
		if !m.Alive {
			j.Merc = append(append([]itemreq.Reason{}, j.Merc...), itemreq.Reason{Code: itemreq.Dead, Text: "the mercenary is dead"})
		}
	}
	return j
}

// CheckPlayerEquip refuses what the judge says the hero cannot use, with the structured reasons, like the leveler.
func (u *uworld) CheckPlayerEquip(s *d2.Snapshot, id uint32, body int) (string, error) {
	if rs := u.heroReq[id]; len(rs) > 0 {
		return "thing", &itemreq.Error{Err: errors.New("thing: " + rs[0].Text), Reasons: rs}
	}
	return u.gworld.CheckPlayerEquip(s, id, body)
}

func (u *uworld) data() status.InventoryData {
	d := u.gworld.data()
	for i := range d.Inventory {
		if u.broken[d.Inventory[i].ID] {
			d.Inventory[i].MaxDur = 20
		}
	}
	d.Merc = u.merc
	return d
}

func newUWorld(t *testing.T, class d2.Class) (*Desk, *uworld) {
	t.Helper()
	w := &gworld{t: t, class: class, worn: map[string]*gitem{}, failEq: map[uint32]error{}, inTown: true}
	u := &uworld{gworld: w, heroReq: map[uint32][]itemreq.Reason{}, mercReq: map[uint32][]itemreq.Reason{}, broken: map[uint32]bool{}}
	now := time.Unix(1_800_000_000, 0)
	d, err := New(Config{Path: filepath.Join(t.TempDir(), StateFile), Snap: w.snap, Inventory: u.data,
		Now: func() time.Time { now = now.Add(time.Second); return now }})
	if err != nil {
		t.Fatal(err)
	}
	d.SetEquipper(u)
	d.SetUnequipper(w)
	d.SetHands(w)
	d.SetServices(&worldStash{fakeSv: &fakeSv{tome: 1}, w: w})
	return d, u
}

func str(n int) itemreq.Reason {
	return itemreq.Reason{Code: itemreq.Requirements, Stat: "str", Need: n, Have: 30, Text: fmt.Sprintf("needs %d strength (you have 30)", n)}
}

func (v View) pack(id uint32) ItemView {
	for _, it := range v.Backpack {
		if it.ID == id {
			return it
		}
	}
	return ItemView{}
}

// The hero's verdict on a pack item is the equip check's: ok exactly when a slot takes it, for every pack item, for two
// classes (one rule, two views).
func TestUsableHeroIsTheEquipCheck(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		d, u := newUWorld(t, class)
		kit(u.gworld, class)
		u.w().pack = append(u.w().pack, u.w().item("sword", 9, "Other Class Sword", false, "rhand", d2.Barbarian))
		v := d.View()
		for _, it := range v.Backpack {
			if it.Usable == nil || it.Usable.Hero == nil {
				t.Fatalf("%v: %d has no verdict", class, it.ID)
			}
			if it.Usable.Hero.OK != (len(it.CanWear) > 0) {
				t.Errorf("%v: item %d: ok=%v canWear=%v", class, it.ID, it.Usable.Hero.OK, it.CanWear)
			}
			if !it.Usable.Hero.OK && len(it.Usable.Hero.Reasons) == 0 {
				t.Errorf("%v: item %d refused without a reason", class, it.ID)
			}
			if it.Usable.Merc != nil {
				t.Errorf("%v: no mercenary but a merc verdict", class)
			}
		}
		// the war staff is refused by the hands rule while the shield is worn: the item itself is fine
		if r := v.pack(3).Usable.Hero.Reasons; len(r) == 0 || !strings.Contains(r[0].Text, "two-handed") {
			t.Errorf("%v: staff reasons %+v", class, r)
		}
	}
}

func (u *uworld) w() *gworld { return u.gworld }

func TestUsableReasonsMercWornAndNonEquipment(t *testing.T) {
	d, u := newUWorld(t, d2.Paladin)
	kit(u.w(), d2.Paladin)
	u.heroReq[4] = []itemreq.Reason{str(47)} // the cap in the pack
	u.heroReq[1] = []itemreq.Reason{str(40)} // the sword worn: a stat change put it under its requirement
	u.heroReq[7] = []itemreq.Reason{str(33)} // the club in the stash
	u.mercReq[3] = []itemreq.Reason{{Code: itemreq.MercType, Text: "a desert cannot use that"}}
	u.broken[5] = true
	u.w().pack = append(u.w().pack, u.w().item("potion", 8, "Potion", false, "", d2.Paladin))
	u.merc = &status.MercInv{Hired: true, Alive: true, Class: merc.ClassDesert, Level: 14, Str: 60, Dex: 40, GearKnown: true,
		Equipment: map[string]*status.Item{"mrhand": {ID: 500, Code: "sword", Quality: "magic", Identified: true}}}
	v := d.View()

	cap := v.pack(4)
	if cap.Usable.Hero.OK || cap.Usable.Hero.Reasons[0].Need != 47 || cap.Usable.Hero.Reasons[0].Have != 30 {
		t.Errorf("cap: %+v", cap.Usable.Hero)
	}
	if !cap.Usable.Merc.OK || len(cap.Usable.Merc.Reasons) != 0 {
		t.Errorf("cap for the merc: %+v", cap.Usable.Merc)
	}
	if st := v.pack(3).Usable.Merc; st.OK || st.Reasons[0].Code != itemreq.MercType {
		t.Errorf("staff for the merc: %+v", st)
	}
	if w := v.Equipment["rhand"].Usable.Hero; w.OK || w.Reasons[0].Stat != "str" {
		t.Errorf("a worn item under its requirement must be reported: %+v", w)
	}
	if s := v.Stash[0].Usable.Hero; s.OK {
		t.Errorf("stash item: %+v", s)
	}
	// broken is a warning: the game still lets it be worn, so ok does not change
	ring := v.pack(5).Usable.Hero
	if !ring.OK || len(ring.Reasons) != 1 || ring.Reasons[0].Code != itemreq.Broken || !ring.Reasons[0].Warn {
		t.Errorf("broken ring: %+v", ring)
	}
	if it := v.pack(8); it.Usable != nil || it.Req != nil {
		t.Errorf("a potion has no verdict: %+v", it)
	}
	// his own gear: only his side
	mw := v.Merc.Equipment["mrhand"]
	if mw == nil || mw.Usable == nil || mw.Usable.Hero != nil || mw.Usable.Merc == nil || !mw.Usable.Merc.OK {
		t.Errorf("merc worn: %+v", mw)
	}
	// a dead mercenary: every item reads dead for him
	u.merc.Alive = false
	if m := d.View().pack(4).Usable.Merc; m.OK || m.Reasons[0].Code != itemreq.Dead {
		t.Errorf("dead merc: %+v", m)
	}
}

// Old readers: the new fields are omitted when there is nothing to say, and the schema number moved.
func TestUsableSchemaAndOmission(t *testing.T) {
	d, u := newUWorld(t, d2.Sorceress)
	kit(u.w(), d2.Sorceress)
	v := d.View()
	if v.Schema < 8 { // 8 added req and usable; later schemas only add
		t.Errorf("schema %d", v.Schema)
	}
	b, _ := json.Marshal(ItemView{})
	if strings.Contains(string(b), "usable") || strings.Contains(string(b), `"req"`) {
		t.Errorf("empty item view: %s", b)
	}
}

// A set that names an item under its requirement: Apply skips it with the reason and puts nothing in its place.
func TestSetApplySkipsItemsUnderTheirRequirement(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		d, u := newUWorld(t, class)
		kit(u.w(), class)
		u.w().worn["head"] = u.w().item("helm", 8, "Great Helm", false, "head", class)
		gs, err := d.SaveSet("Full", "")
		if err != nil {
			t.Fatal(err)
		}
		// the great helm goes to the stash and the player's strength dropped under what the club in the stash needs
		u.w().remove(8)
		u.w().stash = append(u.w().stash, u.w().item("helm", 8, "Great Helm", false, "head", class))
		u.heroReq[8] = []itemreq.Reason{str(47)}
		pv, err := d.PreviewSet(gs.ID)
		if err != nil {
			t.Fatal(err)
		}
		var head SetSlotPreview
		for _, sp := range pv.Slots {
			if sp.Slot == "head" {
				head = sp
			}
		}
		if head.Status != "blocked" || len(head.Reasons) != 1 || head.Reasons[0].Need != 47 || !strings.Contains(head.Why, "47 strength") {
			t.Fatalf("%v: head slot: %+v", class, head)
		}
		for _, st := range pv.Steps {
			if st.Slot == "head" {
				t.Errorf("%v: a step for the blocked slot: %+v", class, st)
			}
		}
		// the Cap in the pack must not take its place
		if len(u.w().pack) < 2 || u.w().pack[1].name != "Cap" {
			t.Fatal("setup")
		}
		if _, ok := u.w().worn["head"]; ok {
			t.Errorf("%v: something is worn in the blocked slot", class)
		}
	}
}

// Goldens (both classes and a mercenary): the item fields req and usable of the pack, the worn items and his gear.
func TestUsableViewGoldens(t *testing.T) {
	for _, class := range []d2.Class{d2.Paladin, d2.Sorceress} {
		for _, withMerc := range []bool{false, true} {
			d, u := newUWorld(t, class)
			kit(u.w(), class)
			u.heroReq[4] = []itemreq.Reason{str(47), {Code: itemreq.Class, Need: "Paladin", Have: class.String(), Text: "Paladin only"}}
			u.heroReq[1] = []itemreq.Reason{str(40)}
			u.broken[5] = true
			name := "view_" + class.String() + "_nomerc"
			if withMerc {
				name = "view_" + class.String() + "_desert"
				u.mercReq[3] = []itemreq.Reason{{Code: itemreq.MercType, Text: "a desert cannot use that"}}
				u.merc = &status.MercInv{Hired: true, Alive: true, Class: merc.ClassDesert, Level: 14, Str: 60, Dex: 40, GearKnown: true,
					Equipment: map[string]*status.Item{"mrhand": {ID: 500, Code: "spear", Quality: "magic", Identified: true}}}
			}
			v := d.View()
			out := struct {
				Schema    int                  `json:"schema"`
				Equipment map[string]*ItemView `json:"equipment"`
				Backpack  []ItemView           `json:"backpack"`
				Merc      *MercView            `json:"merc,omitempty"`
			}{8, v.Equipment, v.Backpack, v.Merc} // the number these goldens were made at
			b, _ := json.MarshalIndent(out, "", " ")
			b = append(b, '\n')
			path := filepath.Join("..", "watch", "testdata", "itemusable", name+".json")
			if os.Getenv("UPDATE") != "" {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, b, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(want) != string(b) {
				t.Errorf("%s differs from %s (UPDATE=1 rewrites it)\n got: %s", name, path, b)
			}
		}
	}
}
