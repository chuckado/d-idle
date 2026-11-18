package lootrules

import "fmt"

// Options are the few knobs of the simple editor (docs/lootrules.md "Wave 2a"). They never invent rules: each one edits
// a named rule of the preset (by id), so the result is an ordinary rule set that the full engine, the preview and the
// shadow stage handle like any other.
type Options struct {
	// SellGear is the least quality of magic-or-better gear the bot picks up only to sell it: "magic" (the preset's
	// own line), "rare" (rare and crafted) or "never".
	SellGear string `json:"sellGear,omitempty"`
	// MinRune is the lowest rune number kept (1 = El, so every rune; 1..33). Lower runes are left on the ground and
	// are not stashed.
	MinRune int `json:"minRune,omitempty"`
	// DropSockets false switches off the socketed-base rules (runeword bases on the ground, 4+ socket bases in the pack).
	KeepSockets *bool `json:"keepSockets,omitempty"`
	// MinGold is the smallest gold pile picked up when the bot does not need gold (0 = every pile).
	MinGold int `json:"minGold,omitempty"`
}

// Defaults are the options that leave a preset as it is.
func (o Options) normalized() Options {
	if o.SellGear == "" {
		o.SellGear = "magic"
	}
	if o.MinRune < 1 {
		o.MinRune = 1
	}
	return o
}

// Validate refuses an option outside its range, naming the field.
func (o Options) Validate() Errors {
	var es Errors
	if o.SellGear != "" && o.SellGear != "magic" && o.SellGear != "rare" && o.SellGear != "never" {
		es = append(es, Error{Path: "options.sellGear", Msg: "magic, rare or never"})
	}
	if o.MinRune < 0 || o.MinRune > 33 {
		es = append(es, Error{Path: "options.minRune", Msg: "1 to 33"})
	}
	if o.MinGold < 0 || o.MinGold > 100000 {
		es = append(es, Error{Path: "options.minGold", Msg: "0 to 100000"})
	}
	return es
}

func leaf(f, op string, v any) Cond {
	return Cond{F: f, Op: op, V: mustRaw(v)}
}

// Customize returns a copy of base with the options applied. Rules the options name but the base does not have are left
// out silently (a preset may not carry them); everything else is untouched, ids and order included.
func Customize(base *RuleSet, o Options) (*RuleSet, error) {
	if es := o.Validate(); len(es) > 0 {
		return nil, es
	}
	o = o.normalized()
	b, err := base.Marshal()
	if err != nil {
		return nil, err
	}
	rs, err := Parse(b)
	if err != nil {
		return nil, err
	}
	for i := range rs.Rules {
		r := &rs.Rules[i]
		switch r.ID {
		case "g-magic-rare":
			switch o.SellGear {
			case "rare":
				r.When = &Cond{F: "quality", Op: "in", V: mustRaw([]string{"rare", "crafted"})}
				r.Reason = "rare item, sells"
			case "never":
				r.Off = true
			}
		case "g-runes", "t-runes":
			if o.MinRune > 1 {
				w := *r.When
				r.When = &Cond{All: []Cond{w, leaf("runeno", "ge", o.MinRune)}}
				r.Reason = fmt.Sprintf("rune %d or higher", o.MinRune)
			}
		case "g-rw2", "g-rw3", "g-rw4", "t-sockets":
			if o.KeepSockets != nil && !*o.KeepSockets {
				r.Off = true
			}
		case "g-spare-gold":
			if o.MinGold > 0 {
				w := *r.When
				r.When = &Cond{All: []Cond{w, leaf("qty", "ge", o.MinGold)}}
				r.Reason = fmt.Sprintf("gold pile of %d or more", o.MinGold)
			}
		}
	}
	rs.Options = &o
	return rs, nil
}
