package drill

import (
	"context"
	"fmt"
	"time"
)

// Runner runs drills on one World.
type Runner struct {
	W            World
	Repeat       int           // runs per drill (0: 1)
	DefaultLimit time.Duration // hard limit of one run when the drill sets none (0: 5 minutes)
	Grace        time.Duration // how long to wait for a body that ignores its context after the limit (0: 5 s)
	Logf         func(format string, a ...any)
}

func (r *Runner) logf(f string, a ...any) {
	if r.Logf != nil {
		r.Logf(f, a...)
	}
}

// Group is the outcome of RunGroup.
type Group struct {
	Results []Result `json:"results"`
	Stopped string   `json:"stopped,omitempty"` // why the remaining drills did not run
}

// gate is the P35 check: reads the UI, and on a dirty one tries to recover once. It returns the first observation,
// whether the UI was clean at first, and whether it is clean now.
func (r *Runner) gate(ctx context.Context) (first UI, wasClean, nowClean bool, err error) {
	first, err = r.W.UI(ctx)
	if err != nil {
		return first, false, false, fmt.Errorf("ui read: %w", err)
	}
	if first.Clean() {
		return first, true, true, nil
	}
	if rerr := r.W.Recover(ctx); rerr != nil {
		r.logf("recover: %v", rerr)
	}
	again, err := r.W.UI(ctx)
	if err != nil {
		return first, false, false, fmt.Errorf("ui read after recovery: %w", err)
	}
	return first, false, again.Clean(), nil
}

func p35(where, id string, ui UI, recovered bool) Result {
	reason := fmt.Sprintf("dirty UI %s drill %s", where, id)
	obs := ui.String()
	if recovered {
		obs += "; recovered with one Escape/close"
	} else {
		obs += "; recovery failed, group stopped"
	}
	return Result{ID: "P35." + where + "." + id, Primitive: "P35", Verdict: Fail, Reason: reason,
		Intended: "cursor empty, no panel, no dialog", Observed: obs, Deterministic: true}
}

// RunGroup runs the drills in order. A dirty UI that one recovery does not clear stops the group.
func (r *Runner) RunGroup(ctx context.Context, drills []Drill) Group {
	var g Group
	for i, d := range drills {
		if len(d.Needs) > 0 {
			g.Results = append(g.Results, Result{ID: d.ID, Primitive: d.Primitive, Title: d.Title, Verdict: Blocked,
				Reason: "missing: " + joinNeeds(d.Needs), Deterministic: true})
			r.logf("%s", g.Results[len(g.Results)-1].Line())
			continue
		}
		res, stop := r.runDrill(ctx, d, &g)
		g.Results = append(g.Results, res)
		r.logf("%s", res.Line())
		if stop != "" {
			g.Stopped = stop
			for _, rest := range drills[i+1:] {
				g.Results = append(g.Results, Result{ID: rest.ID, Primitive: rest.Primitive, Title: rest.Title, Verdict: Blocked,
					Reason: "not run: " + stop, Deterministic: true})
			}
			return g
		}
		if ctx.Err() != nil {
			g.Stopped = ctx.Err().Error()
			return g
		}
	}
	return g
}

func joinNeeds(n []string) string {
	s := ""
	for i, x := range n {
		if i > 0 {
			s += "; "
		}
		s += x
	}
	return s
}

