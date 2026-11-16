package lootrules

import _ "embed"

//go:embed starter.json
var starterJSON []byte

// Starter returns the built-in Starter preset: today's pickit (ground) and tidy policy (pack) as one rule set.
func Starter() *RuleSet {
	rs, err := Parse(starterJSON)
	if err != nil {
		panic("lootrules: bad embedded starter: " + err.Error())
	}
	return rs
}
