package scenario

import (
	"bytes"
	"strings"
)

// Provenance says which binaries and which machine state produced a result, so two results can be compared honestly
// (design 2026-11-21 section 2.2 H5/H6: one "pair" once ran two different bot builds). Host is a role name, never a
// machine name.
type Provenance struct {
	BotCommit string  `json:"bot_commit,omitempty"`
	BotDirty  bool    `json:"bot_dirty,omitempty"`
	BotSHA256 string  `json:"bot_sha256,omitempty"`
	D2DSHA256 string  `json:"d2d_sha256,omitempty"`
	Host      string  `json:"host,omitempty"`
	Load1     float64 `json:"load1,omitempty"`
	// Pace is virtual seconds per wall second over the whole run (0 when unknown).
	Pace float64 `json:"pace,omitempty"`
}

// RunLengthLine is what the bot logs when its virtual clock reached the budget (cmd/d2bot stopAtVirtual).
const RunLengthLine = "virtual run length reached"

// RunLengthReached reports whether the bot log shows the virtual budget was played to the end. A bot stopped by the wall
// limit (--run-for) or killed does not log it, and its state is not the state the scenario asked about: not judged.
func RunLengthReached(botLog []byte) bool { return bytes.Contains(botLog, []byte(RunLengthLine)) }

// ParseBuildInfo reads `go version -m <binary>` output for the VCS stamp: the revision and whether the tree was dirty.
func ParseBuildInfo(out string) (commit string, dirty bool) {
	for _, l := range strings.Split(out, "\n") {
		f := strings.Fields(l)
		if len(f) < 2 || f[0] != "build" {
			continue
		}
		k, v, ok := strings.Cut(f[1], "=")
		if !ok {
			continue
		}
		switch k {
		case "vcs.revision":
			commit = v
		case "vcs.modified":
			dirty = v == "true"
		}
	}
	return
}

// PaceOf is virtual seconds per wall second; 0 when the wall time is unknown.
func PaceOf(virtualMin float64, wallSec int) float64 {
	if wallSec <= 0 {
		return 0
	}
	return virtualMin * 60 / float64(wallSec)
}
