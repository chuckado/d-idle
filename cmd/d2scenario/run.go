package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dinput"
	"d2lod-bot/internal/questboard"
	"d2lod-bot/internal/scenario"
)

// cmdRun plans, then runs every run in its own lab slot (tools/d2dslot.sh), at most -parallel at once, and aggregates the
// results. It must run ON THE BUILDER (docs/process/proving-ground.md section "Running"); -dry-run prints the commands.
func cmdRun(args []string, w io.Writer) error {
	fs := newFlags("run")
	dir := fs.String("dir", defaultDir, "scenario directory")
	tier := fs.String("tier", "fast", "fast, full or nightly")
	changed := fs.String("changed", "", "changed files (merge gate selection)")
	ids := fs.String("id", "", "ids or id prefixes ending in '.'")
	class := fs.String("class", "", "one class")
	seed := fs.Int64("seed", 0, "one seed")
	out := fs.String("out", "provingground-out", "output directory (results.json, per-run files)")
	parallel := fs.Int("parallel", 2, "lab slots at once (at most 4)")
	dry := fs.Bool("dry-run", false, "print what would run, start nothing")
	slot := fs.String("slot", "tools/d2dslot.sh", "slot wrapper")
	bot := fs.String("bot", os.Getenv("D2BOT_BIN"), "d2bot binary (env D2BOT_BIN)")
	if _, err := parseArgs(fs, args); err != nil {
		return err
	}
	seedSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "seed" {
			seedSet = true
		}
	})
	scs, err := loadAll(*dir)
	if err != nil {
		return err
	}
	if *ids != "" && !flagSet(fs, "tier") {
		*tier = scenario.TierNightly // naming scenarios selects them whatever their tier
	}
	sel := selectorFrom(*tier, *changed, *ids, *class, *seed, seedSet)
	picked := scenario.Select(scs, sel)
	runs, err := scenario.Expand(picked, sel)
	if err != nil {
		return err
	}
	by := map[string]scenario.Scenario{}
	var results []scenario.Result
	for _, s := range picked {
		by[s.ID] = s
		if s.Pending() {
			results = append(results, scenario.PendingResult(s))
		}
	}
	if len(runs) == 0 && len(results) == 0 {
		return fmt.Errorf("no scenario matches the selection (tier %s, id %q): nothing to run", *tier, *ids)
	}
	if *parallel > scenario.MaxLabSlots {
		*parallel = scenario.MaxLabSlots
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return err
	}
	self, _ := os.Executable()
	var todo []scenario.Run
	for _, r := range runs {
		s := by[r.ID]
		if why := blockedBefore(s, r); why != "" {
			results = append(results, scenario.Result{Version: scenario.ResultVersion, Run: r, Status: scenario.Blocked, Reason: why, BudgetMin: s.Budget.VirtualMin})
			continue
		}
		todo = append(todo, r)
	}
	cmdFor := func(r scenario.Run) []string {
		a := []string{*slot, "run", "-w", "3000", "-t", strconv.Itoa(by[r.ID].Budget.WallMin*60 + 600), "--", self, "exec-one",
			"-dir", *dir, "-id", r.ID, "-class", r.Class, "-seed", strconv.FormatInt(r.Seed, 10), "-out", *out, "-bot", *bot}
		return a
	}
	if *dry {
		fmt.Fprintf(w, "DRY RUN: %d runs would start, %d slots at once, nothing started\n", len(todo), *parallel)
		for _, r := range todo {
			fmt.Fprintln(w, "  "+strings.Join(cmdFor(r), " "))
		}
		for _, r := range results {
			fmt.Fprintf(w, "  not run: %-30s %s %s\n", r.ID+"/"+r.Class, r.Status, r.Reason)
		}
		return nil
	}
	if *bot == "" || os.Getenv("D2D_BIN") == "" {
		return fmt.Errorf("run needs -bot (or D2BOT_BIN) and D2D_BIN (the lab d2d) in the environment")
	}
	var mu sync.Mutex
	sem := make(chan struct{}, *parallel)
	var wg sync.WaitGroup
	for _, r := range todo {
		wg.Add(1)
		sem <- struct{}{}
		go func(r scenario.Run) {
			defer wg.Done()
			defer func() { <-sem }()
			a := cmdFor(r)
			c := exec.Command(a[0], a[1:]...)
			c.Stdout, c.Stderr = os.Stderr, os.Stderr
			err := c.Run()
			res, rerr := readResult(filepath.Join(*out, resultName(r)))
			if rerr != nil { // the run wrapper died before writing a result: that is a blocked run, never a pass
				res = scenario.Result{Version: scenario.ResultVersion, Run: r, Status: scenario.Blocked, Reason: fmt.Sprintf("no result written (slot or runner failed: %v)", err), BudgetMin: by[r.ID].Budget.VirtualMin}
			}
			mu.Lock()
			results = append(results, res)
			mu.Unlock()
		}(r)
	}
	wg.Wait()
	f, err := os.Create(filepath.Join(*out, "results.json"))
	if err != nil {
		return err
	}
	if err := scenario.WriteJSON(f, results); err != nil {
		return err
	}
	f.Close()
	scenario.Report(w, results)
	if !scenario.Gate(results) {
		return exitErr{3, "NOT GREEN"}
	}
	return nil
}

