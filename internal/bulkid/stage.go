package bulkid

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

// ErrNoRoom is what a Mover returns when the target grid has no spot for the item.
var ErrNoRoom = errors.New("no room")

// Mover is the game side of the stash<->pack moves (cmd wires the item verb or the town's click path; tests use a fake).
// Pack and Stash re-read the hero each call: a move is only believed when the read-back shows it.
type Mover interface {
	Pack(ctx context.Context) ([]Item, error)
	Stash(ctx context.Context) ([]Item, error)
	// Move puts the item into the first free spot of the place (PlacePack or PlaceStash); ErrNoRoom when none fits.
	Move(ctx context.Context, id uint32, to string) error
}

// Journal remembers which items were taken out of the stash, BEFORE each move, so a restart at any point can put them
// back. Save is called after every change; a nil Save keeps it in memory only.
type Journal struct {
	IDs  []uint32             `json:"ids"`
	Save func([]uint32) error `json:"-"`
}

func (j *Journal) add(id uint32) error {
	if !slices.Contains(j.IDs, id) {
		j.IDs = append(j.IDs, id)
	}
	return j.save()
}

func (j *Journal) drop(id uint32) error {
	j.IDs = slices.DeleteFunc(j.IDs, func(x uint32) bool { return x == id })
	return j.save()
}

func (j *Journal) save() error {
	if j.Save == nil {
		return nil
	}
	return j.Save(append([]uint32(nil), j.IDs...))
}

// Stager moves stash picks into the pack in batches and back.
type Stager struct {
	M     Mover
	J     *Journal
	Chunk int                              // items per batch (identify.chunk)
	Logf  func(format string, args ...any) // one line per batch; nil: silent
}

func (s *Stager) logf(f string, a ...any) {
	if s.Logf != nil {
		s.Logf(f, a...)
	}
}

// StageResult says what arrived.
type StageResult struct {
	Moved  []uint32 // verified in the pack
	Left   []uint32 // still in the stash (no room or a failed move)
	Reason string   // why Left is not empty
}

func has(items []Item, id uint32) bool {
	return slices.ContainsFunc(items, func(it Item) bool { return it.ID == id })
}

// StageIn moves the ids from the stash to the pack, Chunk at a time. Each id is journaled before its move and each
// batch is read back before the next starts; an id that does not show in the pack ends the staging (Left keeps it, the
// journal keeps it for Return). A full pack ends it quietly: nothing is forced and nothing is dropped.
func (s *Stager) StageIn(ctx context.Context, ids []uint32) (StageResult, error) {
	var res StageResult
	chunk := max(s.Chunk, 1)
	for start := 0; start < len(ids); start += chunk {
		batch := ids[start:min(start+chunk, len(ids))]
		var tried []uint32
		stop := ""
		for _, id := range batch {
			if err := s.J.add(id); err != nil {
				return res, fmt.Errorf("bulkid: journal: %w", err)
			}
			err := s.M.Move(ctx, id, PlacePack)
			if errors.Is(err, ErrNoRoom) {
				_ = s.J.drop(id) // it never left the stash
				stop = "the pack has no room"
				break
			}
			if err != nil {
				stop = "move failed: " + err.Error()
				tried = append(tried, id)
				break
			}
			tried = append(tried, id)
		}
		pack, err := s.M.Pack(ctx)
		if err != nil {
			return res, fmt.Errorf("bulkid: read pack: %w", err)
		}
		ok := 0
		for _, id := range tried {
			if has(pack, id) {
				res.Moved = append(res.Moved, id)
				ok++
			} else {
				// reported moved or not, it is not in the pack: the read decides; Return finds it in the stash
				res.Left = append(res.Left, id)
				if stop == "" {
					stop = "an item did not arrive in the pack"
				}
			}
		}
		s.logf("bulkid: batch %d moved %d verified %d", start/chunk+1, len(tried), ok)
		if stop != "" {
			res.Reason = stop
			for _, id := range ids {
				if !slices.Contains(res.Moved, id) && !slices.Contains(res.Left, id) {
					res.Left = append(res.Left, id)
				}
			}
			return res, nil
		}
	}
	return res, nil
}

// ReturnResult says where the journaled items ended.
type ReturnResult struct {
	Back   []uint32 // moved pack -> stash and verified
	Kept   []uint32 // still in the pack (the stash is full)
	Gone   []uint32 // in neither place: an unresolved item stays in the journal
	Reason string
}

// Return puts every journaled item that is in the pack back into the stash, verifying each, and clears the journal for
// the ones that are in the stash. An item that cannot go back stays in the pack (never dropped or sold) and in the
// journal.
func (s *Stager) Return(ctx context.Context) (ReturnResult, error) {
	var res ReturnResult
	for _, id := range slices.Clone(s.J.IDs) {
		pack, err := s.M.Pack(ctx)
		if err != nil {
			return res, fmt.Errorf("bulkid: read pack: %w", err)
		}
		stash, err := s.M.Stash(ctx)
		if err != nil {
			return res, fmt.Errorf("bulkid: read stash: %w", err)
		}
		switch {
		case has(stash, id):
			if err := s.J.drop(id); err != nil {
				return res, err
			}
		case !has(pack, id):
			res.Gone = append(res.Gone, id)
		default:
			err := s.M.Move(ctx, id, PlaceStash)
			if err == nil {
				if st, rerr := s.M.Stash(ctx); rerr == nil && has(st, id) {
					res.Back = append(res.Back, id)
					if err := s.J.drop(id); err != nil {
						return res, err
					}
					continue
				}
			}
			res.Kept = append(res.Kept, id)
			if errors.Is(err, ErrNoRoom) {
				res.Reason = "the stash is full: items stay in the pack"
			} else if res.Reason == "" {
				res.Reason = "an item did not go back to the stash"
			}
		}
	}
	return res, nil
}
