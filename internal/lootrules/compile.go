package lootrules

import (
	"encoding/json"
	"strings"
)

type tri int8

const (
	fal tri = iota
	tru
	unk // a field the decision point cannot read: the rule neither matches nor fails
)

type matcher func(*Item, *Ctx) tri

func b2t(b bool) tri {
	if b {
		return tru
	}
	return fal
}

func and(ms []matcher) matcher {
	return func(it *Item, cx *Ctx) tri {
		r := tru
		for _, m := range ms {
			switch m(it, cx) {
			case fal:
				return fal
			case unk:
				r = unk
			}
		}
		return r
	}
}

func or(ms []matcher) matcher {
	return func(it *Item, cx *Ctx) tri {
		r := fal
		for _, m := range ms {
			switch m(it, cx) {
			case tru:
				return tru
			case unk:
				r = unk
			}
		}
		return r
	}
}

func not(m matcher) matcher {
	return func(it *Item, cx *Ctx) tri {
		switch m(it, cx) {
		case tru:
			return fal
		case fal:
			return tru
		}
		return unk
	}
}

func cmp(op string, a, b int) bool {
	switch op {
	case "eq":
		return a == b
	case "ne":
		return a != b
	case "ge":
		return a >= b
	case "gt":
		return a > b
	case "le":
		return a <= b
	}
	return a < b
}

func magicPlus(q string) bool {
	switch q {
	case "magic", "set", "rare", "unique", "crafted":
		return true
	}
	return false
}

// statsHidden: an unidentified magic or better item shows no property values (S: the game hides them), and an adapter may
// also flag HStats (d2d ground labels carry text only).
func statsHidden(it *Item) bool {
	return it.Hidden&HStats != 0 || magicPlus(it.Quality) && (!it.Identified || it.Hidden&HIdentified != 0)
}

