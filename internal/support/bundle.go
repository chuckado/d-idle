package support

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Limits. MaxZip is what the receiver accepts (docs/support-bundle.md); the client stays under it.
const (
	MaxZip       = 20 << 20
	MaxTextFile  = 1 << 20   // after redaction; a longer file keeps its tail
	MaxShot      = 4 << 20   // optional screenshot
	MaxSave      = 256 << 10 // optional character save (a .d2s is about 8 KiB)
	PreviewLines = 40
)

// Allowed lists the only file names a bundle can contain (allowlist, not blocklist). A collector that offers any
// other name is ignored and the name is reported as dropped.
var Allowed = map[string]bool{
	"version.json": true, "system.json": true, "knobs.json": true, "status.json": true, "glance.json": true,
	"summary-minute.json": true, "summary-area.json": true, "summary-ledger.json": true, "d2d-trace.json": true,
	"streamstats.json": true, "log.txt": true, "errors.txt": true,
}

// Optional files, only with their option on.
const (
	ShotName = "screenshot.png"
	SavePfx  = "save/"
)

// Options are the two checkboxes of the sheet; both default to off.
type Options struct {
	IncludeScreenshot bool `json:"includeScreenshot"`
	IncludeSave       bool `json:"includeSave"`
}

// Input is one file a collector offers.
type Input struct {
	Name string
	Data []byte
}

// Inputs is what the bot side collected. Extra are literal strings of this machine to scrub (host, user, data dir).
type Inputs struct {
	Files []Input
	Extra []string
	Now   time.Time
	// LogHours keeps only log lines of the last hours that carry a leading RFC3339 timestamp; 0: keep all (the byte cap still applies).
	LogHours int
}

// File is one line of the preview.
type File struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

// Bundle is a built, scanned bundle. Zip is only sent when Findings is empty.
type Bundle struct {
	ID         string    `json:"id"`
	Created    time.Time `json:"created"`
	Files      []File    `json:"files"`
	LogPreview string    `json:"logPreview"`
	Findings   []Finding `json:"findings"`
	Dropped    []string  `json:"dropped,omitempty"`
	Options    Options   `json:"options"`
	ZipSize    int       `json:"zipSize"`
	Zip        []byte    `json:"-"`
}

// Sendable is the fail-closed test: no finding, a zip, within the cap.
func (b *Bundle) Sendable() bool {
	return b != nil && len(b.Findings) == 0 && len(b.Zip) > 0 && len(b.Zip) <= MaxZip
}

