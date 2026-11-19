package env

// InvdeskTuning holds the inventory-desk knobs (docs/design/item-verbs.md). Owner: the itemverbs card.
type InvdeskTuning struct {
	Verbs int // 1: do the player's equip through the backend's item verbs when it advertises them (click path otherwise)
	// Instant: 1 lets a verb-capable item action the player gives run at once through the verbs (the fast lane,
	// docs/inventory-actions.md); needs Verbs on too. Default 0 until the live check has run.
	Instant int
	// MaxSets is how many gear sets the player may keep per hero (docs/gearsets.md). Owner: the gearsets card.
	MaxSets int
}

func DefaultInvdeskTuning() InvdeskTuning { return InvdeskTuning{Verbs: 0, Instant: 0, MaxSets: 8} }

func invdeskKnobs() []Knob {
	return []Knob{
		{Key: "invdesk.verbs", Label: "Item verbs", Unit: "bool", Default: 0, Min: 0, Max: 1, NoHandicap: true,
			Desc: "1: a player's equip command goes through the private fork's item verbs (the server's own item handlers) when the connected d2d advertises them in caps; the click path stays the fallback and the oracle. 0 (default): always the click path. Off until the user has seen it work live (docs/design/item-verbs.md)."},
		{Key: "invdesk.instant", Label: "Instant item actions", Unit: "bool", Default: 0, Min: 0, Max: 1, NoHandicap: true,
			Desc: "1: with invdesk.verbs on and a backend that advertises the verbs, an item action the player gives (equip, unequip, drop, identify, and stash, take and gold at the chest) runs at once through the verbs instead of waiting for the bot's next safe moment. 0 (default): every action queues. Off until the user has seen the fast lane work live (docs/inventory-actions.md, Instant actions)."},
		{Key: "invdesk.maxSets", Label: "Gear sets per hero", Unit: "sets", Default: 8, Min: 1, Max: 20, NoHandicap: true,
			Desc: "How many named gear sets the player may save per hero on the inventory page (docs/gearsets.md). A set is applied only when the player clicks Apply; the bot never chooses one."},
	}
}
