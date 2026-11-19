// Package inputdata embeds the input capability sets per backend build (docs/process/sandbox-devp14.md; the planner
// is internal/input).
package inputdata

import "embed"

// FS holds caps/*.json.
//
//go:embed caps/*.json
var FS embed.FS
