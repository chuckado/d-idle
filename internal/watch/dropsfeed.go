package watch

import (
	"net/http"
	"sort"
	"strconv"
)

// GET /api/drops: the items the hero picked up, newest first (the Items tab of the log drawer reads it through /api/log).
// The source is the away log's drop events, which the looter emits when an item reaches the pack. U: the log does not
// record what the bot did with the item afterwards (kept, sold, equipped), so a row says "picked up" and nothing more.
const (
	DropsRoute  = "/api/drops"
	dropsSchema = "watch.drops/1"
	dropsMax    = 40
)

// DropRow is one pickup.
type DropRow struct {
	ID      string  `json:"id"`
	Ago     float64 `json:"ago"`
	Item    string  `json:"item"`
	Quality string  `json:"quality,omitempty"`
	Area    string  `json:"area,omitempty"`
}

// DropsView is the reply of DropsRoute.
type DropsView struct {
	Schema string    `json:"schema"`
	Rows   []DropRow `json:"rows"`
}

// DropsHandler serves DropsRoute from the away log; now gives the log's time base.
func DropsHandler(l *AwayLog, now func() float64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := DropsView{Schema: dropsSchema, Rows: []DropRow{}}
		evs := l.Events()
		n := now()
		for i := len(evs) - 1; i >= 0 && len(v.Rows) < dropsMax; i-- {
			e := evs[i]
			if e.Kind != AwayDrop || e.Item == "" {
				continue
			}
			ago := n - e.T
			if ago < 0 {
				ago = 0
			}
			v.Rows = append(v.Rows, DropRow{ID: "d" + strconv.Itoa(i), Ago: ago, Item: e.Item, Quality: e.Quality, Area: e.Area})
		}
		sort.SliceStable(v.Rows, func(i, j int) bool { return v.Rows[i].Ago < v.Rows[j].Ago })
		writeJSON(w, http.StatusOK, v)
	})
}
