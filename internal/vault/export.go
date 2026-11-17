package vault

import (
	"encoding/json"
	"fmt"
	"io"
)

// ExportVersion is the format of the export file (docs/vault.md).
const ExportVersion = 1

// Import limits: a hand-made file must not exhaust memory.
const (
	MaxImportBytes = 64 << 20
	MaxImportItems = 5000
)

type exportFile struct {
	Version int    `json:"version"`
	Hero    string `json:"hero"`
	Items   []Item `json:"items"`
}

// Export writes the live records (oldest first) as one JSON document.
func (s *Store) Export(w io.Writer) error {
	recs, _ := s.Query(Filter{Order: "old"})
	ef := exportFile{Version: ExportVersion, Hero: s.hero, Items: make([]Item, len(recs))}
	for i := range recs {
		ef.Items[i] = recs[i].Item
	}
	return json.NewEncoder(w).Encode(ef)
}

// ImportResult counts what happened to each item.
type ImportResult struct {
	Added, Duplicate, Rejected, Blocked int
}

// Import validates the whole file first (version, size, every item) and only then adds. Items go through the
// normal Add path: dedupe applies, capacity is respected, and import never trims (it stops at "full" and
// counts the rest as Blocked). The hero in the file is informational: items are stored under this vault's hero.
func (s *Store) Import(r io.Reader) (ImportResult, error) {
	b, err := io.ReadAll(io.LimitReader(r, MaxImportBytes+1))
	if err != nil {
		return ImportResult{}, err
	}
	if len(b) > MaxImportBytes {
		return ImportResult{}, ErrTooLarge
	}
	var ef exportFile
	if err := json.Unmarshal(b, &ef); err != nil {
		return ImportResult{}, fmt.Errorf("vault import: %w", err)
	}
	if ef.Version != ExportVersion {
		return ImportResult{}, fmt.Errorf("vault import: unsupported version %d", ef.Version)
	}
	if len(ef.Items) > MaxImportItems {
		return ImportResult{}, ErrTooLarge
	}
	for i := range ef.Items {
		if err := s.check(&ef.Items[i]); err != nil {
			return ImportResult{}, fmt.Errorf("vault import: item %d: %w", i, err)
		}
	}
	var res ImportResult
	for _, it := range ef.Items {
		s.mu.Lock()
		full := s.live >= s.capacity()
		s.mu.Unlock()
		if full {
			res.Blocked++
			continue
		}
		switch _, err := s.Add(it); err {
		case nil:
			res.Added++
		case ErrDuplicate:
			res.Duplicate++
		default:
			return res, err
		}
	}
	return res, nil
}
