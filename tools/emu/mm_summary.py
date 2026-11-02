#!/usr/bin/env python3
"""MONSTERMISSILE: turn mm_launch.py's output into the markdown table of the trace doc.

    python3 mm_summary.py launch.txt TABLES_DIR

One line per (monster family, missile row): game.exe's physical range and element ranges in points per difficulty (V run, the
missile's own stats right after FUN_0059fa30), the row's ToHit, blockable (physical max > 0), and the reach in subtiles
(Range x Vel x 0.046875: path velocity ((Vel << 8) x 75 / 100) / 4096 per frame, VelLev x level / 8 added to Vel at the level; D + S).
"""
import ast
import csv
import os
import sys


def rows(d, name):
    r = list(csv.reader(open(os.path.join(d, name), encoding="latin-1"), delimiter="\t"))
    return [dict(zip(r[0], x)) for x in r[1:] if x and x[0] and x[0] != "Expansion"]


def num(s):
    try: return int(s)
    except ValueError: return 0


def pts(lo, hi):
    if hi == 0: return "-"
    f = lambda v: (f"{v / 256:.1f}".rstrip("0").rstrip(".") if v >= 26 or v == 0 else "<0.1")
    return f"{f(lo)}-{f(hi)}"


def main():
    out, d = {}, sys.argv[2]
    mis = {m["Missile"]: m for m in rows(d, "missiles.txt")}
    mons = {m["Id"]: m for m in rows(d, "monstats.txt")}

    def chance(mid, mode):
        m = mons[mid]
        got = []
        for k in (1, 2, 3):
            if m.get(f"El{k}Mode") == mode:
                ps = [num(m.get(f"El{k}Pct{s_}", "")) for s_ in ("", "(N)", "(H)")]
                if any(0 < x < 100 for x in ps): got.append("/".join(str(x) for x in ps))
        return got
    for line in open(sys.argv[1]):
        f = line.rstrip("\n").split("\t")
        if len(f) < 24 or f[5] != "own": continue
        diff, mid, mode, src, mn = int(f[0]), f[1], f[2], f[3], f[4]
        ph = (int(f[12]), int(f[13]))
        el = ast.literal_eval(f[-1])
        out.setdefault((mn, src, mid), {})[diff] = (ph, dict((int(a, 16), b) for a, b in el))
    modes = {}
    for line in open(sys.argv[1]):
        f = line.split("\t")
        if len(f) > 4 and f[5] == "own": modes[(f[4], f[3], f[1])] = f[2]
    names = {0x30: "fire", 0x32: "ltng", 0x34: "mag", 0x36: "cold"}
    seen = set()
    print("| row (id) | shooters (source) | physical pts N / NM / H | element pts N / NM / H | poison total pts (per-frame x len / 256) N / NM / H | ToHit | blockable | reach |")
    print("|---|---|---|---|---|---|---|---|")
    for (mn, src, mid), v in sorted(out.items(), key=lambda kv: (int(mis[kv[0][0]]["Id"]), kv[0][2])):
        key = (mn, tuple(sorted((k, d_[0], tuple(sorted(d_[1].items()))) for k, d_ in v.items())))
        if key in seen: continue
        seen.add(key)
        r = mis[mn]
        phs = " / ".join(pts(*v[k][0]) for k in (0, 1, 2) if k in v)
        els, pois = [], []
        for k in (0, 1, 2):
            if k not in v: continue
            e = v[k][1]
            parts = [f"{n} {pts(e.get(s, 0), e[s + 1])}" for s, n in names.items() if e.get(s + 1)]
            els.append(", ".join(parts) or "-")
            pois.append(f"{e[0x39] * e.get(0x3b, 0) / 256:.1f}" if e.get(0x3a) else "-")
        vel = num(r["Vel"]); rng = num(r["Range"])
        fam = sorted({m for (n_, s_, m) in out if n_ == mn})
        ch = chance(mid, modes.get((mn, src, mid), "")) if src.startswith("Miss") else []
        shooters = ", ".join(f"{m} ({s_})" for (n_, s_, m) in sorted(out) if n_ == mn and (n_, s_, m) == (mn, src, mid)) + (f" [El chance {ch[0]} %: element is a per-hit chance]" if ch else "") + (" [weapon: random]" if mid == "bloodraven" else "")
        block = "yes" if any(v[k][0][1] > 0 for k in v) else "no"
        print(f"| {mn} ({r['Id']}) | {shooters} | {phs} | {' / '.join(els)} | {' / '.join(pois)} | {'1' if r['ToHit'] else '-'} | {block} | {rng * vel * 0.046875:.1f} (vel {vel}, range {rng}) |")


if __name__ == "__main__": main()
