package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"d2lod-bot/internal/scenario"
)

// soakRun is one run of the soak: its verdict, the invariant hits and what the fault proxy injected.
type soakRun struct {
	Run      scenario.Run   `json:"run"`
	Status   string         `json:"status"`
	Reason   string         `json:"reason,omitempty"`
	Hits     map[string]int `json:"invariants,omitempty"`
	Injected map[string]int `json:"injected,omitempty"`
	Survived bool           `json:"survived"`
	Result   string         `json:"result,omitempty"`
}

// soakReport is the whole soak: the runs and the tallies. Hard and Soft split the hits by invariants.json.
type soakReport struct {
	Faults   string         `json:"faults"`
	Duration string         `json:"duration"`
	Runs     []soakRun      `json:"runs"`
	Hard     map[string]int `json:"hard_hits"`
	Soft     map[string]int `json:"soft_hits"`
	Injected map[string]int `json:"injected"`
	Survived int            `json:"survived"`
	Died     int            `json:"died"`
	Blocked  int            `json:"blocked"`
}

// survived is the hero's own view of the run: it played to the end (not BLOCKED), life is above zero at the end and the
// monitor never saw `life-zero-alive`. A RED run can still have survived (a hard invariant hit is a bot defect, not a death).
func survived(res scenario.Result) bool {
	if res.Status == scenario.Blocked || res.State == nil {
		return false
	}
	if res.State.LifePct != nil && *res.State.LifePct <= 0 {
		return false
	}
	return res.State.Invariants["life-zero-alive"] == 0
}

func (p invariantPolicy) isHard(name string) bool {
	for _, n := range p.Hard {
		if n == name {
			return true
		}
	}
	return false
}

func addSoakRun(rep *soakReport, pol invariantPolicy, r scenario.Run, res scenario.Result) {
	sr := soakRun{Run: r, Status: res.Status, Reason: res.Reason, Survived: survived(res)}
	if res.State != nil {
		sr.Hits = res.State.Invariants
		for n, c := range res.State.Invariants {
			if pol.isHard(n) {
				rep.Hard[n] += c
			} else {
				rep.Soft[n] += c
			}
		}
	}
	if res.Faults != nil {
		sr.Injected = res.Faults.Counts
		for k, c := range res.Faults.Counts {
			rep.Injected[k] += c
		}
	}
	switch {
	case res.Status == scenario.Blocked:
		rep.Blocked++
	case sr.Survived:
		rep.Survived++
	default:
		rep.Died++
	}
	rep.Runs = append(rep.Runs, sr)
}

// soakLoop runs the rows in turn, again with the next seed after each full pass, until the duration is used up (at least one
// run). now and run are parameters so a test needs neither a clock nor a game.
func soakLoop(ctx context.Context, rows []scenario.Run, d time.Duration, now func() time.Time, run func(scenario.Run) scenario.Result, pol invariantPolicy) soakReport {
	rep := soakReport{Duration: d.String(), Hard: map[string]int{}, Soft: map[string]int{}, Injected: map[string]int{}}
	start := now()
	for pass := int64(0); len(rows) > 0; pass++ {
		for _, r := range rows {
			if len(rep.Runs) > 0 && (now().Sub(start) >= d || ctx.Err() != nil) {
				return rep
			}
			r.Seed += pass
			addSoakRun(&rep, pol, r, run(r))
		}
	}
	return rep
}

func (r soakReport) print(w io.Writer) {
	fmt.Fprintf(w, "SOAK faults=%s duration=%s runs=%d survived=%d died=%d blocked=%d\n", r.Faults, r.Duration, len(r.Runs), r.Survived, r.Died, r.Blocked)
	fmt.Fprintf(w, "injected: %s\n", tally(r.Injected))
	fmt.Fprintf(w, "hard invariant hits: %s\n", tally(r.Hard))
	fmt.Fprintf(w, "soft invariant hits: %s\n", tally(r.Soft))
	for _, s := range r.Runs {
		fmt.Fprintf(w, "  %-34s %-7s %-7s survived=%-5v hits=%s %s\n", s.Run.ID+"/"+s.Run.Class+"/"+strconv.FormatInt(s.Run.Seed, 10), s.Status, "", s.Survived, tally(s.Hits), s.Reason)
	}
}

func tally(m map[string]int) string {
	if len(m) == 0 {
		return "none"
	}
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	for i, k := range ks {
		ks[i] = fmt.Sprintf("%s=%d", k, m[k])
	}
	return strings.Join(ks, " ")
}

