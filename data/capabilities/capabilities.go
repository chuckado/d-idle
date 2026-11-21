// Package capdata embeds the skill capability table (data/capabilities/capabilities.json). The loader and the planner
// that reads it are in internal/agency.
package capdata

import _ "embed"

//go:embed capabilities.json
var JSON []byte

// Proven lists the skills with a green Proving Ground row per backend; a tier of supported needs one (internal/delivery).
//
//go:embed proven.json
var Proven []byte
