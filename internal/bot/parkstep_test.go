package bot

import (
	"strings"
	"testing"
	"time"

	"d2lod-bot/internal/task"
)

func notFound(step string) task.Result {
	return task.Result{Outcome: task.OutRetry, StepID: step, Kind: task.KindNotFound,
		Err: task.Errorf(task.KindNotFound, "object class 30 not found")}
}

// A quest step that is not found twice is parked, the reason is published, a different step is not touched, and the
// park ends after parkMinutes. The rule reads only the step id and the knobs, so it is the same for every class and
// build; two groups below stand for the two quests of the log.
func TestParkStepAfterRepeatedBlocks(t *testing.T) {
	g := soakyRig(t, 1)
	r := g.r
	ex := task.NewExecutor(nil, nil)
	r.loop.task.ex = ex
	r.loop.task.stepTitle = "Take the Inifuss bark"
	tree, cellar := "quest.search_for_cain/tree", "quest.forgotten_tower/to-cellar"

	if r.parkStep(notFound(tree)) {
		t.Fatal("the first block only waits")
	}
	if r.parkStep(notFound(cellar)) {
		t.Fatal("a different step counts on its own")
	}
	if _, ok := ex.Blocked("forgotten_tower"); ok {
		t.Fatal("another step must not be parked")
	}
	if !r.parkStep(notFound(tree)) {
		t.Fatal("the second identical block parks the step")
	}
	if _, ok := ex.Blocked("search_for_cain"); !ok {
		t.Fatal("group not parked in the executor")
	}
	if w := r.journeyWhy; !strings.Contains(w, "Inifuss bark") || !strings.Contains(w, "parked 15 min") {
		t.Fatalf("journeyBlocked text: %q", w)
	}
	r.journeyOK() // a farm step ran: the reason stays while parked
	if r.journeyWhy == "" {
		t.Fatal("the reason must outlive a farm step")
	}
	g.clock.Sleep(14 * time.Minute)
	r.unparkDue(ex)
	if _, ok := ex.Blocked("search_for_cain"); !ok {
		t.Fatal("unparked too early")
	}
	g.clock.Sleep(2 * time.Minute)
	r.unparkDue(ex)
	if _, ok := ex.Blocked("search_for_cain"); ok {
		t.Fatal("still parked after parkMinutes")
	}
	r.journeyOK()
	if r.journeyWhy != "" {
		t.Fatalf("reason not cleared: %q", r.journeyWhy)
	}
	// non-quest steps and other kinds are never parked
	if r.parkStep(notFound("level.14")) || r.parkStep(notFound("level.14")) {
		t.Fatal("a level step is not a quest step")
	}
	r.e.Tune.Bot.ParkAfter = 0
	if r.parkStep(notFound(tree)) || r.parkStep(notFound(tree)) {
		t.Fatal("parkAfter 0 turns the rule off")
	}
}
