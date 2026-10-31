package d2dcombat

import (
	"context"
	"fmt"
	"sync"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dinput"
	"d2lod-bot/internal/env"
)

// Meter is the rig's clock: it steps the game through the Pacer and then reads "state" once, so the smoke tool can
// total damage taken, mana spent and deaths per virtual time without a second reader. One extra verb per Sleep.
//
// Two shapes: New's Meter is the clock itself (P). Attach's Meter wraps the backend clock (Clk) and is opt-in: until
// Enable is called it only forwards, so the bot's per-sleep cost is unchanged (docs/fixmeter.md).
type Meter struct {
	P   *d2dinput.Pacer
	C   *d2dctl.Client
	Ctx context.Context

	// Clk is the clock to wrap when P is nil. Off (Attach only) skips sampling.
	Clk env.Clock
	Off bool
	// Mons reads the snapshot whose monsters feed damage dealt and kills (nil: those metrics are not available).
	// Hostile filters which monsters count (nil: all).
	Mons    func() (*d2.Snapshot, error)
	Hostile func(*d2.Snapshot, d2.Unit) bool

	mu   sync.Mutex
	m    Totals
	last d2dctl.State
	have bool
	seen map[uint32]int // hostile monster id -> life at the last sample (alive ones only)
}

// Totals is what the Meter saw.
type Totals struct {
	Damage    int // sum of life drops between samples (life units as the channel reports them, V hit points)
	ManaSpent int // sum of mana drops between samples (potions and regeneration make this approximate)
	Deaths    int // transitions to life 0
	MinLife   int
	Samples   int

	// Monster-side totals, derived from the life of hostile monsters in the "monsters" listing (hp cur/max, V in
	// docs/d2d-backend.md; the derivation itself is UNVERIFIED live, docs/fixmeter.md). Valid only when DealtOK.
	DamageDealt int // sum of life drops of hostile monsters between samples, all sources (the player, minions, traps)
	Kills       int // hostile monsters seen alive that were then seen dying or dead
	DealtOK     bool
	// Hits and misses (swing results) are not carried by the d2d state: always not available, never a number.
}

// String is the one-line form for reports: numbers that d2d cannot supply print "n/a", never 0.
func (t Totals) String() string {
	dealt, kills := "n/a", "n/a"
	if t.DealtOK {
		dealt, kills = fmt.Sprint(t.DamageDealt), fmt.Sprint(t.Kills)
	}
	return fmt.Sprintf("taken=%d dealt=%s kills=%s hits=n/a misses=n/a mana=%d deaths=%d minlife=%d samples=%d",
		t.Damage, dealt, kills, t.ManaSpent, t.Deaths, t.MinLife, t.Samples)
}

func (m *Meter) clock() env.Clock {
	if m.P != nil {
		return m.P
	}
	return m.Clk
}

func (m *Meter) Now() time.Time { return m.clock().Now() }

func (m *Meter) Sleep(d time.Duration) {
	m.clock().Sleep(d)
	m.mu.Lock()
	off := m.Off
	m.mu.Unlock()
	if off {
		return
	}
	m.sample()
}

// Enable turns sampling on (Attach's Meter starts off) and forgets the last sample.
func (m *Meter) Enable() {
	m.mu.Lock()
	m.Off, m.have, m.seen = false, false, nil
	m.mu.Unlock()
}

// sampleMonsters folds one snapshot into the monster totals. Caller holds mu.
func (m *Meter) sampleMonsters(s *d2.Snapshot) {
	if s == nil {
		return
	}
	if m.seen == nil {
		m.seen = map[uint32]int{}
	}
	m.m.DealtOK = true
	now := map[uint32]*d2.Unit{}
	for i := range s.Monsters {
		now[s.Monsters[i].ID] = &s.Monsters[i]
	}
	for id, was := range m.seen {
		u, ok := now[id]
		switch {
		case !ok: // left the table: neither a kill nor damage (the same rule as combat.noteKills)
			delete(m.seen, id)
		case !u.Alive():
			m.m.DamageDealt += was // the last life read while alive; the dying read is not trusted, overkill is not counted
			m.m.Kills++
			delete(m.seen, id)
		default:
			if u.Monster != nil && u.Monster.Life < was {
				m.m.DamageDealt += was - u.Monster.Life
			}
			if u.Monster != nil {
				m.seen[id] = u.Monster.Life
			}
		}
	}
	for id, u := range now {
		if _, ok := m.seen[id]; ok || !u.Alive() || u.Type != d2.TypeMonster || u.Monster == nil || u.Monster.MaxLife <= 0 {
			continue
		}
		if m.Hostile == nil || m.Hostile(s, *u) {
			m.seen[id] = u.Monster.Life
		}
	}
}

func (m *Meter) sample() {
	ctx := m.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	st, _, err := m.C.State(ctx)
	var snap *d2.Snapshot
	if m.Mons != nil {
		snap, _ = m.Mons()
	}
	if err != nil || st.Life.Max == 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sampleMonsters(snap)
	if m.have {
		if d := m.last.Life.Cur - st.Life.Cur; d > 0 {
			m.m.Damage += d
		}
		if d := m.last.Mana.Cur - st.Mana.Cur; d > 0 {
			m.m.ManaSpent += d
		}
		if st.Life.Cur <= 0 && m.last.Life.Cur > 0 {
			m.m.Deaths++
		}
	}
	if !m.have || st.Life.Cur < m.m.MinLife {
		m.m.MinLife = st.Life.Cur
	}
	m.m.Samples++
	m.last, m.have = st, true
}

// Totals returns the counters so far.
func (m *Meter) Totals() Totals {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.m
}

// Reset forgets the last sample (call after a cheat that sets life or mana, so the jump is not counted).
func (m *Meter) Reset() {
	m.mu.Lock()
	m.have, m.seen = false, nil
	m.mu.Unlock()
}
