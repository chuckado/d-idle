package questboard

import (
	"testing"
	"time"
)

// A refused start is remembered and the chunk's view carries the Why-not row (word, last refusal); a chunk that moves
// has none. Must fail while ChunkRow has no WhyNot. Two classes: the row is data, not per quest.
func TestDeskWhyNotRow(t *testing.T) {
	for _, class := range []string{"Sorceress", "Barbarian"} {
		r := newRig(t, class, 5)
		d := r.desk(nil)
		now := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
		d.cfg.Now = func() time.Time { return now }
		find := func(q, c string) *ChunkRow {
			d.invalidate()
			for _, row := range d.View().Quests {
				for i := range row.Chunks {
					if row.ID == q && (c == "" || row.Chunks[i].ID == c) {
						return &row.Chunks[i]
					}
				}
			}
			return nil
		}
		if a := d.Start(brq, ""); a.OK {
			t.Fatalf("%s: start was not refused: %+v", class, a)
		}
		cr := find(brq, "")
		if cr == nil || cr.WhyNot == nil {
			t.Fatalf("%s: no Why-not row: %+v", class, cr)
		}
		if cr.WhyNot.Word != cr.Readiness || cr.WhyNot.LastRefused == nil || !cr.WhyNot.LastRefused.Equal(now) {
			t.Fatalf("%s: row %+v readiness %q", class, cr.WhyNot, cr.Readiness)
		}
	}
}
