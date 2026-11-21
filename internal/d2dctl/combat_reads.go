package d2dctl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// The combat READ verbs of the private fork (dlv-reads, ADR 0026 decision Q4 and Q5; docs/process/combat-reads.md): `debug minions`,
// `debug skillstate`, `debug combatlog [n]`. They are reads under ADR 0022 (never CHEATED), flag-gated in the fork (`--read-combat`,
// implied by `--lab-verbs`) and listed in `caps`. Every row says which of its fields a player can see (Visible) and which are
// test-only numbers (Hidden): the hero decides on visible fields only, scenarios and expectations may use both.
const (
	CapReadMinions      = "read.minions"
	CapReadSkillState   = "read.skillstate"
	CapReadCombatEvents = "read.combat-events"
)

// ErrCombatRead is wrapped by every failure of the three verbs; the message names the verb.
var ErrCombatRead = errors.New("d2dctl: combat read")

func combatReadErr(verb string, err error) error {
	return fmt.Errorf("%w: debug %s: %w", ErrCombatRead, verb, err)
}

// Minion is one summon of the hero (golem, skeleton, mage, revive, raven, wolf, vine, hydra, totem, decoy, shadow; never the
// hireling). X and Y are cells as `monsters`. Skill is the Skills.txt id that made it. TTL is the ticks of 40 ms left of a
// timed summon, -1 for one that never ends; it is HIDDEN (the screen shows no timer). Everything else is visible.
type Minion struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Owner   int      `json:"owner"`
	Skill   int      `json:"skill"`
	Life    int      `json:"life"`
	MaxLife int      `json:"maxlife"`
	X       float64  `json:"x"`
	Y       float64  `json:"y"`
	Mode    string   `json:"mode"`
	Here    bool     `json:"here"`
	TTL     int      `json:"ttl"`
	Visible bool     `json:"visible"`
	Hidden  []string `json:"hidden"`
}

// IsHidden reports whether the named field is a test-only read.
func (m Minion) IsHidden(field string) bool { return inList(m.Hidden, field) }

// Minions is the reply of `debug minions`.
type Minions struct {
	Tick    int      `json:"tick"`
	Hero    int      `json:"hero"`
	Minions []Minion `json:"minions"`
}

// SkillRow is one skill of the hero in `debug skillstate`. Base is the points, Level the level with gear, Mana the whole mana
// points of the next cast, Cooldown the ticks of 40 ms until the skill's own delay allows another cast (0 ready), Usable
// whether the engine would start it now, Left/Right whether it is on that button, Charges the charge-ups held. Status is the
// engine's refusal code (0 ok, 1 mana, 2 weapon, 3 no row, 4 form, 5 passive, 6 aura, 7 no level, 8 delay) and is HIDDEN.
type SkillRow struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Base     int      `json:"base"`
	Level    int      `json:"level"`
	Mana     int      `json:"mana"`
	Cooldown int      `json:"cooldown"`
	Usable   bool     `json:"usable"`
	Left     bool     `json:"left"`
	Right    bool     `json:"right"`
	Charges  int      `json:"charges"`
	Status   int      `json:"status"`
	Visible  []string `json:"visible"`
	Hidden   []string `json:"hidden"`
}

// IsHidden reports whether the named field is a test-only read.
func (r SkillRow) IsHidden(field string) bool { return inList(r.Hidden, field) }

// SkillState is the reply of `debug skillstate`. Left and Right are the skill ids on the buttons.
type SkillState struct {
	Tick   int        `json:"tick"`
	Class  int        `json:"class"`
	Left   int        `json:"left"`
	Right  int        `json:"right"`
	Skills []SkillRow `json:"skills"`
}

// Skill returns the row of a skill id.
func (s *SkillState) Skill(id int) (SkillRow, bool) {
	for _, r := range s.Skills {
		if r.ID == id {
			return r, true
		}
	}
	return SkillRow{}, false
}

