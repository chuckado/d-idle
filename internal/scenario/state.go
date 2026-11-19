package scenario

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

// StateVersion is the format of the observed game state ("didle.scenario-state/1").
const StateVersion = "didle.scenario-state/1"

// State is what the observer read from the game AFTER the run: facts of the world, never log lines. A fact that was not
// observed is nil or absent, and an assertion on it is UNOBSERVED (the scenario is BLOCKED, not red).
type State struct {
	Version string `json:"version"`
	// Source says where the facts came from: "d2d-observer" (a read-only d2d loaded the saved hero), "fixture" (tests).
	Source string `json:"source"`
	// Has lists the fact groups the observer actually read (items, objects): an empty list of items is a fact only when
	// "items" is in Has.
	Has   []string `json:"has,omitempty"`
	Area  *int     `json:"area,omitempty"` // levels.txt id
	Level *int     `json:"level,omitempty"`
	Gold  *int     `json:"gold,omitempty"`
	Stash *int     `json:"stash,omitempty"`
	// LifePct is current life in percent of maximum.
	LifePct *int `json:"life_pct,omitempty"`
	// Quests maps the fork's quest record index to its 16-bit word (docs/forkcheats.md).
	Quests map[int]uint16 `json:"quests,omitempty"`
	// Items lists every item the hero carries: Code is the base code (bks, bkd), Where pack, stash, belt, body, cursor.
	Items []StateItem `json:"items,omitempty"`
	// Objects are the level objects read with debug objects: class and state (NU, ON ...).
	Objects []StateObject `json:"objects,omitempty"`
	// Portals are the open portals of `debug portals` (which, level the portal stands in).
	Portals []StatePortal `json:"portals,omitempty"`
	// AreaChanges counts area changes during the run when the observer could count them (nil: not observable). It is the
	// number of `zone entry` lines of the bot's own log after the first (one line per area change, internal/bot/lifelog.go):
	// S, the bot's report of what the game told it, not a game-side counter. AreaTrail is the area of each entry.
	AreaChanges *int  `json:"area_changes,omitempty"`
	AreaTrail   []int `json:"area_trail,omitempty"`
	// WritesAfterStart is the number of gameplay-affecting writes after the run start (d2cheat ledger), -1 unknown.
	WritesAfterStart int `json:"writes_after_start"`
}

// StateItem is one carried item.
type StateItem struct {
	Code  string `json:"code"`
	Where string `json:"where"`
}

// StatePortal is one row of `debug portals`.
type StatePortal struct {
	Which int `json:"which"`
	Level int `json:"level"`
}

// StateObject is one level object.
type StateObject struct {
	Class int    `json:"class"`
	State string `json:"state"`
}

// Assertion kinds. Each reads one kind of fact from State.
const (
	AItemInPack     = "item_in_pack"   // code, min (default 1): items of that code in pack, belt, stash or body
	AItemAbsent     = "item_absent"    // code
	AQuestBits      = "quest_bits"     // record, set (mask all set), clear (mask all clear)
	AAreaIs         = "area_is"        // area
	AAreaIn         = "area_in"        // areas
	AGoldBetween    = "gold_between"   // min, max (purse)
	AStashBetween   = "stash_between"  // min, max
	ALevelAtLeast   = "level_at_least" // min
	ALifePctAtLeast = "life_pct_at_least"
	AObjectState    = "object_state"    // class, state: at least one object of the class has the state
	AAreaChangesLE  = "area_changes_le" // max
	APortalOpen     = "portal_open"     // area: a portal stands in that level (debug portals); needs a live observation
	ANoWrites       = "no_writes"       // writes_after_start == 0 (the CRUTCH RULE line)
)

// Assert is one assertion on game state.
type Assert struct {
	Kind   string `json:"kind"`
	Code   string `json:"code,omitempty"`
	Where  string `json:"where,omitempty"` // item_in_pack: pack (default: anywhere carried), stash
	Record int    `json:"record,omitempty"`
	Set    uint16 `json:"set,omitempty"`
	Clear  uint16 `json:"clear,omitempty"`
	Area   int    `json:"area,omitempty"`
	Areas  []int  `json:"areas,omitempty"`
	Class  int    `json:"class,omitempty"`
	State  string `json:"state,omitempty"`
	Min    int    `json:"min,omitempty"`
	Max    *int   `json:"max,omitempty"`
	Why    string `json:"why,omitempty"`
}

