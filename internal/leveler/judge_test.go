package leveler_test

import (
	"errors"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/itemreq"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/merc"
	"d2lod-bot/internal/town/sim"
)

// The 21 cases of docs/ui/redesign-2026-11-14/I-character-screen.md section 10, bound to real rows of the game tables
// ($D2_MPQ_DIR, skipped without them): each item is looked up by a predicate on the table, never by a typed code, so a
// case follows the data. What the tables or the rule cannot settle is said in the case, not guessed.

var slotsOfBody = []int{d2.BodyHead, d2.BodyNeck, d2.BodyTorso, d2.BodyRHand, d2.BodyLHand, d2.BodyRRing, d2.BodyLRing, d2.BodyBelt, d2.BodyFeet, d2.BodyGloves}

type judgeRig struct {
	t  *testing.T
	r  *sim.Rig
	lv *leveler.Leveler
	db *data.DB
}

func newJudgeRig(t *testing.T) *judgeRig {
	r := sim.NewRig(t)
	return &judgeRig{t: t, r: r, lv: leveler.New(r.Env, leveler.Options{Snapshot: r.W.Snapshot}), db: r.Env.Data}
}

// find returns the first base (table order) that satisfies pred.
func (g *judgeRig) find(what string, pred func(*data.ItemBase) bool) *data.ItemBase {
	g.t.Helper()
	for _, b := range g.db.Items {
		if b != nil && pred(b) {
			g.t.Logf("%s: %s (%s) lvl %d str %d dex %d", what, b.Name, b.Code, b.LevelReq, b.ReqStr, b.ReqDex)
			return b
		}
	}
	g.t.Fatalf("no table row for %s", what)
	return nil
}

func (g *judgeRig) hero(cls d2.Class, lvl, str, dex int) {
	p := &g.r.W.Player
	p.Class, p.Level, p.Strength, p.Dexterity = cls, lvl, str, dex
}

// pack puts a fresh normal quality, identified item of that base in the pack and judges it for the hero; ok is the
// equip path's own answer (some slot takes it).
func (g *judgeRig) pack(b *data.ItemBase) (itemreq.Judgement, bool, d2.Unit) {
	u := g.r.W.Add(b.Code, d2.ContInventory, 0, 0)
	g.r.W.Items[len(g.r.W.Items)-1] = u
	s, _ := g.r.W.Snapshot()
	ok := false
	for _, body := range slotsOfBody {
		if _, err := g.lv.CheckPlayerEquip(s, u.ID, body); err == nil {
			ok = true
		}
	}
	j := g.lv.Judge(s, &u, itemreq.Item{Code: b.Code, Identified: true}, nil)
	return j, ok, u
}

func codes(rs []itemreq.Reason) []string {
	var out []string
	for _, r := range rs {
		out = append(out, r.Code+"/"+r.Stat)
	}
	return out
}

func has(rs []itemreq.Reason, code, stat string) *itemreq.Reason {
	for i := range rs {
		if rs[i].Code == code && rs[i].Stat == stat {
			return &rs[i]
		}
	}
	return nil
}

func typeIs(g *judgeRig, b *data.ItemBase, t string) bool { return g.db.IsA(b.Type, t) }

func plain(_ *judgeRig, b *data.ItemBase) bool {
	return b.Code != "" && !b.Quest && b.Spawnable
}

