package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"d2lod-bot/internal/boss"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dctl/fake"
	"d2lod-bot/internal/d2dworld"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/nav"
)

// On d2d Env.World is nil: the boss presets must come from the navigator's world, so the fork's monster preset (the
// Countess's spawn point, read through the wire with the knob on) reaches the boss script; knob off or verb absent
// leaves it out (today's behaviour).
func TestBossPresetsOnD2dCarrySuperuniqueSpawn(t *testing.T) {
	for _, c := range []struct {
		name       string
		knob, verb bool
		want       bool
	}{{"knob on", true, true, true}, {"knob off", false, true, false}, {"verb absent", true, false, false}} {
		t.Run(c.name, func(t *testing.T) {
			path, cleanup, err := fake.TempSocket()
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()
			fix := fake.Sample()
			fix.Level = "25 30.70 13.70 57 41 0 0"
			if c.verb {
				doc, _ := json.Marshal(&d2dctl.LevelMonsters{Level: 25, Unit: "cell", Presets: []d2dctl.MonsterPreset{
					{Index: 2, Kind: "superunique", ID: 740, Super: 6, Name: "The Countess", X: 21.5, Y: 8.5, Room: 4}}})
				fix.LevelMonsters = string(doc)
			}
			srv, err := fake.New(path, fix, fake.Options{})
			if err != nil {
				t.Fatal(err)
			}
			defer srv.Close()
			cl := d2dctl.New(path, d2dctl.Options{Timeout: 2 * time.Second})
			defer cl.Close()
			w := d2dworld.New(d2dworld.Client{Client: cl}, d2dworld.Options{LayoutRead: func() bool { return c.knob }})
			nv := nav.New(&env.Env{})
			nv.UseWorld(w)
			g := &bossGame{e: &env.Env{Log: slog.New(slog.NewTextHandler(io.Discard, nil))}, nv: nv, logPresets: map[int]int{}}
			ps, err := g.Presets(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			got := false
			for _, p := range ps {
				got = got || (p.Kind == boss.PresetNPC && p.ID == 740 && p.Name == "The Countess")
			}
			if got != c.want {
				t.Fatalf("superunique preset present=%v want %v: %+v", got, c.want, ps)
			}
		})
	}
}
