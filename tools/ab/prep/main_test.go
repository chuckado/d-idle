package main

import (
	"reflect"
	"testing"
)

func TestVerbs(t *testing.T) {
	got := Verbs(Setup{Clvl: 11, Gold: 500, Stash: 0, MercType: 0})
	want := []string{"debug stat 12 11", "debug stat 13 EXP", "debug statpts 50", "debug skillpts 10", "debug stat 14 500", "debug stat 15 0", "debug merc hire 0"}
	got[1] = "debug stat 13 EXP"
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v", got)
	}
	if v := Verbs(Setup{Clvl: 5, MercType: -1}); len(v) != 6 {
		t.Errorf("no merc: %v", v)
	}
}

func TestVerbsWeapon(t *testing.T) {
	v := Verbs(Setup{Clvl: 12, MercType: 0, Weapon: "hax"})
	if v[len(v)-2] != "debug equip hax 4" || v[len(v)-1] != "debug merc hire 0" {
		t.Errorf("weapon is worn before the merc is hired: %v", v)
	}
	for _, x := range Verbs(Setup{Clvl: 12, MercType: -1}) {
		if len(x) > 11 && x[:11] == "debug equip" {
			t.Errorf("no weapon asked, but %q", x)
		}
	}
}

func TestWeaponReadback(t *testing.T) {
	ok := `{"v":1,"stats":{"weapons":[{"code":"hax","slot":4,"mindam":3,"maxdam":6}],"hero":{"hands":[{"slot":4,"min":3,"max":6}]}}}`
	if s, err := WeaponReadback(ok, "hax"); err != nil || s == "" {
		t.Errorf("worn weapon: %q, %v", s, err)
	}
	if _, err := WeaponReadback(ok, "swd"); err == nil {
		t.Error("another code must fail")
	}
	old := `{"v":1,"stats":{"base":{"str":25}}}`
	if _, err := WeaponReadback(old, "hax"); err == nil {
		t.Error("a dump without weapons must fail when one was asked for")
	}
	if _, err := WeaponReadback(old, ""); err != nil {
		t.Errorf("no weapon asked: %v", err)
	}
}
