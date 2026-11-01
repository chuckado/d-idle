package resource

import (
	"testing"
	"time"
)

func TestRecoveryCheck(t *testing.T) {
	tu := DefaultTuning()
	t0 := time.Unix(1000, 0)
	in := RecoveryInput{Gold: 37, Shortfall: 250, SellMin: 2, Now: t0}
	if v := RecoveryCheck(in, tu); !v.Start {
		t.Fatalf("broke with a kit shortfall must start: %+v", v)
	}
	in.Gold = 600
	if v := RecoveryCheck(in, tu); v.On {
		t.Fatalf("gold above the kit gold: %+v", v)
	}
	in.Gold, in.Shortfall = 300, 30
	if v := RecoveryCheck(in, tu); v.On {
		t.Fatalf("affordable kit: %+v", v)
	}
	in = RecoveryInput{Gold: 37, Shortfall: 250, Now: t0.Add(5 * time.Minute), LastEnd: t0}
	if v := RecoveryCheck(in, tu); v.On {
		t.Fatalf("cooldown: %+v", v)
	}
	in.LowKit = true
	if v := RecoveryCheck(in, tu); !v.Start {
		t.Fatalf("low kit overrides the cooldown: %+v", v)
	}
	in.Shortfall, in.Gold = 30, 40
	if v := RecoveryCheck(in, tu); v.On {
		t.Fatalf("low kit but affordable: %+v", v)
	}
	in.LowKit, in.Shortfall, in.Gold = false, 250, 37
	run := RecoveryInput{Was: true, Gold: 950, Sellable: 3, SellMin: 2, Since: t0, Now: t0.Add(10 * time.Minute)}
	if v := RecoveryCheck(run, tu); !v.On || v.End {
		t.Fatalf("loot still unsold: %+v", v)
	}
	run.Sellable = 0
	if v := RecoveryCheck(run, tu); !v.End {
		t.Fatalf("target reached: %+v", v)
	}
	run.Gold, run.Now = 100, t0.Add(46*time.Minute)
	if v := RecoveryCheck(run, tu); !v.End {
		t.Fatalf("time cap: %+v", v)
	}
	tu.RecoveryOn = 0
	if v := RecoveryCheck(in, tu); v.On {
		t.Fatalf("off: %+v", v)
	}
}

func TestRecoveryForceBypassesCooldown(t *testing.T) {
	tu := DefaultTuning()
	t0 := time.Unix(1000, 0)
	in := RecoveryInput{Gold: 0, Shortfall: 250, Now: t0.Add(5 * time.Minute), LastEnd: t0, Force: true}
	if v := RecoveryCheck(in, tu); !v.Start {
		t.Fatalf("forced start: %+v", v)
	}
	in.Gold = 600
	if v := RecoveryCheck(in, tu); v.On {
		t.Fatalf("force must not start with the kit paid: %+v", v)
	}
}

func TestKitLow(t *testing.T) {
	tu := DefaultTuning()
	tu.RecoveryPotionFloor = 2
	if !KitLow(1, tu) || KitLow(2, tu) {
		t.Fatal("floor 2: 1 is low, 2 is not")
	}
	tu.RecoveryPotionFloor = 0
	if KitLow(0, tu) {
		t.Fatal("floor 0 disables the override")
	}
}
