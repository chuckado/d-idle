package loot

import (
	"fmt"
	"log/slog"
	"sync"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/lootrules"
)

// Wave 2a of the loot rules engine (docs/lootrules.md "Wave 2a"): per-stage switch-over behind loot.rulesEngine.
// 0 off, 1 the ground stage, 2 ground and tidy. 3 is reserved for the sell chore, which needs a user decision on five
// kinds and stays on today's code at every value. The engine only ever replaces ONE decision at a time and falls back
// to today's answer for that decision when it errors or panics (logged once per message): a bug here must never lose
// an item.

// Engine stages (loot.rulesEngine).
const (
	StageOff    = 0
	StageGround = 1
	StageTidy   = 2
)

// Engine holds the active rule set and its compiled forms per hero class and build.
type Engine struct {
	mu     sync.Mutex
	rs     *lootrules.RuleSet
	comp   map[string]*lootrules.Compiled
	warned map[string]bool
}

// DefaultEngine is the process-wide engine every Looter starts with (the Starter preset until SetActiveRules).
var DefaultEngine = NewEngine(lootrules.Starter())

// NewEngine decides with the rule set rs.
func NewEngine(rs *lootrules.RuleSet) *Engine {
	return &Engine{rs: rs, comp: map[string]*lootrules.Compiled{}, warned: map[string]bool{}}
}

// SetRules swaps the rule set (the store calls this after a save); the compiled forms are rebuilt on demand.
func (e *Engine) SetRules(rs *lootrules.RuleSet) {
	e.mu.Lock()
	e.rs, e.comp = rs, map[string]*lootrules.Compiled{}
	e.mu.Unlock()
}

// Rules returns the active rule set.
func (e *Engine) Rules() *lootrules.RuleSet {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.rs
}

// SetActiveRules makes rs the rule set of the engine and of the shadow stage.
func SetActiveRules(rs *lootrules.RuleSet) {
	DefaultEngine.SetRules(rs)
	DefaultShadow.SetRules(rs)
}

// SetRules swaps the shadow's rule set.
func (s *Shadow) SetRules(rs *lootrules.RuleSet) {
	s.mu.Lock()
	s.rs, s.comp = rs, map[string]*lootrules.Compiled{}
	s.mu.Unlock()
}

func (e *Engine) compiled(class, build string) (*lootrules.Compiled, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	k := class + "|" + build
	if c, ok := e.comp[k]; ok {
		return c, nil
	}
	c, err := lootrules.Compile(e.rs, class, build)
	if err != nil {
		return nil, err
	}
	e.comp[k] = c
	return c, nil
}

// decide asks the engine for one item; any error or panic is returned, never thrown.
func (e *Engine) decide(it *lootrules.Item, cx *lootrules.Ctx) (d lootrules.Decision, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("loot rules engine panic: %v", r)
		}
	}()
	c, err := e.compiled(cx.Class, cx.Build)
	if err != nil {
		return d, fmt.Errorf("loot rules engine: compile for class %q build %q: %w", cx.Class, cx.Build, err)
	}
	return c.Decide(it, cx), nil
}

// fallback logs an engine failure once per distinct message.
func (e *Engine) fallback(log *slog.Logger, err error) {
	e.mu.Lock()
	first := !e.warned[err.Error()]
	e.warned[err.Error()] = true
	e.mu.Unlock()
	if first && log != nil {
		log.Warn("loot rules engine failed; today's code decides this item", "err", err.Error())
	}
}

// EngineStage is the knob loot.rulesEngine clamped to what is implemented (3 acts as 2: the sell stage is not built).
func (l *Looter) EngineStage() int {
	if l == nil || l.engine == nil || l.e == nil || l.e.Tune == nil {
		return StageOff
	}
	return min(max(l.e.Tune.Loot.RulesEngine, 0), StageTidy)
}

// SetEngine replaces the looter's engine (tests); nil switches every engine stage off for this looter.
func (l *Looter) SetEngine(e *Engine) { l.engine = e }

// Engine returns the engine the looter uses.
func (l *Looter) Engine() *Engine { return l.engine }

// ground-stage verdict: the engine's decision in the pickit's vocabulary, so every filter that follows (strictness,
// minimum priority, per-cell value, small gold piles, pack room) applies to it unchanged.
func (p *ShadowPass) groundVerdict(d lootrules.Decision) Verdict {
	v := Verdict{Tier: d.Tier, Prio: d.Prio, Rule: d.Rule, Reason: d.Why, NeedID: d.Pending}
	switch {
	case d.Action == lootrules.Pickup:
		v.Action = Keep
	case d.Action == lootrules.Ignore && d.Fate == lootrules.Sell:
		v.Action = Sell
	default:
		v.Action = Drop // the pickit's "leave it": passes() refuses it
		v.NeedID = false
	}
	return v
}

// GroundVerdict is the ground stage (loot.rulesEngine >= 1): the engine's verdict for the unit, or ok=false when the
// stage is off or the engine failed (then the caller keeps today's verdict). stats is the unit's stat list if the
// caller read it.
func (p *ShadowPass) GroundVerdict(u d2.Unit, stats d2.Stats) (v Verdict, ok bool) {
	if p == nil || p.stage < StageGround || u.Item == nil {
		return v, false
	}
	it, err := p.item(ShadowGround, u, "", false, stats)
	if err == nil {
		it.Place = lootrules.PlaceGround
		var d lootrules.Decision
		if d, err = p.l.engine.decide(&it, p.cx); err == nil {
			return p.groundVerdict(d), true
		}
	}
	p.l.engine.fallback(p.l.e.Log, err)
	return v, false
}

// TidyDecision is the tidy stage (loot.rulesEngine >= 2): the engine's decision for one pack unit folded into the tidy
// vocabulary (keep, stash, sell, drop, ignore), or ok=false when the stage is off or the engine failed.
func (p *ShadowPass) TidyDecision(u d2.Unit, mark string, neverSell bool) (action, rule, why string, ok bool) {
	if p == nil || p.stage < StageTidy || u.Item == nil {
		return "", "", "", false
	}
	it, err := p.item(ShadowTidy, u, mark, neverSell, nil)
	if err == nil {
		it.Place = lootrules.PlacePack
		var d lootrules.Decision
		if d, err = p.l.engine.decide(&it, p.cx); err == nil {
			return norm(ShadowTidy, d.Action), d.Rule, d.Why, true
		}
	}
	p.l.engine.fallback(p.l.e.Log, err)
	return "", "", "", false
}
