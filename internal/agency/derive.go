package agency

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/power"
)

// Allocation is what the player did with the character: hard points only (items and +skills are not counted).
type Allocation struct {
	Class         string
	Clvl          int
	Skills        map[string]int // hard points by skills.txt name
	Str, Dex      int            // attribute points added, not totals
	Vit, Ene      int
	QuestSkillPts int
}

func (a Allocation) spent() int {
	n := 0
	for _, p := range a.Skills {
		n += p
	}
	return n
}

func (a Allocation) clone() Allocation {
	b := a
	b.Skills = map[string]int{}
	for k, v := range a.Skills {
		b.Skills[k] = v
	}
	return b
}

// Weapon is what the bot has read of the equipped weapon. Avg 0 means not read.
type Weapon struct {
	Class string  // AnimData weapon class, "1HS"
	WSM   int     // weapon speed modifier
	Avg   float64 // average weapon damage
}

// Env is everything Derive reads besides the allocation.
type Env struct {
	Skills   *power.SkillData
	Profiles *power.Profiles
	Caps     *Caps
	Rules    *Rules // optional; enables the prereq_only warning
	Weapon   Weapon
	FCR, IAS int
	GearRead bool         // FCR and IAS were read from gear (0 is then a real 0)
	Combat   combat.Build // combat line of the class profile ("" = class default); used for plain Attack and contact
}

// NewEnv loads the shipped data.
func NewEnv() (Env, error) {
	sd, err := power.LoadSkillData()
	if err != nil {
		return Env{}, err
	}
	pr, err := power.Load()
	if err != nil {
		return Env{}, err
	}
	cp, err := LoadCaps()
	if err != nil {
		return Env{}, err
	}
	return Env{Skills: sd, Profiles: pr, Caps: cp}, nil
}

// Warning codes.
const (
	WNoAttack       = "no_attack"
	WTourist        = "tourist"
	WSingleElement  = "single_element"
	WUnsupportedPts = "unsupported_points"
	WWeaponUnread   = "weapon_unread"
	WManaStarved    = "mana_starved"
	WPrereqOnly     = "prereq_only"
	WSynergyUnknown = "synergy_unknown"
)

// Warning is one honest statement about what the bot cannot or should not do with this allocation.
type Warning struct{ Code, Text string }

// Choice is one decision with its reason.
type Choice struct {
	Skill      string
	Level      int
	Role       string
	Delivery   string
	Element    string
	Tier       string
	DPS        float64 // 0 = unknown (see Flags)
	ManaPerSec float64
	PerUse     float64
	Flags      []string
	Why        string
}

// Plan is how the bot would play the allocation.
type Plan struct {
	Main       *Choice  // what the character does: the best attack, or the top summon for a summoner
	Attack     *Choice  // the best personal attack (== Main unless the archetype is summoner); nil when none
	Secondary  *Choice  // best attack of another element than Attack (immunes)
	Candidates []Choice // every attack considered, best first
	Buffs      []Choice
	Aura       *Choice
	Summons    []Choice
	Movement   []Choice
	Archetype  string // "" when it could not be resolved
	// Best is the attack with the highest estimate; Prior is the combat profile's own pick (the hand-written band order);
	// Attack is the one the plan names: the Prior unless the estimate beats it by Caps.Planner.PriorMargin (docs/agency.md).
	Best, Prior string
	PriorAura   string // the combat profile's aura pick ("" none)
	// Spec and Arch are what the survival model needs to judge the hero with this plan (Arch is nil when the archetype
	// could not be resolved).
	Spec       power.BuildSpec
	Arch       *power.Archetype
	ManaPerSec float64
	ManaBudget float64
	Unplayed   Unplayed
	Warnings   []Warning
}

// Unplayed is points sitting in skills the engine cannot use.
type Unplayed struct {
	Count  int
	Skills map[string]int
}

// Has reports whether a warning code is present.
func (p Plan) Has(code string) bool {
	for _, w := range p.Warnings {
		if w.Code == code {
			return true
		}
	}
	return false
}

