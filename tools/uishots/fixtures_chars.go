package main

import (
	"net/http"
	"net/url"
	"time"

	"d2lod-bot/internal/watch"
)

// charsFixture serves /api/characters for the Characters page (internal/watch/web/characters.js): the page's query chars=
// picks the state (the poll carries it as the Referer). The shapes are the goldens of internal/watch/testdata/characters; the
// rules come from the real handler. Times are relative to now, so "3 h ago" stays true.
//
//	1 one hero, 3 (default) three, 8 eight of every class with a hardcore, a dead and a damaged one, long names, deleted (archives),
//	switching, creating, empty, off, nosave, measured (the backend cannot change anything)
func charsFixture(w http.ResponseWriter, r *http.Request) {
	mode := "3"
	if ref, err := url.Parse(r.Referer()); err == nil && ref.Query().Get("chars") != "" {
		mode = ref.Query().Get("chars")
	}
	switch mode {
	case "off":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"code":"off","error":"characters are not available"}`))
		return
	case "nosave":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"code":"no_save_dir","error":"the save folder is not known yet (the first save creates it)"}`))
		return
	}
	ago := func(d time.Duration) string { return time.Now().Add(-d).UTC().Format(time.RFC3339) }
	hero := func(name, class, className string, level int, last time.Duration, extra map[string]any) map[string]any {
		m := map[string]any{"name": name, "class": class, "className": className, "level": level, "hardcore": false, "expansion": true, "lastPlayed": ago(last), "active": false, "archived": []any{}}
		for k, v := range extra {
			m[k] = v
		}
		return m
	}
	arch := func(name, class string, level int, d time.Duration) map[string]any {
		return map[string]any{"id": name + "@20261114T120000Z", "name": name, "class": class, "level": level, "archivedAt": ago(d)}
	}
	chars := []any{
		hero("Holy", "paladin", "Paladin", 24, 5*time.Minute, map[string]any{"active": true}),
		hero("Zapp", "sorceress", "Sorceress", 9, 3*time.Hour, nil),
		hero("Arrow", "amazon", "Amazon", 31, 26*time.Hour, map[string]any{"expansion": false}),
	}
	deleted := []any{}
	caps := map[string]any{"create": true, "switch": true, "delete": true, "restore": true}
	v := map[string]any{"schema": "watch.characters/1", "active": "Holy", "caps": caps}
	switch mode {
	case "1":
		chars = chars[:1]
	case "8":
		chars = append(chars,
			hero("Bone", "necromancer", "Necromancer", 18, 3*24*time.Hour, nil),
			hero("Rage", "barbarian", "Barbarian", 40, 9*24*time.Hour, map[string]any{"hardcore": true}),
			hero("Wolf", "druid", "Druid", 12, 40*24*time.Hour, map[string]any{"hardcore": true, "died": true}),
			hero("Shade", "assassin", "Assassin", 27, 2*time.Hour, map[string]any{"damaged": "the save fails its check (checksum mismatch)"}),
			hero("Newb", "sorceress", "Sorceress", 1, 90*time.Second, map[string]any{"archived": []any{arch("Newb", "sorceress", 5, 48*time.Hour)}}))
		deleted = []any{arch("Mage", "sorceress", 33, 4*time.Hour), arch("Tank", "paladin", 20, 72*time.Hour)}
	case "long":
		chars = []any{
			hero("Mmmmmmmmmmmmmmm", "paladin", "Paladin", 24, 5*time.Minute, map[string]any{"active": true}),
			hero("Wolfgang-Amadeus", "sorceress", "Sorceress", 99, 3*time.Hour, map[string]any{"damaged": "the save fails its check: the file ends early (section sizes)"}),
			hero("Zzzzzzzzzzzzzzz", "necromancer", "Necromancer", 100, 26*time.Hour, map[string]any{"hardcore": true, "died": true}),
		}
		deleted = []any{arch("Iiiiiiiiiiiiiii", "druid", 12, 5*time.Hour)}
	case "deleted":
		deleted = []any{arch("Mage", "sorceress", 33, 4*time.Hour), arch("Tank", "paladin", 20, 72*time.Hour), arch("Zeta", "assassin", 7, 40*24*time.Hour)}
	case "switching":
		v["switch"] = map[string]any{"from": "Holy", "to": "Zapp", "since": ago(5 * time.Second)}
	case "creating":
		v["switch"] = map[string]any{"from": "Holy", "to": "Newb", "creating": true, "since": ago(5 * time.Second)}
	case "empty":
		chars, v["active"] = []any{}, ""
	case "measured":
		caps["create"], caps["switch"] = false, false
		caps["delete"], caps["restore"] = false, false
		caps["why"] = "a measured run is in progress (D2BOT_RUN)"
	}
	v["characters"], v["deleted"] = chars, deleted
	js(w, v)
}

// charsRules is the real rules handler (the class table and the name rule of the backend).
var charsRules = func() http.Handler { a := watch.NewCharactersAPI(); return a.RulesHandler() }()
