"""monsterskill: classify the replay mismatches of msk_run.sh (game.exe act vs d2d act), from $D2_SCRATCH/run/stderr.txt.
Each mismatch is two lines: `mon_ais AI: <scene...> | | | act n x y unit wait event | pace steps | ctrl x3 | seed` (game.exe) and
`  got act n x y unit wait event | pace steps | ctrl x3 | seed` (d2d). Run: tools/emu/run-on-worker.sh msk_cls.py"""
import collections
import os

ACT = "idle a1 a2 s2 skill walk run approach keep circle wander none die around home point untraced s3 s4 seq fly gone s1 sc mode_at move_to spawn direct warp".split()
p = os.path.join(os.environ.get("D2_SCRATCH", "/tmp/d2agent-monsterskill"), "run", "stderr.txt")
L = open(p).read().split("\n")
cnt = collections.Counter()
ex = {}
i = 0
while i < len(L) - 1:
    if L[i].startswith("mon_ais") and L[i + 1].startswith("  got"):
        g = L[i].split("|")
        want = g[3].split()
        got = L[i + 1][6:].split("|")
        gw = got[0].split()
        key = (ACT[int(want[0])], ACT[int(gw[0])], "pace %s/%s" % (want[7], got[1].split()[0]),
               "fields " + ",".join(n for n, a, b in zip("act n x y unit wait event".split(), want, gw) if a != b))
        cnt[key] += 1
        ex.setdefault(key, (L[i][:260], L[i + 1]))
        i += 2
    else:
        i += 1
for k, v in cnt.most_common():
    print(v, k)
print("total", sum(cnt.values()))
