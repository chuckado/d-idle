"""Back-off step cadence from a d2d soak log: consecutive 'out of mana: backing off' steps (to X,Y, step n -> n+1),
virtual seconds between them (minute lines = 60 virtual s) and the distance between their targets (subtiles)."""
import bisect
import json
import math
import sys
from datetime import datetime


def ts(s):
    s = s[:26] + s[-6:] if len(s) > 32 else s
    return datetime.fromisoformat(s).timestamp()


minutes, steps = [], []
for line in open(sys.argv[1]):
    if '"msg":"minute"' in line:
        minutes.append(ts(json.loads(line)["time"]))
    elif "out of mana: backing off" in line and "\"step\"" in line:
        j = json.loads(line)
        steps.append((ts(j["time"]), j["step"], j["to"]["X"], j["to"]["Y"]))


def virt(t):
    i = max(0, min(bisect.bisect_right(minutes, t) - 1, len(minutes) - 2))
    a, b = minutes[i], minutes[i + 1]
    return 60.0 * (i + (t - a) / max(b - a, 1e-6))


dts, ds = [], []
for p, q in zip(steps, steps[1:]):
    if q[1] == p[1] + 1:
        dts.append(virt(q[0]) - virt(p[0]))
        ds.append(math.hypot(q[2] - p[2], q[3] - p[3]))
dts.sort()
ds.sort()
q = lambda a, f: round(a[int(f * (len(a) - 1))], 2)
print("steps", len(steps), "pairs", len(dts))
print("virtual s between steps p10/p50/p90", q(dts, .1), q(dts, .5), q(dts, .9), "sum min", round(sum(dts) / 60, 1))
print("subtiles between targets p10/p50/p90", q(ds, .1), q(ds, .5), q(ds, .9))
print("subtiles per virtual s (sum/sum)", round(sum(ds) / sum(dts), 2))
