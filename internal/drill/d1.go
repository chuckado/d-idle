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
}

func (p PickInfo) String() string {
	return fmt.Sprintf("seen %d, picked [%s], skipped [%s], failed [%s]", p.Seen, strings.Join(p.Picked, "; "), strings.Join(p.Skipped, "; "), strings.Join(p.Failed, "; "))
}

// FillInfo is the belt-fill chore's own report.
type FillInfo struct {
	Moved, Failed int
	Err           error
}

// BeltLab is the instance as the D1 drills use it. Setup helpers are game actions or logged cheats; FillBelt, Pickit
// and Drink call the bot's real functions (town.Town.FillBelt, loot.Looter.PickUp, combat Fighter.Act).
type BeltLab interface {
	View(ctx context.Context) (View, error)
	Tick(ctx context.Context, n int) error
	PlanCols(level int) [4]Kind

	GiveGround(ctx context.Context, r *Run, code string) error // cheat "debug give": an item at the player's feet
	ClearPack(ctx context.Context, r *Run) error               // cheat "debug clearinv"
	SetLife(ctx context.Context, r *Run, points int) error     // cheat "debug stat 6"
	SetMana(ctx context.Context, r *Run, points int) error     // cheat "debug stat 8"
	EquipBelt(ctx context.Context, r *Run, code string) error  // cheat "debug equip <code> 8"
	PickupRaw(ctx context.Context) (int, error)                // the engine's own "cmd pickup" on every ground item, in order
	DrainBelt(ctx context.Context) error                       // drink every belt potion with "cmd belt" until the belt is empty

	FillBelt(ctx context.Context) (FillInfo, error)
	Pickit(ctx context.Context) (PickInfo, error)
	Drink(ctx context.Context, k Kind) error
}

const (
	missingModifier = "fork verb `click <x> <y> shift|ctrl` (click modifiers): d2d has no shift/ctrl-click behaviour at all"
	missingMerc     = "cheat/verb to hire or revive a mercenary on a fresh character (Kashya's reward needs the Blood Raven quest chain)"
)

// D1 returns the belt drill group: P1 belt fill, P2 belt rows by belt type, P3 drink by column, P7 ground potions.
// belts is the list of belt item codes P2 equips.
func D1(l BeltLab, belts []string) []Drill {
	return []Drill{
		{ID: "D1.P7a", Primitive: "P7", Title: "engine destination of a ground potion (empty belt)", Limit: 4 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return p7a(ctx, r, l) }},
		{ID: "D1.P7", Primitive: "P7", Title: "pickit takes a ground potion while the belt is empty and the pack holds 3", Limit: 4 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return p7(ctx, r, l) }},
		{ID: "D1.P1", Primitive: "P1", Title: "belt fill from the pack, right columns", Limit: 4 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return p1(ctx, r, l) }},
		{ID: "D1.P1s", Primitive: "P1", Title: "shift-click pack to belt", Needs: []string{missingModifier}},
		{ID: "D1.P2", Primitive: "P2", Title: "belt rows by belt type, overflow stays in the pack", Limit: 12 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return p2(ctx, r, l, belts) }},
		{ID: "D1.P3", Primitive: "P3", Title: "drink by belt column", Limit: 4 * time.Minute, Do: func(ctx context.Context, r *Run) Outcome { return p3(ctx, r, l) }},
		{ID: "D1.P3m", Primitive: "P3", Title: "merc drink by belt column", Needs: []string{missingMerc}},
	}
}

