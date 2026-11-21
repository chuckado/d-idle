package charsel

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/saves"
)

// Schemas of the replies (additive; fixtures in internal/watch/testdata/characters).
const (
	ListSchema    = "watch.characters/1"
	ResultSchema  = "watch.characters.result/1"
	IntentSchema  = "watch.charswitch/1"
	intentFile    = "charswitch.json"
	switchTimeout = 30 * time.Second
)

// Refusal is a named no. Status is the HTTP status the route answers with.
type Refusal struct {
	Status int    `json:"-"`
	Code   string `json:"code"`
	Error  string `json:"error"`
}

func no(status int, code, msg string) *Refusal { return &Refusal{status, code, msg} }

// Intent is the file between the bot that leaves and the supervisor that starts the next character.
type Intent struct {
	Schema     string `json:"schema"`
	From       string `json:"from"`
	To         string `json:"to"`
	Create     bool   `json:"create,omitempty"`
	ClassIndex int    `json:"classIndex"` // d2d --start-class of a created character; -1 = load the save
	At         string `json:"at"`
}

// ReadIntent returns the pending switch of an instance, if any.
func ReadIntent(inst string) (Intent, bool) {
	b, err := os.ReadFile(filepath.Join(inst, intentFile))
	var in Intent
	if err != nil || json.Unmarshal(b, &in) != nil || in.Schema != IntentSchema || in.To == "" {
		return Intent{}, false
	}
	return in, true
}

// staleAfter: an intent whose bot never exited (a crash between the write and the exit) must not block every verb.
const staleAfter = 5 * time.Minute

// pending is ReadIntent that drops a stale intent.
func (m *Manager) pending() (Intent, bool) {
	in, ok := ReadIntent(m.Inst)
	if !ok {
		return in, false
	}
	if t, err := time.Parse(time.RFC3339, in.At); err != nil || m.now().Sub(t) > staleAfter {
		ClearIntent(m.Inst)
		return Intent{}, false
	}
	return in, true
}

// ClearIntent removes it (the supervisor does, after it has swapped the state).
func ClearIntent(inst string) { _ = os.Remove(filepath.Join(inst, intentFile)) }

func writeIntent(inst string, in Intent) error {
	b, _ := json.Marshal(in)
	tmp := filepath.Join(inst, intentFile+".tmp")
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(inst, intentFile))
}

// Caps tells the page what this backend can do, so it can hide or explain a verb instead of failing.
type Caps struct {
	Create  bool   `json:"create"`
	Switch  bool   `json:"switch"`
	Delete  bool   `json:"delete"`
	Restore bool   `json:"restore"`
	Why     string `json:"why,omitempty"` // why Create and Switch are off
}

// SwitchView is a switch in flight.
type SwitchView struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Creating bool   `json:"creating,omitempty"`
	Since    string `json:"since"`
}

// ListView is GET /api/characters (watch.characters/1).
type ListView struct {
	Schema     string       `json:"schema"`
	Active     string       `json:"active"`
	Caps       Caps         `json:"caps"`
	Switch     *SwitchView  `json:"switch,omitempty"`
	Characters []Entry      `json:"characters"`
	Deleted    []ArchiveRef `json:"deleted"` // archives of characters that are no longer in the folder (restorable)
}

// Result answers the four POST verbs.
type Result struct {
	Schema  string      `json:"schema"`
	Action  string      `json:"action"`
	Name    string      `json:"name"`
	Archive string      `json:"archive,omitempty"`
	Switch  *SwitchView `json:"switch,omitempty"`
}

