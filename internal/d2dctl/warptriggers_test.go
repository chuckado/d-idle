package d2dctl

import (
	"context"
	"testing"
	"time"

	"d2lod-bot/internal/d2dctl/fake"
)

const warpTriggersJSON = `{"level":2,"unit":"subtile","scan":40,"reach_rule":5,"warps":[{"index":0,"to":8,"unit":[335,140],"select":[-40,-60,80,60],"click":[[138,333,337],[139,332,338]],"reach":[[140,330,340]],"truncated":false}]}`

func TestWarpTriggersOverTheWire(t *testing.T) {
	path, cleanup, err := fake.TempSocket()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	fix := fake.Sample()
	fix.WarpTriggers = warpTriggersJSON
	srv, err := fake.New(path, fix, fake.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	c := New(path, Options{Timeout: 2 * time.Second})
	defer c.Close()
	w, err := c.WarpTriggers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if w.Level != 2 || w.ReachRule != 5 || len(w.Warps) != 1 {
		t.Fatalf("%+v", w)
	}
	if t0 := w.Warps[0]; t0.To != 8 || t0.Unit != [2]int{335, 140} || len(t0.Click) != 2 || t0.Click[1] != (Run{139, 332, 338}) || t0.Truncated {
		t.Fatalf("warp %+v", t0)
	}
}

func TestWarpTriggersMissingVerbIsAnError(t *testing.T) {
	path, cleanup, err := fake.TempSocket()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	srv, err := fake.New(path, fake.Sample(), fake.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	c := New(path, Options{Timeout: 2 * time.Second})
	defer c.Close()
	if _, err := c.WarpTriggers(context.Background()); err == nil {
		t.Fatal("stock d2d answers err: want an error")
	}
}