// Derive turns an allocation into a plan. Errors are bad input (unknown class or skill, no level), never a silent
// default.
func Derive(a Allocation, env Env) (Plan, error) {
	var pl Plan
	class := lower(a.Class)
	if !validClass(class) {
		return pl, fmt.Errorf("agency: unknown class %q", a.Class)
	}
	if a.Clvl < 1 {
		return pl, fmt.Errorf("agency: allocation of %s has clvl %d", class, a.Clvl)
	}
	if env.Skills == nil || env.Profiles == nil || env.Caps == nil {
		return pl, errors.New("agency: Env needs Skills, Profiles and Caps (use NewEnv)")
	}
	for _, n := range sortedKeys(a.Skills) {
		if a.Skills[n] < 0 {
			return pl, fmt.Errorf("agency: skill %q has %d points", n, a.Skills[n])
		}
		if !known(n, env.Caps, env.Profiles, env.Skills, env.Rules) {
			return pl, fmt.Errorf("agency: unknown skill %q (not in the capability table, power skillKinds, skilldamage or the rules)", n)
		}
	}
	cls, err := d2Class(class)
	if err != nil {
		return pl, err
	}
	prof, err := combat.LookupProfile(cls, env.Combat)
	if err != nil {
		return pl, err
	}
	d := &deriver{a: a, class: class, env: env, prof: prof, kinds: mergedKinds(a, env)}
	d.prior = ProfilePick(prof, a.Clvl, a.Skills)
	pl = d.run()
	pl.PriorAura = ProfileAura(prof, a.Clvl, a.Skills)
	return pl, nil
}

func d2Class(name string) (d2.Class, error) {
	for c := d2.Amazon; c <= d2.Assassin; c++ {
		if lower(c.String()) == name {
			return c, nil
		}
	}
	return 0, fmt.Errorf("agency: no d2 class for %q", name)
}

// mergedKinds is power's skillKinds plus what the capability table says for skills power does not list (Fire Bolt and
// Ice Bolt have no kind there, so Resolve would fail on them: found by the prototype).
func mergedKinds(a Allocation, env Env) map[string]string {
	k := map[string]string{}
	for n, v := range env.Profiles.SkillKinds {
		k[n] = v
	}
	for n := range a.Skills {
		if _, ok := k[n]; ok {
			continue
		}
		if c, ok := env.Caps.Skills[n]; ok {
			switch {
			case c.Role == RoleSummon:
				k[n] = "summon"
			case c.Role == RoleAttack && c.Delivery == "melee":
				k[n] = "melee"
			case c.Role == RoleAttack && c.Delivery == "projectile" && c.Element == "phys":
				k[n] = "missile"
			case c.Role == RoleAttack:
				k[n] = "spell"
			}
		}
	}
	k[power.PlainAttack] = "melee"
	return k
}

type deriver struct {
	a     Allocation
	class string
	env   Env
	prof  *combat.Profile
	kinds map[string]string
	prior string // the profile's pick, "" none
}

// ProfileAura is the combat profile's fighting aura for a character with these hard points: the first aura of the
// ladder it has a point in that is not a resting aura; "" when none.
func ProfileAura(prof *combat.Profile, clvl int, skills map[string]int) string {
	for _, a := range prof.AuraLadder(clvl, func(s string) bool { return skills[s] > 0 }) {
		if a.RestLifePct == 0 && skills[a.Name] > 0 {
			return a.Name
		}
	}
	return ""
}

// ProfilePick is the combat profile's current attack choice for a character with these hard points: the first skill of
// the active band that the character has a point in (the plain swing needs none); "" when the band names none.
func ProfilePick(prof *combat.Profile, clvl int, skills map[string]int) string {
	for _, n := range prof.AttackNames(clvl, func(s string) bool { return skills[s] > 0 }) {
		if skills[n] > 0 || n == power.PlainAttack {
			return n
		}
	}
	return ""
}

func (d *deriver) cap(n string) (Cap, bool) { return d.env.Caps.Lookup(n, d.kinds) }

