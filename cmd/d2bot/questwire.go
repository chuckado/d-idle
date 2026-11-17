package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync/atomic"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/questboard"
	"d2lod-bot/internal/watch"
)

// The quest board wiring (slice S4, docs/questboard.md "API"): the margin query behind the readiness words, the per-hero
// plan store, the desk behind GET and POST /api/quests and the glance rows. Everything here is inert while the knob
// quests.board is 0: the Exec answers "off" and boardStep never runs.

// questStoreDir is where the plan records live. The files are named <hero>.quests.json, so a Character Select switch needs
// no parking (internal/charsel/perchar.json lists only files that carry no hero name); each hero keeps its own file and its
// .bak next to it.
func questStoreDir(stateDir string) string { return filepath.Join(stateDir, "quests") }

// survShadow is the survival model the margin query reads (set by newSurvivalShadow; nil without game tables).
var survShadow atomic.Pointer[survivalShadow]

// questMargin is bot.Deps.QuestMargin: read only, never writes the ledger and never touches the zone choice.
func questMargin(s *d2.Snapshot, boss, diff string, potions int) (questboard.Hero, *questboard.Encounter, bool) {
	w := survShadow.Load()
	if w == nil {
		return questboard.Hero{}, nil, false
	}
	return w.questMargin(s, boss, diff, potions)
}

// wireQuestBoard gives the runner the board maker and the margin query, and returns the desk the watch routes use (nil
// without a state directory: the plan records need a place to live). Safe to call with the knob at 0.
func wireQuestBoard(stateDir string, ref *runnerRef, r *bot.Runner, logf func(msg string, kv ...any)) *questboard.Desk {
	if ref == nil || r == nil || stateDir == "" {
		return nil
	}
	b, err := questboard.Load()
	if err != nil {
		logf("quest board off", "err", err)
		return nil
	}
	dir := questStoreDir(stateDir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		logf("quest board off: no plan directory", "err", err)
		return nil
	}
	r.SetQuestBoardMaker(func(hero string) (*questboard.Board, *questboard.Store, error) {
		st, err := questboard.Open(b, dir, hero)
		return b, st, err
	})
	r.SetQuestMargin(questMargin)
	d := questboard.NewDesk(b, questboard.DeskConfig{
		Exec:    r.QuestBoard,
		Paused:  func() bool { return r.PlayerPauseView().State != bot.PauseRunning },
		Explain: r.QuestReadiness,
		Class:   r.QuestClass,
		Meter:   r.QuestMeter,
	})
	ref.qd = d
	return d
}

// bindQuests mounts the quest board routes on the watch screen.
func bindQuests(f *watch.Fun, d *questboard.Desk) {
	if f == nil || d == nil {
		return
	}
	api := watch.NewQuestsAPI()
	api.Bind(d)
	f.QuestBoard = api
}

func init() {
	// The video card strip and the Needs-you rows of the quest board (design 4.6, slice S4).
	glanceExtra = append(glanceExtra, func(src *watch.GlanceSources, d glanceDeps) {
		if d.Quests == nil {
			return
		}
		src.QuestOn = func() bool {
			qd := d.Quests()
			return qd != nil && qd.View().Enabled
		}
		src.Quest = func() (watch.GlanceQuest, bool) {
			qd := d.Quests()
			if qd == nil {
				return watch.GlanceQuest{}, false
			}
			s := qd.View().Strip
			if s == nil {
				return watch.GlanceQuest{}, false
			}
			return watch.GlanceQuest{Quest: s.Quest, Title: s.QuestTitle, Chunk: s.Chunk, Part: s.Title, State: string(s.State), Step: s.Step, NPC: s.NPC}, true
		}
		prev := src.Needs
		src.Needs = func() []watch.GlanceNeed {
			var out []watch.GlanceNeed
			if prev != nil {
				out = prev()
			}
			if qd := d.Quests(); qd != nil {
				out = append(out, questNeeds(qd.View().Needs)...)
			}
			return out
		}
	})
}

func questNeeds(rows []questboard.NeedRow) []watch.GlanceNeed {
	var out []watch.GlanceNeed
	for _, n := range rows {
		g := watch.GlanceNeed{ID: n.ID, Kind: n.Kind, Sev: n.Sev, Text: n.Text, Actions: []watch.GlanceVerb{}}
		for _, a := range n.Actions {
			body, _ := json.Marshal(a.Body)
			g.Actions = append(g.Actions, watch.GlanceVerb{Label: a.Label, Post: a.Post, Body: body})
		}
		out = append(out, g)
	}
	return out
}
