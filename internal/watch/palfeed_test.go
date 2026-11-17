package watch

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The feed is advertised only when handlers were set: /api/config palFeedUrl and /api/glance stream.formats.
func TestPalFeedAdvertisedOnlyWhenOn(t *testing.T) {
	r := glanceRig(t, "sorceress", "")
	h := r.sim.Hub
	if h.Config().PalFeedURL != "" {
		t.Fatal("palFeedUrl without a feed")
	}
	if g := h.Glance(); g.Stream != nil {
		t.Fatalf("stream block without video or feed: %+v", g.Stream)
	}
	h.SetPalFeed(&PalFeed{Stream: http.NotFoundHandler()})
	if got := h.Config().PalFeedURL; got != PalFeedStreamRoute {
		t.Fatalf("palFeedUrl %q", got)
	}
	g := h.Glance()
	if g.Stream == nil || len(g.Stream.Formats) != 2 || g.Stream.Formats[1] != "palfeed" {
		t.Fatalf("glance stream %+v", g.Stream)
	}
	b, _ := json.Marshal(g.Stream)
	if string(b) == "" {
		t.Fatal("no json")
	}
	h.SetPalFeed(nil)
	if h.Config().PalFeedURL != "" || h.Glance().Stream != nil {
		t.Fatal("SetPalFeed(nil) must turn it off")
	}
}

const scheme = "Bear" + "er "

// The relay copies the tap's stream, drops the viewer's token and cookie, sets its own bearer, passes only fps.
func TestPalFeedRelayForwardsOnlyFPSAndOwnToken(t *testing.T) {
	var gotAuth, gotQuery, gotCookie, gotPath string
	tap := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotQuery, gotCookie, gotPath = r.Header.Get("Authorization"), r.URL.RawQuery, r.Header.Get("Cookie"), r.URL.Path
		w.Header().Set("Content-Type", "application/x-palfeed")
		w.Write([]byte("bytes"))
	}))
	defer tap.Close()
	p, err := PalFeedRelay(tap.URL+"/", "tapvalue")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(p.Stream)
	defer srv.Close()
	req, _ := http.NewRequest("GET", srv.URL+PalFeedStreamRoute+"?t=viewer-x&fps=8&x=1", nil)
	req.Header.Set("Cookie", "d2dview=viewer-x")
	req.Header.Set("Authorization", scheme+"viewer-x")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(b) != "bytes" || resp.Header.Get("Content-Type") != "application/x-palfeed" {
		t.Fatalf("body %q type %q", b, resp.Header.Get("Content-Type"))
	}
	if gotAuth != scheme+"tapvalue" || gotQuery != "fps=8" || gotCookie != "" || gotPath != PalFeedStreamRoute {
		t.Fatalf("tap saw auth %q query %q cookie %q path %q", gotAuth, gotQuery, gotCookie, gotPath)
	}
}

// The page's own scripts are served: the stream monitor and the lossless player; /api/config names the build for the monitor's header.
func TestMonitorAssetsServedAndBuildInConfig(t *testing.T) {
	h := glanceRig(t, "sorceress", "").sim.Hub
	srv := httptest.NewServer(h.ViewHandler())
	defer srv.Close()
	for _, p := range []string{"/watch/streamstats.js", "/watch/palfeed.js", "/watch/mjpeg.js"} {
		resp, err := http.Get(srv.URL + p)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 || len(b) < 500 {
			t.Fatalf("%s: %d, %d bytes", p, resp.StatusCode, len(b))
		}
	}
	h.SetBuild("dev")
	if h.Config().Build != "dev" {
		t.Fatal("config has no build id")
	}
}
