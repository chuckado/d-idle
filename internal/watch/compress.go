package watch

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
)

// The Vault page loads 12 files and polls an API; the app used to send them raw with Cache-Control no-store, so every visit paid
// every byte again (docs/probes/vault-perf-2026-11-23.md). Embedded assets are immutable per build: they get an ETag from their
// bytes, "no-cache" (the browser revalidates, the answer is a bodyless 304 until the build changes) and a gzip copy made once.
type packed struct {
	etag    string
	raw, gz []byte
}

var packedAssets sync.Map // asset path -> *packed

func packAsset(name string, b []byte) *packed {
	if v, ok := packedAssets.Load(name); ok {
		return v.(*packed)
	}
	sum := sha256.Sum256(b)
	p := &packed{etag: `"` + hex.EncodeToString(sum[:8]) + `"`, raw: b}
	var z bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&z, gzip.BestCompression)
	zw.Write(b)
	zw.Close()
	if z.Len() < len(b) {
		p.gz = z.Bytes()
	}
	packedAssets.Store(name, p)
	return p
}

func acceptsGzip(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}

// serveAsset writes an embedded file with the validators and the gzip copy described above. hdr sets the page's own headers.
func serveAsset(w http.ResponseWriter, r *http.Request, name string, b []byte, hdr func(http.ResponseWriter)) {
	hdr(w)
	if strings.HasPrefix(w.Header().Get("Content-Type"), "text/html") {
		// pages stay as they were: a middleware (the update banner) edits their bytes, so a validator or a gzip copy would be wrong
		w.Header().Set("Cache-Control", "no-store")
		w.Write(b)
		return
	}
	p := packAsset(name, b)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("ETag", p.etag)
	w.Header().Add("Vary", "Accept-Encoding")
	if m := r.Header.Get("If-None-Match"); m != "" && strings.Contains(m, p.etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	if p.gz != nil && acceptsGzip(r) {
		w.Header().Set("Content-Encoding", "gzip")
		w.Write(p.gz)
		return
	}
	w.Write(p.raw)
}

// gzipJSON compresses an API answer when the client accepts it and the body is big enough to be worth it.
func gzipJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !acceptsGzip(r) {
			next.ServeHTTP(w, r)
			return
		}
		g := &gzipWriter{ResponseWriter: w}
		next.ServeHTTP(g, r)
		g.finish()
	})
}

type gzipWriter struct {
	http.ResponseWriter
	code int
	buf  bytes.Buffer
}

func (g *gzipWriter) WriteHeader(c int) { g.code = c }
func (g *gzipWriter) Write(b []byte) (int, error) {
	return g.buf.Write(b)
}

func (g *gzipWriter) finish() {
	h := g.ResponseWriter.Header()
	if g.code == 0 {
		g.code = http.StatusOK
	}
	if g.buf.Len() >= 1024 && h.Get("Content-Encoding") == "" && strings.Contains(h.Get("Content-Type"), "json") {
		var z bytes.Buffer
		zw, _ := gzip.NewWriterLevel(&z, 5)
		zw.Write(g.buf.Bytes())
		zw.Close()
		h.Set("Content-Encoding", "gzip")
		h.Add("Vary", "Accept-Encoding")
		g.ResponseWriter.WriteHeader(g.code)
		g.ResponseWriter.Write(z.Bytes())
		return
	}
	g.ResponseWriter.WriteHeader(g.code)
	g.ResponseWriter.Write(g.buf.Bytes())
}
