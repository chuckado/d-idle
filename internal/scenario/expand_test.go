package scenario

import (
	"strings"
	"testing"
)

func TestCountZoneEntries(t *testing.T) {
	log := []byte(`{"level":"INFO","msg":"zone entry","area":1,"clvl":10}
{"level":"INFO","msg":"arrival","area":5}
{"level":"INFO","msg":"zone entry","area":5}
not json zone entry
{"level":"INFO","msg":"zone entry","area":"2"}
{"level":"INFO","msg":"zone entry","area":5}
`)
	trail, ok := CountZoneEntries(log)
	if !ok || len(trail) != 4 || trail[0] != 1 || trail[2] != 2 {
		t.Fatalf("trail %v ok %v", trail, ok)
	}
	if AreaChangeCount(trail) != 3 || AreaChangeCount(nil) != 0 {
		t.Fatal("count is entries after the first")
	}
	if _, ok := CountZoneEntries([]byte("{}\n")); ok {
		t.Fatal("no zone entry line: not observed")
	}
}

func TestAreaChangesAssertionUsesTheCount(t *testing.T) {
	mx := 12
	as := []Assert{{Kind: AAreaChangesLE, Max: &mx}}
	if r := Evaluate(as, State{}); !r[0].Unobserved {
		t.Fatal("no count: unobserved")
	}
	n := 12
	if r := Evaluate(as, State{AreaChanges: &n}); !r[0].OK {
		t.Fatalf("12 <= 12: %+v", r[0])
	}
	n = 13
	if r := Evaluate(as, State{AreaChanges: &n}); r[0].OK || r[0].Unobserved {
		t.Fatalf("13 must be red: %+v", r[0])
	}
}

func TestPortalOpen(t *testing.T) {
	as := []Assert{{Kind: APortalOpen, Area: 4}}
	if r := Evaluate(as, State{}); !r[0].Unobserved {
		t.Fatal("portals not read: unobserved")
	}
	st := State{Has: []string{"portals"}, Portals: []StatePortal{{Which: 0, Level: 38}, {Which: 1, Level: 1}}}
	if r := Evaluate(as, st); r[0].OK || r[0].Unobserved {
		t.Fatalf("no portal in 4: red, got %+v", r[0])
	}
	st.Portals = append(st.Portals, StatePortal{Which: 2, Level: 4})
	if r := Evaluate(as, st); !r[0].OK {
		t.Fatalf("%+v", r[0])
	}
	if (Assert{Kind: APortalOpen}).Validate() == nil {
		t.Fatal("portal_open needs area")
	}
}

func TestGroundItemsAreNotCarried(t *testing.T) {
	st := State{Has: []string{"items", "ground"}, Items: []StateItem{{Code: "rin", Where: "ground"}}}
	carried := Evaluate([]Assert{{Kind: AItemInPack, Code: "rin"}}, st)
	if carried[0].OK {
		t.Fatal("a ring on the floor is not in the pack")
	}
	if r := Evaluate([]Assert{{Kind: AItemAbsent, Code: "rin", Where: "ground"}}, st); r[0].OK {
		t.Fatal("the ring still lies there: red")
	}
	st.Items = nil
	if r := Evaluate([]Assert{{Kind: AItemAbsent, Code: "rin", Where: "ground"}}, st); !r[0].OK {
		t.Fatal("floor empty: green")
	}
	st.Has = []string{"items"}
	if r := Evaluate([]Assert{{Kind: AItemAbsent, Code: "rin", Where: "ground"}}, st); !r[0].Unobserved {
		t.Fatal("floor not read (reloaded hero): unobserved, never a pass")
	}
}

func TestSetupItemQualityGroundAndLife(t *testing.T) {
	s := baseScenario()
	s.Pre = Pre{Level: 10, Start: StartField, Area: 2, LifePct: 15,
		Items: []Item{{Code: "rin", Where: "ground", Quality: 6, Unid: true}, {Code: "cap", Quality: 4, Unid: true}}}
	for _, class := range []string{"Paladin", "Sorceress"} {
		su, err := BuildSetup(s, class, 3)
		if err != nil {
			t.Fatal(err)
		}
		var lines []string
		for _, v := range su.Verbs {
			lines = append(lines, v.Line)
			if !strings.Contains(v.Reason, class) {
				t.Errorf("reason lacks the class: %q", v.Reason)
			}
		}
		all := strings.Join(lines, "|")
		for _, want := range []string{"debug give rin 6 unid", "debug give cap pack q4 unid", "debug goto 2"} {
			if !strings.Contains(all, want) {
				t.Errorf("%s: missing %q in %s", class, want, all)
			}
		}
		if su.LifePct != 15 || len(su.Unsupported) != 0 {
			t.Errorf("life %d unsupported %v", su.LifePct, su.Unsupported)
		}
	}
	s.Pre.Items = []Item{{Code: "rin", Where: "ground"}}
	if su, _ := BuildSetup(s, "Paladin", 3); len(su.Unsupported) == 0 {
		t.Fatal("a ground item without a quality is not buildable")
	}
}

func TestDriverValidation(t *testing.T) {
	s := baseScenario()
	s.Trigger.Driver = DriverBoardTaps
	if s.Validate() == nil {
		t.Fatal("board-taps needs a quest")
	}
	s.Trigger.TapQuest = "search_for_cain"
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	if s.TappedQuest() != "search_for_cain" {
		t.Fatal(s.TappedQuest())
	}
	s.Pre.Board = &QuestRef{Quest: "a", Chunk: "b"}
	if s.Validate() == nil {
		t.Fatal("the driver accepts the chunk itself")
	}
	s.Pre.Board, s.Trigger.Driver = nil, "mouse"
	if s.Validate() == nil {
		t.Fatal("unknown driver")
	}
}
