package talk

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/env"
	"d2lod-bot/internal/quest"
	"d2lod-bot/internal/task"
	"d2lod-bot/internal/town"
)

// Services is the part of *town.Town the dialog layer drives (the game I/O): open an NPC by class, click a bottom-anchored
// line, and the two item pickers that follow a Talk. *town.Town implements it; tests use a fake.
type Services interface {
	OpenForTest(ctx context.Context, class int) (d2.Unit, error)
	PickLine(ctx context.Context, npc d2.Unit, idx, n int) error
	Socket(ctx context.Context, w town.SocketWant) (town.ServiceReport, error)
	Imbue(ctx context.Context, w town.ImbueWant) (town.ServiceReport, error)
}

// townDialog is the quest.Dialog over the Townsman: open the NPC (walk, click, wait out the greeting), click bottom-
// anchored lines, close.
type townDialog struct {
	t   Services
	e   *env.Env
	npc d2.Unit
}

func (d *townDialog) Talk(ctx context.Context, class int) error {
	n, err := d.t.OpenForTest(ctx, class)
	d.npc = n
	return err
}
func (d *townDialog) Choose(ctx context.Context, row, of int) error {
	err := d.t.PickLine(ctx, d.npc, row, of)
	d.e.Clock.Sleep(700 * time.Millisecond) // the next menu (or the speech) comes up
	return err
}
func (d *townDialog) Close(ctx context.Context) error { return d.e.In.CloseAll() }

// Adapter talks to town NPCs for the task layer. Option paths are rows of quest.Menus: Akara's "Talk" then
// "Den of Evil" is the only one seen live (docs/saves.md: it wrote the Den's quest word); the others are U and a
// wrong row shows up as a failed step (needs_attention), never as a guessed click.
type Adapter struct {
	T      Services
	E      *env.Env
	Status func(quest string) quest.Status // row visibility; nil assumes every conditional row is shown
}

// akaraRespecSeen: Akara's main menu has the fourth line "Reset Stat/Skill Points" once she paid the Den reward
// (the save word that proves it lags until Save and Exit, so the process remembers).
var akaraRespecSeen atomic.Bool

// Talk runs one dialog path. The Den reward claim retries with the other main-menu line count when the skill point
// does not come (a click one line off opens the shop or hits the title) after resetting the dialog.
func (d Adapter) Talk(ctx context.Context, npc, pick string) error {
	// The item pickers (town/imbue.go, U live): Larzuk's Talk is part of the socket flow; Charsi first takes the Malus
	// with the plain Talk, then the Imbue line and the item.
	switch {
	case npc == task.NPCLarzuk && pick == "sockets":
		_, err := d.T.Socket(ctx, town.SocketWant{})
		return err
	case npc == task.NPCCharsi && pick == "imbue:best":
		if err := d.talkOnce(ctx, npc, pick, false); err != nil {
			return err
		}
		_, err := d.T.Imbue(ctx, town.ImbueWant{})
		return err
	}
	if npc != task.NPCAkara || pick != "reward" {
		return d.talkOnce(ctx, npc, pick, false)
	}
	var err error
	for try := 0; try < 3 && ctx.Err() == nil; try++ {
		flip := try == 1 // second attempt: assume the other line count
		if err = d.talkOnce(ctx, npc, pick, flip); err == nil {
			akaraRespecSeen.Store(true)
			return nil
		}
		d.E.Log.Warn("akara reward not paid, retrying", "try", try+1, "err", err)
	}
	return err
}

func (d Adapter) talkOnce(ctx context.Context, npc, pick string, flipRespec bool) error {
	act := 0
	if s, err := d.E.Snapshot(); err == nil && s.Player != nil {
		act = s.Player.Act + 1
	}
	class, ok := ClassIn(npc, act)
	if !ok {
		return fmt.Errorf("no NPC class known for %s", npc)
	}
	path, ok := Paths[[2]string{npc, pick}]
	if pick == "" {
		path, ok = []string{"Talk"}, true
	}
	if !ok {
		return fmt.Errorf("dialog option %q of %s is not implemented", pick, npc)
	}
	var before int
	if s, err := d.E.Snapshot(); err == nil && s.Player != nil {
		before = s.Player.SkillPoints
	}
	began := d.E.Clock.Now()
	d.E.Log.Info("npc: talk", "npc", npc, "pick", pick, "class", class, "path", path)
	dlg := &townDialog{t: d.T, e: d.E}
	if err := dlg.Talk(ctx, class); err != nil {
		_ = dlg.Close(ctx)
		return fmt.Errorf("open %s: %w", npc, err)
	}
	// Rows with a quest condition appear and disappear with the quest states (the line count places the click).
	base := d.Status
	if base == nil {
		base = func(string) quest.Status { return quest.Started }
	}
	respec := akaraRespecSeen.Load() != flipRespec
	st := func(q string) quest.Status {
		switch q {
		case quest.RespecRow:
			if respec {
				return quest.Started
			}
			return quest.Needed
		case "den_of_evil":
			// The save lags and the reward is pending in memory: a reward claim never opens the topic list.
			if pick == "reward" {
				return quest.Started
			}
		}
		return base(q)
	}
	err := quest.PickPath(ctx, dlg, class, st, path...)
	// the speech: wait for the reward (a skill point) up to 8 s when we expect one, else a short pause
	deadline := d.E.Clock.Now().Add(2500 * time.Millisecond)
	if npc == task.NPCAkara && pick == "reward" {
		deadline = d.E.Clock.Now().Add(8 * time.Second)
	}
	for d.E.Clock.Now().Before(deadline) && ctx.Err() == nil {
		if s, serr := d.E.Snapshot(); serr == nil && s.Player != nil && s.Player.SkillPoints > before {
			break
		}
		d.E.Clock.Sleep(300 * time.Millisecond)
	}
	_ = dlg.Close(ctx)
	d.E.Log.Info("npc visit: talk done", "npc", npc, "path", path, "ms", d.E.Clock.Now().Sub(began).Milliseconds(), "err", err)
	if err != nil {
		return fmt.Errorf("talk to %s %v: %w", npc, path, err)
	}
	// The Den reward is a skill point: the one claim whose effect shows in memory at once (docs/task.md L1).
	if npc == task.NPCAkara && pick == "reward" {
		s, serr := d.E.Snapshot()
		d.E.Log.Info("akara talk", "skillPointsBefore", before, "after", func() int {
			if serr == nil && s.Player != nil {
				return s.Player.SkillPoints
			}
			return -1
		}())
		if serr != nil || s.Player == nil || s.Player.SkillPoints <= before {
			return fmt.Errorf("no skill point after talking to Akara (quest not complete, or the dialog missed)")
		}
	}
	return nil
}

// RespecSeen reports whether this process has seen Akara pay the Den reward; SetRespecSeen sets it (tests, drills).
func RespecSeen() bool     { return akaraRespecSeen.Load() }
func SetRespecSeen(v bool) { akaraRespecSeen.Store(v) }
