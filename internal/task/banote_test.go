package task

import "testing"

// cainfix3 F1: the refusal for a banned field area names the deaths and the end time when the bot supplies them.
func TestFieldReadyBanNote(t *testing.T) {
	w := World{Banned: map[Area]bool{4: true}, BanNote: map[Area]string{4: "3 deaths (until 14:05)"}}
	if f := fieldReady(4)(w); f == nil || f.Kind != KindDanger || f.Why != "not ready: area banned after 3 deaths (until 14:05)" {
		t.Fatalf("fault %+v", f)
	}
	w.BanNote = nil
	if f := fieldReady(4)(w); f == nil || f.Why != "not ready: area banned after repeated deaths" {
		t.Fatalf("fault %+v", f)
	}
}
