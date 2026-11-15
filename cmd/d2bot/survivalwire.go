package main

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/danger"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/power"
	"d2lod-bot/internal/survival"
)

// Survival slice S-C (docs/survival/ledger.md): the per-zone ledger learns from the run, the controller names the zone
// it would farm and why, and a words-only block is served for the UI. SHADOW ONLY: nothing here changes what the bot does.
const (
	shadowTravelMin   = 2.0 // U: minutes of travel assumed for every candidate zone but the current one
	shadowKitPotions  = 4   // the kit rule's healing potions (docs/bot kit), U for this use
	shadowReviveQuote = 400 // U: the real quote is read in town; swap7 paid 360 in Normal
)

// survFight is the hook the fighter's OnSummary chain calls (senseshistory.go); nil until the shadow is wired.
var survFight atomic.Pointer[func(combat.FightSummary)]

// survGlance is what /api/survival serves; nil until the first decision.
var survGlance atomic.Pointer[survival.Glance]

// survivalGlance is the watch.SurvivalSource of the bot's hub.
func survivalGlance() any {
	if g := survGlance.Load(); g != nil {
		return g
	}
	return nil
}

type predKey = survival.ZoneKey

type shadowPred struct{ takenPerKill, secPerKill float64 }

type survivalShadow struct {
	mu       sync.Mutex
	db       *data.DB
	m        *power.Model
	tab      *danger.Table
	p        *survival.Params
	led      *survival.Ledger
	path     string
	power    float64
	lastDiff string
	pred     map[predKey]shadowPred
	warned   map[string]bool
}

// wireSurvival returns the ZoneChoice and ZoneObserve hooks; both nil without game tables or a loadable model. A ledger
// file that cannot be read starts empty (and a file from a newer build is left alone: saves are skipped).
func wireSurvival(db *data.DB, ledgerPath string) (func(*d2.Snapshot, int, string, int, []bot.ZoneCand) string, func(bot.ZoneEvent)) {
	tab := dangerTable(db)
	if tab == nil {
		return nil, nil
	}
	m, err := power.NewModel(tab.Params)
	if err != nil {
		slog.Warn("survival: no power model", "err", err)
		return nil, nil
	}
	p, err := survival.LoadParams()
	if err != nil {
		slog.Warn("survival: no params", "err", err)
		return nil, nil
	}
	w := &survivalShadow{db: db, m: m, tab: tab, p: p, path: ledgerPath, pred: map[predKey]shadowPred{}, warned: map[string]bool{}}
	led, _, err := survival.LoadLedger(ledgerPath, p)
	if err != nil {
		slog.Warn("survival: ledger not loaded, starting empty and not saving", "err", err)
		w.path = ""
	}
	w.led = led
	f := w.fight
	survFight.Store(&f)
	return w.choice, w.observe
}

func (w *survivalShadow) warn(err error) string {
	if k := err.Error(); !w.warned[k] {
		w.warned[k] = true
		slog.Warn("survival: no choice", "err", err)
	}
	return ""
}

func (w *survivalShadow) save() {
	if w.path == "" {
		return
	}
	if err := w.led.Save(w.path); err != nil && !w.warned["save"] {
		w.warned["save"] = true
		slog.Warn("survival: ledger not saved", "err", err)
	}
}

