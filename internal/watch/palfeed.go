package watch

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// The lossless picture feed on the watch screen (docs/video-feed.md section 9). The codec and the handlers are
// internal/watch/palfeed; this file only lets the Hub advertise the feed and relays it from the tap.

// PalFeedRoutes are the two routes of the feed; ViewHandler serves them when SetPalFeed gave it handlers.
var PalFeedRoutes = []string{PalFeedStreamRoute, PalFeedCapRoute}

const (
	PalFeedStreamRoute = "/stream.pfeed"
	PalFeedCapRoute    = "/api/palfeed"
)

// PalFeed holds the two handlers of the feed (the stream and the capability probe). Either may be nil.
type PalFeed struct{ Stream, Capability http.Handler }

// SetPalFeed turns the feed on for the page: /api/config carries palFeedUrl and /api/glance lists "palfeed" under
// stream.formats. A nil p or a nil Stream turns it off. Handlers that are set are served on PalFeedRoutes by ViewHandler
// (a host that mounts the routes itself, as d2dview does, passes handlers only for advertising: nil Capability and a
// Stream that it also mounts). Call before serving.
func (h *Hub) SetPalFeed(p *PalFeed) { h.pal = p }

func (h *Hub) palFeedURL() string {
	if h.pal == nil || h.pal.Stream == nil {
		return ""
	}
	return PalFeedStreamRoute
}

// PalFeedRelay serves the tap's feed through this viewer: the bytes are copied, not decoded, so the watch screen costs
// the d2bot process nothing but the copy. base is the tap's base URL, token its bearer token ("" none).
func PalFeedRelay(base, token string) (*PalFeed, error) {
	u, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil {
		return nil, err
	}
	rp := &httputil.ReverseProxy{
		FlushInterval: -1, // every message at once
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(u)
			r.Out.Header.Del("Cookie")
			r.Out.Header.Del("Authorization")
			if token != "" {
				r.Out.Header.Set("Authorization", "Bearer "+token)
			}
			r.Out.URL.RawQuery = keepFPS(r.In.URL.Query())
		},
	}
	return &PalFeed{Stream: rp, Capability: rp}, nil
}

// keepFPS passes only the low-rate query value to the tap: the page's token is checked by this viewer, not forwarded.
func keepFPS(q url.Values) string {
	out := url.Values{}
	if v := q.Get("fps"); v != "" {
		out.Set("fps", v)
	}
	return out.Encode()
}

// streamFormats is stream.formats of /api/glance.
func (h *Hub) streamFormats() []string {
	f := []string{"mjpeg"}
	if h.palFeedURL() != "" {
		f = append(f, "palfeed")
	}
	return f
}
