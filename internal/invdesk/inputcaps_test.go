package invdesk

import (
	"context"
	"strings"
	"testing"

	"d2lod-bot/internal/d2"
)

// The input.* names (Q13) stand for the lab names only behind the knob input.planner: with the knob off a build that
// lists only the new names is treated as having no verb (the click path, byte for byte as before).
func TestInputCapsEquipOnlyBehindPlannerKnob(t *testing.T) {
	for _, c := range []struct {
		class d2.Class
		good  uint32
	}{{d2.Paladin, 1}, {d2.Sorceress, 2}} {
		t.Run(c.class.String(), func(t *testing.T) {
			planner := false
			fv := &fakeVerbs{caps: []string{"state.caps", InputCapItemEquip}}
			d, eq, ve := verbRig(t, c.class, fv, on)
			ve.InputCaps = func() bool { return planner }
			d.RequestEquip(c.good, "rhand")
			lines := d.RunPending(context.Background())
			if len(fv.equips) != 0 || len(eq.worn) != 1 || !strings.Contains(lines[0], "(how=click)") {
				t.Fatalf("planner off must not use input.item-equip: verb %v click %v %v", fv.equips, eq.worn, lines)
			}

			fv2 := &fakeVerbs{caps: []string{"state.caps", InputCapItemEquip}}
			d, eq, ve = verbRig(t, c.class, fv2, on)
			ve.InputCaps = func() bool { return true }
			d.RequestEquip(c.good, "rhand")
			lines = d.RunPending(context.Background())
			if len(fv2.equips) != 1 || len(eq.worn) != 0 || !strings.Contains(lines[0], "(how=verb)") {
				t.Fatalf("planner on must use input.item-equip: verb %v click %v %v", fv2.equips, eq.worn, lines)
			}
		})
	}
}

func TestInputCapsMoverAliasEveryVerb(t *testing.T) {
	inputOnly := []string{"state.caps", InputCapItemMove, InputCapItemIdent, InputCapItemDrop, InputCapGoldStash}
	for _, planner := range []bool{false, true} {
		fv := &fakeMoverVerbs{caps: inputOnly}
		w := &world{where: d2.Container(0)}
		m := newMover(fv, w, true)
		m.InputCaps = func() bool { return planner }
		ctx := context.Background()
		d1, _ := m.Identify(ctx, 7)
		d2_, _ := m.Drop(ctx, 7)
		_, d3, _ := m.Gold(ctx, 100, true)
		got := d1 || d2_ || d3
		if got != planner {
			t.Fatalf("planner %v: handled identify %v drop %v gold %v", planner, d1, d2_, d3)
		}
		if planner && len(fv.calls) < 3 {
			t.Fatalf("planner on: calls %v", fv.calls)
		}
		if !planner && len(fv.calls) != 0 {
			t.Fatalf("planner off: verbs were called %v", fv.calls)
		}
	}
}
