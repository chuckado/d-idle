package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/d2dmerc"
	"d2lod-bot/internal/drill"
	"d2lod-bot/internal/input"
	"d2lod-bot/internal/leveler"
	"d2lod-bot/internal/nav"
)

func (l *lab) Field(ctx context.Context, r *drill.Run, level int) error { return l.Goto(ctx, r, level) }

func (l *lab) Invincible(ctx context.Context, r *drill.Run) error {
	for _, v := range []string{"debug stat 7 512000", "debug stat 6 512000"} {
		if _, err := r.Cheat(ctx, v, "life 2000 (24.8 fixed point): the walking drills cross monster areas with the bot's navigator and no combat, the throwaway must survive"); err != nil {
			return err
		}
	}
	return l.Tick(ctx, 10)
}

func (l *lab) beltKinds(s *d2.Snapshot) (hp, mp int) {
	for _, u := range s.Inventory.Container(d2.ContBelt) {
		switch drill.PotionKind(l.code(u.Class)) {
		case "hp":
			hp++
		case "mp":
			mp++
		}
	}
	return
}

func (l *lab) ArmBelt(ctx context.Context, r *drill.Run) (string, error) {
	s, err := l.be.Env.Snapshot()
	if err != nil || s.Inventory == nil {
		return "", fmt.Errorf("no inventory: %v", err)
	}
	hp, mp := l.beltKinds(s)
	if mp == 0 {
		// the starting belt holds one healing potion in each of its 4 columns: a mana potion has no column. Drink the
		// last column (a game action) so the engine belts the mana potion there.
		if err := l.be.Client.Drink(ctx, 3); err != nil {
			return "", fmt.Errorf("freeing a belt column: %w", err)
		}
		_, _ = l.be.Client.Step(ctx, 40)
		if err := l.GiveGround(ctx, r, "mp1"); err != nil {
			return "", err
		}
		_, _ = l.be.Client.Step(ctx, 30)
		if _, err := l.PickupRaw(ctx); err != nil {
			return "", err
		}
		if s, err = l.be.Env.Snapshot(); err == nil {
			hp, mp = l.beltKinds(s)
		}
	}
	return fmt.Sprintf("belt hp %d mp %d", hp, mp), nil
}

func (l *lab) GiveSkills(ctx context.Context, r *drill.Run) error {
	for _, v := range []string{"debug points 36 1", "debug points 38 1"} {
		if _, err := r.Cheat(ctx, v, "learn the skill at level 1: the bind and cast drill needs known skills"); err != nil {
			return err
		}
	}
	return l.Tick(ctx, 10)
}

func (l *lab) BindCount(ctx context.Context) (int, error) { return l.rig.BindSkills() }

func (l *lab) mana() (int, error) {
	s, err := l.be.Env.Snapshot()
	if err != nil || s.Player == nil {
		return 0, fmt.Errorf("no player: %v", err)
	}
	return s.Player.Mana, nil
}

func (l *lab) CastMana(ctx context.Context, skill int) (int, int, error) {
	if _, err := l.rig.BindSkills(); err != nil {
		return 0, 0, err
	}
	s, err := l.be.Env.Snapshot()
	if err != nil || s.Player == nil {
		return 0, 0, fmt.Errorf("no player: %v", err)
	}
	m0 := s.Player.Mana
	if err := l.rig.In.CastAt(fmt.Sprintf("sk%d", skill), input.Point{X: s.Player.X + 10, Y: s.Player.Y}); err != nil {
		return m0, m0, err
	}
	_, _ = l.be.Client.Step(ctx, 25)
	m1, err := l.mana()
	return m0, m1, err
}

func (l *lab) SafetyAt(ctx context.Context, r *drill.Run, life, mana int) (string, error) {
	if err := l.SetLife(ctx, r, life); err != nil {
		return "", err
	}
	if err := l.SetMana(ctx, r, mana); err != nil {
		return "", err
	}
	// the Fighter's potion gap runs on the env clock, which only moves through Clock.Sleep (Client.Step bypasses it)
	l.be.Env.Clock.Sleep(1500 * time.Millisecond)
	s, err := l.be.Env.Snapshot()
	if err != nil {
		return "", err
	}
	return l.rig.Fighter.Safety(s).String(), nil
}

