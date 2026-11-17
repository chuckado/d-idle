package main

import (
	"bytes"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/url"
	"strings"

	"d2lod-bot/internal/watch"
)

// The phone's storage tab with data that looks like a real hero's (card uiphone, bug report 2026-11-15, a 412 px screen): a full
// pack with 2 by 3 and 2 by 4 items, a stash with items, a seven-potion belt whose items carry their slot number as x, a long item
// name, the selected item with its action buttons, the town errands, and game art on every cell (the first fit fixtures had no
// art, so the label painted over the art never showed). fx=phone turns all of it on; the art is a fixture picture, not a game one.

// phoneArt is the size in cells of the art of an item code, so a tall item gets a tall picture.
var phoneArt = map[string][2]int{"8ss": {1, 3}, "rin": {1, 1}, "tsc": {1, 1}, "isc": {1, 1}, "tbk": {1, 2}, "hla": {2, 3}, "gth": {2, 4}, "tar": {2, 2}, "buc": {2, 2}, "pik": {2, 4},
	"jew": {1, 1}, "cm1": {1, 1}, "cm2": {1, 2}, "ltp": {2, 3}, "amu": {1, 1}, "hp1": {1, 1}, "mp1": {1, 1}, "hp2": {1, 1}, "sst": {1, 3}, "cap": {2, 2}, "qui": {2, 3}, "lgl": {2, 2},
	"lbl": {2, 1}, "lbt": {2, 2}, "gsw": {1, 1}}

// fxKind is the kind word the server sends from the tables (itemMeta GoalType).
var fxKind = map[string]string{"jew": "jewel", "rin": "ring", "amu": "amulet", "cm1": "charm", "cm2": "charm", "gsw": "gem", "hp1": "potion", "mp1": "potion", "hp2": "potion", "zzz": "jewel"}

