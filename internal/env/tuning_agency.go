package env

// AgencyTuning holds the knobs of the stat and skill agency work (docs/agency.md, ADR 0020). Owner: the agencyp1a card.
type AgencyTuning struct {
	// Planner is the skill planner: 0 off, 1 shadow (logs what it would pick next to what the combat profile picks, serves
	// GET /api/hero/plan; the hero plays exactly as before).
	Planner int
}

func DefaultAgencyTuning() AgencyTuning { return AgencyTuning{Planner: 0} }

func agencyKnobs() []Knob {
	return []Knob{
		{Key: "agency.planner", Label: "Skill planner", Unit: "mode", Default: 0, Min: 0, Max: 1, NoHandicap: true,
			Desc: "0 (default): off. 1: shadow. Once a minute the hero logs a `planner shadow:` line with the attack and aura the planner would pick from the skills the player actually bought, next to what the combat profile picks, and GET /api/hero/plan and the Points page show the derived rotation, the warnings and the viability line. Shadow only: it never changes what the hero does (docs/agency.md)."},
	}
}
