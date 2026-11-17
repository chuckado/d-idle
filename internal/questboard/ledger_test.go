package questboard

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// memLedger is the wallet's rules in memory: a key pays once; fail makes the write fail (a crash before the append).
type memLedger struct {
	mu   sync.Mutex
	paid map[string]int
	fail int
}

func newMemLedger() *memLedger { return &memLedger{paid: map[string]int{}} }

func key(quest, chunk, diff string) string {
	if chunk == "" {
		chunk = "-"
	}
	return "quest/" + quest + "/" + chunk + "/" + diff
}

func (m *memLedger) Earn(quest, chunk, diff string, amount int) (LedgerResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := key(quest, chunk, diff)
	if m.fail > 0 {
		m.fail--
		return LedgerResult{}, errors.New("disk full")
	}
	if _, dup := m.paid[k]; dup {
		return LedgerResult{Key: k, Already: true}, nil
	}
	m.paid[k] = amount
	return LedgerResult{Key: k, Credited: amount}, nil
}

func (m *memLedger) Paid(quest, chunk, diff string) (int, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n, ok := m.paid[key(quest, chunk, diff)]
	return n, ok
}

func (m *memLedger) total() int {
	n := 0
	for _, v := range m.paid {
		n += v
	}
	return n
}

// handedIn runs the Den of Evil chunk to the hand-in and returns the result.
func (r *rig) handIn() HandInResult {
	r.t.Helper()
	if err := r.x.Accept(den, clr); err != nil {
		r.t.Fatal(err)
	}
	r.tick(4)
	res, err := r.x.HandIn(context.Background(), den, clr)
	if err != nil {
		r.t.Fatal(err)
	}
	return res
}

func TestEtherAwardOnceAndRepeat(t *testing.T) {
	for _, bd := range builds {
		t.Run(bd.class, func(t *testing.T) {
			r := newRig(t, bd.class, 6)
			r.led = newMemLedger()
			res := r.handIn()
			if res.Outcome != HandInDone || len(res.Awarded) != 1 || res.Awarded[0].Amount != 2 || res.Awarded[0].Key != "quest/den_of_evil/-/normal" || res.Awarded[0].Kind != KindEther {
				t.Fatalf("awarded: %+v", res)
			}
			again, err := r.x.HandIn(context.Background(), den, clr)
			if err != nil || !again.Already || len(again.Awarded) != 0 {
				t.Fatalf("repeat: %+v %v", again, err)
			}
			// a restart and the sweep book nothing twice
			r.restart()
			if got := r.x.Settle(); len(got) != 0 || r.led.(*memLedger).total() != 2 {
				t.Fatalf("settle after restart: %+v total %d", got, r.led.(*memLedger).total())
			}
		})
	}
}

// The process dies (or the write fails) after the evidence says done and before the ledger has the key: the next
// look at the evidence pays it, once.
func TestEtherCrashBetweenEvidenceAndLedger(t *testing.T) {
	r := newRig(t, "Paladin", 6)
	m := newMemLedger()
	r.led = m
	m.fail = 1
	res := r.handIn()
	if res.Outcome != HandInDone || len(res.Awarded) != 0 || m.total() != 0 {
		t.Fatalf("write failed: %+v total %d", res, m.total())
	}
	r.restart()
	d := r.desk(nil)
	reply := d.HandIn(den, "") // the evidence says done: this is the settle
	if !reply.OK || reply.Already || len(reply.Awarded) != 1 || m.total() != 2 {
		t.Fatalf("settle: %+v total %d", reply, m.total())
	}
	if again := d.HandIn(den, ""); !again.Already || len(again.Awarded) != 0 || m.total() != 2 {
		t.Fatalf("repeat: %+v", again)
	}
}

