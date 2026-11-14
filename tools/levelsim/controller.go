package levelsim

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"d2lod-bot/internal/power"
	"d2lod-bot/internal/survival"
)

// Controller replay (survival slice S-C, docs/survival/replay-2026-11-14.md): feeds a recorded bot log through the
// ledger and the shadow controller exactly as the bot's hooks do (cmd/d2bot/survivalwire.go): fights, chickens, deaths
// and merc deaths are observed, and Choose runs at every `planner` pick and every `minute` line, with the planner's own
// choice next to it. Candidates are the areas the planner named in the log (the plan's steps), judged by the same
// margin; reachability is NOT modelled (docs/survival/ledger.md deliberate limits). Log only; nothing here is shipped.

// CtlOpts selects the model inputs of a controller replay.
type CtlOpts struct {
	ShadowOpts
	Params *survival.Params
}

// CtlRow is one decision of the replay.
type CtlRow struct {
	T        time.Time
	Kind     string // "pick" (a planner line stood) or "minute"
	Clvl     int
	Cur      int // the zone the hero was in
	Planner  int // the planner's standing pick
	Shadow   int
	Word     string // margin word of the shadow's choice
	Code     string // why code
	Hole     bool
	Fresh    bool
	MercDead bool
	Revive   string // "yes", "no:<code>" or "" without a dead merc
	Chick    int    // chickens since the previous row
	Deaths   int
	MercDied int
}

// CtlReport is the result of one controller replay.
type CtlReport struct {
	Rows      []CtlRow
	Names     map[int]string
	KHero     float64
	KZone     map[int]float64
	CandAreas []int
	Dial      string
	Skipped   []string
}

