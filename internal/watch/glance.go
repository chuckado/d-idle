package watch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"
)

// GET /api/glance (watch.glance/1): the cockpit's band and strip in one read-only reply (docs/ui/redesign-2026-11-14,
// J-second-attempt.md section 2): the status sentence, what needs the player, five health lights and the progress
// numbers. It is built from what the Hub already holds (hero, uptime, orders, whyline); nothing is polled for it. A
// source that is not wired answers light "unknown" or its block is left out. The fields do not change after the first
// release: the panels and the golden files build on them.
//
// Survival margin (user answer 2026-11-14): shown as words only, never a number. Until the survival controller feeds
// /api/glance (docs/survival/design-2026-11-14.md section 8.1, slice S-C) the word comes from the readiness light of
// /api/hero's stats block; sentence.margin and health.readiness.word carry it ("safe", "risky", "too hard").

const (
	GlanceRoute  = "/api/glance"
	glanceSchema = "watch.glance/1"

	glancePoints  = 30              // one-minute samples kept for the xp/h and gold/h series
	glanceEvery   = time.Minute     // spacing of two samples
	glanceMaxNeed = 20              // rows of "needs you" kept
	glanceTextMax = 120             // bound of a reply text
	glanceWarm    = 2 * time.Minute // a rate over less time is noise (same as xpRateWarm)
)

// GlanceView is the /api/glance reply.
type GlanceView struct {
	Schema   string         `json:"schema"`
	Seq      uint64         `json:"seq"`
	Sentence GlanceSentence `json:"sentence"`
	Needs    []GlanceNeed   `json:"needs"`
	Health   GlanceHealth   `json:"health"`
	Progress GlanceProgress `json:"progress"`
	Journey  *GlanceJourney `json:"journey,omitempty"`
	// Survival is the block of GET /api/survival (docs/survival/ledger.md, slice S-C) as the bot gave it, words only; absent
	// while no source is bound (an old bot, S-C not merged).
	Survival json.RawMessage `json:"survival,omitempty"`
	// Paused: the game itself is frozen by the player's Pause (docs/pause.md); PausedSince is the RFC 3339 time it began.
	Paused      bool   `json:"paused,omitempty"`
	PausedSince string `json:"pausedSince,omitempty"`
	// Stream is the server side of the game picture (rates since the last read); absent without -watch-video. A friend can
	// screenshot it next to the page's own fps line.
	Stream *StreamStats `json:"stream,omitempty"`
	// Intake is the observed damage intake (docs/survival/ledger.md "Observed intake"); absent while no source is bound or
	// before the first minute line. Additive: the cockpit turns it into light, moderate or heavy later.
	Intake *GlanceIntake `json:"intake,omitempty"`
}

// GlanceIntake is the life lost per minute averaged over Window minute lines.
type GlanceIntake struct {
	LifeLostPerMin float64 `json:"lifeLostPerMin"`
	Window         int     `json:"window"`
}

// GlanceSentence is the line of the band: a state chip (Word), the Place, and the Reason, with the popover's content.
// Tone is ok, warn, danger or info.
type GlanceSentence struct {
	Tone   string  `json:"tone"`
	Word   string  `json:"word"`
	Place  string  `json:"place"`
	Reason string  `json:"reason"`
	Why    HeroWhy `json:"why"`
	Uptime string  `json:"uptime"`
	Voice  string  `json:"voice"`
	Margin string  `json:"margin,omitempty"` // safe, risky or too hard; absent while the bot reports no readiness
}

// GlanceNeed is one row of "needs you". Sev is info, warn or danger; an action posts Body to the answer route Post.
type GlanceNeed struct {
	ID      string       `json:"id"`
	Kind    string       `json:"kind"`
	Sev     string       `json:"sev"`
	Text    string       `json:"text"`
	Sub     string       `json:"sub,omitempty"`
	Actions []GlanceVerb `json:"actions"`
}

