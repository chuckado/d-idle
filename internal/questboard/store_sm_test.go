package questboard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func smStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	st, err := Open(smBoard(t), dir, "Hero")
	if err != nil {
		t.Fatal(err)
	}
	st.now = func() time.Time { return smT0 }
	return st, filepath.Join(dir, "Hero.quests.json")
}

func smRec(quest, chunk string) Record {
	return Record{Quest: quest, Chunk: chunk, Diff: "normal", At: smT0, Mode: "train"}
}

func seed(t *testing.T, st *Store, rs ...Record) {
	t.Helper()
	if err := st.update(func(p *Plan) error { p.Accepted = rs; return nil }); err != nil {
		t.Fatal(err)
	}
}

// The v1 shapes the import has to read (appendix A of the design: one open record, a chain, two open records,
// a record of a done chunk, a record whose prerequisite has none).
func TestImportV1Shapes(t *testing.T) {
	const q = "search_for_cain"
	doneScroll := func(quest, chunk string, diff int) (Tier, bool) { return TierProxy, quest == q && chunk == "scroll" }
	cases := []struct {
		name string
		recs []Record
		done DoneFn
		want map[string]RunStatus // chunk -> status after load
	}{
		{"one open", []Record{smRec(q, "scroll")}, nil, map[string]RunStatus{"scroll": RunRunning}},
		{"chain", []Record{smRec(q, "scroll"), smRec(q, "read")}, nil, map[string]RunStatus{"scroll": RunRunning, "read": RunQueued}},
		{"two open records", []Record{smRec(q, "read"), smRec(q, "scroll")}, nil, map[string]RunStatus{"scroll": RunRunning, "read": RunQueued}},
		{"record of a done chunk", []Record{smRec(q, "scroll"), smRec(q, "read")}, doneScroll, map[string]RunStatus{"scroll": RunFinished, "read": RunRunning}},
		{"prerequisite without a record (bug 3)", []Record{smRec(q, "stones")}, nil, map[string]RunStatus{"scroll": RunRunning, "read": RunQueued, "stones": RunQueued}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st, _ := smStore(t)
			seed(t, st, c.recs...)
			m, err := st.LoadSM(c.done)
			if err != nil {
				t.Fatal(err)
			}
			got := map[string]RunStatus{}
			for _, r := range m.Runs {
				got[r.Chunk] = r.Status
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %v want %v", got, c.want)
			}
			if v := m.Check(st.b, "normal"); len(v) != 0 {
				t.Fatalf("violations %v", v)
			}
			if c.done != nil && !m.Latched(q, "scroll", "normal") {
				t.Fatal("a finished import must carry a latch")
			}
		})
	}
}

func TestLoadSMIsIdempotentAndPersists(t *testing.T) {
	st, path := smStore(t)
	seed(t, st, smRec("search_for_cain", "stones"))
	if _, err := st.LoadSM(nil); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(path)
	if _, err := st.LoadSM(nil); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Fatal("a second load rewrote the file")
	}
	// running is not stored (3.7): the head is recomputed at load
	if strings.Contains(string(first), `"running"`) || strings.Contains(string(first), `"held"`) {
		t.Fatalf("running or held was stored:\n%s", first)
	}
}

// Dual write: Accepted equals the live and finished runs after every UpdateSM, and an older build (a struct without
// Runs and Latches, the v1 Plan) reads the file and finds its records.
func TestDualWriteAndOldBuildRead(t *testing.T) {
	st, path := smStore(t)
	if err := st.UpdateSM(nil, func(m *SM) error {
		_, _, err := m.Start(st.b, "normal", "search_for_cain", "stones", "player", smT0, nil)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	var env struct {
		Version int
		Data    json.RawMessage
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	if env.Version != PlanVersion || PlanVersion != 1 {
		t.Fatalf("version %d: the file stays version 1 so an older build still loads it", env.Version)
	}
	var v1 struct {
		Accepted []Record `json:"accepted"`
	}
	if err := json.Unmarshal(env.Data, &v1); err != nil {
		t.Fatal(err)
	}
	var keys []string
	for _, r := range v1.Accepted {
		keys = append(keys, r.Chunk)
	}
	if !reflect.DeepEqual(keys, []string{"scroll", "read", "stones"}) {
		t.Fatalf("old build sees %v", keys)
	}
	// the old path withdrew a record (knob back at 0): the next SM load stops the run instead of resurrecting it
	if err := st.Cancel("search_for_cain", "stones", 0); err != nil {
		t.Fatal(err)
	}
	m, _ := st.LoadSM(nil)
	for _, r := range m.Runs {
		if r.Chunk == "stones" && r.Status != RunStopped {
			t.Fatalf("withdrawn record is %s", r.Status)
		}
	}
}

func TestUpdateSMErrorSavesNothing(t *testing.T) {
	st, path := smStore(t)
	seed(t, st, smRec("search_for_cain", "scroll"))
	before, _ := os.ReadFile(path)
	err := st.UpdateSM(nil, func(m *SM) error {
		m.Stop(st.b, "normal", "search_for_cain", "scroll")
		return os.ErrInvalid
	})
	after, _ := os.ReadFile(path)
	if err != os.ErrInvalid || string(before) != string(after) {
		t.Fatalf("err %v, file changed %v", err, string(before) != string(after))
	}
}

func TestStopPersistsAndStoppedAreBounded(t *testing.T) {
	st, _ := smStore(t)
	for i := 0; i < maxStoppedRuns+5; i++ {
		if err := st.UpdateSM(nil, func(m *SM) error {
			if _, _, err := m.Start(st.b, "normal", "den_of_evil", "clear_den", "player", smT0, nil); err != nil {
				return err
			}
			m.Stop(st.b, "normal", "den_of_evil", "clear_den")
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	m, _ := st.LoadSM(nil)
	if len(m.Runs) > maxStoppedRuns || len(m.Live(st.b, "normal")) != 0 {
		t.Fatalf("%d runs kept", len(m.Runs))
	}
}