func resultName(r scenario.Run) string {
	return strings.NewReplacer("/", "_", " ", "_").Replace(r.Key()) + ".json"
}

func readResult(path string) (scenario.Result, error) {
	var r scenario.Result
	return r, readJSON(path, &r)
}

// blockedBefore names a reason when the run cannot start at all.
func blockedBefore(s scenario.Scenario, r scenario.Run) string {
	for _, n := range s.Needs {
		if n == "sandbox-d2d-binary" && os.Getenv("D2D_BIN_SANDBOX") == "" {
			return "needs the shipped sandbox d2d build: set D2D_BIN_SANDBOX"
		}
	}
	if s.Build == "sandbox" && s.Pre.Start == scenario.StartField {
		return "a field start warps the hero with a lab verb and hands the lab d2d to the bot; the shipped build refuses the gated verbs, so its field state cannot be set up (use a town start, or the lab build)"
	}
	su, err := scenario.BuildSetup(s, r.Class, r.Seed)
	if err != nil {
		return err.Error()
	}
	if len(su.Unsupported) > 0 {
		return "setup not buildable: " + strings.Join(su.Unsupported, "; ")
	}
	return ""
}

// cmdExecOne is one run inside a slot (D2D_SOCK, D2D_BOT_ADDR, D2D_SLOT_DIR from tools/d2dslot.sh).
func cmdExecOne(args []string) error {
	fs := newFlags("exec-one")
	dir := fs.String("dir", defaultDir, "scenario directory")
	id := fs.String("id", "", "scenario id")
	class := fs.String("class", "", "class")
	seed := fs.Int64("seed", 0, "seed")
	out := fs.String("out", "", "output directory")
	bot := fs.String("bot", "", "d2bot binary")
	if _, err := parseArgs(fs, args); err != nil {
		return err
	}
	scs, err := loadAll(*dir)
	if err != nil {
		return err
	}
	s, err := findScenario(scs, *id)
	if err != nil {
		return err
	}
	r := scenario.Run{ID: s.ID, Class: *class, Seed: *seed}
	start := time.Now()
	res := execOne(s, r, *out, *bot)
	res.Version, res.Run = scenario.ResultVersion, r
	if res.Status == "" { // not blocked by the infrastructure: judge the observed state
		if res.State == nil {
			res = withVerdict(res, blocked("no state was observed"))
		} else {
			res.Asserts = scenario.Evaluate(s.Assert, *res.State)
			res.Status, res.Reason = scenario.Verdict(res.Asserts)
		}
	}
	res.BudgetMin = s.Budget.VirtualMin
	res.WallSec = int(time.Since(start).Seconds())
	res.Prov = collectProv(*bot)
	res.Prov.Pace = scenario.PaceOf(res.VirtualMin, res.WallSec)
	if res.State != nil {
		res.Digest = scenario.DigestState(*res.State)
	}
	res.AssertDigest = scenario.AssertDigest(res)
	b, _ := json.MarshalIndent(res, "", " ")
	return os.WriteFile(filepath.Join(*out, resultName(r)), b, 0o644)
}

