package d2dctl

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"d2lod-bot/internal/cheat"
)

// KindD2DDebug is the cheat ledger kind of every d2d "debug" verb that changes the game (docs/cheats.md).
const KindD2DDebug cheat.Kind = "d2d-debug"

// SkillDebug sends the lab fork's skill-drill verbs (docs/forkdbg.md): debug mana full, debug states set|clear,
// debug mon set, debug cast. They exist only in the private lab binary and refuse unless the game is offline
// (--no-save). Each call is a cheat or a test action: it is written to the ledger BEFORE it is sent (an empty reason
// or an unwritable ledger stops it, nothing is sent) and completed with the reply. Nothing else in this package
// may send these verbs.
type SkillDebug struct {
	C         *Client
	Ledger    *cheat.Ledger
	Actor     string // the drill or agent; required
	Instance  string
	Character string
}

var skillDebugSeq atomic.Int64

// MonEdit says what `debug mon set` changes on one monster; nil fields stay as they are. Life, MaxLife are points,
// LifeFrac the 256ths already taken off life (0..255), Dmg the monster's attack range, Res absolute resistances by
// name (phys, magic, fire, light, cold, poison; -100..100).
type MonEdit struct {
	Life, LifeFrac, MaxLife *int
	Dmg                     *[2]int
	Res                     map[string]int
}

var monResNames = map[string]bool{"phys": true, "magic": true, "fire": true, "light": true, "cold": true, "poison": true}

// ManaFull refills the player's mana to its maximum. The reply is "mana <a> -> <b> (max <m>)".
func (d *SkillDebug) ManaFull(ctx context.Context, reason string) (string, error) {
	return d.run(ctx, "mana full", reason, false)
}

// StatesSet puts the state of skill (Frozen Armor is 40, States.txt row 10) on the player as a cast starts it. level
// 0 keeps the skill's own level (at least 1); otherwise 1..99.
func (d *SkillDebug) StatesSet(ctx context.Context, skill, level int, reason string) (string, error) {
	if skill < 0 || skill > 1000 || level < 0 || level > 99 {
		return "", errors.New("d2dctl: states set: skill 0..1000, level 0..99")
	}
	line := fmt.Sprintf("states set %d", skill)
	if level > 0 {
		line += fmt.Sprintf(" %d", level)
	}
	return d.run(ctx, line, reason, false)
}

// StatesClear removes the state of skill from the player; skill -1 removes all of them.
func (d *SkillDebug) StatesClear(ctx context.Context, skill int, reason string) (string, error) {
	if skill < -1 || skill > 1000 {
		return "", errors.New("d2dctl: states clear: skill -1 (all) or 0..1000")
	}
	if skill < 0 {
		return d.run(ctx, "states clear all", reason, false)
	}
	return d.run(ctx, fmt.Sprintf("states clear %d", skill), reason, false)
}

// MonSet edits one live monster (its unit id, the Unit of Monsters()).
func (d *SkillDebug) MonSet(ctx context.Context, unit int, e MonEdit, reason string) (string, error) {
	if unit < 0 {
		return "", errors.New("d2dctl: mon set: unit id >= 0")
	}
	parts := []string{fmt.Sprintf("mon set %d", unit)}
	rng := func(name string, p *int, lo, hi int) error {
		if p == nil {
			return nil
		}
		if *p < lo || *p > hi {
			return fmt.Errorf("d2dctl: mon set: %s %d..%d", name, lo, hi)
		}
		parts = append(parts, fmt.Sprintf("%s %d", name, *p))
		return nil
	}
	for _, f := range []struct {
		name   string
		p      *int
		lo, hi int
	}{{"life", e.Life, 1, 1000000}, {"frac", e.LifeFrac, 0, 255}, {"maxlife", e.MaxLife, 1, 1000000}} {
		if err := rng(f.name, f.p, f.lo, f.hi); err != nil {
			return "", err
		}
	}
	if e.Dmg != nil {
		if e.Dmg[0] < 0 || e.Dmg[1] > 100000 || e.Dmg[0] > e.Dmg[1] {
			return "", errors.New("d2dctl: mon set: dmg 0..100000, lo <= hi")
		}
		parts = append(parts, fmt.Sprintf("dmg %d %d", e.Dmg[0], e.Dmg[1]))
	}
	names := make([]string, 0, len(e.Res))
	for n := range e.Res {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if !monResNames[n] || e.Res[n] < -100 || e.Res[n] > 100 {
			return "", fmt.Errorf("d2dctl: mon set: res %q %d: phys|magic|fire|light|cold|poison, -100..100", n, e.Res[n])
		}
		parts = append(parts, fmt.Sprintf("res %s %d", n, e.Res[n]))
	}
	if len(parts) == 1 {
		return "", errors.New("d2dctl: mon set: nothing to change")
	}
	return d.run(ctx, strings.Join(parts, " "), reason, false)
}

