package lootrules

import (
	"fmt"
	"sort"
)

type crule struct {
	Rule
	m       matcher
	scopeOf map[string]bool
}

// Compiled is a rule set bound to one hero (class and build): disabled rules and other classes' rules are gone and every
// condition is a closure, so a decision is a linear walk with no parsing and no allocation. Safe for concurrent use.
type Compiled struct {
	rs        *RuleSet
	rules     map[string][]crule // by scope
	protected []Group
	supplies  Group
	acked     bool
	// UsesActors: some rule reads the hero's or the mercenary's numbers, so decisions must be redone when
	// Ctx.Fingerprint changes.
	UsesActors bool
}

// Compile validates and compiles rs for one hero. class and build are data keys (d2.Class.String(), the leveler build
// name); an unknown class simply matches only the rules without a class list.
func Compile(rs *RuleSet, class, build string) (*Compiled, error) {
	if es := rs.Validate(); len(es) > 0 {
		return nil, es
	}
	groups := rs.allGroups()
	c := &Compiled{rs: rs, rules: map[string][]crule{}, supplies: groups["supplies"]}
	for _, g := range protectedGroups {
		c.protected = append(c.protected, groups[g])
	}
	c.acked = has(rs.Ack, "drop-valuable")
	for _, r := range rs.Rules {
		if r.Off || len(r.Classes) > 0 && !has(r.Classes, class) || len(r.Builds) > 0 && !has(r.Builds, build) {
			continue
		}
		cr := crule{Rule: r, scopeOf: map[string]bool{}}
		if r.When != nil {
			c.UsesActors = c.UsesActors || usesActors(*r.When)
			cr.m = compileCond(*r.When, groups)
		}
		for _, sc := range r.effScope() {
			cr.scopeOf[sc] = true
			c.rules[sc] = append(c.rules[sc], cr)
		}
	}
	return c, nil
}

func why(a, reason string, it *Item) string {
	v := map[string]string{Pickup: "Will pick up", Keep: "Kept", KeepMerc: "Kept for your mercenary", Ask: "Asking you", Stash: "Will stash", Sell: "Will sell", Drop: "Will drop",
		Identify: "Will identify", Ignore: "Left alone"}[a]
	if a == Sell || a == Drop {
		return fmt.Sprintf("%s: %s, value %d", v, reason, it.Value)
	}
	return v + ": " + reason
}

// Decide answers for one item. Order, none of it data: place, the player's mark, start kit, quest item; then the rules of
// the item's scope (first match wins, else the scope default); then the floors (stash full, supplies, valuables).
func (c *Compiled) Decide(it *Item, cx *Ctx) Decision {
	sc := it.Place
	if !scopes[sc] {
		return Decision{Action: Ignore, Rule: "invariant:place", Why: "Left alone: only ground, backpack and stash items are decided"}
	}
	switch {
	case sc == PlaceGround && it.Quest:
		return Decision{Action: Pickup, Rule: "invariant:quest", Why: "Will pick up: quest item", Prio: 100}
	case sc == PlaceGround:
	case it.Mark == "keep":
		return Decision{Action: Ignore, Rule: "invariant:mark", Why: "Kept: you marked it keep"}
	case it.Mark == "sell":
		return Decision{Action: Ignore, Rule: "invariant:mark", Why: "Will sell: you marked it for sale"}
	case it.StartItem:
		return Decision{Action: Keep, Rule: "invariant:start", Why: "Kept: starting gear"}
	case it.Quest && (sc == PlacePack || sc == PlaceSell):
		return Decision{Action: Keep, Rule: "invariant:quest", Why: "Kept: quest item"}
	case it.Quest:
		return Decision{Action: Ignore, Rule: "invariant:quest", Why: "Left alone: quest item"}
	}
	pending := false
	for i := range c.rules[sc] {
		r := &c.rules[sc][i]
		t := tru
		if r.m != nil {
			t = r.m(it, cx)
		}
		if t == unk {
			pending = true
			if r.OnUnknown != "match" {
				continue
			}
			t = tru
		}
		if t == fal {
			continue
		}
		d := c.finish(r.Action, r.Reason, r.ID, r.Prio, r.Pickup, r.Confirm, sc, it, cx)
		d.Pending, d.Tier = pending, r.Tier
		return d
	}
	a := c.rs.Defaults[sc]
	if a == "" || a == Ignore {
		return Decision{Action: Ignore, Rule: "default", Why: "Left alone: no rule", Pending: pending}
	}
	d := c.finish(a, "default", "default", 0, nil, false, sc, it, cx)
	d.Pending = pending
	return d
}

