package agency

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/power"
)

// Modes of the knob agency.planner.
const (
	ModeOff    = 0
	ModeShadow = 1
)

// Input is what the planner reads about the hero at one moment. All of it comes from the snapshot the bot already has.
type Input struct {
	Hero     string
	Alloc    Allocation // class, clvl, hard points by name, attribute points added
	Weapon   Weapon     // Avg 0 = not read
	FCR      int        // gear speeds (power.Hero.Attack carries them); GearRead says 0 is a real 0
	IAS      int
	GearRead bool
	Combat   string     // the combat line of the hero's plan table ("" = class default profile)
	Power    power.Hero // the survival model's reading of the hero (power.HeroFromPlayer)
	Diff     string     // lower case difficulty name of the hero's game
	Act      int        // the act the hero is in
}

// Planner is the shadow planner: the capability table, the survival model and the zone table.
type Planner struct {
	Env   Env
	Model *power.Model // nil: no viability line
	Tab   *danger.Table
}

// Evaluate derives the plan for the input and judges the act the hero is in. A viability that cannot be judged leaves
// the view's Viability nil and says why in Note; only a plan that cannot be derived is an error.
func (p *Planner) Evaluate(in Input) (View, Plan, error) {
	env := p.Env
	env.Weapon, env.FCR, env.IAS, env.GearRead = in.Weapon, in.FCR, in.IAS, in.GearRead
	env.Combat = combatBuild(in.Combat)
	pl, err := Derive(in.Alloc, env)
	if err != nil {
		return View{}, pl, err
	}
	var via *Viability
	note := ""
	if p.Model != nil && p.Tab != nil && in.Diff != "" && in.Act > 0 {
		if v, err := Judge(p.Model, p.Tab, in.Power, pl, in.Diff, in.Act); err != nil {
			note = "viability not judged: " + err.Error()
		} else {
			via = &v
		}
	} else {
		note = "viability not judged: no zone table or game state"
	}
	v := ViewOf(pl, pl.Prior, via)
	v.Hero, v.Class, v.Level, v.Note = in.Hero, in.Alloc.Class, in.Alloc.Clvl, note
	return v, pl, nil
}

// ShadowLine is the `planner shadow:` log line: what the plan would play against what the combat profile plays now.
func ShadowLine(v View, pl Plan) string {
	pick := func(c *Choice) string {
		if c == nil {
			return "-"
		}
		return fmt.Sprintf("%s/%d", c.Skill, c.Level)
	}
	aura, auraProf := "-", orDash(pl.PriorAura)
	if pl.Aura != nil {
		aura = pl.Aura.Skill
	}
	var codes []string
	for _, w := range pl.Warnings {
		codes = append(codes, w.Code)
	}
	agree := v.Agrees != nil && *v.Agrees
	via := "none"
	if v.Viability != nil {
		via = fmt.Sprintf("%s:%s", fmt.Sprintf("act%d", v.Viability.Act), strings.ReplaceAll(v.Viability.Word, " ", "_"))
	}
	return fmt.Sprintf("planner shadow: clvl=%d class=%s plan=%s profile=%s agree=%t best=%s aura=%s auraProfile=%s arch=%s unplayed=%d warnings=%s via=%s",
		v.Level, v.Class, pick(pl.Attack), orDash(pl.Prior), agree, orDash(pl.Best), aura, auraProf, orDash(pl.Archetype), pl.Unplayed.Count, orDash(strings.Join(codes, ",")), via)
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// Limiter lets a line through at most once per interval; the clock is the caller's (the game clock under lockstep).
type Limiter struct {
	mu    sync.Mutex
	every time.Duration
	last  time.Time
}

// NewLimiter is a limiter with the interval (a minute for the shadow line).
func NewLimiter(every time.Duration) *Limiter { return &Limiter{every: every} }

// Allow reports whether a line may be written at now, and records it.
func (l *Limiter) Allow(now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.last.IsZero() && now.Sub(l.last) < l.every {
		return false
	}
	l.last = now
	return true
}
