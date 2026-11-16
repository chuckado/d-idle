package lootrules

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

func mustRaw(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic("lootrules: " + err.Error())
	}
	return b
}

// PresetInfo is one row of the preset list.
type PresetInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PresetStarter is the id of the shipped default; "class:<Name>" ids are the Starter reduced to one hero class.
const PresetStarter = "starter"

// PresetCustom marks a hand-edited set (reordered, switched off, or sent whole).
const PresetCustom = "custom"

// ErrConflict means the stored set changed since the client read it (HTTP 409); Seq is the stored sequence number.
type ErrConflict struct{ Seq int }

func (e ErrConflict) Error() string {
	return fmt.Sprintf("the rule set changed since you opened it (now version %d); reload and apply your change again", e.Seq)
}

// Store keeps the player's rule set: one JSON file, atomic write, one previous copy (the .bak file) and a change counter.
// Nothing here talks to the game. A missing file is the Starter preset; a damaged file is moved aside (.bad) and the
// Starter is used, so a bad edit can never stop the bot (the load error is kept in LoadErr for the page).
type Store struct {
	mu       sync.Mutex
	path     string
	classes  []string
	cur      *RuleSet
	fresh    bool // nothing saved yet and nothing loaded: the default set is active
	onChange func(*RuleSet)
	// LoadErr is why the file on disk was not used at open, "" when it was fine or absent.
	LoadErr string
}

// OpenStore reads path (may not exist). classes are the hero class names the per-class presets are made for.
func OpenStore(path string, classes []string) *Store {
	s := &Store{path: path, classes: append([]string(nil), classes...)}
	s.cur = Starter()
	s.cur.Seq, s.cur.Preset = 1, PresetStarter
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		s.fresh = true
		return s
	}
	if err == nil {
		var rs *RuleSet
		if rs, err = Parse(b); err == nil {
			s.cur = rs
			if s.cur.Seq < 1 {
				s.cur.Seq = 1
			}
			return s
		}
	}
	s.LoadErr = err.Error()
	_ = os.Rename(path, path+".bad")
	return s
}

// OnChange registers the function called with the new active set after every successful save (the bot swaps its engine).
func (s *Store) OnChange(f func(*RuleSet)) { s.mu.Lock(); s.onChange = f; s.mu.Unlock() }

// Active returns the current set (treat it as read-only).
func (s *Store) Active() *RuleSet { s.mu.Lock(); defer s.mu.Unlock(); return s.cur }

// Presets lists the Starter and one preset per class.
func (s *Store) Presets() []PresetInfo {
	out := []PresetInfo{{PresetStarter, "Starter", "Today's behaviour: the pickit and the tidy table written as rules. Same for every class, with the throwing-weapon rule for the class that uses them."}}
	cl := append([]string(nil), s.classes...)
	sort.Strings(cl)
	for _, c := range cl {
		out = append(out, PresetInfo{"class:" + c, "Starter for " + c, "The Starter with only the rules that apply to a " + c + " (the class tags are removed so you can edit them freely)."})
	}
	return out
}

// Preset builds the preset id as a fresh rule set (no seq).
func (s *Store) Preset(id string) (*RuleSet, error) {
	base := Starter()
	switch {
	case id == PresetStarter:
	case strings.HasPrefix(id, "class:"):
		cls := strings.TrimPrefix(id, "class:")
		if !has(s.classes, cls) {
			return nil, Errors{{Path: "preset", Msg: fmt.Sprintf("unknown class %q", cls)}}
		}
		var rules []Rule
		for _, r := range base.Rules {
			if len(r.Classes) == 0 || has(r.Classes, cls) {
				r.Classes = nil
				rules = append(rules, r)
			}
		}
		base.Rules, base.Name = rules, "Starter for "+cls
	default:
		return nil, Errors{{Path: "preset", Msg: fmt.Sprintf("unknown preset %q", id)}}
	}
	base.Preset, base.Seq = id, 0
	return base, nil
}