func blocked(why string, a ...any) scenario.Result {
	return scenario.Result{Status: scenario.Blocked, Reason: fmt.Sprintf(why, a...)}
}

// execOne does prep, bot and observation for one run. Every infrastructure failure is BLOCKED, never GREEN or RED: a
// red verdict is only ever a failed assertion on observed state.
func execOne(s scenario.Scenario, r scenario.Run, out, bot string) scenario.Result {
	su, err := scenario.BuildSetup(s, r.Class, r.Seed)
	if err != nil {
		return blocked("%v", err)
	}
	if len(su.Unsupported) > 0 {
		return blocked("setup not buildable: %s", strings.Join(su.Unsupported, "; "))
	}
	cidx, _ := scenario.ClassIndex(r.Class)
	hero := scenario.HeroName(s.ID, r.Class, r.Seed)
	inst := instanceID(r) // the bot wants [a-z0-9-]{1,16}
	work := filepath.Join(os.Getenv("D2D_SLOT_DIR"), "pg-"+inst)
	if os.Getenv("D2D_SLOT_DIR") == "" {
		work = filepath.Join(os.TempDir(), "pg-"+inst)
	}
	_ = os.RemoveAll(work)
	data := filepath.Join(work, "data")
	instDir := filepath.Join(data, "instances", inst)
	home, state := filepath.Join(instDir, "d2d"), filepath.Join(instDir, "state")
	for _, d := range []string{home, state} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return blocked("%v", err)
		}
	}
	sock := os.Getenv("D2D_SOCK")
	if sock == "" {
		sock = filepath.Join(work, "d2d.sock")
	}
	ledger := filepath.Join(instDir, "setup-ledger.jsonl")
	trace := filepath.Join(out, strings.TrimSuffix(resultName(r), ".json")+".bot.log")

	// 1. SETUP on a lab d2d (lab verbs on), every verb ledgered with its reason before it is sent.
	bin, gamedata := os.Getenv("D2D_BIN"), os.Getenv("D2D_DATA")
	prep := &d2dinput.Process{Bin: bin, Data: gamedata, Sock: sock, ConfigDir: filepath.Join(home, "config"), DataDir: filepath.Join(home, "data"),
		Nice: 10, ExtraArgs: []string{"--lockstep", "--lab-verbs", "--toggle", "autoloot=off,autoparty=off"}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.Budget.WallMin+15)*time.Minute)
	defer cancel()
	if err := prep.Start(ctx, d2dinput.Spec{Name: hero, Class: cidx, Seed: r.Seed}); err != nil {
		return blocked("prep d2d start: %v", err)
	}
	if err := prep.WaitInGame(ctx, 3*time.Minute); err != nil {
		_ = prep.Stop()
		return blocked("prep d2d not in game: %v", err)
	}
	c := prep.Controller().C
	if err := applySetup(ctx, c, su, ledger); err != nil {
		_ = prep.Stop()
		return blocked("setup: %v", err)
	}
	if _, err := c.Step(ctx, 50); err != nil {
		_ = prep.Stop()
		return blocked("settle: %v", err)
	}
	if err := writeBoardAndTuning(state, hero, su); err != nil {
		_ = prep.Stop()
		return blocked("board record: %v", err)
	}
	if err := writeBotConfig(data, su); err != nil {
		_ = prep.Stop()
		return blocked("bot config: %v", err)
	}
	res := scenario.Result{Tags: su.Tags, Trace: trace}

	if su.Field {
		return runField(ctx, s, r, res, prep, c, bot, data, sock, inst, trace)
	}
	if err := prep.SaveExit(ctx); err != nil {
		return blocked("prep save: %v", err)
	}
	before := newestSave(home)

	// 2. The real bot plays on a d2d WITHOUT lab verbs: gated debug verbs are refused, so no gameplay write after the start.
	botBin := bot
	d2dBot := bin
	if s.Build == "sandbox" {
		d2dBot = os.Getenv("D2D_BIN_SANDBOX")
	}
	refused, taps, berr := runBot(ctx, s, r, botBin, d2dBot, gamedata, data, home, sock, inst, hero, trace, false)
	res.Taps = taps
	if why := botFailure(trace); why != "" {
		return withVerdict(res, blocked("the bot exited with an error, it did not play: %s", why))
	}
	if why := cutReason(trace); why != "" {
		return withVerdict(res, blocked("%s", why))
	}
	if !after(newestSave(home), before) {
		res.State = &scenario.State{Version: scenario.StateVersion, Source: "none", WritesAfterStart: refused}
		return withVerdict(res, blocked("the bot run left no newer save (bot exit: %v): nothing to observe", berr))
	}
	res.VirtualMin = s.Budget.VirtualMin // the bot stops on the virtual clock (D2DSOAK_HOURS), so a normal exit is the full budget

	// 3. Observe: a read-only d2d loads the saved hero; the facts are read from the game, not from logs.
	obs := &d2dinput.Process{Bin: bin, Data: gamedata, Sock: sock, ConfigDir: filepath.Join(home, "config"), DataDir: filepath.Join(home, "data"),
		Nice: 10, ExtraArgs: []string{"--lockstep", "--toggle", "autoloot=off,autoparty=off"}}
	if err := obs.Start(ctx, d2dinput.Spec{Name: hero, Class: -1}); err != nil {
		return withVerdict(res, blocked("observer start: %v", err))
	}
	defer obs.Stop()
	if err := obs.WaitInGame(ctx, 3*time.Minute); err != nil {
		return withVerdict(res, blocked("observer not in game: %v", err))
	}
	st := observe(ctx, obs.Controller().C, refused, false)
	st.Source = "d2d-observer-reload"
	addTrail(&st, trace)
	addRunFacts(&st, s, r, trace)
	res.State = &st
	return withVerdict(res, scenario.Result{})
}

