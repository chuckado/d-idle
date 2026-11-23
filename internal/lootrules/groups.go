package lootrules

import "strings"

// builtinGroups are the group names every rule set may use. Item type codes are S (the repo's own tables), U against the
// live ItemBase.Type strings. A rule set may add groups, not redefine a built-in one.
var builtinGroups = map[string]Group{
	"runes":    {Types: []string{"rune"}},
	"gems":     {TypePrefixes: []string{"gem", "skul"}},
	"charms":   {Types: []string{"scha", "mcha", "lcha"}},
	"jewels":   {Types: []string{"jewl"}},
	"jewelry":  {Types: []string{"ring", "amul"}},
	"circlets": {Types: []string{"circ"}},
	"potions":  {Types: []string{"hpot", "mpot", "rpot", "spot", "apot", "wpot"}},
	"supplies": {Codes: []string{"tbk", "ibk", "tsc", "isc", "key"}},
	"gold":     {Types: []string{"gold"}},
	"gear":     {Types: []string{"weap", "armo"}},
	"javelins": {Types: []string{"jave", "ajav"}},
}

// protectedGroups may never be dropped or sold by a rule without the player's confirmation (and, for supplies, never).
var protectedGroups = []string{"runes", "gems", "charms", "jewels"}

func (g Group) has(it *Item) bool {
	for _, t := range g.Types {
		for _, x := range it.Types {
			if x == t {
				return true
			}
		}
	}
	for _, p := range g.TypePrefixes {
		for _, x := range it.Types {
			if strings.HasPrefix(x, p) {
				return true
			}
		}
	}
	for _, c := range g.Codes {
		if globMatch(c, it.Code) {
			return true
		}
	}
	return false
}

// globMatch supports ? and * only (base codes are 3 to 4 characters).
func globMatch(p, s string) bool {
	if p == "" {
		return s == ""
	}
	switch p[0] {
	case '*':
		for i := 0; i <= len(s); i++ {
			if globMatch(p[1:], s[i:]) {
				return true
			}
		}
		return false
	case '?':
		return s != "" && globMatch(p[1:], s[1:])
	}
	return s != "" && s[0] == p[0] && globMatch(p[1:], s[1:])
}
