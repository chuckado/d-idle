package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/points"
	"d2lod-bot/internal/watch"
)

// mountBuilds serves the Points page and its Build tab (/watch/points, ?tab=build) from the real points desk over a fixture
// Paladin: the cards and details come from the shipped build catalogue (internal/plandata), the build files are a few of its
// real plan table names with an experimental one, and auto-allocation is on. No real hero data. Card buildcards.
func mountBuilds(m *http.ServeMux) {
	p := d2.Player{Class: d2.Paladin, Level: 14, Strength: 30, Vitality: 40, StatPoints: 3, SkillPoints: 1}
	tb := points.Tables{
		Classes: []*data.CharClass{{Name: "Paladin", Code: "pal", Str: 25, LifePerVit: 12, ManaPerEnergy: 6}},
		Skills:  []*data.Skill{{ID: 96, Key: "Might", Name: "Might", Class: "pal", Page: 1, ReqLevel: 1, MaxLevel: 20, Aura: true}},
	}
	row := func(n int, s string, pts int) points.PreviewRow {
		return points.PreviewRow{Clvl: n, Skills: []points.PreviewSkill{{Skill: s, Points: pts}}}
	}
	dir, _ := os.MkdirTemp("", "uishots-points-")
	newDesk := func(hero string) *points.Desk {
		// a locked build that took over its later table (the allocator handoff line, cards #1155 and #1176): the Build tab and the Points page
		// then carry the long plan sentence and the spent / unspent / waiting text
		if hero == "Locked" {
			os.WriteFile(filepath.Join(dir, "Locked.points.json"), []byte(`{"autopilot":true,"build":"paladin-normal30","locked":true,"lockedAtLevel":12,"plan":"paladin-normal30-v2","planAtLevel":30}`), 0o600)
		}
		return points.NewDesk(points.Config{
			Hero: hero, T: tb, Dir: dir,
			AutoOn: func() bool { return true },
			Builds: func() []points.BuildInfo {
				pre := []points.PreviewRow{row(2, "Might", 1), row(4, "Holy Fire", 1), row(6, "Blessed Hammer", 2)}
				mk := func(name, tier string) points.BuildInfo {
					return points.BuildInfo{Name: name, Label: name, Class: "paladin", Tier: tier, Preview: pre}
				}
				ex := mk("paladin_hammerdin_endgame90", points.BuildExperimental)
				ex.Needs = []points.Need{{Skill: "Concentration", Tier: "unsupported", Reason: "no rule"}}
				return []points.BuildInfo{mk("paladin-normal30", points.BuildSupported), mk("paladin-normal30-v2", points.BuildSupported), mk("paladin_hammerdin_guide", points.BuildSupported), ex, mk("paladin-zealot-lite", points.BuildSupported)}
			},
			Respec: func() points.RespecParams { return points.RespecParams{BaseGold: 1000, PerLevel: 100, FirstFree: 0} },
			Player: func() *d2.Player { q := p; return &q },
			Fresh:  func(context.Context) (*d2.Player, error) { q := p; return &q, nil },
		})
	}
	// the page's query plan=1 (the poll carries it as the Referer) picks the hero whose build is locked on its later plan; every other
	// builds page keeps the unlocked hero (the Choose flow needs it)
	serve := func(hero string) http.Handler {
		api := watch.NewPointsAPI()
		api.Bind(newDesk(hero))
		return api.Handler()
	}
	plain, locked := serve("Fixture"), serve("Locked")
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Referer(), "plan=1") {
			locked.ServeHTTP(w, r)
			return
		}
		plain.ServeHTTP(w, r)
	})
	for _, r := range watch.PointsRoutes {
		m.Handle(r, h)
	}
}
