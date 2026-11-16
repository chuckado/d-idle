package charsel

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Per-character bot state (docs/charselect.md section "State"): the bot keeps it per INSTANCE, not per character, so a
// switch parks the old character's files under <inst>/characters/<name>/ and brings the new character's back. The list
// of paths is data (perchar.json); a file that is not listed stays shared (profile level: onboarding, visit, currency,
// the ladder, layout and proposals keyed by profile).

//go:embed perchar.json
var perCharJSON []byte

type perChar struct {
	Paths []string `json:"paths"` // globs relative to the instance dir
}

func loadPerChar() []string {
	var p perChar
	if err := json.Unmarshal(perCharJSON, &p); err != nil || len(p.Paths) == 0 {
		panic("charsel: perchar.json")
	}
	return p.Paths
}

// UnnamedState is the folder name for state written before any character name was known.
const UnnamedState = "_unnamed"

func owned(inst string) ([]string, error) {
	var out []string
	for _, g := range loadPerChar() {
		m, err := filepath.Glob(filepath.Join(inst, g))
		if err != nil {
			return nil, err
		}
		out = append(out, m...)
	}
	return out, nil
}

// SwapState parks the per-character files of from and unparks those of to. A target without parked state starts with
// none (the bot creates its defaults). On a failure it puts everything back and returns the error; nothing is deleted.
func SwapState(inst, from, to string) error {
	if from == "" {
		from = UnnamedState
	}
	if err := ValidFileName(to); err != nil {
		return err
	}
	if err := ValidFileName(from); err != nil {
		return err
	}
	if from == to {
		return nil
	}
	root := filepath.Join(inst, "characters")
	if err := move(inst, filepath.Join(root, from)); err != nil {
		return fmt.Errorf("charsel: parking the state of %s: %w", from, err)
	}
	src := filepath.Join(root, to)
	if _, err := os.Stat(src); err != nil {
		return nil // no parked state: a fresh character
	}
	if _, err := unpark(inst, src); err != nil {
		_, _ = unpark(inst, filepath.Join(root, from))
		return fmt.Errorf("charsel: bringing back the state of %s: %w", to, err)
	}
	return nil
}

// move renames every owned path of inst into dst (keeping relative paths) ; on a failure it moves back what it moved.
func move(inst, dst string) (err error) {
	paths, err := owned(inst)
	if err != nil {
		return err
	}
	var moved []string
	defer func() {
		if err != nil {
			for _, r := range moved {
				_ = os.Rename(filepath.Join(dst, r), filepath.Join(inst, r))
			}
		}
	}()
	for _, p := range paths {
		rel, _ := filepath.Rel(inst, p)
		to := filepath.Join(dst, rel)
		if err = os.MkdirAll(filepath.Dir(to), 0o700); err != nil {
			return err
		}
		if _, e := os.Stat(to); e == nil {
			return fmt.Errorf("%s is already parked", rel)
		}
		if err = os.Rename(p, to); err != nil {
			return err
		}
		moved = append(moved, rel)
	}
	return nil
}

// unpark moves a parked tree back into inst, refusing to overwrite a live path.
func unpark(inst, src string) ([]string, error) {
	var moved []string
	err := filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		to := filepath.Join(inst, rel)
		if _, err := os.Stat(to); err == nil {
			return fmt.Errorf("%s exists", rel)
		}
		if err := os.MkdirAll(filepath.Dir(to), 0o700); err != nil {
			return err
		}
		if err := os.Rename(p, to); err != nil {
			return err
		}
		moved = append(moved, rel)
		return nil
	})
	if err != nil {
		for _, r := range moved {
			_ = os.Rename(filepath.Join(inst, r), filepath.Join(src, r))
		}
		return nil, err
	}
	return moved, nil
}

// ActiveFile names the file in the instance dir that holds the character the supervisor last switched to. A container
// restart (entrypoint.sh) reads it, so it does not fall back to the name in its environment and start the wrong hero.
const ActiveFile = "active-char"

// WriteActive records name as the active character, atomically (temp file, then rename).
func WriteActive(inst, name string) error {
	if err := ValidName(name); err != nil {
		return err
	}
	tmp := filepath.Join(inst, ActiveFile+".tmp")
	if err := os.WriteFile(tmp, []byte(name+"\n"), 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, filepath.Join(inst, ActiveFile)); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// ReadActive returns the recorded active character, or "" when there is none or it is not a valid name.
func ReadActive(inst string) string {
	b, err := os.ReadFile(filepath.Join(inst, ActiveFile))
	if err != nil {
		return ""
	}
	n := strings.TrimSpace(string(b))
	if ValidName(n) != nil {
		return ""
	}
	return n
}
