package vaultkeep

import (
	"context"
	"testing"

	"d2lod-bot/internal/bulkid"
	"d2lod-bot/internal/env"
)

type rescueTools struct {
	chaosTools
	rescued bool
}

func (t rescueTools) Rescued() bool { return t.rescued }

// The viewer note appears only while identify.policy counts on Cain (1 or 2) and he is not rescued.
func TestRevealNoteWhileCainNotRescued(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(3)
			base := r.k.Tools.(chaosTools)
			for _, c := range []struct {
				policy  int
				rescued bool
				want    string
			}{
				{env.IdentifyCainFree, false, bulkid.WhyNoCain},
				{env.IdentifyScrolls, false, bulkid.WhyNoCain},
				{env.IdentifyCainFree, true, ""},
				{env.IdentifyManual, false, ""},
				{env.IdentifyNever, false, ""},
			} {
				r.k.Tools = rescueTools{base, c.rescued}
				p := c.policy
				r.k.IdentifyPolicy = func() int { return p }
				if got := r.k.RevealPrice(context.Background()).Note; got != c.want {
					t.Errorf("policy %d rescued %v: note %q want %q", c.policy, c.rescued, got, c.want)
				}
			}
			r.k.IdentifyPolicy = nil
			r.k.Tools = rescueTools{base, false}
			if got := r.k.RevealPrice(context.Background()).Note; got != "" {
				t.Errorf("no policy hook: note %q", got)
			}
		})
	}
}
