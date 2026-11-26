package plandata

import (
	"encoding/json"
	"testing"
)

type areasFile struct {
	Areas []struct {
		ID   int    `json:"id"`
		Act  int    `json:"act"`
		Town bool   `json:"town"`
		Name string `json:"name"`
	} `json:"areas"`
	WPs map[string][]int `json:"waypoints_by_act"`
}

func loadAreas(t *testing.T) (map[int]int, map[int]bool, map[int]bool) {
	t.Helper()
	b, err := FS.ReadFile("areas.json")
	if err != nil {
		t.Fatal(err)
	}
	var f areasFile
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	act, town, wp := map[int]int{}, map[int]bool{}, map[int]bool{}
	for _, a := range f.Areas {
		act[a.ID], town[a.ID] = a.Act, a.Town
	}
	for _, id := range f.WPs["2"] {
		wp[id] = true
	}
	return act, town, wp
}

type admRoad struct {
	ID         string         `json:"id"`
	Areas      []int          `json:"areas"`
	MinLevel   map[string]int `json:"minLevel"`
	MinMaxLife map[string]int `json:"minMaxLife"`
}

func loadRoads(t *testing.T) []admRoad {
	t.Helper()
	b, err := FS.ReadFile("admission.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Roads []admRoad `json:"roads"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	return f.Roads
}

// TestAct2PlanShape runs for every class of the file (no class is named here): areas are Act 2 dungeons or fields,
// bands rise, no clvl from the first entry to the last exit is uncovered with or without a merc, and a slice entry
// never sits below the admission road floor of an area it names.
func TestAct2PlanShape(t *testing.T) {
	p, err := LoadAct2Plan()
	if err != nil {
		t.Fatal(err)
	}
	if p.Schema != "act2plan/1" || p.Difficulty != "normal" {
		t.Fatalf("schema %q difficulty %q", p.Schema, p.Difficulty)
	}
	if len(p.Classes) < 2 {
		t.Fatalf("want at least 2 classes (build flexibility), have %d", len(p.Classes))
	}
	act, town, wp := loadAreas(t)
	roads := loadRoads(t)
	archetypes := map[string]bool{}
	for name, c := range p.Classes {
		archetypes[c.Archetype] = true
		if c.Build == "" || c.Archetype == "" || len(c.Slices) == 0 {
			t.Fatalf("%s: build, archetype and slices needed", name)
		}
		ids := map[string]bool{}
		for i, s := range c.Slices {
			if ids[s.ID] || s.ID == "" || s.Kind != "farm" {
				t.Errorf("%s: slice %q repeated, empty or not a farm", name, s.ID)
			}
			ids[s.ID] = true
			if s.Entry < 1 || s.Entry >= s.Exit || s.NoMercAdd < 0 || s.Entry+s.NoMercAdd > s.Exit {
				t.Errorf("%s %s: band [%d,%d] add %d", name, s.ID, s.Entry, s.Exit, s.NoMercAdd)
			}
			for _, a := range s.Areas {
				if act[a] != 2 || town[a] {
					t.Errorf("%s %s: area %d is not an Act 2 field or dungeon", name, s.ID, a)
				}
				for _, r := range roads {
					for _, ra := range r.Areas {
						if ra == a && s.Entry < r.MinLevel[c.Archetype] {
							t.Errorf("%s %s: entry %d below road %s floor %d", name, s.ID, s.Entry, r.ID, r.MinLevel[c.Archetype])
						}
						if ra == a && s.MinMaxLife < r.MinMaxLife[c.Archetype] {
							t.Errorf("%s %s: life gate %d below road %s life %d", name, s.ID, s.MinMaxLife, r.ID, r.MinMaxLife[c.Archetype])
						}
					}
				}
			}
			switch s.WayIn {
			case "walk":
			case "waypoint":
				if s.ViaWP == 0 {
					t.Errorf("%s %s: waypoint way_in needs via_wp", name, s.ID)
				}
			default:
				t.Errorf("%s %s: way_in %q", name, s.ID, s.WayIn)
			}
			if s.ViaWP != 0 && !wp[s.ViaWP] {
				t.Errorf("%s %s: via_wp %d has no Act 2 waypoint", name, s.ID, s.ViaWP)
			}
			if i > 0 && (s.Entry < c.Slices[i-1].Entry || s.Exit < c.Slices[i-1].Exit) {
				t.Errorf("%s: %s bands not monotone", name, s.ID)
			}
		}
		for _, merc := range []bool{true, false} {
			if l, ok := act2Gap(c.Slices, merc); !ok {
				t.Errorf("%s: no slice holds clvl %d (merc %v)", name, l, merc)
			}
		}
	}
	if len(archetypes) < 2 {
		t.Errorf("the plan covers one archetype only: %v", archetypes)
	}
}

func act2Gap(sl []Act2Slice, merc bool) (int, bool) {
	entry := func(s Act2Slice) int {
		if merc {
			return s.Entry
		}
		return s.Entry + s.NoMercAdd
	}
	lo, hi := 1<<30, 0
	for _, s := range sl {
		lo, hi = min(lo, entry(s)), max(hi, s.Exit)
	}
	for l := lo; l <= hi; l++ {
		held := false
		for _, s := range sl {
			held = held || entry(s) <= l && l <= s.Exit
		}
		if !held {
			return l, false
		}
	}
	return 0, true
}

// TestAct2PlanJoinsAct1 checks that the Act 2 plan starts inside the Act 1 plan of the same class (no clvl gap at the hop)
// for every class both files carry.
func TestAct2PlanJoinsAct1(t *testing.T) {
	p, err := LoadAct2Plan()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := FS.ReadFile("levelplan-normal.json")
	var l1 struct {
		Classes map[string]struct {
			Slices []struct {
				Kind string `json:"kind"`
				Exit int    `json:"exit"`
			} `json:"slices"`
		} `json:"classes"`
	}
	if err := json.Unmarshal(b, &l1); err != nil {
		t.Fatal(err)
	}
	for name, c := range p.Classes {
		a1, ok := l1.Classes[name]
		if !ok {
			t.Errorf("%s has no Act 1 plan rows", name)
			continue
		}
		last := 0
		for _, s := range a1.Slices {
			last = max(last, s.Exit)
		}
		if c.Slices[0].Entry > last {
			t.Errorf("%s: Act 2 starts at clvl %d, Act 1 plan ends at %d", name, c.Slices[0].Entry, last)
		}
	}
}

// TestAct2AdmissionRoads: every Act 2 dungeon area of the plan has a road row, each row names Act 2 areas only, carries
// evidence fields (threshold governance) and has both archetypes.
func TestAct2AdmissionRoads(t *testing.T) {
	act, _, _ := loadAreas(t)
	b, _ := FS.ReadFile("admission.json")
	var f struct {
		Roads []struct {
			admRoad
			Evidence    *float64 `json:"evidence"`
			EvidenceRef string   `json:"evidence_ref"`
		} `json:"roads"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	covered := map[int]bool{}
	for _, r := range f.Roads {
		if r.Evidence == nil || r.EvidenceRef == "" {
			t.Errorf("road %s: evidence and evidence_ref needed", r.ID)
		}
		a2 := false
		for _, a := range r.Areas {
			a2 = a2 || act[a] == 2
			covered[a] = true
		}
		if a2 && (r.MinLevel["melee"] <= 0 || r.MinLevel["caster"] <= 0 || r.MinLevel["caster"] < r.MinLevel["melee"]) {
			t.Errorf("road %s: floors %v", r.ID, r.MinLevel)
		}
	}
	p, _ := LoadAct2Plan()
	for _, a := range append([]int{73, 47, 48, 49}, p.Duriel.Areas...) {
		if !covered[a] {
			t.Errorf("area %d has no road row", a)
		}
	}
	for _, c := range p.Classes {
		for _, s := range c.Slices {
			if s.MinMaxLife == 0 {
				continue // open fields carry no life gate and no road row
			}
			for _, a := range s.Areas {
				if !covered[a] {
					t.Errorf("slice %s area %d has a life gate but no road row", s.ID, a)
				}
			}
		}
	}
}

func TestAct2PlanUnknownClass(t *testing.T) {
	p, _ := LoadAct2Plan()
	if _, err := p.For("warlock"); err == nil {
		t.Error("an unknown class must be an error")
	}
	if p.Merc.NPC == "" || p.Merc.Act != 2 || len(p.Merc.Aura) < 2 {
		t.Errorf("merc row: %+v", p.Merc)
	}
}
