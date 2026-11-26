// auraprobe measures what the Paladin aura state is on d2d after each way the bot (or a player) touches it (card
// auralive, docs/probes/paladin-aura-live-2026-11-26.md). Run it inside tools/d2dslot.sh on the worker host with the lab
// binary (D2D_BIN, lab verbs). Reads only after the setup: the aura is read from the `state` line (`aura=` is the
// player's active aura skill id, `lskill=`/`rskill=` the two hands), `debug states` (buffs) and the area. The setup
// (clvl, skill points) goes through the cheat ledger BEFORE the first reading; the one gameplay-moving verb after it is
// `debug goto` (the area change under test), declared in the doc.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"d2lod-bot/internal/cheat"
	"d2lod-bot/internal/combat"
	"d2lod-bot/internal/d2dcombat"
	"d2lod-bot/internal/d2dctl"
	"d2lod-bot/internal/d2denv"
	"d2lod-bot/internal/d2dinput"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/env"
)

const (
	skMight, skHolyFire, skSacrifice, skZeal = 98, 102, 107, 106
)

type probe struct {
	be   *d2denv.Backend
	led  *cheat.Ledger
	char string
	fi   *combat.Fighter
	ctx  context.Context
	t0   time.Time
}

var kv = regexp.MustCompile(`\b(aura|lskill|rskill|level|life|mana|screen)=(\S+)`)

func (p *probe) say(f string, a ...any) {
	fmt.Printf("%s ", time.Now().Format("15:04:05.000"))
	fmt.Printf(f+"\n", a...)
}

// read prints one reading: state line fields, buffs, area.
func (p *probe) read(label string) {
	r, err := p.be.Client.Do(p.ctx, "state")
	fields := map[string]string{}
	if err == nil {
		for _, m := range kv.FindAllStringSubmatch(strings.Join(append(r.Lines, r.Rest), " "), -1) {
			fields[m[1]] = m[2]
		}
	}
	b, berr := p.be.Client.Do(p.ctx, "debug states")
	buffs := ""
	if berr == nil {
		buffs = strings.TrimSpace(strings.Join(append(b.Lines, b.Rest), " "))
	}
	dmg := "?"
	if d, derr := p.be.Client.Do(p.ctx, "dump stats"); derr == nil {
		if m := regexp.MustCompile(`"dmgpct":(-?\d+)`).FindStringSubmatch(strings.Join(append(d.Lines, d.Rest), " ")); m != nil {
			dmg = m[1]
		}
	}
	area := -1
	if s, e := p.be.Env.Snapshot(); e == nil && s.Player != nil {
		area = s.Player.Area
	}
	p.say("READ %-34s aura=%s lskill=%s rskill=%s area=%d dmgpct=%s | %s", label, fields["aura"], fields["lskill"], fields["rskill"], area, dmg, buffs)
}

func (p *probe) monsters(label string) {
	r, err := p.be.Client.Do(p.ctx, "monsters")
	out := ""
	if r != nil {
		out = strings.Join(append(r.Lines, r.Rest), " | ")
	}
	if i := strings.Index(out, "#1\t"); i > 0 { // keep only monster #1's row
		out = out[max(0, strings.LastIndex(out[:i], "|")+1) : i+4]
	}
	p.say("MON  %-34s err=%v %s", label, err, out)
}

func (p *probe) step(n int) {
	_, err := p.be.Client.Step(p.ctx, n)
	if err != nil {
		p.say("step err %v", err)
	}
}

func (p *probe) cheat(verb, reason string) {
	e := cheat.Entry{Time: time.Now().UTC(), Actor: "auralive", Instance: "slot" + os.Getenv("D2D_SLOT"), Character: p.char, Kind: "d2d-debug",
		Params: map[string]any{"verb": verb}, Reason: reason}
	if err := p.led.Append(e); err != nil {
		p.say("ledger refused: %v", err)
		os.Exit(1)
	}
	r, err := p.be.Client.Do(p.ctx, verb)
	out := ""
	if r != nil {
		out = strings.TrimSpace(strings.Join(append(r.Lines, r.Rest), " "))
	}
	e.Time, e.Success, e.After = time.Now().UTC(), err == nil, map[string]any{"reply": out}
	_ = p.led.Append(e)
	p.say("CHEAT %q -> %s err=%v", verb, out, err)
}

