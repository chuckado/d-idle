package watch

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"d2lod-bot/internal/points"
)

// The points routes (docs/points.md; design review 2026-11-19 04 section 7 P0):
//
//	GET  /api/hero/points                the sheet: banked points, per-stat rows with the preview line, worn-gear needs, autopilot, the recent trail
//	POST /api/hero/points/stat           {"stat":"str|dex|vit|ene"} one stat point, by verb, read back
//	POST /api/hero/points/skill          {"skill":<id>,"confirm":false} one skill point after the legality check; an unsupported attack asks for confirm
//	POST /api/hero/points/autopilot      {"on":true|false} the per-hero switch (default off: points are banked and you are asked)
//
// The POSTs sit on the answer listener behind policyGate like every control route. Spending is always an explicit action
// of the player: nothing here runs on its own. Each POST is audited (client class and user-agent family, never an address).

const (
	PointsRoute  = "/api/hero/points"
	PointsPrefix = "/api/hero/points/"
)

// PointsRoutes are the viewer routes the handler serves: the sheet, the page and its two assets.
var PointsRoutes = []string{PointsRoute, BuildRoute, "/watch/points", "/watch/points.js", "/watch/buildcards.js", "/watch/points.css"}

// BuildRoute is the automatic allocator's build choice: GET on the viewer listener (the picker and the lock state), POST
// {"build": name, "confirm": bool} on the answer listener (docs/autoalloc.md). Idempotent; refused while locked.
const BuildRoute = "/api/hero/build"

// BuildDesk is the part of the points desk the build routes need; a desk without it answers 501.
type BuildDesk interface {
	BuildView() points.BuildView
	SetBuild(name string, confirm bool) points.Reply
	BuildDetail(id string) (points.BuildDetail, error)
}

// PointsDesk is what the routes need from the points service (points.Desk).
type PointsDesk interface {
	View(ctx context.Context) points.Sheet
	Recent(n int) []points.Event
	SpendStat(ctx context.Context, key string) points.Reply
	SpendSkill(ctx context.Context, id int, confirm bool) points.Reply
	SetAutopilot(on bool) points.Reply
	Record(points.Event)
	Waiting() (text string, stat, skill int, ok bool)
}

// PointsAPI holds the desk once it exists (late-bound like the quest board).
type PointsAPI struct {
	mu sync.Mutex
	d  PointsDesk
}

func NewPointsAPI() *PointsAPI { return &PointsAPI{} }

func (a *PointsAPI) Bind(d PointsDesk) { a.mu.Lock(); a.d = d; a.mu.Unlock() }

func (a *PointsAPI) desk() PointsDesk { a.mu.Lock(); defer a.mu.Unlock(); return a.d }

type pointsView struct {
	points.Sheet
	Recent []points.Event `json:"recent,omitempty"`
}

// Handler serves the viewer GET.
func (a *PointsAPI) Handler() http.Handler {
	return readOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PointsRoute:
		case BuildRoute:
			bd, _ := a.desk().(BuildDesk)
			if bd == nil {
				writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "the build picker is not available yet"})
				return
			}
			if id := r.URL.Query().Get("detail"); id != "" { // one card's detail, on demand (card buildcards)
				det, err := bd.BuildDetail(id)
				if err != nil {
					writeJSON(w, http.StatusNotFound, policyReply{Code: "no_build", Error: err.Error()})
					return
				}
				writeJSON(w, 200, det)
				return
			}
			writeJSON(w, 200, bd.BuildView())
			return
		case "/watch/points":
			pactHdr(w, "text/html; charset=utf-8")
			io.WriteString(w, pointsPage)
			return
		case "/watch/points.js", "/watch/buildcards.js", "/watch/points.css":
			b, err := webFS.ReadFile("web/" + strings.TrimPrefix(r.URL.Path, "/watch/"))
			if err != nil {
				http.NotFound(w, r)
				return
			}
			ct := "text/javascript; charset=utf-8"
			if strings.HasSuffix(r.URL.Path, ".css") {
				ct = "text/css; charset=utf-8"
			}
			pactHdr(w, ct)
			w.Write(b)
			return
		default:
			http.NotFound(w, r)
			return
		}
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "the points desk is not available yet"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		writeJSON(w, 200, pointsView{Sheet: d.View(ctx), Recent: d.Recent(10)})
	}))
}

type pointsBody struct {
	Stat    string `json:"stat"`
	Skill   *int   `json:"skill"`
	Confirm bool   `json:"confirm"`
	On      *bool  `json:"on"`
}

