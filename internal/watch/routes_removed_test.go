package watch

import "testing"

// The user deleted these viewer routes on 2026-11-23: nothing in the UI called them. The unlock engine stays.
func TestRemovedRoutesStayGone(t *testing.T) {
	r, _ := bestiaryRig(t, "Sorceress")
	for _, p := range []string{"/api/bestiary", "/api/atlas", "/api/unlocks"} {
		for _, base := range []string{r.view, r.answer} {
			if resp, _ := r.do("GET", base, p, funToken, ""); resp.StatusCode == 200 {
				t.Errorf("GET %s on %s answers %d, want no handler", p, base, resp.StatusCode)
			}
		}
	}
}
