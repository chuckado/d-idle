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
	group := flag.String("group", "D1", "drill group (D1)")
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
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: lvl}))

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

	var drills []drill.Drill
	switch strings.ToUpper(*group) {
	case "D1":
		drills = drill.D1(l, strings.Split(*belts, ","))
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

	// The character stays alive and in town.
	if s, err := be.Env.Snapshot(); err == nil && s.Player != nil {
		fmt.Printf("END character %s life %d/%d in town: %v\n", *name, s.Player.Life, s.Player.MaxLife, s.Player.InTown())
	}
	fmt.Printf("\n%s", g.Table())
	fmt.Printf("cheat ledger: %s (actor drills:<drill id>, kind d2d-debug)\n", lp)
	if err := g.Write(*out, *date+"-"+strings.ToLower(*group)); err != nil {
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
