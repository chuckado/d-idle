package watch

import (
	"log"
	"net"
	"strings"

	"d2lod-bot/internal/questboard"
)

// remoteClass names where a request came from without printing the address: loopback, lan or other. A proxy on the same
// host makes every client look like loopback; the user agent family and the confirm flag still tell the clients apart.
func remoteClass(addr string) string {
	h, _, err := net.SplitHostPort(addr)
	if err != nil {
		h = addr
	}
	ip := net.ParseIP(strings.Trim(h, "[]"))
	switch {
	case ip == nil:
		return "other"
	case ip.IsLoopback():
		return "loopback"
	case ip.IsPrivate(), ip.IsLinkLocalUnicast():
		return "lan"
	}
	return "other"
}

// agentFamily reduces a User-Agent to a family word; the raw header (build numbers, device names) is never kept.
func agentFamily(ua string) string {
	l := strings.ToLower(ua)
	for _, f := range []struct{ sub, name string }{
		{"curl/", "curl"}, {"go-http-client", "go"}, {"python", "python"}, {"edg/", "edge"}, {"firefox", "firefox"},
		{"chrome", "chrome"}, {"safari", "safari"},
	} {
		if strings.Contains(l, f.sub) {
			return f.name
		}
	}
	if ua == "" {
		return ""
	}
	return "other"
}

// safeID keeps what a quest id can contain; anything else (a probe, a log-injection attempt) becomes "?".
func safeID(s string) string {
	if s == "" {
		return ""
	}
	if len(s) > 64 {
		return "?"
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return "?"
		}
	}
	return s
}

// logQuestAction writes the INFO line of one API action. It never prints a header, a token or an address.
func logQuestAction(e questboard.AuditEvent) {
	log.Printf("quests: action=%s quest=%s chunk=%s path=%s remote=%s agent=%s confirm=%t ok=%t already=%t status=%d code=%s reason=%q",
		e.Action, e.Quest, e.Chunk, e.Path, e.Remote, e.Agent, e.Confirm, e.OK, e.Already, e.Status, e.Code, e.Reason)
}

// safePath keeps a request path to the characters a quest route has.
func safePath(p string) string {
	if len(p) > 100 {
		return "?"
	}
	for _, c := range p {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-' || c == '/') {
			return "?"
		}
	}
	return p
}