// Build redacts and scans the inputs and zips them. It never returns a zip it could not scan clean; on a finding the
// bundle is returned with Findings set (and no Zip) so the sheet can show what blocked it.
func Build(in Inputs, o Options) (*Bundle, error) {
	if in.Now.IsZero() {
		in.Now = time.Now()
	}
	b := &Bundle{Created: in.Now.UTC(), Options: o}
	var idb [8]byte
	if _, err := rand.Read(idb[:]); err != nil {
		return nil, err
	}
	b.ID = hex.EncodeToString(idb[:])
	type entry struct {
		name string
		data []byte
	}
	var out []entry
	saves := 0
	for _, f := range in.Files {
		switch {
		case Allowed[f.Name]:
			txt := prepare(f.Name, f.Data, in)
			if len(txt) > MaxTextFile {
				txt = txt[len(txt)-MaxTextFile:]
			}
			out = append(out, entry{f.Name, txt})
			b.Findings = append(b.Findings, Scan(f.Name, string(txt), in.Extra...)...)
		case f.Name == ShotName && o.IncludeScreenshot:
			if !isImage(f.Data) || len(f.Data) > MaxShot {
				b.Dropped = append(b.Dropped, f.Name+" (not an image or too large)")
				continue
			}
			out = append(out, entry{f.Name, f.Data})
		case strings.HasPrefix(f.Name, SavePfx) && o.IncludeSave:
			if !hasSaveMagic(f.Data) || len(f.Data) > MaxSave {
				b.Dropped = append(b.Dropped, "a save (not a character file or too large)")
				continue
			}
			saves++ // the character name is never in the path
			out = append(out, entry{fmt.Sprintf("%scharacter%d.d2s", SavePfx, saves), f.Data})
		case f.Name == ShotName || strings.HasPrefix(f.Name, SavePfx):
			// offered but not ticked: left out without a word, it was never part of the bundle
		default:
			b.Dropped = append(b.Dropped, f.Name)
		}
	}
	// the second line: a character save anywhere but save/ (a collector slip) blocks the bundle
	for _, e := range out {
		if hasSaveMagic(e.data) && !strings.HasPrefix(e.name, SavePfx) {
			b.Findings = append(b.Findings, Finding{File: e.name, Line: 1, Class: "save", Context: "looks like a character save outside save/"})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	manifest := manifestText(o, in.Now)
	out = append([]entry{{"MANIFEST.txt", []byte(manifest)}}, out...)
	for _, e := range out {
		b.Files = append(b.Files, File{e.name, len(e.data)})
		if e.name == "log.txt" {
			b.LogPreview = tail(string(e.data), PreviewLines)
		}
	}
	if len(b.Findings) > 0 {
		return b, nil
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range out {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: e.name, Method: zip.Deflate, Modified: in.Now.UTC()})
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(e.data); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	if buf.Len() > MaxZip {
		return nil, errors.New("the bundle is larger than the upload limit")
	}
	b.Zip, b.ZipSize = buf.Bytes(), buf.Len()
	return b, nil
}

// prepare redacts one allowlisted file: JSON is parsed and stripped of secret-named keys first (a value-only
// export), then everything goes through Redact; text is time-filtered (logs) and redacted.
func prepare(name string, data []byte, in Inputs) []byte {
	if strings.HasSuffix(name, ".json") {
		var v any
		if json.Unmarshal(data, &v) != nil {
			return []byte("{\"error\": \"not valid JSON, omitted\"}\n")
		}
		data, _ = json.MarshalIndent(dropSecretKeys(v), "", "  ")
		return []byte(Redact(string(data), in.Extra...) + "\n")
	}
	s := string(data)
	if name == "log.txt" || name == "errors.txt" {
		s = recent(s, in.Now, in.LogHours)
	}
	return []byte(Redact(s, in.Extra...))
}

func dropSecretKeys(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, e := range x {
			lk := strings.ToLower(k)
			if strings.Contains(lk, "token") || strings.Contains(lk, "secret") || strings.Contains(lk, "password") || strings.Contains(lk, "passphrase") || strings.Contains(lk, "cookie") || strings.Contains(lk, "key") {
				delete(x, k)
				continue
			}
			x[k] = dropSecretKeys(e)
		}
	case []any:
		for i, e := range x {
			x[i] = dropSecretKeys(e)
		}
	}
	return v
}

// recent drops the lines older than hours that start with an RFC3339 time; lines without one follow the last decision.
func recent(s string, now time.Time, hours int) string {
	if hours <= 0 {
		return s
	}
	cut := now.Add(-time.Duration(hours) * time.Hour)
	var kept []string
	keep := true
	for _, l := range strings.Split(s, "\n") {
		if i := strings.IndexAny(l, " \t"); i > 0 {
			if t, err := time.Parse(time.RFC3339, strings.Trim(l[:i], "[]")); err == nil {
				keep = !t.Before(cut)
			}
		}
		if keep {
			kept = append(kept, l)
		}
	}
	return strings.Join(kept, "\n")
}

func tail(s string, n int) string {
	l := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(l) > n {
		l = l[len(l)-n:]
	}
	return strings.Join(l, "\n")
}

func isImage(b []byte) bool {
	return bytes.HasPrefix(b, []byte("\x89PNG\r\n\x1a\n")) || bytes.HasPrefix(b, []byte{0xFF, 0xD8, 0xFF})
}

func manifestText(o Options, now time.Time) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "d.idle support bundle, built %s\n", now.UTC().Format(time.RFC3339))
	fmt.Fprintf(&sb, "includes screenshot: %v\nincludes character save: %v\n", o.IncludeScreenshot, o.IncludeSave)
	sb.WriteString("Text files are redacted (home paths, IPs, host names, emails, tokens, keys) and scanned before this zip is made.\n")
	return sb.String()
}
