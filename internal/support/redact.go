// Package support builds, redacts, scans and sends the support bundle (docs/support-bundle.md). The bundle is an
// allowlist of named files; every text file goes through Redact and then Scan, and a bundle with a finding is never
// sent (fail closed).
package support

import (
	"regexp"
	"strings"
)

// The rules are shared by Redact (replace) and Scan (detect what is left). A placeholder never matches its own rule.
type rule struct {
	class string
	rx    *regexp.Regexp
	repl  string
	keep  func(m string) bool // true: not a hit (a match that is allowed to stay)
}

var secretName = `(?:[\w.-]*(?:token|secret|passw(?:or)?d|passphrase|cookie|credential|auth(?:orization)?|api[_-]?key|cd[_-]?key|private[_-]?key|license[_-]?key|access[_-]?key)[\w.-]*|key)`

var rules = []rule{
	{"pem", regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?(?:-----END [A-Z ]*PRIVATE KEY-----|$)`), "<key-block>", nil},
	{"bearer", regexp.MustCompile(`(?i)\b(?:bearer|basic)\s+[A-Za-z0-9._~+/=-]{6,}`), "Bearer <token>", nil},
	// name = value, name: value, "name": "value" for names that say secret; the value goes, the name stays
	{"secret", regexp.MustCompile(`(?i)(["']?\b` + secretName + `["']?\s*[:=]\s*)("[^"\n]*"|'[^'\n]*'|[^\s,;&}\]"']+)`), `${1}"<secret>"`, func(m string) bool {
		return strings.Contains(m, "<secret>") || strings.Contains(m, "<token>") || strings.Contains(m, "<redacted>")
	}},
	{"urltoken", regexp.MustCompile(`(?i)([?&](?:token|t|key|auth|sig|code|access_token)=)[^&\s"'<]+`), `${1}<token>`, nil},
	{"urlpath", regexp.MustCompile(`/p/[A-Za-z0-9_-]{12,}`), `/p/<token>`, nil},
	{"uuid", regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`), "<uuid>", nil},
	{"cdkey", regexp.MustCompile(`\b[A-Z0-9]{4,6}(?:-[A-Z0-9]{4,6}){2,4}\b`), "<cdkey>", func(m string) bool { return !strings.ContainsAny(m, "0123456789") }},
	{"keyfile", regexp.MustCompile(`(?i)[\w.-]*\.key\b`), "<keyfile>", nil},
	{"savefile", regexp.MustCompile(`(?i)[\w.-]+\.d2s\b`), "<save>.d2s", func(m string) bool { return strings.EqualFold(m, "<save>.d2s") }},
	{"email", regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)+`), "<email>", nil},
	{"homepath", regexp.MustCompile(`(?i)(?:/home|/Users)/[^/\s"'<:;)]+|[A-Z]:\\Users\\[^\\\s"'<:;)]+`), "<home>", nil},
	{"ip6", regexp.MustCompile(`(?i)\b(?:[0-9a-f]{1,4}:){3,7}[0-9a-f]{1,4}\b`), "<ip6>", nil},
	{"ip", regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`), "<ip>", func(m string) bool {
		for _, o := range strings.Split(m, ".") {
			if len(o) > 3 || (len(o) == 3 && o > "255") {
				return true
			}
		}
		return false
	}},
	{"host", regexp.MustCompile(`(?i)\b[a-z0-9_-]*birbhub[a-z0-9_.-]*|\b(?:[a-z0-9-]+\.)+(?:lan|local|internal|home|ts\.net)\b`), "<host>", nil},
	{"hex", regexp.MustCompile(`\b[0-9a-fA-F]{32,}\b`), "<hex>", func(m string) bool { return len(m) == 40 }}, // a 40-hex build id is public
	{"token", regexp.MustCompile(`[A-Za-z0-9+/_-]{32,}={0,2}`), "<token>", func(m string) bool { return !randomToken(m) }},
}

// randomToken: long, with digits and both cases, so identifiers and plain paths do not count (as cmd/leakcheck).
func randomToken(s string) bool {
	var lo, up, dg int
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z':
			lo++
		case c >= 'A' && c <= 'Z':
			up++
		case c >= '0' && c <= '9':
			dg++
		}
	}
	return lo > 0 && up > 0 && dg > 0
}

// Redact replaces secrets, personal data and machine identity with placeholders. extra are literal strings of this
// machine (host name, user name, data directory) that go too.
func Redact(s string, extra ...string) string {
	for _, e := range extra {
		if rx := literalRx(e); rx != nil {
			s = rx.ReplaceAllString(s, "<local>")
		}
	}
	for _, r := range rules {
		s = r.rx.ReplaceAllStringFunc(s, func(m string) string {
			if r.keep != nil && r.keep(m) {
				return m
			}
			return r.rx.ReplaceAllString(m, r.repl)
		})
	}
	return s
}

var wordLit = regexp.MustCompile(`^[\w-]+$`)

// literalRx matches a literal of this machine; a plain word (user or host name) only as a whole word, so a short
// name does not eat the middle of other words. nil for literals shorter than 3 characters.
func literalRx(e string) *regexp.Regexp {
	if len(e) < 3 {
		return nil
	}
	q := regexp.QuoteMeta(e)
	if wordLit.MatchString(e) {
		q = `(?i)\b` + q + `\b`
	}
	return regexp.MustCompile(q)
}
