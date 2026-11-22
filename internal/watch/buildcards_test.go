package watch

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/points"
)

// web/buildcards.js in a stub DOM (node, skipped without it): the small card and the detail view against a text golden
// (testdata/buildcards_dom.txt) built from the real catalogue through the desk, and the one confirm sheet's focus, Tab, Escape
// and press-outside behaviour. UPDATE=1 rewrites the golden.
func TestBuildCardsDOM(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("no node")
	}
	d := points.NewDesk(points.Config{
		Hero: "Testy", Dir: t.TempDir(),
		Builds: func() []points.BuildInfo {
			return []points.BuildInfo{
				{Name: "paladin-normal30", Label: "paladin normal30", Class: "paladin", Tier: points.BuildSupported, Preview: []points.PreviewRow{{Clvl: 2, Skills: []points.PreviewSkill{{Skill: "Might", Points: 1}}}, {Clvl: 6, Skills: []points.PreviewSkill{{Skill: "Blessed Hammer", Points: 2, Tier: "unsupported"}}, Stats: map[string]int{"str": 5}}}},
				{Name: "paladin_hammerdin_endgame90", Label: "paladin hammerdin endgame90", Class: "paladin", Tier: points.BuildExperimental, Needs: []points.Need{{Skill: "Concentration", Tier: "unsupported", Reason: "no rule"}, {Skill: "Vigor", Tier: "unproven"}}},
			}
		},
		Player: func() *d2.Player { return &d2.Player{Class: d2.Paladin, Level: 12} },
	})
	cards, err := d.Cards("Paladin")
	if err != nil || len(cards) == 0 {
		t.Fatalf("cards: %v", err)
	}
	det, err := d.BuildDetail(cards[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if det.DefenseByDifficulty == nil {
		t.Log("note: the first catalogue build has no defenseByDifficulty; the golden does not show that line")
	}
	b, _ := json.Marshal(map[string]any{"card": cards[0], "detail": det})
	dir := t.TempDir()
	js, _ := webFS.ReadFile("web/buildcards.js")
	os.WriteFile(filepath.Join(dir, "data.json"), b, 0o600)
	os.WriteFile(filepath.Join(dir, "buildcards.cjs"), js, 0o600)
	for _, f := range []string{"nav_dom.cjs", "buildcards_test.cjs", "buildcards_dom.txt"} {
		src, err := os.ReadFile(filepath.Join("testdata", f))
		if err != nil && f != "buildcards_dom.txt" {
			t.Fatal(err)
		}
		os.WriteFile(filepath.Join(dir, f), src, 0o600)
	}
	cmd := exec.Command(node, filepath.Join(dir, "buildcards_test.cjs"), dir)
	if os.Getenv("UPDATE") != "" {
		cmd.Env = append(os.Environ(), "UPDATE="+filepath.Join("testdata", "buildcards_dom.txt"))
		cmd.Dir, _ = os.Getwd()
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, out)
	}
	var res struct{ Fails []string }
	if err := json.Unmarshal(out[strings.Index(string(out), "{"):], &res); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, f := range res.Fails {
		t.Error(f)
	}
	if got, err := os.ReadFile(filepath.Join(dir, "buildcards_dom.got")); err == nil {
		t.Logf("got:\n%s", got)
	}
}
