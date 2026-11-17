package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"

	"d2lod-bot/internal/goals"
	"d2lod-bot/internal/watch"
	"d2lod-bot/internal/watch/palfeed"
)

// serve runs the watch view on a frozen simulated runner (speed 0, so the clock and every list hold still) plus
// fixture answers for the routes the sim does not have: a game frame, the minimap, gear, tally and currency. Everything is
// deterministic, so two runs of the same tree give the same pictures.
func serve(addr, class string) error {
	sim, err := watch.NewSim(class, 0)
	if err != nil {
		return err
	}
	sim.Hub.SetFrameURL("/stream")
	pal := palFixture() // the lossless feed, offered only to pages that ask with pal=1 (see glanceFixture)
	sim.Hub.SetPalFeed(pal)
	sim.Step(30 * time.Second) // one step so the hero, the map and the feed have something, then the clock stays
	// the act map shows a target (Cold Plains, running), two pins and a failed mark, and the Loop and the queue are open
	if o, err := sim.Hub.GiveOrder(watch.OrderGo, "Cold Plains", 30); err != nil {
		log.Printf("uishots: fixture order: %v", err)
	} else {
		sim.Hub.OrderState(o.ID, watch.OrderActive, "")
	}
	for _, a := range []string{"Stony Field", "Dark Wood"} {
		if _, err := sim.Hub.GiveOrder(watch.OrderGo, a, 20); err != nil {
			log.Printf("uishots: fixture pin: %v", err)
		}
	}
	if o, err := sim.Hub.GiveOrder(watch.OrderGo, "Burial Grounds", 20); err == nil {
		sim.Hub.OrderState(o.ID, watch.OrderFailed, "the hero's guard turned it down")
	}
	sim.Hub.SetMapLadder(func() *watch.MapLadder {
		return &watch.MapLadder{LoopOpen: true, QueueOpen: true, QueueSlots: 2, QueueWaiting: 1, QueueTeaser: "Clear the Den of Evil for a second slot."}
	})
	sim.Hub.Whyline().Observe(goals.Why{Code: "fight", Params: map[string]string{"area": "Blood Moor", "life": "72"}}, "Fighting in Blood Moor", "g1", "")
	sim.Hub.SetGoldSource(func() (int, int, bool) { return 1240, 5300, true })
	sim.Hub.SetHeroStatsSource(func() (watch.HeroStats, bool) {
		def := 214
		return watch.HeroStats{Resists: watch.HeroResists{Fire: 12, Cold: 30, Light: -8, Poison: 5},
			Attrs: &watch.HeroAttrs{Str: 38, Dex: 25, Vit: 30, Energy: 22}, Defense: &def, Damage: &watch.HeroDamage{Min: 4, Max: 9},
			Readiness: &watch.HeroReadiness{Light: "amber", Area: "Blood Moor", Reason: "Lightning resistance is under the line for this area."}}, true
	})
	// what /api/glance shows in the band: two things need the player, like the accepted mockup
	sim.Hub.SetGlanceSources(watch.GlanceSources{Needs: func() []watch.GlanceNeed {
		return []watch.GlanceNeed{
			{ID: "gear", Kind: "gear", Sev: "danger", Text: "Weapon broken", Sub: "repair is about 180 gold", Actions: []watch.GlanceVerb{{Label: "Repair now", Post: "/api/repair"}}},
			{ID: "raven", Kind: "proposal", Sev: "warn", Text: "Go for Blood Raven at level 9?", Actions: []watch.GlanceVerb{{Label: "Go now", Post: "/api/orders"}}}}
	}})
	view := sim.Hub.ViewHandler()
	m := http.NewServeMux()
	m.Handle("/", view)
	// no live event stream: headless Chrome's virtual time waits for the network to go idle, and an open SSE request
	// never does; the pages fall back to their polls
	m.HandleFunc("/api/events", http.NotFound)
	m.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		if ref, err := url.Parse(r.Referer()); err == nil && ref.Query().Get("stream") == "live" {
			liveStream(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(frame())
	})
	m.HandleFunc(watch.CommandsRoute, commandsFixture(class))
	m.HandleFunc(watch.ResetRoute, resetFixture)
	m.HandleFunc(watch.CharactersRoute, charsFixture)
	m.Handle(watch.CharactersRulesRoute, charsRules)
	m.HandleFunc(watch.PauseRoute, pauseFixture)
	m.HandleFunc(watch.SupportRoute, supportFixture)
	m.HandleFunc(watch.SupportBundleRoute, supportFixture)
	m.HandleFunc("/api/lootrules", lootRulesFixture)
	for _, p := range []string{"preview", "undo", "import", "export"} {
		m.HandleFunc("/api/lootrules/"+p, lootRulesFixture)
	}
	m.Handle(watch.ZoneClearRoute, zoneClearFixture())
	m.HandleFunc("/api/update/status", updateStatusFixture)
	m.Handle("/watch/update.js", watch.UpdateHandler(fixtureUpdater{}))
	m.HandleFunc(watch.GlanceRoute, glanceFixture(sim.Hub))
	m.HandleFunc(watch.QuestsRoute, questsFixture)
	m.HandleFunc("/api/repair", func(w http.ResponseWriter, r *http.Request) { js(w, repairFixture) })
	m.HandleFunc("/api/tally", func(w http.ResponseWriter, r *http.Request) { js(w, tallyFixture) })
	if b, err := os.ReadFile("internal/watch/testdata/currency-" + class + ".json"); err == nil {
		m.HandleFunc("/api/currency", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write(b)
		})
	}
	m.HandleFunc("/api/alerts", func(w http.ResponseWriter, r *http.Request) {
		js(w, map[string]any{"schema": "watch.alerts/1", "open": false, "seq": 2, "alerts": []any{
			map[string]any{"seq": 1, "t": 12, "rule": "gear", "kind": "gear", "severity": "warn", "free": true, "text": "Short Staff is broken."},
			map[string]any{"seq": 2, "t": 20, "rule": "life", "kind": "life", "severity": "alarm", "free": true, "text": "Life under 25% in Blood Moor."}},
			"outbound": map[string]any{"enabled": false, "consent": ""}, "notifications": map[string]any{}})
	})
	// the first-steps card the "..." menu opens as a sheet, over the real texts; the log drawer reads the routes above
	if ob, err := watch.NewOnboarding(watch.OnboardingConfig{Hero: func() string { return "Sorc" }}); err == nil {
		m.Handle(watch.OnboardingRoute, ob.Handler())
	} else {
		log.Printf("uishots: onboarding: %v", err)
	}
	// the Items tab of the log drawer: quality colours, a long name and an area, newest first
	drops := sim.Hub.AwayLog()
	for i, d := range []watch.AwayEvent{
		{Item: "Short Staff", Quality: "normal", Area: "Blood Moor"}, {Item: "Leather Gloves of the Fox", Quality: "magic", Area: "Cold Plains"},
		{Item: "Rune of Tir", Quality: "normal", Area: "Cold Plains"}, {Item: "Stormshield, a very long unique shield name for the long names row", Quality: "unique", Area: "Den of Evil"},
		{Item: "Civerb's Ward", Quality: "set", Area: "Burial Grounds"}, {Item: "Gloom Plate", Quality: "rare", Area: "Black Marsh"}} {
		d.Kind, d.T = watch.AwayDrop, float64(100+i*40)
		drops.Add(d)
	}
	m.Handle(watch.DropsRoute, watch.DropsHandler(drops, func() float64 { return 400 }))
	m.Handle(watch.LogRoute, sim.Hub.LogHandler(m))
	mountSkills(m)
	mountVault(m)
	if fogStage != "" {
		pub, err := fogPublisher(fogStage)
		if err != nil {
			return err
		}
		m.Handle("/api/senses", (&watch.Senses{Src: pub, Lock: watch.FixedTier(1)}).Handler())
	} else {
		m.HandleFunc("/api/senses", func(w http.ResponseWriter, r *http.Request) { js(w, sensesFixture()) })
	}
	// the unit layer script of the map lives behind the senses routes; a Senses with no source still serves it
	units := (&watch.Senses{}).Handler()
	m.Handle("/watch/senses_units.js", units)
	mountPages(m, sim, class)
	if fogStage != "" { // the Coverage layer is off by default; the demo needs it on
		inner := http.Handler(m)
		m = http.NewServeMux()
		m.HandleFunc("/watch/panels/map/panel.js", func(w http.ResponseWriter, r *http.Request) {
			rec := httptest.NewRecorder()
			inner.ServeHTTP(rec, r)
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			w.Write(bytes.Replace(rec.Body.Bytes(), []byte("covOn = covStore.get() === '1';"), []byte("covOn = true;"), 1))
		})
		m.Handle("/", inner)
	}
	log.Printf("uishots: fixture viewer on http://%s/watch (%s, frozen)", addr, class)
	return http.ListenAndServe(addr, withUpdateScript(mountFixturePanels(m)))
}

