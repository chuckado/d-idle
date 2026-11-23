package main

import "testing"

func TestEnvKnobs(t *testing.T) {
	got := envKnobs(map[string]any{"quests.board": 1}, "nav.fullGrid=1, nav.warpRead=1,bad,x=y,=2")
	if len(got) != 3 || got["nav.fullGrid"] != 1.0 || got["nav.warpRead"] != 1.0 || got["quests.board"] != 1 {
		t.Fatalf("%v", got)
	}
	if envKnobs(nil, "") != nil {
		t.Fatal("empty spec must not create a map")
	}
}
