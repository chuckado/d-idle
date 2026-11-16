package palfeed

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// synth draws a frame of flat areas and noise so tiles differ; size is deliberately not a multiple of the tile.
func synth(w, h int, seed int64, seq uint32) Frame {
	r := rand.New(rand.NewSource(seed))
	pix := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := (y*w + x) * 4
			v := byte((x / 7) ^ (y / 5))
			if r.Intn(9) == 0 {
				v = byte(r.Intn(256))
			}
			pix[o], pix[o+1], pix[o+2], pix[o+3] = v, v/2+byte(y), byte(x), 255
		}
	}
	return Frame{W: w, H: h, Pix: pix, Seq: seq, GameMS: seq * 40}
}

func TestKeyThenDeltasRoundTrip(t *testing.T) {
	for _, tile := range []int{8, 16, 32} {
		a := synth(103, 77, 1, 1)
		b := a
		b.Pix = append([]byte(nil), a.Pix...)
		b.Seq = 2
		for i := 0; i < 40; i++ { // change a few scattered pixels and one block
			o := (i*131%77*103 + i*17%103) * 4
			b.Pix[o] ^= 0x55
		}
		for y := 30; y < 50; y++ {
			for x := 60; x < 90; x++ {
				b.Pix[(y*103+x)*4+1] = 9
			}
		}
		enc := &Encoder{Tile: tile}
		var dec Decoder
		m, err := enc.Encode(nil, &a)
		if err != nil || m[3]&FlagKey == 0 {
			t.Fatalf("key: %v flags %d", err, m[3])
		}
		if err := dec.Apply(m); err != nil || !bytes.Equal(dec.Frame.Pix, a.Pix) {
			t.Fatalf("tile %d: key frame mismatch, err %v", tile, err)
		}
		keyLen := len(m)
		m, err = enc.Encode(&a, &b)
		if err != nil || m[3]&FlagKey != 0 {
			t.Fatalf("delta: %v", err)
		}
		if len(m) >= keyLen {
			t.Errorf("tile %d: delta %d bytes not smaller than key %d", tile, len(m), keyLen)
		}
		if err := dec.Apply(m); err != nil || !bytes.Equal(dec.Frame.Pix, b.Pix) || dec.Frame.Seq != 2 {
			t.Fatalf("tile %d: delta mismatch, err %v", tile, err)
		}
	}
}

func TestIdenticalFrameIsTiny(t *testing.T) {
	a := synth(800, 600, 2, 1)
	b := a
	b.Seq = 2
	enc := &Encoder{}
	m, _ := enc.Encode(&a, &b)
	if len(m) > HeaderSize+64 {
		t.Errorf("identical frame costs %d bytes", len(m))
	}
}

func TestDeltaNeedsItsBase(t *testing.T) {
	a, b, c := synth(64, 48, 3, 1), synth(64, 48, 4, 2), synth(64, 48, 5, 3)
	enc := &Encoder{}
	var dec Decoder
	if err := dec.Apply(mustEnc(t, enc, nil, &a)); err != nil {
		t.Fatal(err)
	}
	// a client that skipped frame 2 gets a delta from 1 to 3, built by the server from what the client holds
	if err := dec.Apply(mustEnc(t, enc, &b, &c)); !errors.Is(err, ErrNeedKey) {
		t.Fatalf("delta on the wrong base: %v", err)
	}
	if dec.Frame.Seq != 1 || !bytes.Equal(dec.Frame.Pix, a.Pix) {
		t.Fatal("a refused delta changed the frame")
	}
	if err := dec.Apply(mustEnc(t, enc, &a, &c)); err != nil || !bytes.Equal(dec.Frame.Pix, c.Pix) {
		t.Fatalf("skipping a frame: %v", err)
	}
	var fresh Decoder
	if err := fresh.Apply(mustEnc(t, enc, &a, &c)); !errors.Is(err, ErrNeedKey) {
		t.Fatalf("joining on a delta: %v", err)
	}
}

func mustEnc(t *testing.T, e *Encoder, base, cur *Frame) []byte {
	t.Helper()
	m, err := e.Encode(base, cur)
	if err != nil {
		t.Fatal(err)
	}
	return append([]byte(nil), m...)
}

