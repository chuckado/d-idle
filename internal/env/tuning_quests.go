package env

// QuestsTuning holds the quest board switch (docs/questboard.md, design docs/design/quest-board-2026-11-16.md).
type QuestsTuning struct {
	Board int // 1: the quest board drives quests (player taps Start and Hand in); 0: the planner drives them as before

	// Opt-in auto-start (slice S8); nothing here acts while Board is 0 or while no quest is marked auto.
	AutoDelaySec int // countdown before an auto-start, in seconds; 0 starts at once
	AwayMin      int // minutes without a watch client before the player counts as away
	AutoHoldMin  int // minutes after the player paused or reset when nothing starts by itself
	AutoGapMin   int // minutes between two auto-started chunks
	AutoDayCap   int // auto-started chunks per 24 hours
}

func DefaultQuestsTuning() QuestsTuning {
	return QuestsTuning{Board: 1, AutoDelaySec: 20, AwayMin: 10, AutoHoldMin: 15, AutoGapMin: 30, AutoDayCap: 6}
}

func questsKnobs() []Knob {
	return []Knob{
		{Key: "quests.board", Label: "Quest board drives quests", Unit: "bool", Default: 1, Min: 0, Max: 1, NoHandicap: true,
			Desc: "1: a quest runs only after the player accepts its current chunk (Start); the bot does that chunk, walks to the NPC and stops, and the player taps Hand in. The planner then no longer starts quests by itself: between chunks the hero acts on the player's mode (Train levels toward the next quest, Idle parks in town, Farm farms loot and gold; docs/questboard.md Modes). 0 (default): today's planner behaviour, nothing changes. Off until the live drill of docs/questboard.md has run."},
		{Key: "quests.autoDelaySec", Label: "Auto-start countdown", Unit: "s", Default: 20, Min: 0, Max: 300, NoHandicap: true,
			Desc: "Seconds an opted-in quest counts down (shown on the page, with Cancel) before it starts by itself; 0 starts at once. Only a SAFE offer ever auto-starts."},
		{Key: "quests.awayMin", Label: "Away after", Unit: "min", Default: 10, Min: 1, Max: 240, NoHandicap: true,
			Desc: "Minutes without any watch client asking for the board before the player counts as away: then a chunk the board started by itself is also handed in by the board. A present player taps Hand in."},
		{Key: "quests.autoHoldMin", Label: "No auto-start after pause or reset", Unit: "min", Default: 15, Min: 0, Max: 240, NoHandicap: true,
			Desc: "No quest starts by itself this many minutes after the player paused or reset the bot."},
		{Key: "quests.autoGapMin", Label: "Minutes between auto-starts", Unit: "min", Default: 30, Min: 0, Max: 600, NoHandicap: true,
			Desc: "At least this long between two auto-started chunks (a bug can then not chain starts)."},
		{Key: "quests.autoDayCap", Label: "Auto-starts per day", Unit: "count", Default: 6, Min: 0, Max: 100, NoHandicap: true,
			Desc: "At most this many auto-started chunks in 24 hours; 0 turns auto-start off."},
	}
}
