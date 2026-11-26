package bot

import (
	"testing"

	"d2lod-bot/internal/admit"
)

// internal/admit does not import the bot, so it repeats the need codes and wakes; this keeps them equal to needTable.
func TestAdmitNeedsMatchNeedTable(t *testing.T) {
	for code, wake := range admit.Wake {
		row, ok := needTable[code]
		if !ok {
			t.Errorf("admit need %q is not a row of needTable", code)
			continue
		}
		if row.Wake != wake {
			t.Errorf("admit wake for %q = %q, needTable has %q", code, wake, row.Wake)
		}
	}
}
