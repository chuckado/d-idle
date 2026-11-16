package watch

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// A viewer that stops reading must not slow the capture or grow a queue: the handler blocks in Write, the viewer's
// channel holds one frame (latest wins) and the misses are counted.
func TestVideoSlowViewerMissesFramesAndNothingQueues(t *testing.T) {
	v, _ := NewVideo(VideoOptions{Capture: pumpCapture(make([]byte, 256<<10), 2*time.Millisecond), MaxFrame: 1 << 20, WriteTimeout: 5 * time.Second})
	srv := httptest.NewServer(http.HandlerFunc(v.handleStream))
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close() // never read
	time.Sleep(800 * time.Millisecond)
	st := v.Stats()
	if in := v.st.in.Load(); in < 100 {
		t.Fatalf("capture delivered only %d frames in 0.8 s: a stalled viewer slowed it", in)
	}
	if st.Missed == 0 {
		t.Fatalf("stats %+v: a stalled viewer must count missed frames", st)
	}
}

func TestVideoFPSCapAndDuplicateSkip(t *testing.T) {
	v, _ := NewVideo(VideoOptions{Capture: pumpCapture([]byte{0xFF, 0xD8, 1, 0xFF, 0xD9}, 5*time.Millisecond, true)})
	srv := httptest.NewServer(http.HandlerFunc(v.handleStream))
	defer srv.Close()
	resp, err := http.Get(srv.URL + "?fps=5")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	rd := bufio.NewReader(resp.Body)
	end, parts := time.Now().Add(1500*time.Millisecond), 0
	for time.Now().Before(end) {
		readPart(t, rd)
		parts++
	}
	if parts > 12 {
		t.Fatalf("%d parts in 1.5 s with fps=5 and a duplicate-only source", parts)
	}
	if v.Stats().Same == 0 {
		t.Fatal("identical frames were not skipped")
	}
}
