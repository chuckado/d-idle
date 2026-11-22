package loot

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/lootrules"
)

// Wave 2a of the loot rules engine (docs/lootrules.md "Wave 2a"): per-stage switch-over behind loot.rulesEngine.
// 0 off, 1 the ground stage, 2 ground and tidy, 3 plus the sell chore (wave 2b, after the user's decision on the five
// sell kinds, docs/lootrules.md section 13). The engine only ever replaces ONE decision at a time and falls back
// to today's answer for that decision when it errors or panics (logged once per message): a bug here must never lose
// an item.

// Engine stages (loot.rulesEngine).
const (
	StageOff    = 0
	StageGround = 1
	StageTidy   = 2
	StageSell   = 3
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

// EngineStage is the knob loot.rulesEngine clamped to the stages that exist.
func (l *Looter) EngineStage() int {
	if l == nil || l.engine == nil || l.e == nil || l.e.Tune == nil {
		return StageOff
	}
	return min(max(l.e.Tune.Loot.RulesEngine, 0), StageSell)
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

// SellCtx is what the sell chore knows and the rules do not: the pack is full (today's trip threshold, or a forced room
// make), an unidentified plain item may be sold as it is, and the priority from which a ground rule's keeper counts as
// valuable (loot.valuableMinPrio).
type SellCtx struct {
	Full          bool
	SellUnid      bool
	ValuableMinPr int
}

// decideSell is the sell stage for one pack item: sell or keep, with the rule and the reason.
//
// 1. The player's mark and the invariants come first (Compiled.Decide, scope sell: a keep mark, the starting kit, quest
// and never-sell items stay; a sell mark sells unless the item is never-sell).
// 2. The rules of scope sell decide the kinds the user settled on 2026-11-16 (docs/lootrules.md section 13): unidentified
// jewelry is sold while no Tome of Identify is in the pack and stashed (left in the pack for the tidy chore) once one is.
// 3. Everything else is judged the way the pickit judges today, with the ground rules of the same set: what they would
// pick up is KEPT (decision 1: good gear picked up on purpose), a rule that only sells (low gems, magic and rare gear)
// SELLS, and what they would leave on the floor (cheap junk with no rule, plain junk, gems the build does not need)
// SELLS when the pack is full (decisions 2 to 4; the tidy drop rule keeps the full pack away from town). An unidentified
// item whose answer waits for identification sells only when the pack is full or the town says it is not worth a scroll.
func decideSell(c *lootrules.Compiled, it *lootrules.Item, cx *lootrules.Ctx, sc SellCtx) (sell bool, d lootrules.Decision) {
	it.Place = lootrules.PlaceSell
	if it.Mark == "sell" && !it.Quest {
		return true, lootrules.Decision{Action: lootrules.Sell, Rule: "invariant:mark", Why: "Will sell: you marked it for sale"}
	}
	d = c.Decide(it, cx)
	if d.Rule != "default" {
		return d.Action == lootrules.Sell, d
	}
	g := *it
	g.Place = lootrules.PlaceGround
	gd := c.Decide(&g, cx)
	needID := gd.Pending && (gd.Action == lootrules.Pickup || gd.Fate == lootrules.Sell)
	switch {
	case needID && !it.Identified:
		sell = (sc.Full || sc.SellUnid) && gd.Prio < sc.ValuableMinPr
	case gd.Action == lootrules.Pickup:
		sell = false
	case gd.Action == lootrules.Ignore && gd.Fate == lootrules.Sell:
		sell = true
	default:
		sell = sc.Full && gd.Tier != "S"
	}
	if sell {
		gd.Why = strings.Replace(gd.Why, "Left alone", "Will sell", 1)
	}
	return sell, gd
}

// SellDecision is the sell stage (loot.rulesEngine 3): whether to sell one unmarked pack unit, or ok=false when the
// stage is off or the engine failed (then the caller keeps today's answer).
func (p *ShadowPass) SellDecision(u d2.Unit, mark string, neverSell bool, sc SellCtx) (sell bool, rule, why string, ok bool) {
	if p == nil || p.stage < StageSell || u.Item == nil {
		return false, "", "", false
	}
	it, err := p.item(ShadowSell, u, mark, neverSell, nil)
	if err == nil {
		sell, rule, why, err = p.l.engine.decideSell(&it, p.cx, sc)
		if err == nil {
			return sell, rule, why, true
		}
	}
	p.l.engine.fallback(p.l.e.Log, err)
	return false, "", "", false
}

func (e *Engine) decideSell(it *lootrules.Item, cx *lootrules.Ctx, sc SellCtx) (sell bool, rule, why string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("loot rules engine panic: %v", r)
		}
	}()
	c, err := e.compiled(cx.Class, cx.Build)
	if err != nil {
		return false, "", "", fmt.Errorf("loot rules engine: compile for class %q build %q: %w", cx.Class, cx.Build, err)
	}
	s, d := decideSell(c, it, cx, sc)
	return s, d.Rule, d.Why, nil
}
