package drill

import (
	"context"
	"testing"
)

type holdStub struct {
	FieldLab
	h ArmorHold
}

func (s holdStub) ArmorHold(context.Context, *Run, int) ArmorHold { return s.h }

func TestD29Verdicts(t *testing.T) {
	base := ArmorHold{Binary: "x", SimMS: 300000, Casts: 2, Samples: 50, PollMS: 2000}
	cases := []struct {
		name string
		mod  func(*ArmorHold)
		want Verdict
	}{
		{"no cut", func(h *ArmorHold) {}, Fail},
		{"cut and retry", func(h *ArmorHold) { h.Cuts, h.CutRetries = 1, 1 }, Pass},
		{"cut no retry", func(h *ArmorHold) { h.Cuts = 1 }, Fail},
		{"absent too long", func(h *ArmorHold) { h.Cuts, h.CutRetries, h.MaxDownMS = 1, 1, 5000 }, Fail},
		{"single poll absence", func(h *ArmorHold) { h.Cuts, h.CutRetries, h.MaxDownMS, h.DownMS = 1, 1, 2000, 2000 }, Pass},
		{"no recast", func(h *ArmorHold) { h.Casts = 1 }, Fail},
	}
	for _, c := range cases {
		h := base
		c.mod(&h)
		ds := D29(holdStub{h: h})
		o := ds[0].Do(context.Background(), &Run{})
		if o.Verdict != c.want {
			t.Errorf("%s: verdict %v (%s), want %v", c.name, o.Verdict, o.Reason, c.want)
		}
	}
}
