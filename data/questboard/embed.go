// Package questboarddata embeds the quest board definitions (docs/questboard.md; the code is internal/questboard).
package questboarddata

import _ "embed"

// Families is families.json (class to family).
//
//go:embed families.json
var Families []byte

// Act1 is act1.json (quests.def/1).
//
//go:embed act1.json
var Act1 []byte

// Act2 is act2.json (quests.def/1).
//
//go:embed act2.json
var Act2 []byte
