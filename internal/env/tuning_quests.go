package env

// QuestsTuning holds the quest board switch (docs/questboard.md, design docs/design/quest-board-2026-11-16.md).
type QuestsTuning struct {
	Board int // 1: the quest board drives quests (player taps Start and Hand in); 0: the planner drives them as before
}

func DefaultQuestsTuning() QuestsTuning { return QuestsTuning{} }

func questsKnobs() []Knob {
	return []Knob{
		{Key: "quests.board", Label: "Quest board drives quests", Unit: "bool", Default: 0, Min: 0, Max: 1, NoHandicap: true,
			Desc: "1: a quest runs only after the player accepts its current chunk (Start); the bot does that chunk, walks to the NPC and stops, and the player taps Hand in. The planner then no longer starts quests by itself: between chunks the hero acts on the player's mode (Train levels toward the next quest, Idle parks in town, Farm farms loot and gold; docs/questboard.md Modes). 0 (default): today's planner behaviour, nothing changes. Off until the live drill of docs/questboard.md has run."},
	}
}
