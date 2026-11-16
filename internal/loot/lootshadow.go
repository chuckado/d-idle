package loot

import (
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/lootrules"
)

// The shadow stage of the loot rules engine (docs/lootrules.md, wave 1): at each of today's item decisions the active
// preset is asked too, and a difference is logged and counted. Log only: the engine never acts, never touches the
// reflex layer, never panics into the loop (Observe recovers) and costs about two microseconds an item (TestShadowBudget).
// The knob is bot.lootShadow (0 off, 1 on, default 1).

// ShadowSchema is the version of the read-only report served at GET /api/lootrules/shadow.
const ShadowSchema = "watch.lootshadow/1"

// Scopes of the decision points that feed the shadow.
const (
	ShadowGround = "ground" // the pickup decision of Looter.candidates
	ShadowTidy   = "tidy"   // the stash and drop decision of the tidy chore (tidy.Decide)
	ShadowSell   = "sell"   // the sell chore's sellable list
)

const (
	shadowLines  = 50 // lines kept for the report
	shadowPerMin = 20 // log lines per minute at most; the rest is only counted
	shadowTop    = 10 // rules listed in the report
)

// ShadowRule is one row of the "top disagreements by rule" table.
type ShadowRule struct {
	Rule    string `json:"rule"`
	Count   int    `json:"count"`
	Example string `json:"example"`
}

// ShadowScope are the totals of one scope.
type ShadowScope struct {
	Decisions     int `json:"decisions"`
	Disagreements int `json:"disagreements"`
}

// ShadowReport is the read-only body of GET /api/lootrules/shadow.
type ShadowReport struct {
	Schema        string                 `json:"schema"`
	Enabled       bool                   `json:"enabled"`
	Preset        string                 `json:"preset"`
	Decisions     int                    `json:"decisions"`
	Disagreements int                    `json:"disagreements"`
	Unique        int                    `json:"unique"` // distinct (scope, area, code, quality) that disagreed: the logged lines
	Errors        int                    `json:"errors"` // items the adapter or the engine could not judge (never acted on)
	Scopes        map[string]ShadowScope `json:"scopes"`
	TopRules      []ShadowRule           `json:"topRules"`
	Last          []string               `json:"last"`
	AvgMicros     float64                `json:"avgMicros"`
	MaxMicros     float64                `json:"maxMicros"`
}

// Shadow collects the disagreements. One per process is shared by the looter and the town chores (DefaultShadow);
// tests make their own. Safe for concurrent use.
type Shadow struct {
	mu     sync.Mutex
	log    *slog.Logger
	now    func() time.Time // the game clock for the log rate
	wall   func() time.Time // host clock on purpose: the cost of a decision, not game time
	rs     *lootrules.RuleSet
	comp   map[string]*lootrules.Compiled
	seen   map[string]bool
	byRule map[string]*ShadowRule
	scopes map[string]*ShadowScope
	last   []string
	logged []time.Time
	errs   int
	total  ShadowScope
	nanos  int64
	maxNs  int64
}

// DefaultShadow is the process-wide shadow every Looter starts with.
var DefaultShadow = NewShadow(lootrules.Starter())

// NewShadow shadows the rule set rs (the Starter preset by default, see DefaultShadow).
func NewShadow(rs *lootrules.RuleSet) *Shadow {
	return &Shadow{rs: rs, wall: time.Now, comp: map[string]*lootrules.Compiled{}, seen: map[string]bool{},
		byRule: map[string]*ShadowRule{}, scopes: map[string]*ShadowScope{}}
}

// Use sets the logger and the game clock that rate-limits the lines (nil keeps the current one).
func (s *Shadow) Use(log *slog.Logger, now func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if log != nil {
		s.log = log
	}
	if now != nil {
		s.now = now
	}
}

// compiled returns the rule set compiled for the hero's class and build (cached; an unknown class is an error that
// names it, and the item is counted as an error, never decided).
func (s *Shadow) compiled(class, build string) (*lootrules.Compiled, error) {
	k := class + "|" + build
	if c, ok := s.comp[k]; ok {
		return c, nil
	}
	c, err := lootrules.Compile(s.rs, class, build)
	if err != nil {
		return nil, fmt.Errorf("loot shadow: compile for class %q build %q: %w", class, build, err)
	}
	s.comp[k] = c
	return c, nil
}

// norm folds an action into what today's decision point can say, so the two sides compare.
func norm(scope, a string) string {
	switch scope {
	case ShadowGround:
		if a == lootrules.Pickup {
			return lootrules.Pickup
		}
		return lootrules.Ignore
	case ShadowSell:
		if a == lootrules.Sell {
			return lootrules.Sell
		}
		return lootrules.Keep
	}
	switch a { // tidy: these three leave the item in the pack
	case lootrules.Identify, lootrules.Ask, lootrules.KeepMerc:
		return lootrules.Keep
	}
	return a
}