func (l *lab) ActDrink(ctx context.Context, action string) (int, int, error) {
	a := combat.DrinkLife
	if action == "drink-mana" {
		a = combat.DrinkMana
	}
	s, err := l.be.Env.Snapshot()
	if err != nil || s.Player == nil {
		return 0, 0, fmt.Errorf("no player: %v", err)
	}
	b := s.Player.Life
	if a == combat.DrinkMana {
		b = s.Player.Mana
	}
	aerr := l.rig.Fighter.Act(ctx, a)
	_, _ = l.be.Client.Step(ctx, 40)
	s, err = l.be.Env.Snapshot()
	if err != nil || s.Player == nil {
		return b, b, fmt.Errorf("no player: %v", err)
	}
	after := s.Player.Life
	if a == combat.DrinkMana {
		after = s.Player.Mana
	}
	return b, after, aerr
}

func (l *lab) Walks(ctx context.Context, n int) []drill.WalkInfo {
	var out []drill.WalkInfo
	s, err := l.be.Env.Snapshot()
	if err != nil || s.Player == nil {
		return out
	}
	ox, oy := s.Player.X, s.Player.Y
	var cands []input.Point
	for _, rad := range []float64{30, 50, 70, 90, 110} {
		for k := 0; k < 12; k++ {
			a := float64(k)*math.Pi/6 + rad/50
			cands = append(cands, input.Point{X: ox + int(rad*math.Cos(a)), Y: oy + int(rad*math.Sin(a))})
		}
	}
	for _, p := range cands {
		if len(out) >= n {
			break
		}
		wctx, cancel := context.WithTimeout(ctx, 120*time.Second)
		t0 := l.tick()
		werr := l.navr.GoTo(wctx, p)
		cancel()
		w := drill.WalkInfo{Tx: p.X, Ty: p.Y, Ticks: l.tick() - t0}
		if s, err := l.be.Env.Snapshot(); err == nil && s.Player != nil {
			w.Fx, w.Fy = s.Player.X, s.Player.Y
			w.Cells = math.Hypot(float64(p.X-s.Player.X), float64(p.Y-s.Player.Y)) / 5
		}
		if errors.Is(werr, nav.ErrNoPath) { // a wall or off-map cell, not a walk: pick another target
			continue
		}
		if werr != nil {
			w.Err = werr.Error()
		}
		out = append(out, w)
	}
	return out
}

func (l *lab) Exits(ctx context.Context, levels []int) []drill.ExitInfo {
	var out []drill.ExitInfo
	for _, want := range levels {
		e := drill.ExitInfo{Want: want}
		if s, err := l.be.Env.Snapshot(); err == nil && s.Player != nil {
			e.From = s.Player.Area
		}
		wctx, cancel := context.WithTimeout(ctx, 4*time.Minute)
		t0 := l.tick()
		err := l.navr.ToArea(wctx, want)
		cancel()
		e.Ticks = l.tick() - t0
		if err != nil {
			e.Err = err.Error()
		}
		if s, err := l.be.Env.Snapshot(); err == nil && s.Player != nil {
			e.To = s.Player.Area
		}
		out = append(out, e)
	}
	return out
}

func (l *lab) LevelUp(ctx context.Context, r *drill.Run, clvl int) error {
	for _, v := range []string{fmt.Sprintf("debug stat 12 %d", clvl), fmt.Sprintf("debug statpts %d", 5*(clvl-1)), fmt.Sprintf("debug skillpts %d", clvl-1)} {
		if _, err := r.Cheat(ctx, v, fmt.Sprintf("a character of level %d with its stat and skill points unspent: the build drill", clvl)); err != nil {
			return err
		}
	}
	return l.Tick(ctx, 10)
}

func (l *lab) leveler() *leveler.Leveler {
	if l.lv == nil {
		l.lv = leveler.New(l.be.Env, leveler.Options{Rebind: func(context.Context) error { _, err := l.rig.BindSkills(); return err }})
	}
	return l.lv
}

