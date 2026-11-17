package lootrules

import (
	"encoding/json"
	"fmt"
	"strings"
)

var actionWords = map[string]string{
	Pickup: "Pick up", Keep: "Keep", Stash: "Stash", Sell: "Sell", Drop: "Drop on the floor", Identify: "Identify",
	Ignore: "Leave alone", Ask: "Ask you about", KeepMerc: "Keep for the mercenary", Vault: "Record in the vault",
}

var opWords = map[string]string{"eq": "is", "ne": "is not", "ge": "at least", "gt": "more than", "le": "at most", "lt": "less than", "in": "is one of", "contains": "contains"}

// Describe is the rule as one plain sentence for the Loot Rules tab: where it applies, what it does, and when.
func Describe(r Rule) string {
	where := map[string]string{PlaceGround: "On the ground", PlacePack: "In the pack", PlaceStash: "In the stash", PlaceSell: "At the vendor"}
	var sc []string
	for _, s := range r.effScope() {
		sc = append(sc, where[s])
	}
	act := actionWords[r.Action]
	if act == "" {
		act = r.Action
	}
	out := strings.Join(sc, " and ") + ": " + strings.ToLower(act[:1]) + act[1:]
	if r.When != nil {
		out += " when " + plain(*r.When, false)
	}
	if len(r.Classes) > 0 {
		out += " (only " + strings.Join(r.Classes, ", ") + ")"
	}
	return out
}

func contains(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

var labels map[string]string

func labelOf(f string) string {
	if labels == nil {
		labels = map[string]string{}
		for _, d := range Fields() {
			labels[d.Name] = strings.ToLower(d.Label)
		}
	}
	if l, ok := labels[f]; ok {
		return l
	}
	return f
}

func plain(c Cond, nested bool) string {
	switch {
	case len(c.All) > 0:
		return join(c.All, " and ", nested, false)
	case len(c.Any) > 0:
		return join(c.Any, " or ", nested, true)
	case c.Not != nil:
		return "not (" + plain(*c.Not, false) + ")"
	}
	val := valueText(c.V)
	switch {
	case c.F == "need":
		return "the bot needs more " + val
	case c.Op == "eq" && (val == "true" || val == "false"):
		if val == "true" {
			return "it is " + labelOf(c.F)
		}
		return "it is not " + labelOf(c.F)
	}
	op := opWords[c.Op]
	if op == "" {
		op = c.Op
	}
	return labelOf(c.F) + " " + op + " " + val
}

func join(cs []Cond, sep string, nested, paren bool) string {
	var parts []string
	for _, c := range cs {
		parts = append(parts, plain(c, true))
	}
	s := strings.Join(parts, sep)
	if nested && paren {
		return "(" + s + ")"
	}
	return s
}

func valueText(v json.RawMessage) string {
	var l []any
	if json.Unmarshal(v, &l) == nil {
		var parts []string
		for _, x := range l {
			parts = append(parts, fmt.Sprint(x))
		}
		if len(parts) > 8 {
			parts = append(parts[:8], fmt.Sprintf("and %d more", len(l)-8))
		}
		return strings.Join(parts, ", ")
	}
	var x any
	if json.Unmarshal(v, &x) == nil {
		return fmt.Sprint(x)
	}
	return string(v)
}