func TestJudgeHeroCases(t *testing.T) {
	g := newJudgeRig(t)
	noClass := func(b *data.ItemBase) bool { return plain(g, b) && classOfType(g, b) == "" }

	// 1: Sorceress 7, Str 38: a one-handed sword she meets
	g.hero(d2.Sorceress, 7, 38, 20)
	sword := g.find("1: sword", func(b *data.ItemBase) bool {
		return noClass(b) && b.Kind == data.KindWeapon && typeIs(g, b, "swor") && !b.TwoHanded && b.ReqStr > 0 && b.ReqStr <= 38 && b.ReqDex <= 20 && b.LevelReq <= 7
	})
	if j, ok, _ := g.pack(sword); !ok || len(j.Hero) != 0 || !j.Wearable {
		t.Errorf("1: ok=%v reasons=%v", ok, codes(j.Hero))
	}

	// 2: Str unmet, with numbers
	big := g.find("2: strength above 38", func(b *data.ItemBase) bool {
		return noClass(b) && b.Kind == data.KindWeapon && typeIs(g, b, "swor") && !b.TwoHanded && b.ReqStr > 38 && b.ReqDex <= 20 && b.LevelReq <= 7
	})
	j, ok, _ := g.pack(big)
	r := has(j.Hero, itemreq.Requirements, "str")
	if ok || r == nil || r.Need != big.ReqStr || r.Have != 38 || r.Text == "" {
		t.Fatalf("2: ok=%v reasons=%v %+v", ok, codes(j.Hero), r)
	}
	t.Logf("2: %q", r.Text)

	// 3: level unmet
	lv := g.find("3: level above 7", func(b *data.ItemBase) bool {
		return noClass(b) && b.Kind == data.KindArmor && typeIs(g, b, "tors") && b.LevelReq > 7 && b.ReqStr <= 38
	})
	if j, ok, _ := g.pack(lv); ok || has(j.Hero, itemreq.Requirements, "lvl") == nil || j.Req.Lvl != lv.LevelReq {
		t.Errorf("3: ok=%v reasons=%v req %+v", ok, codes(j.Hero), j.Req)
	}

	// 4, 5, 6: class items
	pshield := g.find("4: Paladin shield", func(b *data.ItemBase) bool {
		return b.Spawnable && typeIs(g, b, "shld") && classOfType(g, b) == "pal" && b.ReqStr <= 38 && b.LevelReq <= 7
	})
	orb := g.find("6: Sorceress orb", func(b *data.ItemBase) bool {
		return b.Spawnable && b.Kind == data.KindWeapon && classOfType(g, b) == "sor" && b.ReqStr <= 38 && b.LevelReq <= 7
	})
	for _, cls := range []d2.Class{d2.Sorceress, d2.Paladin} {
		g.hero(cls, 7, 38, 20)
		for _, c := range []struct {
			b     *data.ItemBase
			own   d2.Class
			case_ string
		}{{pshield, d2.Paladin, "4/5"}, {orb, d2.Sorceress, "6"}} {
			j, ok, _ := g.pack(c.b)
			if cls == c.own {
				if !ok || len(j.Hero) != 0 {
					t.Errorf("%s: %v wearing own class item: ok=%v %v", c.case_, cls, ok, codes(j.Hero))
				}
				continue
			}
			rc := has(j.Hero, itemreq.Class, "")
			if ok || rc == nil || rc.Have != cls.String() || len(j.Req.Class) != 1 {
				t.Errorf("%s: %v with the other class's item: ok=%v %v req %+v", c.case_, cls, ok, codes(j.Hero), j.Req)
			}
		}
	}

	// 7: Amazon, a bow with Dex above hers. The numbers are reported either way; whether the equip path takes a bow for
	// an Amazon at all is read from the rule below (see the gap note).
	g.hero(d2.Amazon, 20, 40, 25)
	bow := g.find("7: bow, dex above 25", func(b *data.ItemBase) bool {
		return noClass(b) && b.Kind == data.KindWeapon && typeIs(g, b, "bow") && b.ReqDex > 25 && b.LevelReq <= 20
	})
	j, ok, _ = g.pack(bow)
	if r := has(j.Hero, itemreq.Requirements, "dex"); ok || r == nil || r.Need != bow.ReqDex || r.Have != 25 {
		t.Errorf("7: ok=%v %v", ok, codes(j.Hero))
	}
	if r := has(j.Hero, itemreq.Slot, ""); r != nil {
		t.Logf("7: GAP (V in the tables, the equip rule is unchanged): slotsFor refuses every missile weapon for every class, so even a bow an Amazon meets reads %q", r.Text)
	}

	// 8: Barbarian primal helm; a Druid cannot wear it
	phlm := g.find("8: primal helm", func(b *data.ItemBase) bool {
		return b.Spawnable && typeIs(g, b, "helm") && classOfType(g, b) == "bar" && b.ReqStr <= 38 && b.LevelReq <= 7
	})
	g.hero(d2.Barbarian, 7, 38, 20)
	if j, ok, _ := g.pack(phlm); !ok || len(j.Hero) != 0 {
		t.Errorf("8 barbarian: ok=%v %v", ok, codes(j.Hero))
	}
	g.hero(d2.Druid, 7, 38, 20)
	if j, ok, _ := g.pack(phlm); ok || has(j.Hero, itemreq.Class, "") == nil {
		t.Errorf("8 druid: ok=%v %v", ok, codes(j.Hero))
	}

	// 9: Necromancer head, Druid pelt, Assassin claw, each on its class and on another
	for _, c := range []struct {
		name string
		own  d2.Class
		typ  string
	}{{"head", d2.Necromancer, "head"}, {"pelt", d2.Druid, "pelt"}, {"claw", d2.Assassin, "h2h"}} {
		b := g.find("9: "+c.name, func(b *data.ItemBase) bool {
			return b.Spawnable && typeIs(g, b, c.typ) && classOfType(g, b) != "" && b.ReqStr <= 38 && b.ReqDex <= 20 && b.LevelReq <= 7
		})
		g.hero(c.own, 7, 38, 20)
		if j, ok, _ := g.pack(b); !ok || len(j.Hero) != 0 {
			t.Errorf("9 %s on its class: ok=%v %v", c.name, ok, codes(j.Hero))
		}
		g.hero(d2.Sorceress, 7, 38, 20)
		if j, ok, _ := g.pack(b); ok || has(j.Hero, itemreq.Class, "") == nil {
			t.Errorf("9 %s on another class: ok=%v %v", c.name, ok, codes(j.Hero))
		}
	}

	// 10: an unidentified magic ring
	g.hero(d2.Sorceress, 7, 38, 20)
	ring := g.find("10: ring", func(b *data.ItemBase) bool { return b.Spawnable && typeIs(g, b, "ring") })
	u := g.r.W.Add(ring.Code, d2.ContInventory, 0, 0)
	u.Item.Quality, u.Item.Flags = d2.QualityMagic, 0
	g.r.W.Items[len(g.r.W.Items)-1] = u
	s, _ := g.r.W.Snapshot()
	jr := g.lv.Judge(s, &u, itemreq.Item{Code: ring.Code}, nil)
	if has(jr.Hero, itemreq.Identified, "") == nil {
		t.Errorf("10: %v", codes(jr.Hero))
	}
	if _, err := g.lv.CheckPlayerEquip(s, u.ID, d2.BodyRRing); err == nil {
		t.Error("10: the equip path takes an unidentified ring")
	}

	g.hero(d2.Sorceress, 7, 38, 30)
	// 12: a two-hander while a shield is worn: the item is fine, the hands rule refuses, with its own code
	two := g.find("12: two-handed sword", func(b *data.ItemBase) bool {
		return noClass(b) && b.Kind == data.KindWeapon && typeIs(g, b, "swor") && b.TwoHanded && b.ReqStr <= 38 && b.LevelReq <= 7
	})
	sh := g.find("12: shield", func(b *data.ItemBase) bool {
		return noClass(b) && typeIs(g, b, "shld") && b.ReqStr <= 38 && b.LevelReq <= 7
	})
	worn := g.r.W.Add(sh.Code, d2.ContEquipment, 0, 0)
	worn.Item.Body = d2.BodyLHand
	g.r.W.Items[len(g.r.W.Items)-1] = worn
	j2, _, u2 := g.pack(two)
	s, _ = g.r.W.Snapshot()
	_, err := g.lv.CheckPlayerEquip(s, u2.ID, d2.BodyRHand)
	var ie *itemreq.Error
	if len(j2.Hero) != 0 || !errors.Is(err, leveler.ErrOffHand) || !errors.As(err, &ie) || ie.Reasons[0].Code != itemreq.TwoHand {
		t.Errorf("12: item reasons %v, err %v", codes(j2.Hero), err)
	}
	t.Log("12: nospace (a full pack) is not part of the equip path today (U): the view cannot report it")

	// 15 javelin: U in the doc, refused for the Desert Guard
	jav := g.find("15: javelin", func(b *data.ItemBase) bool { return b.Spawnable && b.Kind == data.KindWeapon && typeIs(g, b, "jave") })
	if has(judgeMerc(g, s, jav, itemreq.Merc{Class: merc.ClassDesert, Level: 14, Str: 60, Dex: 40, Alive: true}), itemreq.MercType, "") == nil {
		t.Error("15 javelin: no merctype")
	}

	// 21: an ethereal item: the requirement is reported as the base number and marked approximate (U until a probe)
	g.hero(d2.Sorceress, 7, 38, 20)
	eth := g.r.W.Add(sword.Code, d2.ContInventory, 1, 0)
	eth.Item.Flags |= d2.FlagEthereal
	g.r.W.Items[len(g.r.W.Items)-1] = eth
	s, _ = g.r.W.Snapshot()
	if je := g.lv.Judge(s, &eth, itemreq.Item{Code: sword.Code, Identified: true}, nil); !je.Req.Approx || je.Req.Str != sword.ReqStr {
		t.Errorf("21: %+v", je.Req)
	}
}

