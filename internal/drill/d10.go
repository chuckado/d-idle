package drill

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// D10: fault recovery on d2d, the Gate 1 injected fault (docs/review-strategy.md "d2d gate ladder", soak-gates
// "Injected fault recovery on d2d"). The managed bot loop runs on a hidden slot, the slot's d2d process is killed with
// SIGKILL, and the criteria below are read from the bot's own log (the minute lines), the process table and the
// character's save file. This file is the pure part (parse and verdict); cmd/d2drill/d10_lab.go starts the processes.

// Minute is the bot's once-a-minute log line (internal/bot/observe.go minuteLine), the fields the drill needs.
type Minute struct {
	Time     time.Time
	Area     int
	Clvl     int
	XP       int64
	Kills    int64
	Restarts int
	XPRisk   int64
	XPLost   int64
}

// ParseMinute reads one line of the bot's JSON log; ok is false for any other line.
func ParseMinute(line string) (Minute, bool) {
	if !strings.Contains(line, `"msg":"minute"`) {
		return Minute{}, false
	}
	var raw struct {
		Time     time.Time `json:"time"`
		Msg      string    `json:"msg"`
		Area     int       `json:"area"`
		Clvl     int       `json:"clvl"`
		XP       int64     `json:"xp"`
		Kills    int64     `json:"kills"`
		Restarts int       `json:"restarts"`
		XPRisk   int64     `json:"xpRisk"`
		XPLost   int64     `json:"xpLost"`
	}
	if err := json.Unmarshal([]byte(line), &raw); err != nil || raw.Msg != "minute" {
		return Minute{}, false
	}
	return Minute{raw.Time, raw.Area, raw.Clvl, raw.XP, raw.Kills, raw.Restarts, raw.XPRisk, raw.XPLost}, true
}

// Event is a bot log line that tells which game restarts happened and why.
type Event struct {
	Time time.Time
	Kind string // "leave" (a Save and Exit of the game, any cause), "savepoint" (planned), "lost" (the failsafe for a vanished d2d)
}

// ParseEvent classifies a line of the bot's JSON log; ok is false for lines that are no restart event.
func ParseEvent(line string) (Event, bool) {
	var kind string
	switch {
	case strings.Contains(line, "Save and Exit requested (chicken / leave game)"):
		kind = "leave"
	case strings.Contains(line, `"msg":"save point: saving and re-entering the game"`):
		kind = "savepoint"
	case strings.Contains(line, `"msg":"game lost, restarting it"`):
		kind = "lost"
	default:
		return Event{}, false
	}
	var raw struct {
		Time time.Time `json:"time"`
	}
	if json.Unmarshal([]byte(line), &raw) != nil {
		return Event{}, false
	}
	return Event{raw.Time, kind}, true
}

// D10Facts is everything the live side observed.
type D10Facts struct {
	KillAt    time.Time
	KilledPID int
	Minutes   []Minute // every minute line of the bot, in order, over all bot processes
	Events    []Event  // the restart events of the bot log, in order

	NewD2DAt   time.Time // first moment a d2d with another pid answered on the slot's socket (zero: never seen)
	NewD2DPID  int
	BotExits   int    // bot process exits during the observation (the soak supervisor's restarts)
	SavedPre   int64  // experience in the character's save file just before the kill (-1 unknown)
	SavedEnd   int64  // experience in the save file after the graceful stop (-1 unknown)
	StagePre   string // P35 gate before the run ("clean" or what was open)
	StagePost  string // P35 gate after, on the character loaded from the save
	Leftover   string // `tools/agentreap.sh --list` and `--orphans` lines that name the killed pid or its process group ("" none)
	StillThere bool   // the killed pid is still in the process table (a zombie counts)
	Registry   string // the registry evidence: killed pid, its pgid, the registered group it belongs to
}

// Limits of the pass condition (docs/review-strategy.md fault-injection procedure).
const (
	D10BackWithin = 3 * time.Minute
	D10LoopWithin = 5 * time.Minute
	// D10CountSettle is how long after the first minute line back in game the restarts counter is read: later rises are
	// other restarts (a save point is followed by a fresh game every few minutes, V live D10 pilot runs), not the
	// recovery of the kill, which re-entered the game within 2 s.
	D10CountSettle = 10 * time.Second
)