func js(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}

// palFixture serves the lossless feed of the fixture picture: the real hub and handlers, a frame with a bright 6 px border so
// the pictures show whether the canvas fills the card, and a bounded stream (25 s, like liveStream) so headless Chrome's
// network-idle wait ends.
func palFixture() *watch.PalFeed {
	hub := palfeed.NewHub()
	pix := make([]byte, 800*600*4)
	img := &image.RGBA{Pix: pix, Stride: 800 * 4, Rect: image.Rect(0, 0, 800, 600)}
	if src, err := jpeg.Decode(bytes.NewReader(frame())); err != nil {
		log.Printf("uishots: pal fixture: %v", err)
	} else {
		draw.Draw(img, img.Rect, src, image.Point{}, draw.Src)
	}
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			if x < 6 || y < 6 || x >= 794 || y >= 594 {
				img.Set(x, y, color.RGBA{230, 170, 40, 255})
			}
		}
	}
	hub.Publish(palfeed.Frame{W: 800, H: 600, Pix: pix, Seq: 1}, 0)
	// a moving square at 30 fps, so the stream monitor (?pfdebug) has frames to count; and a steady host sample
	go func() {
		seq := uint32(1)
		for range time.Tick(time.Second / 30) {
			seq++
			f := append([]byte(nil), pix...)
			x := 100 + int(seq%300)
			for y := 250; y < 290; y++ {
				for xx := x; xx < x+40; xx++ {
					o := (y*800 + xx) * 4
					f[o], f[o+1], f[o+2] = 240, 240, 240
				}
			}
			hub.Publish(palfeed.Frame{W: 800, H: 600, Pix: f, Seq: seq, GameMS: seq * 33}, 0)
		}
	}()
	o := palfeed.Options{Auth: func(*http.Request) bool { return true }, Server: func() palfeed.ServerStats {
		return palfeed.ServerStats{ThrottlePct: 35, CPUPct: 12, RingFPS: 30, AtMS: time.Now().UnixMilli()}
	}}
	stream := hub.Stream(o)
	return &watch.PalFeed{
		Capability: hub.Capability(o),
		Stream: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
			defer cancel()
			stream.ServeHTTP(w, r.WithContext(ctx))
		}),
	}
}