// stagePack leaves an empty belt, hp1 and mp1 potions in the pack and nothing on the ground. The potions enter the pack
// because the belt is full when they are picked up (V: the engine puts a ground potion in the belt first); the belt is
// then drunk empty with the real belt verb.
func stagePack(ctx context.Context, r *Run, l BeltLab, hp, mp int) error {
	if err := l.ClearPack(ctx, r); err != nil {
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
	for _, c := range []string{"hp1", "mp1", "hp1"} {
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
	observed := fmt.Sprintf("picked hp1, mp1, hp1 in that order: belt [%s], pack [%s], ground %d", v.BeltString(), v.PackString(), len(v.Ground))
	dest := "pack"
	if len(v.Belt) > 0 {
		dest = "belt"
	}
	if len(v.Belt)+countPotions(v.Pack)+len(v.Ground) != 3 || len(v.Ground) != 0 {
		return Failed("a potion was lost or stayed on the ground", intended, observed)
	}
	return Passed("observed destination: "+dest+" (engine fills the lowest free box, not the plan column: mp1 landed in box "+boxOf(v, "mp1")+", the plan puts mana in column "+planCol(l.PlanCols(v.Level), "mp")+")", intended, observed)
}

func boxOf(v View, code string) string {
	for b, c := range v.Belt {
		if c == code {
			return fmt.Sprint(b)
		}
	}
	return "none"
}

func planCol(p [4]Kind, k Kind) string {
	for i, c := range p {
		if c == k {
			return fmt.Sprint(i)
		}
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
	intended := "belt empty, pack hp1x3, 2 hp1 on the ground -> both end in the belt (directly or after a fill), ground empty"
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
	case inBelt == 2 && len(v.Ground) == 0:
		return Passed("pickit took both potions and they are in the belt", intended, observed)
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

func p1(ctx context.Context, r *Run, l BeltLab) Outcome {
	if err := stagePack(ctx, r, l, 3, 1); err != nil {
		return Failed("setup: "+err.Error(), "", "")
	}
	pre, err := l.View(ctx)
	if err != nil {
		return Failed("read: "+err.Error(), "", "")
	}
	plan := l.PlanCols(pre.Level)
	intended := fmt.Sprintf("plan columns %v: belt 0:hp1 1:hp1 2:mp1 3:hp1 (hp in hp columns, mp in the mp column), pack holds none of them", plan)
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
	if left := countPotions(v.Pack); left != 0 {
		return Failed(fmt.Sprintf("%d potions are still in the pack", left), intended, observed)
	}
	var bad []string
	for box, code := range v.Belt {
		col := box % 4
		if k := PotionKind(code); plan[col] != k {
			bad = append(bad, fmt.Sprintf("%s in column %d (plan %s)", code, col, plan[col]))
		}
	}
	sort.Strings(bad)
	if len(bad) > 0 {
		return Failed("potions in the wrong columns: "+strings.Join(bad, ", "), intended, observed)
	}
	if len(v.Belt) != 4 {
		return Failed(fmt.Sprintf("belt holds %d of 4 potions", len(v.Belt)), intended, observed)
	}
	return Passed("every potion belted in its plan column, pack empty of potions", intended, observed)
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
	for _, c := range []string{"hp1", "hp1", "mp1", "hp1"} { // pickup order gives the plan layout: boxes 0..3 = hp hp mp hp
		if err := l.GiveGround(ctx, r, c); err != nil {
			return Failed("setup: "+err.Error(), "", "")
		}
	}
	if _, err := l.PickupRaw(ctx); err != nil {
		return Failed("setup pickup: "+err.Error(), "", "")
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
	if start != "0:hp1 1:hp1 2:mp1 3:hp1" {
		return Failed("setup did not give the plan layout", "belt 0:hp1 1:hp1 2:mp1 3:hp1", start)
	}
	intended := "each drink removes exactly one potion of the asked kind from its column, life or mana rises within 2 s, the other kind is untouched; with no healing potion left the drink reports an error"
	var steps []string
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
		steps = append(steps, fmt.Sprintf("drink %s: err=%v removed box %v, life %d->%d mana %d->%d", k, derr, gone, low.Life, after.Life, low.Mana, after.Mana))
		hasKind := false
		for _, c := range before.Belt {
			if PotionKind(c) == k {
				hasKind = true
			}
		}
		if !hasKind {
			if derr == nil {
				fails = append(fails, fmt.Sprintf("drink %s with none in the belt reported success", k))
			}
			return
		}
		switch {
		case derr != nil:
			fails = append(fails, fmt.Sprintf("drink %s failed: %v", k, derr))
		case len(gone) != 1:
			fails = append(fails, fmt.Sprintf("drink %s removed %d potions", k, len(gone)))
		case PotionKind(before.Belt[gone[0]]) != k:
			fails = append(fails, fmt.Sprintf("drink %s consumed %s from box %d", k, before.Belt[gone[0]], gone[0]))
		case rise <= 0:
			fails = append(fails, fmt.Sprintf("drink %s: no rise within 2 s", k))
		}
	}
	for i := 0; i < 3; i++ {
		drink("hp")
	}
	drink("mp")
	drink("hp") // none left
	end, _ := l.View(ctx)
	observed := fmt.Sprintf("start belt [%s]; %s; end belt [%s]", start, strings.Join(steps, " | "), end.BeltString())
	if len(end.Belt) != 0 {
		fails = append(fails, "belt not empty at the end: "+end.BeltString())
	}
	if len(fails) > 0 {
		return Failed(strings.Join(fails, "; "), intended, observed)
	}
	return Passed("3 healing columns and the mana column each drunk once from the right box, none left reported an error", intended, observed)
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
