package saves

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"d2lod-bot/internal/cheat/d2s"
)

// The snapshot library: a directory of named, immutable snapshots of a whole character
// (<lib>/<id>/manifest.json plus the file set), copied into other prefixes under a new name by Place.
// It is separate from the rotating per-instance backups (snapshot.go). Layout and rename rules: docs/snaplib.md.

const (
	ManifestName    = "manifest.json"
	ManifestVersion = 1
	statExperience  = 13 // stat ids of the character block (d2s: statBits)
	statGold        = 14
	statStashGold   = 15
)

var (
	ErrExists   = errors.New("saves: the target already has a character with that name")
	ErrSameName = errors.New("saves: the new name must differ from the snapshot's name")
	ErrBadID    = errors.New("saves: bad snapshot id")
	ErrManifest = errors.New("saves: snapshot does not match its manifest")
)

// Manifest describes one library snapshot. Every field is read from the files or given by the caller;
// nothing here is a claim about the game's behaviour.
type Manifest struct {
	Version      int               `json:"version"`
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Class        int               `json:"class"`
	ClassName    string            `json:"class_name"`
	Level        int               `json:"clvl"`
	XP           uint32            `json:"xp"`
	Gold         uint32            `json:"gold"`
	StashGold    uint32            `json:"stash_gold"`
	Difficulty   string            `json:"difficulty"` // the active difficulty (bit 7 of the header bytes)
	Act          int               `json:"act"`        // 1-based act of the active difficulty
	Hardcore     bool              `json:"hardcore"`
	Expansion    bool              `json:"expansion"`
	MapID        uint32            `json:"map_id"`
	KitHash      string            `json:"kit_hash"` // sha256 of the skill/item section of the .d2s
	Files        map[string]string `json:"files"`    // file name -> sha256
	Date         time.Time         `json:"date"`
	BotCommit    string            `json:"bot_commit,omitempty"`
	LedgerStatus string            `json:"ledger_status"` // "honest" or "cheated" at snapshot time (caller reads the ledger)
	SourceRun    string            `json:"source_run,omitempty"`
	SourceInst   string            `json:"source_instance,omitempty"`
}

// nameRE is the in-game rule for a character name (S: 2-15 letters, at most one - or _ not first or last);
// the tool is stricter than the file format on purpose, a name the game would not create is not placed.
var nameRE = regexp.MustCompile(`^[A-Za-z]+[-_]?[A-Za-z]+$`)

// CheckNewName validates a name for Place.
func CheckNewName(name string) error {
	if len(name) < 2 || len(name) > 15 || !nameRE.MatchString(name) {
		return fmt.Errorf("%w: %q (2-15 letters, optionally one - or _ inside)", ErrBadName, name)
	}
	return nil
}

func sum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// libFile reports whether file belongs to character name in the library: the character files plus a
// shared stash .d2x (U: offline 1.14d has none, V no .d2x in any prefix listing, docs/research/character-factory.md).
func libFile(file, name string) bool {
	return isCharFile(file, name) || file == name+".d2x"
}

func readSet(dir, name string) (map[string][]byte, error) {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := map[string][]byte{}
	for _, e := range ents {
		if e.Type().IsRegular() && libFile(e.Name(), name) {
			b, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				return nil, err
			}
			out[e.Name()] = b
		}
	}
	if _, ok := out[name+".d2s"]; !ok {
		return nil, fmt.Errorf("%w: %s.d2s in %s", ErrNoFiles, name, dir)
	}
	return out, nil
}

// SnapInfo is what the caller supplies to LibSnapshot besides the files.
type SnapInfo struct {
	ID           string // default <name>-<UTC stamp>
	BotCommit    string
	SourceRun    string
	SourceInst   string
	LedgerStatus string // required: "honest" or "cheated"
}

