package combat

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"d2lod-bot/internal/d2"
	"d2lod-bot/internal/input"
)

// LeftPopupGeometry places the icons of the LEFT-skill popup on the 800x600 client. VERIFIED LIVE (CLASSPAL, Paladin
// Palli, docs/classes/smoke-paladin.md): the button is at (140,578); the grid grows to the right and up from x=104;
// the bottom row (y=480) holds Attack, the row above (y=433) the Paladin combat skills in tree order Sacrifice 104,
// Smite 153, Holy Bolt 202, Zeal 251 (pitch 49). Only skills with leftskill=1 are listed, so rows are counted among
// those. The row spacing of 47 above row 0 is measured once; every pick is verified by reading LeftSkill back.
type LeftPopupGeometry struct {
	BtnX, BtnY int
	Col0X      int // x of the first (leftmost) icon column
	ColDX      int // column pitch, to the right
	Row0Y      int // y of the bottom row
	RowDY      int // row pitch, upward
}

// LeftPopup is the geometry SetLeft uses.
var LeftPopup = LeftPopupGeometry{BtnX: 140, BtnY: 578, Col0X: 104, ColDX: 49, Row0Y: 480, RowDY: 47}

// leftCells returns the candidate screen cells of skill id in the left popup, most likely first: the cell the
// tree position predicts (row, then column order within the ListRow), then the rest of its row.
func (f *Fighter) leftCells(p *d2.Player, id int) []input.Screen {
	type ent struct{ id, row, tr, tc int }
	byRow := map[int][]ent{}
	haveAttack := false
	for _, sk := range p.Skills {
		def := f.cat.Skill(sk.ID)
		r, ok := f.listRows[sk.ID]
		if def == nil || !ok || r < 0 || !(def.LeftSkill || sk.ID == 0) {
			continue
		}
		haveAttack = haveAttack || sk.ID == 0
		byRow[r] = append(byRow[r], ent{sk.ID, r, def.Row, def.Column})
	}
	row, ok := f.listRows[id]
	// The basic Attack is on every class's popup (bottom row, first column) but the d2d dump lists only trained skills
	// and the selected one, so it is absent from p.Skills while another skill is selected (live: a Paladin out of mana
	// never swung: "skill 0 has no left popup cell"). VERIFIED LIVE for the Paladin popup (geometry comment above);
	// other classes U, the tree order puts it first in its row for all of them.
	if !haveAttack {
		ar, aok := f.listRows[0]
		if !aok || ar < 0 {
			ar = 1 << 30
			for r := range byRow {
				ar = min(ar, r)
			}
			if ar == 1<<30 {
				ar = 0
			}
		}
		byRow[ar] = append(byRow[ar], ent{0, ar, -1, -1})
		if id == 0 {
			row, ok = ar, true
		}
	}
	if !ok || row < 0 {
		return nil
	}
	if _, ok := byRow[row]; !ok {
		return nil
	}
	var rows []int
	for r := range byRow {
		rows = append(rows, r)
	}
	sort.Ints(rows)
	idx := sort.SearchInts(rows, row)
	es := byRow[row]
	sort.Slice(es, func(i, j int) bool {
		if es[i].tr != es[j].tr {
			return es[i].tr < es[j].tr
		}
		if es[i].tc != es[j].tc {
			return es[i].tc < es[j].tc
		}
		return es[i].id < es[j].id
	})
	g := LeftPopup
	y := g.Row0Y - idx*g.RowDY
	var first, rest []input.Screen
	for c, e := range es {
		cell := input.Screen{X: g.Col0X + c*g.ColDX, Y: y}
		if e.id == id {
			first = append(first, cell)
		} else {
			rest = append(rest, cell)
		}
	}
	return append(first, rest...)
}

func (f *Fighter) openLeftPopup(ctx context.Context) error {
	if err := f.e.In.Click(input.Left, LeftPopup.BtnX, LeftPopup.BtnY); err != nil {
		return err
	}
	return f.sleep(ctx, 400*time.Millisecond)
}

// readHands returns the current left and right skill ids after the game had time to apply a pick.
func (f *Fighter) readHands(ctx context.Context) (left, right int, err error) {
	if err = f.sleep(ctx, 200*time.Millisecond); err != nil {
		return
	}
	s, err := f.read(ctx)
	if err != nil {
		return
	}
	return s.Player.LeftSkill, s.Player.RightSkill, nil
}

