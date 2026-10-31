package drill

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// DR returns the repair group (P16): the repair drills of D3 and the poor-purse and price-table drills of DRILLS4.
// Run it with `-repair-below 100` so the chore also repairs the half-worn rows of the price table.
func DR(l TownLab) []Drill {
	lim := 15 * time.Minute
	do := func(f func(context.Context, *Run, TownLab) Outcome) func(context.Context, *Run) Outcome {
		return func(ctx context.Context, r *Run) Outcome { return f(ctx, r, l) }
	}
	runs := 0
	return []Drill{
		{ID: "D3.r", Primitive: "P16", Title: "worn items at durability 1, gold enough: repair restores durability, price logged", Limit: lim, Do: do(d3r)},
		{ID: "D3.r0", Primitive: "P16", Title: "worn items at durability 1, gold 0: no repair progress, no loop, no gold lost", Limit: lim, Do: do(d3r0)},
		{ID: "D3.r1", Primitive: "P16", Title: "worn items at durability 1, gold 5 (below the price): the smith refuses, gold unchanged, UI clean", Limit: lim, Do: func(ctx context.Context, r *Run) Outcome { runs++; return d3r1(ctx, r, l, runs) }},
		{ID: "D3.rt", Primitive: "P16", Title: "repair price table: one worn item of each base damaged alone, gold paid per base and durability lost", Limit: 40 * time.Minute, Do: do(d3rt)},
	}
}

// d3r1: the purse holds less than the price. The bot's chore (before fix-bank) does not look at gold, so the smith
// window is opened and the repair-all button pressed; d2d answers by not repairing.
//
// The town code latches a repair that changed nothing until the lowest durability changes (internal/town/basicsfix.go),
// so each repeat starts from a different durability (1, 2, ...): a repeat in the same process must not read the latch of
// the run before it as the bot's answer to the poor purse.
func d3r1(ctx context.Context, r *Run, l TownLab, run int) Outcome {
	if err := l.ClearPack(ctx, r); err != nil {
		return setupFail(err)
	}
	if err := l.SetGold(ctx, r, 5); err != nil {
		return setupFail(err)
	}
	if _, err := r.Cheat(ctx, fmt.Sprintf("debug durability all %d", run), "every worn item to a low durability (below the repair knob; it differs per repeat so the repair latch of the run before does not hide the first call)"); err != nil {
		return setupFail(err)
	}
	if err := l.Tick(ctx, 20); err != nil {
		return setupFail(err)
	}
	v0, err := l.TView(ctx)
	if err != nil {
		return setupFail(err)
	}
	if v0.DurMax <= 0 || 100*v0.DurCur/v0.DurMax >= l.RepairKnob() {
		return Blocks(fmt.Sprintf("gear at %d/%d is not below the repair knob %d", v0.DurCur, v0.DurMax, l.RepairKnob()))
	}
	ci := l.Chores(ctx, "repair")
	v1, err := l.TView(ctx)
	if err != nil {
		return setupFail(err)
	}
	obs := fmt.Sprintf("durability %d/%d -> %d/%d; gold %d -> %d; steps [%s] err %q ticks %d; panels %s",
		v0.DurCur, v0.DurMax, v1.DurCur, v1.DurMax, v0.Gold, v1.Gold, strings.Join(ci.Steps, "; "), ci.Err, ci.Ticks, panelsOf(v1.Panels))
	o := Outcome{Intended: "gold below the price: the visit is skipped or refused, gold unchanged, durability unchanged, UI clean", Observed: obs,
		Sig: fmt.Sprintf("unchanged=%v gold=%d ran=%v", v1.Gold == v0.Gold && v1.DurCur == v0.DurCur, v1.Gold, !strings.Contains(strings.Join(ci.Steps, ";"), "repair skipped"))}
	switch {
	case v1.Gold != v0.Gold || v1.DurCur != v0.DurCur:
		o.Verdict, o.Reason = Fail, "gold or durability changed below the price"
	case len(v1.Panels) > 0:
		o.Verdict, o.Reason = Fail, "panel left open"
	case !strings.Contains(strings.Join(ci.Steps, ";"), "repair skipped"):
		o.Verdict, o.Reason = Fail, "the chore walked to the smith although the purse (5) cannot pay: the repair step does not check gold (finding F4 of DRILLS3; fix-bank pending). Gold and durability are unchanged and the UI is clean"
	default:
		o.Verdict, o.Reason = Pass, "the chore skipped the visit"
	}
	return o
}

// RepairBase is one base of the price table: item code, body slot (1 head, 3 body armour, 4 right hand, 5 left hand,
// 9 boots, 10 gloves).
type RepairBase struct {
	Code string
	Slot int
}

// RepairBases is what the price table covers: armour of every weight class, weapons, shields, boots and gloves, all
// normal-quality bases of the Act 1 smith's range.
var RepairBases = []RepairBase{{"cap", 1}, {"lea", 3}, {"qui", 3}, {"spl", 3}, {"ful", 3}, {"hax", 4}, {"ssd", 4}, {"lsd", 4}, {"sst", 4}, {"buc", 5}, {"kit", 5}, {"lbt", 9}, {"lgl", 10}}

