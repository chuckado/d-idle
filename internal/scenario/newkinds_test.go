package scenario

import (
	"fmt"
	"strings"
	"testing"
)

// PGFORMAT: skill_points, log_line, only_class and trigger.build, each must-fail and each against Paladin and Sorceress.

// levels builds a class-order base_levels list of 30 with the given skill ids set.
func levels(class string, set map[int]int) []int {
	l := make([]int, 30)
	for id, n := range set {
		l[id-classSkillBase[strings.ToLower(class)]] = n
	}
	return l
}

func TestSkillPointsTwoClasses(t *testing.T) {
	cases := []struct {
		class      string
		skill, own int // Fire Ball 47 for the Sorceress, Blessed Hammer 112 for the Paladin
	}{{"Sorceress", 47, 16}, {"Paladin", 112, 20}}
	for _, c := range cases {
		lv := ip(c.own)
		st := State{Class: c.class, Level: ip(30), SkillLevels: levels(c.class, map[int]int{c.skill: c.own})}
		ok := Assert{Kind: ASkillPoints, Skill: c.skill, Min: c.own, Max: lv}
		if r := Evaluate([]Assert{ok}, st)[0]; !r.OK {
			t.Errorf("%s exact: %+v", c.class, r)
		}
		for name, a := range map[string]Assert{
			"too few":  {Kind: ASkillPoints, Skill: c.skill, Min: c.own + 1},
			"too many": {Kind: ASkillPoints, Skill: c.skill, Max: ip(c.own - 1)},
			"unspent":  {Kind: ASkillPoints, Skill: c.skill + 1, Min: 1},
		} {
			if r := Evaluate([]Assert{a}, st)[0]; r.OK || r.Unobserved {
				t.Errorf("%s %s: must fail, got %+v", c.class, name, r)
			}
		}
		// the other class's skill id is not this class's skill: unobserved, never a pass
		other := 47
		if c.class == "Sorceress" {
			other = 112
		}
		if r := Evaluate([]Assert{{Kind: ASkillPoints, Skill: other, Min: 1}}, st)[0]; !r.Unobserved {
			t.Errorf("%s foreign skill: %+v", c.class, r)
		}
		// at_level: another level is unobserved (BLOCKED), the right one decides
		if r := Evaluate([]Assert{{Kind: ASkillPoints, Skill: c.skill, Min: 1, AtLevel: 31}}, st)[0]; !r.Unobserved {
			t.Errorf("%s at_level mismatch: %+v", c.class, r)
		}
		if r := Evaluate([]Assert{{Kind: ASkillPoints, Skill: c.skill, Min: 1, AtLevel: 30}}, st)[0]; !r.OK {
			t.Errorf("%s at_level: %+v", c.class, r)
		}
		if r := Evaluate([]Assert{ok}, State{Class: c.class})[0]; !r.Unobserved {
			t.Errorf("%s no dump: %+v", c.class, r)
		}
	}
	for _, a := range []Assert{{Kind: ASkillPoints, Min: 1}, {Kind: ASkillPoints, Skill: 47}} {
		if a.Validate() == nil {
			t.Errorf("%+v validates", a)
		}
	}
}

func logJSON(t0 int, lines ...string) []byte {
	var b strings.Builder
	for i, l := range lines {
		fmt.Fprintf(&b, `{"time":"2026-11-20T10:%02d:00Z","level":"INFO","msg":%q,"inst":"x"}`+"\n", t0+i, l)
	}
	return []byte(b.String())
}

func TestLogLineWindowAndCounts(t *testing.T) {
	for _, class := range []string{"Paladin", "Sorceress"} {
		log := logJSON(0, "start", "planner shadow: clvl=18 class=sorceress agree=true", "x", "planner shadow: clvl=18 class=paladin agree=false", "planner shadow: no plan", "end")
		as := []Assert{
			{Kind: ALogLine, Pattern: `^planner shadow: clvl=\d+ .*agree=true`},
			{Kind: ALogLine, Pattern: `agree=false`, Max: ip(0)},
			{Kind: ALogLine, Pattern: `^planner shadow:`, Min: 3, Max: ip(3), FromMin: 0},
			{Kind: ALogLine, Pattern: `agree=true`, FromMin: 3, ToMin: 5},
		}
		lines, ok := CollectLog(as, log, 5)
		if !ok {
			t.Fatal("log not read")
		}
		st := State{Class: class, Has: []string{"log"}, LogLines: lines}
		rs := Evaluate(as, st)
		want := []bool{true, false, true, false} // agree=false is present, and agree=true is before minute 3 (virtual 5 min over 5 log minutes)
		for i, r := range rs {
			if r.OK != want[i] || r.Unobserved {
				t.Errorf("%s assert %d: %+v", class, i, r)
			}
		}
		if status, _ := Verdict(rs); status != Red {
			t.Errorf("%s verdict %s", class, status)
		}
	}
}

