package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"d2lod-bot/internal/scenario"
)

// cutReason is the BLOCKED reason of a bot run that did not play its virtual budget to the end (killed by the wall
// limit or exited early without an error): its state is not what the scenario asked about, so it is not judged.
func cutReason(trace string) string {
	b, err := os.ReadFile(trace)
	if err != nil || scenario.RunLengthReached(b) {
		return ""
	}
	return "the bot run was cut before its virtual budget (no '" + scenario.RunLengthLine + "' in its log): not judged"
}

// fileSHA256 is the hex sha256 of a file, "" when unreadable.
func fileSHA256(p string) string {
	f, err := os.Open(p)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

// collectProv gathers the provenance of a run: the bot's VCS stamp (go version -m; D2LOD_BOT_COMMIT when the binary
// carries none), the binary hashes, the host ROLE (D2D_HOST_ROLE, never a machine name) and the load.
func collectProv(bot string) *scenario.Provenance {
	p := &scenario.Provenance{BotSHA256: fileSHA256(bot), D2DSHA256: fileSHA256(os.Getenv("D2D_BIN")), Host: os.Getenv("D2D_HOST_ROLE")}
	if p.Host == "" {
		p.Host = "unknown"
	}
	if out, err := exec.Command("go", "version", "-m", bot).Output(); err == nil {
		p.BotCommit, p.BotDirty = scenario.ParseBuildInfo(string(out))
	}
	if p.BotCommit == "" {
		p.BotCommit = os.Getenv("D2LOD_BOT_COMMIT")
	}
	if b, err := os.ReadFile("/proc/loadavg"); err == nil {
		if f := strings.Fields(string(b)); len(f) > 0 {
			p.Load1, _ = strconv.ParseFloat(f[0], 64)
		}
	}
	return p
}
