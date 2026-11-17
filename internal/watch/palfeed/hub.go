package palfeed

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Hub holds the newest frame and wakes the clients. It never queues: a slow client simply skips to the newest frame,
// because each client's delta is computed against the frame that client last received (latest frame wins).
type Hub struct {
	mu    sync.Mutex
	cur   *Frame
	wake  chan struct{} // closed and replaced on each Publish
	free  [][]byte      // pixel buffers of released frames, reused by Publish (no 1.9 MB allocation per frame)
	pubAt time.Time     // when the newest frame was published

	clients atomic.Int32
	bytes   atomic.Int64 // message bytes written to all clients, cumulative
	msgs    atomic.Int64
	keys    atomic.Int64 // messages that were key frames or as large as one (see Options.BigMsg)
	slow    atomic.Int64 // pauses the bandwidth guard added
	pubs    atomic.Int64 // frames published
}

// Clients is the number of connected stream clients; the host stops the ring lease when it is 0 and no MJPEG viewer is left.
func (h *Hub) Clients() int { return int(h.clients.Load()) }

// Stats are the cumulative counters (messages, bytes, big messages, guard pauses).
func (h *Hub) Stats() (msgs, bytes, big, paused int64) {
	return h.msgs.Load(), h.bytes.Load(), h.keys.Load(), h.slow.Load()
}

func NewHub() *Hub { return &Hub{wake: make(chan struct{})} }

// Age is the time since the newest frame was published (0 before the first): "ring lag" for the stats overlay.
func (h *Hub) Age() time.Duration {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.pubAt.IsZero() {
		return 0
	}
	return time.Since(h.pubAt)
}

// frameRef counts the holders of a published frame (the hub's current slot and every client's base and in-flight
// frame); the pixel buffer goes back to the free list when the last one lets go.
type frameRef struct{ n atomic.Int32 }

func (h *Hub) release(f *Frame) {
	if f == nil || f.ref == nil {
		return
	}
	if f.ref.n.Add(-1) == 0 {
		h.mu.Lock()
		if len(h.free) < 6 {
			h.free = append(h.free, f.Pix)
		}
		h.mu.Unlock()
	}
}

// buf returns a pixel buffer of n bytes from the free list or a new one. Caller holds no lock.
func (h *Hub) buf(n int) []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i := len(h.free) - 1; i >= 0; i-- {
		if cap(h.free[i]) >= n && cap(h.free[i]) < 2*n {
			b := h.free[i][:n]
			h.free = append(h.free[:i], h.free[i+1:]...)
			return b
		}
	}
	return make([]byte, n)
}

// Publish stores a copy of f as the newest frame. Safe from the one goroutine that reads the d2d ring. stride is the
// source's bytes per row (the ring header's stride; 0 or 4*W: packed); rows are repacked when it is larger.
func (h *Hub) Publish(f Frame, stride int) {
	n := 4 * f.W * f.H
	pix := h.buf(n)
	if stride > 4*f.W && len(f.Pix) >= stride*(f.H-1)+4*f.W {
		for y := 0; y < f.H; y++ {
			copy(pix[y*4*f.W:(y+1)*4*f.W], f.Pix[y*stride:])
		}
	} else {
		copy(pix, f.Pix)
	}
	ref := &frameRef{}
	ref.n.Store(1) // the hub's slot
	c := &Frame{W: f.W, H: f.H, Pix: pix, Seq: f.Seq, GameMS: f.GameMS, ref: ref}
	h.pubs.Add(1)
	h.mu.Lock()
	old := h.cur
	h.cur = c
	h.pubAt = time.Now()
	w := h.wake
	h.wake = make(chan struct{})
	h.mu.Unlock()
	close(w)
	h.release(old)
}

