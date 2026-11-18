package main

import (
	"os"
	"path/filepath"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/watch"
)

// The player's Pause command (docs/pause.md): the runner holds the request, this file hands it to the watch route. A
// measured run (D2BOT_RUN) never installs it, so the route answers 503 "off" there.

// pauseDesk returns the desk for GET and POST /api/pause, or nil in a measured run or without a runner. The request is
// persisted next to the other state so a restart stays paused.
func pauseDesk(r *bot.Runner, stateDir string) watch.PauseDesk {
	if r == nil || os.Getenv("D2BOT_RUN") != "" {
		return nil
	}
	if stateDir != "" {
		r.SetPausePath(filepath.Join(stateDir, "playerpause.json"))
	}
	return pauseAdapter{r}
}

type pauseAdapter struct{ r *bot.Runner }

func (a pauseAdapter) View() any       { return a.r.PlayerPauseView() }
func (a pauseAdapter) Set(on bool) any { return a.r.PlayerPause(on) }
func (a pauseAdapter) SetMode(on bool, freeze *bool) any {
	return a.r.PlayerPauseMode(on, freeze)
}

var (
	_ watch.PauseDesk     = pauseAdapter{}
	_ watch.PauseModeDesk = pauseAdapter{}
)
