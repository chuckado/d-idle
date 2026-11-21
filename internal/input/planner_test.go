package input

import (
	"errors"
	"testing"
	"testing/fstest"
)

func TestPlannerOrderAndCounts(t *testing.T) {
	var did []string
	state := ""
	a := func(verbSets bool) Action {
		return Action{Name: "x", Cap: "cap",
			Verb: func() error {
				did = append(did, "verb")
				if verbSets {
					state = "v"
				}
				return nil
			},
			Key:    func() error { did = append(did, "key"); state = "k"; return nil },
			Click:  func() error { did = append(did, "click"); state = "c"; return nil },
			Verify: func() (bool, error) { return state != "", nil }}
	}

	p := &Planner{Caps: &BackendCaps{Verbs: []string{"cap"}}}
	state, did = "", nil
	if path, err := p.Run(a(true)); err != nil || path != PathVerb || len(did) != 1 {
		t.Fatalf("%v %v %v", path, err, did)
	}
	state, did = "", nil
	if path, err := p.Run(a(false)); err != nil || path != PathKey || len(did) != 2 {
		t.Fatalf("verb without effect must fall to key: %v %v %v", path, err, did)
	}
	q := &Planner{} // no capability: the verb is never tried
	state, did = "", nil
	if path, _ := q.Run(a(true)); path != PathKey || did[0] != "key" {
		t.Fatalf("%v %v", path, did)
	}
	if v, k, c, f := p.Counts(); v != 1 || k != 1 || c != 0 || f != 0 {
		t.Fatalf("%d %d %d %d", v, k, c, f)
	}
	if p.Summary() != "1/1/0" {
		t.Fatal(p.Summary())
	}
}

func TestPlannerNoPathAndErrors(t *testing.T) {
	p := &Planner{}
	_, err := p.Run(Action{Name: "n", Click: func() error { return nil }, Verify: func() (bool, error) { return false, nil }})
	if !errors.Is(err, ErrNoPath) {
		t.Fatal(err)
	}
	boom := errors.New("boom")
	_, err = p.Run(Action{Name: "n", Click: func() error { return boom }, Verify: func() (bool, error) { return true, nil }})
	if !errors.Is(err, boom) {
		t.Fatal(err)
	}
	if _, _, _, f := p.Counts(); f != 2 {
		t.Fatalf("failed %d", f)
	}
	if (*Planner)(nil).On() {
		t.Fatal("nil planner is off")
	}
}

func TestPlannerLiveCapsAskedOnce(t *testing.T) {
	asked := 0
	p := &Planner{CapsFn: func() []string { asked++; return []string{"cap"} }}
	for i := 0; i < 3; i++ {
		state := false
		a := Action{Name: "x", Cap: "cap", Verb: func() error { state = true; return nil }, Verify: func() (bool, error) { return state, nil }}
		if path, err := p.Run(a); err != nil || path != PathVerb {
			t.Fatal(path, err)
		}
	}
	if asked != 1 {
		t.Fatalf("asked %d", asked)
	}
}

func TestBackendCaps(t *testing.T) {
	fsys := fstest.MapFS{"caps/a.json": {Data: []byte(`{"schema":"inputcaps/1","backend":"a","verbs":["input.select"]}`)},
		"caps/b.json": {Data: []byte(`{"schema":"inputcaps/1","backend":"b","verbs":[],"bogus":1}`)}}
	c, err := BackendCapsByName(fsys, "a")
	if err != nil || !c.Has(CapSelect) || c.Has("other") {
		t.Fatalf("%v %v", c, err)
	}
	if _, err := BackendCapsByName(fsys, "b"); err == nil {
		t.Fatal("unknown field must be an error")
	}
	if _, err := BackendCapsByName(fsys, "zzz"); err == nil {
		t.Fatal("unknown backend must be an error")
	}
	if (*BackendCaps)(nil).Has("x") {
		t.Fatal("nil caps has nothing")
	}
}
