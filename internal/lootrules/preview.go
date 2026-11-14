package lootrules

// Sample is one recent drop or pack item for the dry run: the item as the bot saw it and the context at that moment.
type Sample struct {
	ID   uint32
	Item Item
	Ctx  Ctx
}

// RuleCount is one row of the live preview.
type RuleCount struct {
	Rule  string   `json:"rule"` // rule id, "default" or "invariant:<name>"
	Count int      `json:"count"`
	IDs   []uint32 `json:"ids"`
}

// Preview decides every sample (first match attribution) and counts per rule, in rule order, defaults and invariants last.
func (c *Compiled) Preview(samples []Sample) (counts []RuleCount, decisions map[uint32]Decision) {
	decisions = map[uint32]Decision{}
	idx := map[string]int{}
	var order []string
	for _, sc := range []string{PlaceGround, PlacePack, PlaceStash} {
		for _, r := range c.rules[sc] {
			if _, ok := idx[r.ID]; !ok {
				idx[r.ID] = len(order)
				order = append(order, r.ID)
			}
		}
	}
	rows := make([]RuleCount, len(order))
	for i, id := range order {
		rows[i].Rule = id
	}
	extra := map[string]*RuleCount{}
	var extraOrder []string
	for i := range samples {
		s := &samples[i]
		d := c.Decide(&s.Item, &s.Ctx)
		decisions[s.ID] = d
		if j, ok := idx[d.Rule]; ok {
			rows[j].Count++
			rows[j].IDs = append(rows[j].IDs, s.ID)
			continue
		}
		x := extra[d.Rule]
		if x == nil {
			x = &RuleCount{Rule: d.Rule}
			extra[d.Rule] = x
			extraOrder = append(extraOrder, d.Rule)
		}
		x.Count++
		x.IDs = append(x.IDs, s.ID)
	}
	counts = rows
	for _, id := range extraOrder {
		counts = append(counts, *extra[id])
	}
	return counts, decisions
}

// Change is one item whose answer differs between two rule sets ("with this change: 2 items would be sold instead of kept").
type Change struct {
	ID       uint32 `json:"id"`
	From, To string
}

// Diff compares what two compiled sets decide for the same samples.
func Diff(a, b *Compiled, samples []Sample) []Change {
	var out []Change
	for i := range samples {
		s := &samples[i]
		x, y := a.Decide(&s.Item, &s.Ctx), b.Decide(&s.Item, &s.Ctx)
		if x.Action != y.Action {
			out = append(out, Change{s.ID, x.Action, y.Action})
		}
	}
	return out
}

// Shadow reports a rule that can never match because an earlier rule of the same scope already catches everything it
// would: an earlier rule with no condition (and no class or build list), or the same condition. A conservative check:
// it finds no false positives, but misses a rule that is only shadowed by a combination of earlier rules.
type Shadow struct {
	Rule, By string
}

func (rs *RuleSet) Shadowed() []Shadow {
	var out []Shadow
	for i, r := range rs.Rules {
		if r.Off {
			continue
		}
	scopes:
		for _, sc := range r.effScope() {
			for _, e := range rs.Rules[:i] {
				if e.Off || !has(e.effScope(), sc) || len(e.Classes) > 0 || len(e.Builds) > 0 {
					continue
				}
				if e.When == nil || sameCond(e.When, r.When) {
					out = append(out, Shadow{r.ID, e.ID})
					break scopes
				}
			}
		}
	}
	return out
}

func sameCond(a, b *Cond) bool {
	if a == nil || b == nil {
		return a == b
	}
	x, _ := jsonOf(*a)
	y, _ := jsonOf(*b)
	return x == y
}