// Validate checks that the fields the kind needs are present.
func (a Assert) Validate() error {
	switch a.Kind {
	case AItemInPack, AItemAbsent:
		if a.Code == "" {
			return fmt.Errorf("%s needs code", a.Kind)
		}
	case AQuestBits:
		if a.Set == 0 && a.Clear == 0 {
			return fmt.Errorf("quest_bits needs set or clear")
		}
	case AAreaIs:
		if a.Area == 0 {
			return fmt.Errorf("area_is needs area")
		}
	case AAreaIn:
		if len(a.Areas) == 0 {
			return fmt.Errorf("area_in needs areas")
		}
	case AGoldBetween, AStashBetween:
		if a.Max == nil && a.Min == 0 {
			return fmt.Errorf("%s needs min or max", a.Kind)
		}
	case ALevelAtLeast, ALifePctAtLeast:
		if a.Min <= 0 {
			return fmt.Errorf("%s needs min", a.Kind)
		}
	case AObjectState:
		if a.Class == 0 || a.State == "" {
			return fmt.Errorf("object_state needs class and state")
		}
	case APortalOpen:
		if a.Area == 0 {
			return fmt.Errorf("portal_open needs area")
		}
	case AAreaChangesLE:
		if a.Max == nil {
			return fmt.Errorf("area_changes_le needs max")
		}
	case ANoWrites:
	default:
		return fmt.Errorf("unknown assertion kind %q", a.Kind)
	}
	return nil
}

// AssertResult is the outcome of one assertion.
type AssertResult struct {
	Kind       string `json:"kind"`
	OK         bool   `json:"ok"`
	Unobserved bool   `json:"unobserved,omitempty"`
	Want       string `json:"want"`
	Got        string `json:"got"`
	Why        string `json:"why,omitempty"`
}

// Evaluate checks every assertion against st.
func Evaluate(as []Assert, st State) []AssertResult {
	out := make([]AssertResult, 0, len(as))
	for _, a := range as {
		r := evalOne(a, st)
		r.Kind, r.Why = a.Kind, a.Why
		out = append(out, r)
	}
	return out
}

func unobs(want, what string) AssertResult {
	return AssertResult{Want: want, Got: what + " not observed", Unobserved: true}
}

