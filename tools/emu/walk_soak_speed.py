"""Realised displacement speed from a d2d soak log (status x,y every few wall seconds; minute lines = 60 virtual s).
usage: python3 speed.py bot.log [...]"""
import bisect
import json
import math
import sys
from datetime import datetime


def ts(s):
    s = s[:26] + s[-6:] if len(s) > 32 else s
    return datetime.fromisoformat(s).timestamp()


for path in sys.argv[1:]:
    minutes, status, rungs = [], [], 0
    for line in open(path):
        if '"msg":"minute"' in line or '"msg":"status"' in line:
            try:
                j = json.loads(line)
            except ValueError:
                continue
            t = ts(j["time"])
            if j["msg"] == "minute":
                minutes.append(t)
            else:
                status.append((t, j.get("area"), j.get("x"), j.get("y"), j.get("state"), j.get("hostileNear", 0)))
    if len(minutes) < 3:
        continue

    def virt(t):  # virtual seconds at wall time t (60 per minute line, linear between)
        i = bisect.bisect_right(minutes, t) - 1
        i = max(0, min(i, len(minutes) - 2))
        a, b = minutes[i], minutes[i + 1]
        return 60.0 * (i + (t - a) / max(b - a, 1e-6))
    by = {}
    speeds = []
    for p, q in zip(status, status[1:]):
        if p[1] != q[1] or p[2] is None or q[2] is None:
            continue
        dv = virt(q[0]) - virt(p[0])
        if dv <= 0 or dv > 120:
            continue
        d = math.hypot(q[2] - p[2], q[3] - p[3])
        if d > 400:
            continue
        key = p[4] + ("+hostile" if p[5] else "")
        s = by.setdefault(key, [0.0, 0.0, 0])
        s[0] += d
        s[1] += dv
        s[2] += 1
        speeds.append(d / dv)
    speeds.sort()
    q = lambda f: round(speeds[int(f * (len(speeds) - 1))], 2)
    print(path, "virtual min", len(minutes), "pairs", len(speeds), "displacement cells/vs p50", q(.5), "p90", q(.9), "p99", q(.99), "max", q(1))
    tot = [0.0, 0.0]
    for k, (d, t, n) in sorted(by.items(), key=lambda kv: -kv[1][1]):
        tot[0] += d
        tot[1] += t
        print(f"  {k:22s} pairs {n:5d} virt {t / 60:7.1f} min  {d / t:5.2f} cells/vs")
    print(f"  all: {tot[0] / tot[1]:.2f} cells per virtual s")
