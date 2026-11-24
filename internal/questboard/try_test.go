package questboard

import (
	"errors"
	"strings"
	"testing"
)

const (
	burial = "sisters_burial_grounds"
	raven  = "blood_raven"
)

// The owner's try anyway (ADR 0030 amended 2026-11-24): a readiness refusal ("too hard for now", a floor) is skipped for
// ONE attempt of one chunk; the plain accept still refuses; the try ends when the run is stopped or the hero dies. Both
// builds' floors differ (legacy floors by class family), the override does not. Must fail without AcceptTry's CanStart bypass.
func TestTryAnywayStartsOneNotReadyChunk(t *testing.T) {
	for _, class := range smClasses {
		t.Run(class, func(t *testing.T) {
			r := newRigSM(t, class, 1, true)
			st, _ := r.x.State()
			var view ChunkView
			for _, v := range st.Views {
				if v.ID == burial {
					view = v.Chunks[0]
				}
			}
			if view.State != Available || view.Why == "" {
				t.Skipf("the fake hero does not leave %s available: %+v", burial, view)
			}
			if err := r.x.Accept(burial, raven); err == nil || !strings.Contains(err.Error(), "not ready") {
				t.Fatalf("the plain accept must refuse a not ready chunk: %v", err)
			}
			if r.x.TryKey() != "" {
				t.Fatal("a refused accept set the override")
			}
			refused, err := r.x.AcceptTry(burial, raven)
			if err != nil || refused == "" {
				t.Fatalf("try anyway: %q %v", refused, err)
			}
			if r.x.TryKey() != ChunkKey(burial, raven) {
				t.Fatalf("TryKey %q", r.x.TryKey())
			}
			if d := r.next(); d.Kind != DecRun || d.Chunk != raven || r.x.TryKey() == "" {
				t.Fatalf("the try chunk must run: %+v key %q", d, r.x.TryKey())
			}
			// a death ends the override, the run itself stays
			d := r.next()
			r.x.Report(d, StepResult{Outcome: StepDied, Why: "died"})
			if r.x.TryKey() != "" {
				t.Fatal("a death must end the override")
			}
			// one shot: stopped, the next plain accept refuses again, and so does a try only when it is not Available
			if err := r.x.Cancel(burial, raven); err != nil {
				t.Fatal(err)
			}
			r.next()
			if err := r.x.Accept(burial, raven); err == nil {
				t.Fatal("the override outlived the run")
			}
		})
	}
}

// Only a readiness refusal is overridable: a locked or a done part is refused with ErrNotOverridable, and a restart ends
// the override (memory only).
func TestTryAnywayRefusesWhatIsNotReadiness(t *testing.T) {
	for _, class := range smClasses {
		t.Run(class, func(t *testing.T) {
			r := newRigSM(t, class, 1, true)
			var no ErrNotOverridable
			if _, err := r.x.AcceptTry("search_for_cain", "stones"); !errors.As(err, &no) {
				t.Fatalf("a later part of a quest is locked, not overridable: %v", err)
			}
			if _, err := r.x.AcceptTry(burial, raven); err == nil {
				r.restart()
				if r.x.TryKey() != "" {
					t.Fatal("a restart must end the override")
				}
			}
		})
	}
}
