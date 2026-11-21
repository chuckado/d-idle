package main

import (
	"os"

	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/watch"
)

// recallDesk returns the desk for /api/recall, or nil in a measured run (D2BOT_RUN: a command that walks the hero is not
// part of a measured run) or without a runner.
func recallDesk(r *bot.Runner) watch.RecallDesk {
	if r == nil || os.Getenv("D2BOT_RUN") != "" {
		return nil
	}
	return recallAdapter{r}
}

type recallAdapter struct{ r *bot.Runner }

func (a recallAdapter) View() any   { return a.r.RecallView() }
func (a recallAdapter) Recall() any { return a.r.RecallNow() }
func (a recallAdapter) Cancel() any { return a.r.RecallCancel() }

var _ watch.RecallDesk = recallAdapter{}
