package survival

import (
	"fmt"
	"math"
	"time"

	"d2lod-bot/internal/store"
)

// LedgerVersion is the schema version of the ledger document (store.File envelope). The ledger lives in its own file
// next to deaths.json, so a bot without this package never reads it (design 10: additive, versioned).
const LedgerVersion = 1

// ZoneKey is one (area, difficulty) pair.
type ZoneKey struct {
	Area int
	Diff string
}

func (k ZoneKey) String() string { return fmt.Sprintf("%d/%s", k.Area, k.Diff) }

// Event kinds a caller reports.
const (
	EvFight     = "fight" // Kills, Seconds, Taken, PredTaken, PredKillSec, Potions, XP, GoldIn, GoldOut
	EvChicken   = "chicken"
	EvDeath     = "death"
	EvMercDeath = "merc_death"
)

// Event is one observation. Taken is life lost in the fight; PredTaken and PredKillSec are what the model predicted for
// the same kills (the caller knows them; the ledger only forms the ratios).
type Event struct {
	Kind                                          string
	Kills, Seconds, Taken, PredTaken, PredKillSec float64
	Potions, XP, GoldIn, GoldOut                  float64
}

// Ev is one hard event kept for the step-down rule and the revive rule.
type Ev struct {
	T    time.Time `json:"t"`
	Kind string    `json:"kind"`
}

// ZoneStats is the ledger row of one zone. The decayed sums halve every HalfLifeMinutes of field time in the zone.
type ZoneStats struct {
	Kills      float64 `json:"kills"`
	Seconds    float64 `json:"seconds"`
	Taken      float64 `json:"taken"`
	PredTaken  float64 `json:"predTaken"`
	Potions    float64 `json:"potions"`
	Chickens   float64 `json:"chickens"`
	MercDeaths float64 `json:"mercDeaths"`
	Deaths     float64 `json:"deaths"`
	XP         float64 `json:"xp"`
	GoldIn     float64 `json:"goldIn"`
	GoldOut    float64 `json:"goldOut"`

	Hard      float64   `json:"hard"` // product of the hard-event multipliers, relaxing with time away (>= 1; 0 = 1)
	HardAt    time.Time `json:"hardAt"`
	PowerAt   float64   `json:"powerAt"` // the hero power the row was learned at
	Recent    []Ev      `json:"recent,omitempty"`
	LeftAt    time.Time `json:"leftAt"`
	LeftPower float64   `json:"leftPower"`
	NoRevive  time.Time `json:"noReviveUntil"`
}

// Ledger is the persisted document: per zone rows and the pooled offence correction.
type Ledger struct {
	Zones    map[string]*ZoneStats `json:"zones"`
	HeroSec  float64               `json:"heroSec"`  // observed seconds of fighting, pooled
	HeroPred float64               `json:"heroPred"` // predicted seconds for the same kills
	HeroKill float64               `json:"heroKills"`
	P        *Params               `json:"-"`
}

// NewLedger is an empty ledger under p.
func NewLedger(p *Params) *Ledger { return &Ledger{Zones: map[string]*ZoneStats{}, P: p} }

// LedgerFile is the store file of the ledger at path.
func LedgerFile(path string) *store.File[Ledger] {
	return &store.File[Ledger]{Path: path, Version: LedgerVersion}
}

// LoadLedger reads the ledger; a missing file is an empty ledger. A file from a newer build is an error and is not
// touched (store.ErrFutureVersion).
func LoadLedger(path string, p *Params) (*Ledger, store.Info, error) {
	l, info, err := LedgerFile(path).Load()
	if err != nil {
		return NewLedger(p), info, err
	}
	if l.Zones == nil {
		l.Zones = map[string]*ZoneStats{}
	}
	l.P = p
	return &l, info, nil
}

// Save writes the ledger.
func (l *Ledger) Save(path string) error { return LedgerFile(path).Save(*l) }

func (l *Ledger) row(k ZoneKey) *ZoneStats {
	if s := l.Zones[k.String()]; s != nil {
		return s
	}
	s := &ZoneStats{}
	l.Zones[k.String()] = s
	return s
}

