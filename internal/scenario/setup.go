package scenario

import (
	"fmt"
	"sort"

	"d2lod-bot/internal/telemetry"
)

// Verb is one lab verb of the declared SETUP with the reason written to the ledger before it is sent.
type Verb struct {
	Line   string `json:"line"`
	Reason string `json:"reason"`
}

// Setup is what the builder produces from a Pre: lab verbs for the prep d2d (lab verbs on), the board record for the
// hero's plan store, and the provenance tags. It is deterministic: the same scenario, class and seed give the same Setup
// (no clock, no random, no checked-in save blob).
type Setup struct {
	Verbs []Verb `json:"verbs"`
	// Board is the chunk the player accepted; the runner writes it with the real questboard store (<hero>.quests.json).
	Board *QuestRef `json:"board,omitempty"`
	// Tuning are knob overrides for the bot run (state/tuning.json), after the scenario's own Trigger.Knobs.
	Tuning map[string]any `json:"tuning,omitempty"`
	// Build is trigger.build: the runner writes it as levelerBuild into the bot's config.json.
	Build string `json:"build,omitempty"`
	// Tags classify the setup for docs/probes/test-characters.md: CHEATED when the state could not be reached by play
	// (a warped position, a given quest item, a hand-set quest word), SETUP-only otherwise.
	Tags []string `json:"tags"`
	// LifePct is the life percentage the runner brings the hero to AFTER the verbs, with `debug hit <dmg>` (V docs/forkcheats.md:
	// the real hurt path), the damage computed from the live life: 0 means leave it alone. It is a gameplay write before the
	// start, a declared SETUP.
	LifePct int `json:"life_pct,omitempty"`
	// Field is true when the hero is warped and the live d2d is handed to the bot (pre.start field).
	Field bool `json:"field"`
	// Unsupported lists declared preconditions the builder cannot construct yet (no verb known). A run with any is BLOCKED.
	Unsupported []string `json:"unsupported,omitempty"`
}

