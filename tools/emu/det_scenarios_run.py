# SPDX-License-Identifier: GPL-3.0-or-later
"""Run and compare the det scenarios of det_scenarios_lab.py (card dethash2).

    # inside one d2d slot (tools/d2dslot.sh run -- ...): every scenario on every build, results.txt
    python3 det_scenarios_run.py run BINDIR OUTDIR SCENARIOS TAGS
    # no slot needed: diff the logs against a base tag
    python3 det_scenarios_run.py compare OUTDIR SCENARIOS BASE TAGS
    # the target's hp / chill / stun per block of a hit scenario
    python3 det_scenarios_run.py table OUTDIR SCENARIO TAG

SCENARIOS and TAGS are comma lists. The binary of tag X is BINDIR/d2d.X; a trailing 2 on the tag (base2) reruns
the binary of the tag without it, the self-check that a build repeats its hashes.
"""

import os
import re
import subprocess
import sys

HERE = os.path.dirname(os.path.abspath(__file__))


def run_all(bindir, outdir, scenarios, tags):
    os.makedirs(outdir, exist_ok=True)
    results = os.path.join(outdir, "results.txt")
    with open(results, "w") as res:
        for scen in scenarios:
            for tag in tags:
                env = dict(os.environ, D2D_BIN=os.path.abspath(os.path.join(bindir, "d2d." + tag.removesuffix("2"))))
                proc = subprocess.run([sys.executable, os.path.join(HERE, "det_scenarios_lab.py"), scen, tag, outdir],
                                      env=env, capture_output=True, text=True, check=False)
                line = proc.stdout.strip().splitlines()[-1] if proc.stdout.strip() else "%s %s FAILED %s" % (scen, tag, proc.stderr.strip()[-200:])
                print(line, flush=True)
                res.write(line + "\n")
                res.flush()


def load(outdir, scen, tag):
    """The log of one run as a list of (label, [lines]) snapshots."""
    snaps, label = [], "-"
    with open(os.path.join(outdir, "det8_%s_%s.txt" % (scen, tag))) as log:
        for line in log:
            line = line.rstrip("\n")
            if line.startswith("SNAP "):
                label = line[5:]
                snaps.append((label, []))
            elif line[:2] in ("M|", "S|", "P|", "G|", "T|") and snaps:
                snaps[-1][1].append(line)
    return snaps


def compare(outdir, scenarios, base, tags):
    for scen in scenarios:
        ref = load(outdir, scen, base)
        for tag in tags:
            cur = load(outdir, scen, tag)
            diffs = [(a[0], [x for x in set(a[1]) ^ set(b[1])]) for a, b in zip(ref, cur, strict=False) if a[1] != b[1]]
            if not diffs and len(ref) == len(cur):
                print("%-14s %-8s same as %s" % (scen, tag, base))
                continue
            kinds = {}
            for _, lines in diffs:
                for text in lines:
                    kinds[text[0]] = kinds.get(text[0], 0) + 1
            first = diffs[0] if diffs else ("(length)", [])
            print("%-14s %-8s DIFFERS in %d of %d snapshots, first %r, lines by section %s" % (scen, tag, len(diffs), len(ref), first[0], kinds))
            for text in sorted(first[1])[:6]:
                print("      " + text[:170])


def table(outdir, scen, tag):
    """hp / chill / stun / mode of the newest monster per snapshot (the hit scenarios' target)."""
    for label, lines in load(outdir, scen, tag):
        rows = [line for line in lines if line.startswith("S|#")]
        who = rows[-1][2:] if rows else "(target dead or gone)"
        who = re.sub(r"\s+at=\S+", "", who)
        player = [line[2:] for line in lines if line.startswith("P|")]
        print("%-34s %s  player %s" % (label, who, player[0].split("\t")[2:4] if player else ""))


def main(argv):
    cmd = argv[1]
    if cmd == "run":
        run_all(argv[2], argv[3], argv[4].split(","), argv[5].split(","))
    elif cmd == "compare":
        compare(argv[2], argv[3].split(","), argv[4], argv[5].split(","))
    elif cmd == "table":
        table(argv[2], argv[3], argv[4])
    else:
        raise SystemExit(__doc__)


if __name__ == "__main__":
    main(sys.argv)
