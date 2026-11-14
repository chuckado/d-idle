package lootrules

import "fmt"

// Requirement predicates (docs/lootrules.md section 2.4). The hero's and the mercenary's numbers are inputs of the
// evaluation, not facts of the item, so a cached decision that used them is stale when they change (Fingerprint).

func (cx *Ctx) actor(who string) *Actor {
	if who == "merc" {
		return cx.Merc
	}
	return cx.Hero
}

// never is the "usable_from" answer for an item the wearer can not wear at any level (class, hireling type, strength or
// dexterity short). Stat points could change the last two: U, the rule sees "never" until the stats rise.
const never = 999

// blocked reports what keeps the actor from wearing the item apart from level: class or hireling type, strength,
// dexterity. unknown = the actor's numbers are unread.
func blocked(it *Item, a *Actor, who string, plus Plus) (permanent bool, unknown bool) {
	if a == nil {
		return true, false
	}
	if a.Level == 0 {
		return false, true
	}
	if who == "merc" {
		if !has(it.MercTypes, a.Class) {
			return true, false
		}
	} else if it.ClassOnly != "" && it.ClassOnly != a.Class {
		return true, false
	}
	return it.ReqStr > a.Str+plus.Str || it.ReqDex > a.Dex+plus.Dex, false
}

func usable(it *Item, a *Actor, plus Plus, who string) tri {
	perm, unk0 := blocked(it, a, who, plus)
	switch {
	case unk0:
		return unk
	case perm:
		return fal
	}
	return b2t(it.ReqLvl <= a.Level+plus.Lvl)
}

// usableFrom is the lowest level at which the actor can wear the item: 0 now, the required level when only the level
// stands in the way, 999 when something else does.
func usableFrom(it *Item, a *Actor, who string) (int, bool) {
	perm, unk0 := blocked(it, a, who, Plus{})
	switch {
	case unk0:
		return 0, true
	case perm:
		return never, false
	case it.ReqLvl <= a.Level:
		return 0, false
	}
	return it.ReqLvl, false
}

// Fingerprint changes when the hero's or the mercenary's class, level, strength or dexterity changes. A caller that
// caches decisions re-evaluates them (L2, L3 cards) when it changes; Compiled.UsesActors says whether it matters.
func (cx *Ctx) Fingerprint() string {
	f := func(a *Actor) string {
		if a == nil {
			return "-"
		}
		return fmt.Sprintf("%s/%d/%d/%d", a.Class, a.Level, a.Str, a.Dex)
	}
	return f(cx.Hero) + "|" + f(cx.Merc)
}
