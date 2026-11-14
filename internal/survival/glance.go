package survival

// Glance is the survival part of /api/glance (design 8.1), behind the UI's feature flag. WORDS ONLY: the margin is
// "safe", "risky" or "too hard" and no field here carries a margin, xp or gold number (user decision 2026-11-14,
// decisions-2026-11-14.md item 2). The numbers stay in the `zone choice:` log line and in /api/danger for developers.
type Glance struct {
	Schema       string     `json:"schema"`
	Zone         GlanceZone `json:"zone"`
	Alternatives []Alt      `json:"alternatives"`
	Merc         *Why       `json:"merc,omitempty"` // why he stays dead, for the Status sentence
	Fresh        bool       `json:"fresh,omitempty"`
	NeedsYou     bool       `json:"needsYou,omitempty"`
}

// GlanceZone is the zone the shadow controller would farm.
type GlanceZone struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Margin string `json:"margin"`
	Dial   string `json:"dial"`
	Why    Why    `json:"why"`
}

// GlanceSchema is the schema id of the block.
const GlanceSchema = "survival.glance/1"

// GlanceOf builds the words-only block from a decision; merc is the revive verdict when the merc is dead (nil otherwise).
func GlanceOf(d Decision, dial string, merc *ReviveDecision) *Glance {
	g := &Glance{Schema: GlanceSchema, Zone: GlanceZone{ID: d.Area, Name: d.Name, Margin: d.Margin, Dial: dial, Why: d.Why},
		Alternatives: append([]Alt{}, d.Alts...), Fresh: d.Fresh, NeedsYou: d.NeedsYou}
	if merc != nil && !merc.Revive {
		w := merc.Why
		g.Merc = &w
	}
	return g
}
