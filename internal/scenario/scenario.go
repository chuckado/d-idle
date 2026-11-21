// Package scenario is the data model and the pure logic of the Proving Ground (docs/process/proving-ground.md): a
// scenario declares a minimal precondition (Pre), how the real bot is driven (Trigger, Budget) and what must hold in
// GAME STATE afterwards (Assert). Nothing here talks to a game: the runner (cmd/d2scenario) builds the state with the
// setup builder, plays it on a lockstep d2d slot and reads a State back; Evaluate turns that into a verdict.
package scenario

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"d2lod-bot/internal/leveler"
)

// Version is the scenario file format ("didle.scenario/1"); a file with another version is refused.
const Version = "didle.scenario/1"

// Tiers: fast is the merge-gate subset (seconds to a few virtual minutes each), full is every scenario times the
// class matrix (nightly, long lane), nightly is the act runs (later).
const (
	TierFast    = "fast"
	TierFull    = "full"
	TierNightly = "nightly"
)

// Kinds of scenario.
const (
	KindStep      = "step"      // one quest step or chunk alone, from its own precondition
	KindBehaviour = "behaviour" // something the bot does outside quests (chores, belt, recovery, flail, loot ...)
	KindE2E       = "e2e"       // a chain played end to end
)

// Start modes. "town": the hero is saved in town and the bot walks out itself (no gameplay write after the start, a
// save cannot hold a field position). "field": the runner warps the hero with a lab verb and hands the live d2d to the
// bot (attach); the position is declared SETUP and the run is CHEATED-position, mechanics only.
const (
	StartTown  = "town"
	StartField = "field"
)

// File is one scenario file: a list of scenarios sharing a version.
type File struct {
	Version   string     `json:"version"`
	Scenarios []Scenario `json:"scenarios"`
}

// Scenario is one case. Class and seed come from the matrix at run time, never from the file's pre-state builder.
type Scenario struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
	Tier  string `json:"tier"`
	// Status "pending" marks a skeleton: it is listed and counted but never run (PendingReason says what is missing).
	Status        string `json:"status,omitempty"`
	PendingReason string `json:"pending_reason,omitempty"`
	// Quest links a step scenario to the quest board data (data/questboard) so the arbiter's shadow mode and the
	// questspec checker can drive from the same declaration.
	Quest *QuestRef `json:"quest,omitempty"`
	// Step is the declared action and target; Pre and Assert are the declared pre and post facts of that step.
	Step   *StepDecl `json:"step,omitempty"`
	Covers []string  `json:"covers,omitempty"` // behaviour ids of behaviours.json
	// Touches are path prefixes of the code this scenario guards; the merge gate runs the fast scenarios whose prefixes
	// match the files a merge changed (empty: always part of the fast gate).
	Touches []string `json:"touches,omitempty"`
	Pre     Pre      `json:"pre"`
	Trigger Trigger  `json:"trigger"`
	Budget  Budget   `json:"budget"`
	Assert  []Assert `json:"assert"`
	Matrix  Matrix   `json:"matrix"`
	// Needs lists capabilities the d2d build must have (verbs, dump keys). A missing one makes the run BLOCKED, not red.
	Needs []string `json:"needs,omitempty"`
	// D2D selects the build: "lab" (default) or "sandbox" (the shipped build of the friend sandbox, env D2D_BIN_SANDBOX).
	Build string `json:"build,omitempty"`
	// Provenance is the text for docs/probes/test-characters.md: what is CHEATED or IMPOSSIBLE about the setup.
	Provenance string `json:"provenance,omitempty"`
	Notes      string `json:"notes,omitempty"`
}

// QuestRef names a chunk of data/questboard.
type QuestRef struct {
	Quest string `json:"quest"`
	Chunk string `json:"chunk"`
}

// StepDecl is the action/target declaration shared with the arbiter and the questspec checker.
type StepDecl struct {
	Action string   `json:"action"` // interact, talk, travel, kill, chore ...
	Target string   `json:"target"`
	Nodes  []string `json:"nodes,omitempty"` // task-layer node ids this scenario exercises
}

// Pre is the declared precondition. Every field is optional; the builder applies only what is set.
type Pre struct {
	Start      string    `json:"start,omitempty"` // town (default) or field
	Area       int       `json:"area,omitempty"`  // levels.txt id, only for start=field
	Level      int       `json:"level,omitempty"`
	Gold       int       `json:"gold,omitempty"`
	Stash      int       `json:"stash,omitempty"`
	Difficulty string    `json:"difficulty,omitempty"` // normal (default), nightmare, hell
	Waypoints  []int     `json:"waypoints,omitempty"`  // waypoint bits to activate (internal/saves waypointTable order)
	Quests     []QWord   `json:"quests,omitempty"`     // quest records written as the hero's own state
	Board      *QuestRef `json:"board,omitempty"`      // chunk the player accepted (the plan store)
	Items      []Item    `json:"items,omitempty"`
	Belt       []string  `json:"belt,omitempty"` // potion codes in belt order
	Merc       *int      `json:"merc,omitempty"` // hireling.txt Id, nil: none
	LifePct    int       `json:"life_pct,omitempty"`
}

