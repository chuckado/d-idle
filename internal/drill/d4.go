package drill

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// PackLab is the instance as the D4 drills use it: the bot's pickit, MakeRoom and Transmute, plus setup.
type PackLab interface {
	TownLab
	PickitLog(ctx context.Context) (PickInfo, []string, error) // loot.Looter.PickUp, with the bot's loot log lines of the call
	GroundCodes(ctx context.Context) ([]string, error)
	FillPack(ctx context.Context, r *Run, code string, n int) (int, error) // gives and picks up n items one by one with the engine's pickup; returns how many are in the pack
	PackFree(ctx context.Context, w, h int) (bool, error)                  // town.PackFits
	MakeRoom(ctx context.Context, w, h int) (bool, string, error)          // vendor open, town.MakeRoom
	Transmute(ctx context.Context, codes []string) ([]string, error)
	PurseNow(ctx context.Context) (int, error)
}

// D4 returns the pack group: P9 pickup by rule, P4 pack full and room, P8 gold into a full pack, P6 cube.
func D4(l PackLab) []Drill {
	lim := 10 * time.Minute
	do := func(f func(context.Context, *Run, PackLab) Outcome) func(context.Context, *Run) Outcome {
		return func(ctx context.Context, r *Run) Outcome { return f(ctx, r, l) }
	}
	return []Drill{
		{ID: "D4.a", Primitive: "P9", Title: "pickit on a mixed pile: the report agrees with the end state, each item logged once, a rerun does not retry", Limit: lim, Do: do(d4a)},
		{ID: "D4.b", Primitive: "P4", Title: "pack full of 1x1 items: the pickit leaves a new item once, no loop, no panel left open", Limit: lim, Do: do(d4b)},
		{ID: "D4.c", Primitive: "P4", Title: "pack full: MakeRoom at a vendor frees a 2x2 hole (sells or refuses, no loop)", Limit: lim, Do: do(d4c)},
		{ID: "D4.g", Primitive: "P8", Title: "gold pile with a full pack: the purse rises by the pile", Limit: lim, Do: do(d4g)},
		{ID: "D4.cap", Primitive: "P8", Title: "purse near the carry cap (10000 per level, level 1): a pile that does not fit", Limit: lim, Do: do(d4cap)},
		{ID: "D4.x", Primitive: "P6", Title: "cube: three chipped amethysts become one flawed amethyst, the cube is kept", Limit: lim, Do: do(d4x)},
	}
}

func groundString(c []string) string {
	if len(c) == 0 {
		return "empty"
	}
	m := map[string]int{}
	for _, x := range c {
		m[x]++
	}
	return countString(m)
}

func giveAll(ctx context.Context, r *Run, l PackLab, codes ...string) error {
	for _, c := range codes {
		if err := l.GiveGround(ctx, r, c); err != nil {
			return err
		}
	}
	return l.Tick(ctx, 30)
}

func d4a(ctx context.Context, r *Run, l PackLab) Outcome {
	if err := l.ClearPack(ctx, r); err != nil {
		return setupFail(err)
	}
	if err := giveAll(ctx, r, l, "gcv", "cap", "rin", "hp1"); err != nil {
		return setupFail(err)
	}
	// items an earlier drill left on the ground (refused ones stay: there is no verb that clears the ground) are in
	// every count below; the checks use differences
	g0, _ := l.GroundCodes(ctx)
	v0, _ := l.TView(ctx)
	p1, log1, err := l.PickitLog(ctx)
	if err != nil {
		return setupFail(fmt.Errorf("pickit: %w", err))
	}
	v1, _ := l.TView(ctx)
	g1, _ := l.GroundCodes(ctx)
	p2, log2, _ := l.PickitLog(ctx)
	v2, _ := l.TView(ctx)
	g2, _ := l.GroundCodes(ctx)
	obs := fmt.Sprintf("ground %s; call 1: %s | pack %s belt %d, ground %s; call 2: %s | pack %s, ground %s; log lines call1 %d call2 %d",
		groundString(g0), p1, v1.PackString(), v1.Belt, groundString(g1), p2, v2.PackString(), groundString(g2), len(log1), len(log2))
	o := Outcome{Intended: "report = end state: picked items in the pack or belt, skipped items still on the ground, rerun picks nothing new", Observed: obs}
	picked := len(p1.Picked)
	held := func(v TownView) int {
		n := v.Belt
		for _, c := range v.Pack {
			n += c
		}
		return n
	}
	have := held(v1) - held(v0)
	o.Sig = fmt.Sprintf("picked=%d held+%d ground-%d rerun=%d", picked, have, len(g0)-len(g1), len(p2.Picked))
	switch {
	case picked != have:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("the report says %d picked, the pack and belt gained %d", picked, have)
	case len(g1) != len(g0)-picked:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("ground went from %d to %d items for %d picked", len(g0), len(g1), picked)
	case len(p2.Picked) != 0 || v2.PackString() != v1.PackString():
		o.Verdict, o.Reason = Fail, "the second call picked or changed something"
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("picked %d of %d, left %d, rerun quiet", picked, len(g0), len(g1))
	}
	return o
}