// EvalD10 applies the pass condition: the bot is back in game within 3 min (a minute line, which exists only in
// game), the restarts counter rose by exactly 1, saved XP did not drop, the loop resumed within 5 min of that (a kill,
// XP or area change), the killed process is gone from the process table and the registry, and the P35 gate was clean
// before and after.
func EvalD10(f D10Facts) Outcome {
	var pre, post []Minute
	for _, m := range f.Minutes {
		if m.Time.Before(f.KillAt) {
			pre = append(pre, m)
		} else {
			post = append(post, m)
		}
	}
	o := Outcome{Intended: "d2d killed with SIGKILL: back in game within 3 min, restarts +1, saved XP not lower, loop resumes within 5 min, no process of the kill left, P35 clean before and after"}
	if len(pre) == 0 {
		return Blocks("no minute line before the kill: the bot never reached the loop")
	}
	last := pre[len(pre)-1]
	var bad []string
	obs := []string{fmt.Sprintf("kill pid %d at %s (%s)", f.KilledPID, f.KillAt.Format("15:04:05"), f.Registry),
		fmt.Sprintf("pre-kill minute: area %d clvl %d xp %d kills %d restarts %d", last.Area, last.Clvl, last.XP, last.Kills, last.Restarts)}
	if len(post) == 0 {
		bad = append(bad, "no minute line after the kill: the bot never came back into a game")
		obs = append(obs, "no minute line after the kill")
	} else {
		first := post[0]
		back := first.Time.Sub(f.KillAt)
		obs = append(obs, fmt.Sprintf("first minute line after the kill +%s: area %d xp %d kills %d restarts %d xpLost %d", back.Round(time.Second), first.Area, first.XP, first.Kills, first.Restarts, first.XPLost))
		if back > D10BackWithin {
			bad = append(bad, fmt.Sprintf("back in game after %s, limit %s", back.Round(time.Second), D10BackWithin))
		}
		end := post[len(post)-1]
		inWin := first
		for _, m := range post {
			if m.Time.Sub(first.Time) <= D10CountSettle {
				inWin = m
			}
		}
		delta := inWin.Restarts - last.Restarts
		// Attribution: the recovery of the kill is one leave (kind "lost"); every other unplanned leave in the window
		// (a fresh game after a save point, V live) raises the counter by itself.
		var lost, leaves, saves int
		for _, e := range f.Events {
			if e.Time.Before(f.KillAt) || e.Time.After(inWin.Time) {
				continue
			}
			switch e.Kind {
			case "lost":
				lost++
			case "leave":
				leaves++
			case "savepoint":
				saves++
			}
		}
		other := max(leaves-lost-saves, 0) // leaves that are neither the recovery nor a planned save point
		obs = append(obs, fmt.Sprintf("restart events up to then: %d game-lost recoveries, %d planned save points, %d other new games (%d leaves in all)", lost, saves, other, leaves))
		obs = append(obs, fmt.Sprintf("restarts counter of the minute line: %d before, %d at +%s (delta %d), %d on the last line +%s (rises after that are other restarts)",
			last.Restarts, inWin.Restarts, inWin.Time.Sub(f.KillAt).Round(time.Second), delta, end.Restarts, end.Time.Sub(f.KillAt).Round(time.Second)))
		switch {
		case delta != 1:
			bad = append(bad, fmt.Sprintf("restarts counter rose by %d within %s of the first minute line back in game, want exactly 1 (the game-lost recovery of d2d is not counted: internal/bot/loop.go gameLost calls noteReentry, not noteRecycle)", delta, D10CountSettle))
		case other > 0:
			bad = append(bad, fmt.Sprintf("restarts counter rose by 1 but %d other new game(s) happened in the window: the rise is theirs, the recovery itself is not counted", other))
		}
		if lost != 1 {
			bad = append(bad, fmt.Sprintf("%d game-lost recoveries logged, want exactly 1", lost))
		}
		if first.XP < f.SavedPre && f.SavedPre >= 0 {
			bad = append(bad, fmt.Sprintf("xp after the restart %d is below the saved xp %d before the kill", first.XP, f.SavedPre))
		}
		resumed := ""
		for _, m := range post[1:] {
			if m.Time.Sub(first.Time) > D10LoopWithin {
				break
			}
			switch {
			case m.Kills > first.Kills:
				resumed = fmt.Sprintf("kills %d -> %d", first.Kills, m.Kills)
			case m.XP > first.XP:
				resumed = fmt.Sprintf("xp %d -> %d", first.XP, m.XP)
			case m.Area != first.Area:
				resumed = fmt.Sprintf("area %d -> %d", first.Area, m.Area)
			}
			if resumed != "" {
				break
			}
		}
		if resumed == "" {
			bad = append(bad, "the loop did not resume within 5 min of the first minute line (no kill, XP or area change)")
			obs = append(obs, "loop: no event")
		} else {
			obs = append(obs, "loop resumed: "+resumed)
		}
	}
	if !f.NewD2DAt.IsZero() {
		obs = append(obs, fmt.Sprintf("new d2d pid %d seen +%s", f.NewD2DPID, f.NewD2DAt.Sub(f.KillAt).Round(time.Second)))
	} else {
		bad = append(bad, "no new d2d process was seen on the slot's socket")
		obs = append(obs, "no new d2d")
	}
	obs = append(obs, fmt.Sprintf("saved xp (save file) before the kill %d, after the graceful stop %d; bot process exits %d", f.SavedPre, f.SavedEnd, f.BotExits))
	if f.SavedPre >= 0 && f.SavedEnd >= 0 && f.SavedEnd < f.SavedPre {
		bad = append(bad, fmt.Sprintf("saved xp fell from %d to %d", f.SavedPre, f.SavedEnd))
	}
	if f.StagePre != "clean" || f.StagePost != "clean" {
		bad = append(bad, fmt.Sprintf("P35: before %q, after %q", f.StagePre, f.StagePost))
	}
	obs = append(obs, fmt.Sprintf("P35 before %s, after %s", f.StagePre, f.StagePost))
	if f.StillThere {
		bad = append(bad, "the killed pid is still in the process table")
	}
	if f.Leftover != "" {
		bad = append(bad, "registry or orphan list names the killed process: "+f.Leftover)
	}
	obs = append(obs, "leftovers: "+map[bool]string{true: "none", false: f.Leftover}[f.Leftover == "" && !f.StillThere])
	o.Observed = strings.Join(obs, " | ")
	o.Sig = fmt.Sprintf("bad=%d", len(bad)) // the two runs differ in numbers; the verdict is what repeats
	if len(bad) > 0 {
		o.Verdict, o.Reason = Fail, strings.Join(bad, "; ")
		return o
	}
	o.Verdict, o.Reason = Pass, "d2d killed and restarted by the bot, counters and saves as required"
	return o
}
