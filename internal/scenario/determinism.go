package scenario

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// DigestState hashes the observed state in a canonical form (sorted items and objects) so two runs of one seed can be
// compared byte for byte, not only by verdict. The source and write-count fields are part of it on purpose.
func DigestState(st State) string {
	c := st
	c.Items = append([]StateItem(nil), st.Items...)
	sort.Slice(c.Items, func(i, j int) bool {
		if c.Items[i].Code != c.Items[j].Code {
			return c.Items[i].Code < c.Items[j].Code
		}
		return c.Items[i].Where < c.Items[j].Where
	})
	c.Objects = append([]StateObject(nil), st.Objects...)
	sort.Slice(c.Objects, func(i, j int) bool {
		if c.Objects[i].Class != c.Objects[j].Class {
			return c.Objects[i].Class < c.Objects[j].Class
		}
		return c.Objects[i].State < c.Objects[j].State
	})
	c.LogLines = nil        // the bot log is not game state and its place in the run follows the wall clock
	b, _ := json.Marshal(c) // map keys of Quests are sorted by encoding/json
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:8])
}

// d1 is the assertion digest of a result: the recorded one, else computed from its asserts (so results written before
// the digest existed can be compared); "" when the result carries neither.
func d1(r Result) string {
	if r.AssertDigest != "" {
		return r.AssertDigest
	}
	if len(r.Asserts) > 0 {
		return AssertDigest(r)
	}
	return ""
}

// Divergence is one run key whose two executions disagree.
type Divergence struct {
	Key    string `json:"key"`
	A      string `json:"a"`
	B      string `json:"b"`
	Detail string `json:"detail"`
}

// CheckDeterminism pairs the results of two executions of the same plan by run key. The same seed must give the same
// verdict and the same assertion digest (both hard, returned in verdictDiffs); the whole-state digest is only reported
// as drift in digestDiffs (the bot loop is not yet fully deterministic, docs/d2d-determinism.md). An assertion digest
// is compared only when both results carry asserts or one (a result with neither is not). A flaky scenario is a
// failed scenario: any verdict or assertion-digest difference makes ok false.
func CheckDeterminism(a, b []Result) (ok bool, verdictDiffs, digestDiffs []Divergence, missing []string) {
	bm := map[string]Result{}
	for _, r := range b {
		bm[r.Key()] = r
	}
	ok = true
	for _, ra := range a {
		rb, found := bm[ra.Key()]
		if !found {
			missing = append(missing, ra.Key())
			ok = false
			continue
		}
		delete(bm, ra.Key())
		if ra.Status != rb.Status {
			verdictDiffs = append(verdictDiffs, Divergence{Key: ra.Key(), A: ra.Status, B: rb.Status, Detail: fmt.Sprintf("%s vs %s", ra.Reason, rb.Reason)})
			ok = false
		} else if da, db := d1(ra), d1(rb); da != "" && db != "" && da != db {
			verdictDiffs = append(verdictDiffs, Divergence{Key: ra.Key(), A: da, B: db, Detail: "same verdict, different assertion digest"})
			ok = false
		} else if ra.Digest != rb.Digest {
			digestDiffs = append(digestDiffs, Divergence{Key: ra.Key(), A: ra.Digest, B: rb.Digest, Detail: "same verdict, different state digest"})
		}
	}
	for k := range bm {
		missing = append(missing, k)
		ok = false
	}
	sort.Strings(missing)
	return
}
