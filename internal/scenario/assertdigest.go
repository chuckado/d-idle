package scenario

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// discreteKinds are the assertion kinds whose observed value is a fact, not an amount: a quest word, an area, a hand
// skill, an object state. Two runs that read a different value there are not the same run even when both pass.
// Every other kind (gold, stash, life, area changes, item counts, log lines) is an amount that moves with shopping and
// timing, so only its pass or fail bucket enters the digest (design 2026-11-21 section 2.6).
var discreteKinds = map[string]bool{
	AQuestBits: true, AAreaIs: true, AAreaIn: true, AObjectState: true, APortalOpen: true,
	ARightSkillIs: true, ALeftSkillIs: true, ASkillPoints: true, ANoWrites: true,
}

// AssertDigest is the assertion digest D1: the run key and, for each evaluated assertion in order, its kind and bucket
// (ok, failed, unobserved) and, for a discrete kind, the value read. It is deliberately close to the verdict: it gates
// determinism together with it, while the whole-state digest (DigestState) is only a drift report.
func AssertDigest(r Result) string {
	var b strings.Builder
	b.WriteString(r.Key())
	for _, a := range r.Asserts {
		b.WriteString("\n" + a.Kind + "|")
		switch {
		case a.Unobserved:
			b.WriteString("unobserved")
		case a.OK:
			b.WriteString("ok")
		default:
			b.WriteString("fail")
		}
		if discreteKinds[a.Kind] && !a.Unobserved {
			b.WriteString("|" + a.Got)
		}
	}
	h := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(h[:8])
}
