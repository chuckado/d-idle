package drill

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Kind is a potion kind of the belt plan: "hp" or "mp".
type Kind string

// View is the end state the D1 drills read through the StateReader (never the bot's log lines).
type View struct {
	Life, MaxLife, Mana, MaxMana, Level int
	BeltSlots                           int            // potion boxes the reader derives from the equipped belt (4 without one)
	Belt                                map[int]string // belt box (row*4+column) -> item code
	BeltRaw                             []string       // "gx,gy code" as the reader reports them
	Pack                                map[string]int // item code -> count in the pack grid
	Ground                              []string       // item codes on the ground
}

// BeltString is "0:hp1 1:hp1 3:mp1" in box order.
func (v View) BeltString() string {
	if len(v.Belt) == 0 {
		return "empty"
	}
	var box []int
	for b := range v.Belt {
		box = append(box, b)
	}
	sort.Ints(box)
	var s []string
	for _, b := range box {
		s = append(s, fmt.Sprintf("%d:%s", b, v.Belt[b]))
	}
	return strings.Join(s, " ")
}

// PackString is "hp1x3 mp1x1" (potions only, sorted).
func (v View) PackString() string { return countString(v.Pack) }

func countString(m map[string]int) string {
	var k []string
	for c, n := range m {
		if n > 0 {
			k = append(k, fmt.Sprintf("%sx%d", c, n))
		}
	}
	if len(k) == 0 {
		return "none"
	}
	sort.Strings(k)
	return strings.Join(k, " ")
}

// PotionKind maps an item code to "hp" or "mp" ("" for anything else): hp1..hp5 and mp1..mp5.
func PotionKind(code string) Kind {
	if len(code) == 3 && (code[:2] == "hp" || code[:2] == "mp") && code[2] >= '1' && code[2] <= '5' {
		return Kind(code[:2])
	}
	return ""
}

// PickInfo is the bot pickit's decision for one call.
type PickInfo struct {
	Picked  []string // "name (verdict)"
	Skipped []string // "name: reason"
	Failed  []string
	Seen    int
	Log     []string // the bot's log lines of the call (warnings and info), for the signature checks
}

func (p PickInfo) String() string {
	return fmt.Sprintf("seen %d, picked [%s], skipped [%s], failed [%s]", p.Seen, strings.Join(p.Picked, "; "), strings.Join(p.Skipped, "; "), strings.Join(p.Failed, "; "))
}

// FillInfo is the belt-fill chore's own report.
type FillInfo struct {
	Moved, Failed int
	Err           error
	Log           []string // the bot's log lines of the call
}

// BeltLab is the instance as the D1 drills use it. Setup helpers are game actions or logged cheats; FillBelt, Pickit
// and Drink call the bot's real functions (town.Town.FillBelt, loot.Looter.PickUp, combat Fighter.Act).
type BeltLab interface {
	View(ctx context.Context) (View, error)
	Tick(ctx context.Context, n int) error
	// PlanCols is the belt plan the town fill and the pickit use for the character (class, level, difficulty).
	PlanCols(ctx context.Context) ([4]Kind, error)

	GiveGround(ctx context.Context, r *Run, code string) error // cheat "debug give": an item at the player's feet
	ClearPack(ctx context.Context, r *Run) error               // cheat "debug clearinv"
	SetLife(ctx context.Context, r *Run, points int) error     // cheat "debug stat 6"
	SetMana(ctx context.Context, r *Run, points int) error     // cheat "debug stat 8"
	EquipBelt(ctx context.Context, r *Run, code string) error  // cheat "debug equip <code> 8"
	PickupRaw(ctx context.Context) (int, error)                // the engine's own "cmd pickup" on every ground item, in order
	DrainBelt(ctx context.Context) error                       // drink every belt potion with "cmd belt" until the belt is empty

	// ShiftClickPack is a game action with the click modifier verb: shift-left-click on the first pack item with this code.
	// The bot has no such verb on d2d (it uses "cmd belt put", D1.P1 covers that), so the drill drives the click itself.
	ShiftClickPack(ctx context.Context, code string) error

	FillBelt(ctx context.Context) (FillInfo, error)
	Pickit(ctx context.Context) (PickInfo, error)
	Drink(ctx context.Context, k Kind) error
}

