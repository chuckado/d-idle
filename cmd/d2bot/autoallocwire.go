package main

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"d2lod-bot/internal/agency"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/points"
	"d2lod-bot/internal/power"
)

// The automatic allocator on the real bot (docs/autoalloc.md): the build list for the picker, the planner that says what
// a locked build wants next, and the loop that asks the desk to spend one point at a time. The desk does the spending
// by the same verbs and the same legality checks as the Points page.

// previewLevels is how many levels of a build the picker shows (the normal30 tables end at 30).
const previewLevels = 30

type allocWire struct {
	db    *data.DB
	caps  *agency.Caps
	kinds map[string]string
}

func newAllocWire(db *data.DB, log *slog.Logger) *allocWire {
	if db == nil {
		return nil
	}
	caps, err := agency.LoadCaps()
	if err != nil {
		log.Warn("auto-alloc: no capability table, the picker shows no tiers", "err", err)
		return &allocWire{db: db}
	}
	caps.Overlay(db.SkillByName)
	w := &allocWire{db: db, caps: caps}
	if p, err := power.Load(); err == nil {
		w.kinds = p.SkillKinds
	}
	return w
}

// tier is the capability tier of a skill by name ("" when unknown or without a table).
func (w *allocWire) tier(name string) (tier, why string) {
	if w.caps == nil {
		return "", ""
	}
	c, ok := w.caps.Lookup(name, w.kinds)
	if !ok {
		return agency.Unsupported, "no capability row: the fighter has no rule for it"
	}
	return c.Tier, c.Note
}

// builds lists every build file with its class, preview and what it needs. A build is experimental when any skill of its
// first levels is not supported; passives and unknown-class skills never count (a passive is supported by rule).
func (w *allocWire) builds() []points.BuildInfo {
	var out []points.BuildInfo
	for _, name := range leveler.TableNames() {
		b, err := leveler.LoadBuild(name, "")
		if err != nil {
			continue
		}
		bi := points.BuildInfo{Name: name, Label: strings.ReplaceAll(name, "_", " "), Class: b.Class, Tier: points.BuildSupported}
		seen := map[string]bool{}
		for _, r := range b.Rows {
			if r.Clvl > previewLevels {
				break
			}
			pr := points.PreviewRow{Clvl: r.Clvl}
			for a, n := range r.Attr {
				if n > 0 {
					if pr.Stats == nil {
						pr.Stats = map[string]int{}
					}
					pr.Stats[leveler.StatKey(a)] = n
				}
			}
			for _, s := range r.Skills {
				t, why := w.tier(s.Skill)
				pr.Skills = append(pr.Skills, points.PreviewSkill{Skill: s.Skill, Points: s.Points, Tier: t})
				if t != "" && t != agency.Supported && !seen[s.Skill] {
					seen[s.Skill] = true
					bi.Needs = append(bi.Needs, points.Need{Skill: s.Skill, Tier: t, Reason: why})
					bi.Tier = points.BuildExperimental
				}
			}
			bi.Preview = append(bi.Preview, pr)
		}
		out = append(out, bi)
	}
	return out
}

// next is the build's remaining steps for the hero, stats first then skills, each group in table order.
func (w *allocWire) next(build string, p *d2.Player) ([]points.Step, error) {
	b, err := leveler.LoadBuild(build, "")
	if err != nil {
		return nil, err
	}
	have := leveler.HaveFor(w.db, p, nil)
	acts := b.Plan(have, leveler.DBRule(w.db, leveler.ClassCode(w.db, p.Class.String())), nil)
	var out []points.Step
	for _, a := range acts {
		if a.Kind == leveler.KindStat {
			out = append(out, points.Step{Kind: "stat", Stat: leveler.StatKey(a.Attr), Name: a.String()})
			continue
		}
		sk := w.db.SkillByName(a.Skill)
		if sk == nil {
			continue
		}
		out = append(out, points.Step{Kind: "skill", Skill: sk.ID, Name: a.Skill})
	}
	return out, nil
}

func (w *allocWire) offBuild(build string, p *d2.Player) []string {
	b, err := leveler.LoadBuild(build, "")
	if err != nil {
		return nil
	}
	owned := leveler.OwnedSkills(w.db, p)
	for n := range owned {
		if sk := w.db.SkillByName(n); sk == nil || sk.Page == 0 {
			delete(owned, n)
		}
	}
	return b.OffBuild(owned)
}

// fillAutoAlloc adds the allocator's functions to the desk config.
func fillAutoAlloc(cfg *points.Config, db *data.DB, e *env.Env, log *slog.Logger) {
	w := newAllocWire(db, log)
	if w == nil {
		return
	}
	cfg.AutoOn = func() bool { return e.Tune != nil && e.Tune.Agency.AutoAlloc != 0 }
	cfg.Builds, cfg.Next, cfg.OffBuild = w.builds, w.next, w.offBuild
	cfg.Respec = func() points.RespecParams {
		t := env.DefaultRespecTuning()
		if e.Tune != nil {
			t = e.Tune.Respec
		}
		return points.RespecParams{BaseGold: t.BaseGold, PerLevel: t.PerLevel, PerRespecPct: t.PerRespecPct, FirstFree: t.FirstFree, MinGold: t.MinGold, MaxGold: t.MaxGold}
	}
	cfg.Logf = func(msg string, kv ...any) { log.Info(msg, kv...) }
}

// runAutoAlloc asks the desk for one point every few seconds while the knob is on and a build is locked. The desk checks
// the calm moment, the game's rules and the read-back; this loop only paces it.
func runAutoAlloc(ctx context.Context, d *points.Desk, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			d.AutoAllocOnce(cctx)
			cancel()
		}
	}
}
