package d2dctl_test

import (
	"bufio"
	"context"
	"errors"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"d2lod-bot/internal/cheat"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2dctl/fake"
)

// payServer answers `debug paygold` lines with the scripted replies in order.
func payServer(t *testing.T, replies ...string) (path string, lines func() []string) {
	t.Helper()
	path, cleanup, err := fake.TempSocket()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(cleanup)
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	var mu sync.Mutex
	var got []string
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer c.Close()
				sc := bufio.NewScanner(c)
				for sc.Scan() {
					mu.Lock()
					n := len(got)
					got = append(got, sc.Text())
					mu.Unlock()
					c.Write([]byte(replies[min(n, len(replies)-1)]))
				}
			}()
		}
	}()
	return path, func() []string { mu.Lock(); defer mu.Unlock(); return append([]string(nil), got...) }
}

func newPay(t *testing.T, measured bool, replies ...string) (*d2dctl.Pay, func() []string, *cheat.Ledger) {
	t.Helper()
	path, lines := payServer(t, replies...)
	c := d2dctl.New(path, d2dctl.Options{Timeout: 2 * time.Second})
	t.Cleanup(func() { c.Close() })
	led := cheat.NewLedger(filepath.Join(t.TempDir(), "cheats.jsonl"))
	return &d2dctl.Pay{C: c, Ledger: led, Actor: "bot:vault", Instance: "i0", Character: "Hero", Measured: func() bool { return measured }}, lines, led
}

func TestPayGoldLineLedgerAndReply(t *testing.T) {
	p, lines, led := newPay(t, false,
		`{"ok":true,"paid":5800,"stash_before":7200,"stash_after":1400,"order":"g-1"}`+"\nok\n",
		`{"ok":true,"paid":0,"already":true,"stash_before":1400,"stash_after":1400,"order":"g-1"}`+"\nok\n")
	r, err := p.PayGold(context.Background(), 5800, "g-1")
	if err != nil || r.Paid != 5800 || r.StashAfter != 1400 || r.Already {
		t.Fatalf("pay: %+v %v", r, err)
	}
	r, err = p.PayGold(context.Background(), 5800, "g-1")
	if err != nil || r.Paid != 0 || !r.Already {
		t.Fatalf("repeat: %+v %v", r, err)
	}
	if got := strings.Join(lines(), "|"); got != "debug paygold 5800 g-1|debug paygold 5800 g-1" {
		t.Fatalf("lines %q", got)
	}
	es, err := led.List(cheat.Filter{Kind: d2dctl.KindD2DDebug})
	if err != nil || len(es) != 4 || es[0].Success || !es[1].Success || es[0].Reason != "vault gold reveal order g-1" || es[1].Actor != "bot:vault" {
		t.Fatalf("ledger %d %v %+v", len(es), err, es)
	}
}

func TestPayGoldRefusals(t *testing.T) {
	cases := []struct {
		reply string
		want  error
	}{
		{"err vault pay off: debug paygold needs d2d --vault-pay\n", d2dctl.ErrPayOff},
		{"err measured run: paygold is refused\n", d2dctl.ErrPayMeasured},
		{"err stash: 9000 asked, the stash holds 10\n", d2dctl.ErrPayShort},
		{"err amount: a positive whole number\n", d2dctl.ErrPayAmount},
	}
	for _, c := range cases {
		p, _, led := newPay(t, false, c.reply)
		_, err := p.PayGold(context.Background(), 9000, "g-2")
		if !errors.Is(err, c.want) {
			t.Fatalf("%q: %v", c.reply, err)
		}
		es, _ := led.List(cheat.Filter{})
		if len(es) != 2 || es[1].Success || es[1].Error == "" {
			t.Fatalf("a refusal is ledgered as failed: %+v", es)
		}
	}
}

func TestPayGoldMeasuredNeverSends(t *testing.T) {
	p, lines, led := newPay(t, true, "ok\n")
	if _, err := p.PayGold(context.Background(), 10, "g-3"); !errors.Is(err, d2dctl.ErrPayMeasured) {
		t.Fatalf("measured: %v", err)
	}
	if _, err := (&d2dctl.Pay{}).PayGold(context.Background(), 10, "g-3"); err == nil {
		t.Fatal("a Pay without a ledger must refuse")
	}
	if es, _ := led.List(cheat.Filter{}); len(lines()) != 0 || len(es) != 0 {
		t.Fatalf("nothing sent, nothing ledgered: %v %d", lines(), len(es))
	}
	p, _, _ = newPay(t, false, "ok\n")
	if _, err := p.PayGold(context.Background(), 0, "g-4"); !errors.Is(err, d2dctl.ErrPayAmount) {
		t.Fatalf("zero: %v", err)
	}
	if _, err := p.PayGold(context.Background(), 5, "bad/id"); err == nil {
		t.Fatal("bad id")
	}
}

func TestPayFakeMirrorsTheFork(t *testing.T) {
	f := &d2dctl.PayFake{Stash: 7200, Purse: 300}
	ctx := context.Background()
	r, err := f.PayGold(ctx, 5800, "o1")
	if err != nil || r.StashAfter != 1400 || f.Stash != 1400 || f.Purse != 300 {
		t.Fatalf("%+v %v", r, err)
	}
	if r, err = f.PayGold(ctx, 5800, "o1"); err != nil || r.Paid != 0 || !r.Already || f.Stash != 1400 {
		t.Fatalf("repeat %+v %v", r, err)
	}
	if _, err = f.PayGold(ctx, 1401, "o2"); !errors.Is(err, d2dctl.ErrPayShort) {
		t.Fatalf("overdraft %v", err)
	}
	if _, err = f.PayGold(ctx, -1, "o3"); !errors.Is(err, d2dctl.ErrPayAmount) {
		t.Fatalf("negative %v", err)
	}
	f.Measured = true
	if _, err = f.PayGold(ctx, 1, "o4"); !errors.Is(err, d2dctl.ErrPayMeasured) || f.Stash != 1400 {
		t.Fatalf("measured %v", err)
	}
}
