package d2dctl

import (
	"context"
	"testing"
)

func TestBaseAttrsAndPointVerbs(t *testing.T) {
	s, c := startStub(t, map[string]string{
		"dump stats":    `{"v":1,"stats":{"base":{"str":20,"dex":25,"vit":30,"ene":60},"bonus":[3,0,5,10]}}` + "\nok\n",
		"cmd stat 3 1":  "ok\n",
		"cmd skillpt 2": "ok\n",
	}, nil)
	ctx := context.Background()
	a, ok, err := c.BaseAttrs(ctx)
	if err != nil || !ok || a != [4]int{20, 25, 30, 60} {
		t.Fatalf("%v %v %v", a, ok, err)
	}
	if err := c.CmdStat(ctx, StatVitality, 1); err != nil {
		t.Fatal(err)
	}
	if err := c.CmdSkillPt(ctx, 2); err != nil {
		t.Fatal(err)
	}
	if got := s.seen[len(s.seen)-1]; got != "cmd skillpt 2" {
		t.Errorf("last verb %q", got)
	}
}

func TestBaseAttrsStockD2D(t *testing.T) {
	_, c := startStub(t, map[string]string{"dump stats": "err unknown verb dump\n"}, nil)
	if _, ok, err := c.BaseAttrs(context.Background()); ok || err != nil {
		t.Errorf("stock d2d: ok %v err %v", ok, err)
	}
}