// compileCond builds the closure tree. The condition has been validated.
func compileCond(c Cond, groups map[string]Group) matcher {
	switch {
	case len(c.All) > 0:
		return and(compileAll(c.All, groups))
	case len(c.Any) > 0:
		return or(compileAll(c.Any, groups))
	case c.Not != nil:
		return not(compileCond(*c.Not, groups))
	}
	var num int
	var f float64
	var b bool
	var s string
	var l []string
	if json.Unmarshal(c.V, &f) == nil {
		num = int(f)
	}
	_ = json.Unmarshal(c.V, &b)
	_ = json.Unmarshal(c.V, &s)
	_ = json.Unmarshal(c.V, &l)
	op := c.Op
	numOf := func(get func(*Item, *Ctx) (int, bool)) matcher {
		return func(it *Item, cx *Ctx) tri {
			v, hidden := get(it, cx)
			if hidden {
				return unk
			}
			return b2t(cmp(op, v, num))
		}
	}
	itemNum := func(v func(*Item) int, h uint8) matcher {
		return numOf(func(it *Item, _ *Ctx) (int, bool) { return v(it), h != 0 && it.Hidden&h != 0 })
	}
	ctxNum := func(v func(*Ctx) int) matcher {
		return numOf(func(_ *Item, cx *Ctx) (int, bool) { return v(cx), false })
	}
	boolOf := func(v func(*Item) bool, h uint8) matcher {
		return func(it *Item, _ *Ctx) tri {
			if h != 0 && it.Hidden&h != 0 {
				return unk
			}
			return b2t(v(it) == b)
		}
	}
	enum := func(v func(*Item, *Ctx) string, h uint8) matcher {
		return func(it *Item, cx *Ctx) tri {
			if h != 0 && it.Hidden&h != 0 {
				return unk
			}
			got, hit := v(it, cx), false
			if op == "in" {
				hit = has(l, got)
			} else {
				hit = got == s
			}
			return b2t(hit != (op == "ne"))
		}
	}
	switch c.F {
	case "ilvl":
		return itemNum(func(it *Item) int { return it.ILvl }, HILvl)
	case "reqlvl":
		return itemNum(func(it *Item) int { return it.ReqLvl }, 0)
	case "sockets":
		return itemNum(func(it *Item) int { return it.Sockets }, 0)
	case "value":
		return itemNum(func(it *Item) int { return it.Value }, HValue)
	case "qty":
		return itemNum(func(it *Item) int { return it.Qty }, 0)
	case "clvl":
		return ctxNum(func(cx *Ctx) int { return cx.CLvl })
	case "gold":
		return ctxNum(func(cx *Ctx) int { return cx.Gold })
	case "packfull":
		return ctxNum(func(cx *Ctx) int { return cx.PackFullPct })
	case "idtome":
		return ctxNum(func(cx *Ctx) int {
			if cx.IDTome {
				return 1
			}
			return 0
		})
	case "stashfull":
		return ctxNum(func(cx *Ctx) int { return cx.StashFullPct })
	case "runeno":
		return itemNum(func(it *Item) int { return it.RuneNo }, 0)
	case "grade":
		return itemNum(func(it *Item) int { return it.Grade }, 0)
	case "maxsockets":
		return itemNum(func(it *Item) int { return it.MaxSockets }, 0)
	case "act":
		return numOf(func(_ *Item, cx *Ctx) (int, bool) { return cx.Act, cx.Act == 0 })
	case "owned":
		return numOf(func(it *Item, cx *Ctx) (int, bool) {
			if cx.Owned == nil {
				return 0, false
			}
			return cx.Owned(it), false
		})
	case "basetier":
		return enum(func(it *Item, _ *Ctx) string { return it.BaseTier }, 0)
	case "reqstr":
		return itemNum(func(it *Item) int { return it.ReqStr }, 0)
	case "reqdex":
		return itemNum(func(it *Item) int { return it.ReqDex }, 0)
	case "durpct":
		return itemNum(func(it *Item) int {
			if it.MaxDur <= 0 {
				return 100
			}
			return it.Durability * 100 / it.MaxDur
		}, 0)
	case "broken":
		return boolOf(func(it *Item) bool { return it.MaxDur > 0 && it.Durability <= 0 }, 0)
	case "class_restricted":
		return func(it *Item, _ *Ctx) tri { return b2t(it.ClassOnly != "" && has(l, it.ClassOnly)) }
	case "usable_by":
		who, plus := s, Plus{}
		if c.Plus != nil {
			plus = *c.Plus
		}
		return func(it *Item, cx *Ctx) tri { return usable(it, cx.actor(who), plus, who) }
	case "hero_usable_from":
		return numOf(func(it *Item, cx *Ctx) (int, bool) { return usableFrom(it, cx.actor("hero"), "hero") })
	case "merc_usable_from":
		return numOf(func(it *Item, cx *Ctx) (int, bool) { return usableFrom(it, cx.actor("merc"), "merc") })
	case "ethereal":
		return boolOf(func(it *Item) bool { return it.Ethereal }, HIdentified)
	case "identified":
		return boolOf(func(it *Item) bool { return it.Identified }, HIdentified)
	case "stackable":
		return boolOf(func(it *Item) bool { return it.Stackable }, 0)
	case "quest":
		return boolOf(func(it *Item) bool { return it.Quest }, 0)
	case "quality":
		return enum(func(it *Item, _ *Ctx) string { return it.Quality }, HQuality)
	case "class":
		return enum(func(_ *Item, cx *Ctx) string { return cx.Class }, 0)
	case "build":
		return enum(func(_ *Item, cx *Ctx) string { return cx.Build }, 0)
	case "difficulty":
		return enum(func(_ *Item, cx *Ctx) string { return cx.Difficulty }, 0)
	case "type":
		return func(it *Item, _ *Ctx) tri {
			for _, t := range it.Types {
				if has(l, t) {
					return tru
				}
			}
			return fal
		}
	case "code":
		return func(it *Item, _ *Ctx) tri {
			for _, p := range l {
				if globMatch(p, it.Code) {
					return tru
				}
			}
			return fal
		}
	case "group":
		gs := make([]Group, len(l))
		for i, n := range l {
			gs[i] = groups[n]
		}
		return func(it *Item, _ *Ctx) tri {
			for _, g := range gs {
				if g.has(it) {
					return tru
				}
			}
			return fal
		}
	case "name":
		want := strings.ToLower(s)
		return func(it *Item, _ *Ctx) tri { return b2t(strings.Contains(strings.ToLower(it.Name), want)) }
	case "need":
		return func(it *Item, cx *Ctx) tri { return b2t(cx.Needs != nil && cx.Needs(s, it)) }
	}
	name := strings.TrimPrefix(c.F, "stat:")
	return numOf(func(it *Item, _ *Ctx) (int, bool) { return statOf(it, name), statsHidden(it) })
}

func compileAll(cs []Cond, groups map[string]Group) []matcher {
	ms := make([]matcher, len(cs))
	for i, c := range cs {
		ms[i] = compileCond(c, groups)
	}
	return ms
}

func has(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}
