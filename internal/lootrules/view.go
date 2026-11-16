package lootrules

// ViewSchema is the version of the GET /api/lootrules answer.
const ViewSchema = "watch.lootrules/1"

// Recent is one decision of the live bot, newest first in the view, for the "why was this kept or sold" list.
type Recent struct {
	Scope string `json:"scope"` // ground, tidy or sell: the decision point
	Item  string `json:"item"`
	Today string `json:"today"` // what the bot's own code did
	Rules string `json:"rules"` // what the rule set says
	Rule  string `json:"rule"`  // id of the rule that decided, "default" or "invariant:<name>"
	Why   string `json:"why"`
}

// ViewDoc is the answer of GET /api/lootrules.
type ViewDoc struct {
	Schema    string            `json:"schema"`
	Seq       int               `json:"seq"`
	Active    *RuleSet          `json:"active"`
	Plain     map[string]string `json:"plain"` // rule id -> one sentence
	Protected []InvariantInfo   `json:"protected"`
	Presets   []PresetInfo      `json:"presets"`
	Classes   []string          `json:"classes"`
	Options   Options           `json:"options"` // the effective options (defaults filled in)
	Shadowed  []Shadow          `json:"shadowed"`
	Recent    []Recent          `json:"recent"`
	Stage     int               `json:"stage"` // knob loot.rulesEngine now
	LoadError string            `json:"loadError,omitempty"`
	SellKinds []SellKind        `json:"sellKinds"`      // the five kinds the sell stage decides, in plain words
	CanUndo   bool              `json:"canUndo"`        // an earlier saved version exists
	Hero      string            `json:"hero,omitempty"` // the hero class the per-hero store last saw
}

// View assembles the answer from the store, the live stage and the recent decisions.
func (s *Store) View(stage int, recent []Recent) ViewDoc {
	rs := s.Active()
	pl := map[string]string{}
	for _, r := range rs.Rules {
		pl[r.ID] = Describe(r)
	}
	o := Options{}
	if rs.Options != nil {
		o = *rs.Options
	}
	cl := append([]string(nil), s.classes...)
	if recent == nil {
		recent = []Recent{}
	}
	return ViewDoc{Schema: ViewSchema, Seq: rs.Seq, Active: rs, Plain: pl, Protected: Invariants(), Presets: s.Presets(),
		Classes: cl, Options: o.normalized(), Shadowed: rs.Shadowed(), Recent: recent, Stage: stage, LoadError: s.LoadErr, CanUndo: s.HasUndo(), SellKinds: SellKinds()}
}

// SellKind is one of the five things the sell stage decides, as the user settled them on 2026-11-16 (docs/lootrules.md
// section 13). The sheet shows them in plain words.
type SellKind struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

// SellKinds lists the five sell decisions.
func SellKinds() []SellKind {
	return []SellKind{
		{"good-gear", "Good gear it picked up on purpose", "Kept. Selling something the bot walked to and took as good a minute ago would be a contradiction."},
		{"cheap-junk", "Cheap junk under 20 gold with no rule", "Sold at the next vendor visit once the pack is getting full. Every cell it holds is a cell a better item cannot use."},
		{"low-gems", "Surplus low gems", "Sold. Only the gems your build or the Horadric Cube still needs go to the stash."},
		{"plain-junk", "Plain junk", "Sold in town. It is only dropped on the ground when the pack is full away from town."},
		{"jewelry", "Rings, amulets and circlets", "Sold unidentified until you own a Tome of Identify, then identified and stashed."},
	}
}