// stdDurability is the normal-quality maximum durability of each base (Armor.txt and Weapons.txt durability). `debug equip`
// sometimes makes an ethereal item (V live: cap 7, hax 15, that is half plus one): the smith cannot repair those (a
// game rule), so the drill equips again, a few times, until the maximum is the normal one.
var stdDurability = map[string]int{"cap": 12, "lea": 24, "qui": 20, "spl": 30, "ful": 70, "hax": 28, "ssd": 24, "lsd": 44, "sst": 20, "buc": 12, "kit": 30, "lbt": 12, "lgl": 12}

// RepairRow is one measurement: gold paid to restore one item of this base from Cur to Max.
type RepairRow struct {
	Code      string
	Slot      int
	Max, Cur  int
	Paid      int // purse before minus after the chore
	Spent     int // the chore's own report
	Restored  bool
	Err       string
	NoDur     bool // the base has no durability
	Ethereal  bool // the equipped item was ethereal on every try: the smith does not repair it
	ChoreTick int
}

func (r RepairRow) String() string {
	if r.NoDur {
		return fmt.Sprintf("%s slot %d: no durability", r.Code, r.Slot)
	}
	s := fmt.Sprintf("%s slot %d dur %d/%d lost %d paid %d", r.Code, r.Slot, r.Cur, r.Max, r.Max-r.Cur, r.Paid)
	if r.Ethereal {
		return s + " (ethereal item: not repairable, expected)"
	}
	if !r.Restored {
		s += " NOT RESTORED"
	}
	if r.Spent != r.Paid {
		s += fmt.Sprintf(" (chore reports %d)", r.Spent)
	}
	if r.Err != "" {
		s += " err " + r.Err
	}
	return s
}

func d3rt(ctx context.Context, r *Run, l TownLab) Outcome {
	if err := l.ClearPack(ctx, r); err != nil {
		return setupFail(err)
	}
	var rows []RepairRow
	bad := 0
	for _, b := range RepairBases {
		if err := l.SetGold(ctx, r, 10000); err != nil {
			return setupFail(err)
		}
		// an already empty slot answers with an error: that is the state wanted, so it is not one
		max, eth := 0, false
		for try := 0; try < 4; try++ {
			// an already empty slot answers with an error: that is the state wanted, so it is not one
			_, _ = r.Cheat(ctx, fmt.Sprintf("debug unequip %d", b.Slot), fmt.Sprintf("empty body slot %d so the next item's maximum durability is the change in the worn total (price table, base %s)", b.Slot, b.Code))
			_ = l.Tick(ctx, 12)
			v0, err := l.TView(ctx)
			if err != nil {
				return setupFail(err)
			}
			if _, err := r.Cheat(ctx, fmt.Sprintf("debug equip %s %d", b.Code, b.Slot), fmt.Sprintf("wear one %s: the price table damages and repairs one base at a time", b.Code)); err != nil {
				rows = append(rows, RepairRow{Code: b.Code, Slot: b.Slot, Err: "equip: " + err.Error()})
				bad++
				break
			}
			if _, err := r.Cheat(ctx, "debug durability all 9999", "every worn item to its maximum durability (capped): only the item under test is damaged next"); err != nil {
				return setupFail(err)
			}
			_ = l.Tick(ctx, 12)
			v1, err := l.TView(ctx)
			if err != nil {
				return setupFail(err)
			}
			max = v1.DurMax - v0.DurMax
			std, known := stdDurability[b.Code]
			eth = known && max > 0 && max != std
			if !eth {
				break
			}
		}
		if max <= 0 {
			if len(rows) == 0 || rows[len(rows)-1].Code != b.Code {
				rows = append(rows, RepairRow{Code: b.Code, Slot: b.Slot, NoDur: true})
			}
			continue
		}
		for _, to := range []int{1, max / 2} {
			if to < 1 || (to == 1 && max/2 == 1) {
				continue
			}
			if _, err := r.Cheat(ctx, fmt.Sprintf("debug durability %d %d", b.Slot, to), fmt.Sprintf("damage the %s alone to durability %d of %d: one row of the repair price table", b.Code, to, max)); err != nil {
				rows = append(rows, RepairRow{Code: b.Code, Slot: b.Slot, Max: max, Cur: to, Err: "durability: " + err.Error()})
				bad++
				continue
			}
			_ = l.Tick(ctx, 12)
			a, err := l.TView(ctx)
			if err != nil {
				return setupFail(err)
			}
			ci := l.Chores(ctx, "repair")
			z, err := l.TView(ctx)
			if err != nil {
				return setupFail(err)
			}
			row := RepairRow{Code: b.Code, Slot: b.Slot, Max: max, Ethereal: eth, Cur: to, Paid: a.Gold - z.Gold, Spent: ci.Spent, Restored: z.DurCur == z.DurMax, Err: ci.Err, ChoreTick: ci.Ticks}
			if (!row.Restored || row.Paid <= 0) && !eth {
				bad++
			}
			rows = append(rows, row)
		}
	}
	var lines []string
	for _, x := range rows {
		lines = append(lines, x.String())
	}
	obs := strings.Join(lines, "; ")
	o := Outcome{Intended: "per base: durability restored to max, gold paid > 0 (the price is the measurement)", Observed: obs, Sig: obs}
	if bad > 0 {
		o.Verdict, o.Reason = Fail, fmt.Sprintf("%d of %d rows not restored or not paid (or a cheat step failed)", bad, len(rows))
		return o
	}
	o.Verdict, o.Reason = Pass, fmt.Sprintf("%d rows measured, all restored and paid", len(rows))
	return o
}