// choice is the bot.Deps.ZoneChoice hook: judge every candidate, pick, log, and publish the words-only block.
func (w *survivalShadow) choice(s *d2.Snapshot, area int, diff string, potions int, cands []bot.ZoneCand) string {
	w.mu.Lock()
	defer w.mu.Unlock()
	p := s.Player
	hero, a, err := heroOf(w.m, w.db, s, potions, w.warn)
	if err != nil {
		return w.warn(err)
	}
	dial, err := w.m.P.Dial(w.m.P.DefaultDial)
	if err != nil {
		return w.warn(err)
	}
	build := func(h power.Hero) ([]survival.Candidate, error) {
		var out []survival.Candidate
		for _, c := range cands {
			z, err := w.tab.Find(c.Area, diff)
			if err != nil {
				continue // an area the profile does not cover is not a candidate
			}
			zm, err := w.m.Zone(h, a, z, dial.ID)
			if err != nil {
				return nil, err
			}
			k := survival.ZoneKey{Area: c.Area, Diff: diff}
			if n := float64(z.Encounters.Typical.N); n > 0 {
				w.pred[k] = shadowPred{zm.Typical.Terms.Taken / n, zm.Typical.Terms.PerKill}
			}
			tm := shadowTravelMin
			if c.Area == area {
				tm = 0
			}
			out = append(out, survival.Candidate{Key: k, Name: z.Name, ZM: zm, Eco: survival.Economy(z, zm, h.Level, w.p), TravelMin: tm, Depleted: c.Depleted})
		}
		return out, nil
	}
	cs, err := build(hero)
	if err != nil {
		return w.warn(err)
	}
	need := 0
	if potions >= 0 && potions < shadowKitPotions {
		need = (shadowKitPotions - potions) * int(w.p.PotionPrice)
	}
	mercDead := hero.Merc.Hired && !hero.Merc.Alive
	if mercDead {
		need += shadowReviveQuote
	}
	w.power, w.lastDiff = survival.PowerOf(hero), diff
	in := survival.Input{Now: time.Now(), Dial: dial, Current: survival.ZoneKey{Area: area, Diff: diff}, Power: w.power,
		Purse: survival.Purse{Gold: p.Gold, Need: need}, Cands: cs, FreshAllowed: true, OffenceUnread: hero.OffenceUnread()}
	d, err := w.led.Choose(in)
	if err != nil {
		return w.warn(err)
	}
	lines := d.Line()
	var merc *survival.ReviveDecision
	if mercDead {
		alive := hero
		alive.Merc.Alive = true
		with, err1 := build(alive)
		if err1 == nil {
			rd := w.led.Revive(survival.ReviveInput{Base: in, With: with, Without: cs, Quote: shadowReviveQuote, Reserve: shadowKitPotions * int(w.p.PotionPrice), Gold: p.Gold})
			merc = &rd
			lines += "\nzone revive: " + rd.Why.Code + ": " + rd.Why.Text
		}
	}
	survGlance.Store(survival.GlanceOf(d, dial.ID, merc))
	return lines
}

// observe is the bot.Deps.ZoneObserve hook.
func (w *survivalShadow) observe(ev bot.ZoneEvent) {
	w.mu.Lock()
	defer w.mu.Unlock()
	k := survival.ZoneKey{Area: ev.Area, Diff: ev.Diff}
	now := ev.At
	if now.IsZero() {
		now = time.Now()
	}
	switch ev.Kind {
	case bot.ZoneEvChicken:
		w.led.Observe(k, now, survival.Event{Kind: survival.EvChicken}, w.power)
	case bot.ZoneEvDeath:
		w.led.Observe(k, now, survival.Event{Kind: survival.EvDeath}, w.power)
	case bot.ZoneEvMercDied:
		w.led.Observe(k, now, survival.Event{Kind: survival.EvMercDeath}, w.power)
	case bot.ZoneEvArea:
		w.led.Observe(k, now, survival.Event{Kind: survival.EvXP, XP: ev.XP}, w.power)
	default:
		return
	}
	w.save()
}

// fight feeds one finished fight (the fighter's OnSummary) with the model's prediction for the same kills, which is
// what k_zone and k_hero compare against. Without a prediction for the zone (no decision yet) the fight is skipped.
func (w *survivalShadow) fight(fs combat.FightSummary) {
	w.mu.Lock()
	defer w.mu.Unlock()
	pr, ok := w.pred[survival.ZoneKey{Area: fs.Area, Diff: w.lastDiff}]
	if !ok || fs.Kills <= 0 {
		return
	}
	n := float64(fs.Kills)
	w.led.Observe(survival.ZoneKey{Area: fs.Area, Diff: w.lastDiff}, time.Now(), survival.Event{Kind: survival.EvFight, Kills: n, Seconds: fs.Sec,
		Taken: float64(fs.Damage), PredTaken: pr.takenPerKill * n, PredKillSec: pr.secPerKill * n, Potions: float64(fs.Potions)}, w.power)
}
