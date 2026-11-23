package main

import (
	"fmt"
	"strings"
	"testing"
)

func wordsLine(d int, set map[int]int) string {
	f := []string{fmt.Sprint(d)}
	for i := 0; i < 48; i++ {
		f = append(f, fmt.Sprint(set[i]))
	}
	return strings.Join(f, "\t")
}

func TestQuestWordsReadsTheActiveDifficulty(t *testing.T) {
	lines := []string{wordsLine(0, map[int]int{3: 12, 20: 7}), wordsLine(1, map[int]int{20: 9, 40: 0x2011}), wordsLine(2, nil)}
	w := questWords(lines, "active=1")
	if w[20] != 9 || w[40] != 0x2011 || len(w) != 48 {
		t.Fatalf("normal difficulty words wrong: %v", w)
	}
	if w := questWords(lines, "active=0"); w[20] != 7 || w[3] != 12 {
		t.Fatalf("difficulty 0: %v", w)
	}
}

// must-fail control: a missing verb, a short row or no active marker is nil, never a guess.
func TestQuestWordsRefusesWhatItCannotRead(t *testing.T) {
	good := wordsLine(0, nil)
	for name, c := range map[string]struct {
		lines []string
		rest  string
	}{
		"no active":  {[]string{good}, ""},
		"other diff": {[]string{good}, "active=2"},
		"short row":  {[]string{"0\t1 2 3"}, "active=0"},
		"err reply":  {nil, "active=0"},
		"not number": {[]string{strings.Replace(good, "0", "x", 2)}, "active=0"},
	} {
		if w := questWords(c.lines, c.rest); w != nil {
			t.Errorf("%s: want nil, got %d words", name, len(w))
		}
	}
}

func TestMergeKeepsTheListedWords(t *testing.T) {
	dst := map[int]uint16{3: 12}
	n := mergeQuestWords(dst, map[int]uint16{3: 99, 20: 7})
	if n != 1 || dst[3] != 12 || dst[20] != 7 {
		t.Fatalf("merge: %d %v", n, dst)
	}
}