func (l *Ledger) peek(k ZoneKey) *ZoneStats { return l.Zones[k.String()] }

func decay(seconds, halfLifeMin float64) float64 { return math.Pow(0.5, seconds/(halfLifeMin*60)) }

func (s *ZoneStats) hard() float64 {
	if s.Hard < 1 {
		return 1
	}
	return s.Hard
}

// hardNow is the hard-event multiplier relaxed toward 1 by the time since the last hard event.
func (l *Ledger) hardNow(s *ZoneStats, now time.Time) float64 {
	h := s.hard()
	if h <= 1 || s.HardAt.IsZero() {
		return 1
	}
	return 1 + (h-1)*decay(now.Sub(s.HardAt).Seconds(), l.P.HardHalfLifeMinutes)
}

func (l *Ledger) clamp(v float64) float64 { return math.Min(l.P.KMax, math.Max(l.P.KMin, v)) }

// Observe adds one event to the zone's row. power is the hero power estimate now (Refresh handles a changed one).
func (l *Ledger) Observe(k ZoneKey, now time.Time, e Event, power float64) {
	s := l.row(k)
	l.refresh(s, power)
	s.LeftAt = time.Time{}
	switch e.Kind {
	case EvFight:
		f := decay(e.Seconds, l.P.HalfLifeMinutes)
		s.Kills, s.Seconds, s.Taken, s.PredTaken = s.Kills*f, s.Seconds*f, s.Taken*f, s.PredTaken*f
		s.Potions, s.XP, s.GoldIn, s.GoldOut = s.Potions*f, s.XP*f, s.GoldIn*f, s.GoldOut*f
		s.Chickens, s.MercDeaths, s.Deaths = s.Chickens*f, s.MercDeaths*f, s.Deaths*f
		s.Kills += e.Kills
		s.Seconds += e.Seconds
		s.Taken += e.Taken
		s.PredTaken += e.PredTaken
		s.Potions += e.Potions
		s.XP += e.XP
		s.GoldIn += e.GoldIn
		s.GoldOut += e.GoldOut
		h := decay(e.Seconds, l.P.HeroHalfLifeMinutes)
		l.HeroSec, l.HeroPred, l.HeroKill = l.HeroSec*h+e.Seconds, l.HeroPred*h+e.PredKillSec, l.HeroKill*h+e.Kills
	case EvChicken, EvDeath, EvMercDeath:
		mul := map[string]float64{EvChicken: l.P.ChickenMul, EvDeath: l.P.DeathMul, EvMercDeath: l.P.MercDeathMul}[e.Kind]
		s.Hard = l.hardNow(s, now) * mul // the earlier events have relaxed by now; the new one lands in full
		s.HardAt = now
		switch e.Kind {
		case EvChicken:
			s.Chickens++
		case EvDeath:
			s.Deaths++
		case EvMercDeath:
			s.MercDeaths++
		}
		s.Recent = append(s.Recent, Ev{T: now, Kind: e.Kind})
		if len(s.Recent) > 16 {
			s.Recent = s.Recent[len(s.Recent)-16:]
		}
		if e.Kind == EvMercDeath && l.count(s, now, EvMercDeath) >= 2 {
			s.NoRevive = now.Add(time.Duration(l.P.NoReviveMinutes * float64(time.Minute)))
		}
	}
}

// refresh drops what was learned about a different hero: when the power moved by more than PowerChange since the row
// was learned, the average relaxes (sums reset) and the hard multiplier goes back to 1 (design 5.4).
func (l *Ledger) refresh(s *ZoneStats, power float64) {
	if power <= 0 {
		return
	}
	if s.PowerAt > 0 && math.Abs(power/s.PowerAt-1) > l.P.PowerChange {
		s.Kills, s.Seconds, s.Taken, s.PredTaken, s.Potions = 0, 0, 0, 0, 0
		s.Hard, s.HardAt = 1, time.Time{}
	}
	if s.PowerAt == 0 || math.Abs(power/s.PowerAt-1) > l.P.PowerChange {
		s.PowerAt = power
	}
}