func d4b(ctx context.Context, r *Run, l PackLab) Outcome {
	if err := l.ClearPack(ctx, r); err != nil {
		return setupFail(err)
	}
	n, err := l.FillPack(ctx, r, "rin", 40)
	if err != nil {
		return setupFail(err)
	}
	if n < 20 {
		return setupFail(fmt.Errorf("only %d rings in a full pack", n))
	}
	if err := giveAll(ctx, r, l, "gcy"); err != nil {
		return setupFail(err)
	}
	p1, log1, err := l.PickitLog(ctx)
	if err != nil {
		return setupFail(fmt.Errorf("pickit: %w", err))
	}
	p2, log2, _ := l.PickitLog(ctx)
	g, _ := l.GroundCodes(ctx)
	v, _ := l.TView(ctx)
	gemLeft := 0
	for _, c := range g {
		if c == "gcy" {
			gemLeft++
		}
	}
	full := 0
	for _, ln := range append(append([]string{}, log1...), log2...) {
		if strings.Contains(ln, "pack full") {
			full++
		}
	}
	obs := fmt.Sprintf("pack rings %d; call 1: %s; call 2: %s; ground %s; 'pack full' lines over both calls %d (1 per item and call allowed, %d gems on the ground); panels %s", v.Pack["rin"], p1, p2, groundString(g), full, gemLeft, panelsOf(v.Panels))
	o := Outcome{Intended: "the gem stays on the ground, at most one 'pack full' line per call, nothing picked, no panel", Observed: obs, Sig: fmt.Sprintf("picked=%d/%d gemLeft>0=%v oneLinePerItemPerCall=%v", len(p1.Picked), len(p2.Picked), gemLeft > 0, full == 2*gemLeft)}
	switch {
	case len(p1.Picked)+len(p2.Picked) != 0:
		o.Verdict, o.Reason = Fail, "something was picked into a full pack"
	case full > 2*gemLeft:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("%d 'pack full' lines for %d gems and 2 calls: an item is logged more than once per call", full, gemLeft)
	case len(v.Panels) > 0:
		o.Verdict, o.Reason = Fail, "panel left open"
	default:
		o.Verdict, o.Reason = Pass, "full pack: item left once per call, no loop"
	}
	return o
}

func d4c(ctx context.Context, r *Run, l PackLab) Outcome {
	if err := l.ClearPack(ctx, r); err != nil {
		return setupFail(err)
	}
	n, err := l.FillPack(ctx, r, "rin", 40)
	if err != nil || n < 20 {
		return setupFail(fmt.Errorf("could only put %d rings in the pack (%v)", n, err))
	}
	if fit, _ := l.PackFree(ctx, 2, 2); fit {
		return setupFail(fmt.Errorf("a 2x2 hole exists before MakeRoom"))
	}
	ok, note, merr := l.MakeRoom(ctx, 2, 2)
	fit, _ := l.PackFree(ctx, 2, 2)
	v, _ := l.TView(ctx)
	obs := fmt.Sprintf("MakeRoom(2,2) -> %v err %v; hole now %v; rings left %d; gold %d; %s; panels %s", ok, merr, fit, v.Pack["rin"], v.Gold, note, panelsOf(v.Panels))
	o := Outcome{Intended: "a free 2x2 hole after MakeRoom, or a clean refusal", Observed: obs, Sig: fmt.Sprintf("ok=%v hole=%v rings=%d", ok, fit, v.Pack["rin"])}
	switch {
	case ok != fit:
		o.Verdict, o.Reason = Fail, "MakeRoom's answer disagrees with the pack"
	case !ok:
		o.Verdict, o.Reason = Fail, "MakeRoom could not free a 2x2 hole from a pack of plain rings"
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("hole made, %d rings left", v.Pack["rin"])
	}
	return o
}