func (l *lab) Spend(ctx context.Context) drill.SpendInfo {
	var si drill.SpendInfo
	s, err := l.be.Env.Snapshot()
	if err != nil || s.Player == nil {
		si.Err = fmt.Sprint("no player: ", err)
		return si
	}
	rep := l.leveler().Do(ctx, s)
	si.StatBefore, si.StatLeft, si.SkillBefore, si.SkillLeft = rep.StatBefore, rep.StatLeft, rep.SkillBefore, rep.SkillLeft
	si.Planned, si.Done, si.Refused = len(rep.Planned), len(rep.Done), rep.Refused
	if rep.Err != nil {
		si.Err = rep.Err.Error()
	}
	if rep.Skipped != "" && si.Err == "" {
		si.Err = "skipped: " + rep.Skipped
	}
	_, _ = l.be.Client.Step(ctx, 20)
	if s, err := l.be.Env.Snapshot(); err == nil && s.Player != nil {
		si.Str, si.Dex, si.Vit, si.Ene = s.Player.Strength, s.Player.Dexterity, s.Player.Vitality, s.Player.Energy
		var sk []string
		for _, k := range s.Player.Skills {
			if k.Level > 0 {
				sk = append(sk, fmt.Sprintf("%d:%d", k.ID, k.Level))
			}
		}
		si.Skills = strings.Join(sk, ",")
		if si.StatLeft == 0 && si.StatBefore > 0 && s.Player.StatPoints != 0 {
			si.StatLeft = s.Player.StatPoints
		}
		if si.SkillLeft == 0 && si.SkillBefore > 0 && s.Player.SkillPoints != 0 {
			si.SkillLeft = s.Player.SkillPoints
		}
	}
	return si
}

