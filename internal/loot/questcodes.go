package loot

import (
	"encoding/json"
	"sync"

	questitemsdata "d2lod-bot/data/questitems"
)

var (
	questCodesOnce sync.Once
	questCodes     map[string]bool
)

// QuestCode reports whether a base code is a quest item of Acts 1 to 5: the Act 1 and 2 list is data
// (data/questitems/codes.json), Acts 3 to 5 are QuestItemCodes. It backs the items.txt quest column, never replaces it:
// a quest item is never sold, dropped, vaulted or left on the ground (sandbox hero 2026-11-21: the Scroll of Inifuss
// and the Key to the Cairn Stones were in no never-sell list).
func QuestCode(code string) bool {
	questCodesOnce.Do(func() {
		questCodes = map[string]bool{}
		var f struct{ Act1, Act2 []string }
		if err := json.Unmarshal(questitemsdata.JSON, &f); err != nil {
			panic("data/questitems/codes.json: " + err.Error())
		}
		for _, c := range append(f.Act1, f.Act2...) {
			questCodes[c] = true
		}
		for _, c := range QuestItemCodes {
			questCodes[c] = true
		}
	})
	return questCodes[code]
}