// Manager implements the Characters desk. Its hooks are what a backend supplies; a nil Leave means the backend cannot
// switch (Caps says so).
type Manager struct {
	Store      *Store
	Inst       string                      // instance dir: the intent file and the parked state live here
	Active     func() string               // the character the bot plays now ("" before it knows)
	Measured   bool                        // D2BOT_RUN is set: no verb that changes anything
	Supervised bool                        // a supervisor restarts the bot with the next --char
	Leave      func(context.Context) error // save and leave the game; an error means nothing changed
	Begin      func()                      // optional: tell the runner a leave is under way (it must not rejoin)
	Exit       func(code int)              // ends the process after ExitDelay so the reply goes out first
	ExitCode   int
	ExitDelay  time.Duration
	Now        func() time.Time
	SaveDir    func() string           // resolves the save folder each call (a d2d finds it after its first save); nil keeps Store.SaveDir
	Verify     func(path string) error // the save check at switch time; nil = saves.VerifyFile (checksum, sections)

	mu sync.Mutex
}

func (m *Manager) now() time.Time {
	if m.Now != nil {
		return m.Now().UTC()
	}
	return time.Now().UTC()
}

func (m *Manager) caps() Caps {
	c := Caps{Delete: !m.Measured, Restore: !m.Measured}
	switch {
	case m.Measured:
		c.Why = "a measured run is in progress (D2BOT_RUN)"
	case m.Leave == nil:
		c.Why = "this backend cannot leave and rejoin a game"
	case !m.Supervised:
		c.Why = "switching needs the bot to run under its supervisor (D2BOT_SUPERVISE=1 or --supervise)"
	default:
		c.Create, c.Switch = true, true
	}
	return c
}

// sync takes the resolved save folder into the Store; callers hold mu.
func (m *Manager) sync() {
	if m.SaveDir != nil {
		m.Store.SaveDir = m.SaveDir()
	}
}

func (m *Manager) active() string {
	if m.Active == nil {
		return ""
	}
	return m.Active()
}

// List is GET /api/characters.
func (m *Manager) List() (any, *Refusal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sync()
	es, del, err := m.Store.List(m.active())
	if err != nil {
		if errors.Is(err, ErrNoSave) {
			return nil, no(http.StatusServiceUnavailable, "no_save_dir", "the save folder is not known yet (the first save creates it)")
		}
		return nil, no(http.StatusInternalServerError, "list_failed", err.Error())
	}
	if del == nil {
		del = []ArchiveRef{}
	}
	v := ListView{Schema: ListSchema, Active: m.active(), Caps: m.caps(), Characters: es, Deleted: del}
	if in, ok := m.pending(); ok {
		v.Switch = &SwitchView{From: in.From, To: in.To, Creating: in.Create, Since: in.At}
	}
	return v, nil
}

// gate is the refusals every changing verb shares.
func (m *Manager) gate() *Refusal {
	if m.Measured {
		return no(409, "measured_run", "a measured run never changes characters (D2BOT_RUN is set)")
	}
	if _, ok := m.pending(); ok {
		return no(409, "switch_in_progress", "a switch is already under way")
	}
	return nil
}

func (m *Manager) gateSwitch() *Refusal {
	if r := m.gate(); r != nil {
		return r
	}
	if m.Leave == nil {
		return no(501, "unsupported", m.caps().Why)
	}
	if !m.Supervised {
		return no(409, "no_supervisor", m.caps().Why)
	}
	return nil
}

// Create asks for a new level 1 character and switches to it. hardcore stays unbuilt (docs/decisions.md Current truth).
func (m *Manager) Create(name, class string, hardcore bool) (any, *Refusal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sync()
	if r := m.gateSwitch(); r != nil {
		return nil, r
	}
	if hardcore {
		return nil, no(409, "hardcore_unbuilt", "hardcore characters are not built yet")
	}
	if err := ValidName(name); err != nil {
		return nil, no(400, "bad_name", err.Error())
	}
	cl, err := ClassByID(class)
	if err != nil {
		return nil, no(400, "bad_class", err.Error())
	}
	if _, err := m.Store.Find(name); err == nil {
		return nil, no(409, "exists", "a character named "+name+" exists (names are not case sensitive)")
	} else if !errors.Is(err, ErrNotFound) && !errors.Is(err, ErrNoSave) {
		return nil, no(500, "list_failed", err.Error())
	}
	if sameName(name, m.active()) {
		return nil, no(409, "exists", name+" is the character playing now")
	}
	return m.leave(Intent{From: m.active(), To: name, Create: true, ClassIndex: cl.D2DIndex}, "create")
}

