package ether

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Earner turns the bot's observed events into Earn calls (V7b, docs/ether.md "Earn call sites"). Every call is
// idempotent through the wallet's keys, so a restart or a replayed journal books nothing twice; the only state
// here is the unpaid kill count of this process, which a restart loses on purpose (the trickle just starts over).
type Earner struct {
	W   *Wallet
	Log func(string) // one line per failure; nil: silent

	mu    sync.Mutex
	kills int
}

// Hero names who an event happened to. Difficulty is the wallet's word: normal, nm or hell.
type Hero struct {
	Name       string
	Difficulty string
}

func (e *Earner) earn(source, sub, diff string) {
	if _, err := e.W.Earn(source, sub, diff, 0); err != nil && e.Log != nil {
		e.Log("ether: " + err.Error())
	}
}

// First books "a first for this hero": kind is "monster" or "area", name what was met. The key holds the hero, so a
// second hero earns its own firsts.
func (e *Earner) First(h Hero, kind, name string) {
	if h.Name == "" || name == "" {
		return
	}
	e.earn("first", h.Name+"/"+kind+"/"+name, "")
}

// Kill books a boss kill (once per boss, difficulty and account), the hero's first of that monster and, every
// `per` kills, a step of the daily trickle.
func (e *Earner) Kill(h Hero, monster string, boss bool) {
	if monster == "" {
		return
	}
	e.First(h, "monster", monster)
	if boss && difficulties[h.Difficulty] {
		e.earn("boss", monster, h.Difficulty)
	}
	src, ok := e.W.opt.Sources.find("kills")
	if !ok || src.Per <= 0 {
		return
	}
	e.mu.Lock()
	e.kills++
	due := e.kills >= src.Per
	if due {
		e.kills = 0
	}
	e.mu.Unlock()
	if due {
		if _, err := e.W.Trickle("kills"); err != nil && e.Log != nil {
			e.Log("ether: " + err.Error())
		}
	}
}

// Trickle books the next free batch of today for a daily capped source: key <source>/<day>/<k>. The wallet's own
// daily cap decides when there is nothing left (Capped).
func (w *Wallet) Trickle(source string) (Result, error) {
	src, ok := w.opt.Sources.find(source)
	if !ok {
		return Result{}, fmt.Errorf("%w %q", ErrUnknownSource, source)
	}
	if src.DailyCap <= 0 || src.Amount <= 0 {
		return Result{}, fmt.Errorf("%w: %s is not a capped trickle", ErrBadCredit, source)
	}
	day := w.opt.Now().Format("2006-01-02")
	for k := 1; k <= src.DailyCap/src.Amount+1; k++ {
		r, err := w.Earn(source, day+"/"+strconv.Itoa(k), "", 0)
		if err != nil || !r.Already {
			return r, err
		}
	}
	return Result{Capped: true}, nil
}

// Level books every clvl row of the milestone table the level has reached, and the difficulty rows the hero has
// reached (diff_nm, diff_hell). Account-wide, once each.
func (e *Earner) Level(h Hero, level int) {
	src, ok := e.W.opt.Sources.find("milestone")
	if !ok {
		return
	}
	for id := range src.Rows {
		if n, ok := strings.CutPrefix(id, "clvl"); ok {
			if t, err := strconv.Atoi(n); err == nil && level >= t {
				e.earn("milestone", id, "")
			}
		}
	}
	e.Reached(h)
}

// Reached books the difficulty rows the hero's difficulty has reached (nightmare counts for normal's row too).
func (e *Earner) Reached(h Hero) {
	if h.Difficulty == "nm" || h.Difficulty == "hell" {
		e.earn("milestone", "diff_nm", "")
	}
	if h.Difficulty == "hell" {
		e.earn("milestone", "diff_hell", "")
	}
}

// Paid reports what a credit key paid (the full key, for example quest/den_of_evil/-/normal); false when it is not booked.
func (w *Wallet) Paid(key string) (int, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	i, ok := w.keys[key]
	if !ok || i < 0 || i >= len(w.entries) {
		return 0, false
	}
	return w.entries[i].Amount, true
}
