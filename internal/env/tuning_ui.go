package env

// UITuning holds the viewer's management switches (docs/process/management-fast-path.md).
type UITuning struct {
	// ManagementFast: 1 lets the player's management actions (equip, unequip, drop, identify, stash, gold) run at once
	// through the item verbs and opens the Recall now command. It stands for invdesk.verbs, invdesk.instant and
	// input.planner together for the desk's lane and executors. Default 0.
	ManagementFast int
}

func DefaultUITuning() UITuning { return UITuning{} }

func uiKnobs() []Knob {
	return []Knob{
		{Key: "ui.managementFast", Label: "Immediate management", Unit: "bool", Default: 0, Min: 0, Max: 1, NoHandicap: true,
			Desc: "1: the viewer's management actions (equip, unequip, drop, identify, stash and take, gold) run at once through the backend's item verbs, not behind the bot's goal loop, and Recall now (POST /api/recall) takes the hero to town at the next safe moment. Needs a backend that advertises the item verbs; without them, in a named run, or when the lane refuses, the action queues as before and the page names the reason. Driving orders (farm, quest, pause) stay queued. 0 (default): everything queues. Off until the user has seen it work live (docs/process/management-fast-path.md)."},
	}
}
