package main

import (
	"net/http"
	"net/url"
	"strings"

	"d2lod-bot/internal/support"
	"d2lod-bot/internal/watch"
)

// supportFixture serves the support bundle sheet (internal/watch/web/supportui.js) for the fit pages: POST
// /api/support/bundle answers a preview whose variant is the page's query support=blocked|nosend (the Referer), else a
// normal one (files, a long log line, upload set up). It never builds or sends anything.
func supportFixture(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == watch.SupportRoute {
		js(w, map[string]any{"schema": "watch.support/1", "configured": true})
		return
	}
	variant := ""
	if ref, err := url.Parse(r.Referer()); err == nil {
		variant = ref.Query().Get("support")
	}
	p := watch.SupportPreview{Schema: "watch.support/1", ID: "fixture", Configured: variant != "nosend", CanSend: variant == "", CanSave: variant != "blocked",
		Findings: []support.Finding{}, ZipSize: 41234,
		Files: []support.File{{Name: "MANIFEST.txt", Size: 312}, {Name: "errors.txt", Size: 1840}, {Name: "glance.json", Size: 5312}, {Name: "knobs.json", Size: 9120},
			{Name: "log.txt", Size: 182044, Lines: 2100, Redactions: 14, Excerpt: "2026-11-15T10:00:01Z INFO entered Blood Moor\n... (2076 lines not shown here; open the file to read them) ...\n2026-11-15T10:00:03Z INFO killed 7 monsters"}, {Name: "readiness.json", Size: 2210}, {Name: "system.json", Size: 190}, {Name: "version.json", Size: 120}},
		LogPreview: strings.Join([]string{"2026-11-15T10:00:01Z INFO entered Blood Moor life=212/300",
			"2026-11-15T10:00:02Z WARN path blocked, retrying with a wider search around <local>/instances/i1/state a very long unbroken token-free line to test wrapping on a phone width",
			"2026-11-15T10:00:03Z INFO killed 7 monsters, picked up 2 items"}, "\n")}
	if variant == "blocked" {
		p.CanSend, p.ZipSize = false, 0
		p.Findings = []support.Finding{{File: "log.txt", Line: 42, Class: "ip", Context: "2026-11-15 peer [***] refused"}, {File: "errors.txt", Line: 3, Class: "email", Context: "mail to [***] failed"}}
	}
	js(w, p)
}
