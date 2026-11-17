package questboard

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"d2lod-bot/internal/store"
)

// PlanVersion is the on-disk version of the plan-record file.
const PlanVersion = 1

// Record is an accepted Start tap. It is a plan record, not handler progress: it says the player consented to
// this chunk, never how far the handler got (design 1.5).
type Record struct {
	Quest string    `json:"quest"`
	Chunk string    `json:"chunk"`
	Diff  string    `json:"difficulty"`
	At    time.Time `json:"at"`
	Mode  string    `json:"mode,omitempty"` // the mode that was active (train, idle, farm)
}

// SkipRecord is a player Skip. Chunk "" skips the whole quest.
type SkipRecord struct {
	Quest string    `json:"quest"`
	Chunk string    `json:"chunk,omitempty"`
	Diff  string    `json:"difficulty"`
	At    time.Time `json:"at"`
}

// AutoRecord is the opt-in to auto-start, per quest and difficulty (default off).
type AutoRecord struct {
	Quest string    `json:"quest"`
	Diff  string    `json:"difficulty"`
	Since time.Time `json:"since"`
}

// Plan is the whole per-hero store: only the player's decisions, never derived progress.
type Plan struct {
	Accepted []Record     `json:"accepted,omitempty"`
	Skipped  []SkipRecord `json:"skipped,omitempty"`
	Auto     []AutoRecord `json:"autoStart,omitempty"`
	Seen     []string     `json:"seen,omitempty"` // payoff toasts already shown
	Mode     string       `json:"mode,omitempty"` // the between-chunks mode the player picked (S4 stores it, S6 acts on it)
}

// Modes are the between-chunks modes (user decision 2026-11-16); the default is Train.
var Modes = []string{"train", "idle", "farm"}

// DefaultMode is used while the player never picked one.
const DefaultMode = "train"

// ValidMode reports whether m is one of Modes.
func ValidMode(m string) bool {
	for _, x := range Modes {
		if x == m {
			return true
		}
	}
	return false
}

func (p Plan) accepted(diff int, quest, chunk string) bool {
	for _, r := range p.Accepted {
		if r.Quest == quest && r.Chunk == chunk && r.Diff == Difficulties[diff] {
			return true
		}
	}
	return false
}

func (p Plan) skipped(diff int, quest, chunk string) bool {
	for _, r := range p.Skipped {
		if r.Quest == quest && r.Diff == Difficulties[diff] && (r.Chunk == "" || r.Chunk == chunk) {
			return true
		}
	}
	return false
}

// touched reports whether the player ever started anything of the quest in this difficulty (the board ran it).
func (p Plan) touched(diff int, quest string) bool {
	for _, r := range p.Accepted {
		if r.Quest == quest && r.Diff == Difficulties[diff] {
			return true
		}
	}
	return false
}

// AutoStart reports the opt-in for a quest in a difficulty.
func (p Plan) AutoStart(diff int, quest string) bool {
	for _, r := range p.Auto {
		if r.Quest == quest && r.Diff == Difficulties[diff] {
			return true
		}
	}
	return false
}

// check rejects records that name an unknown quest, chunk or difficulty (an old record after a data change).
func (p Plan) check(b *Board) error {
	for _, r := range p.Accepted {
		if _, err := b.Chunk(r.Quest, r.Chunk); err != nil {
			return fmt.Errorf("plan record: %w", err)
		}
		if _, err := DiffIndex(r.Diff); err != nil {
			return fmt.Errorf("plan record: %w", err)
		}
	}
	for _, r := range p.Skipped {
		if r.Chunk == "" {
			if _, err := b.Quest(r.Quest); err != nil {
				return fmt.Errorf("plan skip: %w", err)
			}
		} else if _, err := b.Chunk(r.Quest, r.Chunk); err != nil {
			return fmt.Errorf("plan skip: %w", err)
		}
		if _, err := DiffIndex(r.Diff); err != nil {
			return fmt.Errorf("plan skip: %w", err)
		}
	}
	for _, r := range p.Auto {
		if _, err := b.Quest(r.Quest); err != nil {
			return fmt.Errorf("plan autoStart: %w", err)
		}
		if _, err := DiffIndex(r.Diff); err != nil {
			return fmt.Errorf("plan autoStart: %w", err)
		}
	}
	return nil
}