// SetLeft makes skill id the left-hand skill through the left popup and verifies it by reading LeftSkill back. A cell
// that gave another skill is not an error, the next candidate is tried. The right hand (the aura) is never touched.
func (f *Fighter) SetLeft(ctx context.Context, id int) error {
	snap, err := f.read(ctx)
	if err != nil {
		return err
	}
	p := snap.Player
	if p.LeftSkill == id {
		return nil
	}
	if f.e.Planner.On() {
		return f.setPlanned(ctx, "select-left", id, true, func() error {
			if len(f.listRows) == 0 {
				return fmt.Errorf("%w: popup rows not loaded (SetListRows)", ErrBindFailed)
			}
			cells := f.leftCells(p, id)
			if c, ok := f.leftCell[id]; ok {
				cells = append([]input.Screen{c}, cells...)
			}
			if len(cells) == 0 {
				return fmt.Errorf("%w: skill %d has no left popup cell", ErrBindFailed, id)
			}
			return f.leftPopupPick(ctx, id, cells)
		}, nil)
	}
	if len(f.listRows) == 0 {
		return fmt.Errorf("%w: popup rows not loaded (SetListRows)", ErrBindFailed)
	}
	cells := f.leftCells(p, id)
	if c, ok := f.leftCell[id]; ok { // a cell that worked before goes first
		cells = append([]input.Screen{c}, cells...)
	}
	if len(cells) == 0 {
		return fmt.Errorf("%w: skill %d has no left popup cell", ErrBindFailed, id)
	}
	return f.leftPopupPick(ctx, id, cells)
}

// leftPopupPick opens the left popup and clicks each candidate cell until the hand reads back as id.
func (f *Fighter) leftPopupPick(ctx context.Context, id int, cells []input.Screen) error {
	tried := map[input.Screen]bool{}
	for _, c := range cells {
		if tried[c] {
			continue
		}
		tried[c] = true
		if err := f.openLeftPopup(ctx); err != nil {
			return err
		}
		if err := f.e.In.Click(input.Left, c.X, c.Y); err != nil { // picking an icon closes the popup
			return err
		}
		left, _, err := f.readHands(ctx)
		if err != nil {
			return err
		}
		if left == id {
			f.leftCell[id] = c
			f.info("left skill set", "skill", id, "cell", c)
			return nil
		}
	}
	return fmt.Errorf("%w: left skill %d not found in %d popup cells", ErrBindFailed, id, len(tried))
}

// SetRight makes skill id the right-hand skill: by its verified hotkey when it has one (about 200 ms, no mana), else
// through the right popup. The left hand is not touched.
func (f *Fighter) SetRight(ctx context.Context, id int) error {
	snap, err := f.read(ctx)
	if err != nil {
		return err
	}
	p := snap.Player
	if p.RightSkill == id {
		return nil
	}
	if f.e.Planner.On() {
		var key func() error
		if k := f.keys[id]; k != "" && !f.D2D {
			key = func() error { return f.e.In.Key(k) }
		}
		return f.setPlanned(ctx, "select-right", id, false, func() error { return f.rightPopupPick(ctx, p, id) }, key)
	}
	if key := f.keys[id]; key != "" && !f.D2D { // d2d has pseudo keys only: the popup is the way
		if err := f.e.In.Key(key); err != nil {
			return err
		}
		if _, r, err := f.readHands(ctx); err != nil || r == id {
			return err
		}
	}
	return f.rightPopupPick(ctx, p, id)
}

// rightPopupPick clicks the candidate cells of the right popup until the hand reads back as id.
func (f *Fighter) rightPopupPick(ctx context.Context, p *d2.Player, id int) error {
	for _, c := range f.popupCells(p, id) {
		if err := f.openPopup(ctx); err != nil {
			return err
		}
		if err := f.e.In.Click(input.Left, c.X, c.Y); err != nil {
			return err
		}
		if _, r, err := f.readHands(ctx); err != nil || r == id {
			return err
		}
	}
	return fmt.Errorf("%w: right skill %d", ErrBindFailed, id)
}

// setPlanned selects skill id on a hand through the input planner: the select verb when the backend advertises it,
// else key (nil: none), else click, each verified by reading the hand back (07-input-layer.md 4.2).
func (f *Fighter) setPlanned(ctx context.Context, name string, id int, left bool, click, key func() error) error {
	a := input.Action{Name: name, Cap: input.CapSelect, Key: key, Click: click,
		Verify: func() (bool, error) {
			l, r, err := f.readHands(ctx)
			if left {
				return l == id, err
			}
			return r == id, err
		}}
	if sel, ok := f.e.In.(input.SkillSelector); ok {
		a.Verb = func() error { return sel.SelectSkillID(id, left) }
	}
	path, err := f.e.Planner.Run(a)
	if err != nil {
		if errors.Is(err, input.ErrNoPath) && !errors.Is(err, ErrBindFailed) {
			return fmt.Errorf("%w: %s %d: %v", ErrBindFailed, name, id, err)
		}
		return err
	}
	f.info("skill selected", "hand", name, "skill", id, "path", string(path))
	return nil
}
