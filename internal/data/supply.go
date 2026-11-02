package data

// Per-kill supply of a Normal Act 1 / 2 kill, from docs/research/treasure-trace.md (V run: exact class odds checked by
// 104,000 native rolls of game.exe's roller; the pile formula is game.exe FUN_00557ab0). One player, MF and GF 0.
// The numbers describe the monster, not the character, so nothing here depends on the class or build.

// KillKind is what was killed.
type KillKind int

const (
	KillPlain  KillKind = iota // a Normal-rarity monster of a melee (H2H) class
	KillCaster                 // shamans, archers: fewer gold piles
	KillChampion
	KillUnique
	KillSuperUnique
)

// Supply is what one kill of a kind drops on average (trace section 5 table).
type Supply struct {
	Items     float64 // drops per kill (a plain kill: 1 - NoDrop 62.5 percent)
	GoldPiles float64 // gold piles per kill
	GoldMul   float64 // coin multiplier of a pile (champions x5: gld,mul=1280)
	Potions   float64 // healing + mana + rejuvenation potions per kill
}

var supplyTable = [...]Supply{
	KillPlain:       {Items: 0.375, GoldPiles: 0.131, GoldMul: 1, Potions: 0.057},
	KillCaster:      {Items: 0.375, GoldPiles: 0.1125, GoldMul: 1, Potions: 0.070},
	KillChampion:    {Items: 3, GoldPiles: 0.6, GoldMul: 5, Potions: 2},
	KillUnique:      {Items: 5, GoldPiles: 0, GoldMul: 1, Potions: 4},
	KillSuperUnique: {Items: 6, GoldPiles: 0, GoldMul: 1, Potions: 4},
}

// SupplyOf returns the per-kill supply of a kind.
func SupplyOf(k KillKind) Supply { return supplyTable[k] }

// GoldPileMean is the mean coins of one pile dropped by a monster of level mlvl: coins = ilvl + rand(5 ilvl) with
// ilvl = the monster level (at least 1), so ilvl .. 6 ilvl - 1, mean 3.5 ilvl - 0.5 (trace section 4, V run).
func GoldPileMean(mlvl int) float64 {
	return 3.5*float64(max(mlvl, 1)) - 0.5
}

// GoldPerKill is the mean gold of one kill of a kind at monster level mlvl (Blood Moor plain: 0.39).
func GoldPerKill(k KillKind, mlvl int) float64 {
	s := supplyTable[k]
	return s.GoldPiles * s.GoldMul * GoldPileMean(mlvl)
}

// PileAtLeast is the share of piles at monster level mlvl holding n or more coins (coins are uniform on
// ilvl .. 6 ilvl - 1): ilvl 1 never reaches 10, ilvl 2 20 percent, 3 53, 4 70, 5 80, 8 95.
func PileAtLeast(mlvl, n int) float64 {
	lo := max(mlvl, 1)
	hi := 6*lo - 1
	if n <= lo {
		return 1
	}
	if n > hi {
		return 0
	}
	return float64(hi-n+1) / float64(hi-lo+1)
}

// PileMinWorth is the largest pile size that at least share (0..1) of the piles at monster level mlvl reach, so a
// pile bar set from it never filters out most of an area. share <= 0 returns 0 (no bar).
func PileMinWorth(mlvl int, share float64) int {
	if share <= 0 {
		return 0
	}
	lo := max(mlvl, 1)
	n := 6*lo - 1
	for n > lo && PileAtLeast(lo, n) < share {
		n--
	}
	return n
}
