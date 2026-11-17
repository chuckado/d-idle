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
	return QuestsTuning{AutoDelaySec: 20, AwayMin: 10, AutoHoldMin: 15, AutoGapMin: 30, AutoDayCap: 6}
}

func questsKnobs() []Knob {
	return []Knob{
		{Key: "quests.board", Label: "Quest board drives quests", Unit: "bool", Default: 0, Min: 0, Max: 1, NoHandicap: true,
			Desc: "1: a quest runs only after the player accepts its current chunk (Start); the bot does that chunk, walks to the NPC and stops, and the player taps Hand in. The planner then no longer starts quests by itself: between chunks the hero acts on the player's mode (Train levels toward the next quest, Idle parks in town, Farm farms loot and gold; docs/questboard.md Modes). 0 (default): today's planner behaviour, nothing changes. Off until the live drill of docs/questboard.md has run."},
	}
}