const (
	missingMerc = "a cheat that lowers the merc's life: `debug merc hire|kill|revive` exist (D8.a), but a hired merc stands at full life and `debug hit` hits the player only, so a merc heal cannot be seen (missing: `debug merc hurt <dmg>`)"
)

// D1 returns the belt drill group: P1 belt fill, P2 belt rows by belt type, P3 drink by column, P7 ground potions.
// belts is the list of belt item codes P2 equips.
func D1(l BeltLab, belts []string) []Drill {
	return []Drill{
		{ID: "D1.P7a", Primitive: "P7", Title: "engine destination of a ground potion (empty belt)", Limit: 4 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return p7a(ctx, r, l) }},
		{ID: "D1.P7", Primitive: "P7", Title: "pickit takes a ground potion while the belt is empty and the pack holds 3", Limit: 4 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return p7(ctx, r, l) }},
		{ID: "D1.P1", Primitive: "P1", Title: "belt fill from the pack, right columns", Limit: 4 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return p1(ctx, r, l) }},
		{ID: "D1.P1b", Primitive: "P1", Title: "fix-belt2: sash, pack of 6 healing + 4 mana, empty belt: every planned kind keeps a column, no 'marked full' line", Limit: 4 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return p1b(ctx, r, l) }},
		{ID: "D1.P7b", Primitive: "P7", Title: "fix-belt2/fix-pickup: healing in 3 columns, healing on the ground: the potion does not take the mana column", Limit: 4 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return p7b(ctx, r, l) }},
		{ID: "D1.P1s", Primitive: "P1", Title: "shift-click pack to belt (click modifier verb, driven directly: the bot does not use it on d2d)", Limit: 4 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return p1s(ctx, r, l) }},
		{ID: "D1.P2", Primitive: "P2", Title: "belt rows by belt type, overflow stays in the pack", Limit: 12 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return p2(ctx, r, l, belts) }},
		{ID: "D1.P3", Primitive: "P3", Title: "drink by belt column", Limit: 4 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return p3(ctx, r, l) }},
		{ID: "D1.P3m", Primitive: "P3", Title: "merc drink by belt column", Needs: []string{missingMerc}},
	}
}

// testBelt is the belt every drill but P2 wears, so a drill does not depend on which drill ran before it (the engine
// has no verb to take a belt off): a Sash, 8 boxes in the engine (V, D1.P2).
const testBelt = "lbl"

// stagePack leaves an empty belt, hp1 and mp1 potions in the pack and nothing on the ground. The potions enter the pack
// because the belt is full when they are picked up (V: the engine puts a ground potion in the belt first); the belt is
// then drunk empty with the real belt verb.
func stagePack(ctx context.Context, r *Run, l BeltLab, hp, mp int) error {
	// loose potions an earlier drill left at the character's feet would be picked up with the staged ones (DRILLS8 F14:
	// P1s saw hp1x4 instead of x2): take them first, then ClearPack empties the pack again
	if _, err := l.PickupRaw(ctx); err != nil {
		return err
	}
	if err := l.ClearPack(ctx, r); err != nil {
		return err
	}
	if err := l.DrainBelt(ctx); err != nil {
		return err
	}
	if err := l.EquipBelt(ctx, r, testBelt); err != nil {
		return err
	}
	if err := l.DrainBelt(ctx); err != nil {
		return err
	}
	slots := 4
	if v, err := l.View(ctx); err == nil && v.BeltSlots > 0 {
		slots = v.BeltSlots
	}
	for i := 0; i < slots+hp+mp; i++ {
		code := "hp1"
		if i >= slots+hp {
			code = "mp1"
		}
		if err := l.GiveGround(ctx, r, code); err != nil {
			return err
		}
	}
	if _, err := l.PickupRaw(ctx); err != nil {
		return err
	}
	if err := l.DrainBelt(ctx); err != nil {
		return err
	}
	v, err := l.View(ctx)
	if err != nil {
		return err
	}
	if len(v.Belt) != 0 || len(v.Ground) != 0 || v.Pack["hp1"] != hp || v.Pack["mp1"] != mp {
		return fmt.Errorf("stage failed: belt %s, pack %s, ground %v (wanted pack hp1x%d mp1x%d)", v.BeltString(), v.PackString(), v.Ground, hp, mp)
	}
	return nil
}

