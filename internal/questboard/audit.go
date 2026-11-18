package questboard

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"
)

// AuditMax is the size of the audit ring: the last this many board actions per hero.
const AuditMax = 500

// StoppedWindow is how long a surprise Stop stays visible on the board.
const StoppedWindow = 24 * time.Hour

// AuditEvent is one action that came through the quest API (docs/questboard.md "Audit"). It carries no credential:
// the source is a class, never an address or a token.
type AuditEvent struct {
	At      time.Time `json:"at"`
	Action  string    `json:"action"` // start, handin, cancel, retry, auto, mode, dismiss
	Quest   string    `json:"quest,omitempty"`
	Chunk   string    `json:"chunk,omitempty"`
	Path    string    `json:"path"`
	Remote  string    `json:"remote"`            // loopback, lan or other
	Agent   string    `json:"agent,omitempty"`   // user agent family
	Confirm bool      `json:"confirm,omitempty"` // the two-tap confirm flag was on the request
	OK      bool      `json:"ok"`
	Already bool      `json:"already,omitempty"`
	Status  int       `json:"status"`
	Code    string    `json:"code,omitempty"`
	Reason  string    `json:"reason,omitempty"`
}

// Audit is the ring file <hero>.audit.jsonl next to the plan file. Concurrent use is safe within one process.
type Audit struct {
	path string
	mu   sync.Mutex
}

var (
	auditMu  sync.Mutex
	auditFor = map[string]*Audit{}
)

// AuditOf returns the ring that belongs to a plan store (one Audit per file, so writers share the lock).
func AuditOf(st *Store) *Audit {
	if st == nil {
		return nil
	}
	p := strings.TrimSuffix(st.file.Path, ".quests.json") + ".audit.jsonl"
	auditMu.Lock()
	defer auditMu.Unlock()
	a := auditFor[p]
	if a == nil {
		a = &Audit{path: p}
		auditFor[p] = a
	}
	return a
}

// Append adds one event and trims the file to AuditMax lines.
func (a *Audit) Append(e AuditEvent) error {
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	old, _ := os.ReadFile(a.path)
	lines := bytes.Split(bytes.TrimRight(old, "\n"), []byte("\n"))
	if len(old) == 0 {
		lines = nil
	}
	lines = append(lines, line)
	if len(lines) > AuditMax {
		lines = lines[len(lines)-AuditMax:]
	}
	out := append(bytes.Join(lines, []byte("\n")), '\n')
	tmp := a.path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, a.path)
}

// Recent returns the newest events first (n <= 0: all). A line that does not parse is skipped.
func (a *Audit) Recent(n int) []AuditEvent {
	a.mu.Lock()
	f, err := os.Open(a.path)
	a.mu.Unlock()
	if err != nil {
		return nil
	}
	defer f.Close()
	var all []AuditEvent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		var e AuditEvent
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

// StoppedRow is a quest the player stopped through the API in the last day (View.Stopped).
type StoppedRow struct {
	Quest      string    `json:"quest"`
	QuestTitle string    `json:"questTitle"`
	Chunk      string    `json:"chunk,omitempty"`
	At         time.Time `json:"at"`
}

// Record appends the event of an API action to the hero's ring and refreshes the board. It does nothing without a hero.
func (d *Desk) Record(e AuditEvent) {
	if e.At.IsZero() {
		e.At = d.cfg.Now().UTC()
	}
	_, st := d.exec()
	if a := AuditOf(st); a != nil {
		_ = a.Append(e)
	}
	d.invalidate()
}

// stopped lists the real stops (a cancel that stopped something) of the last day, newest first, one per quest, unless
// the quest runs again or is done.
func (d *Desk) stopped(st *Store, v View) []StoppedRow {
	a := AuditOf(st)
	if a == nil {
		return nil
	}
	now := d.cfg.Now()
	seen := map[string]bool{}
	var out []StoppedRow
	for _, e := range a.Recent(0) {
		if e.Action != "cancel" || !e.OK || e.Already || e.Quest == "" || seen[e.Quest] || now.Sub(e.At) > StoppedWindow {
			continue
		}
		seen[e.Quest] = true
		if v.Running != nil && v.Running.Quest == e.Quest || v.Awaiting != nil && v.Awaiting.Quest == e.Quest || v.Blocked != nil && v.Blocked.Quest == e.Quest {
			continue
		}
		done := false
		for _, q := range v.Quests {
			if q.ID == e.Quest && q.State == Done {
				done = true
			}
		}
		if !done {
			out = append(out, StoppedRow{Quest: e.Quest, QuestTitle: d.title(e.Quest), Chunk: e.Chunk, At: e.At})
		}
	}
	return out
}

// AuditPathOf is the ring file of a store (for tests and the support doc).
func AuditPathOf(st *Store) string {
	if a := AuditOf(st); a != nil {
		return a.path
	}
	return ""
}

// ExecForTest returns what the desk's Exec hook returns (the watch tests read the ring through it).
func (d *Desk) ExecForTest() (*Exec, *Store) { return d.exec() }