// QWord is one quest record (fork layout: act 1 records 0..6, docs/forkcheats.md) and its 16-bit word.
type QWord struct {
	Record int    `json:"record"`
	Word   uint16 `json:"word"`
}

// Item is one item given to the pack (or stash, belt, body) by a declared SETUP verb.
type Item struct {
	Code  string `json:"code"`
	Where string `json:"where,omitempty"` // pack (default), ground (at the hero's feet), stash, body
	Count int    `json:"count,omitempty"`
	// Quality is the item quality of the `debug give` verb (4 magic, 6 rare: docs/forkcheats.md, V); 0 leaves the code's default.
	Quality int `json:"quality,omitempty"`
	// Unid gives the item unidentified (verb flag `unid`).
	Unid bool `json:"unid,omitempty"`
}

// Trigger says how the real bot is driven. Goals is the bot's --goals value; Knobs are tuning overrides.
type Trigger struct {
	Goals string         `json:"goals,omitempty"`
	Knobs map[string]any `json:"knobs,omitempty"`
	// Build is the leveler build (stat and skill table) the hero plays by name, for example "sorceress_fire_normal30": the
	// runner writes it as levelerBuild into the bot's config file. Empty keeps the class default. It must be a table of
	// the class of every matrix class (Validate).
	Build string `json:"build,omitempty"`
	// Driver plays the part of the person: "" (none) or "board-taps", which runs the bot like the friend sandbox does
	// (--ladder --watch, the quest board is built) and taps Start and Hand in on the board of Quest (V docs/probes/cainl3/poll.py).
	Driver string `json:"driver,omitempty"`
	// TapQuest names the quest whose board is tapped; empty means the scenario's own quest (a behaviour scenario names it).
	TapQuest string `json:"tap_quest,omitempty"`
	// TapChunks limits the taps to these chunks of the scenario's quest (empty: every chunk, a whole chain).
	TapChunks []string `json:"tap_chunks,omitempty"`
}

// TappedQuest is the quest id the board driver taps ("" when the scenario has no driver).
func (s Scenario) TappedQuest() string {
	switch {
	case s.Trigger.Driver != DriverBoardTaps:
		return ""
	case s.Trigger.TapQuest != "":
		return s.Trigger.TapQuest
	}
	return s.Quest.Quest
}

// DriverBoardTaps is the Trigger.Driver value that taps the quest board.
const DriverBoardTaps = "board-taps"

// Budget bounds the run in VIRTUAL minutes (the lockstep clock); WallMin is a safety net in wall minutes.
type Budget struct {
	VirtualMin float64 `json:"virtual_min"`
	WallMin    int     `json:"wall_min,omitempty"`
}

// Matrix is the class and seed set of the full tier; the fast tier uses the first class and first seed unless the
// run names them. Classes are names (Paladin, Sorceress ...), resolved by ClassIndex.
type Matrix struct {
	Classes []string `json:"classes"`
	Seeds   []int64  `json:"seeds"`
}

// classIndex is d2d's UI-order index for a NEW character (internal/d2dinput.Spec; V 1 Necromancer, 2 Paladin, 4
// Sorceress; the others follow the character screen order and are U).
var classIndex = map[string]int{"Amazon": 0, "Necromancer": 1, "Paladin": 2, "Barbarian": 3, "Sorceress": 4, "Druid": 5, "Assassin": 6}

// ClassIndex returns the d2d class index or an error naming the class (never a silent Sorceress).
func ClassIndex(name string) (int, error) {
	i, ok := classIndex[name]
	if !ok {
		return 0, fmt.Errorf("scenario: unknown class %q", name)
	}
	return i, nil
}

// LoadFile reads and validates one scenario file; unknown fields are an error (a typo must not silently drop a fact).
func LoadFile(path string) (File, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	var f File
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return File{}, fmt.Errorf("%s: %w", path, err)
	}
	if f.Version != Version {
		return File{}, fmt.Errorf("%s: version %q, want %q", path, f.Version, Version)
	}
	for i := range f.Scenarios {
		f.Scenarios[i].applyDefaults()
		if err := f.Scenarios[i].Validate(); err != nil {
			return File{}, fmt.Errorf("%s: %w", path, err)
		}
	}
	return f, nil
}