// Refresh applies the power-change rule to a zone's row (the controller calls it for every candidate).
func (l *Ledger) Refresh(k ZoneKey, power float64) {
	if s := l.peek(k); s != nil {
		l.refresh(s, power)
	}
}

// count is the number of events of the kinds in the window ending at now ("" counts chickens and merc deaths).
func (l *Ledger) count(s *ZoneStats, now time.Time, kinds ...string) int {
	n := 0
	w := time.Duration(l.P.WindowMinutes * float64(time.Minute))
	for _, e := range s.Recent {
		if now.Sub(e.T) > w {
			continue
		}
		for _, k := range kinds {
			if e.Kind == k {
				n++
			}
		}
	}
	return n
}

// KHero is the offence correction pooled over all zones: observed kill seconds over predicted, shrunk toward 1 with few
// kills. It is 1 with no data.
func (l *Ledger) KHero() float64 {
	if l.HeroPred <= 0 || l.HeroKill <= 0 {
		return 1
	}
	w := l.HeroKill / (l.HeroKill + l.P.ShrinkKills)
	return l.clamp(1 + w*(l.HeroSec/l.HeroPred-1))
}

// KZone multiplies the zone's predicted damage taken (design 5.4): the observed ratio net of KHero (which already
// carries the offence error), shrunk with few kills, times the hard-event multiplier relaxed by time away.
func (l *Ledger) KZone(k ZoneKey, now time.Time) float64 {
	s := l.peek(k)
	if s == nil {
		return 1
	}
	soft := 1.0
	if s.PredTaken > 0 && s.Kills > 0 {
		w := s.Kills / (s.Kills + l.P.ShrinkKills)
		soft = 1 + w*((s.Taken/s.PredTaken)/l.KHero()-1)
	}
	return l.clamp(soft * l.hardNow(s, now))
}

// ForcedOut reports whether the observed cost of the zone in the window crosses the dial's step-down row, and why.
func (l *Ledger) ForcedOut(k ZoneKey, dial string, now time.Time) (bool, string) {
	s := l.peek(k)
	if s == nil {
		return false, ""
	}
	row, err := l.P.stepDown(dial)
	if err != nil {
		return false, ""
	}
	if row.Deaths > 0 && l.count(s, now, EvDeath) >= row.Deaths {
		return true, "died"
	}
	if row.Either > 0 && l.count(s, now, EvChicken, EvMercDeath) >= row.Either {
		return true, "hard_events"
	}
	return false, ""
}

// MarkLeft records that the hero left k for cost at this power: the zone stays closed until the power moves by more than
// PowerChange. Any later observation in the zone (the hero is back in it) lifts the lock (design 5.4).
func (l *Ledger) MarkLeft(k ZoneKey, now time.Time, power float64) {
	s := l.row(k)
	s.LeftAt, s.LeftPower = now, power
}

// Locked reports whether a left-for-cost lock on k still holds at this power.
func (l *Ledger) Locked(k ZoneKey, power float64) bool {
	s := l.peek(k)
	if s == nil || s.LeftAt.IsZero() {
		return false
	}
	if s.LeftPower > 0 && power > 0 && math.Abs(power/s.LeftPower-1) > l.P.PowerChange {
		return false
	}
	return true
}

// NoReviveNow reports whether the zone is closed to reviving (two merc deaths in the window, not yet expired).
func (l *Ledger) NoReviveNow(k ZoneKey, now time.Time) bool {
	s := l.peek(k)
	return s != nil && now.Before(s.NoRevive)
}

// MercMinutes is the expected field minutes until the merc dies again in k: the observed rate, else the prior.
func (l *Ledger) MercMinutes(k ZoneKey) float64 {
	if s := l.peek(k); s != nil && s.MercDeaths >= 0.5 && s.Seconds > 0 {
		return math.Max(1, s.Seconds/60/s.MercDeaths)
	}
	return l.P.MercPriorMinutes
}
