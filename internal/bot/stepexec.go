package bot

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	questspecdata "d2lod-bot/data/questspec"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/quest"
	"d2lod-bot/internal/questspec"
	"d2lod-bot/internal/task"
)

// The thin adapter of the generic step executor (internal/questspec/exec.go, knob quests.stepExec, card stepexec1).
// The engine and the step row are data and pure; this file only maps the four verbs onto the bot's capabilities and
// the typed outcome back onto the task layer's errors. First chunk on it: the Cain scroll (take_scroll).

// execObjectNames gives the object adapter's name for a class id (the Objects capability still clicks by name).
var execObjectNames = map[int]string{quest.ObjTreeInifuss: TreeObject}

var (
	execOnce  sync.Once
	execSteps map[string]questspec.Step
	execErr   error
)

// execStep returns the step row "<chain>/<step>" of the embedded Act 1 file.
func execStep(chain, step string) (questspec.Step, error) {
	execOnce.Do(func() {
		execSteps = map[string]questspec.Step{}
		raw, err := questspecdata.FS.ReadFile("act1.json")
		if err != nil {
			execErr = err
			return
		}
		cs, err := questspec.LoadFile(raw)
		if err != nil {
			execErr = err
			return
		}
		for _, c := range cs {
			for _, s := range c.Steps {
				execSteps[c.ID+"/"+s.ID] = s
			}
		}
	})
	if execErr != nil {
		return questspec.Step{}, execErr
	}
	s, ok := execSteps[chain+"/"+step]
	if !ok {
		return questspec.Step{}, fmt.Errorf("stepexec: no step %s/%s in the embedded chains", chain, step)
	}
	return s, nil
}

// execBackend implements questspec.Backend over the bot's capabilities.
type execBackend struct {
	a taskActs
}

// Approach: the object and dialog adapters walk to their target themselves, so the step has nothing to do here.
func (b *execBackend) Approach(context.Context, questspec.Target, int) error { return nil }

func (b *execBackend) Act(ctx context.Context, verb string, t questspec.Target) error {
	a := b.a
	td := a.r.d.Task
	switch verb {
	case "interact":
		name, ok := execObjectNames[t.Class]
		if !ok {
			return &questspec.FatalError{Err: fmt.Errorf("no object name for class %d", t.Class)}
		}
		if td == nil || td.Objects == nil {
			return &questspec.FatalError{Err: errors.New("no object interaction capability")}
		}
		return a.tag(td.Objects.Interact(ctx, name))
	case "pickup":
		if a.r.d.Loot == nil {
			return &questspec.FatalError{Err: errors.New("no looter")}
		}
		name, ok := questItemCodes[t.Code]
		if !ok {
			return &questspec.FatalError{Err: fmt.Errorf("no item name for code %q", t.Code)}
		}
		rep, err := a.r.d.Loot.PickUp(ctx, PickOpts{Radius: 30, Names: []string{name}})
		switch {
		case err != nil && task.Classify(a.tag(err)) != task.KindNotFound:
			return a.tag(err)
		case err != nil || rep.Picked == 0:
			return questspec.ErrNothing
		}
		return nil
	}
	return &questspec.FatalError{Err: fmt.Errorf("verb %q has no adapter yet", verb)}
}

// Holds reads the post from a fresh snapshot: has_item(code) is the pack, nothing else is known yet.
func (b *execBackend) Holds(c questspec.Clause) bool {
	r := b.a.r
	s, err := r.d.Snapshot()
	if err != nil || s == nil || s.State != d2.StateInGame {
		s = r.latest()
	}
	if s == nil {
		return false
	}
	held := map[string]int{}
	r.scanQuestItems(s, held)
	for _, p := range c {
		if p.Name == "has_item" && len(p.Args) == 1 && held[questItemCodes[p.Args[0]]] > 0 {
			return true
		}
	}
	return false
}

func (b *execBackend) Sleep(d time.Duration) { b.a.r.e.Clock.Sleep(d) }

// runExecStep runs one data step through the engine and maps the outcome to the task layer's errors: a refusal reads
// "not ready" (the typed refusal path of outcome.go classifies it), a fault is KindNotFound like the old handler.
func (a taskActs) runExecStep(ctx context.Context, chain, step string) error {
	s, err := execStep(chain, step)
	if err != nil {
		return task.Wrap(task.KindFatal, err)
	}
	o := questspec.RunStep(ctx, s, &execBackend{a: a})
	a.r.e.Log.Info("step executor", "step", chain+"/"+step, "outcome", o.Kind.String(), "acts", o.Attempts)
	switch o.Kind {
	case questspec.Done:
		return nil
	case questspec.Refused:
		return task.Errorf(task.KindDanger, "not ready (%s): %s", o.Need, o.Why)
	case questspec.Interrupted:
		return a.tag(o.Err)
	case questspec.Progress:
		return task.Errorf(task.KindNotFound, "step %s/%s in progress", chain, step)
	}
	return a.tag(task.Errorf(task.KindNotFound, "%v", o.Err))
}

// treeViaExecutor is the Cain scroll chunk on the executor: click the tree, take the scroll, wait for it in the pack.
func (a taskActs) treeViaExecutor(ctx context.Context) error {
	if err := a.runExecStep(ctx, "search_for_cain", "take_scroll"); err != nil {
		return err
	}
	a.setPseudo(interactPseudo[TreeObject])
	return nil
}
