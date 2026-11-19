package main

import (
	"log/slog"
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/agency"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/plandata"
	"d2lod-bot/internal/power"
	"d2lod-bot/internal/watch"
)

// The skill planner in shadow (docs/agency.md, knob agency.planner): from the skills the hero actually has it derives the
// rotation, the aura and a viability line, logs `planner shadow:` once a minute next to what the combat profile picks, and
// serves GET /api/hero/plan. Nothing reads its result: the hero plays as before.

// agencyWire is the planner on the real bot.
type agencyWire struct {
	pl    *agency.Planner
	db    *data.DB
	tune  func() env.AgencyTuning
	hero  func() string
	log   *slog.Logger
	lim   *agency.Limiter
	now   func() time.Time
	combo sync.Map // table name -> combat line

	mu    sync.Mutex
	snap  *d2.Snapshot
	view  *agency.View
	viewT time.Time
}

// wireAgency returns nil without game tables or when the capability table or the model cannot load (one warning).
func wireAgency(db *data.DB, e *env.Env, hero func() string, now func() time.Time, log *slog.Logger) *agencyWire {
	if db == nil || len(db.Skills) == 0 || e == nil {
		return nil
	}
	aenv, err := agency.NewEnv()
	if err != nil {
		log.Warn("agency: no planner", "err", err)
		return nil
	}
	if miss := aenv.Caps.Overlay(db.SkillByName); len(miss) > 0 {
		log.Info("agency: capability rows without a skills.txt row", "skills", strings.Join(miss, ","))
	}
	p := &agency.Planner{Env: aenv}
	if tab := dangerTable(db); tab != nil {
		if m, err := power.NewModel(tab.Params); err == nil {
			p.Model, p.Tab = m, tab
		} else {
			log.Warn("agency: no survival model, the viability line is off", "err", err)
		}
	}
	if now == nil {
		now = time.Now
	}
	return &agencyWire{pl: p, db: db, hero: hero, log: log, lim: agency.NewLimiter(time.Minute), now: now,
		tune: func() env.AgencyTuning {
			if e.Tune == nil {
				return env.DefaultAgencyTuning()
			}
			return e.Tune.Agency
		}}
}

// chainSnap runs the hooks in order on every snapshot.
func chainSnap(fns ...func(*d2.Snapshot)) func(*d2.Snapshot) {
	return func(s *d2.Snapshot) {
		for _, f := range fns {
			if f != nil {
				f(s)
			}
		}
	}
}

// OnSnapshot keeps the last snapshot and, when the knob says shadow, writes the minute's line. Quick: the planner runs at most
// once a minute here.
func (w *agencyWire) OnSnapshot(s *d2.Snapshot) {
	if w == nil || s == nil || s.Player == nil {
		return
	}
	w.mu.Lock()
	w.snap = s
	w.mu.Unlock()
	if w.tune().Planner != 1 || !w.lim.Allow(w.now()) {
		return
	}
	v, pl, err := w.eval(s)
	if err != nil {
		w.log.Warn("planner shadow: no plan", "err", err)
		return
	}
	w.store(v)
	w.log.Info(agency.ShadowLine(v, pl))
}

func (w *agencyWire) store(v agency.View) {
	w.mu.Lock()
	w.view, w.viewT = &v, w.now()
	w.mu.Unlock()
}

// View is the API's source: a fresh evaluation of the last snapshot at most every ten seconds; the "off" view with the
// reason when the knob is 0.
func (w *agencyWire) View() any {
	if w == nil {
		return agency.View{Schema: "didle.plan/1", Mode: "off", Note: "no game tables", Rotation: []agency.ViewSkill{}, Warnings: []agency.ViewWarn{}, ShadowOnly: true}
	}
	if w.tune().Planner == 0 {
		return agency.View{Schema: "didle.plan/1", Mode: "off", Note: "the planner is off (knob agency.planner 0)", Rotation: []agency.ViewSkill{}, Warnings: []agency.ViewWarn{}, ShadowOnly: true}
	}
	w.mu.Lock()
	s, v, at := w.snap, w.view, w.viewT
	w.mu.Unlock()
	if s == nil {
		return agency.View{Schema: "didle.plan/1", Enabled: true, Mode: "shadow", Note: "no snapshot yet", Rotation: []agency.ViewSkill{}, Warnings: []agency.ViewWarn{}, ShadowOnly: true}
	}
	if v != nil && w.now().Sub(at) < 10*time.Second {
		return *v
	}
	nv, _, err := w.eval(s)
	if err != nil {
		return agency.View{Schema: "didle.plan/1", Enabled: true, Mode: "shadow", Note: "no plan: " + err.Error(), Rotation: []agency.ViewSkill{}, Warnings: []agency.ViewWarn{}, ShadowOnly: true}
	}
	w.store(nv)
	return nv
}

// eval reads the hero out of the snapshot and runs the planner.
func (w *agencyWire) eval(s *d2.Snapshot) (agency.View, agency.Plan, error) {
	p := s.Player
	class := strings.ToLower(p.Class.String())
	owned := leveler.OwnedSkills(w.db, p)
	a := agency.Allocation{Class: class, Clvl: p.Level, Skills: owned}
	if int(p.Class) < len(w.db.Classes) {
		a.Ene = max(0, p.Energy-w.db.Classes[p.Class].Int)
	}
	hero := power.HeroFromPlayer(p, s.Merc, -1)
	in := agency.Input{Hero: w.hero(), Alloc: a, Combat: w.combatLine(class, p, owned), Power: hero, Diff: agencyDiff(p.Difficulty), Act: data.ActOfArea(p.Area)}
	if hero.DmgMax > 0 {
		in.Weapon.Avg = float64(hero.DmgMin+hero.DmgMax) / 2
	}
	in.FCR, in.IAS, in.GearRead = hero.Attack.FCR, hero.Attack.IAS, hero.Attack.FCRKnown && hero.Attack.IASKnown
	return w.pl.Evaluate(in)
}

func agencyDiff(d d2.Difficulty) string {
	return [...]string{"normal", "nightmare", "hell"}[min(max(int(d), 0), 2)]
}

// combatLine is the combat line of the plan table the leveler would pick for the hero (the profile the bot fights by).
func (w *agencyWire) combatLine(class string, p *d2.Player, owned map[string]int) string {
	name := leveler.SelectBuild(class, p.Level, owned)
	if c, ok := w.combo.Load(name); ok {
		return c.(string)
	}
	line := ""
	if raw, err := plandata.FS.ReadFile("build-" + name + ".json"); err == nil {
		if b, err := leveler.ParseBuild(raw); err == nil {
			line = b.Combat
		}
	}
	w.combo.Store(name, line)
	return line
}

// bindPlan mounts GET /api/hero/plan on the watch screen.
func bindPlan(f *watch.Fun, w *agencyWire) {
	if f == nil || w == nil {
		return
	}
	api := watch.NewPlanAPI()
	api.Bind(w.View)
	f.Plan = api
}

