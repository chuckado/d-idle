package loot

import (
	"reflect"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/lootrules"
)

func TestAdapterItemFacts(t *testing.T) {
	l, db := synthLooter(t)
	a := l.Rules()
	a.classes = map[string]string{"ama": "Amazon", "dru": "Druid"} // the synthetic charstats rows carry no class code

	// An unidentified magic Crystal Sword in the pack, stats unreadable.
	u := synthUnit(t, db, 1, "crs", d2.QualityMagic, 0, d2.ContInventory)
	it, err := a.Item(u, ItemOpts{Place: lootrules.PlacePack})
	if err != nil {
		t.Fatal(err)
	}
	if it.Code != "crs" || it.Quality != "magic" || it.ILvl != 20 || it.ReqStr != 43 || it.ReqLvl != 25 || it.BaseTier != "exceptional" || it.MaxSockets != 5 {
		t.Errorf("facts %+v", it)
	}
	if !reflect.DeepEqual(it.Types, []string{"swor", "mele", "weap"}) {
		t.Errorf("types %v", it.Types)
	}
	if !reflect.DeepEqual(it.MercTypes, []string{"barbarian", "ironwolf"}) {
		t.Errorf("merc types %v", it.MercTypes)
	}
	if it.Hidden&lootrules.HStats == 0 || it.Identified || it.Stats != nil {
		t.Errorf("unreadable stats must be flagged: hidden %b stats %v", it.Hidden, it.Stats)
	}

	// The same item as a d2d ground label: no item level, no identified flag, a magic value is only a bound.
	g, _ := a.Item(u, ItemOpts{Place: lootrules.PlaceGround, LabelOnly: true})
	if g.ILvl != 0 || g.Hidden&(lootrules.HILvl|lootrules.HIdentified|lootrules.HValue) != lootrules.HILvl|lootrules.HIdentified|lootrules.HValue {
		t.Errorf("label-only ground: ilvl %d hidden %b", g.ILvl, g.Hidden)
	}
	w := synthUnit(t, db, 2, "sst", d2.QualityNormal, 0, d2.ContGround)
	if g, _ = a.Item(w, ItemOpts{Place: lootrules.PlaceGround, LabelOnly: true}); g.Hidden&lootrules.HValue != 0 || g.Value <= 0 || g.BaseTier != "normal" {
		t.Errorf("white ground item keeps its base value: %+v", g)
	}

	// Stats readable: values shifted, sockets, durability.
	u.Item.Flags = d2.FlagIdentified
	st := d2.Stats{7: 25 << 8, 39: 12, 194: 3, 73: 30, 72: 21}
	r, _ := a.Item(u, ItemOpts{Place: lootrules.PlacePack, Stats: st, Mark: "keep"})
	if r.Stats["life"] != 25 || r.Stats["fireres"] != 12 || r.Sockets != 3 || r.MaxDur != 30 || r.Durability != 21 || r.Mark != "keep" || r.Hidden&lootrules.HStats != 0 {
		t.Errorf("stats %+v hidden %b", r, r.Hidden)
	}
}

func TestAdapterBaseDerivedFields(t *testing.T) {
	l, db := synthLooter(t)
	a := l.Rules()
	a.classes = map[string]string{"ama": "Amazon", "dru": "Druid"}
	get := func(code string) lootrules.Item {
		it, err := a.Item(synthUnit(t, db, 9, code, d2.QualityNormal, d2.FlagIdentified, d2.ContInventory), ItemOpts{Place: lootrules.PlacePack, Stats: d2.Stats{}})
		if err != nil {
			t.Fatal(err)
		}
		return it
	}
	if get("r33").RuneNo != 33 || get("r01").RuneNo != 1 || get("rin").RuneNo != 0 {
		t.Error("rune numbers")
	}
	if get("skc").Grade != 1 || get("skz").Grade != 5 || get("rin").Grade != 0 {
		t.Error("skull grades")
	}
	if c := get("ams").ClassOnly; c != "Amazon" {
		t.Errorf("amazon javelin class %q", c)
	}
	if c := get("dr1").ClassOnly; c != "Druid" || len(get("dr1").MercTypes) != 0 {
		t.Errorf("pelt class %q merc %v", c, get("dr1").MercTypes)
	}
	if got := get("spr").MercTypes; !reflect.DeepEqual(got, []string{"desert"}) {
		t.Errorf("spear %v", got)
	}
	if got := get("sbw").MercTypes; !reflect.DeepEqual(got, []string{"rogue"}) {
		t.Errorf("bow %v", got)
	}
	if got := get("clm").MercTypes; len(got) != 0 {
		t.Errorf("two-handed sword %v", got)
	}
	if !get("jav").Stackable {
		t.Error("javelin stack")
	}
	if _, err := a.Item(d2.Unit{ID: 3, Class: 9999, Item: &d2.Item{}}, ItemOpts{}); err == nil {
		t.Error("unknown base must be an error")
	}
	bad := d2.Unit{ID: 4, Class: 0}
	if _, err := a.Item(bad, ItemOpts{}); err == nil {
		t.Error("a unit without item data must be an error")
	}
}