func (l *lab) WornCodes(ctx context.Context) (string, error) {
	r, err := l.be.Client.Do(ctx, "dump items")
	if err != nil {
		return "", err
	}
	line := strings.TrimSpace(r.Rest)
	for _, ln := range r.Lines {
		if strings.TrimSpace(ln) != "" {
			line = ln
			break
		}
	}
	var raw struct {
		Items []struct {
			Code, Loc string
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return "", err
	}
	var c []string
	for _, it := range raw.Items {
		if it.Loc == "body" {
			c = append(c, it.Code)
		}
	}
	sort.Strings(c)
	if len(c) == 0 {
		return "nothing worn", nil
	}
	return strings.Join(c, ","), nil
}

func (l *lab) EquipCap(ctx context.Context, r *drill.Run) (bool, string, error) {
	if err := l.ClearPack(ctx, r); err != nil {
		return false, "", err
	}
	if err := l.GiveGround(ctx, r, "cap"); err != nil {
		return false, "", err
	}
	_, _ = l.be.Client.Step(ctx, 30)
	if _, err := l.PickupRaw(ctx); err != nil {
		return false, "", err
	}
	w0, _ := l.WornCodes(ctx)
	s, err := l.be.Env.Snapshot()
	if err != nil {
		return false, "", err
	}
	rep := l.leveler().Equip(ctx)
	_, _ = l.be.Client.Step(ctx, 20)
	w1, _ := l.WornCodes(ctx)
	msg := fmt.Sprintf("worn before the equip call %s (pack %d items), swaps %v failed %v err %v", w0, len(s.Inventory.Container(d2.ContInventory)), rep.Swaps, rep.Failed, rep.Err)
	return strings.Contains(w1, "cap"), msg, nil
}

func (l *lab) DieAndRespawn(ctx context.Context, r *drill.Run) (string, error) {
	if _, err := r.Cheat(ctx, "debug die", "kill the throwaway in the field: the death and corpse drill"); err != nil {
		return "", err
	}
	_, _ = l.be.Client.Step(ctx, 40)
	s, _ := l.be.Env.Snapshot()
	dead := s != nil && s.Player != nil && combat.IsDead(s.Player)
	rerr := l.rig.Fighter.Respawn(ctx, 40*time.Second)
	_, _ = l.be.Client.Step(ctx, 40)
	s, _ = l.be.Env.Snapshot()
	area, life := -1, -1
	if s != nil && s.Player != nil {
		area, life = s.Player.Area, s.Player.Life
	}
	desc := fmt.Sprintf("dead=%v, Respawn err %v, then area %d life %d", dead, rerr, area, life)
	if rerr != nil {
		return desc, rerr
	}
	return desc, nil
}

func (l *lab) TakeCorpse(ctx context.Context, r *drill.Run) (string, error) {
	// the corpse lies where the character died: the arrival cell of the field goto used before the death
	c := l.be.Client
	before, _ := c.Corpses(ctx)
	if err := l.Goto(ctx, r, 2); err != nil {
		return "", err
	}
	mid, _ := c.Corpses(ctx)
	if err := c.CorpseInteract(ctx, 0); err != nil {
		return fmt.Sprintf("corpses %d/%d", len(before.Items), len(mid.Items)), err
	}
	_, _ = c.Step(ctx, 200)
	after, _ := c.Corpses(ctx)
	return fmt.Sprintf("corpses listed %d (town) %d (field) -> %d after cmd interact -3000", len(before.Items), len(mid.Items), len(after.Items)), nil
}

func (l *lab) Cycles(ctx context.Context, n int) (string, error) {
	var out []string
	for i := 0; i < n; i++ {
		t0 := time.Now()
		if err := l.be.Cycler.LeaveGame(ctx); err != nil {
			return strings.Join(out, "; "), fmt.Errorf("cycle %d leave: %w", i+1, err)
		}
		if err := l.be.Cycler.EnterGame(ctx, ""); err != nil {
			return strings.Join(out, "; "), fmt.Errorf("cycle %d enter: %w", i+1, err)
		}
		_, _ = l.be.Client.Step(ctx, 30)
		v, err := l.TView(ctx)
		if err != nil {
			return strings.Join(out, "; "), fmt.Errorf("cycle %d read: %w", i+1, err)
		}
		out = append(out, fmt.Sprintf("cycle %d: in town %v, panels %s, pack %s, %.0fs", i+1, v.InTown, panelsOf(v.Panels), v.PackString(), time.Since(t0).Seconds()))
		if !v.InTown || len(v.Panels) > 0 {
			return strings.Join(out, "; "), fmt.Errorf("cycle %d did not end in town with a clean UI", i+1)
		}
	}
	return strings.Join(out, "; "), nil
}

var _ drill.GameLab = (*lab)(nil)

func panelsOf(p []string) string {
	if len(p) == 0 {
		return "none"
	}
	return strings.Join(p, ",")
}

func (l *lab) MercCheat(ctx context.Context, r *drill.Run, verb string) error {
	_, err := r.Cheat(ctx, verb, "the merc drills need a mercenary in a known state: no other way to hire, kill or revive one on a fresh character")
	return err
}

func (l *lab) MercState(ctx context.Context) (drill.MercInfo, error) {
	r, err := l.be.Client.Do(ctx, "dump merc")
	if err != nil {
		return drill.MercInfo{}, err
	}
	line := strings.TrimSpace(r.Rest)
	for _, ln := range r.Lines {
		if strings.TrimSpace(ln) != "" {
			line = ln
			break
		}
	}
	var raw struct {
		Merc struct {
			Present bool   `json:"present"`
			Dead    bool   `json:"dead"`
			Code    string `json:"code"`
			Life    [2]int `json:"life"`
		} `json:"merc"`
	}
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return drill.MercInfo{}, err
	}
	m := raw.Merc
	return drill.MercInfo{Present: m.Present, Dead: m.Dead, Code: m.Code, Life: m.Life[0], Max: m.Life[1]}, nil
}

func (l *lab) Revive(ctx context.Context) (int, error) {
	rv := d2dmerc.Reviver{C: l.be.Client}
	return rv.Revive(ctx)
}

func (l *lab) DrinkMerc(ctx context.Context, col int) error { return l.be.Ctl.Belt(col, true) }