func (c *Compiled) finish(a, reason, id string, prio int, pickup *bool, confirm bool, sc string, it *Item, cx *Ctx) Decision {
	fateOf := a
	if sc == PlaceGround && a != Pickup && a != Ignore {
		take := pickup != nil && *pickup || pickup == nil && (a == Keep || a == KeepMerc || a == Stash)
		a = map[bool]string{true: Pickup, false: Ignore}[take]
	}
	switch a {
	case Stash:
		if m := c.rs.Limits.StashMaxPct; m > 0 && cx.StashFullPct >= m {
			return Decision{Action: Keep, Rule: id, Why: "Kept: the stash is full (" + reason + ")"}
		}
	case Sell, Drop:
		if c.supplies.has(it) {
			return Decision{Action: Keep, Rule: "invariant:supplies", Why: "Kept: tomes, scrolls and keys are never sold or dropped"}
		}
		if a == Drop && !dropSafe(it) || c.isProtected(it) {
			if !(confirm && c.acked) {
				return Decision{Action: Ignore, Rule: "invariant:valuable", Why: "Left alone: a valuable item is never sold or dropped without your confirmation"}
			}
		}
	}
	d := Decision{Action: a, Rule: id, Why: why(a, reason, it), Prio: prio}
	if fateOf != a {
		d.Fate = fateOf
	}
	return d
}

// dropSafe is the floor under every drop: only low, normal and superior items; magic or better, or an item whose quality
// is not readable, needs a confirmed rule and the set's "drop-valuable" acknowledgement.
func dropSafe(it *Item) bool {
	if it.Hidden&HQuality != 0 {
		return false
	}
	switch it.Quality {
	case "low", "normal", "superior":
		return true
	}
	return false
}

func (c *Compiled) isProtected(it *Item) bool {
	for _, g := range c.protected {
		if g.has(it) {
			return true
		}
	}
	return false
}

// Trace is one line of the "why not" explanation: how every rule of the scope judged the item.
type Trace struct {
	Rule   string `json:"rule"`
	Result string `json:"result"` // match, no, unknown
}

// Explain lists every rule's verdict for the item (the preview endpoint uses it); it does not stop at the first match.
func (c *Compiled) Explain(it *Item, cx *Ctx) []Trace {
	var out []Trace
	for i := range c.rules[it.Place] {
		r := &c.rules[it.Place][i]
		t := tru
		if r.m != nil {
			t = r.m(it, cx)
		}
		out = append(out, Trace{r.ID, [...]string{"no", "match", "unknown"}[t]})
	}
	return out
}

// PlanItem pairs an id with an item.
type PlanItem struct {
	ID   uint32
	Item *Item
}

// Entry is one planned action.
type Entry struct {
	ID uint32
	Decision
}

// Plan decides every pack item and returns the actions of the given kinds (default stash and drop), identify first, then
// stash (best value first), sell (best first), drop (cheapest first), at most Limits.MaxActions in total; ties by id so
// the order is the same every run.
func (c *Compiled) Plan(items []PlanItem, cx *Ctx, kinds ...string) []Entry {
	if len(kinds) == 0 {
		kinds = []string{Stash, Drop}
	}
	by := map[string][]Entry{}
	val := map[uint32]int{}
	for _, pi := range items {
		it := *pi.Item
		it.Place = PlacePack
		d := c.Decide(&it, cx)
		if has(kinds, d.Action) {
			by[d.Action] = append(by[d.Action], Entry{pi.ID, d})
			val[pi.ID] = it.Value
		}
	}
	var out []Entry
	for _, k := range []string{Identify, Stash, Sell, Drop} {
		es := by[k]
		sort.SliceStable(es, func(i, j int) bool {
			a, b := val[es[i].ID], val[es[j].ID]
			if a != b {
				return (a > b) != (k == Drop)
			}
			return es[i].ID < es[j].ID
		})
		out = append(out, es...)
	}
	if n := c.rs.Limits.MaxActions; len(out) > n {
		out = out[:n]
	}
	return out
}

func usesActors(c Cond) bool {
	switch c.F {
	case "usable_by", "hero_usable_from", "merc_usable_from":
		return true
	}
	for _, x := range c.All {
		if usesActors(x) {
			return true
		}
	}
	for _, x := range c.Any {
		if usesActors(x) {
			return true
		}
	}
	return c.Not != nil && usesActors(*c.Not)
}
