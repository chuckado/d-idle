package bot

import (
	"context"
	"d2lod-bot/internal/d2"
	"errors"
	"strings"
	"testing"

	"d2lod-bot/internal/task"
)

type fakeObjects struct{ clicked []string }

func (o *fakeObjects) Interact(_ context.Context, name string) error {
	o.clicked = append(o.clicked, name)
	return nil
}

func TestTaskActsCubeAndMercNeedTheirCapability(t *testing.T) {
	g := newRig(t)
	acts := taskActs{r: g.r}
	g.r.d.Task = &TaskDeps{}
	if err := acts.Transmute(context.Background(), []string{"a"}); task.Classify(err) != task.KindPrereq {
		t.Fatalf("cube without capability: %v", err)
	}
	if err := acts.HireMerc(context.Background(), 3, "Cold"); task.Classify(err) != task.KindPrereq {
		t.Fatalf("merc without capability: %v", err)
	}
	var gotIn []string
	var gotAct int
	var gotAura string
	g.r.d.Task = &TaskDeps{
		Cube: func(_ context.Context, in []string) error { gotIn = in; return nil },
		HireMerc: func(_ context.Context, act int, aura string) error {
			gotAct, gotAura = act, aura
			return errors.New("no gold")
		},
	}
	if err := acts.Transmute(context.Background(), []string{task.ItemFlail, task.ItemEye}); err != nil || len(gotIn) != 2 {
		t.Fatalf("cube: %v %v", err, gotIn)
	}
	if err := acts.HireMerc(context.Background(), 3, "Cold"); err == nil || gotAct != 3 || gotAura != "Cold" {
		t.Fatalf("merc: %v %d %q", err, gotAct, gotAura)
	}
}

