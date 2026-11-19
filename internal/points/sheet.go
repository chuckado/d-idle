// Package points is the player's stat and skill point desk (docs/points.md, design review 2026-11-19 04 section 7 P0):
// the points sheet (what the hero has, what a point would do, what the worn gear needs), the legality of one more point,
// the spend through the d2d client-message verbs with a read-back, the per-hero autopilot switch (default off: points are
// banked and the player is asked) and the audit trail of every action that came through the API. It never decides
// anything: nothing here spends a point without an explicit player action or an autopilot the player turned on.
package points

import (
	"fmt"
	"strings"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
)

// Schema is the version of the JSON the sheet is served as.
const Schema = "didle.points/1"

// StatDef names one of the four attributes. ID is the stat id the d2d `cmd stat` verb takes (V: dump_head_late.json
// raw[0..3] = str, ene, dex, vit; fork apps/d2d/panels.hpp kStatButtons {0 str, 2 dex, 3 vit, 1 ene}).
type StatDef struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// Stats is the attribute list in the character panel's order.
var Stats = [4]StatDef{{"str", "Strength", 0}, {"dex", "Dexterity", 2}, {"vit", "Vitality", 3}, {"ene", "Energy", 1}}

// StatByKey finds an attribute by its key ("str", "dex", "vit", "ene").
func StatByKey(key string) (StatDef, bool) {
	for _, s := range Stats {
		if s.Key == strings.ToLower(key) {
			return s, true
		}
	}
	return StatDef{}, false
}

// Verbs says whether the backend can spend a point by a protocol verb (no clicks). Via names the route in words.
type Verbs struct {
	Stat  bool   `json:"stat"`
	Skill bool   `json:"skill"`
	Via   string `json:"via"`
	Why   string `json:"why,omitempty"` // when not available: what to do about it
}

// GearReq is what one worn item asks of the hero. The bot never equips (docs/gear-policy.md); the sheet only shows it.
type GearReq struct {
	Slot string `json:"slot"`
	Item string `json:"item"`
	Str  int    `json:"str,omitempty"`
	Dex  int    `json:"dex,omitempty"`
	// ShortStr and ShortDex are how many points the hero lacks for the item (0 when met).
	ShortStr int `json:"shortStr,omitempty"`
	ShortDex int `json:"shortDex,omitempty"`
}

// Effect is what one point of an attribute gives, from the class table (charstats.txt, quarter points) and the game's
// fixed rules. Zero means "not this attribute".
type Effect struct {
	Life    float64 `json:"life,omitempty"`
	Mana    float64 `json:"mana,omitempty"`
	Stamina float64 `json:"stamina,omitempty"`
	AR      float64 `json:"ar,omitempty"`      // attack rating, +5 per Dexterity (S: community-known rule; the class ToHitFactor is the base, not the slope)
	Defence float64 `json:"defence,omitempty"` // +0.25 per Dexterity (S)
	// Notes are the effects the tables do not give a number for (block needs the shield, damage needs the weapon).
	Notes []string `json:"notes,omitempty"`
}

// StatRow is one attribute of the sheet.
type StatRow struct {
	StatDef
	Now       int    `json:"now"`       // as the game shows it (base plus item bonus where the backend cannot tell them apart)
	Base      int    `json:"base"`      // base points only; equals Now when the backend gives no split
	ClassBase int    `json:"classBase"` // the class's starting value
	Spent     int    `json:"spent"`     // Base - ClassBase
	Effect    Effect `json:"effect"`
	Preview   string `json:"preview"`             // the one line for the + button
	NeedsGear int    `json:"needsGear,omitempty"` // the highest requirement among the worn items
	CanAdd    bool   `json:"canAdd"`
	Why       string `json:"why,omitempty"` // when CanAdd is false
}

// Sheet is the whole points page for one hero.
type Sheet struct {
	Schema      string    `json:"schema"`
	Enabled     bool      `json:"enabled"` // resource.pointsApi
	Hero        string    `json:"hero,omitempty"`
	Class       string    `json:"class,omitempty"`
	Level       int       `json:"level,omitempty"`
	StatPoints  int       `json:"statPoints"`
	SkillPoints int       `json:"skillPoints"`
	Banked      int       `json:"banked"`
	Waiting     string    `json:"waiting,omitempty"` // "3 stat points and 2 skill points waiting for you"
	Autopilot   bool      `json:"autopilot"`
	Verbs       Verbs     `json:"verbs"`
	Stats       []StatRow `json:"stats"`
	Gear        []GearReq `json:"gear,omitempty"`
	// SkillsPage is where the skill tree with the per-skill verdicts lives.
	SkillsPage string `json:"skillsPage"`
	Note       string `json:"note,omitempty"` // the honest line when the sheet is partial
}

// Input is everything BuildSheet reads; the caller (cmd/d2bot/pointswire.go) fills it from the newest snapshot.
type Input struct {
	Hero      string
	Player    *d2.Player
	Class     *data.CharClass
	Base      *[4]int // str, dex, vit, ene without items, nil when the backend cannot split them
	Gear      []GearReq
	Verbs     Verbs
	Autopilot bool
	Enabled   bool
}

