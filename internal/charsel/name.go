package charsel

import (
	"fmt"
	"strings"
)

// ValidName applies the Diablo II character name rule: 2 to 15 letters, at most one hyphen or underscore, never first or
// last (S: the game's own rule). Duplicates are checked by the Store, because they need the folder.
func ValidName(n string) error {
	if len(n) < NameMin || len(n) > NameMax {
		return fmt.Errorf("a character name has %d to %d characters", NameMin, NameMax)
	}
	sep := 0
	for i, r := range n {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		case r == '-' || r == '_':
			if sep++; sep > 1 || i == 0 || i == len(n)-1 {
				return fmt.Errorf("a name may hold one hyphen or underscore, not at its start or end")
			}
		default:
			return fmt.Errorf("a character name holds letters, and one hyphen or underscore at most")
		}
	}
	return nil
}

// sameName compares names the way the file system of a friend's PC does (case-insensitive).
func sameName(a, b string) bool { return strings.EqualFold(a, b) }
