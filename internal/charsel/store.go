package charsel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Store is one save folder plus the archive next to it. Nothing here asks the game: callers decide whether a character
// may be touched (the Manager refuses the active one).
type Store struct {
	SaveDir    string // folder holding <Name>.d2s and its companions
	ArchiveDir string // <ArchiveDir>/<Name>/<stamp>/ (a delete never leaves the machine)
	StateRoot  string // parked per-character bot state: <StateRoot>/<Name>/ (state.go); copied into the archive
	Now        func() time.Time
}

func (s *Store) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

// Errors the Manager maps to refusals.
var (
	ErrNotFound = errors.New("charsel: no such character")
	ErrExists   = errors.New("charsel: a character with that name exists")
	ErrNoSave   = errors.New("charsel: the save folder is not known yet")
)

// Entry is one row of the list (watch.characters/1).
type Entry struct {
	Name       string       `json:"name"`
	Class      string       `json:"class"`
	ClassName  string       `json:"className"`
	Level      int          `json:"level"`
	Hardcore   bool         `json:"hardcore"`
	Died       bool         `json:"died,omitempty"`
	Expansion  bool         `json:"expansion"`
	LastPlayed string       `json:"lastPlayed,omitempty"`
	Active     bool         `json:"active"`
	Damaged    string       `json:"damaged,omitempty"` // why the header could not be read; the row can still be archived
	Archived   []ArchiveRef `json:"archived"`
}

// ArchiveRef names one archived copy: ID is "<Name>@<stamp>", the value POST restore takes.
type ArchiveRef struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Class      string `json:"class,omitempty"`
	Level      int    `json:"level,omitempty"`
	ArchivedAt string `json:"archivedAt"`
}

const stampFmt = "20060102T150405Z"

type archiveMeta struct {
	Schema     string `json:"schema"`
	Name       string `json:"name"`
	Class      string `json:"class,omitempty"`
	Level      int    `json:"level,omitempty"`
	ArchivedAt string `json:"archivedAt"`
}

const metaSchema = "watch.characters.archive/1"

var stampRe = regexp.MustCompile(`^\d{8}T\d{6}Z$`)

// isCharFile reports whether file (a base name) belongs to character name: .d2s .key .map .ma0 to .ma9 (V: docs/saves.md).
func isCharFile(file, name string) bool {
	if !strings.HasPrefix(file, name+".") {
		return false
	}
	ext := file[len(name)+1:]
	return ext == "d2s" || ext == "key" || ext == "map" || (len(ext) == 3 && ext[:2] == "ma" && ext[2] >= '0' && ext[2] <= '9')
}

// List reads the headers of every .d2s in the folder. A file whose header does not read is listed as damaged under its
// file name. active marks one row (case-insensitive).
func (s *Store) List(active string) ([]Entry, []ArchiveRef, error) {
	if s.SaveDir == "" {
		return nil, nil, ErrNoSave
	}
	arch, err := s.archives()
	if err != nil {
		return nil, nil, err
	}
	des, err := os.ReadDir(s.SaveDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, nil, err
	}
	out := []Entry{}
	live := map[string]bool{}
	for _, de := range des {
		n, ok := strings.CutSuffix(de.Name(), ".d2s")
		if !ok || de.IsDir() {
			continue
		}
		e := Entry{Name: n, Archived: []ArchiveRef{}}
		h, err := ReadHeader(filepath.Join(s.SaveDir, de.Name()))
		if err != nil {
			e.Damaged = err.Error()
		} else {
			e.Name, e.Class, e.ClassName, e.Level = h.Name, h.Class.ID, h.Class.Name, h.Level
			e.Hardcore, e.Died, e.Expansion = h.Hardcore, h.Died, h.Expansion
			if !h.LastPlayed.IsZero() {
				e.LastPlayed = h.LastPlayed.Format(time.RFC3339)
			}
			if h.Name != n {
				e.Name, e.Damaged = n, "the name in the file ("+h.Name+") differs from the file name"
			}
		}
		e.Active = active != "" && sameName(e.Name, active)
		live[strings.ToLower(e.Name)] = true
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	var deleted []ArchiveRef
	for _, a := range arch {
		if live[strings.ToLower(a.Name)] {
			for i := range out {
				if sameName(out[i].Name, a.Name) {
					out[i].Archived = append(out[i].Archived, a)
				}
			}
		} else {
			deleted = append(deleted, a)
		}
	}
	return out, deleted, nil
}

// Unsaved lists the heroes that have parked bot state under StateRoot but no .d2s in the save folder (a hero the game never saved,
// or whose save was removed by hand), apart from the one playing now: a new hero is listed only after its first save, so a folder
// like this is a hero that is waiting for it. It never reads game files.
func (s *Store) Unsaved(active string, saved []Entry) []string {
	out := []string{}
	if s.StateRoot == "" {
		return out
	}
	des, err := os.ReadDir(s.StateRoot)
	if err != nil {
		return out
	}
	have := map[string]bool{}
	for _, e := range saved {
		have[strings.ToLower(e.Name)] = true
	}
	for _, de := range des {
		n := de.Name()
		if !de.IsDir() || ValidFileName(n) != nil || have[strings.ToLower(n)] || (active != "" && sameName(n, active)) {
			continue
		}
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i]) < strings.ToLower(out[j]) })
	return out
}

