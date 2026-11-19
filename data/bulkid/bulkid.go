// Package bulkiddata embeds the bulk identify filter table (data/bulkid/filter.json). The engine is in internal/bulkid.
package bulkiddata

import _ "embed"

// Filter is filter.json.
//
//go:embed filter.json
var Filter []byte
