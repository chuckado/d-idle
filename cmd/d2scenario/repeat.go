package main

import (
	"fmt"

	"d2lod-bot/internal/scenario"
)

// aggregateRepeats is the verdict of one scenario run -repeat times: GREEN when at least need runs are GREEN (the first
// green result stands for the row), else the first RED (or BLOCKED) result with the tally in its reason. A row with a thin
// fight margin gives GREEN and RED on the same commit (V matrix note 68, drill.act2.radament.l30), so one run proves nothing.
func aggregateRepeats(rs []scenario.Result, need int) scenario.Result {
	green := 0
	var firstGreen, firstBad *scenario.Result
	for i := range rs {
		switch rs[i].Status {
		case scenario.Green:
			green++
			if firstGreen == nil {
				firstGreen = &rs[i]
			}
		default:
			if firstBad == nil {
				firstBad = &rs[i]
			}
		}
	}
	tally := fmt.Sprintf("%d/%d green, %d needed", green, len(rs), need)
	if green >= need && firstGreen != nil {
		r := *firstGreen
		if len(rs) > 1 {
			r.Reason = tally
		}
		return r
	}
	if firstBad == nil { // fewer runs than needed, all green
		r := *firstGreen
		r.Status, r.Reason = scenario.Red, tally
		return r
	}
	r := *firstBad
	if len(rs) > 1 {
		r.Reason = tally + "; first failure: " + r.Reason
	}
	return r
}

// repeatNeed is the green count a repeat of n runs needs: two of three, in general the ceiling of two thirds.
func repeatNeed(n int) int { return (2*n + 2) / 3 }
