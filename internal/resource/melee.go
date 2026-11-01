package resource

// The player's plain Attack (game.exe do 1, FUN_0056f070): to-hit and damage as docs/research/sk-melee-trace.md
// reads them (V read, V run against tools/emu/sk_melee.py). Every input is data the caller supplies from the
// character's stats and the weapon's tables; nothing here names a class, skill or weapon.

// BluntUndeadPct is the damage percent points a Blunt weapon adds against undead (FUN_0057b660, V run, trace D2).
const BluntUndeadPct = 50

// BluntType is the ItemTypes code the Blunt rule tests (ItemTypes row 57); staff, scepter and wand reach it through
// their Equiv1 / Equiv2 chain (trace D, the walk itself is I).
const BluntType = "blun"

// WeakMinHitPct is the lowest to-hit chance at which a pack still counts as weak for the melee fallback (U: a
// policy choice, not a game rule; half the swings landing is the least that "a swing is enough" can mean).
const WeakMinHitPct = 50

// Blow is the attacker side of one swing, in the stats game.exe reads.
type Blow struct {
	AR    int // attack rating before percent (AttackRating)
	ARPct int // mastery to-hit + stat 119 + stat 325
	Level int // attacker level (to-hit scales 2 x alvl / (alvl + dlvl))

	Weapon   bool // a weapon is wielded; false swings barehanded (trace D4: Strength counts as a damage percent)
	Min, Max int  // the player's stats 21 / 22 in whole points (weapon base with its enhanced damage, plus flats)
	Flat     int  // stat 111, added to both ends
	Str      int  // used only barehanded
	DmgPct   int  // stat 25 + Str / Dex share (StatDmgPct) + mastery damage + skill ED; everything except the Blunt rule
	MinPct   int  // stat 18 (0 under reading A of trace Q3)
	MaxPct   int  // stat 17 (0 under reading A)
	Blunt    bool // the weapon's type reaches BluntType
}

// Foe is the target side: its defense, level and whether it is undead.
type Foe struct {
	Defense int
	Level   int
	Undead  bool
}

// AttackRating is FUN_00622560: stat 19 + (dex - 7) x 5 + the class's ToHitFactor (V read).
func AttackRating(stat19, dex, toHitFactor int) int { return stat19 + (dex-7)*5 + toHitFactor }

// StatDmgPct is the share of FUN_0057b420's percent that a weapon swing takes from stats: gear stat 25, Strength
// and Dexterity through the weapon's StrBonus / DexBonus (C division), and the mastery damage.
func StatDmgPct(gear, str, dex, strBonus, dexBonus, mastery int) int {
	return gear + str*strBonus/100 + dex*dexBonus/100 + mastery
}

// IsBlunt reports whether weaponType reaches BluntType; isA is the item-type chain walk (data.DB.IsA).
func IsBlunt(isA func(typ, ancestor string) bool, weaponType string) bool {
	return isA != nil && isA(weaponType, BluntType)
}

// HitChance is FUN_0057d9b0's chance in percent (V run, 0 differences): AR' = AR + AR x pct / 100, then
// AR' x 100 / (AR' + DEF), scaled by 2 x alvl / (alvl + dlvl), 5 to 95.
func HitChance(ar, arPct, defense, alvl, dlvl int) int {
	attack := ar + ar*arPct/100
	if defense < 0 {
		attack, defense = attack-defense, 0
	}
	if attack < 0 {
		defense, attack = defense-attack, 0
	}
	chance := 100
	if attack+defense != 0 {
		chance = attack * 100 / (attack + defense)
	}
	chance = chance * 2 * alvl / max(alvl+dlvl, 1)
	return max(5, min(95, chance))
}

// HitPct is the chance this attacker lands a swing on f.
func (m Blow) HitPct(f Foe) int { return HitChance(m.AR, m.ARPct, f.Defense, m.Level, f.Level) }

// fineRange is FUN_0057b420's rolled interval [a, b) in 256ths (V run), with the Blunt rule applied for an
// undead foe when blunt is true.
func (m Blow) fineRange(f Foe, blunt bool) (a, b int) {
	lo, hi, p := m.Min, m.Max, m.DmgPct
	if !m.Weapon {
		if lo <= 0 {
			lo = 1
		}
		if hi <= 1 {
			hi = 2
		}
		p += m.Str
	}
	if blunt && f.Undead {
		p += BluntUndeadPct
	}
	p = max(p, -90)
	l, h := (lo+m.Flat)<<8, (hi+m.Flat)<<8
	if l < 1 {
		l = 0x100
	}
	if h <= l {
		h = l + 0x100
	}
	a = l + (l*(m.MinPct+p))/100
	b = h + (h*(m.MaxPct+p))/100
	return max(a, 0), max(b, 0)
}

// MeanDamage is the mean damage of a landed blow in points (no crit, no resistance), the roll being uniform on
// [a, b) in 256ths. Approximate: crit, deadly strike and the 256ths life tail are not modelled.
func (m Blow) MeanDamage(f Foe) float64 {
	a, b := m.fineRange(f, m.Blunt)
	if b <= a {
		return float64(a) / 256
	}
	return (float64(a) + float64(b-a-1)/2) / 256
}

// Expected is the mean damage of one swing at f, misses included.
func (m Blow) Expected(f Foe) float64 { return float64(m.HitPct(f)) / 100 * m.MeanDamage(f) }

// Swing rolls one blow: rnd(n) is a uniform integer in [0, n). It returns the points dealt (0 on a miss).
func (m Blow) Swing(f Foe, rnd func(n int) int) float64 {
	if rnd(100) >= m.HitPct(f) {
		return 0
	}
	a, b := m.fineRange(f, m.Blunt)
	if b > a {
		a += rnd(b - a)
	}
	return float64(a) / 256
}

// undeadGain is the factor the Blunt rule puts on a landed blow against f: 1 unless the weapon is Blunt and f undead.
func (m Blow) undeadGain(f Foe) float64 {
	if !m.Blunt || !f.Undead {
		return 1
	}
	with := m.MeanDamage(f)
	a, b := m.fineRange(f, false)
	without := float64(a)
	if b > a {
		without += float64(b-a-1) / 2
	}
	if without <= 0 {
		return 1
	}
	return with * 256 / without
}
