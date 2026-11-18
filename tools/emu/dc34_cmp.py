"""decomp34: compare dc34_flee.py output with d2d's rule (ai.cpp:1008-1011: any other monster in DT within ai_distance < 15; ai_distance =
rules/monsters.hpp:468).  usage: dc34_cmp.py flee.out   (a PORT of the rule: the real-build side is the recipe in the trace doc)"""
import sys
def ai_distance(dx, dy):
    dx, dy = abs(dx), abs(dy)
    return (dy + dx * 2) // 2 if dy < dx else (dx + dy * 2) // 2
tot = diff = 0
cls = {}
fell_bad = 0
for line in open(sys.argv[1]):
    if not line.startswith("F "): continue
    head, deaths, tail = line.split(" | ")
    ds = [tuple(int(v) for v in d.split(",")) for d in deaths.split()]
    f = dict(zip(tail.split()[0::2], tail.split()[1::2]))
    away_ok = int(head.split()[4])
    tot += 1
    near = [d for d in ds if d[4] and ai_distance(d[0], d[1]) < 15]
    pred = int(bool(near))
    got = int(f["fled"])
    if pred == got: continue
    diff += 1
    byroom = {}
    for d in ds:
        if d[4]: byroom[d[2]] = byroom.get(d[2], 0) + 1
    why = "ring eviction (room with 5+ deaths)" if any(v > 4 for v in byroom.values()) else "OTHER"
    cls[why] = cls.get(why, 0) + 1
    if why == "OTHER": print("OTHER", line.strip()[:200])
print("cases %d, game fled differs from d2d's rule in %d: %s" % (tot, diff, cls))