// latest returns the newest frame, RETAINED for the caller (who must release it) and the channel closed at the next publish.
func (h *Hub) latest() (*Frame, <-chan struct{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cur != nil && h.cur.ref != nil {
		h.cur.ref.n.Add(1)
	}
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
	// queueing bytes; the delta that follows is built from the newest frame (latest wins). Default 4 MiB/s (the
	// measured mean is 0.49 MB/s at 25 Hz, docs/video-feed.md section 5); negative: no guard.
	MaxBytesPerSec int
	// BigMsg is the size from which a message counts as key-sized (a panel open is 375 to 737 KB); after one the client
	// waits at least BigGap, so panel flicker cannot send a key-sized message more than about once a BigGap. Default
	// 200 KB and 150 ms.
	BigMsg int
	BigGap time.Duration
	// Touch is called with each connection and about once a second while a client is served; the host uses it to keep
	// the ring lease alive for pal-only viewers.
	Touch func()
	// Server returns the host's own numbers for the stats overlay (container CPU throttling, ring lag); nil: none.
	Server func() ServerStats
}

// ServerStats are numbers only the host knows. Zero values mean "not measured".
type ServerStats struct {
	ThrottlePct float64 `json:"throttlePct"` // share of the last second's CPU periods the container was throttled (cgroup cpu.stat)
	ThrottleMS  float64 `json:"throttleMs"`  // throttled time in the last second, ms
	CPUPct      float64 `json:"cpuPct"`      // this process' CPU as a share of one core over the last second
	RingFPS     float64 `json:"ringFps"`     // ring frames per second taken by the host
	AtMS        int64   `json:"atMs"`        // wall clock of the sample, unix ms (aligns server and page series)
}

// Snapshot is the hub's counters at one moment (cumulative since start) plus the host's numbers.
type Snapshot struct {
	Clients int         `json:"clients"`
	Msgs    int64       `json:"msgs"`
	Bytes   int64       `json:"bytes"`
	Big     int64       `json:"big"`   // key-sized messages (key frames and panel opens)
	Holds   int64       `json:"holds"` // times the guard held a client back
	Pubs    int64       `json:"pubs"`  // frames published by the ring reader
	AgeMS   float64     `json:"ageMs"` // age of the newest frame now
	Server  ServerStats `json:"server"`
	AtMS    int64       `json:"atMs"`
}

// Snapshot returns the counters now.
func (h *Hub) Snapshot() Snapshot {
	m, b, big, hold := h.Stats()
	return Snapshot{Clients: h.Clients(), Msgs: m, Bytes: b, Big: big, Holds: hold, Pubs: h.pubs.Load(),
		AgeMS: float64(h.Age().Microseconds()) / 1000, AtMS: time.Now().UnixMilli()}
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
		o.MaxBytesPerSec = 4 << 20
	}
	if o.BigMsg <= 0 {
		o.BigMsg = 200 << 10
	}
	if o.BigGap <= 0 {
		o.BigGap = 150 * time.Millisecond
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
		defer func() { h.release(base) }()
		deltas := 0
		var next time.Time
		var lenb [4]byte
		for {
			cur, wake := h.latest()
			if cur == nil || (base != nil && cur.Seq == base.Seq) {
				touch()
				select {
				case <-wake:
					h.release(cur)
					continue
				case <-time.After(time.Second):
					// A paused game publishes nothing: an empty delta (the same frame again, a few dozen bytes) is the
					// heartbeat that keeps the page's stall watchdog quiet and the write deadline honest.
					if cur == nil || base == nil {
						h.release(cur)
						continue
					}
				case <-r.Context().Done():
					h.release(cur)
					return
				}
			}
			if d := time.Until(next); d > 0 {
				h.release(cur)
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
				h.release(cur)
				return
			}
			binary.LittleEndian.PutUint32(lenb[:], uint32(len(m)))
			touch()
			rc.SetWriteDeadline(time.Now().Add(o.WriteTimeout)) // a client that stopped reading is dropped here
			if _, err := w.Write(lenb[:]); err != nil {
				h.release(cur)
				return
			}
			if _, err := w.Write(m); err != nil {
				h.release(cur)
				return
			}
			fl.Flush()
			h.release(base)
			base = cur
			deltas++
			h.msgs.Add(1)
			h.bytes.Add(int64(len(m)) + 4)
			// The pace is a schedule (the previous slot plus one period), not "now plus one period": the loop's own work
			// (encode, write) would otherwise push every slot later and a frame would be skipped now and then (30 fps
			// became 29 with judder). Behind schedule: restart from now, never catch up in a burst.
			now := time.Now()
			slot := next.Add(time.Second / time.Duration(fps))
			if slot.Before(now) {
				slot = now
			}
			var guard time.Duration
			if len(m) >= o.BigMsg {
				h.keys.Add(1)
				guard = o.BigGap
			}
			if o.MaxBytesPerSec > 0 {
				if g := time.Duration(int64(len(m)) * int64(time.Second) / int64(o.MaxBytesPerSec)); g > time.Second/time.Duration(fps) {
					guard = max(guard, g)
					h.slow.Add(1)
				}
			}
			next = slot
			if g := now.Add(guard); g.After(next) {
				next = g
			}
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
		if o.Touch != nil {
			o.Touch() // a probe keeps the ring awake, so the first frame can arrive
		}
		cur, _ := h.latest()
		defer h.release(cur)
		if cur == nil {
			http.Error(w, "no frame yet", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		// "stats" (cumulative counters and the host's own numbers) is for the page's stream monitor, polled about once a second
		st := h.Snapshot()
		if o.Server != nil {
			st.Server = o.Server()
		}
		b, _ := json.Marshal(struct {
			PaletteFeed bool     `json:"paletteFeed"`
			Version     int      `json:"version"`
			Width       int      `json:"width"`
			Height      int      `json:"height"`
			Stats       Snapshot `json:"stats"`
		}{true, 1, cur.W, cur.H, st})
		_, _ = w.Write(b)
	})
}

func itoa(n int) string { return strconv.Itoa(n) }