// GlanceVerb is a button on a need.
type GlanceVerb struct {
	Label string          `json:"label"`
	Post  string          `json:"post"`
	Body  json.RawMessage `json:"body,omitempty"`
}

// GlanceLight is one health light: Light is ok, warn, bad or unknown, Word the one word under it, Detail the sentence for
// its tooltip.
type GlanceLight struct {
	Light  string `json:"light"`
	Word   string `json:"word"`
	Detail string `json:"detail,omitempty"`
}

// GlanceHealth is the five lights.
type GlanceHealth struct {
	Life      GlanceLight `json:"life"`
	Gear      GlanceLight `json:"gear"`
	Readiness GlanceLight `json:"readiness"`
	Merc      GlanceLight `json:"merc"`
	Supplies  GlanceLight `json:"supplies"`
}

// GlanceProgress is the level meter and the rates. Rates are null while they build up; the series are per hour, one value
// per minute, oldest first.
type GlanceProgress struct {
	Level       GlanceLevel `json:"level"`
	XPPerHour   *float64    `json:"xpPerHour"`
	XPSeries    []float64   `json:"xpSeries"`
	GoldPerHour *float64    `json:"goldPerHour"`
	GoldSeries  []float64   `json:"goldSeries"`
}

// GlanceLevel: From is the level now, To the next one, Pct the share of this level gone, ETASeconds the time to To at the
// present rate (null while the rate builds up).
type GlanceLevel struct {
	From       int      `json:"from"`
	To         int      `json:"to"`
	Pct        *float64 `json:"pct"`
	ETASeconds *int     `json:"etaSeconds"`
}

// GlanceJourney is the chapter strip (from the campaign desk, when a source is bound).
type GlanceJourney struct {
	N        int             `json:"n"`
	Of       int             `json:"of"`
	Chapters []GlanceChapter `json:"chapters"`
	Next     *GlanceNextStep `json:"next,omitempty"`
	Actions  []GlanceVerb    `json:"actions"`
}

// GlanceChapter is one dot of the strip; State is done, open, held, skipped or todo.
type GlanceChapter struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

// GlanceNextStep is the chapter the hero heads for and how ready it is.
type GlanceNextStep struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Band  string `json:"band,omitempty"`
	Text  string `json:"text,omitempty"`
}

// GlanceSources are the optional readers behind the blocks the Hub does not hold itself. Every field may be nil.
type GlanceSources struct {
	Needs    func() []GlanceNeed
	Journey  func() (GlanceJourney, bool)
	Gear     func() (GlanceLight, bool)
	Merc     func() (GlanceLight, bool)
	Supplies func() (GlanceLight, bool)
	Intake   func() (perMin float64, window int, ok bool) // observed life lost per minute (bot.Runner.Intake)
	Survival func() json.RawMessage                       // the `survival` object of /api/survival, nil result: none yet (the feature flag is the binding)
	Voice    func() string                                // the hero's latest voice line
	Paused   func() (paused bool, text string)            // a pause the player asked for (card B1)
	// Frozen: the game is stopped (the frozen Pause) and since when.
	Frozen func() (since time.Time, on bool)
	// Resetting is a Reset the player pressed that has not finished (docs/reset.md); the band says Resetting with the text.
	Resetting func() (on bool, text string)
}

// glanceState is the Hub's part: the sources and the ring of samples.
type glanceState struct {
	mu  sync.Mutex
	src GlanceSources
	pts []glancePt
}

type glancePt struct {
	t      time.Duration
	xp     int64
	gold   int
	goldOK bool
}

// SetGlanceSources binds the optional readers (nil fields stay unbound).
func (h *Hub) SetGlanceSources(s GlanceSources) { h.gl.mu.Lock(); h.gl.src = s; h.gl.mu.Unlock() }

