package main

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"d2lod-bot/internal/buildinfo"
	"d2lod-bot/internal/support"
	"d2lod-bot/internal/watch"
)

// The support bundle desk (docs/support-bundle.md). It only gathers: internal/support redacts, scans and zips, and
// internal/watch serves the routes. Every name offered here must be in support.Allowed or be the two optional files.

const (
	supportLogHours = 6
	supportLogMax   = 4 << 20 // read at most the end of the log
	supportTailJSON = 200
)

type supportDesk struct {
	l        ladderOptions
	stateDir string
	client   *http.Client
	secrets  []string // tokens in use: scrubbed as literals whatever their shape
}

func newSupportAPI(l ladderOptions, stateDir string, tokens ...string) *watch.SupportAPI {
	a := watch.NewSupportAPI()
	a.Bind(&supportDesk{l: l, stateDir: stateDir, client: &http.Client{Timeout: 5 * time.Second}, secrets: tokens})
	return a
}

// knownLiterals are the strings that must never reach the bundle whatever shape they have: the tokens in use, the home
// directory and user names, this machine's name and the configured host names and URLs (the update feed URL carries
// the pack path).
func (d *supportDesk) knownLiterals() []string {
	out := append([]string(nil), d.secrets...)
	for _, k := range []string{"D2DVIEW_TOKEN", support.EnvToken, "HOME", "USER", "LOGNAME", "USERNAME"} {
		out = append(out, os.Getenv(k))
	}
	if h, err := os.UserHomeDir(); err == nil {
		out = append(out, h)
	}
	if u, err := user.Current(); err == nil {
		out = append(out, u.Username)
	}
	if h, err := os.Hostname(); err == nil {
		out = append(out, h)
	}
	for _, k := range []string{support.EnvURL, "D2BOT_UPDATE_FEED", "D2BOT_WATCH_ORIGIN", "D2BOT_WATCH_TAP_URL"} {
		for _, v := range strings.Split(os.Getenv(k), ",") {
			v = strings.TrimSpace(v)
			out = append(out, v)
			if u, err := url.Parse(v); err == nil && u.Host != "" {
				out = append(out, u.Hostname(), strings.Trim(u.Path, "/"))
			}
		}
	}
	out = append(out, d.l.supportLayout.Root, d.stateDir, d.l.currencyDir)
	var keep []string
	for _, s := range out {
		if s = strings.TrimSpace(s); len(s) >= 3 && s != "/" {
			keep = append(keep, s)
		}
	}
	return keep
}

func (d *supportDesk) Sender() support.Sender {
	return support.Sender{URL: os.Getenv(support.EnvURL), Token: os.Getenv(support.EnvToken)}
}

