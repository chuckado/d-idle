package scenario

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
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
	// RightSkill and LeftSkill are the skill ids on the two hands (the lskill and rskill keys of the state verb; the
	// active aura of a Paladin is the right hand).
	RightSkill *int `json:"right_skill,omitempty"`
	LeftSkill  *int `json:"left_skill,omitempty"`
	// Quests maps the fork's quest record index to its 16-bit word (docs/forkcheats.md).
	Quests map[int]uint16 `json:"quests,omitempty"`
	// Items lists every item the hero carries: Code is the base code (bks, bkd), Where pack, stash, belt, body, cursor.
	Items []StateItem `json:"items,omitempty"`
	// Objects are the level objects read with debug objects: class and state (NU, ON ...).
	Objects []StateObject `json:"objects,omitempty"`
	// Stats are the hero's points as the dump reports them: str, dex, vit, ene (base, no items), statpts and skillpts (unspent)
	// and skillsspent (the sum of the hard skill points). Absent keys were not observed.
	Stats map[string]int `json:"stats,omitempty"`
	// Portals are the open portals of `debug portals` (which, level the portal stands in).
	Portals []StatePortal `json:"portals,omitempty"`
	// AreaChanges counts area changes during the run when the observer could count them (nil: not observable). It is the
	// number of `zone entry` lines of the bot's own log after the first (one line per area change, internal/bot/lifelog.go):
	// S, the bot's report of what the game told it, not a game-side counter. AreaTrail is the area of each entry.
	AreaChanges *int  `json:"area_changes,omitempty"`
	AreaTrail   []int `json:"area_trail,omitempty"`
	// Class is the hero's class name of the run (set by the runner), the key of the base skill id of SkillLevels.
	Class string `json:"class,omitempty"`
	// SkillLevels is the dump's skills.base_levels (hard points, class order; the entry i is skill id ClassSkillBase+i, S
	// internal/d2dsrc/tables.go). Absent when the dump has none.
	SkillLevels []int `json:"skill_levels,omitempty"`
	// LogLines are the lines of the bot's log that match a log_line assertion of the scenario (never the whole log), each
	// with its place in the run as virtual minutes (LogLine.T). They come from the bot's log, not from the game: S.
	LogLines []LogLine `json:"log_lines,omitempty"`
	// Board are samples of GET /api/quests during the run and APIReplies the recorded answers of trigger.api (board.go).
	Board      []BoardSample `json:"board,omitempty"`
	APIReplies []APIReply    `json:"api_replies,omitempty"`
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
	AItemInPack     = "item_in_pack"   // code, min (default 1; 0 when max is set), max: items of that code in pack, belt, stash or body
	AItemAbsent     = "item_absent"    // code
	AQuestBits      = "quest_bits"     // record, set (mask all set), clear (mask all clear)
	AAreaIs         = "area_is"        // area
	AAreaIn         = "area_in"        // areas
	AGoldBetween    = "gold_between"   // min, max (purse)
	AStashBetween   = "stash_between"  // min, max
	AWealthBetween  = "wealth_between" // min, max (purse plus stash: a sale and the bank chore move gold between them)
	ALevelAtLeast   = "level_at_least" // min
	ALifePctAtLeast = "life_pct_at_least"
	AObjectState    = "object_state"    // class, state: at least one object of the class has the state
	AAreaChangesLE  = "area_changes_le" // max
	APortalOpen     = "portal_open"     // area: a portal stands in that level (debug portals); needs a live observation
	ARightSkillIs   = "right_skill_is"  // skill (the id on the right hand: the aura, the select verb's effect)
	ALeftSkillIs    = "left_skill_is"   // skill
	ANoWrites       = "no_writes"       // writes_after_start == 0 (the CRUTCH RULE line)
	AStatBetween    = "stat_between"    // code (str dex vit ene statpts skillpts skillsspent), min, max
	ASkillPoints    = "skill_points"    // skill (skills.txt id), min, max, at_level: hard points in one skill (skills.go)
	ABeltAtLeast    = "belt_at_least"   // code (potion family prefix: hp, mp, rv), min: potions of that family in the BELT alone (item_in_pack cannot tell belt from pack when where is unset)
	ALogLine        = "log_line"        // pattern, min (default 1), max, from_min, to_min: lines of the bot log (logline.go)
)

var statKeys = map[string]bool{"str": true, "dex": true, "vit": true, "ene": true, "statpts": true, "skillpts": true, "skillsspent": true}

// Assert is one assertion on game state.
type Assert struct {
	Kind   string   `json:"kind"`
	Code   string   `json:"code,omitempty"`
	Or     []string `json:"or,omitempty"`    // item_in_pack: further codes that count as the same item (bks becomes bkd once Akara reads it)
	Where  string   `json:"where,omitempty"` // item_in_pack: pack (default: anywhere carried), stash
	Record int      `json:"record,omitempty"`
	Set    uint16   `json:"set,omitempty"`
	Clear  uint16   `json:"clear,omitempty"`
	Area   int      `json:"area,omitempty"`
	Skill  int      `json:"skill,omitempty"` // right_skill_is, left_skill_is: skills.txt id
	Areas  []int    `json:"areas,omitempty"`
	Class  int      `json:"class,omitempty"`
	State  string   `json:"state,omitempty"`
	Min    int      `json:"min,omitempty"`
	Max    *int     `json:"max,omitempty"`
	Why    string   `json:"why,omitempty"`
	// OnlyClass limits the assertion to one class of the matrix (other classes skip it); any kind.
	OnlyClass string `json:"only_class,omitempty"`
	// AtLevel (skill_points): the hero must be exactly at this character level when observed, else the fact is UNOBSERVED
	// (the points of a table are only defined at a level). 0: any level.
	AtLevel int `json:"at_level,omitempty"`
	// Pattern (log_line) is a Go regexp matched against `<msg> k=v k=v ...` of each bot log line; FromMin and ToMin
	// bound the window in virtual minutes of the run (ToMin 0: to the end).
	Pattern string  `json:"pattern,omitempty"`
	FromMin float64 `json:"from_min,omitempty"`
	ToMin   float64 `json:"to_min,omitempty"`
	// board, board_heads_le and api_reply (board.go): the chunk, the reason alternatives and the reply fields.
	Quest      string   `json:"quest,omitempty"`
	Chunk      string   `json:"chunk,omitempty"`
	Reason     string   `json:"reason,omitempty"`
	ReasonOr   []string `json:"reason_or,omitempty"`
	When       string   `json:"when,omitempty"`
	Index      *int     `json:"index,omitempty"`
	Status     int      `json:"status,omitempty"`
	OKIs       *bool    `json:"ok,omitempty"`
	AlreadyIs  *bool    `json:"already,omitempty"`
	ChangedHas []string `json:"changed_has,omitempty"`
}

