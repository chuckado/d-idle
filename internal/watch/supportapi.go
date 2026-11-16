package watch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"time"

	"d2lod-bot/internal/support"
)

// The support bundle routes (docs/support-bundle.md): POST /api/support/bundle {includeScreenshot, includeSave} builds,
// redacts and scans a bundle and answers the preview (file list, sizes, the log excerpt, findings); POST
// /api/support/send {id} uploads exactly that bundle and answers the support code; GET /api/support/download?id= is the
// "save to file instead" path; GET /api/support says whether uploading is set up. Never an upload without the press of
// Send, never a bundle with a finding.
const (
	SupportRoute         = "/api/support"
	SupportBundleRoute   = "/api/support/bundle"
	SupportSendRoute     = "/api/support/send"
	SupportDownloadRoute = "/api/support/download"
	SupportFileRoute     = "/api/support/file"
	supportSchema        = "watch.support/1"
	supportKeep          = 15 * time.Minute
)

// SupportDesk is what the routes need from the bot side (cmd/d2bot/supportwire.go).
type SupportDesk interface {
	// Collect gathers the allowlisted inputs. local answers the viewer's own GET routes in process (glance, readiness).
	Collect(o support.Options, local http.Handler) (support.Inputs, error)
	// Sender is the upload endpoint (support.EnvURL, support.EnvToken); an unconfigured sender means save-only.
	Sender() support.Sender
}

// SupportAPI holds the desk and the last built bundles (in memory, a quarter of an hour).
type SupportAPI struct {
	mu      sync.Mutex
	d       SupportDesk
	local   http.Handler
	last    map[string]*support.Bundle
	now     func() time.Time
	build   sync.Mutex      // one build or send at a time
	sending map[string]bool // ids with an upload in flight (a second press is refused)
}

func NewSupportAPI() *SupportAPI {
	return &SupportAPI{last: map[string]*support.Bundle{}, sending: map[string]bool{}, now: time.Now}
}

func (a *SupportAPI) Bind(d SupportDesk) { a.mu.Lock(); a.d = d; a.mu.Unlock() }

func (a *SupportAPI) desk() SupportDesk { a.mu.Lock(); defer a.mu.Unlock(); return a.d }

func (a *SupportAPI) setLocal(h http.Handler) { a.mu.Lock(); a.local = h; a.mu.Unlock() }

// SupportPreview is the answer of the build route.
type SupportPreview struct {
	Schema     string            `json:"schema"`
	ID         string            `json:"id"`
	Files      []support.File    `json:"files"`
	LogPreview string            `json:"logPreview"`
	Findings   []support.Finding `json:"findings"`
	Dropped    []string          `json:"dropped,omitempty"`
	ZipSize    int               `json:"zipSize"`
	CanSend    bool              `json:"canSend"`
	CanSave    bool              `json:"canSave"`
	Configured bool              `json:"configured"`
	Options    support.Options   `json:"options"`
}

func (a *SupportAPI) remember(b *support.Bundle) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.now()
	for id, o := range a.last {
		if now.Sub(o.Created) > supportKeep {
			delete(a.last, id)
		}
	}
	for len(a.last) >= 2 {
		var oldest string
		for id, o := range a.last {
			if oldest == "" || o.Created.Before(a.last[oldest].Created) {
				oldest = id
			}
		}
		delete(a.last, oldest)
	}
	a.last[b.ID] = b
}

func (a *SupportAPI) get(id string) *support.Bundle {
	a.mu.Lock()
	defer a.mu.Unlock()
	b := a.last[id]
	if b != nil && a.now().Sub(b.Created) > supportKeep {
		delete(a.last, id)
		return nil
	}
	return b
}