func (d *deriver) run() Plan {
	var pl Plan
	a := d.a
	spent := a.spent()
	var skipped []string

	// 1. classify every skill with points.
	var attacks, buffs, auras, summons, moves []string
	unplayed := Unplayed{Skills: map[string]int{}}
	for _, n := range sortedKeys(a.Skills) {
		p := a.Skills[n]
		if p <= 0 {
			continue
		}
		c, ok := d.cap(n)
		if !ok || c.Tier == Unsupported {
			if ok && c.Role == RolePassive {
				continue
			}
			unplayed.Count += p
			unplayed.Skills[n] = p
			if ok && c.Role == RoleAttack {
				skipped = append(skipped, n)
			}
			continue
		}
		switch c.Role {
		case RoleAttack:
			attacks = append(attacks, n)
		case RoleBuff, RoleCurse:
			buffs = append(buffs, n)
		case RoleAura:
			auras = append(auras, n)
		case RoleSummon:
			summons = append(summons, n)
		case RoleMovement:
			moves = append(moves, n)
		}
	}
	// the plain weapon swing is the profile's last resort when the profile lists it
	if d.profileHas(power.PlainAttack) {
		attacks = append(attacks, power.PlainAttack)
	}
	pl.Unplayed = unplayed

	// 2. attack candidates with the data-derived damage.
	synUnknown := false
	for _, n := range attacks {
		ch, unknownSyn := d.estimate(n)
		synUnknown = synUnknown || unknownSyn
		pl.Candidates = append(pl.Candidates, ch)
	}
	sort.SliceStable(pl.Candidates, func(i, j int) bool {
		x, y := pl.Candidates[i], pl.Candidates[j]
		if (x.DPS > 0) != (y.DPS > 0) {
			return x.DPS > 0
		}
		if x.DPS != y.DPS {
			return x.DPS > y.DPS
		}
		if a.Skills[x.Skill] != a.Skills[y.Skill] {
			return a.Skills[x.Skill] > a.Skills[y.Skill]
		}
		return x.Skill < y.Skill
	})
	if len(pl.Candidates) > 0 {
		best := pl.Candidates[0]
		best.Why = d.whyAttack(best, pl.Candidates, skipped)
		pl.Best, pl.Prior = best.Skill, d.prior
		// The band order holds knowledge the estimate lacks (multi-strike, life cost): keep the profile's pick unless the
		// estimate beats it by a clear margin (design review 04 section 3.2).
		if i := d.priorIndex(pl.Candidates); i > 0 && best.DPS < pl.Candidates[i].DPS*d.env.Caps.Planner.PriorMargin {
			pr := pl.Candidates[i]
			pr.Why = fmt.Sprintf("the combat profile's pick; %s estimates %.1f dps against %.1f, under the %.1fx margin that overrules it", best.Skill, best.DPS, pr.DPS, d.env.Caps.Planner.PriorMargin)
			pl.Candidates = append([]Choice{pr}, append(pl.Candidates[:i:i], pl.Candidates[i+1:]...)...)
			best = pr
		}
		pl.Attack = &best
		pl.Candidates[0] = best
	}
	// Secondary: best attack of another element than the main one.
	if pl.Attack != nil && pl.Attack.Element != "" {
		for i := range pl.Candidates {
			c := pl.Candidates[i]
			if c.Element != "" && c.Element != pl.Attack.Element && c.DPS > 0 {
				c.Why = fmt.Sprintf("best %s attack (%.1f dps) for targets immune to %s", c.Element, c.DPS, pl.Attack.Element)
				pl.Secondary = &c
				break
			}
		}
	}

	// 3. support roles.
	pl.Buffs = d.choices(buffs, "buff up before a fight; %d points")
	pl.Summons = d.choices(summons, "keep up; %d points")
	pl.Movement = d.choices(moves, "movement skill; %d points")
	if ac := d.choices(auras, ""); len(ac) > 0 {
		sort.SliceStable(ac, func(i, j int) bool {
			if ac[i].Tier != ac[j].Tier {
				return ac[i].Tier == Supported
			}
			return ac[i].Level > ac[j].Level
		})
		x := ac[0]
		x.Why = fmt.Sprintf("aura with the most points among usable ones (%s tier, %d points)", x.Tier, x.Level)
		pl.Aura = &x
	}

	// 4. archetype by power's own rule.
	spec := power.BuildSpec{Class: d.class, Build: "agency", Points: a.Skills, Contact: d.prof.MeleeClass()}
	if pl.Attack != nil {
		spec.Attack = pl.Attack.Skill
		spec.AttackLevel = pl.Attack.Level
	}
	prof := *d.env.Profiles
	prof.SkillKinds = d.kinds
	pl.Spec = spec
	if arch, err := prof.Resolve(spec); err == nil {
		pl.Archetype, pl.Arch = arch.ID, arch
	}
	summoner := pl.Archetype == "summoner"
	if summoner && len(pl.Summons) > 0 {
		sort.SliceStable(pl.Summons, func(i, j int) bool { return pl.Summons[i].Level > pl.Summons[j].Level })
		m := pl.Summons[0]
		m.Why = fmt.Sprintf("summoner archetype (over a third of the points in summon skills): pets kill, %s has the most points (%d)", m.Skill, m.Level)
		pl.Main = &m
	} else {
		pl.Main = pl.Attack
	}

	// 5. mana.
	if pl.Attack != nil {
		pl.ManaPerSec = pl.Attack.ManaPerSec
	}
	pl.ManaBudget = d.manaBudget()

	// 6. warnings.
	d.warn(&pl, summoner, spent, synUnknown)
	return pl
}

