package combat

import (
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/resource"
)

// Stat ids the melee model reads from the player's merged list (docs/research/sk-melee-trace.md: the ids are V read in
// game.exe; that the player's list carries them in these units is U until drill R1, docs/skmeleewire.md).
const (
	statToHit      = 19  // attack rating from gear
	statMinDamage  = 21  // the wielded weapon's minimum, merged into the player's list
	statMaxDamage  = 22  // and its maximum
	statDmgPct     = 25  // gear damage percent
	statDmgFlat    = 111 // flat damage added to both ends
	statToHitPct   = 119 // gear to-hit percent
	statToHitBonus = 325 // progressive to-hit (0 for classes without it)
)

// rangedType is the ItemTypes code of bows and crossbows (S: itemtypes.txt "Missile Weapon"); a left click with one
// shoots, so the melee model does not describe it and the policy keeps its level-only rules.
const rangedType = "miss"

// meleeBlow builds the blow model of the character's plain Attack from the snapshot and the game tables, or nil when
// an input is missing: no table, no player or stat vector, no inventory (the wielded item is unknown), an unknown
// class or weapon row, a ranged or thrown weapon, or a weapon whose damage stats 21 / 22 are not in the list.
// Not modelled (0): the class's weapon mastery to-hit and damage (FUN_00645830), stats 17 / 18 (reading A).
func meleeBlow(p *d2.Player, inv *d2.Inventory, db *data.DB) *resource.Blow {
	if p == nil || inv == nil || db == nil || len(p.Stats) == 0 {
		return nil
	}
	cc := db.ClassByName(p.Class.String())
	if cc == nil {
		return nil
	}
	st := p.Stats
	b := &resource.Blow{
		AR:    resource.AttackRating(int(st.Get(statToHit)), p.Dexterity, cc.ToHitFactor),
		ARPct: int(st.Get(statToHitPct) + st.Get(statToHitBonus)),
		Level: p.Level,
		Flat:  int(st.Get(statDmgFlat)),
		Str:   p.Strength,
		Min:   int(st.Get(statMinDamage)),
		Max:   int(st.Get(statMaxDamage)),
	}
	gear := int(st.Get(statDmgPct))
	u := inv.Equipped(d2.BodyRHand)
	if u == nil {
		b.DmgPct = gear // barehanded: Strength counts as a percent through Blow.Str
		return b
	}
	if u.Class < 0 || u.Class >= len(db.Items) || db.Items[u.Class] == nil || !st.Has(statMinDamage) || !st.Has(statMaxDamage) {
		return nil
	}
	w := db.Items[u.Class]
	if w.Kind != data.KindWeapon || w.Throwable || db.IsA(w.Type, rangedType) {
		return nil
	}
	b.Weapon = true
	b.DmgPct = resource.StatDmgPct(gear, p.Strength, p.Dexterity, w.StrBonus, w.DexBonus, 0)
	b.Blunt = resource.IsBlunt(db.IsA, w.Type)
	return b
}

// meleeFoe is the target side for the pack inside radius: the hostile the blow is least likely to hit (its defense
// and level from monstats at the character's difficulty; U: areas and bosses that override them are not read) and
// whether every one of them is undead. nil when nobody is in range or a hostile's class is unknown.
func meleeFoe(p *d2.Player, hs []Target, radius int, b *resource.Blow) *resource.Foe {
	if p == nil || b == nil {
		return nil
	}
	var foe resource.Foe
	n, undead := 0, true
	for _, h := range hs {
		if h.Dist > float64(radius) {
			continue
		}
		if h.Mon == nil {
			return nil
		}
		f := resource.Foe{Defense: h.Mon.AC[p.Difficulty], Level: h.Mon.Level[p.Difficulty]}
		if n == 0 || b.HitPct(f) < b.HitPct(foe) {
			foe = f
		}
		undead = undead && (h.Mon.LUndead || h.Mon.HUndead)
		n++
	}
	if n == 0 {
		return nil
	}
	foe.Undead = undead
	return &foe
}