// CombatEvent is one journal line of `debug combatlog`. Kind is cast, refuse, hit, miss, block, kill, hurt or missile_hit.
// Source and Target are unit ids (-1 none), Skill a Skills.txt id (-1 none), ByHero whether the hero is the source. Amount and
// Detail are numbers the screen never shows (HIDDEN): cast mana paid and level, refuse the engine status, hit damage and the
// target's life left, kill 1 in Detail for the hero's credit, hurt damage taken and blows landed, missile_hit the Missiles.txt id.
type CombatEvent struct {
	Seq     int      `json:"seq"`
	Tick    int      `json:"tick"`
	Kind    string   `json:"kind"`
	Source  int      `json:"source"`
	Target  int      `json:"target"`
	Skill   int      `json:"skill"`
	ByHero  bool     `json:"by_hero"`
	Amount  int      `json:"amount"`
	Detail  int      `json:"detail"`
	Visible []string `json:"visible"`
	Hidden  []string `json:"hidden"`
}

// IsHidden reports whether the named field is a test-only read.
func (e CombatEvent) IsHidden(field string) bool { return inList(e.Hidden, field) }

// CombatLog is the reply of `debug combatlog [n]`: the last events, oldest first. Total counts every event since the journal
// started (it is not bounded); Enabled is false when the fork runs without the journal.
type CombatLog struct {
	Tick    int           `json:"tick"`
	Total   int           `json:"total"`
	Enabled bool          `json:"enabled"`
	Events  []CombatEvent `json:"events"`
}

// Count returns how many listed events have the kind.
func (l *CombatLog) Count(kind string) int {
	n := 0
	for _, e := range l.Events {
		if e.Kind == kind {
			n++
		}
	}
	return n
}

// Since returns the listed events with Seq above seq (a bookmark taken from an earlier Total).
func (l *CombatLog) Since(seq int) []CombatEvent {
	var out []CombatEvent
	for _, e := range l.Events {
		if e.Seq > seq {
			out = append(out, e)
		}
	}
	return out
}

func inList(list []string, name string) bool {
	for _, s := range list {
		if s == name {
			return true
		}
	}
	return false
}

func parseCombatRead(verb string, r *Reply, into any) error {
	if r == nil || len(r.Lines) == 0 {
		return combatReadErr(verb, errors.New("no reply line"))
	}
	if err := json.Unmarshal([]byte(r.Lines[0]), into); err != nil {
		return combatReadErr(verb, err)
	}
	return nil
}

// ParseMinions reads the JSON line of a `debug minions` reply.
func ParseMinions(r *Reply) (*Minions, error) {
	var m Minions
	if err := parseCombatRead("minions", r, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// ParseSkillState reads the JSON line of a `debug skillstate` reply.
func ParseSkillState(r *Reply) (*SkillState, error) {
	var s SkillState
	if err := parseCombatRead("skillstate", r, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// ParseCombatLog reads the JSON line of a `debug combatlog` reply.
func ParseCombatLog(r *Reply) (*CombatLog, error) {
	var l CombatLog
	if err := parseCombatRead("combatlog", r, &l); err != nil {
		return nil, err
	}
	return &l, nil
}

// Minions asks for the hero's summons (read-only). An `err` reply (stock d2d, a fork without the flag) comes back as the error.
func (c *Client) Minions(ctx context.Context) (*Minions, error) {
	r, err := c.Debug(ctx, "minions")
	if err != nil {
		return nil, combatReadErr("minions", err)
	}
	return ParseMinions(r)
}

// SkillState asks for the state of each skill of the hero (read-only).
func (c *Client) SkillState(ctx context.Context) (*SkillState, error) {
	r, err := c.Debug(ctx, "skillstate")
	if err != nil {
		return nil, combatReadErr("skillstate", err)
	}
	return ParseSkillState(r)
}

// CombatLog asks for the last n events of the combat journal (n <= 0: the fork's default of 64).
func (c *Client) CombatLog(ctx context.Context, n int) (*CombatLog, error) {
	args := []string{"combatlog"}
	if n > 0 {
		args = append(args, strconv.Itoa(n))
	}
	r, err := c.Debug(ctx, args...)
	if err != nil {
		return nil, combatReadErr("combatlog", err)
	}
	return ParseCombatLog(r)
}