func TestEtherNoBackfillAndHandPlayed(t *testing.T) {
	for _, bd := range builds {
		t.Run(bd.class, func(t *testing.T) {
			// done before the board first sees the hero: baseline, never pays
			r := newRig(t, bd.class, 6)
			m := newMemLedger()
			r.led = m
			r.f.DoneQuest[den] = TierSave
			r.f.DoneQuest["sisters_burial_grounds"] = TierSave
			if got := r.x.Settle(); len(got) != 0 || m.total() != 0 {
				t.Fatalf("backfill: %+v", got)
			}
			// the baseline is in the plan file: a restart does not forget it
			r.restart()
			if got := r.x.Settle(); len(got) != 0 {
				t.Fatalf("after restart: %+v", got)
			}
			// done by hand after the board has seen the hero: never accepted, never pays
			r.f.DoneQuest["forgotten_tower"] = TierSave
			if got := r.x.Settle(); len(got) != 0 || m.total() != 0 {
				t.Fatalf("hand-played: %+v", got)
			}
		})
	}
}

func TestEtherSkipNeverAwards(t *testing.T) {
	r := newRig(t, "Sorceress", 6)
	m := newMemLedger()
	r.led = m
	if err := r.x.st.Skip(den, "", 0); err != nil {
		t.Fatal(err)
	}
	r.f.DoneQuest[den] = TierSave // done later by hand
	if got := r.x.Settle(); len(got) != 0 || m.total() != 0 {
		t.Fatalf("skip: %+v", got)
	}
}

func TestEtherDifficultyScaling(t *testing.T) {
	for i, want := range []int{2, 3, 5} {
		r := newRig(t, "Paladin", 40)
		r.f.DiffIdx = i
		m := newMemLedger()
		r.led = m
		if res := r.handIn(); len(res.Awarded) != 1 || res.Awarded[0].Amount != want {
			t.Fatalf("%s: %+v want %d", Difficulties[i], res, want)
		}
	}
}

func TestEtherWalletAbsentIsInert(t *testing.T) {
	r := newRig(t, "Paladin", 6)
	d := r.desk(nil)
	for _, o := range d.View().Offers {
		if o.Pays != nil {
			t.Fatalf("pays without a wallet: %+v", o)
		}
	}
	if res := r.handIn(); res.Outcome != HandInDone || res.Awarded != nil {
		t.Fatalf("%+v", res)
	}
	d.invalidate()
	if h := d.View().History; len(h) != 1 || h[0].Paid != nil {
		t.Fatalf("history: %+v", h)
	}
}

func TestEtherTwoHeroesOneAccount(t *testing.T) {
	m := newMemLedger()
	a := newRig(t, "Paladin", 6)
	a.led = m
	if res := a.handIn(); len(res.Awarded) != 1 {
		t.Fatalf("first hero: %+v", res)
	}
	b := newRig(t, "Sorceress", 6)
	b.hero = "Second"
	b.restart()
	b.led = m
	res := b.handIn()
	if res.Outcome != HandInDone || len(res.Awarded) != 0 || m.total() != 2 {
		t.Fatalf("second hero: %+v total %d", res, m.total())
	}
}

func TestEtherDeskPaysAndHistory(t *testing.T) {
	r := newRig(t, "Paladin", 6)
	m := newMemLedger()
	r.led = m
	d := r.desk(nil)
	var pays []Pay
	for _, o := range d.View().Offers {
		if o.Quest == den {
			pays = o.Pays
		}
	}
	if len(pays) != 1 || pays[0].Kind != KindEther || pays[0].Amount != 2 {
		t.Fatalf("offer pays: %+v", pays)
	}
	if a := d.Start(den, ""); !a.OK {
		t.Fatalf("%+v", a)
	}
	r.tick(4)
	if a := d.HandIn(den, ""); !a.OK {
		t.Fatalf("%+v", a)
	}
	h := waitHand(t, d)
	if h.State != "done" || len(h.Awarded) != 1 || h.Awarded[0].Amount != 2 {
		t.Fatalf("hand-in view: %+v", h)
	}
	v := d.View()
	if len(v.History) != 1 || len(v.History[0].Paid) != 1 {
		t.Fatalf("history: %+v", v.History)
	}
	for _, q := range v.Quests {
		if q.ID == den && (len(q.Paid) != 1 || q.Paid[0].Amount != 2) {
			t.Fatalf("quest row: %+v", q.Paid)
		}
	}
	for _, o := range v.Offers {
		if o.Quest == den {
			t.Fatalf("done quest still offered: %+v", o)
		}
	}
}
