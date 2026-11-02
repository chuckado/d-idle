"""monsterai2: bucket the first-pass replay mismatches (stderr of test_monsters: `mon_ais AI: <case line>` then `  got ...`) by
(AI, wanted act, got act, fields that differ). usage: python3 mai2_resid.py <stderr extract> [more files]  (stdout: the buckets)"""
import collections
import sys

ACTS = ["idle", "a1", "a2", "s2", "skill", "walk", "run", "approach", "keep", "circle", "wander", "none", "die", "around",
        "home", "point", "untraced", "s3", "s4", "seq", "fly", "gone", "s1", "sc", "mode_at", "move_to", "spawn", "direct"]
buckets = collections.defaultdict(list)
samples = {}
show = [a[len("--samples="):] for a in sys.argv[1:] if a.startswith("--samples=")]   # --samples=AI: one raw case per residual bucket of that AI
sys.argv = [a for a in sys.argv if not a.startswith("--samples=")]
for path in sys.argv[1:]:
    lines = open(path).read().split("\n")
    for k, line in enumerate(lines):
        if not line.startswith("mon_ais ") or k + 1 >= len(lines) or not lines[k + 1].startswith("  got "):
            continue
        ai = line[8:line.index(":")]
        case = line[line.index(":") + 2:]
        scene, _, _, tail = case.split(" | ")
        want = [int(v) for v in tail.split()]
        got_text = lines[k + 1][6:].replace("|", " ").split()
        got = [int(v) for v in got_text]
        # got: act n x y unit wait event pace steps c0 c1 c2 seed_lo seed_hi ; want: act n x y unit wait event pace steps c0 c1 c2 seed_lo seed_hi
        names = ["act", "n", "x", "y", "unit", "wait", "event", "pace", "steps", "c0", "c1", "c2", "slo", "shi"]
        diff = [names[i] for i in range(min(len(got), len(want), len(names))) if got[i] != want[i]]
        core = [d for d in diff if d not in ("pace",)]       # `pace` alone is the known pace deviation: bucketed apart
        v = [int(x) for x in scene.split()]
        extra = ""
        if ai == "BloodRaven" and v[53]:                       # case-line fields: mon x y 45/46, foe x y 47/48, free spot 0 (home) 53..55
            def gap(dx, dy, less):                                  # FUN_005dc480: each axis less `less` (size / 2 + 1 = 2), not under 0; (short + 2 x long) / 2
                dx, dy = max(abs(dx) - less, 0), max(abs(dy) - less, 0)
                return (dy + dx * 2) // 2 if dx > dy else (dx + dy * 2) // 2
            tdist, sdist = gap(v[47] - v[54], v[48] - v[55], 2), gap(v[45] - v[54], v[46] - v[55], 2)
            extra = "target>=50:%d self>50:%d self>5:%d" % (tdist >= 50, sdist > 50, sdist > 5)
        elif ai == "BloodRaven": extra = "nohome"
        key = (ai, ACTS[want[0]], ACTS[got[0]], ",".join(core) or "pace only", extra)
        buckets[key].append(scene)
        samples.setdefault(key, (scene, want, got))
for key, scenes in sorted(buckets.items(), key=lambda kv: (kv[0][0], -len(kv[1]))):
    print(len(scenes), *key)
    v = [int(x) for x in scenes[0].split()]
    if len(v) < 42: continue
    print("   sample: ai %d diff %d frame %d target %d dist %d melee %d hit %d life %d ctrl %s aip %s skills %s kin %d/%d corpse %d" % (
        v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7], v[8:11], v[11:19], v[19:24], v[39], v[40], v[41]))
for key, (scene, want, got) in samples.items():
    if show and key[0] == show[0] and key[3] != "pace only":
        v = [int(x) for x in scene.split()]
        print("SAMPLE", key[1:], "dist", v[4], "melee", v[5], "hit", v[6], "life", v[7], "ctrl", v[8:11], "mon", v[45:47], "foe", v[47:49], "home", v[54:56], "want", want[:9], "got", got[:9])
summary = collections.defaultdict(lambda: collections.Counter())
for key, scenes in buckets.items():
    ai, _, _, core, extra = key
    kind = "pace only" if core == "pace only" else ("target>=50 (not in d2d)" if "target>=50:1" in extra else "other")
    summary[ai][kind] += len(scenes); summary[ai]["all"] += len(scenes)
for ai, counter in summary.items():
    print("SUMMARY", ai, dict(counter))