// An object click that proves a pseudo item sets it for the World; a new game forgets it.
func TestInteractSetsPseudoItems(t *testing.T) {
	g := newRig(t)
	objs := &fakeObjects{}
	g.r.d.Task = &TaskDeps{Objects: objs}
	acts := taskActs{r: g.r}
	if _, err := (taskEnv{g.r}).World(context.Background()); err != nil { // the first poll counts the game
		t.Fatal(err)
	}
	if err := acts.Interact(context.Background(), "tal-orifice"); err != nil {
		t.Fatal(err)
	}
	w, err := taskEnv{g.r}.World(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if w.Items[task.ItemStaffPlaced] != 1 {
		t.Fatalf("staff placed pseudo item missing: %v", w.Items)
	}
	_ = acts.Interact(context.Background(), "Horadric Cube chest")
	if w2, _ := (taskEnv{g.r}).World(context.Background()); w2.Items[task.ItemBookRead] != 0 {
		t.Fatal("an unrelated click must not set a pseudo item")
	}
	g.r.obMu.Lock()
	g.r.ob.newGames++
	g.r.obMu.Unlock()
	if w3, _ := (taskEnv{g.r}).World(context.Background()); w3.Items[task.ItemStaffPlaced] != 0 {
		t.Fatal("a new game must forget the pseudo item")
	}
}

// SmashOrb without a town that can swap weapons is a prereq fault and sets nothing; the orb click goes through Interact.
func TestSmashOrbNeedsWeaponSwap(t *testing.T) {
	g := newRig(t)
	objs := &fakeObjects{}
	g.r.d.Task = &TaskDeps{Objects: objs}
	err := taskActs{r: g.r}.SmashOrb(context.Background(), task.ItemWill)
	if task.Classify(err) != task.KindPrereq || len(objs.clicked) != 0 {
		t.Fatalf("got %v clicked %v", err, objs.clicked)
	}
	if w, _ := (taskEnv{g.r}).World(context.Background()); w.Items[task.ItemOrbSmashed] != 0 {
		t.Fatal("a failed smash must not set the pseudo item")
	}
}

func TestWorldReportsMercenaryPseudoItem(t *testing.T) {
	g := newRig(t)
	g.r.d.Task = &TaskDeps{}
	w, err := taskEnv{g.r}.World(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if w.HasMerc != (w.Items[task.ItemMercenary] == 1) {
		t.Fatalf("HasMerc %v but Items %v", w.HasMerc, w.Items)
	}
}

// Reading the scroll and touching Anya prove their pseudo items; the forge needs a town that swaps weapons.
func TestAct5PseudoItemsAndForge(t *testing.T) {
	g := newRig(t)
	objs := &fakeObjects{}
	g.r.d.Task = &TaskDeps{Objects: objs}
	acts := taskActs{r: g.r}
	ctx := context.Background()
	if _, err := (taskEnv{g.r}).World(ctx); err != nil {
		t.Fatal(err)
	}
	if err := acts.UseItem(ctx, task.ItemResScroll); err != nil {
		t.Fatal(err)
	}
	if objs.clicked[0] != "item:"+task.ItemResScroll {
		t.Fatalf("clicked %v", objs.clicked)
	}
	if err := acts.Interact(ctx, AnyaObject); err != nil { // no potion in the rig's pack: a touch
		t.Fatal(err)
	}
	w, _ := taskEnv{g.r}.World(ctx)
	if w.Items[task.ItemScrollRead] != 1 || w.Items[task.ItemAnyaTouched] != 1 || w.Items[task.ItemAnyaThawed] != 0 {
		t.Fatalf("pseudo items %v", w.Items)
	}
	objs.clicked = nil
	err := acts.UseForge(ctx, task.ItemSoulstone, task.ItemHammer)
	if task.Classify(err) != task.KindPrereq || len(objs.clicked) != 1 || objs.clicked[0] != ForgePlaceObject {
		t.Fatalf("got %v clicked %v", err, objs.clicked)
	}
	if w, _ := (taskEnv{g.r}).World(ctx); w.Items[task.ItemForgeUsed] != 0 {
		t.Fatal("a failed forge must not set the pseudo item")
	}
}

// The Inifuss tree click only drops the scroll: the adapter picks it up, and the cairn stones prove the portal pseudo item.
func TestTreePicksUpTheScrollAndStonesSetThePortal(t *testing.T) {
	g := newRig(t)
	objs := &fakeObjects{}
	g.r.d.Task = &TaskDeps{Objects: objs}
	g.r.d.Loot = &dropLoot{fakeLoot: fakeLoot{w: g.w}, w: g.w}
	acts := taskActs{r: g.r}
	ctx := context.Background()
	if _, err := (taskEnv{g.r}).World(ctx); err != nil {
		t.Fatal(err)
	}
	if err := acts.Interact(ctx, TreeObject); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(g.w.log(), "pickup") {
		t.Fatal("no pick-up after the tree")
	}
	if err := acts.Interact(ctx, "cairn-stones"); err != nil {
		t.Fatal(err)
	}
	if w, _ := (taskEnv{g.r}).World(ctx); w.Items[task.ItemPortal] != 1 {
		t.Fatalf("portal pseudo item missing: %v", w.Items)
	}
}

// questObjects is a fake object adapter for the cage: a click flips the quest word of Search for Cain to "reward
// pending" (bit 1, V drill L2: 0x0010 to 0x2012) unless it is told to ignore clicks.
type cageObjects struct {
	q       *fakeQuests
	ignore  bool
	clicked int
}

func (o *cageObjects) Interact(_ context.Context, name string) error {
	if name != CageObject {
		return nil
	}
	o.clicked++
	if !o.ignore {
		o.q.set(task.QCain, task.QuestState{Stage: task.StageCleared, Ev: task.EvMemory})
	}
	return nil
}

func cageRig(t *testing.T, cls d2.Class, objs *cageObjects, q *fakeQuests) *rig {
	g := newRig(t)
	g.w.set(func(w *world) { w.area = 38; w.class = cls })
	g.r.d.Task = &TaskDeps{Objects: objs, Quests: q, QuestsLive: true}
	return g
}

// The cage never opens while a guard lives: the archers stand 2 to 3 cells from the arrival and killed a level 4 hero
// (V drill L2). A fighter that cannot kill them (unreachable, set aside) ends in a Danger fault with no click. Run for a
// melee and a caster class: the step has no class branch, the Fighter decides how.
func TestCageNotOpenedWhileGuardsLive(t *testing.T) {
	for _, cls := range []d2.Class{d2.Sorceress, d2.Paladin} {
		q := &fakeQuests{}
		objs := &cageObjects{q: q}
		g := cageRig(t, cls, objs, q)
		g.fight.clearNothing = true
		for i := 0; i < 3; i++ {
			g.w.addMonster(uint32(20+i), 10, 11+i, 12) // three champions within the clear radius
		}
		err := taskActs{r: g.r}.Interact(context.Background(), CageObject)
		if task.Classify(err) != task.KindDanger || objs.clicked != 0 {
			t.Fatalf("class %v: err %v kind %v clicks %d, want a Danger fault and no click", cls, err, task.Classify(err), objs.clicked)
		}
	}
}

// Guards killed by the normal fighter, then one click, and success is read from the quest word, not from the click.
func TestCageFreedFromTheQuestWord(t *testing.T) {
	for _, cls := range []d2.Class{d2.Sorceress, d2.Paladin} {
		q := &fakeQuests{}
		q.set(task.QCain, task.QuestState{Stage: task.StageStarted, Ev: task.EvMemory})
		objs := &cageObjects{q: q}
		g := cageRig(t, cls, objs, q)
		g.w.addMonster(20, 10, 12, 12)
		g.w.addMonster(21, 10, 14, 12)
		if err := (taskActs{r: g.r}).Interact(context.Background(), CageObject); err != nil {
			t.Fatalf("class %v: %v", cls, err)
		}
		if objs.clicked != 1 || g.w.count("clear") == 0 {
			t.Fatalf("class %v: clicks %d, clear calls %d", cls, objs.clicked, g.w.count("clear"))
		}
	}
}

// A click that does not change the word is not success: the step retries a bounded number of times, then fails.
func TestCageClickWithoutTheWordFails(t *testing.T) {
	q := &fakeQuests{}
	objs := &cageObjects{q: q, ignore: true}
	g := cageRig(t, d2.Sorceress, objs, q)
	err := taskActs{r: g.r}.Interact(context.Background(), CageObject)
	if err == nil || objs.clicked != cageClicks {
		t.Fatalf("err %v clicks %d, want %d clicks and an error", err, objs.clicked, cageClicks)
	}
}

// Without a live quest reader (save only) the click is taken as done: the word is only known after a save.
func TestCageWithoutLiveReaderTakesTheClick(t *testing.T) {
	q := &fakeQuests{}
	objs := &cageObjects{q: q, ignore: true}
	g := cageRig(t, d2.Sorceress, objs, q)
	g.r.d.Task.QuestsLive = false
	if err := (taskActs{r: g.r}).Interact(context.Background(), CageObject); err != nil || objs.clicked != 1 {
		t.Fatalf("err %v clicks %d", err, objs.clicked)
	}
}