func (d *supportDesk) Collect(o support.Options, local http.Handler) (support.Inputs, error) {
	now := time.Now()
	in := support.Inputs{Now: now, LogHours: supportLogHours}
	add := func(name string, b []byte) {
		if len(b) > 0 {
			in.Files = append(in.Files, support.Input{Name: name, Data: b})
		}
	}
	in.Extra = d.knownLiterals()

	ver, _ := json.Marshal(map[string]string{"version": buildinfo.Current(), "image": os.Getenv("D2BOT_IMAGE_TAG"), "instance": d.l.instance})
	add("version.json", ver)
	osName := ""
	if b, err := os.ReadFile("/etc/os-release"); err == nil {
		for _, ln := range strings.Split(string(b), "\n") {
			if v, ok := strings.CutPrefix(ln, "PRETTY_NAME="); ok {
				osName = strings.Trim(v, `"`)
			}
		}
	}
	_, inDocker := os.Stat("/.dockerenv")
	sys, _ := json.Marshal(map[string]any{"os": runtime.GOOS, "arch": runtime.GOARCH, "cpus": runtime.NumCPU(), "go": runtime.Version(), "distro": osName, "container": inDocker == nil})
	add("system.json", sys)
	if d.l.supportKnobs != nil {
		if b, err := json.Marshal(d.l.supportKnobs()); err == nil {
			add("knobs.json", b)
		}
	}
	for name, path := range map[string]string{"glance.json": "/api/glance", "status.json": watch.HeroRoute, "readiness.json": "/api/readiness", "danger.json": "/api/danger"} {
		add(name, watch.FetchLocal(local, path))
	}
	log := d.logTail()
	add("log.txt", log)
	var errs []string
	for _, ln := range strings.Split(string(log), "\n") {
		if u := strings.ToUpper(ln); strings.Contains(u, "ERROR") || strings.Contains(u, "WARN") || strings.Contains(u, "PANIC") {
			errs = append(errs, ln)
		}
	}
	add("errors.txt", []byte(strings.Join(errs, "\n")))
	add("decisions.json", jsonlTail(filepath.Join(d.stateDir, "decisions.jsonl"), supportTailJSON))
	add("ledger.json", jsonlTail(filepath.Join(d.l.currencyDir, "ledger.jsonl"), supportTailJSON))
	add("deaths.json", jsonlTail(filepath.Join(d.stateDir, "deaths.jsonl"), supportTailJSON))

	if o.IncludeScreenshot {
		add(support.ShotName, d.screenshot())
	}
	if o.IncludeSave {
		// U: Instance.SaveDir is a placeholder for fresh setups (internal/paths); a real install keeps saves in the Wine prefix.
		if inst, err := d.l.supportLayout.Instance(d.l.instance); err == nil {
			saves, _ := filepath.Glob(filepath.Join(inst.SaveDir(), "*.d2s"))
			for _, p := range saves {
				if b, err := os.ReadFile(p); err == nil {
					add(support.SavePfx+filepath.Base(p), b)
				}
			}
		}
	}
	return in, nil
}

func (d *supportDesk) logTail() []byte {
	var logs []string
	if d.l.instance != "" {
		if inst, err := d.l.supportLayout.Instance(d.l.instance); err == nil {
			logs = []string{filepath.Join(d.l.supportLayout.InstancesDir(), inst.ID, "logs", "d2bot.log")}
		}
	}
	if len(logs) == 0 {
		logs, _ = filepath.Glob(filepath.Join(d.l.supportLayout.InstancesDir(), "*", "logs", "d2bot.log"))
	}
	var out []byte
	for _, p := range logs {
		out = append(out, readEnd(p, supportLogMax)...)
	}
	return out
}

// readEnd returns the last max bytes of a file starting at a line start.
func readEnd(path string, max int64) []byte {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil
	}
	off := int64(0)
	if st.Size() > max {
		off = st.Size() - max
	}
	b := make([]byte, st.Size()-off)
	if _, err := f.ReadAt(b, off); err != nil && err != io.EOF {
		return nil
	}
	if off > 0 {
		if i := strings.IndexByte(string(b), '\n'); i >= 0 {
			b = b[i+1:]
		}
	}
	return b
}

// jsonlTail wraps the last n valid JSON lines of a .jsonl file as {"lines": [...]}; nil when the file is absent.
func jsonlTail(path string, n int) []byte {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var ring []json.RawMessage
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		if line := sc.Bytes(); json.Valid(line) {
			ring = append(ring, append(json.RawMessage(nil), line...))
			if len(ring) > n {
				ring = ring[1:]
			}
		}
	}
	b, _ := json.Marshal(map[string]any{"lines": ring})
	return b
}

// screenshot reads one frame of the d2d view tap (the same /frame the video poller uses); nil without a tap.
func (d *supportDesk) screenshot() []byte {
	if d.l.tapURL == "" || d.l.videoSource != videoTap {
		return nil
	}
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(d.l.tapURL, "/")+"/frame", nil)
	if err != nil {
		return nil
	}
	if t := os.Getenv("D2DVIEW_TOKEN"); t != "" {
		req.Header.Set("Authorization", "Bearer "+t)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil
	}
	b, _ := io.ReadAll(io.LimitReader(resp.Body, support.MaxShot+1))
	return b
}

var _ watch.SupportDesk = (*supportDesk)(nil)
