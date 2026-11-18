package questboard

import "fmt"

// The between-chunks modes (slice S6, design 3.4): what the quest the hero would train toward is, and the readiness
// meter in words. Pure: nothing here reads the game or changes a zone choice.

// Target is the quest the Train mode works toward.
type Target struct {
	Quest, Chunk, Title string
	Word                Word // readiness of the chunk now (ReadyUnknown: no model)
	Offered             bool // the chunk is Ready: the offer is up and only the player's Start is missing
	Floor               int  // legacy clvl floor of the hero's class family; 0 = none
}

// NextTarget is the lowest-order quest that is not done, skipped or running and whose prerequisites are met (state
// Available or Ready): story quests first, then optional ones, then unlocks (design 3.4). The chunk is the first one
// that is itself Available or Ready. ok is false when nothing is left to train for.
func NextTarget(b *Board, views []QuestView, class string) (Target, bool) {
	for _, kind := range []string{"story", "optional", ""} {
		for _, v := range views {
			q, err := b.Quest(v.ID)
			if err != nil || (kind != "" && q.Kind != kind) || (kind == "" && (q.Kind == "story" || q.Kind == "optional")) {
				continue
			}
			if v.State != Available && v.State != Ready {
				continue
			}
			for _, c := range v.Chunks {
				if c.State != Available && c.State != Ready {
					continue
				}
				t := Target{Quest: v.ID, Chunk: c.ID, Title: q.Title, Word: c.Reason, Offered: c.State == Ready}
				if ch, err := b.Chunk(v.ID, c.ID); err == nil && len(q.Chunks) > 1 {
					t.Title = q.Title + ": " + ch.Title
				}
				if n, ok, err := b.Floor(q, class); err == nil && ok {
					t.Floor = n
				}
				return t, true
			}
		}
	}
	return Target{}, false
}

// Meter is the mode block of GET /api/quests: the mode, what the hero does about it and a readiness meter in words.
type Meter struct {
	Mode       string `json:"mode"`              // the mode the player picked
	Applied    string `json:"applied,omitempty"` // the mode the hero acts on now; differs from Mode until the next safe moment
	Line       string `json:"line"`              // "Training toward Blood Raven: 2 levels to go."
	Target     string `json:"target,omitempty"`  // quest id (train only)
	TargetName string `json:"targetName,omitempty"`
	Word       Word   `json:"word,omitempty"`
	LevelsToGo int    `json:"levelsToGo,omitempty"` // by the legacy clvl floor; 0 when none or reached
	Offered    bool   `json:"offered,omitempty"`    // the target is ready: the player decides
}

// BuildMeter words the mode. running is the title of the accepted chunk ("" when none): a running chunk comes first and
// the mode returns after it. level is the hero's clvl.
func BuildMeter(mode, applied string, t Target, hasTarget bool, level int, running string) Meter {
	m := Meter{Mode: mode, Applied: applied}
	switch {
	case running != "":
		m.Line = "On a quest: " + running + ". " + modeAfter(mode)
	case mode == "idle":
		m.Line = "Waiting in town. Nothing happens in the field until you start a quest or change the mode."
	case mode == "farm":
		m.Line = "Farming loot and gold in a zone the hero has outgrown, not toward the next quest."
	case !hasTarget:
		m.Line = "Training: no quest is left to prepare for here."
	default:
		m.Target, m.TargetName, m.Word, m.Offered = t.Quest, t.Title, t.Word, t.Offered
		if t.Offered {
			m.Line = fmt.Sprintf("Training. %s is ready; tap Start when you want it.", t.Title)
			break
		}
		if t.Floor > level {
			m.LevelsToGo = t.Floor - level
			m.Line = fmt.Sprintf("Training toward %s: %d %s to go.", t.Title, m.LevelsToGo, plural(m.LevelsToGo, "level", "levels"))
			break
		}
		m.Line = fmt.Sprintf("Training toward %s: level is not the gap, the margin still says too hard (gear, potions or resistances).", t.Title)
	}
	return m
}

func modeAfter(mode string) string {
	return "After it the hero returns to " + mode + " mode."
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// Mode returns the mode the hero acts on after the next safe moment (the stored choice).
func (e *Exec) Mode() string { e.mu.Lock(); defer e.mu.Unlock(); return e.o.Mode }