func p7a(ctx context.Context, r *Run, l BeltLab) Outcome {
	if err := l.ClearPack(ctx, r); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	if err := l.DrainBelt(ctx); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	if err := l.EquipBelt(ctx, r, testBelt); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	for _, c := range []string{"hp1", "hp1", "mp1"} {
		if err := l.GiveGround(ctx, r, c); err != nil {
			return Failed("setup: "+err.Error(), "", "")
		}
	}
	if _, err := l.PickupRaw(ctx); err != nil {
		return Failed("engine pickup: "+err.Error(), "", "")
	}
	v, err := l.View(ctx)
	if err != nil {
		return Failed("read: "+err.Error(), "", "")
	}
	intended := "a ground potion picked up with an empty belt goes to the belt (SME) or to the pack: the destination is recorded"
	observed := fmt.Sprintf("picked hp1, hp1, mp1 in that order with the engine's own pickup: belt [%s], pack [%s], ground %d", v.BeltString(), v.PackString(), len(v.Ground))
	dest := "pack"
	if len(v.Belt) > 0 {
		dest = "belt"
	}
	if len(v.Belt)+countPotions(v.Pack)+len(v.Ground) != 3 || len(v.Ground) != 0 {
		return Failed("a potion was lost or stayed on the ground", intended, observed)
	}
	return Passed("observed destination: "+dest+" (engine rule seen: a potion of the kind already in a column goes above it in that column, a new kind takes the lowest free column; the second hp1 landed in box "+boxOf(v, "hp1", 1)+", mp1 in box "+boxOf(v, "mp1", 0)+")", intended, observed)
}

// boxOf is the n-th (0 based, box order) belt box holding code, "none" when there is no such box.
func boxOf(v View, code string, n int) string {
	var box []int
	for b, c := range v.Belt {
		if c == code {
			box = append(box, b)
		}
	}
	sort.Ints(box)
	if n < len(box) {
		return fmt.Sprint(box[n])
	}
	return "none"
}

func countPotions(m map[string]int) int {
	n := 0
	for c, k := range m {
		if PotionKind(c) != "" {
			n += k
		}
	}
	return n
}

func p7(ctx context.Context, r *Run, l BeltLab) Outcome {
	if err := stagePack(ctx, r, l, 3, 0); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	for i := 0; i < 2; i++ {
		if err := l.GiveGround(ctx, r, "hp1"); err != nil {
			return Failed("setup: "+err.Error(), "", "")
		}
	}
	if err := l.Tick(ctx, 25); err != nil {
		return Failed("tick: "+err.Error(), "", "")
	}
	info, perr := l.Pickit(ctx)
	if err := l.Tick(ctx, 50); err != nil {
		return Failed("tick: "+err.Error(), "", "")
	}
	v, err := l.View(ctx)
	if err != nil {
		return Failed("read: "+err.Error(), "", "")
	}
	intended := "belt empty, pack hp1x3, 2 hp1 on the ground -> both end in the belt (directly or after a fill), ground empty, and the pickit report counts them (picked > 0)"
	observed := fmt.Sprintf("belt [%s], pack [%s], ground %d; pickit: %s", v.BeltString(), v.PackString(), len(v.Ground), info)
	if perr != nil {
		observed += "; pickit error: " + perr.Error()
	}
	inBelt := 0
	for _, c := range v.Belt {
		if c == "hp1" {
			inBelt++
		}
	}
	switch {
	case inBelt == 2 && len(v.Ground) == 0 && len(info.Picked) == 0:
		return Failed("the potions are in the belt but the pickit report says picked 0 (arrival bookkeeping: debug log 'item left the ground without arriving in the pack')", intended, observed)
	case inBelt == 2 && len(v.Ground) == 0:
		return Passed(fmt.Sprintf("pickit took both potions, reports picked %d, and they are in the belt", len(info.Picked)), intended, observed)
	case len(v.Ground) > 0:
		return Failed("ground potions left behind with the belt empty: "+skipWhy(info), intended, observed)
	default:
		return Failed("potions were taken but are not in the belt", intended, observed)
	}
}

func skipWhy(i PickInfo) string {
	if len(i.Skipped) == 0 {
		return "no skip reason reported"
	}
	return strings.Join(i.Skipped, "; ")
}

func countKind(p [4]Kind, k Kind) int {
	n := 0
	for _, c := range p {
		if c == k {
			n++
		}
	}
	return n
}

