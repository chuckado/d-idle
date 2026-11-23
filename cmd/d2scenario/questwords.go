package main

import (
	"strconv"
	"strings"
)

// questWords reads `debug quest words` (devp23 labverbs: 3 payload lines "<diff>\t<48 decimal words>", then
// "ok active=<d>"; S source, V the fork's labverbs test) and returns the 48 words of the ACTIVE difficulty by record
// number. It returns nil when the verb is missing or the reply does not parse (a stock or older lab d2d: records 16 to
// 47 then stay unobserved, as before). `debug quest list` stays the primary source; these words only fill records it
// did not print (it lists record numbers up to 15 on the lab builds seen so far, matrix note 30).
func questWords(lines []string, rest string) map[int]uint16 {
	active := -1
	for _, f := range strings.Fields(rest) {
		if v, ok := strings.CutPrefix(f, "active="); ok {
			if n, err := strconv.Atoi(v); err == nil {
				active = n
			}
		}
	}
	if active < 0 {
		return nil
	}
	for _, l := range lines {
		f := strings.Fields(l)
		if len(f) != 49 {
			continue
		}
		if d, err := strconv.Atoi(f[0]); err != nil || d != active {
			continue
		}
		out := make(map[int]uint16, 48)
		for i, w := range f[1:] {
			n, err := strconv.ParseUint(w, 10, 16)
			if err != nil {
				return nil
			}
			out[i] = uint16(n)
		}
		return out
	}
	return nil
}

// mergeQuestWords adds the words of src for records dst lacks and returns the number added.
func mergeQuestWords(dst map[int]uint16, src map[int]uint16) int {
	n := 0
	for r, w := range src {
		if _, ok := dst[r]; !ok {
			dst[r] = w
			n++
		}
	}
	return n
}
