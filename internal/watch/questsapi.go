package watch

import (
	"net/http"
	"strings"
	"sync"

	"d2lod-bot/internal/questboard"
)

// The quest board routes (docs/questboard.md "API"; slice S4 of docs/design/quest-board-2026-11-16.md):
//
//	GET  /api/quests                  the board: offers with readiness words, the running chunk, hand-in, mode, history
//	POST /api/quests/<id>/start       accept the CURRENT chunk of a quest (body {} or {"chunk":"<id>"})
//	POST /api/quests/<id>/handin      the hand-in tap (202: the dialogue runs on the bot loop; read the result from GET)
//	POST /api/quests/<id>/cancel      stop an accepted chunk
//	POST /api/quests/<id>/retry       clear the park of a blocked chunk
//	POST /api/quests/<id>/auto        {"on": true|false} the opt-in to start this quest by itself when it is safe (slice S8)
//	POST /api/quests/away/dismiss     clear the "while you were away" card
//	POST /api/quests/mode             {"mode":"train|idle|farm"} (stored; the modes act from slice S6)
//
// The POSTs sit on the answer listener behind the same gate as every control route (policyGate: origin list, token,
// JSON only, small body). Every reply carries `ok`; a refusal carries `code` and `reason` in words, and a repeated press
// is `ok` with `already`. The page reads state back from the GET, never from a reply.

const (
	QuestsRoute     = "/api/quests"
	QuestsPrefix    = "/api/quests/"
	QuestsModeRoute = "/api/quests/mode"
)

// QuestsDesk is what the routes need from the board (questboard.Desk).
type QuestsDesk interface {
	View() questboard.View
	Start(quest, chunk string) questboard.Reply
	HandIn(quest, chunk string) questboard.Reply
	Cancel(quest, chunk string) questboard.Reply
	Retry(quest, chunk string) questboard.Reply
	SetMode(mode string) questboard.Reply
	SetAuto(quest string, on bool) questboard.Reply
	DismissAway() questboard.Reply
	Record(questboard.AuditEvent)
}

// QuestsAPI holds the desk once it exists (it is late-bound like the other desks).
type QuestsAPI struct {
	mu sync.Mutex
	d  QuestsDesk
}

func NewQuestsAPI() *QuestsAPI { return &QuestsAPI{} }

func (a *QuestsAPI) Bind(d QuestsDesk) { a.mu.Lock(); a.d = d; a.mu.Unlock() }

func (a *QuestsAPI) desk() QuestsDesk { a.mu.Lock(); defer a.mu.Unlock(); return a.d }

// Handler serves the viewer GET.
func (a *QuestsAPI) Handler() http.Handler {
	return readOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != QuestsRoute {
			http.NotFound(w, r)
			return
		}
		d := a.desk()
		if d == nil {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "the quest board is not available yet"})
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, d.View())
	}))
}

type questBody struct {
	Chunk string `json:"chunk"`
	Mode  string `json:"mode"`
	On    *bool  `json:"on"`
	// Confirm is the second tap of a Stop; only recorded, a missing flag does not refuse.
	Confirm bool `json:"confirm"`
}

// PostHandler serves the actions under /api/quests/.
func (a *QuestsAPI) PostHandler(o AnswerOptions) http.Handler {
	return http.HandlerFunc(func(w0 http.ResponseWriter, r *http.Request) {
		w := &statusWriter{ResponseWriter: w0, code: http.StatusOK}
		ev := questboard.AuditEvent{Path: safePath(r.URL.Path), Remote: remoteClass(r.RemoteAddr), Agent: agentFamily(r.UserAgent())}
		var rep questboard.Reply
		var d QuestsDesk
		ran := false
		defer func() {
			ev.Status, ev.OK, ev.Already, ev.Code, ev.Reason = w.code, rep.OK, rep.Already, rep.Code, rep.Reason
			if ev.Action == "" {
				return
			}
			logQuestAction(ev)
			if d != nil && ran {
				d.Record(ev)
			}
		}()
		rest := strings.TrimPrefix(r.URL.Path, QuestsPrefix)
		parts := strings.Split(rest, "/")
		isMode := rest == "mode"
		isAway := rest == "away/dismiss"
		switch {
		case isMode:
			ev.Action = "mode"
		case isAway:
			ev.Action = "dismiss"
		case len(parts) == 2:
			ev.Action, ev.Quest = safeID(parts[1]), safeID(parts[0])
		}
		if !isMode && !isAway && (len(parts) != 2 || parts[0] == "" || parts[1] == "") {
			writeJSON(w, http.StatusNotFound, policyReply{Code: "route", Error: "no such quest route"})
			return
		}
		if !isMode && !isAway {
			switch parts[1] {
			case "start", "handin", "cancel", "retry", "auto":
			default:
				writeJSON(w, http.StatusNotFound, policyReply{Code: "route", Error: "no such quest action"})
				return
			}
		}
		var b questBody
		if !policyGate(o, w, r, &b) {
			return
		}
		ev.Chunk, ev.Confirm = safeID(b.Chunk), b.Confirm
		if isMode {
			ev.Chunk = safeID(b.Mode)
		}
		if dd := a.desk(); dd != nil {
			d = dd
		} else {
			writeJSON(w, http.StatusServiceUnavailable, policyReply{Code: "off", Error: "the quest board is not available yet"})
			return
		}
		switch {
		case isMode:
			if b.Mode == "" {
				writeJSON(w, http.StatusBadRequest, policyReply{Code: "body", Error: `send {"mode": "train"}, "idle" or "farm"`})
				return
			}
			rep = d.SetMode(b.Mode)
		case isAway:
			rep = d.DismissAway()
		case parts[1] == "auto":
			if b.On == nil {
				writeJSON(w, http.StatusBadRequest, policyReply{Code: "body", Error: `send {"on": true} or {"on": false}`})
				return
			}
			rep = d.SetAuto(parts[0], *b.On)
		case parts[1] == "start":
			rep = d.Start(parts[0], b.Chunk)
		case parts[1] == "handin":
			rep = d.HandIn(parts[0], b.Chunk)
		case parts[1] == "cancel":
			rep = d.Cancel(parts[0], b.Chunk)
		default:
			rep = d.Retry(parts[0], b.Chunk)
		}
		ran = true
		code := rep.Status
		if code == 0 {
			code = http.StatusOK
		}
		writeJSON(w, code, rep)
	})
}

// GlanceQuest is the quest strip of the video card: the current chunk, its step and the NPC (additive, absent
// while the board is off or nothing is accepted).
type GlanceQuest struct {
	Quest string `json:"quest"`
	Title string `json:"title"`
	Chunk string `json:"chunk"`
	Part  string `json:"part,omitempty"`
	State string `json:"state"`
	Step  string `json:"step,omitempty"`
	NPC   string `json:"npc,omitempty"`
}
