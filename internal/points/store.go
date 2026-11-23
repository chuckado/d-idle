package points

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Event is one action that came through the API: who (a client class and a user-agent family, never an address or a
// token), what, and how it ended. The same shape as the quest board's audit trail (docs/questboard.md "Audit").
type Event struct {
	At      time.Time `json:"at"`
	Action  string    `json:"action"` // stat, skill, autopilot
	Target  string    `json:"target,omitempty"`
	Path    string    `json:"path"`
	Remote  string    `json:"remote"`
	Agent   string    `json:"agent,omitempty"`
	Confirm bool      `json:"confirm,omitempty"`
	OK      bool      `json:"ok"`
	Status  int       `json:"status"`
	Code    string    `json:"code,omitempty"`
	Reason  string    `json:"reason,omitempty"`
	Before  int       `json:"before,omitempty"`
	After   int       `json:"after,omitempty"`
}

// AuditMax is how many events the ring file keeps.
const AuditMax = 200

// Audit is <hero>.points.audit.jsonl in the state directory.
type Audit struct {
	path string
	mu   sync.Mutex
}

// NewAudit opens (lazily) the audit ring of a hero; dir "" gives a no-op audit.
func NewAudit(dir, hero string) *Audit {
	if dir == "" || safeName(hero) == "" {
		return &Audit{}
	}
	return &Audit{path: filepath.Join(dir, safeName(hero)+".points.audit.jsonl")}
}

// Append adds one event and trims the file to AuditMax lines.
func (a *Audit) Append(e Event) error {
	if a.path == "" {
		return nil
	}
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(a.path), 0o700); err != nil {
		return err
	}
	old, _ := os.ReadFile(a.path)
	var lines [][]byte
	if len(bytes.TrimSpace(old)) > 0 {
		lines = bytes.Split(bytes.TrimRight(old, "\n"), []byte("\n"))
	}
	lines = append(lines, line)
	if len(lines) > AuditMax {
		lines = lines[len(lines)-AuditMax:]
	}
	tmp := a.path + ".tmp"
	if err := os.WriteFile(tmp, append(bytes.Join(lines, []byte("\n")), '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, a.path)
}

// Recent returns up to n events, newest first.
func (a *Audit) Recent(n int) []Event {
	if a.path == "" {
		return nil
	}
	a.mu.Lock()
	f, err := os.Open(a.path)
	a.mu.Unlock()
	if err != nil {
		return nil
	}
	defer f.Close()
	var all []Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		var e Event
		if json.Unmarshal(sc.Bytes(), &e) == nil {
			all = append(all, e)
		}
	}
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}
	if n > 0 && len(all) > n {
		all = all[:n]
	}
	return all
}

// Settings is the per-hero file <hero>.points.json (the hero's plan file; additive fields only, so older files read as
// "no build"). Autopilot is the player's per-hero switch: off (the default) means points are banked and the player is
// asked; on means the bot may spend them from its plan table. Build, Locked, LockedAtLevel and Respecs are the
// automatic allocator's state (docs/autoalloc.md): a locked build stays until a paid respec (a later card) clears it.
type Settings struct {
	Autopilot     bool      `json:"autopilot"`
	Changed       time.Time `json:"changed,omitzero"`
	Build         string    `json:"build,omitempty"`
	Locked        bool      `json:"locked,omitempty"`
	LockedAtLevel int       `json:"lockedAtLevel,omitempty"`
	LockedAt      time.Time `json:"lockedAt,omitzero"`
	Respecs       int       `json:"respecs,omitempty"`
}

// Store reads and writes Settings; safe for concurrent use.
type Store struct {
	path string
	mu   sync.Mutex
	mem  Settings
}

// NewStore opens the settings of a hero; dir "" keeps them in memory only.
func NewStore(dir, hero string) *Store {
	s := &Store{}
	if dir != "" && safeName(hero) != "" {
		s.path = filepath.Join(dir, safeName(hero)+".points.json")
	}
	return s
}

func (s *Store) read() Settings {
	if s.path == "" {
		return s.mem
	}
	var v Settings
	if b, err := os.ReadFile(s.path); err == nil {
		_ = json.Unmarshal(b, &v)
	}
	return v
}

// Get returns the settings; a missing or unreadable file is the default (autopilot off, no build).
func (s *Store) Get() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.read()
}

// update applies f to the settings under the lock and saves them when f reports a change.
func (s *Store) update(f func(*Settings) bool) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur := s.read()
	if !f(&cur) {
		return false, nil
	}
	if s.path == "" {
		s.mem = cur
		return true, nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return false, err
	}
	b, _ := json.MarshalIndent(cur, "", " ")
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o600); err != nil {
		return false, err
	}
	return true, os.Rename(tmp, s.path)
}

// SetAutopilot stores the switch and reports whether it changed.
func (s *Store) SetAutopilot(on bool, now time.Time) (changed bool, err error) {
	return s.update(func(c *Settings) bool {
		if c.Autopilot == on {
			return false
		}
		c.Autopilot, c.Changed = on, now
		return true
	})
}

// LockBuild stores the chosen build and locks it. Locking the same build again changes nothing.
func (s *Store) LockBuild(name string, level int, now time.Time) (changed bool, err error) {
	return s.update(func(c *Settings) bool {
		if c.Locked && c.Build == name {
			return false
		}
		c.Build, c.Locked, c.LockedAtLevel, c.LockedAt, c.Changed = name, true, level, now, now
		return true
	})
}

// ClearBuild is what a paid respec does (the respec card calls it after the game took the refund): the build and the
// lock go, the hero's respec count goes up by one.
func (s *Store) ClearBuild(now time.Time) error {
	_, err := s.update(func(c *Settings) bool {
		c.Build, c.Locked, c.LockedAtLevel, c.LockedAt, c.Changed = "", false, 0, time.Time{}, now
		c.Respecs++
		return true
	})
	return err
}

// safeName keeps a hero name usable as a file stem; anything else gives "".
func safeName(h string) string {
	h = strings.TrimSpace(h)
	if h == "" || len(h) > 40 {
		return ""
	}
	for _, c := range h {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return ""
		}
	}
	return h
}
