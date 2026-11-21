// Package questitemsdata embeds the base codes of the quest items of Acts 1 and 2 that the loot rules and the town
// must never sell, drop or leave on the ground (internal/loot QuestCode). Acts 3 to 5 are loot.QuestItemCodes.
package questitemsdata

import _ "embed"

// JSON is codes.json.
//
//go:embed codes.json
var JSON []byte