func TestBadMessages(t *testing.T) {
	var d Decoder
	for _, m := range [][]byte{nil, []byte("PF"), bytes.Repeat([]byte{1}, 40)} {
		if d.Apply(m) == nil {
			t.Errorf("accepted %q", m)
		}
	}
	a := synth(32, 32, 1, 1)
	m := mustEnc(t, &Encoder{}, nil, &a)
	if d.Apply(m[:len(m)-3]) == nil {
		t.Error("accepted a truncated message")
	}
	if _, err := (&Encoder{}).Encode(nil, &Frame{W: 4, H: 4, Pix: make([]byte, 3)}); err == nil {
		t.Error("encoded a frame with the wrong pixel count")
	}
}

func readMsg(t *testing.T, r *bufio.Reader) []byte {
	t.Helper()
	var l [4]byte
	if _, err := io.ReadFull(r, l[:]); err != nil {
		t.Fatalf("length: %v", err)
	}
	m := make([]byte, binary.LittleEndian.Uint32(l[:]))
	if _, err := io.ReadFull(r, m); err != nil {
		t.Fatalf("message: %v", err)
	}
	return m
}

func TestStreamLatestFrameWinsAndAuth(t *testing.T) {
	h := NewHub()
	o := Options{Auth: func(r *http.Request) bool { return r.Header.Get("Authorization") == "Bearer ok" }, MaxFPS: 1000}
	srv := httptest.NewServer(h.Stream(o))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil || resp.StatusCode != 401 {
		t.Fatalf("no token: %v %v", err, resp)
	}
	resp.Body.Close()

	f1 := synth(96, 64, 1, 1)
	h.Publish(f1, 0)
	req, _ := http.NewRequest("GET", srv.URL, nil)
	req.Header.Set("Authorization", "Bearer ok")
	resp, err = http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 || resp.Header.Get("Content-Type") != "application/x-palfeed" {
		t.Fatalf("stream: %v %v", err, resp)
	}
	defer resp.Body.Close()
	br := bufio.NewReader(resp.Body)
	var dec Decoder
	m := readMsg(t, br)
	if m[3]&FlagKey == 0 || dec.Apply(m) != nil || !bytes.Equal(dec.Frame.Pix, f1.Pix) {
		t.Fatal("join did not start with the key frame")
	}
	// two frames published back to back: the client may see 2 and 3 or only 3, but ends on 3 and every message applies
	f2, f3 := synth(96, 64, 2, 2), synth(96, 64, 3, 3)
	h.Publish(f2, 0)
	h.Publish(f3, 0)
	deadline := time.Now().Add(3 * time.Second)
	for dec.Frame.Seq != 3 && time.Now().Before(deadline) {
		if err := dec.Apply(readMsg(t, br)); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Equal(dec.Frame.Pix, f3.Pix) {
		t.Fatalf("client ended on seq %d", dec.Frame.Seq)
	}
}

func TestCapability(t *testing.T) {
	h := NewHub()
	o := Options{Auth: func(*http.Request) bool { return true }}
	get := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.Capability(o).ServeHTTP(w, httptest.NewRequest("GET", "/api/palfeed", nil))
		return w
	}
	if w := get(); w.Code != 503 {
		t.Fatalf("before a frame: %d", w.Code)
	}
	h.Publish(synth(800, 600, 1, 1), 0)
	if w := get(); w.Code != 200 || w.Body.String() != `{"paletteFeed":true,"version":1,"width":800,"height":600}` {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	if w := httptest.NewRecorder(); true {
		(&Hub{}).Capability(Options{}).ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
		if w.Code != 401 {
			t.Fatalf("nil Auth must refuse, got %d", w.Code)
		}
	}
}

// TestFixtures keeps testdata/ (read by tools/videofeed/palfeed.test.js, the JS decoder's byte-compat test) in step
// with the encoder: it fails if the files differ, and rewrites them with D2FEED_WRITE_FIXTURES=1.
func TestFixtures(t *testing.T) {
	a := synth(96, 70, 11, 1)
	b := synth(96, 70, 11, 2)
	b.Pix = append([]byte(nil), a.Pix...)
	for y := 20; y < 41; y++ {
		for x := 33; x < 70; x++ {
			b.Pix[(y*96+x)*4] ^= 0x3c
		}
	}
	enc := &Encoder{Tile: 16}
	wa, wb := world(100, 100, 1), world(105, 97, 2)
	files := map[string][]byte{
		"a.rgba": a.Pix, "b.rgba": b.Pix,
		"key.pf":   mustEnc(t, enc, nil, &a),
		"delta.pf": mustEnc(t, enc, &a, &b),
		// the scrolled pair: modes 2 and 3 in the delta; wb.sha256 is the SHA-256 of b's RGB bytes
		"wkey.pf":   mustEnc(t, enc, nil, &wa),
		"wdelta.pf": mustEnc(t, enc, &wa, &wb),
		"wb.sha256": []byte(fmt.Sprintf("%x\n", sha256.Sum256(rgbOnly(wb.Pix)))),
	}
	for name, want := range files {
		p := "testdata/" + name
		if os.Getenv("D2FEED_WRITE_FIXTURES") == "1" {
			if err := os.MkdirAll("testdata", 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, want, 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		got, err := os.ReadFile(p)
		if err != nil || !bytes.Equal(got, want) {
			t.Errorf("%s is stale or missing (%v): run D2FEED_WRITE_FIXTURES=1 go test ./internal/watch/palfeed", p, err)
		}
	}
}

func TestPublishRepacksStride(t *testing.T) {
	h := NewHub()
	src := make([]byte, 20*3)
	for y := 0; y < 3; y++ {
		for x := 0; x < 4; x++ {
			src[y*20+x*4] = byte(y*10 + x)
		}
	}
	h.Publish(Frame{W: 4, H: 3, Pix: src, Seq: 1}, 20)
	cur, _ := h.latest()
	if len(cur.Pix) != 48 || cur.Pix[2*16+3*4] != 23 {
		t.Fatalf("repack: len %d", len(cur.Pix))
	}
}

// world draws the view at scroll (ox, oy) of a fixed textured world, with a static HUD strip at the bottom, so a scroll
// moves everything except the HUD.
func world(ox, oy int, seq uint32) Frame {
	const w, h, hud = 240, 160, 24
	pix := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			o := (y*w + x) * 4
			if y >= h-hud {
				pix[o], pix[o+1], pix[o+2], pix[o+3] = byte(x), 200, byte(y), 255
				continue
			}
			wx, wy := x+ox, y+oy
			v := uint32(wx/3)*2654435761 ^ uint32(wy/3)*40503
			pix[o], pix[o+1], pix[o+2], pix[o+3] = byte(v>>8), byte(v>>16), byte(v>>24), 255
		}
	}
	return Frame{W: w, H: h, Pix: pix, Seq: seq, GameMS: seq * 40}
}

func TestScrollIsFoundAndCheap(t *testing.T) {
	a, b := world(100, 100, 1), world(105, 97, 2) // the view moved right 5 and up 3: content shifts by (-5, +3)
	if dx, dy := EstimateShift(&a, &b); dx != -5 || dy != 3 {
		t.Fatalf("shift %d,%d, want -5,3", dx, dy)
	}
	var withShift, without int
	for _, noShift := range []bool{false, true} {
		enc := &Encoder{NoShift: noShift}
		var dec Decoder
		if err := dec.Apply(mustEnc(t, enc, nil, &a)); err != nil {
			t.Fatal(err)
		}
		m := mustEnc(t, enc, &a, &b)
		if err := dec.Apply(m); err != nil || !bytes.Equal(rgbOnly(dec.Frame.Pix), rgbOnly(b.Pix)) {
			t.Fatalf("noShift=%v: mismatch, err %v", noShift, err)
		}
		if noShift {
			without = len(m)
		} else {
			withShift = len(m)
		}
	}
	if withShift*5 > without {
		t.Errorf("scroll coding saves too little: %d B with, %d B without", withShift, without)
	}
	// and a chain of scrolls keeps decoding exactly (the decoder swaps buffers)
	enc := &Encoder{}
	var dec Decoder
	prev := world(0, 0, 1)
	if err := dec.Apply(mustEnc(t, enc, nil, &prev)); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 12; i++ {
		cur := world(i*4, i*2, uint32(i+1))
		if err := dec.Apply(mustEnc(t, enc, &prev, &cur)); err != nil || !bytes.Equal(rgbOnly(dec.Frame.Pix), rgbOnly(cur.Pix)) {
			t.Fatalf("scroll %d: %v", i, err)
		}
		prev = cur
	}
}

// The guard: a key-sized message holds the client back (BigGap), bytes/s is capped, Touch and the client count work.
func TestGuardPacesBigMessagesAndCountsClients(t *testing.T) {
	h := NewHub()
	touched := make(chan struct{}, 8)
	o := Options{Auth: func(*http.Request) bool { return true }, MaxFPS: 1000, BigMsg: 1 << 10, BigGap: 150 * time.Millisecond,
		MaxBytesPerSec: -1, Touch: func() {
			select {
			case touched <- struct{}{}:
			default:
			}
		}}
	srv := httptest.NewServer(h.Stream(o))
	defer srv.Close()
	h.Publish(synth(96, 64, 1, 1), 0)
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	br := bufio.NewReader(resp.Body)
	t0 := time.Now()
	readMsg(t, br) // the key frame (random pixels: well above BigMsg)
	select {
	case <-touched:
	default:
		t.Fatal("Touch not called on connect")
	}
	if h.Clients() != 1 {
		t.Fatalf("clients %d", h.Clients())
	}
	h.Publish(synth(96, 64, 2, 2), 0) // a noise frame: key-sized delta, must wait out BigGap
	readMsg(t, br)
	if d := time.Since(t0); d < 130*time.Millisecond {
		t.Fatalf("second big message after %v, want >= BigGap", d)
	}
	if _, _, big, _ := h.Stats(); big < 2 {
		t.Fatalf("big %d", big)
	}
	resp.Body.Close()
	deadline := time.Now().Add(2 * time.Second)
	for h.Clients() != 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if h.Clients() != 0 {
		t.Fatal("client not released")
	}
}

func TestGuardByteCapSlowsClient(t *testing.T) {
	h := NewHub()
	o := Options{Auth: func(*http.Request) bool { return true }, MaxFPS: 1000, BigMsg: 1 << 30, MaxBytesPerSec: 50 << 10}
	srv := httptest.NewServer(h.Stream(o))
	defer srv.Close()
	h.Publish(synth(96, 64, 1, 1), 0)
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	br := bufio.NewReader(resp.Body)
	t0 := time.Now()
	m := readMsg(t, br)
	h.Publish(synth(96, 64, 2, 2), 0)
	readMsg(t, br)
	want := time.Duration(len(m)) * time.Second / (50 << 10)
	if d := time.Since(t0); d < want*8/10 {
		t.Fatalf("second message after %v, cap wants about %v", d, want)
	}
	if _, _, _, slow := h.Stats(); slow == 0 {
		t.Fatal("guard pause not counted")
	}
}

func TestHeartbeatWhileNothingChanges(t *testing.T) {
	h := NewHub()
	srv := httptest.NewServer(h.Stream(Options{Auth: func(*http.Request) bool { return true }}))
	defer srv.Close()
	f := synth(96, 64, 1, 1)
	h.Publish(f, 0)
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	br := bufio.NewReader(resp.Body)
	var dec Decoder
	if err := dec.Apply(readMsg(t, br)); err != nil {
		t.Fatal(err)
	}
	t0 := time.Now()
	m := readMsg(t, br) // no new frame: the heartbeat comes after about a second
	if d := time.Since(t0); d < 800*time.Millisecond || d > 3*time.Second {
		t.Fatalf("heartbeat after %v", d)
	}
	if m[3]&FlagKey != 0 || len(m) > 200 {
		t.Fatalf("heartbeat is %d bytes, key %v", len(m), m[3]&FlagKey != 0)
	}
	if err := dec.Apply(m); err != nil || !bytes.Equal(dec.Frame.Pix, f.Pix) {
		t.Fatalf("heartbeat does not apply: %v", err)
	}
}

// A source at 30 Hz is delivered at 30 Hz: the pace is a schedule, so the loop's own work never skips a frame.
func TestDeliversEveryFrameOfA30HzSource(t *testing.T) {
	h := NewHub()
	srv := httptest.NewServer(h.Stream(Options{Auth: func(*http.Request) bool { return true }}))
	defer srv.Close()
	base := synth(96, 64, 1, 1)
	h.Publish(base, 0)
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	stop := make(chan struct{})
	const frames = 90
	go func() {
		tk := time.NewTicker(time.Second / 30)
		defer tk.Stop()
		for i := 2; i < 2+frames; i++ {
			select {
			case <-tk.C:
				f := base
				f.Pix = append([]byte(nil), base.Pix...)
				f.Pix[(i*37)%len(f.Pix)] ^= 0xFF
				f.Seq = uint32(i)
				h.Publish(f, 0)
			case <-stop:
				return
			}
		}
	}()
	defer close(stop)
	br := bufio.NewReader(resp.Body)
	var dec Decoder
	got := 0
	deadline := time.Now().Add(5 * time.Second)
	for dec.Frame.Seq < frames+1 && time.Now().Before(deadline) {
		if err := dec.Apply(readMsg(t, br)); err != nil {
			t.Fatal(err)
		}
		got++
	}
	if got < frames-3 {
		t.Fatalf("%d of %d frames delivered at 30 Hz (the pace skipped frames)", got, frames)
	}
}
