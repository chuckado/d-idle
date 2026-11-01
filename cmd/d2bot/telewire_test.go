package main

import (
	"testing"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/nav"
)

func TestWireTeleportKeyFollowsFighterBind(t *testing.T) {
	e := &env.Env{}
	nv := nav.New(e)
	cb := combat.New(e, nv)
	cb.SetCatalog(combat.MapCatalog{Skills: map[int]*data.Skill{54: {ID: 54, Key: "Teleport", Name: "Teleport"}}})
	wireTeleport(nv, cb, nil)
	if nv.HasTeleport() {
		t.Fatal("capability before Teleport is bound")
	}
	if nv.TeleportOK != nil {
		t.Fatal("no tables: every level stays allowed")
	}
	cb.SetHotkeys(map[int]string{54: "f2"})
	if !nv.HasTeleport() {
		t.Fatal("capability missing after the bind")
	}
	if !(moverAdapter{nv}).CanTeleport() {
		t.Fatal("task.World.CanTeleport not fed from the navigator")
	}
	cb.SetHotkeys(nil)
	if (moverAdapter{nv}).CanTeleport() {
		t.Fatal("capability stays after the skill is gone")
	}
}

func TestWireTeleportUnknownAreaAllowed(t *testing.T) {
	e := &env.Env{}
	nv := nav.New(e)
	wireTeleport(nv, combat.New(e, nv), &data.DB{})
	if nv.TeleportOK == nil || !nv.TeleportOK(9999) {
		t.Fatal("an area the tables do not know must not block hops")
	}
}
