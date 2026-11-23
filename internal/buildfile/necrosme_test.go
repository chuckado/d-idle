package buildfile_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"d2lod-bot/internal/agency"
	"d2lod-bot/internal/buildfile"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/plandata"
)

// necro3: the two Necromancer SME builds are build files (docs/design/necromancer-sme-answers-2026-11-22.md); the
// shipped plan tables are their expansion plus engine data (selector, combat line, tier). UPDATE=1 (re)writes both
// table copies after a deliberate change.
var necroSME = []struct{ goals, table, combat, note string }{
	{"testdata/necromancer-summoner-sme.json", "necromancer_summoner_sme", "summoner",
		"NECRO3: Summoner per the d2-sme answers of 2026-11-22. Tier experimental, opt-in by name (levelerBuild necromancer_summoner_sme); defaults unchanged. Point counts after clvl 30 and every number the SME tagged UNVERIFIED (Revive count and duration, curse durations, Blood/Iron Golem prerequisites) are U until read from skills.txt."},
	{"testdata/necromancer-poison-nova-sme.json", "necromancer_poison_nova_sme", "poisonnova",
		"NECRO3: Poison Nova caster per the d2-sme answers of 2026-11-22. Tier experimental, opt-in by name (levelerBuild necromancer_poison_nova_sme); defaults unchanged. Poison Nova, Poison Explosion, Poison Dagger synergy and the Lower Resist prerequisite chain numbers are U until read from skills.txt."},
}

func TestNecromancerSMEBuildsMatchTables(t *testing.T) {
	c, err := agency.LoadCaps()
	if err != nil {
		t.Fatal(err)
	}
	rules := c.BuildRules()
	for _, nb := range necroSME {
		b, err := buildfile.LoadFile(nb.goals, rules)
		if err != nil {
			t.Fatalf("%s: %v", nb.goals, err)
		}
		gen, err := buildfile.ImportPlan(b, rules)
		if err != nil {
			t.Fatalf("%s: %v", nb.goals, err)
		}
		var want map[string]json.RawMessage
		if err := json.Unmarshal(gen, &want); err != nil {
			t.Fatal(err)
		}
		for _, path := range []string{"../plandata/build-" + nb.table + ".json", "../../docs/plan/data/build-" + nb.table + ".json"} {
			raw, err := os.ReadFile(path)
			have := map[string]json.RawMessage{}
			if err == nil {
				if err := json.Unmarshal(raw, &have); err != nil {
					t.Fatal(err)
				}
			}
			if os.Getenv("UPDATE") != "" {
				for k, v := range map[string]any{"build": nb.table, "class": "necromancer", "combat": nb.combat, "selector": nb.table, "tier": "experimental", "note": nb.note} {
					have[k], _ = json.Marshal(v)
				}
				have["rows"] = want["rows"]
				if _, ok := have["schema"]; !ok {
					have["schema"] = want["schema"]
				}
				out, _ := json.MarshalIndent(have, "", " ")
				if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
					t.Fatal(err)
				}
				continue
			}
			if err != nil {
				t.Fatalf("%s: %v (UPDATE=1 writes it)", path, err)
			}
			var a, g bytes.Buffer
			_ = json.Compact(&a, have["rows"])
			_ = json.Compact(&g, want["rows"])
			if a.String() != g.String() {
				t.Errorf("%s: rows differ from the goal file's expansion", path)
			}
		}
		raw, err := plandata.FS.ReadFile("build-" + nb.table + ".json")
		if err != nil {
			continue // UPDATE run: the embed is stale until rebuilt
		}
		lb, err := leveler.ParseBuild(raw)
		if err != nil {
			t.Fatal(err)
		}
		for _, p := range buildfile.CheckTable(lb, rules) {
			t.Errorf("%s: %s", nb.table, p)
		}
	}
}
