package task

import "testing"

// The stones step needs the Key to the Cairn Stones (V lab drill L1: without it every click is ignored): a prereq fault
// with the reason when it is missing, none with it, none once the portal is open or Tristram is reached.
func TestCainStonesNeedTheKey(t *testing.T) {
	var step Step
	for _, n := range cainNodes() {
		if n.Step.ID == id(QCain, "stones") {
			step = n.Step
		}
	}
	if step.Ready == nil {
		t.Fatal("the stones step has no readiness check")
	}
	w := World{Area: StonyField, Items: map[string]int{}}
	if f := step.Ready(w); f == nil || f.Kind != KindPrereq || f.Why == "" {
		t.Fatalf("without the key: %+v", f)
	}
	w.Items[ItemCainSc] = 1
	if f := step.Ready(w); f != nil {
		t.Fatalf("with the key: %+v", f)
	}
	w.Items = map[string]int{ItemPortal: 1}
	if f := step.Ready(w); f != nil {
		t.Fatalf("portal open, key consumed: %+v", f)
	}
	w = World{Area: Tristram}
	if need, _ := step.Need(w); need {
		t.Fatal("the stones step is still needed in Tristram")
	}
}
