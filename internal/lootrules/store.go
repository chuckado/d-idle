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
	next.Seq = s.cur.Seq + 1
	if err := s.write(next); err != nil {
		return nil, err
	}
	s.cur = next
	if s.onChange != nil {
		s.onChange(next)
	}
	return next, nil
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
	if old, err := os.ReadFile(s.path); err == nil {
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
