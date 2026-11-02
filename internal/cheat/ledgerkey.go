package cheat

import "strings"

// KindSetupCopy records a file-level copy of a saved character into a prefix (cmd/d2restore place). It is a
// declared SETUP action (docs/adr/0014): it edits no gameplay value, so it never makes a character cheated,
// but placed after a run-start marker it still counts as a write of that run.
const KindSetupCopy Kind = "setup-copy"

// IsMarker reports kinds that are ledger bookkeeping, not cheats: the setup copy and the run-start marker
// (the literal is the memrules marker kind, run.go).
func (k Kind) IsMarker() bool { return k == KindSetupCopy || k == "runstart" }

// Key identifies a character in one prefix: CharID when an entry carries one, else "<instance>/<name>".
// Five copies of one save in five prefixes (or renamed) are five identities. Entries written before the
// key existed have an instance and a name, so they key the same way; ones without an instance key by name.
func (e Entry) Key() string {
	switch {
	case e.CharID != "":
		return e.CharID
	case e.Instance != "":
		return e.Instance + "/" + e.Character
	}
	return e.Character
}

// CharKey builds the Filter.Key of a character in an instance's prefix.
func CharKey(instance, name string) string { return instance + "/" + name }

// matchKey: case-insensitive like Filter.Character. An entry with no instance cannot be attributed to a
// prefix, so it matches by name (fail closed: it keeps counting against every same-named character).
func matchKey(e Entry, key string) bool {
	if strings.EqualFold(e.Key(), key) {
		return true
	}
	if e.CharID == "" && e.Instance == "" {
		_, name, ok := strings.Cut(key, "/")
		return ok && strings.EqualFold(e.Character, name)
	}
	return false
}
