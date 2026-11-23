package leveler

import (
	"context"
	"errors"
	"testing"

	"d2lod-bot/internal/d2"
)

// verbGame is a hero whose stat points fall by one when the verb is sent, or not when it ignores them.
type verbGame struct {
	pts    int
	ignore bool
	err    error
	sent   []int
}

func (g *verbGame) StatPoint(_ context.Context, attr int) error {
	g.sent = append(g.sent, attr)
	if g.err == nil && !g.ignore {
		g.pts--
	}
	return g.err
}
func (g *verbGame) SkillPoint(context.Context, int) error { return errors.New("not used") }

func verbRig(g *verbGame) *Leveler {
	e := quietEnv()
	l := New(e, Options{Verbs: g})
	l.player = func() (*d2.Player, error) { return &d2.Player{StatPoints: g.pts}, nil }
	return l
}

func TestStatPointsByVerbSkipTheClicks(t *testing.T) {
	g := &verbGame{pts: 3}
	l := verbRig(g)
	acts := []Action{{Kind: KindStat, Attr: Vit}, {Kind: KindStat, Attr: Str}, {Kind: KindStat, Attr: Vit}}
	var rep Report
	st, sk, err := l.runVerbs(context.Background(), acts, nil, &rep)
	if err != nil || len(st) != 0 || len(sk) != 0 || len(rep.Done) != 3 || len(g.sent) != 3 || g.sent[0] != Vit {
		t.Fatalf("rest %v %v err %v done %d sent %v", st, sk, err, len(rep.Done), g.sent)
	}
	if l.Stats().StatPoints != 3 {
		t.Errorf("counter %d", l.Stats().StatPoints)
	}
}

func TestVerbNotTakenFallsBackToClicks(t *testing.T) {
	for name, g := range map[string]*verbGame{"ignored": {pts: 2, ignore: true}, "error": {pts: 2, err: errors.New("err unknown")}} {
		l := verbRig(g)
		acts := []Action{{Kind: KindStat, Attr: Dex}, {Kind: KindStat, Attr: Dex}}
		var rep Report
		st, _, err := l.runVerbs(context.Background(), acts, nil, &rep)
		if err != nil || len(st) != 2 || len(rep.Done) != 0 {
			t.Fatalf("%s: rest %v err %v done %d", name, st, err, len(rep.Done))
		}
		// the verbs stay off: the next call does not send anything
		before := len(g.sent)
		if st, _, _ = l.runVerbs(context.Background(), acts, nil, &rep); len(st) != 2 || len(g.sent) != before {
			t.Errorf("%s: verbs still on after a failure (sent %d -> %d)", name, before, len(g.sent))
		}
	}
}

func TestNoVerbsIsTheOldPath(t *testing.T) {
	l := New(quietEnv(), Options{})
	acts := []Action{{Kind: KindStat, Attr: Vit}}
	st, _, err := l.runVerbs(context.Background(), acts, nil, &Report{})
	if err != nil || len(st) != 1 {
		t.Fatalf("%v %v", st, err)
	}
}