// cmdSoak runs a row under a fault set for a duration and reports invariant hits and survival. Report-only: it exits 0
// unless -strict is given and a hard row was hit. A hit becomes an incident scenario by hand (docs/process/proving-ground.md).
func cmdSoak(args []string, w io.Writer) error {
	fs := newFlags("soak")
	dir := fs.String("dir", defaultDir, "scenario directory")
	ids := fs.String("id", "", "ids or id prefixes ending in '.' (default: every scenario with a faults: set)")
	class := fs.String("class", "", "one class")
	seed := fs.Int64("seed", 0, "first seed")
	faults := fs.String("faults", "", "fault set of cmd/d2scenario/faults.json (default: the row's own)")
	duration := fs.Duration("duration", 30*time.Minute, "wall time of the soak; the run in progress is finished")
	out := fs.String("out", "soak-out", "output directory (soak.json, per-run files)")
	slot := fs.String("slot", "tools/d2dslot.sh", "slot wrapper")
	bot := fs.String("bot", os.Getenv("D2BOT_BIN"), "d2bot binary (env D2BOT_BIN)")
	dry := fs.Bool("dry-run", false, "print the runs of the first pass, start nothing")
	strict := fs.Bool("strict", false, "exit 3 when a hard invariant was hit (default: report only)")
	if _, err := parseArgs(fs, args); err != nil {
		return err
	}
	if *faults != "" {
		if _, err := faultSet(*faults); err != nil {
			return err
		}
	}
	scs, err := loadAll(*dir)
	if err != nil {
		return err
	}
	seedSet := flagSet(fs, "seed")
	sel := selectorFrom(scenario.TierNightly, "", *ids, *class, *seed, seedSet)
	var picked []scenario.Scenario
	for _, s := range scenario.Select(scs, sel) {
		if *ids != "" || s.Faults != "" {
			picked = append(picked, s)
		}
	}
	rows, err := scenario.Expand(picked, sel)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return fmt.Errorf("no row to soak: name one with -id or give a scenario a faults: set")
	}
	by := map[string]scenario.Scenario{}
	for _, s := range picked {
		by[s.ID] = s
		if *faults == "" {
			if _, err := faultSet(s.Faults); err != nil {
				return fmt.Errorf("scenario %s: %w (give -faults or set faults:)", s.ID, err)
			}
		}
	}
	if *dry {
		fmt.Fprintf(w, "DRY RUN: soak %s for %s, first pass of %d runs, nothing started\n", orDefault(*faults, "the rows' own sets"), *duration, len(rows))
		for _, r := range rows {
			fmt.Fprintln(w, "  "+strings.Join(execOneCmd(*slot, *dir, *bot, "", r, by[r.ID].Budget.WallMin, *out, false), " "))
		}
		return nil
	}
	if *bot == "" || os.Getenv("D2D_BIN") == "" {
		return fmt.Errorf("soak needs -bot (or D2BOT_BIN) and D2D_BIN (the lab d2d) in the environment")
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	self, _ := os.Executable()
	pol := loadInvariantPolicy()
	runOne := func(r scenario.Run) scenario.Result {
		o := filepath.Join(*out, fmt.Sprintf("%s-%s-%d", r.ID, r.Class, r.Seed))
		_ = os.MkdirAll(o, 0o755)
		a := execOneCmd(*slot, *dir, *bot, self, r, by[r.ID].Budget.WallMin, o, false)
		c := exec.Command(a[0], a[1:]...)
		c.Stdout, c.Stderr = os.Stderr, os.Stderr
		if *faults != "" {
			c.Env = append(os.Environ(), faultsEnv+"="+*faults)
		}
		err := c.Run()
		res, rerr := readResult(filepath.Join(o, resultName(r)))
		if rerr != nil {
			res = scenario.Result{Version: scenario.ResultVersion, Run: r, Status: scenario.Blocked, Reason: fmt.Sprintf("no result written (slot or runner failed: %v)", err)}
		}
		return res
	}
	rep := soakLoop(context.Background(), rows, *duration, time.Now, runOne, pol)
	rep.Faults = orDefault(*faults, "per row")
	b, _ := json.MarshalIndent(rep, "", " ")
	if err := os.WriteFile(filepath.Join(*out, "soak.json"), b, 0o644); err != nil {
		return err
	}
	rep.print(w)
	if *strict && len(rep.Hard) > 0 {
		return exitErr{3, "HARD INVARIANT HIT"}
	}
	return nil
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

// execOneCmd is the slot-wrapped exec-one command of one run (the run and soak commands share it). self is the d2scenario binary.
func execOneCmd(slot, dir, bot, self string, r scenario.Run, wallMin int, out string, journal bool) []string {
	if self == "" {
		self = "d2scenario"
	}
	a := []string{slot, "run", "-w", "3000", "-t", strconv.Itoa(wallMin*60 + 600), "--", self, "exec-one",
		"-dir", dir, "-id", r.ID, "-class", r.Class, "-seed", strconv.FormatInt(r.Seed, 10), "-out", out, "-bot", bot}
	if journal {
		a = append(a, "-journal")
	}
	return a
}
