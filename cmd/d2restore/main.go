// d2restore: the snapshot library (docs/snaplib.md). snapshot copies a character's file set into the
// library with a manifest, verify re-checks it, list shows it, place copies a snapshot into ANOTHER
// instance's prefix under a new name and logs it in the cheat ledger as a declared setup-copy.
//
//	d2restore snapshot --name N (--inst iN | --src-dir D) --at-menu [--id ID] [--run R] [--bot-commit H]
//	d2restore verify ID | list [--json]
//	d2restore place --snapshot ID --as NEW --inst iN --reason TEXT [--run R] [--allow-cheated]
//
// Common: --lib DIR (default $D2RESTORE_LIB or <data>/snaplib), --ledger FILE. Offline: no game is started.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"d2lod-bot/internal/cheat"
	"d2lod-bot/internal/mem"
	"d2lod-bot/internal/paths"
	"d2lod-bot/internal/saves"
)

// Env holds everything run touches outside its arguments, so tests can fake it.
type Env struct {
	Lib     string
	Ledger  *cheat.Ledger
	Now     func() time.Time
	SaveDir func(inst string) string
	// Preflight refuses unless the prefix of inst is free to receive a character (no instance lock held,
	// no game running); release must be called when the placement is done.
	Preflight func(inst string) (release func(), err error)
}

func defaultEnv() *Env {
	lib := os.Getenv("D2RESTORE_LIB")
	if lib == "" {
		l, err := paths.FromEnv()
		if err != nil {
			l = paths.Layout{Root: os.TempDir() + "/d2bot"}
		}
		lib = filepath.Join(l.Root, "snaplib")
	}
	return &Env{Lib: lib, Ledger: cheat.NewLedger(cheat.DefaultLedgerPath()), Now: time.Now,
		SaveDir: cheat.DefaultSaveDir, Preflight: preflight}
}

func preflight(inst string) (func(), error) {
	l, err := paths.FromEnv()
	if err != nil {
		return nil, err
	}
	if _, err := l.Instance(inst); err != nil {
		return nil, err
	}
	if len(mem.FindGamesFor(inst)) > 0 {
		return nil, fmt.Errorf("a game is running in the prefix of %s", inst)
	}
	return lockAll(l.GameLocks(inst))
}

func main() {
	if err := run(os.Args[1:], defaultEnv(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "d2restore:", err)
		os.Exit(1)
	}
}

