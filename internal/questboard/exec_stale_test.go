package questboard

import "testing"

// A quest that reads done must never keep another quest from starting (user bug, deploy 31): the run of a chunk whose
// own check lags behind the quest word was left live for ever, and a done row has no Stop control.
func TestStaleAcceptDoesNotBlock(t *testing.T) {
	const ftower, countess = "forgotten_tower", "countess"
	cases := []struct {
		name      string
		doneQuest bool // the quest word says done
		doneChunk bool // the chunk's own check says done
		stop      bool // the player stopped the run
		blocks    bool
	}{
		{"active", false, false, false, true},
		{"quest done", true, false, false, false},
		{"chunk done", false, true, false, false},
		{"both done", true, true, false, false},
		{"stopped", false, false, true, false},
		{"stopped quest done", true, false, true, false},
	}
	for _, class := range smClasses {
		for _, c := range cases {
			t.Run(class+"/"+c.name, func(t *testing.T) {
				r := newRigSM(t, class, 30, true)
				if err := r.x.Accept(ftower, countess); err != nil {
					t.Fatal(err)
				}
				if c.stop {
					if err := r.x.Cancel(ftower, countess); err != nil {
						t.Fatal(err)
					}
				}
				if c.doneQuest {
					r.f.DoneQuest[ftower] = TierSave
				}
				if c.doneChunk {
					r.f.ChunkDone[ChunkKey(ftower, countess)] = TierSave
				}
				err := r.x.Accept(den, clr)
				if c.blocks != (err != nil) {
					t.Fatalf("blocks=%v, got err %v", c.blocks, err)
				}
				if !c.blocks {
					r.restart()
					if d := r.next(); d.Quest == ftower {
						t.Fatalf("stale run came back after a restart: %+v", d)
					}
				}
			})
		}
	}
}