// Validate checks that the fields the kind needs are present.
func (a Assert) Validate() error {
	switch a.Kind {
	case AItemInPack, AItemAbsent:
		if a.Code == "" {
			return fmt.Errorf("%s needs code", a.Kind)
		}
	case ABeltAtLeast:
		if a.Code == "" || a.Min <= 0 {
			return fmt.Errorf("belt_at_least needs code (hp, mp or rv) and min")
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
	case AGoldBetween, AStashBetween, AWealthBetween:
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
	case ARightSkillIs, ALeftSkillIs:
		if a.Skill <= 0 {
			return fmt.Errorf("%s needs skill", a.Kind)
		}
	case AStatBetween:
		if !statKeys[a.Code] {
			return fmt.Errorf("stat_between needs code str, dex, vit, ene, statpts, skillpts or skillsspent")
		}
		if a.Max == nil && a.Min == 0 {
			return fmt.Errorf("stat_between needs min or max")
		}
	case ASkillPoints:
		if a.Skill <= 0 {
			return fmt.Errorf("skill_points needs skill (skills.txt id)")
		}
		if a.Max == nil && a.Min == 0 {
			return fmt.Errorf("skill_points needs min or max")
		}
	case ALogLine:
		if err := a.validateLogLine(); err != nil {
			return err
		}
	case ABoard, ABoardHeadsLE, ABoardReasons, AAPIReply:
		return a.validateBoard()
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
		if a.OnlyClass != "" && !strings.EqualFold(a.OnlyClass, st.Class) {
			continue
		}
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
			if (it.Code == a.Code || contains(a.Or, it.Code)) && (a.Where == "" && it.Where != "ground" || it.Where == a.Where) {
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
		if min == 0 && a.Max == nil {
			min = 1
		}
		if a.Max != nil { // an upper bound (a hoard check): min stays as given, 0 by default
			return AssertResult{OK: n >= min && n <= *a.Max, Want: fmt.Sprintf("%d..%d x %s", min, *a.Max, a.Code), Got: fmt.Sprintf("%d x %s", n, a.Code)}
		}
		return AssertResult{OK: n >= min, Want: fmt.Sprintf(">= %d x %s", min, a.Code), Got: fmt.Sprintf("%d x %s", n, a.Code)}
	case ABeltAtLeast:
		if !st.has("items") {
			return unobs("belt "+a.Code, "items")
		}
		n := 0
		for _, it := range st.Items {
			if it.Where == "belt" && len(it.Code) == 3 && strings.HasPrefix(it.Code, a.Code) {
				n++
			}
		}
		return AssertResult{OK: n >= a.Min, Want: fmt.Sprintf(">= %d %s potions in the belt", a.Min, a.Code), Got: fmt.Sprintf("%d in the belt", n)}
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
	case AGoldBetween, AStashBetween, AWealthBetween:
		v, name := st.Gold, "gold"
		switch {
		case a.Kind == AStashBetween:
			v, name = st.Stash, "stash"
		case a.Kind == AWealthBetween && st.Gold != nil && st.Stash != nil:
			w := *st.Gold + *st.Stash
			v, name = &w, "purse and stash"
		case a.Kind == AWealthBetween:
			v, name = nil, "purse and stash"
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
	case AStatBetween:
		want := fmt.Sprintf("%s >= %d", a.Code, a.Min)
		if a.Max != nil {
			want += fmt.Sprintf(" and <= %d", *a.Max)
		}
		v, ok := st.Stats[a.Code]
		if !ok {
			return unobs(want, a.Code)
		}
		return AssertResult{OK: v >= a.Min && (a.Max == nil || v <= *a.Max), Want: want, Got: fmt.Sprintf("%s %d", a.Code, v)}
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
	case ARightSkillIs, ALeftSkillIs:
		v, hand := st.RightSkill, "right"
		if a.Kind == ALeftSkillIs {
			v, hand = st.LeftSkill, "left"
		}
		want := fmt.Sprintf("%s skill %d", hand, a.Skill)
		if v == nil {
			return unobs(want, hand+" skill")
		}
		return AssertResult{OK: *v == a.Skill, Want: want, Got: fmt.Sprintf("%s skill %d", hand, *v)}
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
	case ASkillPoints:
		return evalSkillPoints(a, st)
	case ALogLine:
		return evalLogLine(a, st)
	case ABoard:
		return evalBoard(a, st)
	case ABoardHeadsLE:
		return evalBoardHeads(a, st)
	case ABoardReasons:
		return evalBoardReasons(st)
	case AAPIReply:
		return evalAPIReply(a, st)
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

func contains(l []string, c string) bool {
	for _, x := range l {
		if x == c {
			return true
		}
	}
	return false
}
