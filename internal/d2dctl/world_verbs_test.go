package d2dctl

import (
	"context"
	"errors"
	"testing"
)

func TestWorldVerbReplies(t *testing.T) {
	ctx := context.Background()
	s, c := startStub(t, map[string]string{
		"buy 3 1":      "{\"verb\":\"buy\",\"code\":\"hp1\",\"stock\":3,\"tab\":1,\"price\":30,\"gold_after\":970}\nok\n",
		"buy 9 0":      "err stock no such stock index on that tab\n",
		"buy 2 0":      "err gold not enough gold for hp1\n",
		"sell 7":       "{\"verb\":\"sell\",\"id\":7,\"code\":\"hax\",\"price\":24}\nok\n",
		"sell 8":       "err closed no vendor window is open\n",
		"waypoint 3":   "{\"verb\":\"waypoint\",\"object\":5,\"level\":3,\"from\":1}\nok\n",
		"waypoint 4":   "err inactive that waypoint is not active yet\n",
		"portal":       "{\"verb\":\"portal\",\"portals\":[{\"which\":0,\"level\":2,\"x\":10.5,\"y\":3.0,\"here\":true}]}\nok\n",
		"pickup 12":    "{\"verb\":\"pickup\",\"id\":12,\"code\":\"gld\",\"dist\":4.2}\nok\n",
		"pickup 13 2":  "err far 4.2 cells away, beyond 2\n",
		"pickup 14":    "err unknown verb pickup\n",
		"sell 15":      "err joined a joined game\n",
		"waypoint 5 6": "{\"verb\":\"waypoint\",\"object\":6,\"level\":5,\"from\":1}\nok\n",
	}, nil)
	b, err := c.VerbBuy(ctx, 3, 1, false)
	if err != nil || b.Code != "hp1" || b.Price != 30 || b.GoldAfter != 970 {
		t.Fatalf("buy: %+v %v", b, err)
	}
	_, err = c.VerbBuy(ctx, 9, 0, false)
	if r, ok := IsVerbRefusal(err); !ok || r.Code != VerbErrStock || !r.RuleRefusal() {
		t.Fatalf("stock refusal: %v", err)
	}
	if _, err = c.VerbBuy(ctx, 2, 0, false); err == nil {
		t.Fatal("gold refusal missing")
	} else if r, _ := IsVerbRefusal(err); r == nil || r.Code != VerbErrGold {
		t.Fatalf("gold refusal: %v", err)
	}
	if r, err := c.VerbSell(ctx, 7); err != nil || r.Price != 24 || r.Code != "hax" {
		t.Fatalf("sell: %+v %v", r, err)
	}
	if _, err = c.VerbSell(ctx, 8); err == nil {
		t.Fatal("closed refusal missing")
	} else if r, _ := IsVerbRefusal(err); r == nil || r.Code != VerbErrClosed {
		t.Fatalf("closed: %v", err)
	}
	if r, err := c.VerbWaypoint(ctx, 3, -1); err != nil || r.Level != 3 || r.From != 1 {
		t.Fatalf("waypoint: %+v %v", r, err)
	}
	if r, err := c.VerbWaypoint(ctx, 5, 6); err != nil || r.Object != 6 {
		t.Fatalf("waypoint object: %+v %v", r, err)
	}
	if _, err = c.VerbWaypoint(ctx, 4, -1); err == nil {
		t.Fatal("inactive refusal missing")
	} else if r, _ := IsVerbRefusal(err); r == nil || r.Code != VerbErrInactive {
		t.Fatalf("inactive: %v", err)
	}
	ps, err := c.VerbPortals(ctx)
	if err != nil || len(ps) != 1 || !ps[0].Here || ps[0].Level != 2 {
		t.Fatalf("portals: %+v %v", ps, err)
	}
	if r, err := c.VerbPickup(ctx, 12, 0); err != nil || r.Code != "gld" || r.Dist != 4.2 {
		t.Fatalf("pickup: %+v %v", r, err)
	}
	if _, err = c.VerbPickup(ctx, 13, 2); err == nil {
		t.Fatal("far refusal missing")
	} else if r, _ := IsVerbRefusal(err); r == nil || r.Code != VerbErrFar {
		t.Fatalf("far: %v", err)
	}
	// An older build (unknown verb) and a joined game are NOT rule refusals: the caller falls back to the click path.
	for _, fn := range []func() error{
		func() error { _, err := c.VerbPickup(ctx, 14, 0); return err },
		func() error { _, err := c.VerbSell(ctx, 15); return err },
	} {
		err := fn()
		var r *VerbRefusal
		if err == nil || errors.As(err, &r) {
			t.Fatalf("fallback error expected, got %v", err)
		}
	}
	if got := s.seen[len(s.seen)-1]; got != "sell 15" {
		t.Fatalf("last sent %q", got)
	}
}

func TestVendorStockIndex(t *testing.T) {
	v := Vendor{Items: []VendorItem{{Tab: 0, X: 0, Y: 0}, {Tab: 0, X: 2, Y: 0}, {Tab: 1, X: 0, Y: 0}, {Tab: 1, X: 0, Y: 3}}}
	if i, ok := v.StockIndex(1, 0, 3); !ok || i != 1 {
		t.Fatalf("got %d %v", i, ok)
	}
	if i, ok := v.StockIndex(0, 2, 0); !ok || i != 1 {
		t.Fatalf("got %d %v", i, ok)
	}
	if _, ok := v.StockIndex(2, 0, 0); ok {
		t.Fatal("tab 2 has no items")
	}
}