func judgeMerc(g *judgeRig, s *d2.Snapshot, b *data.ItemBase, m itemreq.Merc) []itemreq.Reason {
	return g.lv.Judge(s, nil, itemreq.Item{Code: b.Code, Identified: true}, &m).Merc
}

func classOfType(g *judgeRig, b *data.ItemBase) string {
	seen := map[string]bool{}
	var walk func(string) string
	walk = func(typ string) string {
		if typ == "" || seen[typ] {
			return ""
		}
		seen[typ] = true
		tt := g.db.ItemType(typ)
		if tt == nil {
			return ""
		}
		if tt.Class != "" {
			return tt.Class
		}
		for _, e := range tt.Equiv {
			if c := walk(e); c != "" {
				return c
			}
		}
		return ""
	}
	return walk(b.Type)
}

func TestJudgeMercCases(t *testing.T) {
	g := newJudgeRig(t)
	g.hero(d2.Sorceress, 30, 60, 60)
	s, _ := g.r.W.Snapshot()
	judge := func(b *data.ItemBase, m itemreq.Merc) []itemreq.Reason {
		j := g.lv.Judge(s, nil, itemreq.Item{Code: b.Code, Identified: true}, &m)
		if !j.Wearable {
			t.Fatalf("%s is not wearable", b.Code)
		}
		return j.Merc
	}
	free := func(b *data.ItemBase) bool { return b.Spawnable && classOfType(g, b) == "" }

	// 15: Desert Guard 14, Str 60: a spear he meets, a polearm above his Str, a bow
	dg := itemreq.Merc{Class: merc.ClassDesert, Level: 14, Str: 60, Dex: 40, Alive: true}
	spear := g.find("15: spear", func(b *data.ItemBase) bool {
		return free(b) && b.Kind == data.KindWeapon && typeIs(g, b, "spea") && !b.Stackable && b.ReqStr <= 60 && b.LevelReq <= 14
	})
	pole := g.find("15: polearm above his strength", func(b *data.ItemBase) bool {
		return free(b) && b.Kind == data.KindWeapon && typeIs(g, b, "pole") && b.ReqStr > 60 && b.LevelReq <= 14
	})
	bow := g.find("15: bow", func(b *data.ItemBase) bool {
		return free(b) && b.Kind == data.KindWeapon && typeIs(g, b, "bow") && !typeIs(g, b, "xbow")
	})
	if rs := judge(spear, dg); len(rs) != 0 {
		t.Errorf("15 spear: %v", codes(rs))
	}
	if r := has(judge(pole, dg), itemreq.Requirements, "str"); r == nil || r.Need != pole.ReqStr || r.Have != 60 {
		t.Errorf("15 polearm: %+v", r)
	}
	if has(judge(bow, dg), itemreq.MercType, "") == nil {
		t.Error("15 bow: no merctype")
	}

	// 16: Rogue: a bow, not a crossbow
	rg := itemreq.Merc{Class: merc.ClassRogue, Level: 14, Str: 40, Dex: 60, Alive: true}
	xbow := g.find("16: crossbow", func(b *data.ItemBase) bool { return free(b) && b.Kind == data.KindWeapon && typeIs(g, b, "xbow") })
	rbow := g.find("16: bow", func(b *data.ItemBase) bool {
		return free(b) && b.Kind == data.KindWeapon && typeIs(g, b, "bow") && !typeIs(g, b, "xbow") && b.ReqDex <= 60 && b.ReqStr <= 40 && b.LevelReq <= 14
	})
	if rs := judge(rbow, rg); len(rs) != 0 {
		t.Errorf("16 bow: %v", codes(rs))
	}
	if has(judge(xbow, rg), itemreq.MercType, "") == nil {
		t.Error("16 crossbow: no merctype")
	}

	// 17: Iron Wolf: a one-handed sword and a shield, not a two-handed sword
	iw := itemreq.Merc{Class: merc.ClassIronWolf, Level: 20, Str: 80, Dex: 60, Alive: true}
	sw1 := g.find("17: one-handed sword", func(b *data.ItemBase) bool {
		return free(b) && b.Kind == data.KindWeapon && typeIs(g, b, "swor") && !b.TwoHanded && b.ReqStr <= 80 && b.LevelReq <= 20
	})
	sw2 := g.find("17: two-handed sword", func(b *data.ItemBase) bool {
		return free(b) && b.Kind == data.KindWeapon && typeIs(g, b, "swor") && b.TwoHanded && b.ReqStr <= 80
	})
	shield := g.find("17: shield", func(b *data.ItemBase) bool {
		return free(b) && typeIs(g, b, "shld") && b.ReqStr <= 80 && b.LevelReq <= 20
	})
	if rs := judge(sw1, iw); len(rs) != 0 {
		t.Errorf("17 sword: %v", codes(rs))
	}
	if rs := judge(shield, iw); len(rs) != 0 {
		t.Errorf("17 shield: %v", codes(rs))
	}
	if has(judge(sw2, iw), itemreq.MercType, "") == nil {
		t.Error("17 two-handed sword (U in the doc, refused): no merctype")
	}
	// the other types carry no shield
	if has(judge(shield, dg), itemreq.MercType, "") == nil {
		t.Error("17 shield on a Desert Guard: no merctype")
	}

	// 18: Barbarian mercenary: a sword and a primal helm
	bm := itemreq.Merc{Class: merc.ClassBarb, Level: 30, Str: 100, Dex: 60, Alive: true}
	phlm := g.find("18: primal helm", func(b *data.ItemBase) bool {
		return b.Spawnable && typeIs(g, b, "helm") && classOfType(g, b) == "bar" && b.ReqStr <= 100 && b.LevelReq <= 30
	})
	if rs := judge(sw1, bm); len(rs) != 0 {
		t.Errorf("18 sword: %v", codes(rs))
	}
	if rs := judge(phlm, bm); len(rs) != 0 {
		t.Errorf("18 primal helm: %v", codes(rs))
	}

	// 19: a dead mercenary
	dead := dg
	dead.Alive = false
	if has(judge(spear, dead), itemreq.Dead, "") == nil {
		t.Error("19: no dead reason")
	}
}

// One rule, two views, over every table row: with nothing worn, an item has no hero reason exactly when some body slot
// takes it, for two classes and a low and a high character.
func TestJudgeAgreesWithTheEquipCheck(t *testing.T) {
	g := newJudgeRig(t)
	for _, cls := range []d2.Class{d2.Sorceress, d2.Paladin} {
		for _, st := range []struct{ lvl, str, dex int }{{1, 10, 10}, {90, 150, 150}, {20, 60, 40}} {
			g.hero(cls, st.lvl, st.str, st.dex)
			n := 0
			for _, b := range g.db.Items {
				if b == nil || b.Code == "" || !(b.Kind == data.KindWeapon || b.Kind == data.KindArmor) {
					continue
				}
				j, ok, _ := g.pack(b)
				if (len(j.Hero) == 0) != ok {
					t.Errorf("%v %+v %s: reasons %v but ok=%v", cls, st, b.Code, codes(j.Hero), ok)
				}
				n++
				g.r.W.Items = g.r.W.Items[:0] // keep the world small
			}
			if n == 0 {
				t.Fatal("no rows")
			}
		}
	}
}
