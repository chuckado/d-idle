#!/usr/bin/env python3
"""knobsummary: one row per run of tools/provingground/knobs3way.sh (card mapverify), from results.json and the bot log.
Usage: knobsummary.py provingground-out/k3-<stamp> [GOAL_MONSTER]    (default goal: The Countess kill)
Columns: condition, class, seed, status, wall s, goal wall s (first log line to the goal kill), goal game s (the last
nav.trace click t before the kill, ms of the bot clock), clicks to the goal (nav.trace click lines: the step count),
and, counted up to the goal kill only: mouth-gate lines (click held / rewritten / sent inside the bound), area gate
refusals, nav stuck, deaths ("character died"); then over the whole run: "nav no usable gate" and deaths. Stdlib only."""
import glob, json, os, re, sys
from datetime import datetime

goal = sys.argv[2] if len(sys.argv) > 2 else "The Countess"

def ts(line):
    m = re.match(r'\{"time":"([^"]+)"', line)
    if not m:
        return None
    t = m.group(1)
    t = re.sub(r"(\.\d{6})\d*", r"\1", t)
    return datetime.fromisoformat(t).timestamp()

print("cond\tclass\tseed\tstatus\twall_s\tgoal_wall_s\tgoal_game_s\tclicks_to_goal\tmouth_gate\tarea_gate_refused\tnav_stuck\tdeaths_to_goal\tno_usable_gate_all\tdeaths_all")
for d in sorted(glob.glob(os.path.join(sys.argv[1], "*-*-*"))):
    cond, cls, seed = os.path.basename(d).split("-")
    res = glob.glob(d + "/results.json")
    st, wall = "NO-RESULT", ""
    if res:
        r = json.load(open(res[0]))["results"][0]
        st, wall = r["status"], r.get("wall_sec", "")
    logs = glob.glob(d + "/*.bot.log")
    if not logs:
        print("\t".join([cond, cls, seed, st, str(wall)])); continue
    t0 = tg = None
    clicks = mg = ag = stuck = deaths = nug = dall = 0
    gt = None
    lastt = 0
    for line in open(logs[0], errors="replace"):
        if '"msg":' not in line:
            continue
        t = ts(line)
        if t0 is None:
            t0 = t
        msg = re.search(r'"msg":"([^"]*)"', line).group(1)
        reached = tg is not None
        if msg == "nav.trace click":
            m = re.search(r'"t":(\d+)', line)
            if m and not reached:
                lastt = int(m.group(1)); clicks += 1
        elif msg == "nav no usable gate":
            nug += 1
        if msg == "character died":
            dall += 1
        if not reached:
            if msg.startswith("nav mouth gate"):
                mg += 1
            elif msg == "area gate refused the walk":
                ag += 1
            elif msg == "nav stuck":
                stuck += 1
            elif msg == "character died":
                deaths += 1
        if msg == "combat kill" and not reached and ('"monster":"%s"' % goal) in line:
            tg, gt = t, lastt
    gw = "%.0f" % (tg - t0) if tg and t0 else "-"
    gg = "%.0f" % (gt / 1000) if tg else "-"
    print("\t".join(str(x) for x in [cond, cls, seed, st, wall, gw, gg, clicks if tg else "-", mg, ag, stuck, deaths, nug, dall]))
