package levelsim

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/power"
)

// shadowTab has an easy field (area 2) and a hard marsh (area 6); the numbers are made up, like the log below.
func shadowTab() *danger.Table {
	mon := func(life, dmg float64) []danger.Monster {
		return []danger.Monster{
			{Key: "m", Share: 0.6, Life: int(life), Def: 20, ToHit: 70, Dmg: dmg, Reach: "melee"},
			{Key: "r", Share: 0.4, Life: int(life), Def: 20, ToHit: 70, Dmg: dmg, Reach: "ranged"}}
	}
	z := func(area, mlvl, n int, life, dmg float64) *danger.Zone {
		return &danger.Zone{Area: area, Name: fmt.Sprintf("zone%d", area), Diff: "normal", MLvl: mlvl, Monsters: mon(life, dmg),
			Encounters: danger.Encounters{
				Typical: danger.Encounter{N: n, Melee: n - 1, Ranged: 1},
				Bad:     danger.Encounter{N: 2 * n, Melee: 2*n - 2, Ranged: 2},
				Elite:   danger.Encounter{N: n + 1, Melee: n, Ranged: 1}}}
	}
	return &danger.Table{Params: danger.DefaultParams(), Zones: []*danger.Zone{z(2, 3, 3, 20, 4), z(6, 9, 6, 140, 30)}}
}

func shadowLog() string {
	t0 := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	var b strings.Builder
	at := func(min int) string { return t0.Add(time.Duration(min) * time.Minute).Format(time.RFC3339) }
	entry := func(min, area, clvl int, merc string) {
		fmt.Fprintf(&b, "time=%s level=INFO msg=\"zone entry\" area=%d clvl=%d life=150/150 heal=4 mana=40/40 tome=1 merc=%s\n", at(min), area, clvl, merc)
	}
	min := 0
	for i := 0; i < 4; i++ { // easy field, long quiet stays: good
		entry(min, 2, 10, "alive")
		min += 12
	}
	for i := 0; i < 4; i++ { // the marsh without a merc: chicken, merc lost or death
		entry(min, 6, 10, "dead")
		fmt.Fprintf(&b, "time=%s level=WARN msg=\"safety decision\" action=chicken area=6 life=10 maxLife=150\n", at(min+2))
		if i%2 == 1 {
			fmt.Fprintf(&b, "time=%s level=WARN msg=\"merc: died\" area=6\n", at(min+3))
		}
		min += 4
	}
	entry(min, 1, 10, "alive") // a town the profile does not hold
	fmt.Fprintf(&b, "time=%s level=INFO msg=minute area=1 clvl=10\n", at(min+1))
	return b.String()
}

func TestShadowReplayPassRule(t *testing.T) {
	m, err := power.NewModel(danger.DefaultParams())
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"bruiser", "caster"} { // two builds
		a, _ := m.P.Archetype(id)
		rep, err := ShadowReplay(strings.NewReader(shadowLog()), ShadowOpts{Tab: shadowTab(), M: m, Arch: a, Diff: "normal"})
		if err != nil {
			t.Fatal(err)
		}
		if len(rep.Stays) != 9 || rep.Bad != 4 || rep.Good != 4 {
			t.Fatalf("%s: %d stays, bad %d good %d\n%s", id, len(rep.Stays), rep.Bad, rep.Good, rep.Text())
		}
		if v := rep.Verdict(); v != "PASS" {
			t.Errorf("%s: verdict %s\n%s", id, v, rep.Text())
		}
		if last := rep.Stays[8]; last.Skipped == "" {
			t.Errorf("%s: the town stay must be skipped", id)
		}
	}
}

func TestShadowReplayVerdictNeedsCases(t *testing.T) {
	m, _ := power.NewModel(danger.DefaultParams())
	a, _ := m.P.Archetype("bruiser")
	rep, err := ShadowReplay(strings.NewReader(""), ShadowOpts{Tab: shadowTab(), M: m, Arch: a, Diff: "normal"})
	if err != nil || !strings.HasPrefix(rep.Verdict(), "n/a") {
		t.Errorf("empty log: %v %v", rep, err)
	}
	if _, err := ShadowReplay(strings.NewReader(""), ShadowOpts{}); err == nil {
		t.Error("missing inputs must be an error")
	}
}

// A model that rates the zone where the hero chickened as safe fails the rule.
func TestShadowReplayFails(t *testing.T) {
	m, _ := power.NewModel(danger.DefaultParams())
	a, _ := m.P.Archetype("bruiser")
	log := strings.ReplaceAll(shadowLog(), "area=6", "area=2") // the chickens now happen in the easy field
	rep, err := ShadowReplay(strings.NewReader(log), ShadowOpts{Tab: shadowTab(), M: m, Arch: a, Diff: "normal"})
	if err != nil {
		t.Fatal(err)
	}
	if v := rep.Verdict(); v != "FAIL" {
		t.Errorf("verdict %s\n%s", v, rep.Text())
	}
}