// frame is a fixed 800x600 picture standing in for the game window: a dark field with a lit diamond.
func frame() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			d := abs(x-400) + abs(y-300)
			c := color.RGBA{20, 16, 12, 255}
			if d < 180 {
				v := uint8(140 - d/2)
				c = color.RGBA{v, v * 3 / 4, v / 3, 255}
			}
			img.Set(x, y, c)
		}
	}
	var b bytes.Buffer
	jpeg.Encode(&b, img, &jpeg.Options{Quality: 80})
	return b.Bytes()
}

// liveStream is a multipart MJPEG stream of the one frame, a part every 300 ms for at most 25 s: the picture tile as it looks with
// a game running (page query stream=live). The reader of the page drops it at the end like any stream.
func liveStream(w http.ResponseWriter, r *http.Request) {
	f := frame()
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=d2frame")
	w.Header().Set("Cache-Control", "no-store")
	fl, _ := w.(http.Flusher)
	end := time.After(25 * time.Second)
	t := time.NewTicker(300 * time.Millisecond)
	defer t.Stop()
	for {
		if _, err := fmt.Fprintf(w, "--d2frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(f)); err != nil {
			return
		}
		w.Write(f)
		w.Write([]byte("\r\n"))
		if fl != nil {
			fl.Flush()
		}
		select {
		case <-r.Context().Done():
			return
		case <-end:
			return
		case <-t.C:
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// sensesFixture is an /api/senses reply: a 48x32 area of rooms and corridors, an exit, a waypoint, the hero and a
// few units (tier 2 open), so the map panel draws something recognisable.
func sensesFixture() map[string]any {
	const w, h = 48, 32
	cells := make([]int, w*h)
	room := func(x0, y0, x1, y1 int) {
		for y := y0; y <= y1; y++ {
			for x := x0; x <= x1; x++ {
				cells[y*w+x] = 1
			}
		}
	}
	room(2, 2, 14, 10)
	room(14, 5, 30, 7)
	room(30, 2, 44, 14)
	room(36, 14, 38, 26)
	room(20, 20, 44, 29)
	room(4, 18, 16, 28)
	room(10, 10, 12, 18)
	for i, c := range cells { // walls around the floor
		if c != 0 {
			continue
		}
		x, y := i%w, i/w
		for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := x+d[0], y+d[1]
			if nx >= 0 && ny >= 0 && nx < w && ny < h && cells[ny*w+nx] == 1 {
				cells[i] = 2
				break
			}
		}
	}
	var rle []int
	for i := 0; i < len(cells); {
		j := i
		for j < len(cells) && cells[j] == cells[i] {
			j++
		}
		rle = append(rle, cells[i], j-i)
		i = j
	}
	return map[string]any{
		"schema": "watch.senses/1", "tiers_unlocked": []int{1, 2}, "version": 1,
		"map": map[string]any{"schema": "watch.senses/1", "areaId": 2, "area": "Blood Moor", "version": 1, "x0": 0, "y0": 0, "step": 1,
			"w": w, "h": h, "rle": rle, "exits": []any{map[string]int{"dest": 3, "x": 43, "y": 8}}, "waypoints": []any{map[string]int{"x": 8, "y": 6}}},
		"hero": map[string]int{"x": 24, "y": 6},
		// a synthetic trail (ages in ms; a 9 s stall at 14,6) and roam coverage, so the pictures show both layers
		"ts": 1000000, "trail_area": 2, "trail": []any{
			map[string]any{"ago": 55000, "x": 4, "y": 5, "e": true}, map[string]any{"ago": 48000, "x": 9, "y": 6},
			map[string]any{"ago": 40000, "x": 14, "y": 6}, map[string]any{"ago": 31000, "x": 19, "y": 6, "a": 1},
			map[string]any{"ago": 20000, "x": 22, "y": 6, "a": 1}, map[string]any{"ago": 8000, "x": 24, "y": 6, "a": 3}},
		"trail_window_s": 1800, "trail_legend": []string{"walk", "combat", "town", "farm", "flee", "rest"},
		"events": []any{
			map[string]any{"ago": 25000, "x": 20, "y": 6, "k": "kill", "t": "Fallen killed", "f": map[string]any{"end": "clear", "kills": 2, "min": 88, "life": 92, "sec": 2.1}},
			map[string]any{"ago": 21000, "x": 22, "y": 6, "k": "kill", "t": "Fallen Shaman killed", "big": true},
			map[string]any{"ago": 12000, "x": 23, "y": 6, "k": "loot", "t": "Picked up Cap"}},
		"coverage": map[string]any{"area": 2, "cell": 8, "cover": 5, "total": 8, "state": "roaming", "legs": 5, "held": 0, "target": []int{33, 6},
			"cells": [][3]int{{6, 6, 1}, {14, 6, 1}, {22, 6, 1}, {24, 12, 2}, {30, 6, 0}, {33, 6, 0}, {40, 8, 0}, {8, 22, 0}}},
		"units": []any{
			map[string]any{"id": 1, "kind": "monster", "x": 34, "y": 6},
			map[string]any{"id": 2, "kind": "monster", "x": 36, "y": 9, "elite": "champion", "name": "Fallen Shaman"},
			map[string]any{"id": 3, "kind": "merc", "x": 22, "y": 6},
			map[string]any{"id": 4, "kind": "npc", "x": 6, "y": 4},
		},
	}
}

// Fixture replies in the shapes of /api/repair (docs/repair.md) and /api/tally (docs/tallystrip.md).
var (
	repairFixture = map[string]any{"mode": "manual", "estimate": 180, "worst": "Short Staff is broken.", "gear": []any{
		map[string]any{"slot": "weapon", "name": "Short Staff", "dur": 0, "max": 18, "pct": 0, "broken": true},
		map[string]any{"slot": "torso", "name": "Quilted Armor", "dur": 5, "max": 24, "pct": 21, "warn": true},
		map[string]any{"slot": "head", "name": "Cap", "dur": 8, "max": 12, "pct": 67},
		map[string]any{"slot": "boots", "name": "Boots", "dur": 10, "max": 12, "pct": 83},
	}}
	tallyFixture = map[string]any{"schema": "watch.tally/1", "enabled": true, "mode": "didle", "strip": "Tally 42 banked", "balance": 42,
		"recent": []any{}, "expedition": map[string]any{"open": true, "credited": 6, "spent": 0, "net": 6, "lines": []any{}},
		"next": []any{map[string]any{"id": "loot_markers", "name": "Loot markers on the map", "cost": 40, "affordable": true, "wired": true}}}
)
