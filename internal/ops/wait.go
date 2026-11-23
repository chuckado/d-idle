package ops

import "time"

// Outcomes of Wait.
const (
	OutFinished = "finished" // the chunk is done
	OutBlocked  = "blocked"  // the chunk is parked or blocked; Code and Text say why
	OutStopped  = "stopped"  // it was running and went back to ready (cancelled)
	OutIdle     = "idle"     // nothing of the quest was running when the wait began
	OutTimeout  = "timeout"
)

// Result is what Wait reports: the evidence a loop script needs without reading logs.
type Result struct {
	Outcome  string
	Quest    string
	Chunk    string
	State    string
	Step     string
	Progress int
	Code     string
	Text     string
}

func active(s string) bool { return s == "accepted" || s == "in_progress" || s == "awaiting_handin" }

// Wait polls the board until the chunk is done, blocked, stopped or the timeout passes. With no chunk it follows the
// quest's active chunk (else its blocked one). poll reads the board; sleep and now are injected for the tests.
func Wait(poll func() (*Board, error), quest, chunk string, timeout, every time.Duration, sleep func(time.Duration), now func() time.Time) Result {
	end := now().Add(timeout)
	res := Result{Quest: quest, Chunk: chunk}
	seenActive := false
	for first := true; ; first = false {
		b, err := poll()
		if err == nil {
			var rows []Row
			for _, r := range b.Rows {
				if r.Quest == quest && (chunk == "" || r.Chunk == chunk) {
					rows = append(rows, r)
				}
			}
			var cur *Row
			if res.Chunk != "" || chunk != "" {
				for i := range rows {
					if rows[i].Chunk == res.Chunk {
						cur = &rows[i]
					}
				}
			} else {
				for _, want := range []func(string) bool{active, func(s string) bool { return s == "blocked" }} {
					for i := range rows {
						if cur == nil && want(rows[i].State) {
							cur = &rows[i]
						}
					}
				}
				if cur == nil && len(rows) > 0 {
					all := true
					for _, r := range rows {
						all = all && (r.State == "done" || r.State == "skipped")
					}
					if all {
						cur = &rows[len(rows)-1]
					}
				}
			}
			if cur != nil {
				res.Chunk, res.State, res.Step, res.Progress, res.Code, res.Text = cur.Chunk, cur.State, cur.Step, cur.Progress, cur.Code, cur.Text
				switch {
				case cur.State == "done" || cur.State == "skipped":
					res.Outcome = OutFinished
					return res
				case cur.State == "blocked":
					res.Outcome = OutBlocked
					return res
				case active(cur.State):
					seenActive = true
				default:
					res.Outcome = OutIdle
					if seenActive {
						res.Outcome = OutStopped
					}
					return res
				}
			} else if first {
				res.Outcome = OutIdle
				return res
			}
		}
		if !now().Before(end) {
			res.Outcome = OutTimeout
			return res
		}
		sleep(every)
	}
}