// BuildSheet is pure: no I/O, the same input gives the same sheet.
func BuildSheet(in Input) Sheet {
	s := Sheet{Schema: Schema, Enabled: in.Enabled, Hero: in.Hero, Autopilot: in.Autopilot, Verbs: in.Verbs, SkillsPage: "/watch/skills", Stats: []StatRow{}}
	p := in.Player
	if p == nil {
		s.Note = "the hero is not in a game yet"
		return s
	}
	s.Class, s.Level = p.Class.String(), p.Level
	s.StatPoints, s.SkillPoints = max(p.StatPoints, 0), max(p.SkillPoints, 0)
	s.Banked = s.StatPoints + s.SkillPoints
	s.Waiting = waiting(s.StatPoints, s.SkillPoints)
	s.Gear = shortages(in.Gear, p)
	if in.Class == nil {
		s.Note = "no class table for " + p.Class.String() + ": per-point effects are not shown"
	}
	now := [4]int{p.Strength, p.Dexterity, p.Vitality, p.Energy}
	for i, d := range Stats {
		r := StatRow{StatDef: d, Now: now[i], Base: now[i]}
		if in.Base != nil {
			r.Base = in.Base[i]
		} else if s.Note == "" {
			s.Note = "the backend cannot split base points from item bonuses: Spent includes gear"
		}
		if in.Class != nil {
			r.ClassBase = classBase(in.Class, d.Key)
			r.Spent = max(r.Base-r.ClassBase, 0)
		}
		r.Effect = EffectOf(in.Class, d.Key)
		r.Preview = Preview(d, r.Effect)
		for _, g := range s.Gear {
			r.NeedsGear = max(r.NeedsGear, map[string]int{"str": g.Str, "dex": g.Dex}[d.Key])
		}
		r.CanAdd, r.Why = StatLegal(s.StatPoints, in.Verbs)
		s.Stats = append(s.Stats, r)
	}
	return s
}

func classBase(c *data.CharClass, key string) int {
	switch key {
	case "str":
		return c.Str
	case "dex":
		return c.Dex
	case "vit":
		return c.Vit
	}
	return c.Int
}

// EffectOf is what one point of the attribute does for the class. charstats.txt holds the gains in quarter points
// (V: internal/cheat/memory_ops.go applies LifePerVit*256/4 per point; its test Sorceress has LifePerVit 8 = 2 life).
func EffectOf(c *data.CharClass, key string) Effect {
	var e Effect
	q := func(v int) float64 { return float64(v) / 4 }
	switch key {
	case "str":
		e.Notes = []string{"worn gear requirements", "damage of weapons that scale with Strength"}
	case "dex":
		e.AR, e.Defence = 5, 0.25
		e.Notes = []string{"chance to block, with a shield"}
	case "vit":
		if c != nil {
			e.Life, e.Stamina = q(c.LifePerVit), q(c.StaminaPerVit)
		}
	case "ene":
		if c != nil {
			e.Mana = q(c.ManaPerEnergy)
		}
	}
	return e
}

// Preview is the one line shown next to a + button.
func Preview(d StatDef, e Effect) string {
	var parts []string
	add := func(v float64, unit string) {
		if v != 0 {
			parts = append(parts, fmt.Sprintf("+%s %s", trim(v), unit))
		}
	}
	add(e.Life, "life")
	add(e.Mana, "mana")
	add(e.Stamina, "stamina")
	add(e.AR, "attack rating")
	add(e.Defence, "defense")
	if len(parts) == 0 {
		parts = append(parts, strings.Join(e.Notes, ", "))
	}
	return "+1 " + d.Name + ": " + strings.Join(parts, ", ")
}

func trim(v float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", v), "0"), ".")
}

func waiting(stat, skill int) string {
	var parts []string
	if stat > 0 {
		parts = append(parts, plural(stat, "stat point"))
	}
	if skill > 0 {
		parts = append(parts, plural(skill, "skill point"))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " and ") + " waiting for you"
}

func plural(n int, w string) string {
	if n == 1 {
		return "1 " + w
	}
	return fmt.Sprintf("%d %ss", n, w)
}

// Waiting is the sentence for the Needs-you row and the return screen, "" when nothing is banked.
func Waiting(stat, skill int) string { return waiting(stat, skill) }

func shortages(g []GearReq, p *d2.Player) []GearReq {
	out := make([]GearReq, 0, len(g))
	for _, r := range g {
		r.ShortStr, r.ShortDex = max(r.Str-p.Strength, 0), max(r.Dex-p.Dexterity, 0)
		out = append(out, r)
	}
	return out
}

// StatLegal says whether one more point can go into an attribute now.
func StatLegal(statPoints int, v Verbs) (bool, string) {
	switch {
	case statPoints < 1:
		return false, "no unspent stat points"
	case !v.Stat:
		return false, "this backend has no stat verb: " + v.Why
	}
	return true, ""
}
