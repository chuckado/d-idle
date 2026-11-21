package points

import (
	"context"
	"d2lod-bot/internal/d2"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// The automatic allocator and the build lock (docs/autoalloc.md). The desk holds no build tables: the wiring hands it
// the list of builds and a planner (Config.Builds, Config.Next), so this file knows no class, skill or build name.

// Build tiers shown to the player; "experimental" needs one extra confirmation when the build is chosen.
const (
	BuildSupported    = "supported"
	BuildExperimental = "experimental"
)

// BuildInfo is one build file a hero of Class may follow.
type BuildInfo struct {
	Name    string       `json:"name"`
	Label   string       `json:"label"`
	Class   string       `json:"class"`
	Tier    string       `json:"tier"`
	Note    string       `json:"note,omitempty"`
	Preview []PreviewRow `json:"preview"`
	Needs   []Need       `json:"needs,omitempty"`
}

// PreviewRow is what the build spends at one character level.
type PreviewRow struct {
	Clvl   int            `json:"clvl"`
	Skills []PreviewSkill `json:"skills,omitempty"`
	Stats  map[string]int `json:"stats,omitempty"`
}

// PreviewSkill is a skill of a preview row with its capability tier.
type PreviewSkill struct {
	Skill  string `json:"skill"`
	Points int    `json:"points"`
	Tier   string `json:"tier,omitempty"`
}

// Need is a skill of the build the hero cannot use well yet (the capability table says so).
type Need struct {
	Skill  string `json:"skill"`
	Tier   string `json:"tier"`
	Reason string `json:"reason"`
}

// Step is one point the build wants next. Kind is "stat" (Stat is the sheet key) or "skill" (Skill is the skills.txt id).
type Step struct {
	Kind  string
	Stat  string
	Skill int
	Name  string
}

// RespecParams is the price model (env.RespecTuning; placeholders, docs/autoalloc.md section 3).
type RespecParams struct{ BaseGold, PerLevel, PerRespecPct, FirstFree, MinGold, MaxGold int }

// RespecPrice is clamp(min, max, base + perLevel x clvl) x (1 + perRespecPct/100 x paid respecs). The first FirstFree
// respecs of a hero are free and do not count for the escalation.
func RespecPrice(p RespecParams, clvl, respecs int) int {
	if respecs < p.FirstFree {
		return 0
	}
	base := p.BaseGold + p.PerLevel*max(clvl, 0)
	base = max(base, p.MinGold)
	if p.MaxGold > 0 {
		base = min(base, p.MaxGold)
	}
	paid := max(respecs-p.FirstFree, 0)
	return base + base*p.PerRespecPct*paid/100
}

// Held is a point the allocator did not spend, with the reason the player reads.
type Held struct {
	Clvl   int    `json:"clvl"`
	Skill  string `json:"skill"`
	Reason string `json:"reason"`
}

// BuildState is the hero's stored choice.
type BuildState struct {
	Build         string `json:"build,omitempty"`
	Label         string `json:"label,omitempty"`
	Tier          string `json:"tier,omitempty"`
	Locked        bool   `json:"locked"`
	LockedAtLevel int    `json:"lockedAtLevel,omitempty"`
	Respecs       int    `json:"respecs"`
}

// RespecView is the respec line of the build page. Available stays false until the fork verb exists (next card).
type RespecView struct {
	Price     int    `json:"price"`
	Available bool   `json:"available"`
	Note      string `json:"note"`
}

// BuildView is GET /api/hero/build.
type BuildView struct {
	Class     string      `json:"class,omitempty"`
	Level     int         `json:"level,omitempty"`
	AutoAlloc bool        `json:"autoAlloc"`
	State     BuildState  `json:"state"`
	Builds    []BuildInfo `json:"builds"`
	Held      []Held      `json:"held,omitempty"`
	OffBuild  []string    `json:"offBuild,omitempty"`
	Respec    RespecView  `json:"respec"`
}

type autoState struct {
	mu   sync.Mutex
	held []Held
}

func (d *Desk) autoOn() bool { return d.c.AutoOn != nil && d.c.AutoOn() }

// Locked says whether a build is locked and the allocator owns the hero's points (the leveler's own spending then stays
// closed so the two never double-spend).
func (d *Desk) Locked() bool {
	s := d.store.Get()
	return d.autoOn() && s.Locked && s.Build != ""
}

func (d *Desk) buildsOf(class string) []BuildInfo {
	var out []BuildInfo
	if d.c.Builds == nil {
		return out
	}
	for _, b := range d.c.Builds() {
		if strings.EqualFold(b.Class, class) {
			out = append(out, b)
		}
	}
	return out
}

// BuildView is the picker and the state.
func (d *Desk) BuildView() BuildView {
	v := BuildView{Builds: []BuildInfo{}, AutoAlloc: d.autoOn()}
	s := d.store.Get()
	v.State = BuildState{Build: s.Build, Locked: s.Locked, LockedAtLevel: s.LockedAtLevel, Respecs: s.Respecs}
	var p *d2.Player
	if d.c.Player != nil {
		p = d.c.Player()
	}
	level := 0
	if p != nil {
		v.Class, v.Level, level = p.Class.String(), p.Level, p.Level
		v.Builds = append(v.Builds, d.buildsOf(v.Class)...)
	}
	for _, b := range v.Builds {
		if b.Name == s.Build {
			v.State.Label, v.State.Tier = b.Label, b.Tier
		}
	}
	rp := RespecParams{}
	if d.c.Respec != nil {
		rp = d.c.Respec()
	}
	v.Respec = RespecView{Price: RespecPrice(rp, level, s.Respecs), Note: "Respec: coming soon. The price is shown from the tuning knobs; paying it needs the respec command of the game, which is not built yet."}
	d.auto.mu.Lock()
	v.Held = append(v.Held, d.auto.held...)
	d.auto.mu.Unlock()
	if p != nil && s.Locked && d.c.OffBuild != nil {
		v.OffBuild = d.c.OffBuild(s.Build, p)
	}
	return v
}

// SetBuild chooses a build and locks it. The same build again is ok (already). Another build while locked is refused:
// the lock ends only with a paid respec.
func (d *Desk) SetBuild(name string, confirm bool) Reply {
	var p *d2.Player
	if d.c.Player != nil {
		p = d.c.Player()
	}
	if p == nil {
		return refuse(http.StatusServiceUnavailable, "no_hero", "the hero is not in a game yet")
	}
	s := d.store.Get()
	var pick *BuildInfo
	for _, b := range d.c.builds() {
		if b.Name == name {
			b := b
			pick = &b
		}
	}
	if pick == nil {
		return refuse(http.StatusBadRequest, "unknown_build", fmt.Sprintf("no build named %q", name))
	}
	if !strings.EqualFold(pick.Class, p.Class.String()) {
		return refuse(http.StatusConflict, "other_class", fmt.Sprintf("build %q is for the %s, this hero is a %s", name, pick.Class, p.Class.String()))
	}
	if s.Locked {
		if s.Build == name {
			return Reply{OK: true, Already: true}
		}
		rp := RespecParams{}
		if d.c.Respec != nil {
			rp = d.c.Respec()
		}
		return refuse(http.StatusConflict, "locked", fmt.Sprintf("locked to %s since level %d; a respec (%d gold) is the only way to change it, and respec is coming soon", s.Build, s.LockedAtLevel, RespecPrice(rp, p.Level, s.Respecs)))
	}
	if pick.Tier == BuildExperimental && !confirm {
		r := refuse(http.StatusConflict, "confirm", fmt.Sprintf("%s is experimental: some of its skills are not proven in the fighter yet. Confirm to lock it anyway", pick.Label))
		r.NeedsConfirm, r.Tier = true, pick.Tier
		return r
	}
	if _, err := d.store.LockBuild(name, p.Level, d.c.Now()); err != nil {
		return refuse(http.StatusInternalServerError, "store", "could not save the build: "+err.Error())
	}
	d.logf("auto-alloc: build locked", "build", name, "clvl", p.Level)
	return Reply{OK: true, Tier: pick.Tier}
}

func (c Config) builds() []BuildInfo {
	if c.Builds == nil {
		return nil
	}
	return c.Builds()
}

func (d *Desk) logf(msg string, kv ...any) {
	if d.c.Logf != nil {
		d.c.Logf(msg, kv...)
	}
}

// AutoAllocOnce spends at most one point from the locked build when the knob is on and the moment is calm. It returns
// what it did ("" when nothing). A skill the hero cannot use yet is held with its reason, never spent silently.
func (d *Desk) AutoAllocOnce(ctx context.Context) (did string) {
	if !d.autoOn() {
		return ""
	}
	s := d.store.Get()
	if !s.Locked || s.Build == "" || d.c.Next == nil {
		return ""
	}
	p, bad := d.precheck()
	if bad != nil || p.StatPoints+p.SkillPoints <= 0 {
		return ""
	}
	steps, err := d.c.Next(s.Build, p)
	if err != nil {
		d.setHeld([]Held{{Clvl: p.Level, Reason: "the build cannot be read: " + err.Error()}})
		return ""
	}
	var held []Held
	defer func() { d.setHeld(held) }()
	for _, st := range steps {
		switch st.Kind {
		case "stat":
			if p.StatPoints <= 0 {
				continue
			}
			if r := d.SpendStat(ctx, st.Stat); r.OK {
				d.logf("auto-alloc: clvl "+fmt.Sprint(p.Level)+" -> +1 "+st.Stat+" (build "+s.Build+")", "build", s.Build)
				return "stat " + st.Stat
			} else if r.Code == "busy" || r.Code == "no_verb" || r.Code == "verb_failed" || r.Code == "not_taken" {
				return ""
			}
		case "skill":
			if p.SkillPoints <= 0 {
				continue
			}
			r := d.SpendSkill(ctx, st.Skill, false)
			if r.OK {
				d.logf("auto-alloc: clvl "+fmt.Sprint(p.Level)+" -> skill "+st.Name+" (build "+s.Build+")", "build", s.Build)
				return "skill " + st.Name
			}
			switch r.Code {
			case "busy", "no_verb", "verb_failed", "not_taken", "read_failed":
				return ""
			}
			reason := r.Reason
			if r.NeedsConfirm {
				reason = "Held for you, spend it on the Skills page to confirm: " + r.Reason
			}
			held = append(held, Held{Clvl: p.Level, Skill: st.Name, Reason: reason})
		}
	}
	return ""
}

func (d *Desk) setHeld(h []Held) {
	d.auto.mu.Lock()
	d.auto.held = h
	d.auto.mu.Unlock()
}
