package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/vault"
	"d2lod-bot/internal/vaultkeep"
	"d2lod-bot/internal/watch"
)

type stubTools struct {
	charges int
	free    bool
}

func (s stubTools) Charges(context.Context) (int, error)      { return s.charges, nil }
func (s stubTools) Free() bool                                { return s.free }
func (s stubTools) Identify(context.Context, int, bool) error { return nil }
func (stubTools) Gold(context.Context) (int, error)           { return 100, nil }
func (stubTools) Sell(context.Context, []int) error           { return nil }

// The page's Reveal and Sell face over a real Keeper: the price before the click, the no-scroll state, free at Cain, the
// refusal reasons, and the opt-in prompt (off by default, a Needs-you row with Yes and Not now when on and short).
func TestActFaceViewAndPrompt(t *testing.T) {
	for _, class := range []string{"sorceress", "paladin"} {
		t.Run(class, func(t *testing.T) {
			root := t.TempDir()
			home := vaultHomeFor(root)
			st, err := home.Open(class)
			if err != nil {
				t.Fatal(err)
			}
			add := func(i, q int, ident bool) string {
				r, err := st.Add(vault.Item{Raw: fmt.Sprintf("%08x", i+1), Code: "rin", Name: fmt.Sprintf("ring %d", i), Quality: q, Ilvl: 10 + i, Identified: ident, Value: 30 + i})
				if err != nil {
					t.Fatal(err)
				}
				return r.ID
			}
			unid, rare, pinned := add(1, 4, false), add(2, 6, true), add(3, 4, true)
			_ = st.Pin(pinned)
			tools := &stubTools{charges: 2}
			short, prompt := 0, 0
			k := &vaultkeep.Keeper{Store: func() *vault.Store { return st }, API: d2dctl.NewVaultFake(class), Enabled: func() bool { return true },
				Tools: tools, Vendor: tools, ShortBy: func() int { return short }, Prompt: func() bool { return prompt != 0 }}
			home.mu.Lock()
			home.keep = k
			home.mu.Unlock()
			activeKeeper.Store(k)
			f := actFace{home}
			ctx := context.Background()

			v := f.Act(ctx)
			if !v.Reveal.Enabled || v.Reveal.Charges != 2 || v.Reveal.PerItem != 1 || v.Reveal.Free || v.Prompt != nil {
				t.Fatalf("view %+v", v)
			}
			tools.charges = 0
			if v = f.Act(ctx); v.Reveal.Enabled || v.Reveal.Code != "noscroll" || v.Reveal.Why != vaultkeep.ReasonNoScroll {
				t.Fatalf("no scroll: %+v", v.Reveal)
			}
			tools.free = true
			if v = f.Act(ctx); !v.Reveal.Enabled || !v.Reveal.Free {
				t.Fatalf("at Cain the price is zero: %+v", v.Reveal)
			}
			if res := f.EnqueueReveal(ctx, []string{unid, rare}, false); !res[0].OK || res[1].OK || res[1].Code != "state" {
				t.Fatalf("reveal results %+v", res)
			}
			if res := f.EnqueueSell(ctx, []string{pinned}, false); res[0].OK || res[0].Code != "pinned" {
				t.Fatalf("pinned sale %+v", res)
			}

			// the prompt: off by default; on and short it lists the unpinned low item, never the rare or the pinned one
			short = 90
			if f.Act(ctx).Prompt != nil || needIDs(t)["vault:sellprompt"] {
				t.Fatal("prompt must be off by default")
			}
			prompt = 1
			p := f.Act(ctx).Prompt
			if p == nil || len(p.IDs) != 1 || p.IDs[0] != unid || p.Short != 90 {
				t.Fatalf("prompt %+v", p)
			}
			if !needIDs(t)["vault:sellprompt"] {
				t.Fatal("no Needs-you row for the prompt")
			}
			f.DismissOffer()
			if f.Act(ctx).Prompt != nil || needIDs(t)["vault:sellprompt"] {
				t.Fatal("dismissed prompt came back")
			}
		})
	}
}

func needIDs(t *testing.T) map[string]bool {
	t.Helper()
	var src watch.GlanceSources
	for _, ex := range glanceExtra {
		ex(&src, glanceDeps{})
	}
	out := map[string]bool{}
	if src.Needs != nil {
		for _, n := range src.Needs() {
			out[n.ID] = true
			if n.ID == "vault:sellprompt" {
				if len(n.Actions) != 2 || n.Actions[0].Post != watch.VaultSellRoute || !json.Valid(n.Actions[0].Body) || n.Actions[1].Post != watch.VaultPromptRoute {
					t.Errorf("prompt actions %+v", n.Actions)
				}
			}
		}
	}
	return out
}
