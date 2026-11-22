package d2dctl_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"d2lod-bot/internal/cheat"
	"d2lod-bot/internal/d2dctl"
)

func TestCreditGoldLineLedgerAndReply(t *testing.T) {
	p, lines, led := newPay(t, false,
		`{"ok":true,"credited":600,"stash_before":1000,"stash_after":1600,"room":2498400,"order":"s-1"}`+"\nok\n",
		`{"ok":true,"credited":0,"already":true,"stash_before":1600,"stash_after":1600,"room":2498400,"order":"s-1"}`+"\nok\n")
	r, err := p.CreditGold(context.Background(), 600, "s-1", 3)
	if err != nil || r.Credited != 600 || r.StashAfter != 1600 || r.Already || r.Room != 2498400 {
		t.Fatalf("credit: %+v %v", r, err)
	}
	if r, err = p.CreditGold(context.Background(), 600, "s-1", 3); err != nil || r.Credited != 0 || !r.Already {
		t.Fatalf("repeat: %+v %v", r, err)
	}
	if got := strings.Join(lines(), "|"); got != "debug creditgold 600 s-1|debug creditgold 600 s-1" {
		t.Fatalf("lines %q", got)
	}
	es, err := led.List(cheat.Filter{Kind: d2dctl.KindD2DDebug})
	if err != nil || len(es) != 4 || es[0].Success || !es[1].Success || es[0].Reason != "player vault sale order s-1" || es[0].Params["items"] != float64(3) && es[0].Params["items"] != 3 {
		t.Fatalf("ledger before the verb, with the reason and the item count: %d %v %+v", len(es), err, es)
	}
}

func TestCreditGoldRefusals(t *testing.T) {
	cases := []struct {
		reply string
		want  error
	}{
		{"err vault pay off: debug creditgold needs d2d --vault-pay\n", d2dctl.ErrPayOff},
		{"err measured run: creditgold is refused\n", d2dctl.ErrPayMeasured},
		{"err room: 9000 asked, the stash has room for 10 (2499990 of 2500000)\n", d2dctl.ErrCreditRoom},
		{"err amount: a positive whole number\n", d2dctl.ErrPayAmount},
	}
	for _, c := range cases {
		p, _, led := newPay(t, false, c.reply)
		if _, err := p.CreditGold(context.Background(), 9000, "s-2", 1); !errors.Is(err, c.want) {
			t.Fatalf("%q: %v", c.reply, err)
		}
		if es, _ := led.List(cheat.Filter{}); len(es) != 2 || es[1].Success || es[1].Error == "" {
			t.Fatalf("a refusal is ledgered as failed: %+v", es)
		}
	}
}

func TestCreditGoldMeasuredNeverSends(t *testing.T) {
	p, lines, led := newPay(t, true, "ok\n")
	if _, err := p.CreditGold(context.Background(), 10, "s-3", 1); !errors.Is(err, d2dctl.ErrPayMeasured) {
		t.Fatalf("measured: %v", err)
	}
	if es, _ := led.List(cheat.Filter{}); len(lines()) != 0 || len(es) != 0 {
		t.Fatalf("nothing sent, nothing ledgered: %v %d", lines(), len(es))
	}
	p, _, _ = newPay(t, false, "ok\n")
	if _, err := p.CreditGold(context.Background(), 0, "s-4", 1); !errors.Is(err, d2dctl.ErrPayAmount) {
		t.Fatalf("zero: %v", err)
	}
}

func TestPayFakeCreditMirrorsTheFork(t *testing.T) {
	f := &d2dctl.PayFake{Stash: d2dctl.StashLimit - 500, Purse: 300}
	ctx := context.Background()
	r, err := f.CreditGold(ctx, 400, "c1", 1)
	if err != nil || r.StashAfter != d2dctl.StashLimit-100 || r.Room != 100 || f.Purse != 300 {
		t.Fatalf("%+v %v", r, err)
	}
	if r, err = f.CreditGold(ctx, 400, "c1", 1); err != nil || r.Credited != 0 || !r.Already || f.Stash != d2dctl.StashLimit-100 {
		t.Fatalf("repeat %+v %v", r, err)
	}
	if _, err = f.CreditGold(ctx, 101, "c2", 1); !errors.Is(err, d2dctl.ErrCreditRoom) || f.Stash != d2dctl.StashLimit-100 {
		t.Fatalf("past the limit: %v", err)
	}
	if r, err = f.CreditGold(ctx, 100, "c3", 1); err != nil || r.Room != 0 || f.Stash != d2dctl.StashLimit {
		t.Fatalf("to the limit: %+v %v", r, err)
	}
	g := &d2dctl.PayFake{Measured: true}
	if _, err = g.CreditGold(ctx, 1, "c4", 1); !errors.Is(err, d2dctl.ErrPayMeasured) {
		t.Fatalf("measured: %v", err)
	}
}

func TestParsePriceLine(t *testing.T) {
	p, err := d2dctl.ParsePriceLine("npc=Akara sell=87 sell_id=530 was=0")
	if err != nil || p.Npc != "Akara" || p.Sell != 87 || p.SellID != 530 || p.Was {
		t.Fatalf("%+v %v", p, err)
	}
	for _, bad := range []string{"", "npc=Akara sell=x sell_id=1 was=0", "npc=A sell=1"} {
		if _, err := d2dctl.ParsePriceLine(bad); err == nil {
			t.Fatalf("%q must fail", bad)
		}
	}
}

func TestCapsListCreditAndPrice(t *testing.T) {
	c := d2dctl.VaultCapsOf([]string{"item.full", "vault.creditgold", "item.price"})
	if !c.CreditGold || !c.Price || c.PayGold {
		t.Fatalf("%+v", c)
	}
}
