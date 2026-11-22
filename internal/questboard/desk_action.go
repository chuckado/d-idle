package questboard

// ActionView is a button of a chunk row: the verb is the POST /api/quests/<id>/<verb> the tap sends.
type ActionView struct {
	Verb     string `json:"verb"`
	Label    string `json:"label"`
	Disabled bool   `json:"disabled,omitempty"`
	Confirm  bool   `json:"confirm,omitempty"` // a stop asks once
	Why      string `json:"why,omitempty"`     // why the button is disabled
}

// chunkAction is the single source of what a row offers. The hand-in case that bit the player (2026-11-21): a chunk that
// reads awaiting_handin with no accepted record (the game counts the field work done, case C5) has no `awaiting` object in the view, so
// its action is Start, which goes straight to the hand-in (Desk.Start); with a run otherwise Hand in. Every other live state
// has one too; the parts that offer nothing (locked, available, done, skipped, ready but not startable) say why in Reason or Why.
func chunkAction(st State, off *OfferView, run *Run, reason *ReasonView, paused bool) *ActionView {
	switch st {
	case Ready:
		if off == nil || !off.Startable {
			return nil
		}
		return &ActionView{Verb: "start", Label: "Start"}
	case Accepted, InProgress:
		return &ActionView{Verb: "cancel", Label: "Stop", Confirm: true}
	case Blocked:
		return &ActionView{Verb: "retry", Label: "Try again"}
	case Awaiting:
		switch {
		case reason != nil && reason.Code == CodeHandInReady:
			a := &ActionView{Verb: "start", Label: "Start"}
			if paused {
				a.Disabled, a.Why = true, "Paused: resume the bot first"
			}
			return a
		case run != nil && run.Status == RunHandingIn:
			return &ActionView{Verb: "handin", Label: "Handing in...", Disabled: true, Why: "The hero is talking to the quest giver"}
		}
		return &ActionView{Verb: "handin", Label: "Hand in"}
	}
	return nil
}
