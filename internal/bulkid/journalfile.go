package bulkid

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// FileJournal loads the journal stored at path (empty when the file is absent) and saves every change atomically, so a
// restart in the middle of a Cain trip finds the ids that still have to go back to the stash. A damaged file is an
// error: guessing would lose track of items.
func FileJournal(path string) (*Journal, error) {
	j := &Journal{}
	b, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
	case err != nil:
		return nil, err
	default:
		if err := json.Unmarshal(b, &j.IDs); err != nil {
			return nil, errors.New("bulkid: journal " + path + " is damaged: " + err.Error())
		}
	}
	j.Save = func(ids []uint32) error {
		if ids == nil {
			ids = []uint32{}
		}
		b, err := json.Marshal(ids)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		tmp := path + ".tmp"
		if err := os.WriteFile(tmp, b, 0o644); err != nil {
			return err
		}
		return os.Rename(tmp, path)
	}
	return j, nil
}