func planString(p [4]Kind) string {
	var s []string
	for i, k := range p {
		s = append(s, fmt.Sprintf("%d:%s", i, k))
	}
	return strings.Join(s, " ")
}

func p1(ctx context.Context, r *Run, l BeltLab) Outcome {
	plan, err := l.PlanCols(ctx)
	if err != nil {
		return Failed("plan: "+err.Error(), "", "")
	}
	if err := l.EquipBelt(ctx, r, testBelt); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	v0, err := l.View(ctx)
	if err != nil {
		return Failed("read: "+err.Error(), "", "")
	}
	rows := max(v0.BeltSlots/4, 1)
	nHP, nMP := countKind(plan, "hp")*rows, countKind(plan, "mp")*rows // exactly what the plan has room for
	if err := stagePack(ctx, r, l, nHP, nMP); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	pre, err := l.View(ctx)
	if err != nil {
		return Failed("read: "+err.Error(), "", "")
	}
	intended := fmt.Sprintf("%s belt (%d boxes), plan columns [%s]: the %d hp1 and %d mp1 of the pack end in the belt, hp in hp columns and mp in mp columns, pack holds none of them", testBelt, v0.BeltSlots, planString(plan), nHP, nMP)
	fi, ferr := l.FillBelt(ctx)
	if err := l.Tick(ctx, 25); err != nil {
		return Failed("tick: "+err.Error(), intended, "")
	}
	v, err := l.View(ctx)
	if err != nil {
		return Failed("read: "+err.Error(), intended, "")
	}
	observed := fmt.Sprintf("before: pack [%s]; fill moved %d failed %d err %v; after: belt [%s], pack [%s]", pre.PackString(), fi.Moved, fi.Failed, ferr, v.BeltString(), v.PackString())
	if ferr != nil || fi.Err != nil {
		return Failed(fmt.Sprintf("fill reported an error: %v", firstErr(ferr, fi.Err)), intended, observed)
	}
	var bad []string
	for box, code := range v.Belt {
		col := box % 4
		if k := PotionKind(code); plan[col] != k {
			bad = append(bad, fmt.Sprintf("%s in column %d (plan %s)", code, col, plan[col]))
		}
	}
	sort.Strings(bad)
	if left := countPotions(v.Pack); left != 0 {
		why := fmt.Sprintf("%d potions are still in the pack with %d free belt boxes", left, v.BeltSlots-len(v.Belt))
		if len(bad) > 0 {
			why += "; potions in the wrong columns: " + strings.Join(bad, ", ")
		}
		return Failed(why, intended, observed)
	}
	if len(bad) > 0 {
		return Failed("potions in the wrong columns: "+strings.Join(bad, ", "), intended, observed)
	}
	if len(v.Belt) != nHP+nMP {
		return Failed(fmt.Sprintf("belt holds %d of %d potions", len(v.Belt), nHP+nMP), intended, observed)
	}
	return Passed("every potion belted in its plan column, pack empty of potions", intended, observed)
}

func p1s(ctx context.Context, r *Run, l BeltLab) Outcome {
	if err := l.EquipBelt(ctx, r, testBelt); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	if err := stagePack(ctx, r, l, 2, 0); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	intended := "empty belt, pack hp1x2: each shift-click on a pack hp1 moves exactly one to the belt (the pickup's box rule), the pack loses one each time; both end in the belt"
	var steps []string
	for i := 1; i <= 2; i++ {
		if err := l.ShiftClickPack(ctx, "hp1"); err != nil {
			return Failed(fmt.Sprintf("shift-click %d: %v", i, err), intended, strings.Join(steps, "; "))
		}
		if err := l.Tick(ctx, 10); err != nil {
			return Failed("tick: "+err.Error(), intended, "")
		}
		v, err := l.View(ctx)
		if err != nil {
			return Failed("read: "+err.Error(), intended, "")
		}
		steps = append(steps, fmt.Sprintf("after click %d: belt [%s], pack [%s]", i, v.BeltString(), v.PackString()))
		if len(v.Belt) != i || v.Pack["hp1"] != 2-i {
			return Failed(fmt.Sprintf("click %d did not move exactly one potion", i), intended, strings.Join(steps, "; "))
		}
	}
	return Passed("the click modifier verb belts one pack potion per shift-click", intended, strings.Join(steps, "; "))
}

func firstErr(a, b error) error {
	if a != nil {
		return a
	}
	return b
}