func (p *probe) cell() (x, y float64) {
	r, err := p.be.Client.Do(p.ctx, "state")
	if err != nil {
		return 20, 20
	}
	m := regexp.MustCompile(`cam=([\d.]+),([\d.]+)`).FindStringSubmatch(strings.Join(append(r.Lines, r.Rest), " "))
	if m == nil {
		return 20, 20
	}
	fmt.Sscanf(m[1], "%f", &x)
	fmt.Sscanf(m[2], "%f", &y)
	return
}

func (p *probe) setRight(id int, name string) {
	err := p.fi.SetRight(p.ctx, id)
	p.say("SetRight(%s %d) popup pick err=%v", name, id, err)
	p.step(10)
}

func main() { os.Exit(run()) }

func run() int {
	home, _ := os.UserHomeDir()
	slotDir := os.Getenv("D2D_SLOT_DIR")
	bin := flag.String("bin", os.Getenv("D2D_BIN"), "lab d2d binary")
	gamedir := flag.String("gamedir", envOr("D2D_DATA", filepath.Join(home, "d2d-lab/data")), "game data")
	cache := flag.String("cache", filepath.Join(home, "d2d-lab/drills/tables"), "tables cache")
	arm := flag.String("arm", "pick", "pick | cast | cmdskill | left")
	seed := flag.Int64("seed", 11, "seed")
	name := flag.String("name", "Auri", "throwaway name")
	flag.Parse()
	if *bin == "" || slotDir == "" {
		fmt.Fprintln(os.Stderr, "auraprobe: run inside tools/d2dslot.sh with D2D_BIN")
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	db, err := data.Open(*gamedir, *cache)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tables:", err)
		return 1
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	tune := env.DefaultTuning()
	hdir := filepath.Join(slotDir, "aurahome")
	_ = os.MkdirAll(hdir, 0o700)
	be, err := d2denv.Build(ctx, d2denv.Config{ID: "auralive", Log: log, Tune: &tune, Data: db, Sock: os.Getenv("D2D_SOCK"), LinkWait: 30 * time.Second,
		Bin: *bin, DataDir: *gamedir, Home: hdir, Nice: 10, LogFile: filepath.Join(hdir, "d2d.out"),
		ExtraArgs: []string{"--toggle", "autoloot=off,autoparty=off", "--lockstep", "--lab-verbs"},
		Spec:      d2dinput.Spec{Name: *name, Class: 2, Seed: *seed, NoSave: true}})
	if err != nil {
		fmt.Fprintln(os.Stderr, "build:", err)
		return 1
	}
	defer be.Close()
	if err := be.Start(ctx, 0); err != nil {
		fmt.Fprintln(os.Stderr, "start:", err)
		return 1
	}
	lp := envOr("D2CHEAT_LEDGER", filepath.Join(home, ".local/share/d2bot/ledger/cheats.jsonl"))
	led := cheat.NewLedger(lp)
	if err := led.Ensure(); err != nil {
		fmt.Fprintln(os.Stderr, "ledger:", err)
		return 1
	}
	d2dctl.Pause = func(ctx context.Context, d time.Duration) error {
		_, err := be.Client.Step(ctx, max(int(d/(40*time.Millisecond)), 1))
		return err
	}
	rig := d2dcombat.Attach(ctx, be.Env, be.Navigator(), be.Ctl)
	rows, err := combat.LoadListRows(*gamedir, db)
	if err != nil {
		fmt.Fprintln(os.Stderr, "list rows:", err)
		return 1
	}
	rig.Fighter.SetListRows(rows)
	p := &probe{be: be, led: led, char: *name, fi: rig.Fighter, ctx: ctx, t0: time.Now()}
	p.say("ARM %s bin=%s", *arm, filepath.Base(*bin))

	// SETUP (before the first reading): level 6 and one hard point each in Might, Holy Fire, Sacrifice, Zeal.
	const why = "auralive SETUP: Paladin that owns the skills under test, before any reading (docs/probes/paladin-aura-live-2026-11-26.md)"
	p.cheat("debug stat 12 6", why)
	for _, id := range []int{skMight, skHolyFire, skSacrifice, skZeal} {
		p.cheat(fmt.Sprintf("debug points %d 1", id), why)
	}
	p.step(10)

	p.read("A baseline")
	x, y := p.cell()
	p.say("player cell %.1f,%.1f", x, y)

	switch *arm {
	case "pick": // A, B, F, D
		p.setRight(skMight, "Might")
		p.read("B after popup pick Might")
		p.step(100)
		p.read("B +100 ticks")
		for i := 1; i <= 3; i++ { // F: the bot's attack path
			_ = be.Client.CmdSkill(ctx, skSacrifice, x+3, y, d2dctl.SkillOpts{})
			p.step(30)
		}
		p.read("F after 3x cmd skill Sacrifice")
		p.setRight(skHolyFire, "Holy Fire")
		p.read("B2 pick Holy Fire (switch)")
		p.setRight(skMight, "Might")
		p.read("B3 pick Might again")
		_ = be.Client.CmdSkill(ctx, skZeal, x+3, y, d2dctl.SkillOpts{Left: true})
		p.step(30)
		p.read("D after cmd skill Zeal ... left")
		p.cheat("debug goto 2", "auralive DIAGNOSIS: the area change under test (Blood Moor), no other effect; mechanics only")
		p.step(20)
		p.read("E after area change (goto 2)")
		p.step(200)
		p.read("E +200 ticks")
		p.cheat("debug goto 1", "auralive DIAGNOSIS: back to the camp, area change under test")
		p.step(20)
		p.read("E2 after goto 1 (town)")
	case "cast": // C: a right click
		p.setRight(skMight, "Might")
		p.read("B after popup pick Might")
		if err := be.Client.RClick(ctx, 400, 250); err != nil {
			p.say("rclick err %v", err)
		}
		p.step(40)
		p.read("C after rclick ground")
		p.step(100)
		p.read("C +100 ticks")
	case "cmdskill": // the aura id through cmd skill without any pick
		_ = be.Client.CmdSkill(ctx, skMight, x+3, y, d2dctl.SkillOpts{})
		p.step(40)
		p.read("G after cmd skill Might (no pick)")
		_ = be.Client.CmdSkill(ctx, skMight, x+3, y, d2dctl.SkillOpts{Left: true})
		p.step(40)
		p.read("G2 after cmd skill Might left")
	case "left": // D: pick, then a different right skill
		p.setRight(skMight, "Might")
		p.read("B after popup pick Might")
		p.setRight(skSacrifice, "Sacrifice as right (non-aura)")
		p.read("D after right skill Sacrifice")
	case "holyfire": // effect check: a monster beside the hero loses life to the picked Holy Fire with no cast
		p.cheat("debug goto 2", "auralive SETUP: field level for the monster, before any reading of this arm")
		p.step(20)
		p.cheat("debug points 102 20", "auralive SETUP: Holy Fire level 21 so a pulse is visible in a 400 life target and the radius is wide")
		p.cheat("debug spawn zombie1 1 near", "auralive SETUP: one monster beside the hero to see aura damage")
		p.cheat("debug mon set 1 life 400 maxlife 400 dmg 0 0", "auralive SETUP: a target that survives the whole window and does no damage")
		p.cheat("debug warp 38 58", "auralive SETUP: the hero next to monster #1, before any reading of this arm")
		p.step(5)
		p.monsters("control, no aura")
		p.step(100)
		p.monsters("control +100 ticks, no aura")
		p.setRight(skHolyFire, "Holy Fire")
		p.read("B after popup pick Holy Fire")
		p.step(100)
		p.monsters("after pick +100 ticks, no cast")
		p.step(100)
		p.monsters("after pick +200 ticks, no cast")
		p.setRight(skSacrifice, "Sacrifice as right (non-aura)")
		p.read("D aura dropped by a non-aura right skill")
		p.monsters("D start")
		p.step(100)
		p.monsters("D +100 ticks (aura off)")
	default:
		p.say("unknown arm")
		return 2
	}
	p.say("DONE writes after the first reading: debug goto only (arm pick)")
	return 0
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
