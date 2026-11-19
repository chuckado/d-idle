package main

import (
	"errors"
	"testing"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/points"
)

func TestPointVerbCapsAndUnknownVerb(t *testing.T) {
	if v := (*pointVerbs)(nil).caps(); v.Stat || v.Skill || v.Why == "" {
		t.Errorf("no client: %+v", v)
	}
	if err := (&pointVerbs{}).note(errors.New("err unknown cmd stat")); err == nil {
		t.Fatal("note must pass the error on")
	}
	pv2 := &pointVerbs{}
	_ = pv2.note(errors.New("err usage: cmd stat <stat> [n]"))
	if v := pv2.caps(); v.Stat || v.Why == "" {
		t.Errorf("after an unknown-verb answer: %+v", v)
	}
	pv3 := &pointVerbs{}
	_ = pv3.note(errors.New("timeout"))
	if pv3.off.Load() != nil {
		t.Error("a timeout is not a missing verb")
	}
}

func TestPointsAreBankedForMeasuredRunsKept(t *testing.T) {
	t.Setenv("D2BOT_RUN", "")
	t.Setenv("D2BOT_GEAR", "")
	if got := withGearPolicy(bot.Config{}); got.Tuning != nil {
		t.Errorf("a normal run keeps the player defaults: %v", got.Tuning)
	}
	t.Setenv("D2BOT_RUN", "g2")
	got := withGearPolicy(bot.Config{Tuning: map[string]float64{pointsDecisionsKey: 0}})
	if got.Tuning[gearDecisionsKey] != 1 || got.Tuning[pointsDecisionsKey] != 0 {
		t.Errorf("a measured run lays the gear key and keeps an explicit points key: %v", got.Tuning)
	}
	if got = withGearPolicy(bot.Config{}); got.Tuning[pointsDecisionsKey] != 1 {
		t.Errorf("a measured run keeps the bot spending: %v", got.Tuning)
	}
}

func TestWornRequirementsTwoClasses(t *testing.T) {
	db := &data.DB{Items: []*data.ItemBase{{Name: "Cap"}, {Name: "Splint Mail", Display: "Splint Mail", ReqStr: 51}, {Name: "Short Sword", ReqStr: 25, ReqDex: 25}}}
	for _, class := range []d2.Class{d2.Sorceress, d2.Paladin} {
		s := &d2.Snapshot{Player: &d2.Player{Class: class}, Inventory: &d2.Inventory{Items: []d2.Unit{
			{Class: 0, Item: &d2.Item{Where: d2.ContEquipment, Body: d2.BodyHead}},
			{Class: 1, Item: &d2.Item{Where: d2.ContEquipment, Body: d2.BodyTorso}},
			{Class: 2, Item: &d2.Item{Where: d2.ContEquipment, Body: d2.BodyRHand}},
		}}}
		g := wornRequirements(s, db)
		if len(g) != 2 || g[0].Slot != "body" || g[0].Str != 51 || g[1].Item != "Short Sword" || g[1].Dex != 25 {
			t.Errorf("%s: %+v", class, g)
		}
	}
	if wornRequirements(nil, db) != nil {
		t.Error("no snapshot")
	}
}

func TestPointsNeedsRow(t *testing.T) {
	if pointsNeeds(nil) != nil {
		t.Error("no desk, no row")
	}
	p := &d2.Player{StatPoints: 2, SkillPoints: 1}
	d := points.NewDesk(points.Config{Player: func() *d2.Player { return p }})
	rows := pointsNeeds(d)
	if len(rows) != 1 || rows[0].Kind != "points" || rows[0].Text != "2 stat points and 1 skill point waiting for you" {
		t.Fatalf("%+v", rows)
	}
	d.SetAutopilot(true)
	if pointsNeeds(d) != nil {
		t.Error("autopilot on: nothing to ask")
	}
	d.SetAutopilot(false)
	p.StatPoints, p.SkillPoints = 0, 0
	if pointsNeeds(d) != nil {
		t.Error("nothing banked, no row")
	}
}