func TestLogLineUnobservedAndValidate(t *testing.T) {
	a := Assert{Kind: ALogLine, Pattern: "x"}
	if r := Evaluate([]Assert{a}, State{})[0]; !r.Unobserved {
		t.Fatalf("no log: %+v", r)
	}
	if _, ok := CollectLog([]Assert{a}, []byte("not json\n"), 5); ok {
		t.Fatal("a non-bot log counted as read")
	}
	for name, bad := range map[string]Assert{
		"no pattern": {Kind: ALogLine}, "bad regexp": {Kind: ALogLine, Pattern: "("},
		"max below min": {Kind: ALogLine, Pattern: "x", Min: 2, Max: ip(1)}, "window": {Kind: ALogLine, Pattern: "x", FromMin: 3, ToMin: 2},
	} {
		if bad.Validate() == nil {
			t.Errorf("%s validates", name)
		}
	}
}

func TestOnlyClassSkips(t *testing.T) {
	as := []Assert{{Kind: ALevelAtLeast, Min: 30, OnlyClass: "Sorceress"}, {Kind: ALevelAtLeast, Min: 5}}
	for _, c := range []struct {
		class string
		n     int
	}{{"Sorceress", 2}, {"Paladin", 1}} {
		if got := Evaluate(as, State{Class: c.class, Level: ip(10)}); len(got) != c.n {
			t.Errorf("%s: %d results, want %d", c.class, len(got), c.n)
		}
	}
}

func TestTriggerBuildMustMatchClass(t *testing.T) {
	s := baseScenario()
	s.Matrix.Classes = []string{"Sorceress"}
	s.Trigger.Build = "sorceress_normal30"
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	s.Matrix.Classes = []string{"Sorceress", "Paladin"}
	if err := s.Validate(); err == nil || !strings.Contains(err.Error(), "matrix class Paladin") {
		t.Fatalf("a Sorceress table for a Paladin: %v", err)
	}
	s.Matrix.Classes = []string{"Paladin"}
	s.Trigger.Build = "paladin-normal30-v2"
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	s.Trigger.Build = "no-such-table"
	if err := s.Validate(); err == nil {
		t.Fatal("unknown build accepted")
	}
}

func TestSetupCarriesBuild(t *testing.T) {
	s := baseScenario()
	s.Matrix.Classes = []string{"Sorceress"}
	s.Trigger.Build = "sorceress_normal30"
	su, err := BuildSetup(s, "Sorceress", 1)
	if err != nil || su.Build != "sorceress_normal30" {
		t.Fatalf("%+v %v", su, err)
	}
}

// Digests of states without the new fields do not change, and the log is not part of a digest.
func TestDigestIgnoresLog(t *testing.T) {
	a := State{Source: "fixture", Level: ip(3)}
	b := a
	b.LogLines = []LogLine{{T: 1.5, Text: "x"}}
	if DigestState(a) != DigestState(b) {
		t.Fatal("log lines changed the digest")
	}
}

func TestItemInPackMax(t *testing.T) {
	two := 2
	st := State{Has: []string{"items"}, Items: []StateItem{{Code: "tsc", Where: "pack"}, {Code: "tsc", Where: "pack"}, {Code: "tsc", Where: "pack"}}}
	if r := Evaluate([]Assert{{Kind: AItemInPack, Code: "tsc", Max: &two}}, st); r[0].OK {
		t.Errorf("3 scrolls passed max 2")
	}
	st.Items = st.Items[:2]
	if r := Evaluate([]Assert{{Kind: AItemInPack, Code: "tsc", Max: &two}}, st); !r[0].OK {
		t.Errorf("2 scrolls failed max 2: %+v", r[0])
	}
	st.Items = nil
	if r := Evaluate([]Assert{{Kind: AItemInPack, Code: "tsc", Max: &two}}, st); !r[0].OK {
		t.Errorf("0 scrolls failed max 2")
	}
}
