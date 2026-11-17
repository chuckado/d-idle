// Package vaultdata embeds the Ethereal Vault tables (data/vault/*.json). The engine is in internal/vault.
package vaultdata

import _ "embed"

// Capacity is capacity.json (tier table).
//
//go:embed capacity.json
var Capacity []byte

// Value is value.json (trim value function).
//
//go:embed value.json
var Value []byte
