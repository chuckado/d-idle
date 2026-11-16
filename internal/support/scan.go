package support

import (
	"bytes"
	"strings"
)

// Finding is one thing the scan still sees after redaction. Context is the line with the match blanked, never the
// match itself.
type Finding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Class   string `json:"class"`
	Context string `json:"context"`
}

// d2sMagic starts every character save file (V: the 1.14d .d2s header, docs/d2-1.14d-memory.md).
var d2sMagic = []byte{0x55, 0xAA, 0x55, 0xAA}

// Scan looks for what Redact should have removed. It is the second line of defence: it runs on the redacted text
// with the same rule table plus the literal extras, and anything it finds blocks the upload.
func Scan(file string, text string, extra ...string) []Finding {
	var out []Finding
	for i, line := range strings.Split(text, "\n") {
		for _, e := range extra {
			if rx := literalRx(e); rx != nil && rx.MatchString(line) {
				out = append(out, Finding{file, i + 1, "local", ctx(rx.ReplaceAllString(line, "[***]"), extra)})
			}
		}
		for _, r := range rules {
			if r.class == "secret" { // a secret assignment is a hit only when its value is not a placeholder
				for _, m := range r.rx.FindAllStringSubmatch(line, -1) {
					if !strings.Contains(m[2], "<") {
						out = append(out, Finding{file, i + 1, r.class, ctx(strings.ReplaceAll(line, m[2], "[***]"), extra)})
					}
				}
				continue
			}
			for _, m := range r.rx.FindAllString(line, -1) {
				if r.keep != nil && r.keep(m) {
					continue
				}
				out = append(out, Finding{file, i + 1, r.class, ctx(strings.ReplaceAll(line, m, "[***]"), extra)})
			}
		}
	}
	return out
}

// ctx is the line as the sheet shows it: the match blanked, everything else the scan could still find redacted,
// so a finding never echoes a second secret of the same line.
func ctx(line string, extra []string) string {
	line = Redact(line, extra...)
	if len(line) > 160 {
		line = line[:160]
	}
	return line
}

// hasSaveMagic reports whether data looks like a character save.
func hasSaveMagic(data []byte) bool { return bytes.HasPrefix(data, d2sMagic) }