// Handler serves the viewer GETs: the status and the download.
func (a *SupportAPI) Handler() http.Handler {
	return readOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "support bundles are not available"})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		if r.URL.Path == SupportFileRoute {
			// one page of one file exactly as the zip holds it: the sheet's "open the file"
			pg, _ := strconv.Atoi(r.URL.Query().Get("page"))
			b := a.get(r.URL.Query().Get("id"))
			if b == nil {
				writeJSON(w, http.StatusNotFound, policyReply{Code: "gone", Error: "build the bundle again"})
				return
			}
			txt, more, ok := b.Page(r.URL.Query().Get("name"), pg)
			if !ok {
				writeJSON(w, http.StatusNotFound, policyReply{Code: "file", Error: "no such file or page"})
				return
			}
			writeJSON(w, 200, map[string]any{"schema": supportSchema, "name": r.URL.Query().Get("name"), "page": pg, "more": more, "text": txt})
			return
		}
		if r.URL.Path == SupportDownloadRoute {
			b := a.get(r.URL.Query().Get("id"))
			if b == nil || len(b.Zip) == 0 {
				writeJSON(w, http.StatusNotFound, policyReply{Code: "gone", Error: "build the bundle again"})
				return
			}
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", `attachment; filename="didle-support-`+b.Created.Format("20060102-1504")+`.zip"`)
			w.Write(b.Zip)
			return
		}
		writeJSON(w, 200, map[string]any{"schema": supportSchema, "configured": d.Sender().Configured()})
	}))
}

// BundleHandler serves POST /api/support/bundle.
func (a *SupportAPI) BundleHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var opt support.Options
		if !policyGate(o, w, r, &opt) {
			return
		}
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "support bundles are not available"})
			return
		}
		if o.Token == "" && !containsStr(o.Origins, r.Header.Get("Origin")) {
			writeJSON(w, http.StatusForbidden, policyReply{Code: "token", Error: "building needs the watch page itself"})
			return
		}
		a.build.Lock()
		defer a.build.Unlock()
		a.mu.Lock()
		local := a.local
		a.mu.Unlock()
		in, err := d.Collect(opt, local)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, policyReply{Code: "collect", Error: "could not gather the files"})
			return
		}
		b, err := support.Build(in, opt)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, policyReply{Code: "build", Error: err.Error()})
			return
		}
		a.remember(b)
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, SupportPreview{Schema: supportSchema, ID: b.ID, Files: b.Files, LogPreview: b.LogPreview, Findings: nonNil(b.Findings),
			Dropped: b.Dropped, ZipSize: b.ZipSize, CanSend: b.Sendable() && d.Sender().Configured(), CanSave: b.Sendable(),
			Configured: d.Sender().Configured(), Options: opt})
	})
}

func nonNil(f []support.Finding) []support.Finding {
	if f == nil {
		return []support.Finding{}
	}
	return f
}

// SendHandler serves POST /api/support/send {id}: the bundle that was shown, nothing else.
func (a *SupportAPI) SendHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ID string `json:"id"`
		}
		if !policyGate(o, w, r, &body) {
			return
		}
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "support bundles are not available"})
			return
		}
		// with no token configured the gate lets any local process in; an upload then needs the page's own origin
		if o.Token == "" && !containsStr(o.Origins, r.Header.Get("Origin")) {
			writeJSON(w, http.StatusForbidden, policyReply{Code: "token", Error: "sending needs the watch page itself"})
			return
		}
		b := a.get(body.ID)
		a.mu.Lock()
		busy := b != nil && a.sending[b.ID]
		if b != nil && !busy {
			a.sending[b.ID] = true
		}
		a.mu.Unlock()
		if busy {
			writeJSON(w, http.StatusConflict, policyReply{Code: "busy", Error: "this bundle is already being sent"})
			return
		}
		if b != nil {
			defer func() { a.mu.Lock(); delete(a.sending, b.ID); a.mu.Unlock() }()
		}
		switch {
		case b == nil:
			writeJSON(w, http.StatusNotFound, policyReply{Code: "gone", Error: "build the bundle again, then send"})
			return
		case !b.Sendable():
			writeJSON(w, http.StatusConflict, policyReply{Code: "blocked", Error: "this bundle still holds private-looking text and will not be sent"})
			return
		}
		a.build.Lock()
		defer a.build.Unlock()
		ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
		defer cancel()
		code, err := d.Sender().Send(ctx, b.Zip)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, policyReply{Code: "send", Error: err.Error()})
			return
		}
		a.mu.Lock()
		delete(a.last, b.ID)
		a.mu.Unlock()
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, map[string]string{"schema": supportSchema, "code": code})
	})
}

// fetchLocal GETs one of the viewer's own routes in process; the support desk uses it for the glance and friends.
func FetchLocal(h http.Handler, path string) []byte {
	if h == nil {
		return nil
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != 200 {
		return nil
	}
	return rec.Body.Bytes()
}
