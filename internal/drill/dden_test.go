package drill

import (
	"context"
	"strings"
	"testing"
)

type denFake struct {
	FieldLab
	area     int
	bounceAt string
}

func (d *denFake) Survive(context.Context, *Run) error         { return nil }
func (d *denFake) Goto(_ context.Context, _ *Run, a int) error { d.area = a; return nil }
func (d *denFake) Tick(context.Context, int) error             { return nil }
func (d *denFake) ToArea(_ context.Context, a int) (int, string) {
	d.area = a
	return a, ""
}
func (d *denFake) DenPhase(_ context.Context, ph string, _ int) (int, []string, string) {
	if ph == d.bounceAt {
		d.area = 2
		return 2, []string{"nav walk callers=a < b"}, ""
	}
	return d.area, nil, ""
}

func TestDenDrillNamesTheBouncingPhase(t *testing.T) {
	f := &denFake{bounceAt: "explore"}
	o := fDen(context.Background(), nil, f, f, 2, 8)
	if o.Verdict != Fail || !strings.Contains(o.Observed, "callers=a < b") || !strings.Contains(o.Reason, "explore") {
		t.Fatalf("%v %q %q", o.Verdict, o.Reason, o.Observed)
	}
	f = &denFake{}
	if o := fDen(context.Background(), nil, f, f, 2, 8); o.Verdict != Pass {
		t.Fatalf("%v %q", o.Verdict, o.Reason)
	}
}
