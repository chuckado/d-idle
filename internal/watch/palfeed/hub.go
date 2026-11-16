package palfeed

import (
	"encoding/binary"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Hub holds the newest frame and wakes the clients. It never queues: a slow client simply skips to the newest frame,
// because each client's delta is computed against the frame that client last received (latest frame wins).
type Hub struct {
	mu   sync.Mutex
	cur  *Frame
	wake chan struct{} // closed and replaced on each Publish

	clients atomic.Int32
	bytes   atomic.Int64 // message bytes written to all clients, cumulative
	msgs    atomic.Int64
	keys    atomic.Int64 // messages that were key frames or as large as one (see Options.BigMsg)
	slow    atomic.Int64 // pauses the bandwidth guard added
}

// Clients is the number of connected stream clients; the host stops the ring lease when it is 0 and no MJPEG viewer is left.
func (h *Hub) Clients() int { return int(h.clients.Load()) }

// Stats are the cumulative counters (messages, bytes, big messages, guard pauses).
func (h *Hub) Stats() (msgs, bytes, big, paused int64) {
	return h.msgs.Load(), h.bytes.Load(), h.keys.Load(), h.slow.Load()
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
	// WriteTimeout bounds each message's write: a viewer that stopped reading without closing is dropped and frees its
	// slot (the servers have no global WriteTimeout, it would cut every long stream). Default 10 s.
	WriteTimeout time.Duration
	// MaxBytesPerSec is the per-client bandwidth guard: after a message of n bytes the next one waits at least
	// n / MaxBytesPerSec, so a heavy stretch (a key frame, a panel opening) lowers that client's rate instead of
	// queueing bytes; the delta that follows is built from the newest frame (latest wins). Default 2 MiB/s (the
	// measured mean is 0.49 MB/s at 25 Hz, docs/video-feed.md section 5); negative: no guard.
	MaxBytesPerSec int
	// BigMsg is the size from which a message counts as key-sized (a panel open is 375 to 737 KB); after one the client
	// waits at least BigGap, so panel flicker cannot send a key-sized message more than about once a BigGap. Default
	// 200 KB and 400 ms.
	BigMsg int
	BigGap time.Duration
	// Touch is called with each connection and about once a second while a client is served; the host uses it to keep
	// the ring lease alive for pal-only viewers.
	Touch func()
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
	if o.WriteTimeout <= 0 {
		o.WriteTimeout = 10 * time.Second
	}
	if o.MaxBytesPerSec == 0 {
		o.MaxBytesPerSec = 2 << 20
	}
	if o.BigMsg <= 0 {
		o.BigMsg = 200 << 10
	}
	if o.BigGap <= 0 {
		o.BigGap = 400 * time.Millisecond
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
		h.clients.Add(1)
		defer h.clients.Add(-1)
		rc := http.NewResponseController(w)
		var touched time.Time
		touch := func() {
			if o.Touch != nil && time.Since(touched) >= time.Second {
				touched = time.Now()
				o.Touch()
			}
		}
		touch()
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
				touch()
				select {
				case <-wake:
					continue
				case <-time.After(time.Second):
					// A paused game publishes nothing: an empty delta (the same frame again, a few dozen bytes) is the
					// heartbeat that keeps the page's stall watchdog quiet and the write deadline honest.
					if cur == nil || base == nil {
						continue
					}
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
			touch()
			rc.SetWriteDeadline(time.Now().Add(o.WriteTimeout)) // a client that stopped reading is dropped here
			if _, err := w.Write(lenb[:]); err != nil {
				return
			}
			if _, err := w.Write(m); err != nil {
				return
			}
			fl.Flush()
			base = cur
			deltas++
			h.msgs.Add(1)
			h.bytes.Add(int64(len(m)) + 4)
			gap := time.Second / time.Duration(fps)
			if len(m) >= o.BigMsg {
				h.keys.Add(1)
				gap = max(gap, o.BigGap)
			}
			if o.MaxBytesPerSec > 0 {
				if g := time.Duration(int64(len(m)) * int64(time.Second) / int64(o.MaxBytesPerSec)); g > gap {
					gap = g
					h.slow.Add(1)
				}
			}
			next = time.Now().Add(gap)
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