func p2(ctx context.Context, r *Run, l BeltLab, belts []string) Outcome {
	const extra = 6
	var obs, bad []string
	for _, code := range belts {
		if err := l.ClearPack(ctx, r); err != nil {
			return Failed("setup: "+err.Error(), "", "")
		}
		if err := l.DrainBelt(ctx); err != nil {
			return Failed("setup: "+err.Error(), "", "")
		}
		if err := l.EquipBelt(ctx, r, code); err != nil {
			return Failed("equip "+code+": "+err.Error(), "", "")
		}
		v, err := l.View(ctx)
		if err != nil {
			return Failed("read: "+err.Error(), "", "")
		}
		slots := v.BeltSlots
		total := slots + extra
		for i := 0; i < total; i++ {
			if err := l.GiveGround(ctx, r, "hp1"); err != nil {
				return Failed("setup: "+err.Error(), "", "")
			}
		}
		if _, err := l.PickupRaw(ctx); err != nil {
			return Failed("engine pickup: "+err.Error(), "", "")
		}
		v, err = l.View(ctx)
		if err != nil {
			return Failed("read: "+err.Error(), "", "")
		}
		inBelt, inPack := len(v.Belt), v.Pack["hp1"]
		s := fmt.Sprintf("%s: reader %d boxes (%d rows), engine belt took %d, pack %d, ground %d", code, slots, slots/4, inBelt, inPack, len(v.Ground))
		obs = append(obs, s)
		switch {
		case inBelt != slots:
			bad = append(bad, fmt.Sprintf("%s: engine belt holds %d, reader says %d boxes", code, inBelt, slots))
		case inBelt+inPack+len(v.Ground) != total:
			bad = append(bad, fmt.Sprintf("%s: %d potions lost", code, total-inBelt-inPack-len(v.Ground)))
		case inPack+len(v.Ground) != extra:
			bad = append(bad, fmt.Sprintf("%s: overflow %d, wanted %d", code, inPack+len(v.Ground), extra))
		}
	}
	intended := fmt.Sprintf("per belt type: potions fill exactly the boxes the reader reports, the %d extra stay (pack), none lost", extra)
	observed := strings.Join(obs, "; ")
	if len(bad) > 0 {
		return Failed(strings.Join(bad, "; "), intended, observed)
	}
	return Passed(fmt.Sprintf("%d belt types: reader boxes equal the engine's capacity, overflow kept, none lost", len(belts)), intended, observed)
}