// sample records the point of this minute. The ring fills while somebody polls /api/glance (the band does every two
// seconds); with nobody watching there is nothing to draw.
func (g *glanceState) sample(now time.Duration, hv HeroView) {
	if hv.XP <= 0 {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if n := len(g.pts); n > 0 {
		last := g.pts[n-1]
		if now < last.t || hv.XP < last.xp { // the clock or the character was reset
			g.pts = g.pts[:0]
		} else if now-last.t < glanceEvery {
			return
		}
	}
	p := glancePt{t: now, xp: hv.XP}
	if hv.Gold != nil {
		p.gold, p.goldOK = *hv.Gold, true
	}
	g.pts = append(g.pts, p)
	if len(g.pts) > glancePoints {
		g.pts = g.pts[len(g.pts)-glancePoints:]
	}
}

// rates are the per-hour values between neighbouring points, oldest first; gold only where both ends know it.
func (g *glanceState) rates() (xp, gold []float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	xp, gold = []float64{}, []float64{}
	for i := 1; i < len(g.pts); i++ {
		a, b := g.pts[i-1], g.pts[i]
		hrs := (b.t - a.t).Hours()
		if hrs <= 0 {
			continue
		}
		xp = append(xp, float64(int(float64(b.xp-a.xp)/hrs)))
		if a.goldOK && b.goldOK {
			gold = append(gold, float64(int(float64(b.gold-a.gold)/hrs)))
		}
	}
	return xp, gold
}

// Glance builds the /api/glance reply.
func (h *Hub) Glance() GlanceView {
	hv := h.Hero()
	up := h.Uptime()
	now := h.cfg.Clock.Now()
	h.gl.sample(now, hv)
	h.gl.mu.Lock()
	src := h.gl.src
	h.gl.mu.Unlock()

	v := GlanceView{Schema: glanceSchema, Seq: up.Seq, Needs: []GlanceNeed{}}
	v.Sentence = glanceSentence(hv, up, src)
	if src.Survival != nil {
		v.Survival = h.glanceSurvival(&v.Sentence, hv, src.Survival())
	}
	if src.Intake != nil {
		if pm, w, ok := src.Intake(); ok {
			v.Intake = &GlanceIntake{LifeLostPerMin: float64(int(pm*10+0.5)) / 10, Window: w}
		}
	}
	for _, n := range h.questionNeeds() { // a question that waits for the player comes first: it fades
		if len(v.Needs) < glanceMaxNeed {
			v.Needs = append(v.Needs, cleanNeed(n))
		}
	}
	if src.Needs != nil {
		for _, n := range src.Needs() {
			if len(v.Needs) < glanceMaxNeed {
				v.Needs = append(v.Needs, cleanNeed(n))
			}
		}
	}
	if src.Frozen != nil {
		if since, on := src.Frozen(); on {
			v.Paused, v.PausedSince = true, since.UTC().Format(time.RFC3339)
		}
	}
	if h.video != nil || h.palFeedURL() != "" {
		var st StreamStats
		if h.video != nil {
			st = h.video.Stats()
		}
		st.Formats = h.streamFormats()
		v.Stream = &st
	}
	v.Health = glanceHealth(hv, src)
	v.Progress = h.glanceProgress(hv)
	if src.Journey != nil {
		if j, ok := src.Journey(); ok {
			j = cleanJourney(j)
			v.Journey = &j
		}
	}
	return v
}

// glanceSentence is the band's line. The reason is the player's order when one runs ("you ordered it, 12 min left"), else the
// whyline's sentence, else what the hero reports it does; it is never empty.
func glanceSentence(hv HeroView, up UptimeView, src GlanceSources) GlanceSentence {
	s := GlanceSentence{Place: hv.Area, Why: hv.Why}
	switch up.State {
	case KindDeath:
		s.Tone, s.Word = "danger", "Dead"
	case KindRecovery:
		s.Tone, s.Word = "warn", "Recovering"
	case KindStall:
		s.Tone, s.Word = "warn", "Waiting"
	case KindProgress:
		s.Tone, s.Word = "ok", firstWord(hv.Activity, "Active")
	default:
		s.Tone, s.Word = "info", "Idle"
	}
	reason := strings.TrimRight(strings.TrimSpace(hv.Why.Now), ".")
	if hv.Goal != nil {
		reason = "you ordered it"
		if hv.Goal.LeftMin != nil {
			reason += fmt.Sprintf(", %d min left", *hv.Goal.LeftMin)
		}
	}
	if reason == "" {
		reason = strings.TrimRight(hv.Activity, ".")
	}
	if reason == "" {
		reason = up.Label
	}
	if !hv.Known {
		s.Tone, s.Word, s.Place = "info", "Waiting", ""
		if reason == "" || up.State == KindIdle {
			reason = "for the bot's first report"
		}
	}
	if src.Paused != nil {
		if p, text := src.Paused(); p {
			s.Tone, s.Word = "info", "Paused"
			if t := safeText(text, glanceTextMax); t != "" {
				reason = t
			}
		}
	}
	if src.Resetting != nil {
		if on, text := src.Resetting(); on {
			s.Tone, s.Word = "info", "Resetting"
			if t := safeText(text, glanceTextMax); t != "" {
				reason = t
			}
		}
	}
	s.Reason = safeText(reason, glanceTextMax)
	if hv.Stats != nil && hv.Stats.Readiness != nil {
		s.Margin = marginWord(hv.Stats.Readiness.Light)
	}
	if s.Margin == "too hard" && s.Tone == "ok" {
		s.Tone = "warn"
	}
	if up.UptimePct != nil {
		s.Uptime = fmt.Sprintf("%d%% of the time making progress", int(*up.UptimePct+0.5))
	}
	if src.Voice != nil {
		s.Voice = safeText(src.Voice(), glanceTextMax)
	}
	return s
}

func firstWord(s, def string) string {
	f := strings.Fields(s)
	if len(f) == 0 {
		return def
	}
	r := []rune(strings.ToLower(f[0]))
	r[0] = unicode.ToUpper(r[0])
	return safeText(string(r), 16)
}

// marginWord maps the readiness light to the words the player sees. The number behind it is never shown.
func marginWord(light string) string {
	switch light {
	case "green":
		return "safe"
	case "amber":
		return "risky"
	case "red":
		return "too hard"
	}
	return ""
}

func glanceHealth(hv HeroView, src GlanceSources) GlanceHealth {
	unk := GlanceLight{Light: "unknown", Word: "unknown"}
	h := GlanceHealth{Life: unk, Gear: unk, Readiness: unk, Merc: unk, Supplies: unk}
	if hv.MaxLife > 0 {
		p := 100 * hv.Life / hv.MaxLife
		l := GlanceLight{Light: "ok", Word: fmt.Sprintf("%d/%d", hv.Life, hv.MaxLife)}
		switch {
		case p < 25:
			l.Light = "bad"
		case p < 50:
			l.Light = "warn"
		}
		h.Life = l
	}
	if hv.Stats != nil && hv.Stats.Readiness != nil {
		r := hv.Stats.Readiness
		l := GlanceLight{Word: marginWord(r.Light), Detail: r.Reason}
		switch r.Light {
		case "green":
			l.Light = "ok"
		case "amber":
			l.Light = "warn"
		case "red":
			l.Light = "bad"
		}
		if l.Word != "" {
			h.Readiness = l
		}
	}
	read := func(f func() (GlanceLight, bool), to *GlanceLight) {
		if f == nil {
			return
		}
		if l, ok := f(); ok {
			l.Word, l.Detail = safeText(l.Word, 20), safeText(l.Detail, glanceTextMax)
			switch l.Light {
			case "ok", "warn", "bad":
				*to = l
			}
		}
	}
	read(src.Gear, &h.Gear)
	read(src.Merc, &h.Merc)
	read(src.Supplies, &h.Supplies)
	return h
}

func (h *Hub) glanceProgress(hv HeroView) GlanceProgress {
	xs, gs := h.gl.rates()
	p := GlanceProgress{XPPerHour: hv.XPPerHour, XPSeries: xs, GoldSeries: gs,
		Level: GlanceLevel{From: hv.Level, To: hv.Level + 1, Pct: hv.XPPct}}
	if hv.Level == 0 {
		p.Level.To = 0
	}
	if hv.XPPerHour != nil && *hv.XPPerHour > 0 && hv.XPNext > hv.XP {
		eta := int(float64(hv.XPNext-hv.XP) / *hv.XPPerHour * 3600)
		p.Level.ETASeconds = &eta
	}
	// gold per hour from the same clock: the newest rate once two samples exist
	if n := len(gs); n > 0 {
		g := gs[n-1]
		p.GoldPerHour = &g
	}
	return p
}

func cleanNeed(n GlanceNeed) GlanceNeed {
	n.ID, n.Kind = safeText(n.ID, 40), safeText(n.Kind, 40)
	n.Text, n.Sub = safeText(n.Text, glanceTextMax), safeText(n.Sub, glanceTextMax)
	switch n.Sev {
	case "info", "warn", "danger":
	default:
		n.Sev = "info"
	}
	if n.Kind == "question" {
		n.Actions = cleanVerbsMax(n.Actions, 8) // every answer of a hero question
	} else {
		n.Actions = cleanVerbs(n.Actions)
	}
	return n
}

// cleanVerbs keeps at most three buttons and only those that post to an API route of this server.
func cleanVerbs(in []GlanceVerb) []GlanceVerb { return cleanVerbsMax(in, 3) }

func cleanVerbsMax(in []GlanceVerb, max int) []GlanceVerb {
	out := []GlanceVerb{}
	for _, a := range in {
		if len(out) == max || !strings.HasPrefix(a.Post, "/api/") || strings.Contains(a.Post, "//") {
			continue
		}
		a.Label = safeText(a.Label, 24)
		if a.Label != "" {
			out = append(out, a)
		}
	}
	return out
}

func cleanJourney(j GlanceJourney) GlanceJourney {
	if len(j.Chapters) > 12 {
		j.Chapters = j.Chapters[:12]
	}
	for i, c := range j.Chapters {
		j.Chapters[i] = GlanceChapter{ID: safeText(c.ID, 40), State: safeText(c.State, 12)}
	}
	if j.Chapters == nil {
		j.Chapters = []GlanceChapter{}
	}
	if j.Next != nil {
		n := *j.Next
		n.ID, n.Title, n.Band, n.Text = safeText(n.ID, 40), safeText(n.Title, 60), safeText(n.Band, 20), safeText(n.Text, glanceTextMax)
		j.Next = &n
	}
	j.Actions = cleanVerbs(j.Actions)
	return j
}

func (h *Hub) glanceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, h.Glance()) }
}

// glanceSurvival embeds the survival block and lets it speak in the sentence: its margin word replaces the readiness
// one and its zone reason fills an empty reason. A block that does not parse, or whose margin is not one of the three
// words, is dropped whole: a number never reaches the page.
func (h *Hub) glanceSurvival(s *GlanceSentence, hv HeroView, raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	var b struct {
		Zone struct {
			Margin string `json:"margin"`
			Why    struct {
				Text string `json:"text"`
			} `json:"why"`
		} `json:"zone"`
	}
	if json.Unmarshal(raw, &b) != nil {
		return nil
	}
	switch b.Zone.Margin {
	case "safe", "risky", "too hard":
		s.Margin = b.Zone.Margin
	case "":
	default:
		return nil
	}
	if hv.Goal == nil && b.Zone.Why.Text != "" && hv.Why.Now == "" {
		s.Reason = safeText(b.Zone.Why.Text, glanceTextMax)
	}
	return raw
}
