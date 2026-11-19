// Package questspecdata embeds the quest chains, the backend capability sets and the allow list (docs/questspec.md; the
// checker is internal/questspec). The executor does not read these files.
package questspecdata

import "embed"

// FS holds act1.json .. , caps/*.json and allow.json.
//
//go:embed *.json caps/*.json
var FS embed.FS
