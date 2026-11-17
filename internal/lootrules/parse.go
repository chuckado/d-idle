package lootrules

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Error is one validation finding; the editor highlights Path ("rules[3].when.all[1].v").
type Error struct {
	Path string `json:"path"`
	Msg  string `json:"msg"`
	Line int    `json:"line,omitempty"` // syntax errors only
	Col  int    `json:"col,omitempty"`
}

func (e Error) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("line %d col %d: %s", e.Line, e.Col, e.Msg)
	}
	return e.Path + ": " + e.Msg
}

// Errors is the list Parse and Validate return.
type Errors []Error

func (es Errors) Error() string {
	var s []string
	for _, e := range es {
		s = append(s, e.Error())
	}
	return strings.Join(s, "; ")
}

var qualities = map[string]bool{"low": true, "normal": true, "superior": true, "magic": true, "set": true, "rare": true, "unique": true, "crafted": true}
var scopes = map[string]bool{PlaceGround: true, PlacePack: true, PlaceStash: true, PlaceSell: true}
var actions = map[string]bool{Ask: true, KeepMerc: true, Pickup: true, Keep: true, Stash: true, Sell: true, Drop: true, Identify: true, Ignore: true, Vault: true}
var needKinds = map[string]bool{"potion": true, "scroll": true, "tome": true, "key": true, "gold": true, "gem": true, "runeword": true, "emptyslot": true, "charmroom": true}

// Parse reads and validates a rule set. A syntax error carries line and column; field errors carry the path of the field.
// Unknown keys are refused so a typo cannot silently disable a rule.
func Parse(b []byte) (*RuleSet, error) {
	var rs RuleSet
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&rs); err != nil {
		var se *json.SyntaxError
		if errors.As(err, &se) {
			line, col := 1, 1
			for _, c := range b[:min(int(se.Offset), len(b))] {
				if c == '\n' {
					line, col = line+1, 1
				} else {
					col++
				}
			}
			return nil, Errors{{Msg: se.Error(), Line: line, Col: col}}
		}
		return nil, Errors{{Path: "$", Msg: err.Error()}}
	}
	if es := rs.Validate(); len(es) > 0 {
		return nil, es
	}
	return &rs, nil
}

// Marshal writes the canonical form (export).
func (rs *RuleSet) Marshal() ([]byte, error) { return json.MarshalIndent(rs, "", " ") }

// allGroups merges the built-in and the set's own groups.
func (rs *RuleSet) allGroups() map[string]Group {
	m := map[string]Group{}
	for k, v := range builtinGroups {
		m[k] = v
	}
	for k, v := range rs.Groups {
		m[k] = v
	}
	return m
}

// Validate returns every finding (empty = valid).
func (rs *RuleSet) Validate() Errors {
	var es Errors
	add := func(p, f string, a ...any) { es = append(es, Error{Path: p, Msg: fmt.Sprintf(f, a...)}) }
	if rs.Schema != Schema {
		add("schema", "got %q, this build reads %q", rs.Schema, Schema)
	}
	if rs.Limits.MaxActions <= 0 {
		add("limits.maxActions", "must be at least 1")
	}
	for k := range rs.Groups {
		if _, bi := builtinGroups[k]; bi {
			add("groups."+k, "built-in group, pick another name")
		}
	}
	for sc, a := range rs.Defaults {
		if !scopes[sc] {
			add("defaults."+sc, "unknown scope")
		} else if a != Ignore && !legal(a, sc) {
			add("defaults."+sc, "action %q is not valid in scope %s", a, sc)
		}
	}
	groups := rs.allGroups()
	seen := map[string]bool{}
	for i, r := range rs.Rules {
		p := fmt.Sprintf("rules[%d]", i)
		switch {
		case r.ID == "":
			add(p+".id", "required")
		case seen[r.ID]:
			add(p+".id", "duplicate id %q", r.ID)
		}
		seen[r.ID] = true
		if r.Reason == "" {
			add(p+".reason", "required (it is the why tooltip)")
		}
		if !actions[r.Action] {
			add(p+".action", "unknown action %q", r.Action)
			continue
		}
		if r.Tier != "" && !has([]string{"S", "A", "B", "C"}, r.Tier) {
			add(p+".tier", "S, A, B or C")
		}
		if r.OnUnknown != "" && r.OnUnknown != "match" && r.OnUnknown != "skip" {
			add(p+".onUnknown", "match or skip")
		}
		for _, sc := range r.Scope {
			if !scopes[sc] {
				add(p+".scope", "unknown scope %q", sc)
			}
		}
		for _, sc := range r.effScope() {
			if scopes[sc] && !legal(r.Action, sc) && !(r.Action != Pickup && sc == PlaceGround && fate(r.Action)) {
				add(p+".action", "%q is not valid in scope %s", r.Action, sc)
			}
		}
		if r.When != nil {
			validCond(*r.When, p+".when", groups, add)
		}
	}
	return es
}

func fate(a string) bool { return a == Keep || a == KeepMerc || a == Stash || a == Sell || a == Drop }