// PostHandler serves the actions under /api/hero/points/.
func (a *PointsAPI) PostHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w0 http.ResponseWriter, r *http.Request) {
		w := &statusWriter{ResponseWriter: w0, code: http.StatusOK}
		ev := points.Event{Path: safePath(r.URL.Path), Remote: remoteClass(r.RemoteAddr), Agent: agentFamily(r.UserAgent())}
		var rep points.Reply
		var d PointsDesk
		ran := false
		defer func() {
			if ev.Action == "" || d == nil || !ran {
				return
			}
			ev.OK, ev.Status, ev.Code, ev.Reason = rep.OK, w.code, rep.Code, rep.Reason
			ev.Before, ev.After = rep.Before, rep.After
			d.Record(ev)
		}()
		action := strings.TrimPrefix(r.URL.Path, PointsPrefix)
		switch action {
		case "stat", "skill", "autopilot":
			ev.Action = action
		default:
			writeJSON(w, http.StatusNotFound, policyReply{Code: "route", Error: "no such points action"})
			return
		}
		var b pointsBody
		if !policyGate(o, w, r, &b) {
			return
		}
		ev.Confirm = b.Confirm
		if dd := a.desk(); dd != nil {
			d = dd
		} else {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "the points desk is not available yet"})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		switch action {
		case "stat":
			ev.Target = safeID(b.Stat)
			if b.Stat == "" {
				writeJSON(w, http.StatusBadRequest, policyReply{Code: "body", Error: `send {"stat": "str"}, "dex", "vit" or "ene"`})
				return
			}
			rep = d.SpendStat(ctx, b.Stat)
		case "skill":
			if b.Skill == nil {
				writeJSON(w, http.StatusBadRequest, policyReply{Code: "body", Error: `send {"skill": <skill id>}`})
				return
			}
			ev.Target = "skill-" + strconv.Itoa(*b.Skill)
			rep = d.SpendSkill(ctx, *b.Skill, b.Confirm)
		default:
			if b.On == nil {
				writeJSON(w, http.StatusBadRequest, policyReply{Code: "body", Error: `send {"on": true} or {"on": false}`})
				return
			}
			ev.Target = map[bool]string{true: "on", false: "off"}[*b.On]
			rep = d.SetAutopilot(*b.On)
		}
		ran = true
		code := rep.Status
		if code == 0 {
			code = http.StatusOK
		}
		writeJSON(w, code, rep)
	})
}

// PointsSection is the return screen's "points waiting" card: shown while stat or skill points are banked (the hero never
// spends them behind the player's back unless the autopilot switch is on). get is read on every build (the desk binds late).
func PointsSection(get func() PointsDesk) ReturnSection {
	return SectionFunc(func(d *ReturnData) *SectionView {
		pd := get()
		if pd == nil {
			return nil
		}
		text, _, _, ok := pd.Waiting()
		if !ok {
			return nil
		}
		return &SectionView{ID: "points", Title: "Points waiting", Lines: []string{text, "The hero is not spending them. Open the Points page to place them: " + PointsPage + "."}}
	})
}

// PointsPage is the path of the page.
const PointsPage = "/watch/points"

type buildBody struct {
	Build   string `json:"build"`
	Confirm bool   `json:"confirm"`
}

// BuildPostHandler serves POST /api/hero/build behind the same gate as the points POSTs. Audited like them.
func (a *PointsAPI) BuildPostHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w0 http.ResponseWriter, r *http.Request) {
		w := &statusWriter{ResponseWriter: w0, code: http.StatusOK}
		var b buildBody
		if !policyGate(o, w, r, &b) {
			return
		}
		d := a.desk()
		bd, _ := d.(BuildDesk)
		if bd == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "the build picker is not available yet"})
			return
		}
		if strings.TrimSpace(b.Build) == "" {
			writeJSON(w, http.StatusBadRequest, policyReply{Code: "body", Error: `send {"build": "<name>"}`})
			return
		}
		rep := bd.SetBuild(b.Build, b.Confirm)
		code := rep.Status
		if code == 0 {
			code = http.StatusOK
		}
		d.Record(points.Event{Path: safePath(r.URL.Path), Remote: remoteClass(r.RemoteAddr), Agent: agentFamily(r.UserAgent()), Action: "build",
			Target: safeID(b.Build), Confirm: b.Confirm, OK: rep.OK, Status: code, Code: rep.Code, Reason: rep.Reason})
		writeJSON(w, code, rep)
	})
}
