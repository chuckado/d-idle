package watch

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"d2lod-bot/internal/support"
)

type fakeSupport struct {
	files  []support.Input
	sender support.Sender
	opts   []support.Options
}

func (f *fakeSupport) Collect(o support.Options, local http.Handler) (support.Inputs, error) {
	f.opts = append(f.opts, o)
	return support.Inputs{Files: f.files, Extra: []string{"venusbox"}}, nil
}
func (f *fakeSupport) Sender() support.Sender { return f.sender }

func supportRig(t *testing.T, files []support.Input, recv http.Handler) (*SupportAPI, *fakeSupport, AnswerOptions, func()) {
	srv := httptest.NewServer(recv)
	d := &fakeSupport{files: files, sender: support.Sender{URL: srv.URL, Token: "up"}}
	api := NewSupportAPI()
	api.Bind(d)
	return api, d, AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}, srv.Close
}

var supportAuth = map[string]string{"Authorization": "Bearer tok"}

func TestSupportBuildSendFlow(t *testing.T) {
	var got int
	api, d, o, done := supportRig(t, []support.Input{{Name: "log.txt", Data: []byte("started from /ho" + "me/alice/x\nok\n")}, {Name: "screenshot.png", Data: []byte("\x89PNG\r\n\x1a\nzz")}},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			got = len(b)
			io.WriteString(w, `{"code":"ABCD2345"}`)
		}))
	defer done()
	bundle, send := api.BundleHandler(o), api.SendHandler(o)
	w := invPost(bundle, SupportBundleRoute, `{}`, supportAuth)
	var p SupportPreview
	json.Unmarshal(w.Body.Bytes(), &p)
	if w.Code != 200 || !p.CanSend || !p.CanSave || p.ID == "" || strings.Contains(p.LogPreview, "alice") || !strings.Contains(p.LogPreview, "<home>") {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	for _, f := range p.Files {
		if strings.Contains(f.Name, "screenshot") {
			t.Fatal("screenshot is off by default")
		}
	}
	if got != 0 {
		t.Fatal("building must never upload")
	}
	if w := invPost(send, SupportSendRoute, `{"id":"nope"}`, supportAuth); w.Code != 404 {
		t.Fatalf("unknown id: %d", w.Code)
	}
	w = invPost(send, SupportSendRoute, `{"id":"`+p.ID+`"}`, supportAuth)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "ABCD2345") || got == 0 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	if w := invPost(send, SupportSendRoute, `{"id":"`+p.ID+`"}`, supportAuth); w.Code != 404 {
		t.Fatal("a sent bundle is forgotten")
	}
	_ = d
}

func TestSupportOptionsPassThrough(t *testing.T) {
	api, d, o, done := supportRig(t, []support.Input{{Name: "log.txt", Data: []byte("x")}}, http.NotFoundHandler())
	defer done()
	invPost(api.BundleHandler(o), SupportBundleRoute, `{"includeScreenshot":true,"includeSave":false}`, supportAuth)
	invPost(api.BundleHandler(o), SupportBundleRoute, `{}`, supportAuth)
	if len(d.opts) != 2 || !d.opts[0].IncludeScreenshot || d.opts[0].IncludeSave || d.opts[1].IncludeScreenshot || d.opts[1].IncludeSave {
		t.Fatalf("%+v", d.opts)
	}
}

