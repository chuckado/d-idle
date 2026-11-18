package watch

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"sync"
	"time"
)

// Profiler captures an on-demand CPU profile of the bot (D2BOT_PPROF_HTTP=1; off by default, never mounted otherwise).
// It sits on the answer listener, behind the same token or passcode gate and Origin rule as the other write routes:
//
//	POST /api/debug/cpuprofile?secs=N   start (N 1..60, default 30); 409 while one runs
//	POST /api/debug/cpuprofile/stop     stop early and keep the file
//	GET  /api/debug/cpuprofile          status as JSON
//	GET  /api/debug/cpuprofile/file     the newest finished profile (go tool pprof -top FILE)
type Profiler struct {
	Dir string // where cpu-<time>.pprof files go; the newest keepProfiles stay

	mu    sync.Mutex
	f     *os.File
	stop  chan struct{}
	path  string // running profile, or the last finished one
	until time.Time
}

const (
	MaxProfileSecs  = 60
	keepProfiles    = 5
	defaultProfSecs = 30
)

// ProfilerRoutes are the paths the answer listener hands to Profiler.Handler.
var ProfilerRoutes = []string{"/api/debug/cpuprofile", "/api/debug/cpuprofile/stop", "/api/debug/cpuprofile/file"}

// Start begins a profile of secs seconds (clamped to 1..MaxProfileSecs) and returns its file.
func (p *Profiler) Start(secs int) (string, error) {
	secs = min(max(secs, 1), MaxProfileSecs)
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.f != nil {
		return "", errors.New("a profile is already running")
	}
	if err := os.MkdirAll(p.Dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(p.Dir, "cpu-"+time.Now().UTC().Format("20060102T150405Z")+".pprof")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return "", err
	}
	if err := pprof.StartCPUProfile(f); err != nil { // another profile (D2DCTL_PPROF) is running
		f.Close()
		os.Remove(path)
		return "", err
	}
	p.f, p.path, p.stop = f, path, make(chan struct{})
	p.until = time.Now().Add(time.Duration(secs) * time.Second)
	go p.run(p.stop, time.Duration(secs)*time.Second)
	p.prune()
	return path, nil
}

func (p *Profiler) run(stop chan struct{}, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
	case <-stop:
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stop == stop { // not a newer profile that started since
		p.stopLocked()
	}
}

// Stop ends the running profile; a no-op when none runs.
func (p *Profiler) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stopLocked()
}

func (p *Profiler) stopLocked() {
	if p.f == nil {
		return
	}
	pprof.StopCPUProfile()
	p.f.Close()
	p.f = nil
	close(p.stop)
}

func (p *Profiler) prune() {
	m, _ := filepath.Glob(filepath.Join(p.Dir, "cpu-*.pprof"))
	for i := 0; i+keepProfiles < len(m); i++ { // names sort by time
		os.Remove(m[i])
	}
}

type profStatus struct {
	Running bool    `json:"running"`
	File    string  `json:"file,omitempty"`
	Left    float64 `json:"leftSecs,omitempty"`
}

func (p *Profiler) status() profStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	return profStatus{Running: p.f != nil, File: p.path, Left: max(0, time.Until(p.until).Seconds())}
}

// Handler serves ProfilerRoutes. o carries the token and origins of the answer listener.
func (p *Profiler) Handler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" && !containsStr(o.Origins, origin) {
			http.Error(w, "origin not allowed", http.StatusForbidden)
			return
		}
		if !o.tokenOK(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		post := r.Method == http.MethodPost
		switch {
		case r.URL.Path == "/api/debug/cpuprofile" && post:
			secs := defaultProfSecs
			if v := r.URL.Query().Get("secs"); v != "" {
				n, err := strconv.Atoi(v)
				if err != nil {
					http.Error(w, "secs must be a number", http.StatusBadRequest)
					return
				}
				secs = n
			}
			if _, err := p.Start(secs); err != nil {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			writeJSON(w, http.StatusAccepted, p.status())
		case r.URL.Path == "/api/debug/cpuprofile/stop" && post:
			p.Stop()
			writeJSON(w, http.StatusOK, p.status())
		case r.URL.Path == "/api/debug/cpuprofile" && r.Method == http.MethodGet:
			writeJSON(w, http.StatusOK, p.status())
		case r.URL.Path == "/api/debug/cpuprofile/file" && r.Method == http.MethodGet:
			st := p.status()
			if st.Running {
				http.Error(w, "still running", http.StatusConflict)
				return
			}
			if st.File == "" {
				http.Error(w, "no profile taken yet", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(st.File)))
			http.ServeFile(w, r, st.File)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
