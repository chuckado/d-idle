package main

import (
	"os"
	"path/filepath"
	"time"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/watch"
)

// The player's Reset command (docs/reset.md). Soft: the runner leaves the game through its Cycler and rejoins. Hard (and
// the fallback of a backend with no Cycler): the process exits with exitStorm exactly like the watchdog's "runner loop
// made no progress" path, so whatever restarts d2bot (--supervise, the container's restart policy) brings it back, and
// the goals, orders, pause and Loop come back from their state files. A measured run (D2BOT_RUN) refuses both.

// hardResetDelay lets the HTTP reply leave before the process exits.
const hardResetDelay = 750 * time.Millisecond

func resetDesk(r *bot.Runner, stateDir string) watch.ResetDesk {
	if r == nil {
		return nil
	}
	if os.Getenv("D2BOT_RUN") == "" && stateDir != "" {
		r.SetResetPath(filepath.Join(stateDir, "playerreset.json"))
	}
	return &resetAdapter{r: r, exit: os.Exit, measured: os.Getenv("D2BOT_RUN") != "", hardOff: os.Getenv("D2BOT_HARD_RESET") == "0"}
}

// resetRunner is the part of bot.Runner the adapter uses.
type resetRunner interface {
	ResetView() bot.ResetView
	CanSoftReset() bool
	PlayerReset(force bool) bot.ResetView
	BeginHardReset() bot.ResetView
}

type resetAdapter struct {
	r        resetRunner
	exit     func(int)
	measured bool
	hardOff  bool
}

func (a *resetAdapter) View() any { return a.r.ResetView() }

func (a *resetAdapter) Set(hard, force bool) (any, *watch.ResetRefusal) {
	if a.measured {
		return nil, &watch.ResetRefusal{Code: "measured_run", Error: "a measured run is never reset (D2BOT_RUN is set)"}
	}
	if v := a.r.ResetView(); v.State == bot.ResetResetting || v.State == bot.ResetRejoining {
		v.Already = true
		return v, nil
	}
	if !hard && a.r.CanSoftReset() {
		return a.r.PlayerReset(force), nil
	}
	if a.hardOff {
		return nil, &watch.ResetRefusal{Code: "hard_off", Error: "restarting the bot is switched off (D2BOT_HARD_RESET=0)"}
	}
	v := a.r.BeginHardReset()
	time.AfterFunc(hardResetDelay, func() { a.exit(exitStorm) })
	return v, nil
}

var _ watch.ResetDesk = (*resetAdapter)(nil)
