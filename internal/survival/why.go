package survival

// Why is one reason in the shape of docs/ui/redesign-2026-11-14/C-orders.md: a closed code the UI and the hero voice can
// phrase, and a plain sentence with no numbers (the watch screen never shows a margin, user decision 2026-11-14).
type Why struct {
	Code string `json:"code"`
	Text string `json:"text"`
}

// The closed list of codes (one file, design 8.1).
const (
	WhyBestPay      = "best_pay_eligible" // the zone pays most among the safe ones
	WhyRiskPaid     = "pays_for_risk"     // a little riskier than the safe zones, and it pays clearly more
	WhyStay         = "stay_current"      // another zone pays more, but not by enough to leave
	WhyDepleted     = "depleted_next"     // the best zone is cleared; the next safe one
	WhyFresh        = "fresh_game"        // the best zone is cleared and a new game beats walking to a poor one
	WhyHole         = "hole_first_field"  // nothing is safe: the easiest field, and the hero needs help
	WhyMarginLow    = "margin_low"        // not safe enough to enter
	WhyStayLine     = "below_stay_line"   // the current zone fell under its stay line
	WhyCost         = "hard_events"       // chickens or merc deaths here crossed the dial's limit
	WhyDied         = "died"              // a death here crossed the dial's limit
	WhyLeftForCost  = "left_for_cost"     // left for cost; closed until the hero is stronger
	WhyMercDead     = "merc_left_dead"    // the merc stays dead and why
	WhyMercRevive   = "merc_revive"       // revive him
	WhyMercNoRevive = "merc_zone_kills"   // the zone kills him twice; farm where he survives or without him
)

// Codes lists every code, for the UI's phrase table and the tests.
var Codes = []string{WhyBestPay, WhyRiskPaid, WhyStay, WhyDepleted, WhyFresh, WhyHole, WhyMarginLow, WhyStayLine, WhyCost, WhyDied,
	WhyLeftForCost, WhyMercDead, WhyMercRevive, WhyMercNoRevive}
