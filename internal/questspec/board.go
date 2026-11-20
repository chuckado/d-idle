package questspec

import (
	"fmt"

	"d2lod-bot/internal/questboard"
)

// CheckBoard is rule 11: the chain describes exactly the chunks and steps of its data/questboard quest. The board stays
// the source the executor reads; a chain that drifts from it (a renamed step, a chunk without steps, a step order that
// disagrees with the chunk order) is a finding, so the two views cannot silently diverge.
func CheckBoard(c *Chain, b *questboard.Board) []Finding {
	var out []Finding
	add := func(step, f string, a ...any) {
		out = append(out, Finding{Rule: 11, Step: step, Msg: fmt.Sprintf(f, a...)})
	}
	if c.Quest == "" {
		return nil
	}
	q, err := b.Quest(c.Quest)
	if err != nil {
		return []Finding{{Rule: 11, Step: "*", Msg: err.Error()}}
	}
	chunkIdx := map[string]int{}
	for i, ch := range q.Chunks {
		chunkIdx[ch.ID] = i
	}
	covered := map[string]bool{}  // chunk/board step
	coveredN := map[string]bool{} // chunk/node
	chunkHas := map[string]bool{}
	last := -1
	for _, s := range c.Steps {
		i, ok := chunkIdx[s.Chunk]
		if !ok {
			add(s.ID, "chunk %q is not a chunk of quest %s", s.Chunk, q.ID)
			continue
		}
		chunkHas[s.Chunk] = true
		if i < last {
			add(s.ID, "chunk %q comes after a later chunk: the chain order disagrees with the board", s.Chunk)
		}
		last = i
		ch := q.Chunks[i]
		for _, bs := range s.Board {
			if !contains(ch.Steps, bs) {
				add(s.ID, "board step %q is not a step of chunk %s (%v)", bs, ch.ID, ch.Steps)
			}
			covered[ch.ID+"/"+bs] = true
		}
		for _, n := range s.Nodes {
			if !contains(ch.Nodes, n) {
				add(s.ID, "node %q is not a node of chunk %s (%v)", n, ch.ID, ch.Nodes)
			}
			coveredN[ch.ID+"/"+n] = true
		}
	}
	for _, ch := range q.Chunks {
		if !chunkHas[ch.ID] {
			add("*", "chunk %s of quest %s has no step in the chain", ch.ID, q.ID)
		}
		for _, st := range ch.Steps {
			if !covered[ch.ID+"/"+st] {
				add("*", "board step %s/%s is realised by no chain step", ch.ID, st)
			}
		}
		for _, n := range ch.Nodes {
			if !coveredN[ch.ID+"/"+n] {
				add("*", "node %s/%s is covered by no chain step", ch.ID, n)
			}
		}
	}
	return out
}