// observe decides the item with the shadow's rules and records a difference. today is the action of the old code in the
// vocabulary of the scope (ground: pickup or ignore; tidy: keep, stash, sell, drop, ignore; sell: sell or keep).
func (s *Shadow) observe(scope string, area int, name string, it *lootrules.Item, cx *lootrules.Ctx, today string, took time.Duration, fold func(lootrules.Decision) string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			s.errs++
		}
	}()
	c, err := s.compiled(cx.Class, cx.Build)
	if err != nil {
		s.errs++
		s.note(err.Error())
		return
	}
	t0 := s.wall()
	it.Place = map[string]string{ShadowGround: lootrules.PlaceGround}[scope]
	if it.Place == "" {
		it.Place = lootrules.PlacePack
	}
	d := c.Decide(it, cx)
	took += s.wall().Sub(t0)
	s.nanos += int64(took)
	s.maxNs = max(s.maxNs, int64(took))
	sc := s.scopes[scope]
	if sc == nil {
		sc = &ShadowScope{}
		s.scopes[scope] = sc
	}
	sc.Decisions++
	s.total.Decisions++
	rules := norm(scope, d.Action)
	if fold != nil {
		rules = fold(d)
	}
	if rules == today {
		return
	}
	sc.Disagreements++
	s.total.Disagreements++
	rr := s.byRule[d.Rule]
	if rr == nil {
		rr = &ShadowRule{Rule: d.Rule}
		s.byRule[d.Rule] = rr
	}
	rr.Count++
	key := fmt.Sprintf("%s|%d|%s|%s", scope, area, it.Code, it.Quality)
	if s.seen[key] {
		return
	}
	s.seen[key] = true
	line := fmt.Sprintf("loot shadow: %s %s today=%s rules=%s rule=%s reason=%s", scope, name, today, rules, d.Rule, d.Why)
	if rr.Example == "" {
		rr.Example = line
	}
	s.note(line)
}

// note keeps the line for the report and logs it unless the per-minute budget is spent. The caller holds mu.
func (s *Shadow) note(line string) {
	s.last = append(s.last, line)
	if len(s.last) > shadowLines {
		s.last = s.last[len(s.last)-shadowLines:]
	}
	if s.log == nil {
		return
	}
	now := time.Time{}
	if s.now != nil {
		now = s.now()
	}
	kept := s.logged[:0]
	for _, t := range s.logged {
		if now.Sub(t) < time.Minute {
			kept = append(kept, t)
		}
	}
	s.logged = kept
	if len(s.logged) >= shadowPerMin {
		return
	}
	s.logged = append(s.logged, now)
	s.log.Info(line)
}

// Report is the read-only view for the API.
func (s *Shadow) Report(enabled bool) ShadowReport {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := ShadowReport{Schema: ShadowSchema, Enabled: enabled, Preset: s.rs.Name, Decisions: s.total.Decisions,
		Disagreements: s.total.Disagreements, Unique: len(s.seen), Errors: s.errs, Scopes: map[string]ShadowScope{},
		Last: append([]string{}, s.last...), MaxMicros: float64(s.maxNs) / 1e3}
	for k, v := range s.scopes {
		r.Scopes[k] = *v
	}
	for _, v := range s.byRule {
		r.TopRules = append(r.TopRules, *v)
	}
	sort.Slice(r.TopRules, func(i, j int) bool {
		if r.TopRules[i].Count != r.TopRules[j].Count {
			return r.TopRules[i].Count > r.TopRules[j].Count
		}
		return r.TopRules[i].Rule < r.TopRules[j].Rule
	})
	if len(r.TopRules) > shadowTop {
		r.TopRules = r.TopRules[:shadowTop]
	}
	if r.Decisions > 0 {
		r.AvgMicros = float64(s.nanos) / float64(r.Decisions) / 1e3
	}
	return r
}

// Reset clears counters and the dedup set (a new session or a test).
func (s *Shadow) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen, s.byRule, s.scopes, s.last, s.logged = map[string]bool{}, map[string]*ShadowRule{}, map[string]*ShadowScope{}, nil, nil
	s.errs, s.total, s.nanos, s.maxNs = 0, ShadowScope{}, 0, 0
}

// SetShadow replaces the looter's shadow (tests); nil switches the shadow stage off for this looter.
func (l *Looter) SetShadow(s *Shadow) { l.shadow = s }

// Shadow returns the shadow the looter feeds.
func (l *Looter) Shadow() *Shadow { return l.shadow }

// ShadowEnabled is the knob bot.lootShadow.
func (l *Looter) ShadowEnabled() bool {
	return l.shadow != nil && l.e.Tune != nil && l.e.Tune.Bot.LootShadow != 0
}

