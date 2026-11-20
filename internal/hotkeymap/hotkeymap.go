// Package hotkeymap generates the hero's skill hotkey map from the planner's slots and writes it into the game's own
// hotkey table (the 16 slots of the character save) through a Binder, then reads it back (ADR 0021, Fable review
// 2026-11-19 07-input-layer.md 4.3 and I7). The map is data per hero: nothing here names a skill, a class or a key.
//
// Status: U until a live check on a d2d build; the unit tests run on a fake Binder, no save file is ever written here.
package hotkeymap

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"d2lod-bot/internal/agency"
)

// Slots is the size of the game's hotkey table (the character save header holds 16 entries; docs/design/fable-review-2026-11-19/07-input-layer.md 4.3, S).
const Slots = 16

// Unbound is what a slot reads back when nothing is bound (the d2d state line prints 65535 for an empty slot; V: the headless start prints it for all 16).
const Unbound = 65535

// Roles of a slot, in the order Generate fills them.
const (
	RoleMain     = "main"
	RoleAttack   = "attack"
	RoleSecond   = "second element"
	RoleMovement = "movement"
	RoleBuff     = "buff"
	RoleAura     = "aura"
	RoleSummon   = "summon"
	RoleTP       = "town portal"
)

// Slot is one hotkey: its index in the table, why it is there and the skill.
type Slot struct {
	Index int
	Role  string
	Skill string
	ID    int
}

// Map is the generated table, filled from index 0 without gaps.
type Map struct {
	Slots []Slot
}

// Options are what Generate needs from the caller.
type Options struct {
	// Resolve maps a skill name to its Skills.txt id (db.SkillByName on the real bot). Required.
	Resolve func(name string) (int, bool)
	// TownPortal is the Skills.txt row name of the skill that opens a portal from the scroll or tome ("" leaves it out).
	TownPortal string
}

// Generate orders the plan's skills: main attack, the attack when it differs (summoner), second element, movement,
// buffs in the plan's upkeep order, aura, summons, town portal. A skill appears once. A skill the plan names that
// Resolve does not know is an error naming it, never a silent gap. More than Slots skills: the first Slots stay.
func Generate(pl agency.Plan, o Options) (Map, error) {
	if o.Resolve == nil {
		return Map{}, errors.New("hotkeymap: no skill resolver")
	}
	var m Map
	seen := map[string]bool{}
	var firstErr error
	add := func(role string, c *agency.Choice) {
		if c == nil || c.Skill == "" {
			return
		}
		addName(&m, seen, &firstErr, o, role, c.Skill)
	}
	add(RoleMain, pl.Main)
	if pl.Main == nil || pl.Attack == nil || pl.Attack.Skill != pl.Main.Skill {
		add(RoleAttack, pl.Attack)
	}
	add(RoleSecond, pl.Secondary)
	for i := range pl.Movement {
		add(RoleMovement, &pl.Movement[i])
	}
	for i := range pl.Buffs {
		add(RoleBuff, &pl.Buffs[i])
	}
	add(RoleAura, pl.Aura)
	for i := range pl.Summons {
		add(RoleSummon, &pl.Summons[i])
	}
	if o.TownPortal != "" {
		addName(&m, seen, &firstErr, o, RoleTP, o.TownPortal)
	}
	if firstErr != nil {
		return Map{}, firstErr
	}
	return m, nil
}

func addName(m *Map, seen map[string]bool, firstErr *error, o Options, role, name string) {
	key := strings.ToLower(name)
	if seen[key] || len(m.Slots) >= Slots {
		return
	}
	id, ok := o.Resolve(name)
	if !ok {
		if *firstErr == nil {
			*firstErr = fmt.Errorf("hotkeymap: skill %q (%s) is not in the skill table", name, role)
		}
		return
	}
	seen[key] = true
	m.Slots = append(m.Slots, Slot{Index: len(m.Slots), Role: role, Skill: name, ID: id})
}

// Binder is the supported route into the hotkey table of one backend. Bind uses the game's own mechanism for the
// slot (the skill popup with hover and the F key, today; a bind-slot message when a build has one) and Read reports
// the table as the game state shows it.
type Binder interface {
	// Writable is how many slots from index 0 the route can write (F1 to F8 on the popup route).
	Writable() int
	Bind(ctx context.Context, slot, skillID int) error
	Read(ctx context.Context) ([Slots]int, error)
}

// Result says what Apply did.
type Result struct {
	Written int // slots bound by this call
	Already int // slots that already held the wanted skill
	Beyond  int // map slots the route cannot write (Writable too small)
}

// ErrReadBack is returned when a slot does not read back as written.
var ErrReadBack = errors.New("hotkeymap: read-back differs")

// Apply writes the map into the hotkey table and reads the whole table back. Slots that already hold the skill are
// not touched; slots beyond Writable are counted, not written; a slot that does not read back as the wanted skill
// is an error (ErrReadBack, the slot named). Slots the map does not cover are left alone.
func Apply(ctx context.Context, b Binder, m Map) (Result, error) {
	var res Result
	if b == nil {
		return res, errors.New("hotkeymap: no binder")
	}
	cur, err := b.Read(ctx)
	if err != nil {
		return res, fmt.Errorf("hotkeymap: read before: %w", err)
	}
	limit := b.Writable()
	var wrote []Slot
	for _, s := range m.Slots {
		switch {
		case s.Index >= limit:
			res.Beyond++
		case cur[s.Index] == s.ID:
			res.Already++
			wrote = append(wrote, s)
		default:
			if err := b.Bind(ctx, s.Index, s.ID); err != nil {
				return res, fmt.Errorf("hotkeymap: bind slot %d (%s): %w", s.Index, s.Skill, err)
			}
			res.Written++
			wrote = append(wrote, s)
		}
	}
	after, err := b.Read(ctx)
	if err != nil {
		return res, fmt.Errorf("hotkeymap: read back: %w", err)
	}
	for _, s := range wrote {
		if after[s.Index] != s.ID {
			return res, fmt.Errorf("%w: slot %d (%s) reads %d, wanted %d", ErrReadBack, s.Index, s.Skill, after[s.Index], s.ID)
		}
	}
	return res, nil
}

// ParseState reads the 16 slots from a d2d "state" line (` hotkeys=a,b,...,p `). A line without the field, with
// fewer than Slots values or a non-number is an error.
func ParseState(line string) ([Slots]int, error) {
	var out [Slots]int
	_, rest, ok := strings.Cut(line, "hotkeys=")
	if !ok {
		return out, errors.New("hotkeymap: no hotkeys field in the state")
	}
	field := strings.Fields(rest)
	if len(field) == 0 {
		return out, errors.New("hotkeymap: empty hotkeys field")
	}
	parts := strings.Split(field[0], ",")
	if len(parts) < Slots {
		return out, fmt.Errorf("hotkeymap: %d hotkey values, want %d", len(parts), Slots)
	}
	for i := 0; i < Slots; i++ {
		var n int
		if _, err := fmt.Sscanf(parts[i], "%d", &n); err != nil {
			return out, fmt.Errorf("hotkeymap: slot %d %q: %w", i, parts[i], err)
		}
		out[i] = n
	}
	return out, nil
}