// instanceID is a stable, valid bot instance id for one run: "pg-" and 12 hex digits of the run key.
func instanceID(r scenario.Run) string {
	h := sha256.Sum256([]byte(r.Key()))
	return "pg-" + hex.EncodeToString(h[:6])
}

// botFailure returns the error text when the bot's own log ends in an "exit" error line (infrastructure: the run did not
// play, so a verdict on the state would be about nothing).
func botFailure(trace string) string {
	b, err := os.ReadFile(trace)
	if err != nil {
		return ""
	}
	for _, l := range strings.Split(string(b), "\n") {
		if strings.Contains(l, `"level":"ERROR","msg":"exit"`) {
			var e struct {
				Err string `json:"err"`
			}
			_ = json.Unmarshal([]byte(l), &e)
			return e.Err
		}
	}
	return ""
}

func after(a, b time.Time) bool { return a.After(b) }

// withVerdict fills Status and Reason: from a blocked infrastructure result when given, else from the assertions.
func withVerdict(res, b scenario.Result) scenario.Result {
	if b.Status != "" {
		res.Status, res.Reason = b.Status, b.Reason
		return res
	}
	return res
}

// newestSave returns the newest .d2s mtime under dir (zero when none).
func newestSave(dir string) time.Time {
	var t time.Time
	_ = filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		if err == nil && !fi.IsDir() && strings.HasSuffix(p, ".d2s") && fi.ModTime().After(t) {
			t = fi.ModTime()
		}
		return nil
	})
	return t
}