// BuildSetup turns a scenario's Pre into the setup for one class and seed. The seed is only used to name the hero so
// two runs never share a save directory; the map seed is passed to d2d by the runner.
func BuildSetup(s Scenario, class string, seed int64) (Setup, error) {
	if _, err := ClassIndex(class); err != nil {
		return Setup{}, err
	}
	p := s.Pre
	reason := func(what string) string {
		return fmt.Sprintf("SETUP: scenario %s (%s, seed %d): %s", s.ID, class, seed, what)
	}
	var su Setup
	add := func(what, line string) { su.Verbs = append(su.Verbs, Verb{Line: line, Reason: reason(what)}) }
	tags := map[string]bool{}

	if p.Level > 1 {
		exp, ok := telemetry.DefaultExpForLevel(p.Level)
		if !ok {
			return Setup{}, fmt.Errorf("scenario %s: no experience for level %d", s.ID, p.Level)
		}
		add("level", fmt.Sprintf("debug stat 12 %d", p.Level))
		add("experience of that level", fmt.Sprintf("debug stat 13 %d", exp))
		add("stat points a played level gives", fmt.Sprintf("debug statpts %d", 5*(p.Level-1)))
		add("skill points a played level gives", fmt.Sprintf("debug skillpts %d", p.Level-1))
	}
	if p.Gold > 0 {
		add("purse", fmt.Sprintf("debug stat 14 %d", p.Gold))
	}
	if p.Stash > 0 {
		add("stash gold", fmt.Sprintf("debug stat 15 %d", p.Stash))
	}
	switch p.Difficulty {
	case "normal", "":
	case "nightmare":
		add("difficulty", "debug difficulty 1") // U: argument form of the fork verb
		tags["difficulty"] = true
	case "hell":
		add("difficulty", "debug difficulty 2")
		tags["difficulty"] = true
	default:
		return Setup{}, fmt.Errorf("scenario %s: difficulty %q", s.ID, p.Difficulty)
	}
	for _, w := range p.Waypoints {
		add(fmt.Sprintf("waypoint bit %d", w), fmt.Sprintf("debug wp %d", w))
		tags["waypoint"] = true
	}
	for _, q := range p.Quests {
		add(fmt.Sprintf("quest record %d", q.Record), fmt.Sprintf("debug quest set %d 0x%04x", q.Record, q.Word))
		tags["quest-word"] = true
	}
	var floor [][2]string // ground items, dropped after the position verb: a drop before it lands where the hero stood (the town)
	for _, it := range p.Items {
		n := it.Count
		if n <= 0 {
			n = 1
		}
		flags := ""
		if it.Quality > 0 {
			flags += fmt.Sprintf(" q%d", it.Quality)
		}
		if it.Unid {
			flags += " unid"
		}
		switch it.Where {
		case "", "pack":
			for i := 0; i < n; i++ {
				add("item "+it.Code, fmt.Sprintf("debug give %s pack%s", it.Code, flags))
			}
		case "ground": // the verb without a place drops the item beside the hero and takes the quality as a bare number (V docs/forkcheats.md)
			if it.Quality == 0 {
				su.Unsupported = append(su.Unsupported, fmt.Sprintf("item %s on the ground needs a quality", it.Code))
				break
			}
			for i := 0; i < n; i++ {
				line := fmt.Sprintf("debug give %s %d", it.Code, it.Quality)
				if it.Unid {
					line += " unid"
				}
				floor = append(floor, [2]string{"item " + it.Code + " on the ground", line})
			}
		default:
			su.Unsupported = append(su.Unsupported, fmt.Sprintf("item %s into %s (no setup verb known)", it.Code, it.Where))
		}
		tags["given-item"] = true
	}
	if len(p.Belt) > 0 {
		add("empty the belt", "debug belt clear")
		for i, code := range p.Belt {
			add("belt box "+fmt.Sprint(i), fmt.Sprintf("debug give %s belt %d", code, i))
		}
	}
	if p.Merc != nil {
		add("hired merc", fmt.Sprintf("debug merc hire %d", *p.Merc))
		tags["merc"] = true
	}
	if p.LifePct > 0 && p.LifePct < 100 {
		su.LifePct = p.LifePct
	}
	if p.Start == StartField {
		add("position (the save cannot hold a field position)", fmt.Sprintf("debug goto %d", p.Area))
		su.Field = true
		tags["position"] = true
	}
	for _, f := range floor {
		add(f[0], f[1])
	}
	su.Board = p.Board
	su.Build = s.Trigger.Build
	if p.Board != nil {
		su.Tuning = map[string]any{"quests.board": 1}
	}
	for k, v := range s.Trigger.Knobs {
		if su.Tuning == nil {
			su.Tuning = map[string]any{}
		}
		su.Tuning[k] = v
	}
	cheated := tags["position"] || tags["waypoint"] || tags["given-item"] || tags["quest-word"] || tags["merc"] || tags["difficulty"]
	if cheated {
		su.Tags = append(su.Tags, "CHEATED")
	} else {
		su.Tags = append(su.Tags, "SETUP-only")
	}
	var names []string
	for t := range tags {
		names = append(names, t)
	}
	sort.Strings(names)
	su.Tags = append(su.Tags, names...)
	return su, nil
}

// HeroName is the throwaway character name of one run: stable per scenario, class and seed, valid for the board store
// (2 to 24 letters, digits, '_' and '-', starting with a letter) and distinct from any played hero.
func HeroName(scenarioID, class string, seed int64) string {
	var b []byte
	b = append(b, 'P', 'g')
	for _, r := range scenarioID {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b = append(b, byte(r))
		}
		if len(b) >= 10 {
			break
		}
	}
	b = append(b, class[0]|0x20)
	n := fmt.Sprint(seed)
	if len(n) > 5 {
		n = n[len(n)-5:]
	}
	return string(b) + n
}