// CastAt makes the player cast skill at the cell (x, y) through the normal skill path (a test action; the cast is
// queued and starts at the next tick). The skill needs a level (Client.Points).
func (d *SkillDebug) CastAt(ctx context.Context, skill int, x, y float64, reason string) (string, error) {
	if skill < 0 || skill > 1000 || x < -4096 || x > 4096 || y < -4096 || y > 4096 {
		return "", errors.New("d2dctl: cast: skill 0..1000, |x| |y| <= 4096")
	}
	return d.run(ctx, fmt.Sprintf("cast %d %s %s", skill, strconv.FormatFloat(x, 'f', -1, 64), strconv.FormatFloat(y, 'f', -1, 64)), reason, true)
}

// CastAtUnit casts skill at the monster with that unit id.
func (d *SkillDebug) CastAtUnit(ctx context.Context, skill, unit int, reason string) (string, error) {
	if skill < 0 || skill > 1000 || unit < 0 {
		return "", errors.New("d2dctl: cast: skill 0..1000, unit >= 0")
	}
	return d.run(ctx, fmt.Sprintf("cast %d unit %d", skill, unit), reason, true)
}

// CastNear casts skill at the nearest targetable monster.
func (d *SkillDebug) CastNear(ctx context.Context, skill int, reason string) (string, error) {
	if skill < 0 || skill > 1000 {
		return "", errors.New("d2dctl: cast: skill 0..1000")
	}
	return d.run(ctx, fmt.Sprintf("cast %d near", skill), reason, true)
}

// run ledgers, sends, completes the entry. test marks a test action (a cast) rather than a state change.
func (d *SkillDebug) run(ctx context.Context, sub, reason string, test bool) (string, error) {
	if d.C == nil || d.Ledger == nil || d.Actor == "" {
		return "", errors.New("d2dctl: SkillDebug needs a client, a ledger and an actor")
	}
	if strings.TrimSpace(reason) == "" {
		return "", errors.New("d2dctl: a debug verb needs a reason for the cheat ledger")
	}
	verb := "debug " + sub
	id := fmt.Sprintf("%s-%d", time.Now().UTC().Format("20060102T150405"), skillDebugSeq.Add(1))
	params := map[string]any{"id": id, "verb": verb}
	if test {
		params["test_action"] = true
	}
	e := cheat.Entry{Time: time.Now().UTC(), Actor: d.Actor, Instance: d.Instance, Character: d.Character, Kind: KindD2DDebug, Params: params, Reason: reason}
	if err := d.Ledger.Append(e); err != nil { // before anything is sent
		return "", err
	}
	r, err := d.C.Do(ctx, verb)
	out := ""
	if r != nil {
		out = strings.TrimSpace(strings.Join(append(append([]string{}, r.Lines...), r.Rest), "\n"))
	}
	e.Time, e.Success, e.After = time.Now().UTC(), err == nil, map[string]any{"reply": out}
	if err != nil {
		e.Error = err.Error()
	}
	if lerr := d.Ledger.Append(e); lerr != nil && err == nil {
		err = lerr
	}
	return out, err
}