// legal: which actions exist where. Fate actions on the ground are derived (see eval.go).
func legal(a, scope string) bool {
	switch scope {
	case PlaceGround:
		return a == Pickup || a == Ignore || a == Vault
	case PlacePack:
		return a != Pickup
	case PlaceStash:
		return a == Keep || a == Sell || a == Ignore
	case PlaceSell:
		return a == Keep || a == Sell || a == Stash || a == Ignore
	}
	return false
}

func (r Rule) effScope() []string {
	if len(r.Scope) == 0 {
		if r.Action == Pickup || r.Action == Vault {
			return []string{PlaceGround}
		}
		if r.Action == Identify {
			return []string{PlacePack}
		}
		return []string{PlaceGround, PlacePack}
	}
	return r.Scope
}

type fieldKind int

const (
	kNum fieldKind = iota
	kBool
	kEnum // eq ne in
	kList // in
	kName // contains
	kNeed
	kWho // usable_by: eq hero|merc
)

func fieldOf(f string) (fieldKind, bool) {
	switch f {
	case "ilvl", "reqlvl", "reqstr", "reqdex", "runeno", "grade", "maxsockets", "owned", "act", "durpct", "hero_usable_from", "merc_usable_from", "sockets", "value", "qty", "clvl", "gold", "packfull", "stashfull", "idtome":
		return kNum, true
	case "ethereal", "identified", "stackable", "quest", "broken":
		return kBool, true
	case "quality", "class", "build", "difficulty", "basetier":
		return kEnum, true
	case "type", "group", "code", "class_restricted":
		return kList, true
	case "name":
		return kName, true
	case "need":
		return kNeed, true
	case "usable_by":
		return kWho, true
	}
	if n, ok := strings.CutPrefix(f, "stat:"); ok && (statNames[n] || n == "ressum") {
		return kNum, true
	}
	return 0, false
}

func validCond(c Cond, p string, groups map[string]Group, add func(string, string, ...any)) {
	forms := 0
	if len(c.All) > 0 {
		forms++
	}
	if len(c.Any) > 0 {
		forms++
	}
	if c.Not != nil {
		forms++
	}
	if c.F != "" {
		forms++
	}
	if forms != 1 {
		add(p, "exactly one of all, any, not or a leaf (f, op, v) per condition")
		return
	}
	for i, x := range c.All {
		validCond(x, fmt.Sprintf("%s.all[%d]", p, i), groups, add)
	}
	for i, x := range c.Any {
		validCond(x, fmt.Sprintf("%s.any[%d]", p, i), groups, add)
	}
	if c.Not != nil {
		validCond(*c.Not, p+".not", groups, add)
	}
	if c.F == "" {
		return
	}
	k, ok := fieldOf(c.F)
	if !ok {
		add(p+".f", "unknown field %q", c.F)
		return
	}
	if c.Plus != nil && c.F != "usable_by" {
		add(p+".plus", "only usable_by takes plus")
	}
	var strs []string
	var num float64
	var b bool
	var s string
	bad := func(want string) { add(p+".v", "%s %s needs %s", c.F, c.Op, want) }
	switch k {
	case kNum:
		if !oneOf(c.Op, "eq", "ne", "ge", "gt", "le", "lt") {
			add(p+".op", "numbers use eq ne ge gt le lt")
		} else if json.Unmarshal(c.V, &num) != nil {
			bad("a number")
		}
	case kBool:
		if c.Op != "eq" {
			add(p+".op", "booleans use eq")
		} else if json.Unmarshal(c.V, &b) != nil {
			bad("true or false")
		}
	case kEnum:
		switch {
		case oneOf(c.Op, "eq", "ne"):
			if json.Unmarshal(c.V, &s) != nil {
				bad("a string")
			} else if c.F == "quality" && !qualities[s] {
				add(p+".v", "unknown quality %q", s)
			}
		case c.Op == "in":
			if json.Unmarshal(c.V, &strs) != nil || len(strs) == 0 {
				bad("a non-empty list")
			}
			for _, q := range strs {
				if c.F == "quality" && !qualities[q] {
					add(p+".v", "unknown quality %q", q)
				}
			}
		default:
			add(p+".op", "use eq ne in")
		}
	case kList:
		if c.Op != "in" {
			add(p+".op", "use in")
		} else if json.Unmarshal(c.V, &strs) != nil || len(strs) == 0 {
			bad("a non-empty list")
		}
		if c.F == "group" {
			for _, g := range strs {
				if _, ok := groups[g]; !ok {
					add(p+".v", "unknown group %q", g)
				}
			}
		}
	case kName:
		if c.Op != "contains" || json.Unmarshal(c.V, &s) != nil || s == "" {
			add(p+".v", "name uses contains with a non-empty string")
		}
	case kWho:
		if c.Op != "eq" || json.Unmarshal(c.V, &s) != nil || s != "hero" && s != "merc" {
			add(p+".v", "usable_by uses eq with hero or merc")
		}
	case kNeed:
		if c.Op != "eq" || json.Unmarshal(c.V, &s) != nil || !needKinds[s] {
			add(p+".v", "need uses eq with one of potion scroll tome key gold gem runeword charmroom")
		}
	}
}

func oneOf(s string, l ...string) bool {
	for _, x := range l {
		if s == x {
			return true
		}
	}
	return false
}