// LibSnapshot copies name's file set from saveDir into <lib>/<id>/ with a manifest. The source .d2s must
// verify; the files are re-stat'ed after reading (as Snapshot does) so a write racing the copy is caught.
// The caller asserts the game is not writing (o.gate). An existing id is never overwritten.
func LibSnapshot(saveDir, lib, name string, info SnapInfo, o Options, now time.Time) (Manifest, error) {
	if err := checkName(name); err != nil {
		return Manifest{}, err
	}
	if err := o.gate(); err != nil {
		return Manifest{}, err
	}
	if info.LedgerStatus != "honest" && info.LedgerStatus != "cheated" {
		return Manifest{}, fmt.Errorf("saves: ledger status %q (want honest or cheated)", info.LedgerStatus)
	}
	files, err := readCharFilesPlus(saveDir, name)
	if err != nil {
		return Manifest{}, err
	}
	f, err := Verify(files[name+".d2s"])
	if err != nil {
		return Manifest{}, fmt.Errorf("saves: source %s.d2s is corrupt, not snapshotting: %w", name, err)
	}
	m := manifestOf(f, files)
	m.Date = now.UTC()
	m.BotCommit, m.SourceRun, m.SourceInst, m.LedgerStatus = info.BotCommit, info.SourceRun, info.SourceInst, info.LedgerStatus
	m.ID = info.ID
	if m.ID == "" {
		m.ID = name + "-" + now.UTC().Format(stampLayout)
	}
	if err := checkName(m.ID); err != nil {
		return Manifest{}, fmt.Errorf("%w: %q", ErrBadID, m.ID)
	}
	if err := os.MkdirAll(lib, 0o755); err != nil {
		return Manifest{}, err
	}
	tmp, err := os.MkdirTemp(lib, ".tmp-")
	if err != nil {
		return Manifest{}, err
	}
	defer os.RemoveAll(tmp)
	for fn, b := range files {
		if err := os.WriteFile(filepath.Join(tmp, fn), b, 0o644); err != nil {
			return Manifest{}, err
		}
	}
	mb, _ := json.MarshalIndent(m, "", "  ")
	if err := os.WriteFile(filepath.Join(tmp, ManifestName), append(mb, '\n'), 0o644); err != nil {
		return Manifest{}, err
	}
	dir := filepath.Join(lib, m.ID)
	if exists(dir) {
		return Manifest{}, fmt.Errorf("saves: snapshot %s already exists", m.ID)
	}
	return m, os.Rename(tmp, dir)
}

// readCharFilesPlus is readCharFiles with the library's file set (adds .d2x).
func readCharFilesPlus(dir, name string) (map[string][]byte, error) {
	files, err := readCharFiles(dir, name)
	if err != nil {
		return nil, err
	}
	if b, err := os.ReadFile(filepath.Join(dir, name+".d2x")); err == nil {
		files[name+".d2x"] = b
	}
	return files, nil
}

func manifestOf(f *d2s.File, files map[string][]byte) Manifest {
	p := ProgressOf(f)
	m := Manifest{Version: ManifestVersion, Name: p.Name, Class: p.Class, Level: p.Level,
		Hardcore: p.Hardcore, Expansion: p.Expansion, MapID: p.MapID, Files: map[string]string{}}
	if p.Class >= 0 && p.Class < len(classNames) {
		m.ClassName = classNames[p.Class]
	}
	m.XP, _ = f.Stat(statExperience)
	m.Gold, _ = f.Stat(statGold)
	m.StashGold, _ = f.Stat(statStashGold)
	for d := 0; d < 3; d++ {
		if b := f.Difficulty(d); b&0x80 != 0 {
			m.Difficulty, m.Act = diffNames[d], int(b&0x7F)+1
		}
	}
	m.KitHash = sum(f.Tail())
	for fn, b := range files {
		m.Files[fn] = sum(b)
	}
	return m
}

// ReadManifest loads <lib>/<id>/manifest.json.
func ReadManifest(lib, id string) (Manifest, error) {
	var m Manifest
	if err := checkName(id); err != nil {
		return m, fmt.Errorf("%w: %q", ErrBadID, id)
	}
	b, err := os.ReadFile(filepath.Join(lib, id, ManifestName))
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return m, fmt.Errorf("saves: manifest %s: %w", id, err)
	}
	if m.Version != ManifestVersion || m.ID != id {
		return m, fmt.Errorf("%w: version %d id %q in directory %q", ErrManifest, m.Version, m.ID, id)
	}
	return m, nil
}

// LibList returns the manifests of the library, oldest first. Unreadable entries are an error, not skipped.
func LibList(lib string) ([]Manifest, error) {
	ents, err := os.ReadDir(lib)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Manifest
	for _, e := range ents {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		m, err := ReadManifest(lib, e.Name())
		if err != nil {
			return out, err
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date.Before(out[j].Date) })
	return out, nil
}

// LibVerify re-hashes every file against the manifest (no extra and no missing files), parses the .d2s
// (checksum, markers, stat block) and compares the header fields with the manifest. Returns the files read.
func LibVerify(lib, id string) (Manifest, map[string][]byte, error) {
	m, err := ReadManifest(lib, id)
	if err != nil {
		return m, nil, err
	}
	dir := filepath.Join(lib, id)
	ents, err := os.ReadDir(dir)
	if err != nil {
		return m, nil, err
	}
	files := map[string][]byte{}
	for _, e := range ents {
		if e.Name() == ManifestName {
			continue
		}
		want, ok := m.Files[e.Name()]
		if !ok {
			return m, nil, fmt.Errorf("%w: unexpected file %s", ErrManifest, e.Name())
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return m, nil, err
		}
		if got := sum(b); got != want {
			return m, nil, fmt.Errorf("%w: %s sha256 %s, manifest %s", ErrManifest, e.Name(), got[:12], want[:12])
		}
		files[e.Name()] = b
	}
	for fn := range m.Files {
		if _, ok := files[fn]; !ok {
			return m, nil, fmt.Errorf("%w: missing file %s", ErrManifest, fn)
		}
	}
	f, err := Verify(files[m.Name+".d2s"])
	if err != nil {
		return m, nil, fmt.Errorf("saves: snapshot %s: %w", id, err)
	}
	if got := manifestOf(f, files); got.Name != m.Name || got.Class != m.Class || got.Level != m.Level ||
		got.XP != m.XP || got.MapID != m.MapID || got.KitHash != m.KitHash {
		return m, nil, fmt.Errorf("%w: .d2s header differs from the manifest", ErrManifest)
	}
	return m, files, nil
}