// priorIndex is the place of the profile's pick among the candidates, -1 when it is not one (no points, unsupported, no
// estimate) or the margin is off.
func (d *deriver) priorIndex(c []Choice) int {
	if d.prior == "" || d.env.Caps.Planner.PriorMargin <= 0 {
		return -1
	}
	for i := range c {
		if c[i].Skill == d.prior && c[i].DPS > 0 {
			return i
		}
	}
	return -1
}

func (d *deriver) profileHas(skill string) bool {
	for _, b := range d.prof.Bands {
		if slices.Contains(b.Attack, skill) {
			return true
		}
	}
	return false
}

func (d *deriver) choices(names []string, why string) []Choice {
	var out []Choice
	for _, n := range names {
		c, _ := d.cap(n)
		ch := Choice{Skill: n, Level: d.a.Skills[n], Role: c.Role, Delivery: c.Delivery, Element: c.Element, Tier: c.Tier}
		if why != "" {
			ch.Why = fmt.Sprintf(why, ch.Level)
		}
		out = append(out, ch)
	}
	return out
}

// estimate runs power's damage estimate for one attack at the allocation's level and synergy points.
func (d *deriver) estimate(n string) (Choice, bool) {
	c, _ := d.cap(n)
	lvl := min(d.a.Skills[n], 20)
	if n == power.PlainAttack {
		lvl = 1
		c = Cap{Role: RoleAttack, Delivery: "melee", Tier: Supported, Element: "phys"}
	}
	ch := Choice{Skill: n, Level: lvl, Role: c.Role, Delivery: c.Delivery, Element: c.Element, Tier: c.Tier}
	at := power.Attack{Skill: n, Level: lvl, SynergyPoints: -1, FCR: d.env.FCR, IAS: d.env.IAS, FCRKnown: d.env.GearRead, IASKnown: d.env.GearRead}
	unknownSyn := false
	if c.Synergies != nil {
		at.SynergyPoints = d.synergyPoints(c.Synergies)
	} else if n != power.PlainAttack {
		unknownSyn = true
	}
	if d.env.Weapon.Avg > 0 {
		at.WeaponAvg, at.WeaponClass, at.WSM = d.env.Weapon.Avg, d.env.Weapon.Class, d.env.Weapon.WSM
	}
	est, err := d.env.Skills.Estimate(d.class, at)
	ch.Flags = est.Flags
	if err != nil {
		ch.Why = "no estimate: " + err.Error()
		return ch, unknownSyn
	}
	if est.Known() {
		ch.DPS, ch.PerUse, ch.ManaPerSec = est.DPS, est.PerUse, est.ManaPerSec
	}
	return ch, unknownSyn
}

// synergyPoints is the uniform points-per-synergy the estimator takes: the mean over the named synergy skills.
func (d *deriver) synergyPoints(names []string) float64 {
	if len(names) == 0 {
		return 0
	}
	t := 0
	for _, n := range names {
		t += d.a.Skills[n]
	}
	return float64(t) / float64(len(names))
}

func (d *deriver) whyAttack(best Choice, all []Choice, skipped []string) string {
	var b strings.Builder
	if best.DPS > 0 {
		fmt.Fprintf(&b, "highest estimated damage per second among %d usable attacks (%.1f dps at level %d)", len(all), best.DPS, best.Level)
	} else {
		fmt.Fprintf(&b, "no attack has a readable damage estimate; chosen by points (%d)", d.a.Skills[best.Skill])
	}
	if len(skipped) > 0 {
		fmt.Fprintf(&b, "; skipped unsupported: %s", strings.Join(skipped, ", "))
	}
	return b.String()
}

