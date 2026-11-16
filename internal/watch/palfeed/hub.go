package palfeed

import (
	"encoding/binary"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Hub holds the newest frame and wakes the clients. It never queues: a slow client simply skips to the newest frame,
// because each client's delta is computed against the frame that client last received (latest frame wins).
type Hub struct {
	mu   sync.Mutex
	cur  *Frame
	wake chan struct{} // closed and replaced on each Publish
}

func NewHub() *Hub { return &Hub{wake: make(chan struct{})} }

// Publish stores a copy of f as the newest frame. Safe from the one goroutine that reads the d2d ring. stride is the
// source's bytes per row (the ring header's stride; 0 or 4*W: packed); rows are repacked when it is larger.
func (h *Hub) Publish(f Frame, stride int) {
	pix := f.Pix
	if stride > 4*f.W && len(pix) >= stride*(f.H-1)+4*f.W {
		pix = make([]byte, 4*f.W*f.H)
		for y := 0; y < f.H; y++ {
			copy(pix[y*4*f.W:(y+1)*4*f.W], f.Pix[y*stride:])
		}
	} else {
		pix = append([]byte(nil), pix...)
	}
	c := Frame{W: f.W, H: f.H, Pix: pix, Seq: f.Seq, GameMS: f.GameMS}
	h.mu.Lock()
	h.cur = &c
	w := h.wake
	h.wake = make(chan struct{})
	h.mu.Unlock()
	close(w)
}

func (h *Hub) latest() (*Frame, <-chan struct{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.cur, h.wake
}

// Options of the HTTP handlers.
type Options struct {
	Auth     func(*http.Request) bool // the viewer's existing watch-token check; nil refuses everything
	MaxFPS   int                      // per-client cap, default 60
	KeyEvery int                      // force a key frame after this many deltas (resync insurance), default 250
	Tile     int                      // see Encoder.Tile
}

// Stream serves GET: a body of messages, each prefixed with its u32 length (little endian), Content-Type
// application/x-palfeed. The first message is a key frame. Ends when the client goes away.
func (h *Hub) Stream(o Options) http.Handler {
	if o.MaxFPS <= 0 {
		o.MaxFPS = 60
	}
	if o.KeyEvery <= 0 {
		o.KeyEvery = 250
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if o.Auth == nil || !o.Auth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		fl, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "no streaming", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/x-palfeed")
		w.Header().Set("Cache-Control", "no-store")
		enc := &Encoder{Tile: o.Tile}
		fps := o.MaxFPS
		if v, err := strconv.Atoi(r.URL.Query().Get("fps")); err == nil && v >= 1 && v < fps { // the page's Low rate toggle
			fps = v
		}
		var base *Frame
		deltas := 0
		var next time.Time
		var lenb [4]byte
		for {
			cur, wake := h.latest()
			if cur == nil || (base != nil && cur.Seq == base.Seq) {
				select {
				case <-wake:
					continue
				case <-r.Context().Done():
					return
				}
			}
			if d := time.Until(next); d > 0 {
				select {
				case <-time.After(d):
				case <-r.Context().Done():
					return
				}
				cur, _ = h.latest() // newest after the pacing sleep
			}
			ref := base
			if deltas >= o.KeyEvery {
				ref, deltas = nil, 0
			}
			m, err := enc.Encode(ref, cur)
			if err != nil {
				return
			}
			binary.LittleEndian.PutUint32(lenb[:], uint32(len(m)))
			// A client that stopped reading blocks here; the server's per-write deadline (as watch.Video) drops it.
			if _, err := w.Write(lenb[:]); err != nil {
				return
			}
			if _, err := w.Write(m); err != nil {
				return
			}
			fl.Flush()
			base = cur
			deltas++
			next = time.Now().Add(time.Second / time.Duration(fps))
		}
	})
}

// Capability serves GET: {"paletteFeed":true,"version":1,"width":W,"height":H} once a frame exists, 503 before (the page
// then keeps the MJPEG img). Same auth as Stream.
func (h *Hub) Capability(o Options) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if o.Auth == nil || !o.Auth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		cur, _ := h.latest()
		if cur == nil {
			http.Error(w, "no frame yet", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(`{"paletteFeed":true,"version":1,"width":` + itoa(cur.W) + `,"height":` + itoa(cur.H) + `}`))
	})
}

func itoa(n int) string { return strconv.Itoa(n) }