func TestAdapterActorsTwoHeroesAndMerc(t *testing.T) {
	l, _ := synthLooter(t)
	a := l.Rules()
	for _, hc := range []struct {
		class d2.Class
		name  string
	}{{d2.Amazon, "Amazon"}, {d2.Sorceress, "Sorceress"}} {
		s := &d2.Snapshot{Player: &d2.Player{Class: hc.class, Level: 12, Strength: 30, Dexterity: 40}}
		h, m := a.ActorsOf(s)
		if h.Class != hc.name || h.Level != 12 || h.Str != 30 || h.Dex != 40 || m != nil {
			t.Errorf("%s: hero %+v merc %+v", hc.name, h, m)
		}
		s.Merc = &d2.MercSheet{Hired: true, Alive: true, Level: 9, Str: 45, Dex: 28, Type: 2}
		if _, m = a.ActorsOf(s); m == nil || m.Class != "desert" || m.Level != 9 || m.Str != 45 {
			t.Errorf("merc %+v", m)
		}
		s.Merc = &d2.MercSheet{Hired: true, Type: 1} // stats not read yet
		if _, m = a.ActorsOf(s); m == nil || m.Class != "rogue" || m.Level != 0 {
			t.Errorf("unread merc must carry level 0: %+v", m)
		}
		s.Merc = &d2.MercSheet{Hired: true, Level: 5, Str: 1, Dex: 1, Type: 77} // not in Hireling.txt
		if _, m = a.ActorsOf(s); m == nil || m.Class != "" || m.Level != 0 {
			t.Errorf("unknown hireling type must stay unknown: %+v", m)
		}
	}
}

func TestAdapterCtxNeedsAndOwned(t *testing.T) {
	l, db := synthLooter(t)
	a := l.Rules()
	book := synthUnit(t, db, 5, "tbk", d2.QualityNormal, d2.FlagIdentified, d2.ContInventory)
	s := &d2.Snapshot{State: d2.StateInGame, Player: &d2.Player{Class: d2.Amazon, Level: 5, Gold: 120, Strength: 20, Dexterity: 25, Area: 1},
		Inventory: &d2.Inventory{Items: []d2.Unit{book}}}
	cx := a.Ctx(s, "javazon", 41, 7)
	if cx.Class != "Amazon" || cx.Build != "javazon" || cx.Difficulty != "Normal" || cx.CLvl != 5 || cx.Gold != 120 || cx.Act != 1 || cx.PackFullPct != 41 || cx.StashFullPct != 7 {
		t.Errorf("ctx %+v", cx)
	}
	if cx.Hero == nil || cx.Hero.Class != "Amazon" {
		t.Error("hero actor")
	}
	tome := &lootrules.Item{Code: "tbk"}
	if cx.Owned(tome) != 1 || cx.Owned(&lootrules.Item{Code: "isc"}) != 0 {
		t.Error("owned counts the pack and the stash")
	}
	if cx.Needs("tome", tome) {
		t.Error("a held tome is not needed")
	}
	if !cx.Needs("tome", &lootrules.Item{Code: "ibk"}) && db.ItemByCode("ibk") != nil {
		t.Error("an unheld tome is needed")
	}
	if cx.Needs("potion", &lootrules.Item{Code: "nope"}) {
		t.Error("an unknown base is never needed")
	}
	// naked hero: a usable weapon fills the empty slot
	if !cx.Needs("emptyslot", &lootrules.Item{Code: "sst"}) || cx.Needs("emptyslot", &lootrules.Item{Code: "rin"}) {
		t.Error("empty slot need")
	}
}
