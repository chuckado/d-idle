package d2

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// ClassTraits are the per-class facts several packages used to write out as a class list (build flexibility audit rows 27
// to 29). Caster: every attack costs mana and life is thin (gear gate, den floor, merc auras). ManaNeed ranks the mana
// dependence: 0 none, 1 occasional, 2 every attack (docs/classes/*.md section 7, S). The numbers live in classtraits.json.
type ClassTraits struct {
	Caster   bool `json:"caster"`
	ManaNeed int  `json:"manaNeed"`
}

//go:embed classtraits.json
var classTraitsJSON []byte

var classTraits = func() map[string]ClassTraits {
	m := map[string]ClassTraits{}
	if err := json.Unmarshal(classTraitsJSON, &m); err != nil {
		panic("d2: classtraits.json: " + err.Error())
	}
	for _, n := range classNames {
		if _, ok := m[n]; !ok {
			panic("d2: classtraits.json has no row for " + n)
		}
	}
	return m
}()

// TraitsOf returns the row of a class; an unknown class is an error naming it, never another class's row.
func TraitsOf(c Class) (ClassTraits, error) {
	t, ok := classTraits[c.String()]
	if !ok {
		return ClassTraits{}, fmt.Errorf("d2: no class traits for %s", c)
	}
	return t, nil
}

// TraitsByName is TraitsOf for the class name strings some packages carry ("Sorceress").
func TraitsByName(name string) (ClassTraits, error) {
	t, ok := classTraits[name]
	if !ok {
		return ClassTraits{}, fmt.Errorf("d2: no class traits for %q", name)
	}
	return t, nil
}

// Caster reports the caster trait; false for an unknown class (a corrupt class byte is not a caster, and TraitsOf names it).
func (c Class) Caster() bool { t, _ := TraitsOf(c); return t.Caster }

// ManaNeed is the mana dependence rank; an unknown class reads 1 (occasional), the old default for classes without a row.
func (c Class) ManaNeed() int {
	if t, err := TraitsOf(c); err == nil {
		return t.ManaNeed
	}
	return 1
}
