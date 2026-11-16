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
	"strconv"
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
	"decisions.json": true, "deaths.json": true, "readiness.json": true, "danger.json": true,
	"streamstats.json": true, "ledger.json": true, "log.txt": true, "errors.txt": true,
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

// File is one line of the preview. Excerpt is the whole redacted text of a small file, else its first and last
// lines; Redactions counts what was replaced in it (the whole text is paged through Bundle.Page).
type File struct {
	Name       string `json:"name"`
	Size       int    `json:"size"`
	Lines      int    `json:"lines,omitempty"`
	Redactions int    `json:"redactions,omitempty"`
	Excerpt    string `json:"excerpt,omitempty"`
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
	text       map[string]string
}

// PageSize is the size of one page of Bundle.Page.
const PageSize = 16 << 10

// Page returns one page of a text file of the bundle exactly as it is in the zip, and whether more follows.
func (b *Bundle) Page(name string, page int) (string, bool, bool) {
	t, ok := b.text[name]
	if !ok || page < 0 || page*PageSize > len(t) {
		return "", false, false
	}
	end := (page + 1) * PageSize
	if end >= len(t) {
		return t[page*PageSize:], false, true
	}
	return t[page*PageSize : end], true, true
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
	b := &Bundle{Created: in.Now.UTC(), Options: o, text: map[string]string{}}
	lits := NewLiterals(in.Extra)
	counts := map[string]int{}
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
			txt, n := prepare(f.Name, f.Data, in, lits)
			if len(txt) > MaxTextFile {
				txt = txt[len(txt)-MaxTextFile:]
			}
			out = append(out, entry{f.Name, txt})
			counts[f.Name] = n
			b.text[f.Name] = string(txt)
			b.Findings = append(b.Findings, ScanL(f.Name, string(txt), lits)...)
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
	b.text["MANIFEST.txt"] = manifest
	out = append([]entry{{"MANIFEST.txt", []byte(manifest)}}, out...)
	for _, e := range out {
		f := File{Name: e.name, Size: len(e.data)}
		if t, ok := b.text[e.name]; ok {
			f.Lines, f.Redactions, f.Excerpt = strings.Count(t, "\n")+1, counts[e.name], excerpt(t)
		}
		b.Files = append(b.Files, f)
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

// prepare redacts one allowlisted file and returns the number of replacements. JSON is parsed, loses every
// secret-named key, has each string value redacted before it is marshalled (so escaping such as \u0026 never hides a
// shape) and goes through the text pass once more; text is time-filtered (logs) and redacted.
func prepare(name string, data []byte, in Inputs, lits Literals) ([]byte, int) {
	if strings.HasSuffix(name, ".json") {
		var v any
		if json.Unmarshal(data, &v) != nil {
			return []byte("{\"error\": \"not valid JSON, omitted\"}\n"), 0
		}
		total := 0
		v = redactStrings(dropSecretKeys(v), lits, &total)
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		enc.Encode(v)
		s, n := RedactN(buf.String(), lits)
		return []byte(s), total + n
	}
	s := string(data)
	if name == "log.txt" || name == "errors.txt" {
		s = recent(s, in.Now, in.LogHours)
	}
	s, n := RedactN(s, lits)
	return []byte(s), n
}

func redactStrings(v any, lits Literals, n *int) any {
	switch x := v.(type) {
	case string:
		s, c := RedactN(x, lits)
		*n += c
		return s
	case map[string]any:
		for k, e := range x {
			x[k] = redactStrings(e, lits, n)
		}
	case []any:
		for i, e := range x {
			x[i] = redactStrings(e, lits, n)
		}
	}
	return v
}

// excerpt is the whole text of a small file, else its first and last 12 lines.
func excerpt(t string) string {
	if len(t) <= 4096 {
		return t
	}
	l := strings.Split(strings.TrimRight(t, "\n"), "\n")
	if len(l) <= 24 {
		return strings.Join(l, "\n")
	}
	return strings.Join(l[:12], "\n") + "\n... (" + strconv.Itoa(len(l)-24) + " lines not shown here; open the file to read them) ...\n" + strings.Join(l[len(l)-12:], "\n")
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
