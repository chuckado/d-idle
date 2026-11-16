package charsel

// What the Characters page needs to validate a name and offer a class without a literal of its own
// (GET /api/characters/rules, watch.characters.rules/1). The numbers are the ones ValidName applies.
const (
	NameMin       = 2
	NameMax       = 15
	NameSeparator = "-_"
	RulesSchema   = "watch.characters.rules/1"
)

// HardcoreBuilt is false until the T4 acceptance (docs/decisions.md Current truth); Create refuses hardcore while it is.
const HardcoreBuilt = false

// ClassView is one class chip.
type ClassView struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NameRule is the name rule in numbers; the page words it.
type NameRule struct {
	Min        int    `json:"min"`
	Max        int    `json:"max"`
	Separators string `json:"separators"` // at most one of these, never first or last
	FoldCase   bool   `json:"foldCase"`   // "Zapp" and "zapp" are the same name
}

// RulesView is the reply of the rules route.
type RulesView struct {
	Schema   string      `json:"schema"`
	Classes  []ClassView `json:"classes"`
	Name     NameRule    `json:"name"`
	Hardcore bool        `json:"hardcore"`
}

// Rules returns the page's rules from the class table and the name constants.
func Rules() RulesView {
	v := RulesView{Schema: RulesSchema, Name: NameRule{Min: NameMin, Max: NameMax, Separators: NameSeparator, FoldCase: true}, Hardcore: HardcoreBuilt}
	for _, c := range classes {
		v.Classes = append(v.Classes, ClassView{ID: c.ID, Name: c.Name})
	}
	return v
}