func init() {
	fxVariants["phone"] = func(v map[string]any) {
		pack := []any{
			item(10, "8ss", "Gothic Staff of Charged Bolt and the Burning Stars", "magic", 0, 0, 1, 3, map[string]any{"canSell": true, "canWear": []string{"rhand"}, "noWearSlot": noSlot("rhand"), "stashOp": "stash", "level": 12,
				"props": []string{"Two-Hand Damage: 2 to 8", "+2 to Charged Bolt (Sorceress Only)", "+12 to Mana", "Requires Level 12"}}),
			item(11, "rin", "Ring", "rare", 1, 0, 1, 1, map[string]any{"props": []string{"Fire Resist +11%", "+12 Life"}, "canWear": []string{"lring", "rring"}, "noWearSlot": noSlot("lring", "rring")}),
			item(13, "tsc", "Scroll of Town Portal", "normal", 1, 1, 1, 1, nil),
			item(14, "isc", "Scroll of Identify", "normal", 1, 2, 1, 1, nil),
			item(21, "tsc", "Scroll of Town Portal", "normal", 1, 3, 1, 1, nil),
			item(41, "hla", "Hard Leather Armor", "normal", 2, 0, 2, 3, map[string]any{"identified": false, "mark": "keep"}),
			item(42, "gth", "Gothic Plate of the Everlasting Dawnbringer", "rare", 4, 0, 2, 4, map[string]any{"req": map[string]any{"str": 80, "lvl": 22}, "usable": usable([]any{need("lvl", 22, 7, "level 22"), need("str", 80, 38, "80 strength")}, nil), "maxDur": 55, "dur": 55}),
			item(18, "tar", "Targe", "magic", 6, 0, 2, 2, map[string]any{"req": map[string]any{"str": 16, "class": []string{"Paladin"}}, "usable": usable([]any{why("class", "only a Paladin can use it")}, nil)}),
			item(43, "buc", "Buckler", "normal", 6, 2, 2, 2, map[string]any{"mark": "sell", "canSell": true}),
			item(19, "pik", "Pike", "normal", 8, 0, 2, 4, map[string]any{"req": map[string]any{"lvl": 9, "str": 60, "dex": 45}, "usable": usable([]any{need("lvl", 9, 7, "level 9"), need("str", 60, 38, "60 strength")}, nil)}),
			item(44, "jew", "Jewel", "rare", 2, 3, 1, 1, nil),
			item(45, "cm1", "Small Charm of Vita", "magic", 3, 3, 1, 1, nil),
		}
		v["backpack"] = pack
		v["belt"] = []any{}
		for i, c := range []string{"hp1", "hp1", "mp1", "mp1", "hp2", "hp2", "mp1"} {
			n := "Minor Healing Potion"
			if c == "mp1" {
				n = "Minor Mana Potion"
			} else if c == "hp2" {
				n = "Light Healing Potion"
			}
			v["belt"] = append(v["belt"].([]any), item(20+i+30, c, n, "normal", i, 0, 1, 1, nil)) // the slot number as x, y 0
		}
		v["stash"] = []any{
			item(30, "rin", "Ring", "magic", 0, 0, 1, 1, nil),
			item(46, "amu", "Amulet of the Everlasting Dawnbringer", "rare", 1, 0, 1, 1, nil),
			item(31, "ltp", "Light Plate", "magic", 2, 0, 2, 3, map[string]any{"stashOp": "take", "props": []string{"Defense: 60", "Cold Resist +11%"}}),
			item(47, "jew", "Jewel", "magic", 4, 0, 1, 1, nil),
			item(48, "cm2", "Large Charm of the Whale", "magic", 5, 0, 1, 2, nil),
			item(49, "rin", "Ring of the Fox", "magic", 0, 1, 1, 1, nil),
			item(50, "cm1", "Small Charm", "normal", 1, 1, 1, 1, nil),
			item(32, "tsc", "Scroll of Town Portal", "normal", 4, 1, 1, 1, nil),
			item(51, "pik", "Pike", "magic", 0, 3, 2, 4, nil),
			item(52, "hla", "Hard Leather Armor", "normal", 2, 3, 2, 3, nil),
			item(53, "tar", "Targe", "normal", 4, 3, 2, 2, nil),
			item(54, "jew", "Jewel", "magic", 0, 2, 1, 1, map[string]any{"identified": false}),
			item(55, "jew", "Jewel", "rare", 1, 2, 1, 1, map[string]any{"identified": false}),
			item(56, "jew", "Rainbow Facet", "unique", 4, 2, 1, 1, nil),
			item(57, "jew", "Jewel", "normal", 5, 2, 1, 1, nil),
			item(58, "gsw", "Flawless Skull", "normal", 5, 1, 1, 1, nil),
			item(59, "amu", "Amulet", "magic", 4, 5, 1, 1, map[string]any{"identified": false}),
			item(60, "zzz", "Unmapped Thing", "magic", 5, 5, 1, 1, nil), // no icon in the set: the type fallback
		}
		for _, l := range []string{"backpack", "belt", "stash"} {
			for _, x := range v[l].([]any) {
				m := x.(map[string]any)
				m["art"] = m["code"]
				m["kind"] = fxKind[m["code"].(string)]
				if m["code"] == "zzz" {
					delete(m, "art")
				}
			}
		}
	}
}

// fxArt are the variants whose pages show fixture art (the page asks /api/art and /watch/art/ with its own address as the Referer).
var fxArt = map[string]bool{"phone": true}

type fixtureArt struct{}

