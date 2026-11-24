package scenario

import (
	"strings"
	"testing"
)

func verbLines(t *testing.T, s Scenario) []string {
	t.Helper()
	su, err := BuildSetup(s, "Paladin", 11)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, v := range su.Verbs {
		out = append(out, v.Line)
	}
	return out
}

func TestSuperSpawnAndObjectAddVerbs(t *testing.T) {
	s := Scenario{ID: "step.x", Pre: Pre{Start: StartField, Area: 28, Quests: []QWord{{Record: 3, Word: 12}},
		Spawn: []Spawn{{Super: "countess"}}, Objects: []ObjectAdd{{Class: 108}, {Class: 392, X: 10, Y: 20}}}}
	got := strings.Join(verbLines(t, s), "\n")
	for _, want := range []string{"debug spawn super countess", "debug object add 108", "debug object add 392 10 20", "debug goto 28"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing verb %q in:\n%s", want, got)
		}
	}
	if i, j := strings.Index(got, "debug goto 28"), strings.Index(got, "debug object add 108"); i > j {
		t.Error("the object is added before the field warp: it would land in the town")
	}
}

// Must-fail: a super spawn with extra fields, an object in town and an empty spawn are refused, never guessed.
func TestStepVerbsRefuseWhatTheyCannotDo(t *testing.T) {
	for name, s := range map[string]Scenario{
		"super with n":   {ID: "a", Pre: Pre{Start: StartField, Area: 1, Spawn: []Spawn{{Super: "countess", N: 2}}}},
		"empty spawn":    {ID: "b", Pre: Pre{Start: StartField, Area: 1, Spawn: []Spawn{{}}}},
		"object in town": {ID: "c", Pre: Pre{Objects: []ObjectAdd{{Class: 108}}}},
		"object class 0": {ID: "d", Pre: Pre{Start: StartField, Area: 1, Objects: []ObjectAdd{{}}}},
	} {
		if _, err := BuildSetup(s, "Paladin", 11); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}
