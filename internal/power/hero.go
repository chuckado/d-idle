package power

import (
	"d2lod-bot/internal/d2"
)

// Quality says how far a power term can be trusted (design 4): measured = read from the game this step, derived =
// computed from tables and a read, default = not readable yet, a named constant stands in.
type Quality int

const (
	Default Quality = iota
	Derived
	Measured
)

func (q Quality) String() string { return [...]string{"default", "derived", "measured"}[q] }

// Stat ids the power estimate reads from the player's stat vector (ItemStatCost ids, S: the same list the melee model
// uses in internal/combat/melee.go; defence and block are U until the live checks of docs/survival/power.md run).
const (
	StatToHit      = 19
	StatBlock      = 20
	StatMinDmg     = 21
	StatMaxDmg     = 22
	StatDmgPct     = 25
	StatArmorClass = d2.StatArmorClass
	StatIAS        = 93  // item_fasterattackrate (ItemStatCost id: S, the same list as above; read when the vector has it)
	StatFCR        = 105 // item_fastercastrate (S)
)

// Term names, closed list: they are the keys of Hero.Quality and the codes of Result.Weakest.
const (
	TermLife    = "life"
	TermPotions = "potions"
	TermDefence = "defence"
	TermBlock   = "block"
	TermResist  = "resist"
	TermOffence = "offence"
	TermMerc    = "merc"
	TermSustain = "sustain"
)

// Hero is the character in the units of the model. A zero quality is Default, so a field nobody fills degrades.
type Hero struct {
	Class       string
	Level       int
	MaxLife     int
	MaxMana     int
	Defence     int // armour class including gear and Dexterity
	BlockPct    int
	Resist      [4]int // fire, cold, lightning, poison
	AR          int    // attack rating before percent (stat 19 plus the dexterity share), 0 = not read
	DmgMin      int    // weapon base damage with enhanced damage (stats 21 and 22), 0 = not read
	DmgMax      int
	DmgPct      int
	Potions     int     // healing potions that can reach the belt
	PotionHeal  float64 // life per potion when known, 0 = use the default
	Sustain     float64 // life per second from regeneration and auras
	MeasuredDPS float64 // kill speed observed against this monster class over about 20 kills, 0 = none
	Merc        Merc
	Attack      Attack // main attack skill and gear speed for the data estimate; Attack.Skill "" = no build known
	Quality     map[string]Quality
}

// Merc is the hireling term. Level 0 with Alive true means the level is not read.
type Merc struct {
	Hired bool
	Alive bool
	Level int
}

// Reader is the seam to the game: what the estimator needs, read once per plan step. The live implementation reads the
// snapshot (HeroFromSnapshot); tests use Fake.
type Reader interface{ Hero() (Hero, error) }

// Fake is a Reader with a fixed hero.
type Fake struct {
	H   Hero
	Err error
}

func (f Fake) Hero() (Hero, error) { return f.H, f.Err }

// q returns the quality of a term; missing is Default.
func (h *Hero) q(term string) Quality { return h.Quality[term] }

func (h *Hero) set(term string, q Quality) {
	if h.Quality == nil {
		h.Quality = map[string]Quality{}
	}
	h.Quality[term] = q
}

// OffenceUnread is true when the weapon terms were not read, so the kill speed is a named constant (quality default).
func (h *Hero) OffenceUnread() bool { return h.q(TermOffence) == Default }

// SetQuality is for callers that fill a Hero by hand.
func (h *Hero) SetQuality(term string, q Quality) { h.set(term, q) }

// HeroFromPlayer builds a Hero from the player read and the merc sheet. potions is the healing potion count the kit
// knows (negative: the belt was not read). An unreadable term keeps quality Default and its field 0, never a guess.
func HeroFromPlayer(p *d2.Player, merc *d2.MercSheet, potions int) Hero {
	h := Hero{Class: p.Class.String(), Level: p.Level, MaxLife: p.MaxLife, MaxMana: p.MaxMana}
	if p.MaxLife > 0 {
		h.set(TermLife, Measured)
	}
	st := p.Stats
	if st.Has(StatArmorClass) {
		h.Defence = int(st.Get(StatArmorClass))
		h.set(TermDefence, Measured)
	}
	if st.Has(StatBlock) {
		h.BlockPct = int(st.Get(StatBlock))
		h.set(TermBlock, Derived) // gear sum only on an older fork; with the "hero" dump the fight model total (no skills, auras)
	}
	h.Resist = [4]int{p.Resist.Fire, p.Resist.Cold, p.Resist.Lightning, p.Resist.Poison}
	h.set(TermResist, Derived) // the penalty basis of the d2d panel value is U (TASK-216)
	if st.Has(StatMinDmg) && st.Has(StatMaxDmg) && st.Get(StatMaxDmg) > 0 {
		h.DmgMin, h.DmgMax = int(st.Get(StatMinDmg)), int(st.Get(StatMaxDmg))
		h.DmgPct = int(st.Get(StatDmgPct))
		h.AR = int(st.Get(StatToHit)) + (p.Dexterity-7)*5
		h.set(TermOffence, Derived)
	}
	if st.Has(StatFCR) {
		h.Attack.FCR, h.Attack.FCRKnown = int(st.Get(StatFCR)), true
	}
	if st.Has(StatIAS) {
		h.Attack.IAS, h.Attack.IASKnown = int(st.Get(StatIAS)), true
	}
	if potions >= 0 {
		h.Potions = potions
		h.set(TermPotions, Derived) // the count is read, the heal of the tier is a constant
	}
	if merc != nil {
		h.Merc = Merc{Hired: merc.Hired, Alive: merc.Alive, Level: merc.Level}
		h.set(TermMerc, Derived)
	}
	return h
}
