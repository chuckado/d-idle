package vaultkeep

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"d2lod-bot/internal/bulkid"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/vault"
)

type etherWallet struct {
	bal       int
	spent     map[string]int
	failSpend bool
}

func (w *etherWallet) Balance() (int, error) { return w.bal, nil }
func (w *etherWallet) Spend(key string, n int) error {
	if w.failSpend {
		return errors.New("disk")
	}
	if _, ok := w.spent[key]; ok {
		return nil
	}
	if n > w.bal {
		return errors.New("short")
	}
	w.bal -= n
	w.spent[key] = n
	return nil
}

// etherRig has NO scroll and no tome charge: an Ether reveal must not need one.
func etherRig(t *testing.T, class string, bal, price int) (*rig, *etherWallet) {
	r := newRig(t, class)
	r.withTools(0)
	w := &etherWallet{bal: bal, spent: map[string]int{}}
	r.k.Ether = &bulkid.EtherPayer{W: w, Price: func() int { return price }}
	return r, w
}

// addSealed records an unidentified item whose affixes the vault holds (never shown until paid).
func (r *rig) addSealed(i int, code string) string {
	r.t.Helper()
	rec := d2dctl.FakeRecord(i, code, 4, 30+i, fmt.Sprintf("roll%d", i))
	it, err := r.st.Add(vault.Item{Raw: rec.Norm, Code: code, Name: "Unidentified " + code, Quality: 4, Ilvl: 30 + i, Value: 40 + i, Sealed: "7:0:20;93:0:10"})
	if err != nil {
		r.t.Fatal(err)
	}
	return it.ID
}

func (r *rig) ether(ids ...string) []Result {
	return r.k.EnqueueRevealEther(context.Background(), ids, false)
}

// Ether replaces the scroll: no charge, no game step, the unlock is immediate, paid once, the affixes appear only now.
func TestEtherReplacesScroll(t *testing.T) {
	for _, class := range classes {
		r, w := etherRig(t, class, 10, 4)
		id := r.addSealed(1, "wsp")
		if rec := r.state(id); rec.Identified || rec.Props != "" || rec.Sealed == "" {
			t.Fatalf("%s: affixes exposed before payment: %+v", class, rec)
		}
		if res := r.ether(id); !res[0].OK {
			t.Fatalf("%s: %+v", class, res[0])
		}
		rec := r.state(id)
		if !rec.Identified || rec.Props != "7:0:20;93:0:10" || rec.Sealed != "" || !rec.Unlocked || r.k.Pending() != 0 {
			t.Fatalf("%s: %+v", class, rec)
		}
		if w.bal != 6 || w.spent["reveal:"+id] != 4 || !r.heroEmpty() {
			t.Fatalf("%s: wallet %+v hero empty %v", class, w, r.heroEmpty())
		}
	}
}

// No double charge: the same tap again, or a second tap after the first, is a no-op.
func TestEtherIdempotent(t *testing.T) {
	r, w := etherRig(t, classes[0], 10, 4)
	id := r.addSealed(1, "wsp")
	r.ether(id)
	res := r.ether(id, id)
	if !res[0].OK || !res[0].Already || !res[1].Already || w.bal != 6 || len(w.spent) != 1 {
		t.Fatalf("%+v wallet %+v", res, w)
	}
}

// A short wallet refuses the whole press: nothing unlocked, nothing spent. Price 0 is off.
func TestEtherShortOrOffRefuses(t *testing.T) {
	for _, class := range classes {
		r, w := etherRig(t, class, 7, 4)
		a, b := r.addSealed(1, "wsp"), r.addSealed(2, "wsp")
		res := r.ether(a, b)
		if res[0].OK || res[1].OK || res[0].Code != "ether" || w.bal != 7 || r.state(a).Identified || r.state(b).Identified {
			t.Fatalf("%s short: %+v", class, res)
		}
		if res := r.ether(a); !res[0].OK || w.bal != 3 {
			t.Fatalf("%s one fits: %+v %+v", class, res, w)
		}
		r2, w2 := etherRig(t, class, 100, 0)
		id := r2.addSealed(1, "wsp")
		if res := r2.ether(id); res[0].OK || res[0].Code != "etheroff" || r2.state(id).Identified || w2.bal != 100 {
			t.Fatalf("%s off: %+v", class, res[0])
		}
	}
}

// A failed charge undoes the unlock: the record stays sealed and nothing is spent.
func TestEtherFailedChargeKeepsSealed(t *testing.T) {
	r, w := etherRig(t, classes[1], 10, 4)
	id := r.addSealed(1, "wsp")
	w.failSpend = true
	res := r.ether(id)
	if res[0].OK || res[0].Code != "ether" {
		t.Fatalf("%+v", res[0])
	}
	if rec := r.state(id); rec.Identified || rec.Props != "" || rec.Sealed == "" || rec.Unlocked || w.bal != 10 {
		t.Fatalf("%+v wallet %+v", rec, w)
	}
	w.failSpend = false
	if res := r.ether(id); !res[0].OK || w.bal != 6 {
		t.Fatalf("retry: %+v %+v", res[0], w)
	}
}

// An item the vault holds no affixes for cannot be unlocked: refused, nothing spent. An identified item costs nothing.
func TestEtherNothingSealedOrIdentified(t *testing.T) {
	r, w := etherRig(t, classes[0], 10, 4)
	plain := r.addUnid(1, "wsp")
	if res := r.ether(plain); res[0].OK || res[0].Code != "nosealed" || w.bal != 10 {
		t.Fatalf("%+v", res[0])
	}
	id := r.addSealed(2, "wsp")
	r.ether(id)
	r.st.Unseal(id) // already identified: refused by the store too
	if err := r.st.Unseal(id); !errors.Is(err, vault.ErrState) {
		t.Fatalf("%v", err)
	}
	if err := r.st.Unseal(plain); !errors.Is(err, vault.ErrNothingSealed) {
		t.Fatalf("%v", err)
	}
	if res := r.ether("v-nope"); res[0].Code != "missing" || w.bal != 6 {
		t.Fatalf("%+v", res[0])
	}
}

// Cain stands here: the reveal is the free normal one and no Ether is charged, even on an Ether tap.
func TestEtherTapAtCainIsFree(t *testing.T) {
	for _, class := range classes {
		r, w := etherRig(t, class, 10, 4)
		r.tools.cain = true
		id := r.addSealed(1, "wsp")
		if res := r.ether(id); !res[0].OK {
			t.Fatalf("%+v", res[0])
		}
		r.run()
		if !r.state(id).Identified || w.bal != 10 || len(w.spent) != 0 {
			t.Fatalf("%s: %+v wallet %+v", class, r.state(id), w)
		}
	}
}

// The scroll path is unchanged and never touches the wallet; with neither a scroll nor Ether it refuses clearly.
func TestPlainRevealNeverSpendsEther(t *testing.T) {
	r, w := etherRig(t, classes[0], 10, 4)
	r.tools.charges = 1
	id := r.addUnid(1, "wsp")
	r.reveal(id)
	r.run()
	if len(w.spent) != 0 || w.bal != 10 || !r.state(id).Identified {
		t.Fatalf("plain reveal: %+v", w)
	}
	r.tools.charges = 0
	other := r.addUnid(2, "wsp")
	if res := r.reveal(other); res[0].OK || res[0].Code != "noscroll" {
		t.Fatalf("%+v", res[0])
	}
}