// Put validates the JSON body (a RuleSet carrying the Seq it was based on) and saves it. With Preset set to a preset id,
// the rules are rebuilt from the preset and Options (the rules in the body are ignored); with Preset "custom" or empty
// the rules are taken as sent. Syntax errors carry line and column, field errors the path of the field.
func (s *Store) Put(body []byte) (*RuleSet, error) {
	in, err := parseLoose(body)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if in.Seq != s.cur.Seq {
		return nil, ErrConflict{Seq: s.cur.Seq}
	}
	next, err := s.build(in)
	if err != nil {
		return nil, err
	}
	next.Seq = s.cur.Seq + 1
	if err := s.write(next); err != nil {
		return nil, err
	}
	s.cur = next
	s.fresh = false
	if s.onChange != nil {
		s.onChange(next)
	}
	return next, nil
}

// build turns a body into the set it would save (preset and options applied, validated, compiled) without saving it.
func (s *Store) build(in *RuleSet) (*RuleSet, error) {
	next := in
	if in.Preset != "" && in.Preset != PresetCustom {
		base, err := s.preset(in.Preset)
		if err != nil {
			return nil, err
		}
		if in.Options != nil {
			if base, err = Customize(base, *in.Options); err != nil {
				return nil, err
			}
		}
		base.Preset = in.Preset
		next = base
	} else if in.Options != nil {
		return nil, Errors{{Path: "options", Msg: "options need a preset to apply to; send \"preset\" too"}}
	}
	next.Preset = orStr(in.Preset, PresetCustom)
	if es := next.Validate(); len(es) > 0 {
		return nil, es
	}
	if _, err := Compile(next, "", ""); err != nil {
		return nil, Errors{{Path: "$", Msg: err.Error()}}
	}
	return next, nil
}

// Build is build for a request body (the preview route): nothing is saved and the seq is not checked.
func (s *Store) Build(body []byte) (*RuleSet, error) {
	in, err := parseLoose(body)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.build(in)
}

func (s *Store) preset(id string) (*RuleSet, error) { return s.Preset(id) }

func orStr(a, b string) string {
	if a == "" {
		return b
	}
	return a
}

// parseLoose reads a body like Parse but does not validate the rules yet: a preset body may carry no rules at all.
func parseLoose(b []byte) (*RuleSet, error) {
	var rs RuleSet
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&rs); err != nil {
		var se *json.SyntaxError
		if errors.As(err, &se) {
			line, col := 1, 1
			for _, c := range b[:min(int(se.Offset), len(b))] {
				if c == '\n' {
					line, col = line+1, 1
				} else {
					col++
				}
			}
			return nil, Errors{{Msg: se.Error(), Line: line, Col: col}}
		}
		return nil, Errors{{Path: "$", Msg: err.Error()}}
	}
	if dec.More() {
		return nil, Errors{{Path: "$", Msg: "one JSON object only"}}
	}
	return &rs, nil
}