func p3(ctx context.Context, r *Run, l BeltLab) Outcome {
	if err := l.ClearPack(ctx, r); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	if err := l.DrainBelt(ctx); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	if err := l.EquipBelt(ctx, r, testBelt); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	for _, c := range []string{"hp1", "hp1", "mp1", "mp1"} { // one at a time: the engine's pickup decides the boxes
		if err := l.GiveGround(ctx, r, c); err != nil {
			return Failed("setup: "+err.Error(), "", "")
		}
		if _, err := l.PickupRaw(ctx); err != nil {
			return Failed("setup pickup: "+err.Error(), "", "")
		}
	}
	v, err := l.View(ctx)
	if err != nil {
		return Failed("read: "+err.Error(), "", "")
	}
	if err := l.SetLife(ctx, r, v.MaxLife); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	if err := l.SetMana(ctx, r, v.MaxMana); err != nil { // the same start for every repeat
		return Failed("setup: "+err.Error(), "", "")
	}
	start := v.BeltString()
	if len(v.Belt) != 4 || countKindIn(v.Belt, "hp") != 2 || countKindIn(v.Belt, "mp") != 2 {
		return Failed("setup did not put 2 hp1 and 2 mp1 in the belt", "2 hp1 and 2 mp1 in the belt", start)
	}
	intended := "every belt potion can be drunk by kind: each drink removes exactly one potion of the asked kind, life or mana rises within 2 s, the other kind is untouched; with none of the kind left the drink reports an error"
	var steps, sigs []string
	var fails []string
	drink := func(k Kind) {
		before, err := l.View(ctx)
		if err != nil {
			fails = append(fails, "read: "+err.Error())
			return
		}
		if k == "hp" {
			err = l.SetLife(ctx, r, 10)
		} else {
			err = l.SetMana(ctx, r, 10)
		}
		if err != nil {
			fails = append(fails, "set low: "+err.Error())
			return
		}
		low, _ := l.View(ctx)
		derr := l.Drink(ctx, k)
		_ = l.Tick(ctx, 50) // 2 s of game time
		after, err := l.View(ctx)
		if err != nil {
			fails = append(fails, "read: "+err.Error())
			return
		}
		gone := removed(before.Belt, after.Belt)
		rise := after.Life - low.Life
		if k == "mp" {
			rise = after.Mana - low.Mana
		}
		steps = append(steps, fmt.Sprintf("drink %s: err=%v removed box %v, belt now [%s], life %d->%d mana %d->%d", k, derr, gone, after.BeltString(), low.Life, after.Life, low.Mana, after.Mana))
		sigs = append(sigs, fmt.Sprintf("%s err=%v removed %v rose=%v", k, derr != nil, gone, rise > 0))
		have := countKindIn(before.Belt, k)
		if have == 0 {
			if derr == nil {
				fails = append(fails, fmt.Sprintf("drink %s with none in the belt reported success", k))
			}
			return
		}
		switch {
		case derr != nil:
			fails = append(fails, fmt.Sprintf("drink %s (%d in the belt) failed: %v", k, have, derr))
		case countKindIn(after.Belt, k) != have-1 || len(after.Belt) != len(before.Belt)-1:
			fails = append(fails, fmt.Sprintf("drink %s: belt went from %d to %d potions", k, len(before.Belt), len(after.Belt)))
		case rise <= 0:
			fails = append(fails, fmt.Sprintf("drink %s: no rise within 2 s", k))
		}
	}
	for _, k := range []Kind{"hp", "mp", "hp", "mp", "hp"} { // the last one has none left
		drink(k)
	}
	end, _ := l.View(ctx)
	observed := fmt.Sprintf("start belt [%s]; %s; end belt [%s]", start, strings.Join(steps, " | "), end.BeltString())
	if len(end.Belt) != 0 {
		fails = append(fails, "belt not empty at the end: "+end.BeltString())
	}
	sig := start + " | " + strings.Join(sigs, " | ") + " | end " + end.BeltString() // life and mana regenerate between reads: not compared
	if len(fails) > 0 {
		o := Failed(strings.Join(fails, "; "), intended, observed)
		o.Sig = sig
		return o
	}
	o := Passed("every potion drunk once from its own column by kind, none left reported an error", intended, observed)
	o.Sig = sig
	return o
}

func countKindIn(belt map[int]string, k Kind) int {
	n := 0
	for _, c := range belt {
		if PotionKind(c) == k {
			n++
		}
	}
	return n
}

// removed lists the boxes that held an item before and hold none after.
func removed(before, after map[int]string) []int {
	var out []int
	for b := range before {
		if _, ok := after[b]; !ok {
			out = append(out, b)
		}
	}
	sort.Ints(out)
	return out
}

// beltGiver is the optional cheat "debug give <code> belt <box>" (a potion straight into a belt box).
type beltGiver interface {
	GiveBelt(ctx context.Context, r *Run, code string, box int) error
}

func logHas(lines []string, sub string) []string {
	var out []string
	for _, l := range lines {
		if strings.Contains(l, sub) {
			out = append(out, l)
		}
	}
	return out
}

