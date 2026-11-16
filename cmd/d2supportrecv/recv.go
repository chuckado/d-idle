package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	maxZip       = 20 << 20 // same as internal/support.MaxZip
	maxEntries   = 64
	maxUnpacked  = 100 << 20                          // zip bomb guard (sum of declared uncompressed sizes)
	codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 32 symbols, no 0/O/1/I
)

var codeRx = regexp.MustCompile(`^[A-Z2-9]{8}$`)

type config struct {
	dir        string // inbox, NOT served by any web server
	tokensFile string // one line per token: <sha256 hex> [label]; delete the line to revoke
	quota      int64  // total bytes in dir; the oldest files go first
	maxFiles   int    // and at most this many files
	tokenRate  int    // uploads per hour per token
	ipRate     int    // requests per hour per client address
	failRate   int    // failed authentications per hour per client address
	trustProxy bool   // read the client address from CF-Connecting-IP / X-Forwarded-For (behind the tunnel)
	now        func() time.Time
}

type server struct {
	cfg    config
	mu     sync.Mutex
	hashes map[string]bool
	mtime  time.Time
	ipN    map[string]*window
	failN  map[string]*window
	tokN   map[string]*window
}

type window struct {
	start time.Time
	n     int
}

func newServer(c config) *server {
	if c.now == nil {
		c.now = time.Now
	}
	return &server{cfg: c, ipN: map[string]*window{}, failN: map[string]*window{}, tokN: map[string]*window{}}
}

// allow counts one event in a one-hour window; false when over limit.
func (s *server) allow(m map[string]*window, key string, limit int) bool {
	now := s.cfg.now()
	w := m[key]
	if w == nil || now.Sub(w.start) >= time.Hour {
		w = &window{start: now}
		m[key] = w
		if len(m) > 10000 { // memory bound: forget windows that ended
			for k, v := range m {
				if now.Sub(v.start) >= time.Hour {
					delete(m, k)
				}
			}
		}
	}
	w.n++
	return w.n <= limit
}

// loadTokens re-reads the token file when it changed, so a revocation needs no restart.
func (s *server) loadTokens() {
	st, err := os.Stat(s.cfg.tokensFile)
	if err != nil {
		s.hashes = nil // a missing file means nobody may upload
		return
	}
	if s.hashes != nil && st.ModTime().Equal(s.mtime) {
		return
	}
	f, err := os.Open(s.cfg.tokensFile)
	if err != nil {
		s.hashes = nil
		return
	}
	defer f.Close()
	h := map[string]bool{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fs := strings.Fields(sc.Text())
		if len(fs) > 0 && !strings.HasPrefix(fs[0], "#") && len(fs[0]) == 64 {
			h[strings.ToLower(fs[0])] = true
		}
	}
	s.hashes, s.mtime = h, st.ModTime()
}

func hashToken(tok string) string {
	sum := sha256.Sum256([]byte(tok))
	return hex.EncodeToString(sum[:])
}

// tokenOK compares the hash of the bearer in constant time against every stored hash.
func (s *server) tokenOK(tok string) (string, bool) {
	if tok == "" {
		return "", false
	}
	got := hashToken(tok)
	ok := false
	for h := range s.hashes {
		if subtle.ConstantTimeCompare([]byte(h), []byte(got)) == 1 {
			ok = true
		}
	}
	return got, ok
}

func (s *server) clientIP(r *http.Request) string {
	if s.cfg.trustProxy {
		if v := r.Header.Get("CF-Connecting-IP"); v != "" {
			return v
		}
		if v := r.Header.Get("X-Forwarded-For"); v != "" {
			return strings.TrimSpace(strings.Split(v, ",")[0])
		}
	}
	h, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return h
}