// ControllerReplay reads a bot log (text key=value lines) and runs the controller over it.
func ControllerReplay(r io.Reader, o CtlOpts) (*CtlReport, error) {
	if o.Tab == nil || o.M == nil || o.Arch == nil || o.Params == nil {
		return nil, fmt.Errorf("levelsim: controller replay needs a danger table, a model, an archetype and survival params")
	}
	dialID := o.Dial
	if dialID == "" {
		dialID = o.M.P.DefaultDial
	}
	dial, err := o.M.P.Dial(dialID)
	if err != nil {
		return nil, err
	}
	diff := o.Diff
	if diff == "" {
		diff = "normal"
	}
	// pass 1: the planner's candidate areas
	lines, areas := []string{}, map[int]bool{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		ln := strings.TrimSpace(sc.Text())
		lines = append(lines, ln)
		if strings.Contains(ln, `msg=planner`) {
			if a, err := strconv.Atoi(textKV(ln)["area"]); err == nil {
				areas[a] = true
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	rep := &CtlReport{Names: map[int]string{}, KZone: map[int]float64{}, Dial: dialID}
	for a := range areas {
		if z, err := o.Tab.Find(a, diff); err == nil {
			rep.CandAreas = append(rep.CandAreas, a)
			rep.Names[a] = z.Name
		} else {
			rep.Skipped = append(rep.Skipped, fmt.Sprintf("area %d not in the profile", a))
		}
	}
	sort.Ints(rep.CandAreas)
	if len(rep.CandAreas) == 0 {
		return nil, fmt.Errorf("levelsim: no planner area of the log is in the zone profile")
	}

	led := survival.NewLedger(o.Params)
	var (
		hero             power.Hero
		gold             int
		cur, plan        int
		last             time.Time
		chick, dth, mdie int
		pending          *CtlRow
		pendingArea      int
	)
	hero = power.Hero{Class: "log", Potions: 6}
	pw := func() float64 { return survival.PowerOf(hero) }
	build := func(h power.Hero) ([]survival.Candidate, error) {
		var out []survival.Candidate
		for _, a := range rep.CandAreas {
			z, _ := o.Tab.Find(a, diff)
			zm, err := o.M.Zone(h, o.Arch, z, dialID)
			if err != nil {
				return nil, err
			}
			tm := 2.0
			if a == cur {
				tm = 0
			}
			out = append(out, survival.Candidate{Key: survival.ZoneKey{Area: a, Diff: diff}, Name: z.Name, ZM: zm, Eco: survival.Economy(z, zm, h.Level, o.Params), TravelMin: tm})
		}
		return out, nil
	}
	decide := func(kind string) error {
		if hero.Level <= 0 || hero.MaxLife <= 0 {
			return nil
		}
		cs, err := build(hero)
		if err != nil {
			return err
		}
		need := 0
		if hero.Potions < 4 {
			need = (4 - hero.Potions) * int(o.Params.PotionPrice)
		}
		dead := hero.Merc.Hired && !hero.Merc.Alive
		const quote = 400 // the bot's constant (U)
		if dead {
			need += quote
		}
		in := survival.Input{Now: last, Dial: dial, Current: survival.ZoneKey{Area: cur, Diff: diff}, Power: pw(), Purse: survival.Purse{Gold: gold, Need: need}, Cands: cs, FreshAllowed: true}
		d, err := led.Choose(in)
		if err != nil {
			return err
		}
		row := CtlRow{T: last, Kind: kind, Clvl: hero.Level, Cur: cur, Planner: plan, Shadow: d.Area, Word: d.Margin, Code: d.Why.Code,
			Hole: d.Hole, Fresh: d.Fresh, MercDead: dead, Chick: chick, Deaths: dth, MercDied: mdie}
		chick, dth, mdie = 0, 0, 0
		if dead {
			alive := hero
			alive.Merc.Alive = true
			with, err := build(alive)
			if err == nil {
				rd := led.Revive(survival.ReviveInput{Base: in, With: with, Without: cs, Quote: quote, Reserve: 4 * int(o.Params.PotionPrice), Gold: gold})
				row.Revive = "no:" + rd.Why.Code
				if rd.Revive {
					row.Revive = "yes"
				}
			}
		}
		rep.Rows = append(rep.Rows, row)
		return nil
	}
	flush := func() error { // a planner pick stands once nothing marked it skipped
		if pending == nil {
			return nil
		}
		plan = pendingArea
		pending = nil
		return decide("pick")
	}
	for _, ln := range lines {
		kv := textKV(ln)
		if ts, err := time.Parse(time.RFC3339, kv["time"]); err == nil {
			if pending != nil && ts.After(last) {
				if err := flush(); err != nil {
					return nil, err
				}
			}
			last = ts
		}
		switch kv["msg"] {
		case "zone entry":
			a, _ := strconv.Atoi(kv["area"])
			m := hero.Merc
			hero = heroOfEntry(kv)
			hero.Merc.Level = m.Level
			if rep.Names[a] != "" {
				cur = a
			}
		case "merc: seen":
			lv, _ := strconv.Atoi(kv["level"])
			hero.Merc = power.Merc{Hired: true, Alive: kv["alive"] == "true", Level: lv}
			hero.SetQuality(power.TermMerc, power.Derived)
		case "merc: died":
			hero.Merc.Hired, hero.Merc.Alive = true, false
			mdie++
			if cur != 0 {
				led.Observe(survival.ZoneKey{Area: cur, Diff: diff}, last, survival.Event{Kind: survival.EvMercDeath}, pw())
			}
		case "safety decision":
			if kv["action"] == "chicken" {
				a, _ := strconv.Atoi(kv["area"])
				if rep.Names[a] != "" {
					chick++
					led.Observe(survival.ZoneKey{Area: a, Diff: diff}, last, survival.Event{Kind: survival.EvChicken}, pw())
				}
			}
		case "character died":
			a, _ := strconv.Atoi(kv["area"])
			if rep.Names[a] != "" {
				dth++
				led.Observe(survival.ZoneKey{Area: a, Diff: diff}, last, survival.Event{Kind: survival.EvDeath}, pw())
			}
		case "area summary":
			a, _ := strconv.Atoi(kv["area"])
			xp, _ := strconv.Atoi(kv["xp"])
			if rep.Names[a] != "" && xp > 0 {
				led.Observe(survival.ZoneKey{Area: a, Diff: diff}, last, survival.Event{Kind: survival.EvXP, XP: float64(xp)}, pw())
			}
		case "combat fight":
			a, _ := strconv.Atoi(kv["area"])
			k, _ := strconv.ParseFloat(kv["kills"], 64)
			if rep.Names[a] == "" || k <= 0 || hero.MaxLife <= 0 {
				break
			}
			z, _ := o.Tab.Find(a, diff)
			zm, err := o.M.Zone(hero, o.Arch, z, dialID)
			if err != nil || z.Encounters.Typical.N <= 0 {
				break
			}
			dmg, _ := strconv.ParseFloat(kv["damage"], 64)
			sec, _ := strconv.ParseFloat(kv["sec"], 64)
			pot, _ := strconv.ParseFloat(kv["potions"], 64)
			n := float64(z.Encounters.Typical.N)
			led.Observe(survival.ZoneKey{Area: a, Diff: diff}, last, survival.Event{Kind: survival.EvFight, Kills: k, Seconds: sec, Taken: dmg,
				PredTaken: zm.Typical.Terms.Taken / n * k, PredKillSec: zm.Typical.Terms.PerKill * k, Potions: pot}, pw())
		case "planner":
			a, _ := strconv.Atoi(kv["area"])
			if rep.Names[a] != "" {
				if err := flush(); err != nil {
					return nil, err
				}
				pending, pendingArea = &CtlRow{}, a
			}
		case "area skipped":
			a, _ := strconv.Atoi(kv["area"])
			if pending != nil && a == pendingArea {
				pending = nil
			}
		case "minute":
			if v, _ := strconv.Atoi(kv["clvl"]); v > 0 {
				hero.Level = v
			}
			if v, _ := strconv.Atoi(kv["maxLife"]); v > 0 {
				hero.MaxLife = v
				hero.SetQuality(power.TermLife, power.Measured)
			}
			if v, err := strconv.Atoi(kv["potions"]); err == nil {
				hero.Potions = v
			}
			gold, _ = strconv.Atoi(kv["gold"])
			if err := flush(); err != nil {
				return nil, err
			}
			if err := decide("minute"); err != nil {
				return nil, err
			}
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}

	rep.KHero = led.KHero()
	for _, a := range rep.CandAreas {
		rep.KZone[a] = led.KZone(survival.ZoneKey{Area: a, Diff: diff}, last)
	}
	return rep, nil
}

func (r *CtlReport) name(a int) string {
	if n := r.Names[a]; n != "" {
		return fmt.Sprintf("%s(%d)", n, a)
	}
	return strconv.Itoa(a)
}

// Text prints runs of consecutive minute rows with the same (zone, planner, shadow, word) as one line, the pick rows
// as their own lines, and the summary.
func (r *CtlReport) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "== controller replay (dial %s), candidates: %s ==\n", r.Dial, strings.Join(func() []string {
		var s []string
		for _, a := range r.CandAreas {
			s = append(s, r.name(a))
		}
		return s
	}(), ", "))
	fmt.Fprintf(&b, "%-8s %-5s %4s %-20s %-20s %-20s %-9s %-18s %3s %3s %3s %s\n", "time", "kind", "clvl", "in", "planner", "shadow", "word", "why", "chk", "dth", "mrc", "revive")
	var prev *CtlRow
	run := 0
	for i := range r.Rows {
		row := &r.Rows[i]
		same := prev != nil && row.Kind == "minute" && prev.Kind == "minute" && row.Cur == prev.Cur && row.Planner == prev.Planner && row.Shadow == prev.Shadow && row.Word == prev.Word && row.MercDead == prev.MercDead && row.Revive == prev.Revive && row.Chick+row.Deaths+row.MercDied == 0
		if same {
			run++
			continue
		}
		if run > 0 {
			fmt.Fprintf(&b, "         ... same for %d more minutes\n", run)
		}
		run = 0
		fmt.Fprintf(&b, "%-8s %-5s %4d %-20s %-20s %-20s %-9s %-18s %3d %3d %3d %s\n", row.T.Format("15:04:05"), row.Kind, row.Clvl, r.name(row.Cur), r.name(row.Planner), r.name(row.Shadow), row.Word, row.Code, row.Chick, row.Deaths, row.MercDied, row.Revive)
		prev = row
	}
	if run > 0 {
		fmt.Fprintf(&b, "         ... same for %d more minutes\n", run)
	}
	var minutes, agree, hole, picks, pickAgree int
	for _, row := range r.Rows {
		if row.Kind == "pick" {
			picks++
			if row.Planner == row.Shadow {
				pickAgree++
			}
			continue
		}
		minutes++
		if row.Cur == row.Shadow {
			agree++
		}
		if row.Hole {
			hole++
		}
	}
	fmt.Fprintf(&b, "minutes %d: shadow stays where the hero was in %d, holes (empty eligible set) %d; planner picks %d, shadow agrees on %d\n", minutes, agree, hole, picks, pickAgree)
	fmt.Fprintf(&b, "k_hero at the end %.2f; k_zone at the end:", r.KHero)
	for _, a := range r.CandAreas {
		fmt.Fprintf(&b, " %s %.2f;", r.name(a), r.KZone[a])
	}
	b.WriteString("\n")
	return b.String()
}

// TSV is one tab-separated line per decision (the uncollapsed rows), for the calibration report.
func (r *CtlReport) TSV() string {
	var b strings.Builder
	b.WriteString("time\tkind\tclvl\tcur\tplanner\tshadow\tword\twhy\thole\tfresh\tmercdead\trevive\tchick\tdeaths\tmercdied\n")
	for _, x := range r.Rows {
		fmt.Fprintf(&b, "%s\t%s\t%d\t%d\t%d\t%d\t%s\t%s\t%t\t%t\t%t\t%s\t%d\t%d\t%d\n", x.T.Format("15:04:05"), x.Kind, x.Clvl, x.Cur, x.Planner, x.Shadow, x.Word, x.Code, x.Hole, x.Fresh, x.MercDead, x.Revive, x.Chick, x.Deaths, x.MercDied)
	}
	return b.String()
}
