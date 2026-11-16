package palfeed

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
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
	h.Publish(f1)
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
	h.Publish(f2)
	h.Publish(f3)
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
	h.Publish(synth(800, 600, 1, 1))
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
	files := map[string][]byte{
		"a.rgba": a.Pix, "b.rgba": b.Pix,
		"key.pf":   mustEnc(t, enc, nil, &a),
		"delta.pf": mustEnc(t, enc, &a, &b),
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