func evalOne(a Assert, st State) AssertResult {
	switch a.Kind {
	case AItemInPack, AItemAbsent:
		n := 0
		for _, it := range st.Items {
			// no place named: anything carried (the floor is not carried)
			if it.Code == a.Code && (a.Where == "" && it.Where != "ground" || it.Where == a.Where) {
				n++
			}
		}
		if !st.has("items") || a.Where == "ground" && !st.has("ground") {
			return unobs("items of "+a.Code, "items")
		}
		if a.Kind == AItemAbsent {
			return AssertResult{OK: n == 0, Want: "no " + a.Code, Got: fmt.Sprintf("%d x %s", n, a.Code)}
		}
		min := a.Min
		if min == 0 {
			min = 1
		}
		return AssertResult{OK: n >= min, Want: fmt.Sprintf(">= %d x %s", min, a.Code), Got: fmt.Sprintf("%d x %s", n, a.Code)}
	case AQuestBits:
		w, ok := st.Quests[a.Record]
		want := fmt.Sprintf("record %d set 0x%04x clear 0x%04x", a.Record, a.Set, a.Clear)
		if !ok {
			return unobs(want, fmt.Sprintf("quest record %d", a.Record))
		}
		return AssertResult{OK: w&a.Set == a.Set && w&a.Clear == 0, Want: want, Got: fmt.Sprintf("0x%04x", w)}
	case AAreaIs:
		if st.Area == nil {
			return unobs(fmt.Sprintf("area %d", a.Area), "area")
		}
		return AssertResult{OK: *st.Area == a.Area, Want: fmt.Sprintf("area %d", a.Area), Got: fmt.Sprintf("area %d", *st.Area)}
	case AAreaIn:
		if st.Area == nil {
			return unobs(fmt.Sprintf("area in %v", a.Areas), "area")
		}
		ok := false
		for _, x := range a.Areas {
			ok = ok || x == *st.Area
		}
		return AssertResult{OK: ok, Want: fmt.Sprintf("area in %v", a.Areas), Got: fmt.Sprintf("area %d", *st.Area)}
	case AGoldBetween, AStashBetween:
		v, name := st.Gold, "gold"
		if a.Kind == AStashBetween {
			v, name = st.Stash, "stash"
		}
		want := fmt.Sprintf("%s >= %d", name, a.Min)
		if a.Max != nil {
			want += fmt.Sprintf(" and <= %d", *a.Max)
		}
		if v == nil {
			return unobs(want, name)
		}
		return AssertResult{OK: *v >= a.Min && (a.Max == nil || *v <= *a.Max), Want: want, Got: fmt.Sprintf("%s %d", name, *v)}
	case ALevelAtLeast:
		if st.Level == nil {
			return unobs(fmt.Sprintf("level >= %d", a.Min), "level")
		}
		return AssertResult{OK: *st.Level >= a.Min, Want: fmt.Sprintf("level >= %d", a.Min), Got: fmt.Sprintf("level %d", *st.Level)}
	case ALifePctAtLeast:
		if st.LifePct == nil {
			return unobs(fmt.Sprintf("life >= %d%%", a.Min), "life")
		}
		return AssertResult{OK: *st.LifePct >= a.Min, Want: fmt.Sprintf("life >= %d%%", a.Min), Got: fmt.Sprintf("life %d%%", *st.LifePct)}
	case AObjectState:
		want := fmt.Sprintf("object class %d state %s", a.Class, a.State)
		if !st.has("objects") {
			return unobs(want, "objects")
		}
		var seen []string
		for _, o := range st.Objects {
			if o.Class == a.Class {
				seen = append(seen, o.State)
				if o.State == a.State {
					return AssertResult{OK: true, Want: want, Got: want}
				}
			}
		}
		sort.Strings(seen)
		return AssertResult{Want: want, Got: fmt.Sprintf("class %d states %v", a.Class, seen)}
	case APortalOpen:
		want := fmt.Sprintf("a portal in area %d", a.Area)
		if !st.has("portals") {
			return unobs(want, "portals")
		}
		var seen []int
		for _, p := range st.Portals {
			seen = append(seen, p.Level)
			if p.Level == a.Area {
				return AssertResult{OK: true, Want: want, Got: want}
			}
		}
		return AssertResult{Want: want, Got: fmt.Sprintf("portals in areas %v", seen)}
	case AAreaChangesLE:
		want := fmt.Sprintf("area changes <= %d", *a.Max)
		if st.AreaChanges == nil {
			return unobs(want, "area change count")
		}
		return AssertResult{OK: *st.AreaChanges <= *a.Max, Want: want, Got: fmt.Sprintf("%d area changes", *st.AreaChanges)}
	case ANoWrites:
		if st.WritesAfterStart < 0 {
			return unobs("writes after start: 0", "write ledger")
		}
		return AssertResult{OK: st.WritesAfterStart == 0, Want: "writes after start: 0", Got: fmt.Sprintf("writes after start: %d", st.WritesAfterStart)}
	}
	return AssertResult{Want: a.Kind, Got: "unknown kind"}
}

func (s State) has(g string) bool {
	for _, x := range s.Has {
		if x == g {
			return true
		}
	}
	return false
}

// CountZoneEntries reads the bot's structured log (one JSON object per line) and returns the area of every `zone entry`
// line (internal/bot/lifelog.go). The count of area CHANGES is len-1: the first line is the arrival of the run. ok is
// false when the log holds no such line at all (a log that is not the bot's, or a run that never started): the fact is
// then not observed.
func CountZoneEntries(log []byte) (trail []int, ok bool) {
	for _, l := range bytes.Split(log, []byte("\n")) {
		if !bytes.Contains(l, []byte(`"zone entry"`)) {
			continue
		}
		var e struct {
			Msg  string `json:"msg"`
			Area any    `json:"area"`
		}
		if json.Unmarshal(l, &e) != nil || e.Msg != "zone entry" {
			continue
		}
		switch v := e.Area.(type) {
		case float64:
			trail = append(trail, int(v))
		case string:
			if n, err := strconv.Atoi(v); err == nil {
				trail = append(trail, n)
			}
		}
	}
	return trail, len(trail) > 0
}

// AreaChangeCount is the number of changes in a trail: entries after the first.
func AreaChangeCount(trail []int) int {
	if len(trail) == 0 {
		return 0
	}
	return len(trail) - 1
}