type ledgerLine struct {
	Time   string `json:"time"`
	Verb   string `json:"verb"`
	Reason string `json:"reason"`
	Reply  string `json:"reply,omitempty"`
}

func applySetup(ctx context.Context, c *d2dctl.Client, su scenario.Setup, ledger string) error {
	lf, err := os.OpenFile(ledger, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer lf.Close()
	enc := json.NewEncoder(lf)
	for _, v := range su.Verbs {
		l := ledgerLine{Time: time.Now().Format(time.RFC3339), Verb: v.Line, Reason: v.Reason}
		rep, err := c.Do(ctx, v.Line)
		if rep != nil {
			l.Reply = strings.Join(append([]string{rep.Rest}, rep.Lines...), " | ")
		}
		_ = enc.Encode(l)
		if err != nil {
			return fmt.Errorf("%s: %w", v.Line, err)
		}
	}
	if su.LifePct > 0 {
		return hurtTo(ctx, c, su.LifePct, enc)
	}
	return nil
}

// hurtTo brings the hero to pct percent of maximum life with `debug hit` (the game's own hurt path, V docs/forkcheats.md),
// a ledgered SETUP verb whose damage is computed from the live life.
func hurtTo(ctx context.Context, c *d2dctl.Client, pct int, enc *json.Encoder) error {
	sv, _, err := c.State(ctx)
	if err != nil || sv.Life.Max <= 0 {
		return fmt.Errorf("life setup: no life read (%v)", err)
	}
	dmg := sv.Life.Cur - sv.Life.Max*pct/100
	if dmg <= 0 {
		return nil
	}
	line := fmt.Sprintf("debug hit %d", dmg)
	l := ledgerLine{Time: time.Now().Format(time.RFC3339), Verb: line, Reason: fmt.Sprintf("SETUP: life to %d%% (%d of %d)", pct, sv.Life.Max*pct/100, sv.Life.Max)}
	rep, err := c.Do(ctx, line)
	if rep != nil {
		l.Reply = rep.Rest
	}
	_ = enc.Encode(l)
	return err
}

// writeBoardAndTuning writes the player's accepted chunk with the real plan store and the knob overrides.
func writeBoardAndTuning(state, hero string, su scenario.Setup) error {
	if len(su.Tuning) > 0 {
		b, _ := json.Marshal(map[string]any{"version": 1, "data": map[string]any{"overrides": su.Tuning}})
		if err := os.WriteFile(filepath.Join(state, "tuning.json"), b, 0o644); err != nil {
			return err
		}
	}
	if su.Board == nil {
		return nil
	}
	b, err := questboard.Load()
	if err != nil {
		return err
	}
	dir := filepath.Join(state, "quests") // cmd/d2bot questStoreDir
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	st, err := questboard.Open(b, dir, hero)
	if err != nil {
		return err
	}
	if err := st.Accept(su.Board.Quest, su.Board.Chunk, 0, "train"); err != nil {
		return err
	}
	// The store refuses a second concurrent chunk; a plan file that holds two is the state under test.
	for _, q := range su.BoardAlso {
		if err := st.AcceptIf(q.Quest, q.Chunk, 0, "train", func(questboard.Record) bool { return true }); err != nil {
			return err
		}
	}
	return nil
}

// runBot starts the real bot like tools/survival/abrun.sh does and returns how many gated writes the d2d refused (the
// bot tried a write after the start; the CRUTCH RULE line) and the bot's exit error.
func runBot(ctx context.Context, s scenario.Scenario, r scenario.Run, bot, d2dBin, gamedata, data, home, sock, inst, hero, trace string, attach bool) (int, []string, error) {
	f, err := os.Create(trace)
	if err != nil {
		return -1, nil, err
	}
	defer f.Close()
	hours := strconv.FormatFloat(s.Budget.VirtualMin/60, 'f', -1, 64)
	args := []string{"--backend", "d2d", "--char", hero, "--seed", strconv.FormatInt(r.Seed, 10),
		"--toggle", "autoloot=off,autoparty=off", "--d2d-linkwait", "60s", "--gamedir", gamedata, "--data", data,
		"--d2d-sock", sock, "--addr", os.Getenv("D2D_BOT_ADDR"), "--goals", s.Trigger.Goals,
		"--run-for", strconv.Itoa(s.Budget.WallMin) + "m"}
	if attach {
		args = append(args, "--attach", "--d2d-attach")
	} else {
		args = append(args, "--managed", "--d2d-class", "-1", "--d2d-args", "--lockstep")
	}
	var tap *tapper
	var apiRun func(context.Context)
	if s.Trigger.Driver == scenario.DriverBoardTaps || len(s.Trigger.API) > 0 {
		tokFile := filepath.Join(data, "wtoken")
		tok, err := newWatchToken(tokFile)
		if err != nil {
			return -1, nil, err
		}
		slot, _ := strconv.Atoi(os.Getenv("D2D_SLOT"))
		vp := 9300 + slot*10 // the L3 harness's port plan; the answer route is the next port
		view, ans := fmt.Sprintf("127.0.0.1:%d", vp), fmt.Sprintf("127.0.0.1:%d", vp+1)
		args = append(args, "--ladder", "--watch", "--watch-addr", view, "--watch-answer-addr", ans, "--watch-token-file", tokFile)
		tlog, _ := os.Create(strings.TrimSuffix(trace, ".bot.log") + ".taps.log")
		if tlog != nil {
			defer tlog.Close()
		} else {
			tlog = os.Stderr
		}
		if s.Trigger.Driver == scenario.DriverBoardTaps {
			tap = newTapper("http://"+view, "http://"+ans, tok, s.TappedQuest(), s.Trigger.TapChunks, tlog)
		}
		if len(s.Trigger.API) > 0 {
			apiRun = func(c context.Context) { runAPICalls(c, "http://"+ans, tok, s.Trigger.API, tlog) }
		}
	}
	tctx, stopTap := context.WithCancel(ctx)
	defer stopTap()
	done := make(chan struct{})
	go func() {
		defer close(done)
		if apiRun != nil {
			apiRun(tctx)
		}
		if tap != nil {
			select {
			case <-tctx.Done():
				return
			case <-time.After(20 * time.Second): // the watch server comes up with the bot
			}
			tap.Run(tctx, 5*time.Second)
		}
	}()
	cmd := exec.CommandContext(ctx, "nice", append([]string{"-n", "10", bot}, args...)...)
	cmd.Env = append(os.Environ(), "D2BOT_INSTANCE="+inst, "D2DSOAK_HOURS="+hours, "D2D_BIN="+d2dBin, "D2D_DATA="+gamedata)
	cmd.Stdout, cmd.Stderr = f, f
	err = cmd.Run()
	stopTap()
	<-done
	b, _ := os.ReadFile(trace)
	var taps []string
	if tap != nil {
		taps = tap.Taps
	}
	return strings.Count(string(b), "err lab verbs off"), taps, err
}

// writeBotConfig writes the bot's config file (<data>/config.json, the default of d2bot --config, S cmd/d2bot/main.go) with
// the leveler build the scenario selects (trigger.build). No build: no file, the bot keeps its defaults.
func writeBotConfig(data string, su scenario.Setup) error {
	if su.Build == "" {
		return nil
	}
	b, _ := json.Marshal(map[string]any{"levelerBuild": su.Build})
	return os.WriteFile(filepath.Join(data, "config.json"), b, 0o644)
}

// addRunFacts adds what the assertions on skill points and bot log lines need: the hero's class and the log lines that
// match a log_line assertion (S: the bot's own log, with a place in the run by a linear map of its wall clock).
func addRunFacts(st *scenario.State, s scenario.Scenario, r scenario.Run, trace string) {
	st.Class = r.Class
	b, err := os.ReadFile(trace)
	if err != nil {
		return
	}
	if lines, ok := scenario.CollectLog(s.Assert, b, s.Budget.VirtualMin); ok {
		st.Has = append(st.Has, "log")
		st.LogLines = lines
	}
}

// addTrail sets the area-change count from the bot's log (S: the bot's own `zone entry` lines).
func addTrail(st *scenario.State, trace string) {
	b, err := os.ReadFile(trace)
	if err != nil {
		return
	}
	if trail, ok := scenario.CountZoneEntries(b); ok {
		n := scenario.AreaChangeCount(trail)
		st.AreaChanges, st.AreaTrail = &n, trail
	}
}

// runField is the field start: the lab d2d (verbs on) was warped by the setup and is now handed to the bot with
// --attach; afterwards the same d2d is read. The write count is unknown (verbs stay on), so no_writes is UNOBSERVED.
func runField(ctx context.Context, s scenario.Scenario, r scenario.Run, res scenario.Result, prep *d2dinput.Process, c *d2dctl.Client, bot, data, sock, inst, trace string) scenario.Result {
	hero := scenario.HeroName(s.ID, r.Class, r.Seed)
	_ = c.Close() // the d2d takes one control client: the bot gets it, then we take it back
	_, taps, berr := runBot(ctx, s, r, bot, os.Getenv("D2D_BIN"), os.Getenv("D2D_DATA"), data, "", sock, inst, hero, trace, true)
	res.Taps = taps
	if why := botFailure(trace); why != "" {
		_ = prep.Stop()
		return withVerdict(res, blocked("the bot exited with an error, it did not play: %s", why))
	}
	if why := cutReason(trace); why != "" {
		_ = prep.Stop()
		return withVerdict(res, blocked("%s", why))
	}
	cc := d2dctl.New(sock, d2dctl.Options{})
	st := observe(ctx, cc, -1, true)
	st.Source = "d2d-live-attach"
	addTrail(&st, trace)
	addRunFacts(&st, s, r, trace)
	res.State = &st
	res.VirtualMin = s.Budget.VirtualMin
	_ = prep.Stop()
	if berr != nil && len(st.Items) == 0 && st.Area == nil {
		return withVerdict(res, blocked("bot run failed and nothing could be observed: %v", berr))
	}
	return res
}

// observe reads the facts of a loaded hero. Item places: loc 1 body, loc 2 belt, panel "stash" the stash, the rest the
// pack (U: the "panel" attribute value for the stash was not seen live).
func observe(ctx context.Context, c *d2dctl.Client, writes int, live bool) scenario.State {
	st := scenario.State{Version: scenario.StateVersion, WritesAfterStart: writes}
	if sv, _, err := c.State(ctx); err == nil {
		l, g := sv.Level, sv.Gold
		st.Level, st.Gold = &l, &g
		if sv.Life.Max > 0 {
			p := 100 * sv.Life.Cur / sv.Life.Max
			st.LifePct = &p
		}
		if v, ok := sv.Int("rskill"); ok {
			st.RightSkill = &v
		}
		if v, ok := sv.Int("lskill"); ok {
			st.LeftSkill = &v
		}
	}
	if li, err := c.Level(ctx); err == nil {
		a := li.ID
		st.Area = &a
	}
	if _, stash, err := c.Gold(ctx); err == nil {
		st.Stash = &stash
	}
	if items, _, err := c.Items(ctx); err == nil {
		st.Has = append(st.Has, "items")
		for _, it := range items {
			where := "pack"
			switch {
			case it.Loc() == 1:
				where = "body"
			case it.Loc() == 2:
				where = "belt"
			case strings.Contains(it.Attrs["panel"], "stash"):
				where = "stash"
			}
			st.Items = append(st.Items, scenario.StateItem{Code: it.Code, Where: where})
		}
	}
	if ps, err := c.Portals(ctx); err == nil {
		st.Has = append(st.Has, "portals")
		for _, p := range ps.Items {
			st.Portals = append(st.Portals, scenario.StatePortal{Which: p.Which, Level: p.Level})
		}
	}
	if gs, err := c.Ground(ctx); err == nil && live { // the floor of the level the hero stands in: only a live observation means it
		st.Has = append(st.Has, "ground")
		for _, g := range gs.Items {
			if !g.IsGold() {
				st.Items = append(st.Items, scenario.StateItem{Code: g.Code, Where: "ground"})
			}
		}
	}
	if rep, err := c.Do(ctx, "dump player,stats,skills"); err == nil {
		st.Stats = pointStats(rep.Lines, rep.Rest)
		st.SkillLevels = skillLevels(rep.Lines, rep.Rest)
	}
	if rep, err := c.Do(ctx, "debug quest list"); err == nil {
		st.Quests = map[int]uint16{}
		for _, l := range rep.Lines {
			f := strings.Fields(l)
			if len(f) < 2 {
				continue
			}
			rec, e1 := strconv.Atoi(f[0])
			w, e2 := strconv.ParseUint(strings.TrimPrefix(f[1], "0x"), 16, 16)
			if e1 == nil && e2 == nil {
				st.Quests[rec] = uint16(w)
			}
		}
	}
	if rep, err := c.Do(ctx, "debug objects"); err == nil {
		st.Has = append(st.Has, "objects")
		for _, l := range rep.Lines { // columns: index x y kind class locked state (V docs/probes/search-for-cain-l0)
			f := strings.Fields(l)
			if len(f) >= 7 {
				if cl, e := strconv.Atoi(f[4]); e == nil {
					st.Objects = append(st.Objects, scenario.StateObject{Class: cl, State: f[len(f)-1]})
				}
			}
		}
	}
	return st
}

// pointStats reads the hero's points out of `dump player,stats,skills` (V dump_head_late.json: player.stat_pts and
// skill_pts, stats.base str/dex/vit/ene, skills.base_levels). nil when the dump is not there (stock d2d) or unreadable.
func pointStats(lines []string, rest string) map[string]int {
	d, ok := readPointDump(lines, rest)
	if !ok {
		return nil
	}
	out := map[string]int{}
	for _, k := range []string{"str", "dex", "vit", "ene"} {
		if v, ok := d.Stats.Base[k]; ok {
			out[k] = v
		}
	}
	if d.Player.StatPts != nil {
		out["statpts"] = *d.Player.StatPts
	}
	if d.Player.SkillPts != nil {
		out["skillpts"] = *d.Player.SkillPts
	}
	if d.Skills.Levels != nil {
		sum := 0
		for _, n := range d.Skills.Levels {
			sum += n
		}
		out["skillsspent"] = sum
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// skillLevels is the dump's skills.base_levels (class order), nil when the dump has none.
func skillLevels(lines []string, rest string) []int {
	if d, ok := readPointDump(lines, rest); ok {
		return d.Skills.Levels
	}
	return nil
}

type pointDump struct {
	Player struct {
		StatPts  *int `json:"stat_pts"`
		SkillPts *int `json:"skill_pts"`
	} `json:"player"`
	Stats struct {
		Base map[string]int `json:"base"`
	} `json:"stats"`
	Skills struct {
		Levels []int `json:"base_levels"`
	} `json:"skills"`
}

// readPointDump decodes the first non-empty line of a `dump player,stats,skills` reply.
func readPointDump(lines []string, rest string) (pointDump, bool) {
	line := strings.TrimSpace(rest)
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			line = l
			break
		}
	}
	var d pointDump
	return d, json.Unmarshal([]byte(line), &d) == nil
}
