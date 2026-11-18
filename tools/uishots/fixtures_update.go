package main

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"strings"

	"d2lod-bot/internal/selfupdate"
	"d2lod-bot/internal/watch"
)

// The update banner of a real install, for the fit suite: update=available (or ready) in the page's query shows the banner
// the way /watch/update.js draws it (the poll carries the query as the Referer). Without it the status is idle, no banner.
type fixtureUpdater struct{ r *http.Request }

func (fixtureUpdater) Status() selfupdate.Status      { return selfupdate.Status{State: "idle"} }
func (fixtureUpdater) Check(context.Context) error    { return nil }
func (fixtureUpdater) Download(context.Context) error { return nil }
func (fixtureUpdater) Apply(string, bool) error       { return nil }
func (fixtureUpdater) Cancel()                        {}
func (fixtureUpdater) Rollback() error                { return nil }
func (fixtureUpdater) Inbox() error                   { return nil }
func (fixtureUpdater) MarkSeen() error                { return nil }

func updateStatusFixture(w http.ResponseWriter, r *http.Request) {
	st := selfupdate.Status{State: "idle", Running: "0.1.2", CheckEnabled: true}
	if ref, err := url.Parse(r.Referer()); err == nil {
		switch ref.Query().Get("update") {
		case "available":
			st.Available = &selfupdate.Available{Version: "0.1.3", Notes: "Fixes for the watch screen and the character screen."}
		case "ready":
			st.State = "ready"
			st.Available = &selfupdate.Available{Version: "0.1.3", Notes: "Fixes for the watch screen and the character screen."}
		}
	}
	js(w, st)
}

// withUpdateScript adds the banner script to the watch HTML like the real server does.
func withUpdateScript(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/watch") || strings.Contains(r.URL.Path[1:], ".") {
			next.ServeHTTP(w, r)
			return
		}
		rec := &captured{h: http.Header{}, code: 200}
		next.ServeHTTP(rec, r)
		body := rec.buf.Bytes()
		if strings.HasPrefix(rec.h.Get("Content-Type"), "text/html") {
			body = bytes.Replace(body, []byte("</body>"), []byte(`<script src="/watch/update.js"></script></body>`), 1)
			rec.h.Del("Content-Length")
		}
		for k, v := range rec.h {
			w.Header()[k] = v
		}
		w.WriteHeader(rec.code)
		w.Write(body)
	})
}

type captured struct {
	h    http.Header
	buf  bytes.Buffer
	code int
}

func (c *captured) Header() http.Header         { return c.h }
func (c *captured) Write(b []byte) (int, error) { return c.buf.Write(b) }
func (c *captured) WriteHeader(code int)        { c.code = code }

var _ = watch.UpdateHandler
