package main

import (
	"bufio"
	"os"
	"strings"
)

// debugWrites are the `debug <sub>` verbs that change the game (V docs/forkcheats.md: act, hit, merc hire|kill|revive|hurt,
// give, durability, belt clear, unequip, spawn, quest set|bit|join, pin). The bot also sends many `debug` reads (players,
// portals, states, missiles, level, objects, presets, blocked, warps, grid, layout: V the journal of a live run), which are
// not writes. U: a new write verb of the fork must be added here.
var debugWrites = map[string]bool{"act": true, "hit": true, "merc": true, "give": true, "durability": true, "belt": true,
	"unequip": true, "spawn": true, "pin": true}

func isWriteVerb(w []string) bool {
	if len(w) < 2 || w[0] != "debug" {
		return false
	}
	if w[1] == "quest" {
		return len(w) > 2 && w[2] != "list"
	}
	return debugWrites[w[1]]
}

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
		if isWriteVerb(w) && !strings.HasPrefix(p[3], "err:") && !strings.HasPrefix(p[3], "fail:") {
			n++
		}
	}
	return n, true
}