func reply(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// ServeHTTP handles POST (the upload) and GET /health. Nothing here logs an address, a token or a code.
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/health") {
		io.WriteString(w, "ok\n")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		reply(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST only"})
		return
	}
	s.mu.Lock()
	ip := s.clientIP(r)
	if !s.allow(s.ipN, ip, s.cfg.ipRate) {
		s.mu.Unlock()
		reply(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit"})
		return
	}
	s.loadTokens()
	bearer := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer"))
	if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		bearer = ""
	}
	th, ok := s.tokenOK(bearer)
	if !ok {
		over := !s.allow(s.failN, ip, s.cfg.failRate)
		s.mu.Unlock()
		if over {
			reply(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit"})
			return
		}
		reply(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if !s.allow(s.tokN, th, s.cfg.tokenRate) {
		s.mu.Unlock()
		reply(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit"})
		return
	}
	s.mu.Unlock()

	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || (mt != "application/zip" && mt != "application/x-zip-compressed") {
		reply(w, http.StatusUnsupportedMediaType, map[string]string{"error": "send application/zip"})
		return
	}
	if r.ContentLength > maxZip {
		reply(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "too large"})
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxZip))
	if err != nil {
		reply(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "too large"})
		return
	}
	if err := checkZip(body); err != nil {
		reply(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	code, err := s.store(body)
	if err != nil {
		reply(w, http.StatusInsufficientStorage, map[string]string{"error": "could not store"})
		return
	}
	reply(w, http.StatusOK, map[string]string{"code": code})
}

// checkZip accepts only a readable zip of plain relative names within the entry and size limits. The receiver never
// unpacks it; the owner's fetch script does, after the same name check.
func checkZip(b []byte) error {
	if !bytes.HasPrefix(b, []byte("PK\x03\x04")) {
		return errString("not a zip")
	}
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return errString("not a zip")
	}
	if len(zr.File) == 0 || len(zr.File) > maxEntries {
		return errString("bad entry count")
	}
	var total uint64
	for _, f := range zr.File {
		n := f.Name
		if n == "" || strings.HasPrefix(n, "/") || strings.Contains(n, "\\") || strings.Contains(n, "\x00") || strings.HasSuffix(n, "/") {
			return errString("bad entry name")
		}
		for _, p := range strings.Split(n, "/") {
			if p == ".." || p == "." || p == "" {
				return errString("bad entry name")
			}
		}
		if f.Mode()&os.ModeSymlink != 0 {
			return errString("bad entry type")
		}
		total += f.UncompressedSize64
		if total > maxUnpacked {
			return errString("too large when unpacked")
		}
	}
	return nil
}

type errString string

func (e errString) Error() string { return string(e) }

// store writes the zip under a fresh random code (O_EXCL, mode 0600), then purges the oldest files over quota.
func (s *server) store(b []byte) (string, error) {
	if err := os.MkdirAll(s.cfg.dir, 0o700); err != nil {
		return "", err
	}
	for try := 0; try < 8; try++ {
		code, err := newCode()
		if err != nil {
			return "", err
		}
		f, err := os.OpenFile(filepath.Join(s.cfg.dir, code+".zip"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, werr := f.Write(b)
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		if werr != nil {
			os.Remove(f.Name())
			return "", werr
		}
		s.purge(code + ".zip")
		return code, nil
	}
	return "", errString("no free code")
}

func newCode() (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	out := make([]byte, 8)
	for i, x := range b {
		out[i] = codeAlphabet[int(x)%len(codeAlphabet)] // 256 is a multiple of 32: no modulo bias
	}
	return string(out), nil
}

// purge deletes the oldest .zip files until the inbox fits the quota and the file limit; keep is never removed.
func (s *server) purge(keep string) {
	ents, err := os.ReadDir(s.cfg.dir)
	if err != nil {
		return
	}
	type fi struct {
		name string
		size int64
		mod  time.Time
	}
	var fs []fi
	var total int64
	for _, e := range ents {
		if !strings.HasSuffix(e.Name(), ".zip") || !codeRx.MatchString(strings.TrimSuffix(e.Name(), ".zip")) {
			continue
		}
		if info, err := e.Info(); err == nil && info.Mode().IsRegular() {
			fs = append(fs, fi{e.Name(), info.Size(), info.ModTime()})
			total += info.Size()
		}
	}
	sort.Slice(fs, func(i, j int) bool { return fs[i].mod.Before(fs[j].mod) })
	n := len(fs)
	for _, f := range fs {
		if total <= s.cfg.quota && n <= s.cfg.maxFiles {
			break
		}
		if f.name == keep {
			continue
		}
		if os.Remove(filepath.Join(s.cfg.dir, f.name)) == nil {
			total -= f.size
			n--
		}
	}
}
