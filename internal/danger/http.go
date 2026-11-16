package danger

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// Route is GET /api/danger[?area=id or name][&difficulty=normal|nightmare|hell]. Read-only; without an area it lists
// the zones of the difficulty (default normal) in short form. Nothing orders anything (S-A: no behaviour change).
const Route = "/api/danger"

// View is the response body.
type View struct {
	Schema string  `json:"schema"`
	Zone   *Zone   `json:"zone,omitempty"`
	Bosses []*Boss `json:"bosses,omitempty"`
	Zones  []Brief `json:"zones,omitempty"`
	Error  string  `json:"error,omitempty"`
}

// Brief is one line of the zone list.
type Brief struct {
	Area int    `json:"area"`
	Name string `json:"name"`
	MLvl int    `json:"mlvl"`
	Bad  int    `json:"badPack"`
}

// Handler serves Route from a table.
func Handler(t *Table) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "GET only", http.StatusMethodNotAllowed)
			return
		}
		v := View{Schema: Schema}
		q := r.URL.Query()
		diff := strings.ToLower(strings.TrimSpace(q.Get("difficulty")))
		if diff == "" {
			diff = DiffNames[0]
		}
		area := strings.TrimSpace(q.Get("area"))
		switch {
		case !validDiff(diff):
			v.Error = "unknown difficulty"
		case area == "":
			for _, z := range t.Zones {
				if z.Diff == diff {
					v.Zones = append(v.Zones, Brief{z.Area, z.Name, z.MLvl, z.Encounters.Bad.N})
				}
			}
		default:
			if z := lookup(t, area, diff); z != nil {
				v.Zone = z
				for _, b := range t.Bosses {
					if b.Area == z.Area && b.Diff == diff {
						v.Bosses = append(v.Bosses, b)
					}
				}
			} else {
				v.Error = "no such area"
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(v)
	})
}

func lookup(t *Table, ref, diff string) *Zone {
	id, err := strconv.Atoi(ref)
	for _, z := range t.Zones {
		if z.Diff == diff && ((err == nil && z.Area == id) || strings.EqualFold(z.Name, ref)) {
			return z
		}
	}
	return nil
}