// A seeded secret that Redact cannot see (a literal of the machine the desk named in Extra, in a name that is not
// redacted by a rule) blocks sending: no upload, 409, the findings are shown without the secret.
func TestSupportBlockedBundleNeverSent(t *testing.T) {
	sent := false
	api, d, o, done := supportRig(t, []support.Input{{Name: "log.txt", Data: []byte("fine\n")}, {Name: "errors.txt", Data: []byte{0x55, 0xAA, 0x55, 0xAA, 'x'}}},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { sent = true }))
	defer done()
	w := invPost(api.BundleHandler(o), SupportBundleRoute, `{}`, supportAuth)
	var p SupportPreview
	json.Unmarshal(w.Body.Bytes(), &p)
	if w.Code != 200 || p.CanSend || p.CanSave || len(p.Findings) == 0 || p.ZipSize != 0 {
		t.Fatalf("%s", w.Body)
	}
	if w := invPost(api.SendHandler(o), SupportSendRoute, `{"id":"`+p.ID+`"}`, supportAuth); w.Code != 409 || sent {
		t.Fatalf("blocked bundle was sent: %d %v", w.Code, sent)
	}
	g := httptest.NewRecorder()
	api.Handler().ServeHTTP(g, httptest.NewRequest("GET", SupportDownloadRoute+"?id="+p.ID, nil))
	if g.Code != 404 {
		t.Fatalf("download of a blocked bundle: %d", g.Code)
	}
	_ = d
}

func TestSupportUnconfiguredSavesOnly(t *testing.T) {
	api := NewSupportAPI()
	api.Bind(&fakeSupport{files: []support.Input{{Name: "log.txt", Data: []byte("hello")}}})
	o := AnswerOptions{Token: "tok"}
	w := invPost(api.BundleHandler(o), SupportBundleRoute, `{}`, supportAuth)
	var p SupportPreview
	json.Unmarshal(w.Body.Bytes(), &p)
	if p.CanSend || !p.CanSave || p.Configured {
		t.Fatalf("%s", w.Body)
	}
	g := httptest.NewRecorder()
	api.Handler().ServeHTTP(g, httptest.NewRequest("GET", SupportDownloadRoute+"?id="+p.ID, nil))
	if g.Code != 200 || g.Header().Get("Content-Type") != "application/zip" || !strings.HasPrefix(g.Body.String(), "PK") || !strings.Contains(g.Header().Get("Content-Disposition"), "attachment") {
		t.Fatalf("%d %v", g.Code, g.Header())
	}
	if w := invPost(api.SendHandler(o), SupportSendRoute, `{"id":"`+p.ID+`"}`, supportAuth); w.Code != 502 || strings.Contains(w.Body.String(), "tok") {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
}

func TestSupportGates(t *testing.T) {
	api := NewSupportAPI()
	o := AnswerOptions{Token: "tok", Origins: []string{"http://ok.test"}}
	if w := invPost(api.BundleHandler(o), SupportBundleRoute, `{}`, supportAuth); w.Code != 503 {
		t.Errorf("unbound: %d", w.Code)
	}
	api.Bind(&fakeSupport{})
	if w := invPost(api.BundleHandler(o), SupportBundleRoute, `{}`, nil); w.Code != 401 {
		t.Errorf("no token: %d", w.Code)
	}
	if w := invPost(api.SendHandler(o), SupportSendRoute, `{"id":"x"}`, map[string]string{"Authorization": "Bearer tok", "Origin": "http://evil.test"}); w.Code != 403 {
		t.Errorf("bad origin: %d", w.Code)
	}
	if w := invPost(api.BundleHandler(o), SupportBundleRoute, `{"unknown":1}`, supportAuth); w.Code != 400 {
		t.Errorf("unknown field: %d", w.Code)
	}
	g := httptest.NewRecorder()
	api.Handler().ServeHTTP(g, httptest.NewRequest("POST", SupportRoute, nil))
	if g.Code != 405 {
		t.Errorf("viewer POST: %d", g.Code)
	}
}

func TestSupportKeepsTwoBundles(t *testing.T) {
	api := NewSupportAPI()
	api.Bind(&fakeSupport{files: []support.Input{{Name: "log.txt", Data: []byte("x")}}})
	o := AnswerOptions{Token: "tok"}
	var ids []string
	for i := 0; i < 3; i++ {
		var p SupportPreview
		json.Unmarshal(invPost(api.BundleHandler(o), SupportBundleRoute, `{}`, supportAuth).Body.Bytes(), &p)
		ids = append(ids, p.ID)
	}
	if api.get(ids[0]) != nil || api.get(ids[2]) == nil {
		t.Fatal("memory is bounded: the oldest goes")
	}
}
