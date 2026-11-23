package main

import (
	"os"

	"d2lod-bot/internal/ops"
	"d2lod-bot/internal/questboard"
)

// startOperator opens the operator socket next to the quest plan files when D2BOT_OPERATOR=1 (never in a measured
// run, never by default). It reuses the desk the viewer routes use; failures only log. docs/operator.md.
func startOperator(stateDir string, d *questboard.Desk, logf func(msg string, kv ...any)) {
	if d == nil || stateDir == "" {
		return
	}
	_, err := ops.Start(stateDir, func() ops.Desk { return d }, os.Getenv, logf)
	if err != nil {
		logf("operator control failed to start", "err", err)
	}
}