func d4g(ctx context.Context, r *Run, l PackLab) Outcome {
	if err := l.ClearPack(ctx, r); err != nil {
		return setupFail(err)
	}
	n, err := l.FillPack(ctx, r, "rin", 40)
	if err != nil || n < 20 {
		return setupFail(fmt.Errorf("could only put %d rings in the pack (%v)", n, err))
	}
	if err := l.SetGold(ctx, r, 0); err != nil {
		return setupFail(err)
	}
	if err := giveAll(ctx, r, l, "gld"); err != nil {
		return setupFail(err)
	}
	g0, _ := l.GroundCodes(ctx)
	p, _, err := l.PickitLog(ctx)
	if err != nil {
		return setupFail(fmt.Errorf("pickit: %w", err))
	}
	v, _ := l.TView(ctx)
	g1, _ := l.GroundCodes(ctx)
	obs := fmt.Sprintf("full pack (%d rings); ground %s -> %s; purse 0 -> %d; %s", v.Pack["rin"], groundString(g0), groundString(g1), v.Gold, p)
	o := Outcome{Intended: "the pile is taken (purse > 0) although the pack has no free cell", Observed: obs, Sig: fmt.Sprintf("purse>0=%v ground=%s", v.Gold > 0, groundString(g1))}
	if v.Gold > 0 {
		o.Verdict, o.Reason = Pass, fmt.Sprintf("gold pile taken into a full pack, purse %d", v.Gold)
	} else {
		o.Verdict, o.Reason = Fail, "the gold pile was not taken with a full pack"
	}
	return o
}

func d4x(ctx context.Context, r *Run, l PackLab) Outcome {
	if err := l.ClearPack(ctx, r); err != nil {
		return setupFail(err)
	}
	for _, c := range []string{"box", "gcv", "gcv", "gcv"} {
		if err := giveAll(ctx, r, l, c); err != nil {
			return setupFail(err)
		}
		if _, err := l.PickupRaw(ctx); err != nil {
			return setupFail(err)
		}
	}
	v0, _ := l.TView(ctx)
	if v0.Pack["box"] != 1 || v0.Pack["gcv"] != 3 {
		return setupFail(fmt.Errorf("staging failed: pack %s", v0.PackString()))
	}
	out, err := l.Transmute(ctx, []string{"gcv", "gcv", "gcv"})
	v1, _ := l.TView(ctx)
	obs := fmt.Sprintf("pack %s -> %s; transmute output %v err %v; panels %s", v0.PackString(), v1.PackString(), out, err, panelsOf(v1.Panels))
	o := Outcome{Intended: "gcv x3 -> gfv x1, cube kept", Observed: obs, Sig: fmt.Sprintf("box=%d gcv=%d gfv=%d", v1.Pack["box"], v1.Pack["gcv"], v1.Pack["gfv"])}
	switch {
	case err != nil:
		o.Verdict, o.Reason = Fail, "transmute failed: "+err.Error()
	case v1.Pack["gfv"] != 1 || v1.Pack["gcv"] != 0 || v1.Pack["box"] != 1:
		o.Verdict, o.Reason = Fail, "end state is not one flawed amethyst plus the cube"
	default:
		o.Verdict, o.Reason = Pass, "chipped x3 became flawed x1, cube kept"
	}
	return o
}

func d4cap(ctx context.Context, r *Run, l PackLab) Outcome {
	if err := l.ClearPack(ctx, r); err != nil {
		return setupFail(err)
	}
	if err := l.SetGold(ctx, r, 9990); err != nil {
		return setupFail(err)
	}
	if err := giveAll(ctx, r, l, "gld"); err != nil {
		return setupFail(err)
	}
	p, _, err := l.PickitLog(ctx)
	if err != nil {
		return setupFail(fmt.Errorf("pickit: %w", err))
	}
	v, _ := l.TView(ctx)
	g, _ := l.GroundCodes(ctx)
	gold := 0
	for _, c := range g {
		if c == "gld" {
			gold++
		}
	}
	obs := fmt.Sprintf("purse 9990 -> %d; gold piles left on the ground %d; %s", v.Gold, gold, p)
	return Outcome{Verdict: Pass, Reason: fmt.Sprintf("recorded: the purse after taking a pile at 9990 gold, clvl 1, is %d (piles left: %d)", v.Gold, gold),
		Intended: "record the carry cap behaviour (SME: 10000 x clvl, U)", Observed: obs, Sig: fmt.Sprintf("purse=%d left>0=%v", v.Gold, gold > 0)}
}
