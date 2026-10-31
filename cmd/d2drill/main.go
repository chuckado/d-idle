// Command d2drill runs the live primitive drills (card m1-liveproof, docs/primitives.md) on ONE hidden d2d instance
// with a throwaway character: setup through the logged cheat tool only, the primitive through the bot's real code path,
// the end state through the StateReader, a P35 UI gate around every drill.
//
// Run it on the worker host inside a slot, under the process registry:
//
//	D2D_AGENT=drills tools/agentrun.sh drills --ttl 3600 -- tools/d2dslot.sh run -t 3600 -- d2drill -group D1
//
// Output: one "DRILL ..." line per drill on stdout, docs/probes/drills/<date>-<group>.txt and .json.
package main

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"d2lod-bot/internal/cheat"
	"d2lod-bot/internal/d2denv"
	"d2lod-bot/internal/d2dinput"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/drill"
	"d2lod-bot/internal/env"
)

func main() { os.Exit(run()) }

func run() int {
	home := os.Getenv("HOME")
	slotDir := os.Getenv("D2D_SLOT_DIR")
	group := flag.String("group", "D1", "drill group (D0 gate self-test, D1 belt, D2 portals and town entry)")
	only := flag.String("only", "", "comma list of drill ids to run (default: the whole group)")
	repeat := flag.Int("repeat", 2, "runs per drill (determinism check)")
	bin := flag.String("bin", envOr("D2D_BIN", filepath.Join(home, "d2d-lab/fork/bin/d2d.lab6")), "d2d binary (the lab build)")
	gamedir := flag.String("gamedir", envOr("D2D_DATA", filepath.Join(home, "d2d-lab/data")), "game data directory (read only)")
	hdir := flag.String("home", filepath.Join(slotDir, "drillhome"), "d2d XDG home (own saves and log)")
	cache := flag.String("cache", filepath.Join(home, "d2d-lab/drills/tables"), "extracted game tables cache")
	sock := flag.String("sock", envOr("D2D_SOCK", ""), "control socket (the slot's)")
	seed := flag.Int64("seed", 11, "map seed")
	class := flag.Int("class", 4, "d2d class index of the new character")
	name := flag.String("name", "Drilly", "throwaway character name")
	out := flag.String("out", "docs/probes/drills", "result directory")
	date := flag.String("date", time.Now().UTC().Format("2006-01-02"), "date prefix of the result files")
	ledger := flag.String("ledger", "", "cheat ledger (default: $D2CHEAT_LEDGER or ~/.local/share/d2bot/ledger/cheats.jsonl)")
	belts := flag.String("belts", "lbl,vbl,mbl,tbl,hbl", "belt item codes P2 equips")
	verbose := flag.Bool("v", false, "debug log of the bot code paths on stderr")
	probe := flag.String("probe", "", "semicolon list of verbs to run and print (debug verbs go through the cheat ledger), then exit; for finding out what a verb does")
	label := flag.String("label", "", "result file label after the date (default: the group, lower case)")
	limit := flag.Duration("limit", 0, "override the per-run hard time limit")
	flag.Parse()
	if *sock == "" || slotDir == "" {
		fmt.Fprintln(os.Stderr, "d2drill: run inside tools/d2dslot.sh (needs D2D_SOCK and D2D_SLOT_DIR)")
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	lvl := slog.LevelWarn
	if *verbose {
		lvl = slog.LevelDebug
	}
	// the lab keeps every record in memory too (the drills quote the bot's failure lines); stderr only shows lvl and up
	rec := newRecorder(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))
	log := slog.New(rec)

	db, err := data.Open(*gamedir, *cache)
	if err != nil {
		fmt.Fprintln(os.Stderr, "d2drill: game tables:", err)
		return 1
	}
	tune := env.DefaultTuning()
	if err := os.MkdirAll(*hdir, 0o700); err != nil {
		fmt.Fprintln(os.Stderr, "d2drill:", err)
		return 1
	}
	be, err := d2denv.Build(ctx, d2denv.Config{ID: "drills", Log: log, Tune: &tune, Data: db, Sock: *sock, LinkWait: 30 * time.Second,
		Bin: *bin, DataDir: *gamedir, Home: *hdir, Nice: 10, LogFile: filepath.Join(*hdir, "d2d.out"),
		ExtraArgs: []string{"--toggle", "autoloot=off,autoparty=off", "--lockstep"},
		Spec:      d2dinput.Spec{Name: *name, Class: *class, Seed: *seed, NoSave: true}})
	if err != nil {
		fmt.Fprintln(os.Stderr, "d2drill: build:", err)
		return 1
	}
	defer be.Close()
	if err := be.Start(ctx, 0); err != nil {
		fmt.Fprintln(os.Stderr, "d2drill: start d2d:", err)
		return 1
	}
	lp := *ledger
	if lp == "" {
		lp = envOr("D2CHEAT_LEDGER", filepath.Join(home, ".local/share/d2bot/ledger/cheats.jsonl")) // not the slot's XDG data dir
	}
	led := cheat.NewLedger(lp)
	if err := led.Ensure(); err != nil {
		fmt.Fprintln(os.Stderr, "d2drill: the cheat ledger must be writable before any cheat:", err)
		return 1
	}
	l := newLab(be, db, led, *name, "slot"+os.Getenv("D2D_SLOT"), log)
	l.rec = rec
	if *probe != "" {
		return l.runProbe(ctx, *probe)
	}

	var drills []drill.Drill
	switch strings.ToUpper(*group) {
	case "D1":
		drills = drill.D1(l, strings.Split(*belts, ","))
	case "D2":
		drills = drill.D2(l)
	case "D0": // harness self-test of the P35 gate: the drill leaves the pack open on purpose
		drills = []drill.Drill{{ID: "D0.gate", Primitive: "P35", Title: "selftest: the drill leaves the inventory open, the gate must flag and recover it", Limit: time.Minute,
			Do: func(ctx context.Context, r *drill.Run) drill.Outcome {
				if err := be.Client.Key(ctx, "i"); err != nil {
					return drill.Failed("key i: "+err.Error(), "", "")
				}
				_, _ = be.Client.Step(ctx, 10)
				return drill.Passed("opened the inventory and returned without closing it", "gate reports a dirty UI after the drill", "inventory open")
			}}}
	default:
		fmt.Fprintln(os.Stderr, "d2drill: unknown group", *group)
		return 2
	}
	if *only != "" {
		want := map[string]bool{}
		for _, id := range strings.Split(*only, ",") {
			want[strings.TrimSpace(id)] = true
		}
		var keep []drill.Drill
		for _, d := range drills {
			if want[d.ID] {
				keep = append(keep, d)
			}
		}
		drills = keep
	}
	if *limit > 0 {
		for i := range drills {
			drills[i].Limit = *limit
		}
	}
	r := &drill.Runner{W: l, Repeat: *repeat, Logf: func(f string, a ...any) { fmt.Printf(f+"\n", a...) }}
	g := r.RunGroup(ctx, drills)

	if strings.ToUpper(*group) == "D2" { // the portal drills end in the field: put the character back in town (logged cheat)
		if _, err := l.Cheat(ctx, "drills:cleanup", "debug goto 1", "return the throwaway character to the Rogue Encampment after the portal drills"); err != nil {
			fmt.Fprintln(os.Stderr, "d2drill: cleanup goto:", err)
		}
		_ = l.Tick(ctx, 25)
	}
	// The character stays alive and in town.
	if s, err := be.Env.Snapshot(); err == nil && s.Player != nil {
		fmt.Printf("END character %s life %d/%d in town: %v\n", *name, s.Player.Life, s.Player.MaxLife, s.Player.InTown())
	}
	fmt.Printf("\n%s", g.Table())
	fmt.Printf("cheat ledger: %s (actor drills:<drill id>, kind d2d-debug)\n", lp)
	if err := g.Write(*out, *date+"-"+cmp.Or(*label, strings.ToLower(*group))); err != nil {
		fmt.Fprintln(os.Stderr, "d2drill: write results:", err)
		return 1
	}
	return 0
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