func run(args []string, env *Env, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: d2restore snapshot|verify|list|place ... (see docs/snaplib.md)")
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	lib := fs.String("lib", env.Lib, "library directory")
	ledger := fs.String("ledger", "", "cheat ledger file (default: the standard one)")
	name := fs.String("name", "", "snapshot: character name")
	inst := fs.String("inst", "", "instance whose prefix is the source (snapshot) or target (place)")
	dir := fs.String("dir", "", "override the save directory of --inst")
	atMenu := fs.Bool("at-menu", false, "snapshot: assert the game is at a menu or not running")
	id := fs.String("id", "", "snapshot: library id (default NAME-UTCSTAMP)")
	snap := fs.String("snapshot", "", "place: library id")
	as := fs.String("as", "", "place: new character name")
	reason := fs.String("reason", "", "place: why (required, goes in the ledger)")
	actor := fs.String("actor", os.Getenv("D2BOT_AGENT"), "ledger actor")
	runID := fs.String("run", "", "run id for the manifest or the ledger line")
	botCommit := fs.String("bot-commit", os.Getenv("D2BOT_COMMIT"), "snapshot: bot commit that played the character")
	allow := fs.Bool("allow-cheated", false, "place: allow a source whose ledger status is cheated (provenance copy-of-cheated)")
	asJSON := fs.Bool("json", false, "list: JSON")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	env.Lib = *lib
	if *ledger != "" {
		env.Ledger = cheat.NewLedger(*ledger)
	}
	saveDir := func() (string, error) {
		switch {
		case *dir != "":
			return *dir, nil
		case *inst != "":
			return env.SaveDir(*inst), nil
		}
		return "", errors.New("need --inst or --dir")
	}
	switch args[0] {
	case "snapshot":
		d, err := saveDir()
		if err != nil || *name == "" {
			return errors.Join(err, requireName(*name))
		}
		key := cheat.CharKey(*inst, *name)
		var cheated bool
		if *inst == "" {
			cheated = env.Ledger.Cheated(time.Time{}, *name) // no prefix: any same-named cheat counts
		} else {
			cheated = env.Ledger.CheatedKey(time.Time{}, key)
		}
		status := "honest"
		if cheated {
			status = "cheated"
		}
		m, err := saves.LibSnapshot(d, env.Lib, *name, saves.SnapInfo{ID: *id, BotCommit: *botCommit,
			SourceRun: *runID, SourceInst: *inst, LedgerStatus: status}, saves.Options{AtMenu: *atMenu}, env.Now())
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "snapshot %s: %s %s clvl %d xp %d %s act %d, %d files, ledger %s\n", m.ID, m.Name, m.ClassName, m.Level, m.XP, m.Difficulty, m.Act, len(m.Files), m.LedgerStatus)
	case "verify":
		if fs.NArg() != 1 {
			return errors.New("verify ID")
		}
		m, _, err := saves.LibVerify(env.Lib, fs.Arg(0))
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "ok %s: %d files, sha256 and .d2s header match (loads-and-continues is a live check)\n", m.ID, len(m.Files))
	case "list":
		ms, err := saves.LibList(env.Lib)
		if err != nil {
			return err
		}
		for _, m := range ms {
			if *asJSON {
				b, _ := json.Marshal(m)
				fmt.Fprintln(out, string(b))
				continue
			}
			fmt.Fprintf(out, "%-32s %-12s %-12s clvl %-3d %-9s act %d %s %s\n", m.ID, m.Name, m.ClassName, m.Level, m.Difficulty, m.Act, m.LedgerStatus, m.Date.Format("2006-01-02"))
		}
	case "place":
		return place(env, out, *snap, *as, *inst, *dir, *reason, *actor, *runID, *allow)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
	return nil
}

func requireName(n string) error {
	if n == "" {
		return errors.New("need --name")
	}
	return nil
}

// place is the guarded copy. Order: arguments, ledger writable, preflight, copy, ledger line. A refusal
// after the arguments are valid is also recorded (Success false), like a refused cheat.
func place(env *Env, out io.Writer, snap, as, inst, dirOverride, reason, actor, run string, allow bool) error {
	switch {
	case snap == "" || as == "" || inst == "":
		return errors.New("place needs --snapshot, --as and --inst")
	case reason == "":
		return errors.New("place needs --reason (the ledger line is the setup declaration)")
	case inst == "i0":
		return errors.New("never place into i0 (the user's desktop game)")
	}
	if actor == "" {
		actor = "unknown"
	}
	if err := env.Ledger.Ensure(); err != nil {
		return fmt.Errorf("ledger not writable, refusing to place: %w", err)
	}
	e := cheat.Entry{Time: env.Now().UTC(), Actor: actor, Instance: inst, Character: as, Kind: cheat.KindSetupCopy,
		Reason: reason, Params: map[string]any{"snapshot": snap, "run": run}}
	refuse := func(err error) error {
		e.Error = err.Error()
		_ = env.Ledger.Append(e)
		return err
	}
	dest := dirOverride
	if dest == "" {
		dest = env.SaveDir(inst)
	}
	release, err := env.Preflight(inst)
	if err != nil {
		return refuse(fmt.Errorf("preflight: %w", err))
	}
	defer release()
	p, err := saves.Place(env.Lib, snap, dest, as, allow)
	if err != nil {
		return refuse(err)
	}
	e.Success = true
	e.Params = map[string]any{"snapshot": snap, "source_name": p.Manifest.Name, "provenance": p.Provenance,
		"sha256_source_d2s": p.SourceD2S, "sha256_placed_d2s": p.PlacedD2S, "files": p.Files,
		"source_ledger_status": p.Manifest.LedgerStatus, "clvl": p.Manifest.Level, "run": run}
	if err := env.Ledger.Append(e); err != nil {
		return fmt.Errorf("placed %s but the ledger line failed: %w", as, err)
	}
	fmt.Fprintf(out, "placed %s in %s: %d files, %s\nsetup: %s %s\n", as, inst, len(p.Files), p.Provenance, p.Provenance, snap)
	return nil
}
