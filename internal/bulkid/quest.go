package bulkid

// Bits of a quest record word (V docs/probes/search-for-cain-l2-2026-11-18.md: record 4, "Search for Cain", went
// 0x0010 -> 0x2012 when the cage opened in the live drill, 0x2011 after the report to Akara).
const (
	questDone          = 0x0001
	questRewardPending = 0x0002 // bit 1: the cage is open, the quest waits for its turn-in
)

// CainRescued reads the Search for Cain quest word: the cage is open (bit 1) or the quest is done (bit 0). Cain's
// position in town is NOT part of this: he leaves the cage by himself and shows up in town later (V l2), so presence
// is a separate check by name or class.
func CainRescued(word uint16) bool { return word&(questDone|questRewardPending) != 0 }
