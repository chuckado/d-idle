package questboard

import (
	"strings"
	"testing"
)

const fams = `{"schema":"quests.families/1","families":{"paladin":"melee","sorceress":"caster"}}`

// act builds a one-act file around the given quest bodies.
func act(n int, quests ...string) []byte {
	return []byte(`{"schema":"quests.def/1","act":` + itoa(n) + `,"source":"t","quests":[` + strings.Join(quests, ",") + `]}`)
}

func itoa(n int) string { return string(rune('0' + n)) }

// quest renders a valid quest row; mod replaces substrings to break one field.
func quest(id string, mod ...string) string {
	s := `{"id":"` + id + `","taskId":"","title":"T","blurb":"b","act":1,"difficulties":["normal","nm","hell"],"kind":"optional",
	 "chunks":[{"id":"c","title":"C","ends":"handin","handin":{"npc":"Akara"},"minutes":[1,2]}],"minutes":[1,2],"risk":"routine","autoStart":"allowed"}`
	for i := 0; i+1 < len(mod); i += 2 {
		if !strings.Contains(s, mod[i]) {
			panic("test fixture: " + mod[i])
		}
		s = strings.Replace(s, mod[i], mod[i+1], 1)
	}
	return s
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name string
		acts [][]byte
		want string // substring that must name the culprit
	}{
		{"unknown act", [][]byte{act(7, quest("a"))}, "unknown act 7"},
		{"quest in wrong act file", [][]byte{act(2, quest("a"))}, `quest "a" is act 1`},
		{"duplicate", [][]byte{act(1, quest("a"), quest("a"))}, `duplicate quest "a"`},
		{"unknown kind", [][]byte{act(1, quest("a", `"optional"`, `"fun"`))}, `unknown kind "fun"`},
		{"unknown risk", [][]byte{act(1, quest("a", `"routine"`, `"mild"`))}, `unknown risk "mild"`},
		{"unknown difficulty", [][]byte{act(1, quest("a", `"hell"`, `"insane"`))}, `unknown difficulty "insane"`},
		{"unknown ends", [][]byte{act(1, quest("a", `"ends":"handin"`, `"ends":"never"`))}, `unknown ends "never"`},
		{"handin without giver", [][]byte{act(1, quest("a", `"handin":{"npc":"Akara"},`, ``))}, "needs a handin giver"},
		{"requires unknown quest", [][]byte{act(1, quest("a", `"kind":"optional",`, `"kind":"optional","requires":["ghost"],`))}, `unknown quest "ghost"`},
		{"requires itself", [][]byte{act(1, quest("a", `"kind":"optional",`, `"kind":"optional","requires":["a"],`))}, "requires itself"},
		{"after unknown chunk", [][]byte{act(1, quest("a", `"ends":"handin"`, `"after":["zz"],"ends":"handin"`))}, `unknown chunk "zz"`},
		{"hazard auto", [][]byte{act(1, quest("a", `"routine"`, `"hazard"`))}, "hazard needs autoStart never"},
		{"bad minutes", [][]byte{act(1, quest("a", `"minutes":[1,2],"risk"`, `"minutes":[5,2],"risk"`))}, "minutes"},
		{"unknown field", [][]byte{act(1, quest("a", `"blurb"`, `"blurbs"`))}, "blurbs"},
		{"meta unknown difficulty", [][]byte{act(1, quest("a", `"kind":"optional",`, `"kind":"optional","meta":{"easy":1},`))}, `unknown difficulty "easy"`},
		{"meta negative", [][]byte{act(1, quest("a", `"kind":"optional",`, `"kind":"optional","meta":{"normal":-1},`))}, "negative"},
		{"unknown family in floor", [][]byte{act(1, quest("a", `"kind":"optional",`, `"kind":"optional","legacyFloor":{"tank":3},`))}, `unknown family "tank"`},
		{"no chunks", [][]byte{act(1, quest("a", `"chunks":[{"id":"c","title":"C","ends":"handin","handin":{"npc":"Akara"},"minutes":[1,2]}]`, `"chunks":[]`))}, "no chunks"},
		{"chunk cycle", [][]byte{act(1, `{"id":"a","taskId":"","title":"T","blurb":"b","act":1,"difficulties":["normal"],"kind":"optional","minutes":[1,2],"risk":"routine","autoStart":"allowed",
		  "chunks":[{"id":"x","title":"X","ends":"item","after":["y"],"minutes":[1,2]},{"id":"y","title":"Y","ends":"item","after":["x"],"minutes":[1,2]}]}`)}, "circular"},
		{"quest cycle", [][]byte{act(1,
			quest("a", `"kind":"optional",`, `"kind":"optional","requires":["b"],`),
			quest("b", `"kind":"optional",`, `"kind":"optional","requires":["a"],`))}, "circular"},
	}
	for _, c := range cases {
		_, err := Parse([]byte(fams), c.acts...)
		if err == nil {
			t.Errorf("%s: accepted", c.name)
		} else if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error %q does not name %q", c.name, err, c.want)
		}
	}
	if _, err := Parse([]byte(`{"schema":"x"}`), act(1, quest("a"))); err == nil {
		t.Error("bad families schema accepted")
	}
	if _, err := Parse([]byte(fams), act(1, quest("a"))); err != nil {
		t.Errorf("valid fixture rejected: %v", err)
	}
}

func TestLookupsNameTheUnknown(t *testing.T) {
	b, _ := Load()
	if _, err := b.Quest("nope"); err == nil || !strings.Contains(err.Error(), `"nope"`) {
		t.Errorf("quest: %v", err)
	}
	if _, err := b.Act(9); err == nil || !strings.Contains(err.Error(), "9") {
		t.Errorf("act: %v", err)
	}
	if _, err := b.Chunk("den_of_evil", "zz"); err == nil || !strings.Contains(err.Error(), `"zz"`) {
		t.Errorf("chunk: %v", err)
	}
	if _, err := DiffIndex("easy"); err == nil {
		t.Error("difficulty accepted")
	}
}
