package vaultkeep

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/bulkid"
)

type etherWallet struct {
	bal   int
	spent map[string]int
}

func (w *etherWallet) Balance() (int, error) { return w.bal, nil }
func (w *etherWallet) Spend(key string, n int) error {
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

func etherRig(t *testing.T, class string, bal, price int) (*rig, *etherWallet) {
	r := newRig(t, class)
	r.withTools(3)
	w := &etherWallet{bal: bal, spent: map[string]int{}}
	r.k.Ether = &bulkid.EtherPayer{W: w, Price: func() int { return price }}
	return r, w
}

// An explicit Ether tap queues, and the Ether leaves only after the verified reveal, once, with the record in the key.
func TestEtherChargedAfterVerifiedReveal(t *testing.T) {
	for _, class := range classes {
		r, w := etherRig(t, class, 10, 4)
		id := r.addUnid(1, "wsp")
		if res := r.k.EnqueueRevealEther(context.Background(), []string{id}, false); !res[0].OK {
			t.Fatalf("%s: %+v", class, res[0])
		}
		if w.bal != 10 {
			t.Fatalf("%s: Ether spent before the reveal", class)
		}
		r.run()
		if !r.state(id).Identified || w.bal != 6 || w.spent["reveal:"+id] != 4 {
			t.Fatalf("%s: identified %v wallet %+v", class, r.state(id).Identified, w)
		}
	}
}

// Must-fail: a short wallet refuses the whole press, queues nothing and spends nothing; price 0 is off.
func TestEtherShortOrOffRefuses(t *testing.T) {
	for _, class := range classes {
		r, w := etherRig(t, class, 3, 4)
		id := r.addUnid(1, "wsp")
		res := r.k.EnqueueRevealEther(context.Background(), []string{id}, false)
		if res[0].OK || res[0].Code != "ether" || w.bal != 3 || r.k.Pending() != 0 {
			t.Fatalf("%s short: %+v pending %d", class, res[0], r.k.Pending())
		}
		r2, w2 := etherRig(t, class, 100, 0)
		id2 := r2.addUnid(1, "wsp")
		if res := r2.k.EnqueueRevealEther(context.Background(), []string{id2}, false); res[0].OK || res[0].Code != "etheroff" || r2.k.Pending() != 0 || w2.bal != 100 {
			t.Fatalf("%s off: %+v", class, res[0])
		}
	}
}

// Must-fail: a plain tap (no Ether) and the free Cain path never touch the wallet.
func TestPlainRevealNeverSpendsEther(t *testing.T) {
	r, w := etherRig(t, classes[0], 10, 4)
	id := r.addUnid(1, "wsp")
	r.reveal(id)
	r.run()
	if len(w.spent) != 0 || w.bal != 10 {
		t.Fatalf("plain reveal spent Ether: %+v", w)
	}
}