func (fixtureArt) PNG(key string) ([]byte, error) {
	w, h := 56, 56
	switch {
	case key == "panel-inv":
		w, h = 320, 246
	case strings.HasPrefix(key, "slot-"):
		w, h = 40, 40
	default:
		if s, ok := phoneArt[strings.TrimPrefix(key, "item-")]; ok {
			w, h = 40*s[0], 40*s[1]
		}
	}
	f := fnv.New32a()
	f.Write([]byte(key))
	r, g, b := uint8(60+f.Sum32()%120), uint8(60+(f.Sum32()>>8)%120), uint8(60+(f.Sum32()>>16)%120)
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			switch {
			case key == "panel-inv":
				img.SetNRGBA(x, y, color.NRGBA{20, 18, 16, 255})
			case strings.HasPrefix(key, "slot-"):
				if (x+y)%7 == 0 {
					img.SetNRGBA(x, y, color.NRGBA{90, 90, 90, 80})
				}
			case strings.HasSuffix(key, "-jew"): // like the game's untinted jewel: a white diamond
				if dx, dy := abs(x-w/2), abs(y-h/2); dx+dy < w/2-3 {
					img.SetNRGBA(x, y, color.NRGBA{235, 235, 235, 255})
				}
			case x > 3 && y > 3 && x < w-4 && y < h-4: // a body with a lighter top so the picture has some detail at any size
				d := uint8(y * 60 / h)
				img.SetNRGBA(x, y, color.NRGBA{r + d/2, g + d/2, b + d/2, 255})
			}
		}
	}
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	return buf.Bytes(), err
}

var artAPI = func() http.Handler { a := watch.NewArtAPI(); a.Bind(fixtureArt{}); return a.Handler() }()

// artFor serves the art routes when the page's own address names a variant with art, and says whether it did.
func artFor(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != "/api/art" && !strings.HasPrefix(r.URL.Path, "/watch/art/") {
		return false
	}
	ref, err := url.Parse(r.Referer())
	if err != nil || !fxArt[ref.Query().Get("fx")] {
		return false
	}
	artAPI.ServeHTTP(w, r)
	return true
}

// defectCSS is, per name, the one defect of the 2026-11-15 screenshot put back on the phone page, unlayered so it wins over the page's own
// rules. The fit pages phonedefect-NAME must FAIL with the verdict mustFail names (fit.go); a pass there means the check is blind to it.
var defectCSS = map[string]string{
	"label":    `.cv .d-cells .item[data-shown] { font-size: 12px !important; }`,
	"acts":     `.cv .cv-sel .cv-acts { max-block-size: 30px !important; overflow: hidden !important; }`,
	"cutchips": `.cv .cv-bar { display: flex !important; flex-wrap: nowrap; overflow: hidden; } .cv .cv-bar > * { flex: none !important; } .cv .cv-chips { overflow: visible !important; mask-image: none !important; }`,
	"lock":     `.cv .cv-side > .cv-lock { white-space: nowrap !important; overflow: hidden !important; text-overflow: ellipsis !important; }`,
	"hover":    `.cv .cv-storez #stash { inline-size: 130% !important; max-inline-size: none !important; }`,
	"narrow":   `.cv #pack, .cv #stash { --u: 20px !important; }`,
	// the Characters page (fixtures_chars.go)
	"chars-clip":    `.cs-name { overflow: hidden !important; white-space: nowrap !important; max-inline-size: 3ch !important; text-overflow: ellipsis !important; }`,
	"chars-wide":    `.cs-card { min-inline-size: 130vw !important; }`,
	"chars-under":   `.cs-main { position: fixed !important; inset: 0 !important; }`,
	"chars-small":   `.cs-card .d-btn { min-block-size: 28px !important; block-size: 28px !important; }`,
	"chars-hole":    `.cs-card { min-block-size: 440px !important; }`,
	"chars-sheet":   `.cs-sheet-card { max-block-size: 160px !important; overflow: hidden !important; }`,
	"chars-covered": `.cs-head { position: fixed !important; inset: 0 0 auto 0 !important; block-size: 300px !important; z-index: 99; background: var(--d-surface-3) !important; }`,
	"resetclip":     `#btn-reset { position: absolute !important; inset-inline-start: -300px !important; }`, // the header Reset pushed out of the window (card resetinstant)
}

func serveDefectCSS(w http.ResponseWriter, r *http.Request) bool {
	n, ok := strings.CutSuffix(strings.TrimPrefix(r.URL.Path, "/__fit/defect-"), ".css")
	if !ok || defectCSS[n] == "" {
		return false
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Write([]byte(defectCSS[n]))
	return true
}
