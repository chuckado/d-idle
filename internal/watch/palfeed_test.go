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

// The relay copies the tap's stream, drops the viewer's token and cookie, sets its own bearer, passes only fps.
func TestPalFeedRelayForwardsOnlyFPSAndOwnToken(t *testing.T) {
	var gotAuth, gotQuery, gotCookie, gotPath string
	tap := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotQuery, gotCookie, gotPath = r.Header.Get("Authorization"), r.URL.RawQuery, r.Header.Get("Cookie"), r.URL.Path
		w.Header().Set("Content-Type", "application/x-palfeed")
		w.Write([]byte("bytes"))
	}))
	defer tap.Close()
	p, err := PalFeedRelay(tap.URL+"/", "tap-secret")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(p.Stream)
	defer srv.Close()
	req, _ := http.NewRequest("GET", srv.URL+PalFeedStreamRoute+"?t=viewer-token&fps=8&x=1", nil)
	req.Header.Set("Cookie", "d2dview=viewer-token")
	req.Header.Set("Authorization", "Bearer viewer-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if string(b) != "bytes" || resp.Header.Get("Content-Type") != "application/x-palfeed" {
		t.Fatalf("body %q type %q", b, resp.Header.Get("Content-Type"))
	}
	if gotAuth != "Bearer tap-secret" || gotQuery != "fps=8" || gotCookie != "" || gotPath != PalFeedStreamRoute {
		t.Fatalf("tap saw auth %q query %q cookie %q path %q", gotAuth, gotQuery, gotCookie, gotPath)
	}
}
