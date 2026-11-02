"""monsterai2: bucket the first-pass replay mismatches (stderr of test_monsters: `mon_ais AI: <case line>` then `  got ...`) by
(AI, wanted act, got act, fields that differ). usage: python3 mai2_resid.py <stderr extract> [more files]  (stdout: the buckets)"""
import collections
import sys

ACTS = ["idle", "a1", "a2", "s2", "skill", "walk", "run", "approach", "keep", "circle", "wander", "none", "die", "around",
        "home", "point", "untraced", "s3", "s4", "seq", "fly", "gone", "s1", "sc", "mode_at", "move_to", "spawn", "direct"]
buckets = collections.defaultdict(list)
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
        key = (ai, ACTS[want[0]], ACTS[got[0]], ",".join(diff))
        buckets[key].append(scene)
for key, scenes in sorted(buckets.items(), key=lambda kv: (kv[0][0], -len(kv[1]))):
    print(len(scenes), *key)
    v = [int(x) for x in scenes[0].split()]
    print("   sample: ai %d diff %d frame %d target %d dist %d melee %d hit %d life %d ctrl %s aip %s skills %s kin %d/%d corpse %d" % (
        v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7], v[8:11], v[11:19], v[19:24], v[39], v[40], v[41]))