// p1b is the fix-belt2 live proof: a Sorceress sash, 6 healing and 4 mana potions in the pack, an empty belt. The plan
// is hp hp mp mp: after the fill the belt must hold a mana column, no healing potion sits in a mana column, no
// "marked full" line is logged and nothing fails.
func p1b(ctx context.Context, r *Run, l BeltLab) Outcome {
	plan, err := l.PlanCols(ctx)
	if err != nil {
		return Failed("plan: "+err.Error(), "", "")
	}
	if err := l.EquipBelt(ctx, r, testBelt); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	if err := stagePack(ctx, r, l, 6, 4); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	pre, _ := l.View(ctx)
	intended := fmt.Sprintf("plan [%s]: the belt gets healing in the hp columns and mana in the mp columns (each planned kind keeps a column), no 'marked full' log line, no failed move", planString(plan))
	fi, ferr := l.FillBelt(ctx)
	if err := l.Tick(ctx, 25); err != nil {
		return Failed("tick: "+err.Error(), intended, "")
	}
	v, err := l.View(ctx)
	if err != nil {
		return Failed("read: "+err.Error(), intended, "")
	}
	marked := logHas(fi.Log, "marked full")
	observed := fmt.Sprintf("before pack [%s]; fill moved %d failed %d err %v; after belt [%s] pack [%s]; 'marked full' lines %d", pre.PackString(), fi.Moved, fi.Failed, ferr, v.BeltString(), v.PackString(), len(marked))
	var bad []string
	hp, mp := 0, 0
	for box, code := range v.Belt {
		k := PotionKind(code)
		if plan[box%4] != k {
			bad = append(bad, fmt.Sprintf("%s in column %d (plan %s)", code, box%4, plan[box%4]))
		}
		switch k {
		case "hp":
			hp++
		case "mp":
			mp++
		}
	}
	sort.Strings(bad)
	o := Outcome{Intended: intended, Observed: observed, Sig: fmt.Sprintf("hp=%d mp=%d bad=%d marked=%d failed=%d", hp, mp, len(bad), len(marked), fi.Failed)}
	switch {
	case ferr != nil:
		o.Verdict, o.Reason = Fail, "fill error: "+ferr.Error()
	case len(marked) > 0:
		o.Verdict, o.Reason = Fail, "the fill still logs a 'marked full' line: "+marked[0]
	case mp == 0:
		o.Verdict, o.Reason = Fail, "no mana potion in the belt: the mana column was lost"
	case len(bad) > 0:
		o.Verdict, o.Reason = Fail, "potions in the wrong columns: "+strings.Join(bad, ", ")
	case fi.Failed > 0:
		o.Verdict, o.Reason = Fail, fmt.Sprintf("%d failed moves", fi.Failed)
	default:
		o.Verdict, o.Reason = Pass, fmt.Sprintf("belt holds %d healing and %d mana, every kind in its planned column, no 'marked full' line", hp, mp)
	}
	return o
}

// p7b is the fix-pickup live proof: healing already fills the three healing-or-spare columns of the sash (both rows),
// a healing potion lies on the ground and the fourth column is the empty mana column. The engine would put it there; the
// pickit must leave it (or send it to the pack), never belt it into the mana column.
func p7b(ctx context.Context, r *Run, l BeltLab) Outcome {
	bg, ok := l.(beltGiver)
	if !ok {
		return Blocks("the lab has no `debug give <code> belt <box>` step")
	}
	if err := l.ClearPack(ctx, r); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	if err := l.DrainBelt(ctx); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	if err := l.EquipBelt(ctx, r, testBelt); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	if err := l.DrainBelt(ctx); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	for _, box := range []int{0, 1, 2, 4, 5, 6} {
		if err := bg.GiveBelt(ctx, r, "hp1", box); err != nil {
			return Failed("setup: "+err.Error(), "", "")
		}
	}
	if err := l.GiveGround(ctx, r, "hp1"); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	if err := l.Tick(ctx, 25); err != nil {
		return Failed("tick: "+err.Error(), "", "")
	}
	pre, _ := l.View(ctx)
	info, perr := l.Pickit(ctx)
	if err := l.Tick(ctx, 50); err != nil {
		return Failed("tick: "+err.Error(), "", "")
	}
	v, err := l.View(ctx)
	if err != nil {
		return Failed("read: "+err.Error(), "", "")
	}
	intended := "healing in columns 0 to 2 (both rows), a healing potion on the ground, column 3 empty: no healing potion ends in column 3 (left on the ground or in the pack)"
	observed := fmt.Sprintf("before belt [%s] ground %d; pickit: %s; after belt [%s] pack [%s] ground %d", pre.BeltString(), len(pre.Ground), info, v.BeltString(), v.PackString(), len(v.Ground))
	if perr != nil {
		observed += "; pickit error: " + perr.Error()
	}
	mana := 0
	for box, code := range v.Belt {
		if box%4 == 3 && PotionKind(code) == "hp" {
			mana++
		}
	}
	where := "left on the ground"
	if len(v.Ground) == 0 {
		where = "taken (pack " + v.PackString() + ")"
	}
	o := Outcome{Intended: intended, Observed: observed, Sig: fmt.Sprintf("col3hp=%d", mana)}
	if mana > 0 {
		o.Verdict, o.Reason = Fail, "the ground healing potion went into the mana column (column 3)"
	} else {
		o.Verdict, o.Reason = Pass, "column 3 stays free of healing: the potion was "+where
	}
	return o
}