var heroName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{1,23}$`)

// Store keeps one hero's plan records in <dir>/<hero>.quests.json, written atomically with a .bak
// (internal/store.File). Calls are serialised.
type Store struct {
	b    *Board
	file store.File[Plan]
	now  func() time.Time
	mu   sync.Mutex
}

// Open returns the store of one hero. The hero name is a Diablo II character name (2 to 24 characters), so
// it can never escape dir.
func Open(b *Board, dir, hero string) (*Store, error) {
	if !heroName.MatchString(hero) {
		return nil, fmt.Errorf("questboard: bad hero name %q", hero)
	}
	return &Store{b: b, now: time.Now, file: store.File[Plan]{
		Path: filepath.Join(dir, hero+".quests.json"), Version: PlanVersion,
	}}, nil
}

// Load returns the records and what the load did (recovery from .bak is reported, a future version is an
// error). A record naming an id the data no longer has is an error too, naming it.
func (s *Store) Load() (Plan, store.Info, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load()
}

func (s *Store) load() (Plan, store.Info, error) {
	p, info, err := s.file.Load()
	if err != nil {
		return Plan{}, info, err
	}
	return p, info, p.check(s.b)
}

func (s *Store) update(f func(*Plan) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, _, err := s.load()
	if err != nil {
		return err
	}
	if err := f(&p); err != nil {
		return err
	}
	return s.file.Save(p)
}

// Accept records the player's Start tap. Callers check CanStart against the derived views first; here the ids
// are validated and a second accepted chunk in the same difficulty is refused ("never two at once", design
// 2.6). Repeating the same tap is a no-op.
func (s *Store) Accept(quest, chunk string, diff int, mode string) error {
	return s.AcceptIf(quest, chunk, diff, mode, nil)
}

// AcceptIf is Accept where a record for which finished returns true (a chunk that is done, kept as the
// proof that the board ran it) no longer counts as running. A nil finished counts every record.
func (s *Store) AcceptIf(quest, chunk string, diff int, mode string, finished func(Record) bool) error {
	if diff < 0 || diff > 2 {
		return fmt.Errorf("questboard: difficulty index %d out of range", diff)
	}
	if _, err := s.b.Chunk(quest, chunk); err != nil {
		return err
	}
	return s.update(func(p *Plan) error {
		for _, r := range p.Accepted {
			if r.Diff != Difficulties[diff] {
				continue
			}
			if r.Quest == quest && r.Chunk == chunk {
				return nil
			}
			if finished != nil && finished(r) {
				continue
			}
			return fmt.Errorf("questboard: %s/%s is already accepted in %s", r.Quest, r.Chunk, r.Diff)
		}
		p.Accepted = append(p.Accepted, Record{quest, chunk, Difficulties[diff], s.now().UTC(), mode})
		return nil
	})
}

// Cancel withdraws an accepted chunk ("Stop this quest", back to ready). Unknown ids are errors; cancelling a
// chunk that is not accepted is a no-op.
func (s *Store) Cancel(quest, chunk string, diff int) error {
	if _, err := s.b.Chunk(quest, chunk); err != nil {
		return err
	}
	return s.update(func(p *Plan) error {
		out := p.Accepted[:0]
		for _, r := range p.Accepted {
			if !(r.Quest == quest && r.Chunk == chunk && r.Diff == Difficulties[diff]) {
				out = append(out, r)
			}
		}
		p.Accepted = out
		return nil
	})
}

// Skip records a player Skip (chunk "" for the whole quest); Unskip removes it.
func (s *Store) Skip(quest, chunk string, diff int) error {
	if err := s.checkIDs(quest, chunk, diff); err != nil {
		return err
	}
	return s.update(func(p *Plan) error {
		for _, r := range p.Skipped {
			if r.Quest == quest && r.Chunk == chunk && r.Diff == Difficulties[diff] {
				return nil
			}
		}
		p.Skipped = append(p.Skipped, SkipRecord{quest, chunk, Difficulties[diff], s.now().UTC()})
		return nil
	})
}

func (s *Store) Unskip(quest, chunk string, diff int) error {
	if err := s.checkIDs(quest, chunk, diff); err != nil {
		return err
	}
	return s.update(func(p *Plan) error {
		out := p.Skipped[:0]
		for _, r := range p.Skipped {
			if !(r.Quest == quest && r.Chunk == chunk && r.Diff == Difficulties[diff]) {
				out = append(out, r)
			}
		}
		p.Skipped = out
		return nil
	})
}

// SetAutoStart turns the opt-in on or off for a quest in one difficulty. It refuses a quest whose data says
// autoStart never (design 1.1).
func (s *Store) SetAutoStart(quest string, diff int, on bool) error {
	q, err := s.b.Quest(quest)
	if err != nil {
		return err
	}
	if diff < 0 || diff > 2 {
		return fmt.Errorf("questboard: difficulty index %d out of range", diff)
	}
	if on && q.AutoStart != "allowed" {
		return fmt.Errorf("questboard: quest %q does not allow auto-start", quest)
	}
	return s.update(func(p *Plan) error {
		out := p.Auto[:0]
		for _, r := range p.Auto {
			if !(r.Quest == quest && r.Diff == Difficulties[diff]) {
				out = append(out, r)
			}
		}
		p.Auto = out
		if on {
			p.Auto = append(p.Auto, AutoRecord{quest, Difficulties[diff], s.now().UTC()})
		}
		return nil
	})
}

// MarkSeen records that a payoff toast key was shown (so it is not shown again after a restart).
func (s *Store) MarkSeen(key string) error {
	return s.update(func(p *Plan) error {
		for _, k := range p.Seen {
			if k == key {
				return nil
			}
		}
		p.Seen = append(p.Seen, key)
		return nil
	})
}

func (s *Store) checkIDs(quest, chunk string, diff int) error {
	if diff < 0 || diff > 2 {
		return fmt.Errorf("questboard: difficulty index %d out of range", diff)
	}
	if chunk == "" {
		_, err := s.b.Quest(quest)
		return err
	}
	_, err := s.b.Chunk(quest, chunk)
	return err
}

// Mode returns the stored mode, DefaultMode when none was picked.
func (s *Store) Mode() (string, error) {
	p, _, err := s.Load()
	if err != nil {
		return "", err
	}
	if p.Mode == "" {
		return DefaultMode, nil
	}
	return p.Mode, nil
}

// SetMode stores the mode; an unknown one is an error naming it. Setting the stored mode again is a no-op write.
func (s *Store) SetMode(m string) error {
	if !ValidMode(m) {
		return fmt.Errorf("questboard: unknown mode %q (train, idle or farm)", m)
	}
	return s.update(func(p *Plan) error { p.Mode = m; return nil })
}
