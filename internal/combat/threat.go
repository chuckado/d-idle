package combat

import (
	"sort"
	"strings"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/input"
)

// Missile threat model (card botthreat; docs/botthreat.md). Everything is read from the loaded tables, nothing is
// keyed on a monster or class name except the AI column (runners).

// framesPerSec and subtilesPerVel: path speed per frame = Vel x 0.046875 subtiles, 25 frames per second
// (S: docs/research/monster-missile-trace.md Q6; the same arithmetic reproduces its reach column).
const (
	framesPerSec   = 25.0
	subtilesPerVel = 0.046875
)

// MissileThreat is what the bot knows about one missiles.txt row.
type MissileThreat struct {
	Name      string
	Speed     float64 // subtiles per second, 0 when the row has no Vel
	Dmg       int     // MaxDamage column (the row's own damage; monster missiles take most damage from their owner: I)
	ToHit     bool    // the row rolls to-hit (blockable, dodgeable)
	Blockable bool    // I: the row carries physical damage or rolls to-hit (V trace: all ToHit rows are blockable)
}

// ThreatFromRow derives a MissileThreat from a missiles.txt row.
func ThreatFromRow(m *data.Missile) MissileThreat {
	return MissileThreat{
		Name: m.Name, Speed: float64(m.Vel) * subtilesPerVel * framesPerSec, Dmg: m.MaxDamage,
		ToHit: m.ToHit, Blockable: m.ToHit || m.MaxDamage > 0,
	}
}

// ThreatTable is keyed by missile id.
type ThreatTable map[int]MissileThreat

// BuildThreatTable reads every row of the loaded missiles table. A nil DB gives a nil table.
func BuildThreatTable(db *data.DB) ThreatTable {
	if db == nil {
		return nil
	}
	t := ThreatTable{}
	for _, m := range db.Missiles {
		t[m.ID] = ThreatFromRow(m)
	}
	return t
}

// Verdict is the answer to "what do I do about this missile".
type Verdict int

const (
	VerdictDodge       Verdict = iota // slow enough to leave its path
	VerdictIgnore                     // too fast to dodge and no shooter to answer
	VerdictKillShooter                // too fast to dodge: the shooter is the target
)

func (v Verdict) String() string { return [...]string{"dodge", "ignore", "kill-shooter"}[v] }

// Classify decides for a missile dist subtiles away. reactSec is the time from poll to the character actually
// moving (I: 0.1-0.3 s for a poll plus a click). An unknown row (no speed) keeps the old behaviour: dodge.
func (t MissileThreat) Classify(dist, reactSec float64, shooterKnown bool) Verdict {
	if t.Speed <= 0 || dist/t.Speed >= reactSec {
		return VerdictDodge
	}
	if shooterKnown {
		return VerdictKillShooter
	}
	return VerdictIgnore
}

// Incoming is the ranked view of the hostile missiles near the player.
type Incoming struct {
	Dodge    []input.Point // positions worth stepping away from
	Shooters []uint32      // owner ids of missiles that cannot be dodged
}

// RankIncoming classifies every hostile missile within radius subtiles, most damaging first. A class missing from
// the table counts as dodgeable (the old rule).
func RankIncoming(snap *d2.Snapshot, tbl ThreatTable, radius, reactSec float64) Incoming {
	var in Incoming
	if snap == nil || snap.Player == nil {
		return in
	}
	me := input.Point{X: snap.Player.X, Y: snap.Player.Y}
	owners := map[uint32]bool{}
	for _, u := range snap.Monsters {
		owners[u.ID] = true
	}
	type scored struct {
		pt  input.Point
		dmg int
	}
	var dodge []scored
	seen := map[uint32]bool{}
	for _, m := range hostileMissileUnits(snap) {
		pt := input.Point{X: m.X, Y: m.Y}
		d := pdist(me, pt)
		if d >= radius {
			continue
		}
		th, ok := tbl[m.Class]
		if !ok {
			dodge = append(dodge, scored{pt, 0})
			continue
		}
		switch th.Classify(d, reactSec, owners[m.OwnerID] && m.OwnerType == 1) {
		case VerdictDodge:
			dodge = append(dodge, scored{pt, th.Dmg})
		case VerdictKillShooter:
			if !seen[m.OwnerID] {
				seen[m.OwnerID] = true
				in.Shooters = append(in.Shooters, m.OwnerID)
			}
		}
	}
	sort.SliceStable(dodge, func(i, j int) bool { return dodge[i].dmg > dodge[j].dmg })
	for _, s := range dodge {
		in.Dodge = append(in.Dodge, s.pt)
	}
	return in
}

// hostileMissileUnits is hostileMissiles with the units kept.
func hostileMissileUnits(snap *d2.Snapshot) []d2.Unit {
	if snap == nil || snap.Player == nil {
		return nil
	}
	friendly := Friendlies(snap, nil)
	var out []d2.Unit
	for _, m := range snap.Missiles {
		if _, ok := friendly[m.OwnerID]; ok || (m.OwnerID == snap.Player.ID && m.OwnerType == 0) {
			continue
		}
		out = append(out, m)
	}
	return out
}

// ShooterPeak is the largest single ranged hit a monster row can deal at difficulty d, from the monstats columns
// (attack 1, attack 2, special 1 and the elemental slots). unblockable is I: a ranged monster whose attack columns
// carry no physical damage but an elemental one (Skeleton Mage, Bighead rows in the trace's list). Zero for a melee row.
func ShooterPeak(m *data.Monster, d d2.Difficulty) (peak int, unblockable bool) {
	if m == nil || !m.Ranged || d < 0 || int(d) > 2 {
		return 0, false
	}
	phys := max(m.A1Max[d], m.A2Max[d])
	elem := 0
	for _, e := range m.Elem {
		elem = max(elem, e.Max)
	}
	peak = max(phys, m.S1Max[d], elem)
	return peak, phys == 0 && elem > 0
}

// runnerAIs are the AI column values whose think sets a pace and runs at the player (V run, docs/research/
// monster-ai-trace.md D1-D5; the Lancer and Arach spellings are U). A walk-away retreat does not break contact.
var runnerAIs = map[string]bool{"zombie": true, "brute": true, "bloodhawk": true, "corruptrogue": true,
	"corruptlancer": true, "arach": true}

// Runs reports whether the monster closes on a retreating player: a runner AI, or a flyer (V: blood hawks charge).
func Runs(m *data.Monster) bool {
	return m != nil && (runnerAIs[strings.ToLower(m.AI)] || m.Flying)
}

// standerAIs wait for the player until it comes inside their aware distance (S: docs/research/mon-wake-trace.md),
// so "does not approach" is normal for them.
var standerAIs = map[string]bool{"fallen": true}

// Stands reports whether a monster that does not come to the player is behaving normally.
func Stands(m *data.Monster) bool { return m != nil && standerAIs[strings.ToLower(m.AI)] }

// anyRunner reports whether a living hostile within d subtiles runs.
func anyRunner(hs []Target, d float64) bool {
	for _, h := range hs {
		if h.Dist <= d && Runs(h.Mon) {
			return true
		}
	}
	return false
}
