package main

import (
	"bufio"
	"os"
	"strings"
)

// writeVerbs are the first words of the lab verbs that change the game (docs/forkcheats.md): a bot that sends one after
// the start of a measured run made a gameplay-affecting write. Reads (dump, state, items, ...) are not listed.
var writeVerbs = map[string]bool{"debug": true} // U: the item and vault verbs of --item-verbs are not counted

// journalWrites counts the mutating commands the bot sent that the d2d did not refuse, from its d2dctl command journal
// (">\t<tick>\t<command>\t<reply>\t<site>"). A field start has no refused-write counter (the verbs are on), so this is
// its writes-after-start line. ok is false when the journal cannot be read: the assertion stays unobserved.
func journalWrites(path string) (int, bool) {
	f, err := os.Open(path)
	if err != nil {
		return 0, false
	}
	defer f.Close()
	n := 0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		p := strings.Split(sc.Text(), "\t")
		if len(p) < 4 || p[0] != ">" {
			continue
		}
		w := strings.Fields(p[2])
		if len(w) > 0 && writeVerbs[w[0]] && !strings.HasPrefix(p[3], "err:") && !strings.HasPrefix(p[3], "fail:") {
			n++
		}
	}
	return n, true
}