func (r *Runner) runDrill(ctx context.Context, d Drill, g *Group) (Result, string) {
	n := max(r.Repeat, 1)
	res := Result{ID: d.ID, Primitive: d.Primitive, Title: d.Title}
	limit := d.Limit
	if limit == 0 {
		limit = r.DefaultLimit
	}
	if limit == 0 {
		limit = 5 * time.Minute
	}
	stop := ""
	for k := 0; k < n; k++ {
		pre, was, now, err := r.gate(ctx)
		if err != nil {
			res.Verdict, res.Reason = Blocked, "UI gate unreadable before the drill: "+err.Error()
			return res, "ui unreadable"
		}
		if !was {
			g.Results = append(g.Results, p35("pre", d.ID, pre, now))
			r.logf("%s", g.Results[len(g.Results)-1].Line())
			if !now {
				res.Verdict, res.Reason = Blocked, "dirty UI before the drill, not run"
				return res, "dirty UI before " + d.ID
			}
		}
		rr, run := r.once(ctx, d, limit)
		rr.UIPre = pre.String()
		post, was, now, err := r.gate(ctx)
		switch {
		case err != nil:
			rr.UIPost = "unreadable: " + err.Error()
			rr.Verdict, rr.Reason = Fail, rr.Reason+" (UI gate unreadable after the drill: "+err.Error()+")"
			stop = "ui unreadable after " + d.ID
		case !was:
			rr.UIPost = post.String()
			g.Results = append(g.Results, p35("post", d.ID, post, now))
			r.logf("%s", g.Results[len(g.Results)-1].Line())
			if rr.Verdict != Blocked {
				rr.Verdict, rr.Reason = Fail, rr.Reason+" (drill left a dirty UI: "+post.String()+")"
			}
			if !now {
				stop = "dirty UI after " + d.ID
			}
		default:
			rr.UIPost = "clean"
		}
		if run != nil {
			rr.Cheats = run.Cheats
		}
		res.Runs = append(res.Runs, rr)
		if stop != "" || ctx.Err() != nil {
			break
		}
	}
	summarise(&res)
	return res, stop
}

// once runs the body under the time limit; a panic or an overrun is a FAIL.
func (r *Runner) once(ctx context.Context, d Drill, limit time.Duration) (RunResult, *Run) {
	run := &Run{W: r.W, Drill: d.ID}
	cctx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	start := time.Now()
	done := make(chan Outcome, 1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				done <- Outcome{Verdict: Fail, Reason: fmt.Sprintf("drill panicked: %v", p)}
			}
		}()
		done <- d.Do(cctx, run)
	}()
	grace := r.Grace
	if grace == 0 {
		grace = 5 * time.Second
	}
	var out Outcome
	select {
	case out = <-done:
	case <-cctx.Done():
		select {
		case out = <-done:
		case <-time.After(grace):
			out = Outcome{Verdict: Fail, Reason: fmt.Sprintf("hard time limit %s: the drill did not return", limit)}
		}
	}
	if out.Verdict == "" {
		out.Verdict, out.Reason = Fail, "drill returned no verdict"
	}
	if cctx.Err() == context.DeadlineExceeded && out.Verdict == Pass {
		out.Verdict, out.Reason = Fail, fmt.Sprintf("passed only after the hard time limit %s", limit)
	}
	return RunResult{Outcome: out, Elapsed: time.Since(start)}, run
}

// summarise folds the repeats: all PASS passes, any BLOCKED blocks, otherwise FAIL; Deterministic compares the
// observed end state and verdict of every repeat.
func summarise(res *Result) {
	if len(res.Runs) == 0 {
		return
	}
	first := res.Runs[0]
	res.Deterministic = true
	verdict := Pass
	for _, rr := range res.Runs {
		if rr.Verdict != first.Verdict || rr.sig() != first.sig() {
			res.Deterministic = false
		}
		switch {
		case rr.Verdict == Blocked:
			verdict = Blocked
		case rr.Verdict == Fail && verdict != Blocked:
			verdict = Fail
		}
	}
	res.Verdict, res.Reason = verdict, first.Reason
	res.Intended, res.Observed = first.Intended, first.Observed
	res.Cheats = first.Cheats
	if !res.Deterministic {
		for i, rr := range res.Runs[1:] {
			if rr.Verdict != first.Verdict || rr.sig() != first.sig() {
				res.Observed = fmt.Sprintf("run1: %s; run%d (%s): %s", first.Observed, i+2, rr.Verdict, rr.Observed)
				if verdict == Pass {
					verdict, res.Verdict = Fail, Fail
					res.Reason = "repeats disagree: " + first.Reason
				}
				break
			}
		}
	}
}
