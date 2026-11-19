// Package capdata embeds the skill capability table (data/capabilities/capabilities.json). The loader and the planner
// that reads it are in internal/agency.
package capdata

import _ "embed"

//go:embed capabilities.json
var JSON []byte