// Rename returns the .d2s image b with the character name replaced and the checksum recomputed (the size
// field is unchanged: the name field is a fixed 16 bytes at 0x14, V: d2s.offName on real files; that no other
// copy of the name lives in the file is S, see docs/snaplib.md). The result must parse.
func Rename(b []byte, name string) ([]byte, error) {
	if err := CheckNewName(name); err != nil {
		return nil, err
	}
	if _, err := Verify(b); err != nil {
		return nil, err
	}
	out := append([]byte(nil), b...)
	const off, width, sumOff = 0x14, 16, 0x0C
	for i := 0; i < width; i++ {
		out[off+i] = 0
	}
	copy(out[off:], name)
	binary.LittleEndian.PutUint32(out[sumOff:], d2s.Checksum(out))
	if f, err := Verify(out); err != nil {
		return nil, err
	} else if f.Name() != name {
		return nil, fmt.Errorf("saves: renamed file reads %q", f.Name())
	}
	return out, nil
}

// Placed is the result of Place.
type Placed struct {
	Manifest   Manifest
	NewName    string
	Files      []string // file names written
	SourceD2S  string   // sha256 of the snapshot's .d2s
	PlacedD2S  string   // sha256 of the written .d2s (differs: name and checksum)
	Provenance string   // "copy-of-honest" or "copy-of-cheated"
}

// Place copies snapshot id of lib into destDir as character newName. It verifies the snapshot first, never
// touches an existing character (any file of that name, compared case-insensitively because the game runs
// under Wine on a case-insensitive file system, makes it refuse; there is no overwrite), renames the .d2s
// with a valid checksum, stages every file under a dot name, then publishes the .d2s last so a half-done
// placement never leaves a loadable character. A source marked cheated is refused unless allowCheated.
// The caller does the prefix preflight (instance lock, running game) before calling.
func Place(lib, id, destDir, newName string, allowCheated bool) (Placed, error) {
	var p Placed
	if err := CheckNewName(newName); err != nil {
		return p, err
	}
	m, files, err := LibVerify(lib, id)
	if err != nil {
		return p, err
	}
	p.Manifest, p.NewName, p.Provenance = m, newName, "copy-of-honest"
	if strings.EqualFold(newName, m.Name) {
		return p, ErrSameName
	}
	if m.LedgerStatus != "honest" {
		if !allowCheated {
			return p, fmt.Errorf("saves: snapshot %s has ledger status %q; refusing (allow-cheated to override)", id, m.LedgerStatus)
		}
		p.Provenance = "copy-of-cheated"
	}
	ents, err := os.ReadDir(destDir)
	if err != nil {
		return p, err
	}
	for _, e := range ents {
		if strings.HasPrefix(strings.ToLower(e.Name()), strings.ToLower(newName)+".") {
			return p, fmt.Errorf("%w: %s in %s", ErrExists, e.Name(), destDir)
		}
	}
	renamed, err := Rename(files[m.Name+".d2s"], newName)
	if err != nil {
		return p, err
	}
	p.SourceD2S, p.PlacedD2S = sum(files[m.Name+".d2s"]), sum(renamed)
	var names []string
	for fn := range files {
		if fn != m.Name+".d2s" {
			names = append(names, fn)
		}
	}
	sort.Strings(names)
	names = append(names, m.Name+".d2s") // published last
	var staged, done []string
	cleanup := func() {
		for _, s := range staged {
			os.Remove(s)
		}
		for _, d := range done {
			os.Remove(d)
		}
	}
	for _, fn := range names {
		out := newName + strings.TrimPrefix(fn, m.Name)
		b := files[fn]
		if fn == m.Name+".d2s" {
			b = renamed
		}
		tmp := filepath.Join(destDir, "."+out+".place")
		staged = append(staged, tmp)
		if err := os.WriteFile(tmp, b, 0o644); err != nil {
			cleanup()
			return p, err
		}
		p.Files = append(p.Files, out)
	}
	for i := range names {
		dst := filepath.Join(destDir, p.Files[i])
		if err := os.Rename(staged[i], dst); err != nil {
			cleanup()
			return p, err
		}
		done = append(done, dst)
	}
	// re-read what is on disk, not what we meant to write
	b, err := os.ReadFile(filepath.Join(destDir, newName+".d2s"))
	if err == nil && sum(b) != p.PlacedD2S {
		err = errors.New("saves: placed .d2s differs from what was written")
	}
	if err == nil {
		_, err = Verify(b)
	}
	if err != nil {
		cleanup()
		return p, err
	}
	return p, nil
}
