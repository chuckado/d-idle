package main

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/charsel"
	"d2lod-bot/internal/paths"
	"d2lod-bot/internal/watch"
)

// The Characters desk (docs/charselect.md): list, create, switch, delete and restore for the Character Select page. The
// d2d backend has no verb for any of them; the list, delete and restore are file operations on the save folder, create
// and switch end the bot with exitSwitch and the SUPERVISOR starts the next one with another --char (supervise.go).

// exitSwitch is the exit code of a deliberate character switch: the supervisor restarts at once with the intent's
// character and does not count it as a crash.
const exitSwitch = 6

// switchDelay lets the HTTP reply leave before the process exits (as hardResetDelay).
const switchDelay = 750 * time.Millisecond

// charactersDesk builds the d2d backend's desk. roots are the folders d2d may keep saves under; leave is the d2d's Save
// and Exit WITHOUT re-entering (Cycler.LeaveGame would restart the old character). A nil leave switches off create and
// switch (the list, delete and restore still work).
func charactersDesk(ip paths.Instance, active func() string, roots []string, r *bot.Runner, leave func(context.Context) error) watch.CharactersDesk {
	m := &charsel.Manager{
		Store:      &charsel.Store{ArchiveDir: filepath.Join(ip.Dir(), "characters-archive"), StateRoot: filepath.Join(ip.Dir(), "characters")},
		Inst:       ip.Dir(),
		Active:     active,
		Measured:   os.Getenv("D2BOT_RUN") != "",
		Supervised: os.Getenv(superviseMarker) != "" && !underFleet(),
		Leave:      leave,
		Exit:       os.Exit,
		ExitCode:   exitSwitch,
		ExitDelay:  switchDelay,
		SaveDir:    saveFolder(roots, active),
	}
	if r != nil {
		m.Begin = func() { r.BeginHardReset() } // the runner must not rejoin the old character while we leave
	}
	return m
}

// saveFolder finds the folder that holds the .d2s files of the d2d under roots: the one with the active character's
// file, else the one with the most. It looks at most five levels down and re-reads at most every 10 s. U: the exact
// subfolder is d2d's (docs/charselect.md); the walk is the same one d2dSaveDir does for the saved exp.
func saveFolder(roots []string, active func() string) func() string {
	var (
		mu   sync.Mutex
		dir  string
		last time.Time
	)
	return func() string {
		mu.Lock()
		defer mu.Unlock()
		if dir != "" {
			if _, err := os.Stat(dir); err == nil && time.Since(last) < 10*time.Second {
				return dir
			}
		}
		last = time.Now()
		best, bestN, bestActive := "", 0, false
		for _, root := range roots {
			count := map[string]int{}
			hasActive := map[string]bool{}
			base := strings.Count(filepath.Clean(root), string(filepath.Separator))
			_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
				if err != nil {
					return nil
				}
				if d.IsDir() {
					if strings.Count(p, string(filepath.Separator))-base > 5 {
						return fs.SkipDir
					}
					return nil
				}
				if n, ok := strings.CutSuffix(d.Name(), ".d2s"); ok {
					dd := filepath.Dir(p)
					count[dd]++
					if a := active(); a != "" && strings.EqualFold(n, a) {
						hasActive[dd] = true
					}
				}
				return nil
			})
			for dd, n := range count {
				if (hasActive[dd] && !bestActive) || (hasActive[dd] == bestActive && n > bestN) {
					best, bestN, bestActive = dd, n, hasActive[dd]
				}
			}
		}
		dir = best
		return dir
	}
}

// switcher is the supervisor's side of a switch: it owns the argument list the children start with.
type switcher struct {
	o       options
	log     *slog.Logger
	cur     []string
	created string // a character the last switch created: its next start loads the save once it exists
	fresh   bool   // the next child starts a character without parked state: it gets --goals again
}

// args is the command line of attempt: childArgs of the current arguments, with the goals of a character that has
// none yet (a created one) and, for a created character that has been saved since, a load instead of a create.
func (w *switcher) args(attempt int, left time.Duration) []string {
	if w.created != "" && w.saved(w.created) {
		w.cur = setArg(w.cur, "--d2d-class", "-1")
		w.created = ""
	}
	a := childArgs(w.cur, attempt > 0, left, w.o.goalSpec != "")
	if w.fresh && attempt > 0 && w.o.goalSpec != "" {
		a = append(a, "--goals", w.o.goalSpec)
	}
	w.fresh = false
	return a
}

func (w *switcher) instDir() (string, bool) {
	l := w.o.layout
	if l.Root == "" {
		l = paths.Layout{Root: w.o.data}
	}
	ip, err := l.Instance(w.o.instance)
	return ip.Dir(), err == nil
}

func (w *switcher) roots(dir string) []string {
	r := []string{filepath.Join(dir, "d2d")}
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		r = append(r, x)
	}
	return r
}

func (w *switcher) saved(name string) bool {
	dir, ok := w.instDir()
	if !ok {
		return false
	}
	d := saveFolder(w.roots(dir), func() string { return name })()
	if d == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(d, name+".d2s"))
	return err == nil
}

// apply reads the intent the bot left, swaps the per-character state and points the arguments at the new character. It
// reports false (restart with the old arguments, state untouched) when there is no intent or the swap fails.
func (w *switcher) apply() bool {
	dir, ok := w.instDir()
	if !ok {
		return false
	}
	in, ok := charsel.ReadIntent(dir)
	if !ok {
		w.log.Error("supervise: exit code for a switch but no intent; restarting the old character")
		return false
	}
	defer charsel.ClearIntent(dir)
	_, noGoals := os.Stat(filepath.Join(dir, "characters", in.To, "goals.json"))
	// the active-char file goes first and is put back when the swap fails, so a container restart never loads a hero
	// whose state was not swapped in
	prev := charsel.ReadActive(dir)
	if err := charsel.WriteActive(dir, in.To); err != nil {
		w.log.Error("supervise: cannot record the active character, staying on the old one", "to", in.To, "err", err)
		return false
	}
	if err := charsel.SwapState(dir, in.From, in.To); err != nil {
		if prev != "" {
			_ = charsel.WriteActive(dir, prev)
		} else {
			_ = os.Remove(filepath.Join(dir, charsel.ActiveFile))
		}
		w.log.Error("supervise: character switch failed, staying on the old character", "from", in.From, "to", in.To, "err", err)
		return false
	}
	w.cur = setArg(setArg(w.cur, "--char", in.To), "--d2d-class", strconv.Itoa(in.ClassIndex))
	w.created, w.fresh = "", noGoals != nil
	if in.Create {
		w.created = in.To
	}
	w.log.Info("supervise: switching character", "from", in.From, "to", in.To, "create", in.Create)
	return true
}

// setArg sets a flag's value in an argument list, replacing `--name v` and `--name=v` and adding it when absent.
func setArg(args []string, name, val string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == name {
			i++
			continue
		}
		if strings.HasPrefix(a, name+"=") {
			continue
		}
		out = append(out, a)
	}
	return append(out, name, val)
}