// Find returns the entry named name (case-insensitive).
func (s *Store) Find(name string) (Entry, error) { return s.FindActive(name, "") }

// FindActive is Find with the active row marked.
func (s *Store) FindActive(name, active string) (Entry, error) {
	es, _, err := s.List(active)
	if err != nil {
		return Entry{}, err
	}
	for _, e := range es {
		if sameName(e.Name, name) {
			return e, nil
		}
	}
	return Entry{}, fmt.Errorf("%w: %q", ErrNotFound, name)
}

// archives lists every archived copy, newest first.
func (s *Store) archives() ([]ArchiveRef, error) {
	names, err := os.ReadDir(s.ArchiveDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []ArchiveRef
	for _, nd := range names {
		if !nd.IsDir() {
			continue
		}
		stamps, _ := os.ReadDir(filepath.Join(s.ArchiveDir, nd.Name()))
		for _, sd := range stamps {
			if !sd.IsDir() || !stampRe.MatchString(sd.Name()) {
				continue
			}
			r := ArchiveRef{ID: nd.Name() + "@" + sd.Name(), Name: nd.Name()}
			if t, err := time.Parse(stampFmt, sd.Name()); err == nil {
				r.ArchivedAt = t.Format(time.RFC3339)
			}
			if b, err := os.ReadFile(filepath.Join(s.ArchiveDir, nd.Name(), sd.Name(), "meta.json")); err == nil {
				var m archiveMeta
				if json.Unmarshal(b, &m) == nil {
					r.Class, r.Level = m.Class, m.Level
				}
			}
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ArchivedAt > out[j].ArchivedAt })
	return out, nil
}

// archivePath resolves an archive id to its folder, only through the listing (an id never becomes a path by itself).
func (s *Store) archivePath(id string) (ArchiveRef, string, error) {
	as, err := s.archives()
	if err != nil {
		return ArchiveRef{}, "", err
	}
	for _, a := range as {
		if a.ID == id {
			n, st, _ := strings.Cut(id, "@")
			return a, filepath.Join(s.ArchiveDir, n, st), nil
		}
	}
	return ArchiveRef{}, "", fmt.Errorf("%w: archive %q", ErrNotFound, id)
}

func (s *Store) stateDir(name string) string {
	if s.StateRoot == "" {
		return ""
	}
	return filepath.Join(s.StateRoot, name)
}

// Archive copies the character (save files and parked state) into a new archive folder and checks every copy byte for
// byte. It removes nothing.
func (s *Store) Archive(name string) (ArchiveRef, error) {
	if err := ValidFileName(name); err != nil {
		return ArchiveRef{}, err
	}
	e, err := s.Find(name)
	if err != nil {
		return ArchiveRef{}, err
	}
	name = e.Name
	stamp := s.now().Format(stampFmt)
	dst := filepath.Join(s.ArchiveDir, name, stamp)
	if _, err := os.Stat(dst); err == nil {
		return ArchiveRef{}, fmt.Errorf("charsel: archive %s@%s exists (two deletes in one second)", name, stamp)
	}
	tmp := dst + ".partial"
	_ = os.RemoveAll(tmp)
	if err := os.MkdirAll(tmp, 0o700); err != nil {
		return ArchiveRef{}, err
	}
	ok := false
	defer func() {
		if !ok {
			_ = os.RemoveAll(tmp)
		}
	}()
	files, err := s.charFiles(name)
	if err != nil {
		return ArchiveRef{}, err
	}
	for _, f := range files {
		if err := copyVerified(filepath.Join(s.SaveDir, f), filepath.Join(tmp, f)); err != nil {
			return ArchiveRef{}, err
		}
	}
	if sd := s.stateDir(name); sd != "" {
		if _, err := os.Stat(sd); err == nil {
			if err := copyTree(sd, filepath.Join(tmp, "state")); err != nil {
				return ArchiveRef{}, err
			}
		}
	}
	m := archiveMeta{Schema: metaSchema, Name: name, Class: e.Class, Level: e.Level, ArchivedAt: s.now().Format(time.RFC3339)}
	mb, _ := json.Marshal(m)
	if err := os.WriteFile(filepath.Join(tmp, "meta.json"), mb, 0o600); err != nil {
		return ArchiveRef{}, err
	}
	if err := os.Rename(tmp, dst); err != nil {
		return ArchiveRef{}, err
	}
	ok = true
	return ArchiveRef{ID: name + "@" + stamp, Name: name, Class: e.Class, Level: e.Level, ArchivedAt: m.ArchivedAt}, nil
}

// Delete archives first and removes the live files only when the archive is complete and verified. It never deletes
// the active character: that is the Manager's rule, kept there because only it knows who is active.
func (s *Store) Delete(name string) (ArchiveRef, error) {
	ref, err := s.Archive(name)
	if err != nil {
		return ArchiveRef{}, err
	}
	files, err := s.charFiles(ref.Name)
	if err != nil {
		return ref, err
	}
	for _, f := range files {
		if err := os.Remove(filepath.Join(s.SaveDir, f)); err != nil {
			return ref, fmt.Errorf("charsel: archived as %s but removing %s failed: %w", ref.ID, f, err)
		}
	}
	if sd := s.stateDir(ref.Name); sd != "" {
		_ = os.RemoveAll(sd)
	}
	return ref, nil
}

// Restore copies an archive back. It refuses when a character of that name (any case) is in the folder; the archive
// stays where it is.
func (s *Store) Restore(id string) (ArchiveRef, error) {
	ref, dir, err := s.archivePath(id)
	if err != nil {
		return ArchiveRef{}, err
	}
	if _, err := s.Find(ref.Name); err == nil {
		return ref, fmt.Errorf("%w: %q", ErrExists, ref.Name)
	}
	des, err := os.ReadDir(dir)
	if err != nil {
		return ref, err
	}
	var done []string
	undo := func() {
		for _, p := range done {
			_ = os.Remove(p)
		}
	}
	if err := os.MkdirAll(s.SaveDir, 0o700); err != nil {
		return ref, err
	}
	for _, de := range des {
		if de.IsDir() || !isCharFile(de.Name(), ref.Name) {
			continue
		}
		p := filepath.Join(s.SaveDir, de.Name())
		if err := copyVerified(filepath.Join(dir, de.Name()), p); err != nil {
			undo()
			return ref, err
		}
		done = append(done, p)
	}
	if len(done) == 0 {
		return ref, fmt.Errorf("charsel: archive %s holds no character files", id)
	}
	if sd := s.stateDir(ref.Name); sd != "" {
		if _, err := os.Stat(filepath.Join(dir, "state")); err == nil {
			_ = os.RemoveAll(sd)
			if err := copyTree(filepath.Join(dir, "state"), sd); err != nil {
				undo()
				return ref, err
			}
		}
	}
	return ref, nil
}

// charFiles lists the save folder files of a character, requiring its .d2s.
func (s *Store) charFiles(name string) ([]string, error) {
	des, err := os.ReadDir(s.SaveDir)
	if err != nil {
		return nil, err
	}
	var out []string
	hasD2S := false
	for _, de := range des {
		if !de.IsDir() && isCharFile(de.Name(), name) {
			out = append(out, de.Name())
			hasD2S = hasD2S || de.Name() == name+".d2s"
		}
	}
	if !hasD2S {
		return nil, fmt.Errorf("%w: %q", ErrNotFound, name)
	}
	return out, nil
}

// ValidFileName guards every name that becomes a path part.
func ValidFileName(n string) error {
	if n == "" || n != filepath.Base(n) || strings.ContainsAny(n, `/\`) || strings.HasPrefix(n, ".") || strings.Contains(n, "@") {
		return fmt.Errorf("charsel: bad character name %q", n)
	}
	return nil
}

func copyVerified(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	tmp := dst + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, bytes.NewReader(b)); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if back, err := os.ReadFile(tmp); err != nil || !bytes.Equal(back, b) {
		_ = os.Remove(tmp)
		return fmt.Errorf("charsel: copy of %s did not read back identical", filepath.Base(src))
	}
	return os.Rename(tmp, dst)
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o700)
		}
		if !d.Type().IsRegular() {
			return nil
		}
		return copyVerified(p, filepath.Join(dst, rel))
	})
}
