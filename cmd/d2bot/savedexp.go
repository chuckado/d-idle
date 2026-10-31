package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"d2lod-bot/internal/saves"
)

// savedExpReader returns bot.Deps.SavedExp: the exp stored in the character's .d2s, re-read only when the file's
// mtime or size changes (the game rewrites it on Save and Exit). ok is false while the name or file is unknown or
// unreadable; a half-written file keeps the last good value.
func savedExpReader(dir func() string, name func() string, read func(string) (int, uint32, error)) func() (uint32, bool) {
	var (
		mu    sync.Mutex
		mtime time.Time
		size  int64
		path  string
		exp   uint32
		have  bool
	)
	return func() (uint32, bool) {
		d, n := dir(), name()
		if d == "" || n == "" {
			return 0, false
		}
		p := filepath.Join(d, n+".d2s")
		st, err := os.Stat(p)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			return 0, false
		}
		if have && p == path && st.ModTime().Equal(mtime) && st.Size() == size {
			return exp, true
		}
		_, e, err := read(p)
		if err != nil {
			return exp, have && p == path
		}
		path, mtime, size, exp, have = p, st.ModTime(), st.Size(), e, true
		return exp, true
	}
}

func savedExpDep(dir string, name func() string) func() (uint32, bool) {
	return savedExpReader(func() string { return dir }, name, saves.ExpFile)
}

// d2dSaveDir returns a dir func for savedExpReader on d2d: the managed slot's d2d writes the character under its own
// home (XDG data dir; the exact subfolders are d2d's), so the folder holding <name>.d2s is searched once and kept; a
// miss is retried at most every retry (the character file appears at the first save). VERIFIED LIVE only through the
// drill harness, which finds the file the same way (docs/probes/drills D10).
func d2dSaveDir(home string, name func() string, retry time.Duration, now func() time.Time) func() string {
	var (
		mu   sync.Mutex
		dir  string
		last time.Time
	)
	return func() string {
		n := name()
		if home == "" || n == "" {
			return ""
		}
		mu.Lock()
		defer mu.Unlock()
		if dir != "" {
			if _, err := os.Stat(filepath.Join(dir, n+".d2s")); err == nil {
				return dir
			}
			dir = ""
		}
		if t := now(); !last.IsZero() && t.Sub(last) < retry {
			return ""
		} else {
			last = t
		}
		_ = filepath.WalkDir(home, func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && d.Name() == n+".d2s" {
				dir = filepath.Dir(p)
				return filepath.SkipAll
			}
			return nil
		})
		return dir
	}
}

// d2dSavedExp is bot.Deps.SavedExp on d2d: the exp in the character's save file under the slot's d2d home, re-read when
// the file changes (d2d rewrites it on Save and Exit, which every save point and game restart does). Between saves it
// is the exp as of the last save, so xpRisk is the exp gained since. Nil without a managed home (attached d2d: the
// observer falls back to the exp at the last entry or restart).
func d2dSavedExp(home string, name func() string) func() (uint32, bool) {
	if home == "" {
		return nil
	}
	return savedExpReader(d2dSaveDir(home, name, 10*time.Second, time.Now), name, saves.ExpFile)
}
