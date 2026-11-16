package merc

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"d2lod-bot/internal/data"
)

// Which hireling types can wear an item (docs/lootrules.md section 1, field MercTypes). The table wear.json is the same
// rule as Allowed (V from the d2d source, docs/research/merc-gear.md section 2.3: Rogue bow, Desert Guard spear and
// polearm, Iron Wolf one-handed sword and shield, Barbarian one-handed sword; helm and body armor for all; no belt,
// gloves, boots or jewelry; javelins, axes, two-handed swords and Druid pelts are U and refused) written as data keyed
// by the item type tree, so a loot rule can ask without a data.DB. WTypeGaps cross-checks the weapon types against the
// Hireling.txt WType columns when the real tables are loaded (a real-table run is not part of the offline tests).

//go:embed wear.json
var wearJSON []byte

type slotWear struct {
	Types     []string `json:"types"`               // the item type or one of its ancestors is one of these
	Not       []string `json:"not,omitempty"`       // ... and none of these
	OneHanded bool     `json:"oneHanded,omitempty"` // weapons: two-handed ones are refused
}

var (
	wearOnce  sync.Once
	wearTable map[string]map[string]slotWear
	wearErr   error
)

func wearTab() (map[string]map[string]slotWear, error) {
	wearOnce.Do(func() { wearErr = json.Unmarshal(wearJSON, &wearTable) })
	return wearTable, wearErr
}

// WearKinds lists the hireling kind names of the table ("rogue", "desert", "ironwolf", "barbarian"), sorted.
func WearKinds() []string {
	t, err := wearTab()
	if err != nil {
		panic("merc: bad embedded wear.json: " + err.Error())
	}
	out := make([]string, 0, len(t))
	for k := range t {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func hasAny(have, want []string) bool {
	for _, w := range want {
		for _, h := range have {
			if h == w {
				return true
			}
		}
	}
	return false
}

// CanWear reports whether a hireling of the kind can wear an item whose type chain (its type and every ancestor, as
// data.DB.IsA walks them) is types. twoHanded is the base row's flag. An unknown kind is an error naming it.
func CanWear(kind string, types []string, twoHanded bool) (bool, error) {
	t, err := wearTab()
	if err != nil {
		return false, err
	}
	slots, ok := t[kind]
	if !ok {
		return false, fmt.Errorf("merc: unknown hireling kind %q (known: %v)", kind, WearKinds())
	}
	for _, s := range slots {
		if hasAny(types, s.Types) && !hasAny(types, s.Not) && !(s.OneHanded && twoHanded) {
			return true, nil
		}
	}
	return false, nil
}

// WearTypes lists, sorted, the hireling kinds that can wear the item; nil when none can.
func WearTypes(types []string, twoHanded bool) []string {
	var out []string
	for _, k := range WearKinds() {
		if ok, _ := CanWear(k, types, twoHanded); ok {
			out = append(out, k)
		}
	}
	return out
}

// TypeChain is the item type and every ancestor through Equiv1 and Equiv2, the list CanWear and the loot rules read.
func TypeChain(db *data.DB, typ string) []string {
	var out []string
	seen := map[string]bool{}
	var walk func(string)
	walk = func(c string) {
		if c == "" || seen[c] {
			return
		}
		seen[c] = true
		out = append(out, c)
		if t := db.ItemType(c); t != nil {
			walk(t.Equiv[0])
			walk(t.Equiv[1])
		}
	}
	walk(typ)
	return out
}

// WTypeGaps lists the WType codes of a Hireling.txt row that no weapon slot of the table covers (an exact code match),
// so a table that disagrees with the game data shows up in a real-table test instead of in play. An unknown hireling
// class is an error.
func WTypeGaps(h *data.Hireling) ([]string, error) {
	kind := KindName(h.Class)
	if kind == "" {
		return nil, fmt.Errorf("merc: hireling %q has unknown class %d", h.Name, h.Class)
	}
	t, err := wearTab()
	if err != nil {
		return nil, err
	}
	var gaps []string
	for _, w := range h.WType {
		if w != "" && !hasAny([]string{w}, t[kind]["weapon"].Types) {
			gaps = append(gaps, w)
		}
	}
	return gaps, nil
}