// Switch leaves the current game at once and starts name.
func (m *Manager) Switch(name string) (any, *Refusal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sync()
	if r := m.gateSwitch(); r != nil {
		return nil, r
	}
	e, err := m.Store.FindActive(name, m.active())
	if err != nil {
		return nil, no(404, "not_found", "no character named "+name)
	}
	if e.Active {
		return nil, no(409, "already_active", e.Name+" is already playing")
	}
	if e.Damaged != "" {
		return nil, no(409, "damaged", e.Name+" cannot be read: "+e.Damaged)
	}
	if e.Hardcore && e.Died {
		return nil, no(409, "dead_hardcore", e.Name+" died in hardcore")
	}
	verify := m.Verify
	if verify == nil {
		verify = func(p string) error { _, err := saves.VerifyFile(p); return err }
	}
	if err := verify(filepath.Join(m.Store.SaveDir, e.Name+".d2s")); err != nil {
		return nil, no(409, "corrupt", e.Name+" fails the save check: "+err.Error())
	}
	return m.leave(Intent{From: m.active(), To: e.Name, ClassIndex: -1}, "switch")
}

func (m *Manager) leave(in Intent, action string) (any, *Refusal) {
	in.Schema, in.At = IntentSchema, m.now().Format(time.RFC3339)
	if err := writeIntent(m.Inst, in); err != nil {
		return nil, no(500, "intent_failed", err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), switchTimeout)
	defer cancel()
	if m.Begin != nil {
		m.Begin()
	}
	if err := m.Leave(ctx); err != nil {
		ClearIntent(m.Inst)
		return nil, no(409, "save_failed", "the current hero could not be saved and left, nothing changed: "+err.Error())
	}
	if m.Exit != nil {
		time.AfterFunc(m.ExitDelay, func() { m.Exit(m.ExitCode) })
	}
	return Result{Schema: ResultSchema, Action: action, Name: in.To,
		Switch: &SwitchView{From: in.From, To: in.To, Creating: in.Create, Since: in.At}}, nil
}

// Delete archives name and then removes it. confirm must be the exact name.
func (m *Manager) Delete(name, confirm string) (any, *Refusal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sync()
	if r := m.gate(); r != nil {
		return nil, r
	}
	if confirm != name {
		return nil, no(400, "confirm_mismatch", "type the exact name of the character to delete it")
	}
	if err := ValidFileName(name); err != nil {
		return nil, no(400, "bad_name", err.Error())
	}
	e, err := m.Store.FindActive(name, m.active())
	if err != nil {
		return nil, no(404, "not_found", "no character named "+name)
	}
	if e.Active {
		return nil, no(409, "active", e.Name+" is playing: switch to another character first")
	}
	ref, err := m.Store.Delete(e.Name)
	if err != nil {
		return nil, no(500, "delete_failed", err.Error())
	}
	return Result{Schema: ResultSchema, Action: "delete", Name: e.Name, Archive: ref.ID}, nil
}

// Restore copies an archive back as a character.
func (m *Manager) Restore(archive string) (any, *Refusal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sync()
	if r := m.gate(); r != nil {
		return nil, r
	}
	if strings.ContainsAny(archive, `/\`) || strings.HasPrefix(archive, ".") {
		return nil, no(400, "bad_archive", "bad archive id")
	}
	ref, err := m.Store.Restore(archive)
	switch {
	case errors.Is(err, ErrNotFound):
		return nil, no(404, "not_found", "no such archive")
	case errors.Is(err, ErrExists):
		return nil, no(409, "exists", "a character named "+ref.Name+" exists; delete or rename it first")
	case err != nil:
		return nil, no(500, "restore_failed", err.Error())
	}
	return Result{Schema: ResultSchema, Action: "restore", Name: ref.Name, Archive: ref.ID}, nil
}
