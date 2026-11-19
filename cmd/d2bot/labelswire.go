package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/watch"
)

// The item labels toggle (docs/video-feed.md, "Item labels"): the watch page's Labels button asks the fork's `labels`
// verb to draw the ground item names. The wish is kept per hero next to the other instance state (labels.json) because
// the game forgets it when its process restarts; the desk applies it when the bot enters a game (Prepare) and on every
// GET, bounded and idempotent. Default off. The bot itself never reads the picture on d2d (it reads `ground` and `dump`),
// so the labels change nothing it does.

const labelsTimeout = 3 * time.Second

type labelsClient interface {
	Labels(ctx context.Context, mode string) (bool, error)
}

type labelsDesk struct {
	c    labelsClient
	path string
	hero func() string
	mu   sync.Mutex
	pref map[string]bool
}

func newLabelsDesk(c labelsClient, stateDir string, hero func() string) *labelsDesk {
	d := &labelsDesk{c: c, hero: hero, pref: map[string]bool{}}
	if stateDir != "" {
		d.path = filepath.Join(stateDir, "labels.json")
		if b, err := os.ReadFile(d.path); err == nil {
			_ = json.Unmarshal(b, &d.pref)
		}
	}
	return d
}

func (d *labelsDesk) want() bool { return d.pref[d.hero()] }

// fail turns a client error into the reply's code: unsupported (this d2d lacks the verb or runs without --item-verbs),
// refused (the game said no) or game (not reachable).
func labelsFail(err error) watch.LabelsState {
	var ce *d2dctl.CmdError
	switch {
	case errors.As(err, &ce) && (strings.HasPrefix(ce.Msg, "lab verbs off") || strings.Contains(ce.Msg, "unknown verb")):
		return watch.LabelsState{Code: "unsupported", Reason: "this game build cannot show item labels"}
	case errors.As(err, &ce):
		return watch.LabelsState{Available: true, Code: "refused", Reason: ce.Msg}
	}
	return watch.LabelsState{Available: true, Code: "game", Reason: "the game is not reachable"}
}

// Get reads the state back and, when the game differs from the hero's wish (a restarted d2d, another hero), applies the
// wish once and reads again.
func (d *labelsDesk) Get() watch.LabelsState {
	ctx, cancel := context.WithTimeout(context.Background(), labelsTimeout)
	defer cancel()
	return d.reconcile(ctx)
}

func (d *labelsDesk) reconcile(ctx context.Context) watch.LabelsState {
	d.mu.Lock()
	defer d.mu.Unlock()
	on, err := d.c.Labels(ctx, "status")
	if err != nil {
		return labelsFail(err)
	}
	if w := d.want(); on != w {
		mode := "off"
		if w {
			mode = "on"
		}
		if on, err = d.c.Labels(ctx, mode); err != nil {
			return labelsFail(err)
		}
	}
	return watch.LabelsState{OK: true, On: on, Available: true}
}

// Set records the wish, applies it and answers what the game reports.
func (d *labelsDesk) Set(on bool) watch.LabelsState {
	ctx, cancel := context.WithTimeout(context.Background(), labelsTimeout)
	defer cancel()
	d.mu.Lock()
	defer d.mu.Unlock()
	mode := "off"
	if on {
		mode = "on"
	}
	got, err := d.c.Labels(ctx, mode)
	if err != nil {
		return labelsFail(err)
	}
	d.pref[d.hero()] = on
	d.save()
	st := watch.LabelsState{OK: got == on, On: got, Available: true}
	if !st.OK {
		st.Code, st.Reason = "refused", "the game did not take it"
	}
	return st
}

func (d *labelsDesk) save() {
	if d.path == "" {
		return
	}
	b, _ := json.Marshal(d.pref)
	tmp := d.path + ".tmp"
	if os.WriteFile(tmp, b, 0o644) == nil {
		_ = os.Rename(tmp, d.path)
	}
}

// Apply is the Prepare hook: it puts the game in the hero's wish when the bot enters a game. It never fails the caller.
func (d *labelsDesk) Apply(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, labelsTimeout)
	defer cancel()
	d.reconcile(ctx)
	return nil
}
