package vaultkeep

import (
	"context"
	"fmt"
	"testing"

	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/vault"
)

// noDecode hides the Decoder of the fake: a fork without `item decode` (an older build).
type noDecode struct{ d2dctl.VaultAPI }

// addBare records an unidentified item that holds NO affixes (the fork sent none) but whose own bytes carry them: the fake
// answers `item decode` for its norm hex with props.
func (r *rig) addBare(i int, code, props string) string {
	r.t.Helper()
	rec := d2dctl.FakeRecord(i, code, 4, 30+i, fmt.Sprintf("bare%d", i))
	it, err := r.st.Add(vault.Item{Raw: rec.Norm, Code: code, Name: "Unidentified " + code, Quality: 4, Ilvl: 30 + i, Value: 40 + i})
	if err != nil {
		r.t.Fatal(err)
	}
	dec := rec
	dec.Props = props
	if r.api.VaultFake.Bytes == nil {
		r.api.VaultFake.Bytes = map[string]d2dctl.VaultRecord{}
	}
	r.api.VaultFake.Bytes[rec.Norm] = dec
	return it.ID
}

// A Sealed-less record is revealable for gold: the order decodes its own bytes first, in the same call, with no scroll, no
// town visit and nothing created in the hero.
func TestGoldRevealDecodesASealedLessRecord(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(0)
			r.withPay(7200, 80)
			var ids []string
			for i := 1; i <= 4; i++ {
				ids = append(ids, r.addBare(i, "wsp", "7:0:20;93:0:10"))
			}
			if r.state(ids[0]).Sealed != "" {
				t.Fatal("the record starts with nothing held")
			}
			res, v := r.gold(ids, 0, false)
			for _, x := range res {
				if !x.OK {
					t.Fatalf("refused: %+v", x)
				}
			}
			if v == nil || v.State != OrderRevealed || v.Done != 4 || v.Paid != 320 {
				t.Fatalf("view %+v", v)
			}
			for _, id := range ids {
				if rec := r.state(id); rec.Props != "7:0:20;93:0:10" || rec.Sealed != "" || !rec.Unlocked || !rec.Identified {
					t.Fatalf("revealed from the decoded bytes: %+v", rec)
				}
			}
			f := r.api.VaultFake
			if f.Decodes != 4 || f.Stats.Creates != 0 || f.Stats.Vaults != 0 || !r.heroEmpty() {
				t.Fatalf("a read: decodes %d, stats %+v, hero empty %v", f.Decodes, f.Stats, r.heroEmpty())
			}
			if r.tools.charges != 0 || r.k.Pending() != 0 || r.k.GoldPending() {
				t.Fatal("no scroll step is planned and nothing waits for a town visit")
			}
			if lines := r.run(); len(lines) != 0 {
				t.Fatalf("a town visit finds nothing to do: %v", lines)
			}
		})
	}
}

// A fork without the verb, or bytes it cannot decode, keep the refusal: nosealed, nothing is charged.
func TestGoldRevealWithoutDecodeStaysRefused(t *testing.T) {
	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			r := newRig(t, class)
			r.withTools(0)
			r.withPay(7200, 80)
			id := r.addBare(1, "wsp", "7:0:20")
			r.k.API = noDecode{chaosAPI{r.api, r.ch}}
			res, v := r.gold([]string{id}, 0, false)
			if v != nil || res[0].OK || res[0].Code != "nosealed" || len(r.pay.Calls) != 0 {
				t.Fatalf("an older fork: %+v view %+v calls %v", res[0], v, r.pay.Calls)
			}
			r.k.API = chaosAPI{r.api, r.ch}
			delete(r.api.VaultFake.Bytes, r.state(id).Raw) // the fork does not know these bytes
			if res, _ := r.gold([]string{id}, 0, false); res[0].Code != "nosealed" || len(r.pay.Calls) != 0 {
				t.Fatalf("undecodable bytes: %+v", res[0])
			}
		})
	}
}

// FillSealed never decodes a record twice and never touches an identified or already sealed one.
func TestFillSealedIsIdempotent(t *testing.T) {
	r := newRig(t, "Paladin")
	bare := r.addBare(1, "wsp", "7:0:20")
	held := r.addSealed(2, "wsp")
	if n := r.k.FillSealed(context.Background(), []string{bare, held, "nope"}); n != 1 {
		t.Fatalf("filled %d", n)
	}
	if n := r.k.FillSealed(context.Background(), []string{bare, held}); n != 0 || r.api.VaultFake.Decodes != 1 {
		t.Fatalf("again: filled %d, decodes %d", n, r.api.VaultFake.Decodes)
	}
	if rec := r.state(bare); rec.Sealed != "7:0:20" || rec.Identified || rec.Props != "" {
		t.Fatalf("held, never shown: %+v", rec)
	}
}