// write saves atomically: temp file in the same directory, fsync, copy of the old file to .bak, rename.
func (s *Store) write(rs *RuleSet) error {
	b, err := rs.Marshal()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	old, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) && s.cur != nil {
		// the first save of a hero: the set it replaces (the preset it started with) is the version Undo goes back to
		old, err = s.cur.Marshal()
	}
	if err == nil {
		if err := os.WriteFile(s.path+".bak", old, 0o600); err != nil {
			return err
		}
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".lootrules-*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	_, werr := f.Write(append(b, '\n'))
	if werr == nil {
		werr = f.Sync()
	}
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		_ = os.Remove(tmp)
		return werr
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// Limits of an imported set: a file from outside is data to check, not trusted.
const (
	maxImportRules   = 400
	maxImportString  = 600
	maxImportGroups  = 60
	maxImportListLen = 400
)

// ErrNoUndo means there is no previous version to go back to.
var ErrNoUndo = errors.New("there is no earlier version to go back to")

// Export is the canonical JSON of the active set (the same bytes the store writes; the set holds no secrets, no paths and
// no host names, only rules, groups and knobs).
func (s *Store) Export() ([]byte, error) { return s.Active().Marshal() }

// checkImport refuses a set that is valid JSON but unreasonably large or odd; Parse already validated the rules.
func checkImport(rs *RuleSet) Errors {
	var es Errors
	add := func(p, f string, a ...any) { es = append(es, Error{Path: p, Msg: fmt.Sprintf(f, a...)}) }
	long := func(p, v string) {
		if len(v) > maxImportString {
			add(p, "longer than %d characters", maxImportString)
		}
	}
	if len(rs.Rules) > maxImportRules {
		add("rules", "more than %d rules", maxImportRules)
	}
	if len(rs.Groups) > maxImportGroups {
		add("groups", "more than %d groups", maxImportGroups)
	}
	long("name", rs.Name)
	long("comment", rs.Comment)
	for i, r := range rs.Rules {
		p := fmt.Sprintf("rules[%d]", i)
		long(p+".id", r.ID)
		long(p+".name", r.Name)
		long(p+".reason", r.Reason)
		if r.When != nil {
			if b, err := json.Marshal(r.When); err == nil && len(b) > 8*maxImportString {
				add(p+".when", "condition too large")
			}
		}
	}
	for k, g := range rs.Groups {
		if len(g.Types)+len(g.TypePrefixes)+len(g.Codes) > maxImportListLen {
			add("groups."+k, "more than %d entries", maxImportListLen)
		}
	}
	return es
}

// Import replaces the active set with the file's: strict parse (unknown keys refused, every rule validated, a compile
// check), size checks, then saved like a custom edit (the previous version goes to .bak, so Undo works). seq is the
// version the client was looking at (409 when the stored set moved on). The file's own seq and preset are ignored.
func (s *Store) Import(body []byte, seq int) (*RuleSet, error) {
	rs, err := Parse(body)
	if err != nil {
		return nil, err
	}
	if es := checkImport(rs); len(es) > 0 {
		return nil, es
	}
	if _, err := Compile(rs, "", ""); err != nil {
		return nil, Errors{{Path: "$", Msg: err.Error()}}
	}
	rs.Preset, rs.Options = PresetCustom, nil
	return s.replace(rs, seq)
}

// Undo goes back to the previous saved version (the .bak file). The undone version becomes the new .bak, so a second
// Undo is a redo. seq is the version the client was looking at.
func (s *Store) Undo(seq int) (*RuleSet, error) {
	s.mu.Lock()
	path := s.path
	s.mu.Unlock()
	b, err := os.ReadFile(path + ".bak")
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNoUndo
	}
	if err != nil {
		return nil, err
	}
	rs, err := Parse(b)
	if err != nil {
		return nil, Errors{{Path: "$", Msg: "the earlier version is damaged: " + err.Error()}}
	}
	return s.replace(rs, seq)
}

// replace saves next as the new active set when the client's seq is current.
func (s *Store) replace(next *RuleSet, seq int) (*RuleSet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if seq != s.cur.Seq {
		return nil, ErrConflict{Seq: s.cur.Seq}
	}
	next.Seq = s.cur.Seq + 1
	if err := s.write(next); err != nil {
		return nil, err
	}
	s.cur = next
	s.fresh = false
	if s.onChange != nil {
		s.onChange(next)
	}
	return next, nil
}

// HasUndo says whether an earlier version exists.
func (s *Store) HasUndo() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := os.Stat(s.path + ".bak")
	return err == nil
}

// UseClassFallback is the per-hero default: a hero whose store file was never saved starts from the preset of its class
// (the Starter reduced to that class) instead of the plain Starter. A saved set, a damaged file or an unknown class are
// left alone. Returns whether the active set changed.
func (s *Store) UseClassFallback(class string) bool {
	s.mu.Lock()
	if !s.fresh || class == "" || !has(s.classes, class) || s.cur.Preset != PresetStarter {
		s.mu.Unlock()
		return false
	}
	s.mu.Unlock()
	rs, err := s.Preset("class:" + class)
	if err != nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.fresh {
		return false
	}
	rs.Seq = s.cur.Seq
	s.cur = rs
	if s.onChange != nil {
		s.onChange(rs)
	}
	return true
}