// LoadDir loads every *.json scenario file of dir (sorted) except behaviours.json, and rejects duplicate ids.
func LoadDir(dir string) ([]Scenario, error) {
	names, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	var all []Scenario
	seen := map[string]string{}
	for _, n := range names {
		if filepath.Base(n) == "behaviours.json" {
			continue
		}
		f, err := LoadFile(n)
		if err != nil {
			return nil, err
		}
		for _, s := range f.Scenarios {
			if prev, dup := seen[s.ID]; dup {
				return nil, fmt.Errorf("scenario id %q in %s and %s", s.ID, prev, n)
			}
			seen[s.ID] = n
			all = append(all, s)
		}
	}
	return all, nil
}

func (s *Scenario) applyDefaults() {
	if s.Pre.Start == "" {
		s.Pre.Start = StartTown
	}
	if s.Pre.Difficulty == "" {
		s.Pre.Difficulty = "normal"
	}
	if s.Build == "" {
		s.Build = "lab"
	}
	if s.Trigger.Goals == "" {
		s.Trigger.Goals = "campaign"
	}
	if s.Budget.WallMin == 0 {
		s.Budget.WallMin = int(s.Budget.VirtualMin/5*1.6) + 10 // abrun.sh: 5x is the slow end of the measured 8x to 60x
	}
}

// Validate checks the scenario alone (cross-file checks are in Check).
func (s Scenario) Validate() error {
	at := "scenario " + s.ID
	if !validID(s.ID) {
		return fmt.Errorf("scenario id %q: lower case letters, digits, '.', '-' and '_' only", s.ID)
	}
	if s.Title == "" {
		return fmt.Errorf("%s: no title", at)
	}
	switch s.Kind {
	case KindStep, KindBehaviour, KindE2E:
	default:
		return fmt.Errorf("%s: kind %q", at, s.Kind)
	}
	switch s.Tier {
	case TierFast, TierFull, TierNightly:
	default:
		return fmt.Errorf("%s: tier %q", at, s.Tier)
	}
	if s.Status != "" && s.Status != "active" && s.Status != "pending" {
		return fmt.Errorf("%s: status %q", at, s.Status)
	}
	if s.Pending() && s.PendingReason == "" {
		return fmt.Errorf("%s: pending without pending_reason", at)
	}
	if s.Kind == KindStep && (s.Quest == nil || s.Step == nil) {
		return fmt.Errorf("%s: a step scenario declares quest and step", at)
	}
	switch s.Trigger.Driver {
	case "":
	case DriverBoardTaps:
		if s.Quest == nil && s.Trigger.TapQuest == "" {
			return fmt.Errorf("%s: driver board-taps needs a quest or trigger.tap_quest", at)
		}
		if s.Pre.Board != nil {
			return fmt.Errorf("%s: driver board-taps accepts the chunk itself: drop pre.board", at)
		}
	default:
		return fmt.Errorf("%s: trigger.driver %q", at, s.Trigger.Driver)
	}
	if s.Trigger.Build != "" {
		lb, err := leveler.LoadBuild(s.Trigger.Build, "")
		if err != nil {
			return fmt.Errorf("%s: trigger.build: %w", at, err)
		}
		for _, c := range s.Matrix.Classes {
			if !strings.EqualFold(c, lb.Class) {
				return fmt.Errorf("%s: trigger.build %q is a %s table, matrix class %s", at, s.Trigger.Build, lb.Class, c)
			}
		}
	}
	if s.Pre.Start != StartTown && s.Pre.Start != StartField {
		return fmt.Errorf("%s: pre.start %q", at, s.Pre.Start)
	}
	if s.Pre.Start == StartField && s.Pre.Area == 0 {
		return fmt.Errorf("%s: pre.start field needs pre.area", at)
	}
	if s.Build != "lab" && s.Build != "sandbox" {
		return fmt.Errorf("%s: d2d %q", at, s.Build)
	}
	if s.Budget.VirtualMin <= 0 {
		return fmt.Errorf("%s: budget.virtual_min must be positive (every scenario has a timeout)", at)
	}
	if len(s.Matrix.Classes) == 0 || len(s.Matrix.Seeds) == 0 {
		return fmt.Errorf("%s: matrix needs classes and seeds", at)
	}
	for _, c := range s.Matrix.Classes {
		if _, err := ClassIndex(c); err != nil {
			return fmt.Errorf("%s: %w", at, err)
		}
	}
	if !s.Pending() && len(s.Assert) == 0 {
		return fmt.Errorf("%s: no assertions", at)
	}
	for i, a := range s.Assert {
		if err := a.Validate(); err != nil {
			return fmt.Errorf("%s: assert %d: %w", at, i, err)
		}
	}
	if s.Pre.Level < 0 || s.Pre.Level > 99 {
		return fmt.Errorf("%s: pre.level %d", at, s.Pre.Level)
	}
	return nil
}

// Pending reports a skeleton.
func (s Scenario) Pending() bool { return s.Status == "pending" }

func validID(id string) bool {
	if id == "" {
		return false
	}
	return strings.IndexFunc(id, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '.' || r == '-' || r == '_')
	}) < 0
}