func (d *deriver) manaBudget() float64 {
	m := d.env.Caps.Mana
	pool := m.Base + m.PerEnergy[d.class]*float64(d.a.Ene) + m.PerLevel[d.class]*float64(d.a.Clvl-1)
	return pool/m.FightSeconds + pool*m.RegenFracPerSec
}

func (d *deriver) warn(pl *Plan, summoner bool, spent int, synUnknown bool) {
	add := func(code, f string, args ...any) {
		pl.Warnings = append(pl.Warnings, Warning{code, fmt.Sprintf(f, args...)})
	}
	a := d.a
	if pl.Attack == nil && !summoner {
		add(WNoAttack, "no usable attack skill (no points in a supported attack and the %s profile has no plain swing)", d.class)
	}
	if pl.Unplayed.Count > 0 {
		var names []string
		for _, n := range sortedKeys(pl.Unplayed.Skills) {
			names = append(names, fmt.Sprintf("%s %d", n, pl.Unplayed.Skills[n]))
		}
		add(WUnsupportedPts, "%d points sit in skills the engine cannot use: %s", pl.Unplayed.Count, strings.Join(names, ", "))
	}
	if pl.Attack != nil && !summoner {
		x := pl.Attack
		if x.Delivery == "melee" && d.env.Weapon.Avg <= 0 {
			add(WWeaponUnread, "%s is a weapon skill and the weapon was not read: damage is unknown, choice made by points", x.Skill)
		} else if slices.Contains(x.Flags, power.FlagWeaponUnread) {
			add(WWeaponUnread, "%s needs weapon damage that was not read", x.Skill)
		}
		if x.Element != "" && x.Element != "phys" && pl.Secondary == nil {
			add(WSingleElement, "every damaging skill is %s: the first %s-immune pack stops the character", x.Element, x.Element)
		}
		if pl.ManaPerSec > pl.ManaBudget {
			add(WManaStarved, "%s costs %.1f mana per second, the sustain budget at clvl %d with %d energy is %.1f", x.Skill, pl.ManaPerSec, a.Clvl, a.Ene, pl.ManaBudget)
		}
		if synUnknown && d.cap2(x.Skill).Synergies == nil && x.Skill != power.PlainAttack {
			add(WSynergyUnknown, "synergies of %s are not in the table: damage is a lower bound", x.Skill)
		}
	}
	// tourist: the main skill and what feeds it hold too small a share of the points.
	if a.Clvl >= 12 && spent > 0 && pl.Main != nil {
		mainPts := 0
		if summoner {
			for _, n := range sortedKeys(a.Skills) {
				if c, _ := d.cap(n); c.Role == RoleSummon {
					mainPts += a.Skills[n]
				}
			}
		} else {
			mainPts = a.Skills[pl.Main.Skill]
			for _, s := range d.cap2(pl.Main.Skill).Synergies {
				mainPts += a.Skills[s]
			}
		}
		if mainPts*100 < 35*spent {
			add(WTourist, "%s and its synergies hold %d of %d spent points (%d percent, under 35) at clvl %d", pl.Main.Skill, mainPts, spent, mainPts*100/spent, a.Clvl)
		}
	}
	if pl.Main == nil && a.Clvl >= 12 && spent > 0 {
		add(WTourist, "no main skill: %d points spread over %d skills", spent, len(a.Skills))
	}
	// prereq_only: a skill with points that no slot of the plan uses and that some other bought skill needs.
	if d.env.Rules != nil {
		used := map[string]bool{}
		for _, c := range append(append(append([]Choice{}, pl.Buffs...), pl.Summons...), pl.Movement...) {
			used[c.Skill] = true
		}
		for _, c := range []*Choice{pl.Main, pl.Attack, pl.Secondary, pl.Aura} {
			if c != nil {
				used[c.Skill] = true
			}
		}
		var only []string
		for _, n := range sortedKeys(a.Skills) {
			if a.Skills[n] <= 0 || used[n] {
				continue
			}
			if c, ok := d.cap(n); ok && c.Role == RolePassive {
				continue
			}
			for m, r := range a.Skills {
				if rs := d.env.Rules.Skills[m].ReqSkills; m != n && r > 0 && slices.Contains(rs[:], n) {
					only = append(only, n)
					break
				}
			}
		}
		if len(only) > 0 {
			add(WPrereqOnly, "points in %s serve only as prerequisites of other bought skills", strings.Join(only, ", "))
		}
	}
}

func (d *deriver) cap2(n string) Cap { c, _ := d.cap(n); return c }
