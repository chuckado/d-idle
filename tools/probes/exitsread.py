#!/usr/bin/env python3
"""exitsread: the engine's own link table per level (card exitsread, docs/probes/exitsread-2026-11-22.md).

Per (level, seed): start a lab d2d, `debug goto L`, then read `debug level`, `debug warps` (tile exits: destination, tile
cell, click cell), `debug layout` (placements of the act) and poll `debug nearby` (levels linked past the edge; d2d builds
neighbours on a worker thread, so the list grows after arrival) until it has been unchanged for SETTLE seconds. One JSON line
per run: {level, seed, rect, warps, nearby_polls, nearby, layout_n, settle_s}. The comparison with the bot's exit data is
tools/probes/exitscmp. Usage: exitsread.py D2D DATA LEVEL[,LEVEL..] SEED[,SEED..] [SETTLE]. Stdlib only."""
import json, os, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session

def rows(text):
    return [l.split("\t") if "\t" in l else l.split() for l in text.split("\n") if l and l != "ok" and not l.startswith("err")]

def probe(d2d, data, level, seed, settle):
    p, cmd = session(d2d, data, seed)
    res = {"level": level, "seed": seed}
    try:
        r = cmd("debug goto %d" % level)
        if not r.startswith("ok"):
            res["error"] = "goto: " + r.strip()[:80]
            return res
        t0 = time.time()
        lv = cmd("debug level").split()
        res["rect"] = lv[1:]
        res["warps"] = [{"dest": int(f[3]), "cell": [float(f[1]), float(f[2])], "unit": [float(f[4]), float(f[5])], "click": [float(f[6]), float(f[7])]} for f in rows(cmd("debug warps")) if len(f) >= 8]
        res["layout_n"] = len(rows(cmd("debug layout")))
        last, stable_since, polls = None, time.time(), 0
        while time.time() - t0 < 60:
            near = sorted(tuple(int(v) for v in f[:3]) for f in rows(cmd("debug nearby")) if len(f) >= 3)
            polls += 1
            if near != last:
                last, stable_since = near, time.time()
            if time.time() - stable_since >= settle:
                break
            time.sleep(1.0)
        res["nearby"] = [{"level": a, "dx": b, "dy": c} for a, b, c in (last or [])]
        res["nearby_polls"], res["settle_s"] = polls, round(stable_since - t0, 1)
    finally:
        p.terminate()
        try:
            p.wait(10)
        except Exception:
            p.kill()
    return res

if __name__ == "__main__":
    d2d, data = sys.argv[1], sys.argv[2]
    levels = [int(x) for x in sys.argv[3].split(",")]
    seeds = [int(x) for x in sys.argv[4].split(",")]
    settle = float(sys.argv[5]) if len(sys.argv) > 5 else 6
    for lv in levels:
        for sd in seeds:
            try:
                r = probe(d2d, data, lv, sd, settle)
            except Exception as e:
                r = {"level": lv, "seed": sd, "error": repr(e)[:120]}
            print(json.dumps(r), flush=True)