// ShadowPass is one snapshot's shadow state: the context is built once, then every decision point of the pass calls
// Observe. A nil pass (shadow off, snapshot unusable) makes Observe a no-op.
type ShadowPass struct {
	l    *Looter
	a    *RulesAdapter
	cx   *lootrules.Ctx
	area int
	max  int
	// ground: today's strictness knob and low-gold mode, which filter by tier after the verdict (passes in pickup.go).
	strict  int
	lowGold bool
}

// SetGround tells the pass the strictness and low-gold mode of the pickup being judged, so the rule's tier is filtered
// the way passes filters the verdict. Without it a ground rule that only sells would never count as a pickup.
func (p *ShadowPass) SetGround(strict int, lowGold bool) {
	if p != nil {
		p.strict, p.lowGold = strict, lowGold
	}
}

// groundFold is the rules side of a ground decision in today's vocabulary: a rule that takes the item (pickup) or one
// that only sells it (the fate sell of an ignore) counts as a pickup when its tier passes the strictness filter.
func (p *ShadowPass) groundFold(d lootrules.Decision) string {
	act := Keep
	switch {
	case d.Action == lootrules.Pickup:
	case d.Action == lootrules.Ignore && d.Fate == lootrules.Sell:
		act = Sell
	default:
		return lootrules.Ignore
	}
	if passes(Verdict{Action: act, Tier: d.Tier}, p.strict, p.lowGold) {
		return lootrules.Pickup
	}
	return lootrules.Ignore
}

// ShadowPass starts the pass for snapshot s. packPct and stashPct are the grid fill the caller measured (0 when it did
// not). Returns nil when the stage is off, the snapshot has no player or the data tables are missing.
func (l *Looter) ShadowPass(s *d2.Snapshot, packPct, stashPct int) (p *ShadowPass) {
	if l == nil || !l.ShadowEnabled() || s == nil || s.Player == nil || s.Inventory == nil || l.e.Data == nil {
		return nil
	}
	defer func() {
		if recover() != nil {
			p = nil
			l.shadow.countError()
		}
	}()
	if l.shadow.log == nil {
		l.shadow.Use(l.e.Log, nil)
	}
	if l.shadow.now == nil && l.e.Clock != nil {
		l.shadow.Use(nil, l.e.Clock.Now)
	}
	a := l.Rules()
	return &ShadowPass{l: l, a: a, cx: a.Ctx(s, l.build, packPct, stashPct), area: s.Player.Area, max: MaxBuyOf(s.Player)}
}

func (s *Shadow) countError() { s.mu.Lock(); s.errs++; s.mu.Unlock() }

// Observe records one decision. stats is the unit's stat list when the caller already read it (nil: read here, or
// unreadable). neverSell marks the town's never-sell list. today is the old decision in the vocabulary of the scope.
func (p *ShadowPass) Observe(scope string, u d2.Unit, today, mark string, neverSell bool, stats d2.Stats) {
	if p == nil || u.Item == nil {
		return
	}
	defer func() {
		if recover() != nil {
			p.l.shadow.countError()
		}
	}()
	t0 := p.l.shadow.wall()
	needs := p.needsStats(u)
	if stats == nil && needs && p.l.stats != nil && u.Addr != 0 {
		stats, _ = p.l.stats(u.Addr)
	}
	o := ItemOpts{Mark: mark, MaxBuy: p.max, NeverSell: neverSell}
	if scope == ShadowGround && u.Item.ILvl == 0 {
		// A ground unit without an item level is a d2d label (docs/d2d-loot.md): no level, no identified flag, no stat
		// list except the amount of a gold pile.
		o.LabelOnly = true
		if !p.a.isGold(u) {
			stats = nil
		}
	}
	switch {
	case stats != nil:
		o.Stats = stats
	case !needs:
		o.Stats = d2.Stats{} // a plain item has no stat list worth reading: readable and empty
	}
	it, err := p.a.Item(u, o)
	if err != nil {
		p.l.shadow.countError()
		return
	}
	name := it.Name
	if name == "" {
		name = it.Code
	}
	var fold func(lootrules.Decision) string
	if scope == ShadowGround {
		fold = p.groundFold
	}
	p.l.shadow.observe(scope, p.area, name, &it, p.cx, today, p.l.shadow.wall().Sub(t0), fold)
}

// needsStats says whether a unit's stat list can matter to a rule: gold (the pile), magic or better (affixes), and the
// bases that can carry the four sockets the tidy row looks at. Everything else is judged from the base row alone.
func (p *ShadowPass) needsStats(u d2.Unit) bool {
	b := p.l.base(u.Class)
	if b == nil {
		return false
	}
	return b.Type == "gold" || u.Item.Quality >= d2.QualityMagic || b.Sockets >= 4 || b.Stackable
}

func (a *RulesAdapter) isGold(u d2.Unit) bool {
	b := a.l.base(u.Class)
	return b != nil && b.Type == "gold"
}
