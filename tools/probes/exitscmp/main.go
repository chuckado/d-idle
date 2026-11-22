// exitscmp compares the engine's own link table (tools/probes/exitsread.py output: tile exits and linked
// neighbours per level) with the bot's static exit data (Levels.txt Vis columns and nav.RouteHints), per level.
//
//	exitscmp -data DIR_WITH_MPQS exitsread.jsonl
//
// Per level it prints the engine links, the data links, and the two differences: "engine only" (the route graph has no
// such edge from the static data) and "data only" (an edge the engine never listed in any probed seed). Edges are
// undirected as nav's route graph adds them (route.go).
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	"d2lod-bot/internal/data"
	"d2lod-bot/internal/nav"
)

type run struct {
	Level  int                   `json:"level"`
	Warps  []struct{ Dest int }  `json:"warps"`
	Nearby []struct{ Level int } `json:"nearby"`
	Error  string                `json:"error"`
}

func main() {
	dir := flag.String("data", "", "directory with the game MPQs")
	flag.Parse()
	if *dir == "" || flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: exitscmp -data DIR exitsread.jsonl")
		os.Exit(2)
	}
	db, err := data.LoadInstall(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	static := map[int]map[int]string{}
	add := func(a, b int, src string) {
		if a == b || a == 0 || b == 0 {
			return
		}
		for _, p := range [][2]int{{a, b}, {b, a}} {
			if static[p[0]] == nil {
				static[p[0]] = map[int]string{}
			}
			if static[p[0]][p[1]] == "" {
				static[p[0]][p[1]] = src
			}
		}
	}
	for _, l := range db.Levels {
		for _, v := range l.Vis {
			add(l.ID, v, "vis")
		}
	}
	for _, h := range nav.RouteHints() {
		add(h[0], h[1], "hint")
	}
	engine := map[int]map[int]string{}
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	for sc.Scan() {
		var r run
		if json.Unmarshal(sc.Bytes(), &r) != nil || r.Error != "" || r.Level == 0 {
			continue
		}
		if engine[r.Level] == nil {
			engine[r.Level] = map[int]string{}
		}
		for _, w := range r.Warps {
			engine[r.Level][w.Dest] = "warp"
		}
		for _, n := range r.Nearby {
			if engine[r.Level][n.Level] == "" {
				engine[r.Level][n.Level] = "border"
			}
		}
	}
	var levels []int
	for l := range engine {
		levels = append(levels, l)
	}
	sort.Ints(levels)
	name := func(id int) string {
		if l := db.AreaByID(id); l != nil {
			return fmt.Sprintf("%d %s", id, l.Name)
		}
		return fmt.Sprint(id)
	}
	for _, l := range levels {
		var only, dataOnly []string
		for d, how := range engine[l] {
			if static[l][d] == "" {
				only = append(only, fmt.Sprintf("%s (%s)", name(d), how))
			}
		}
		for d, src := range static[l] {
			if engine[l][d] == "" {
				dataOnly = append(dataOnly, fmt.Sprintf("%s (%s)", name(d), src))
			}
		}
		sort.Strings(only)
		sort.Strings(dataOnly)
		fmt.Printf("level %s\n  engine only: %v\n  data only:   %v\n", name(l), only, dataOnly)
	}
}
