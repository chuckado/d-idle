#!/usr/bin/env python3
"""One-off catch-up: mark Done the cards whose branch is merged into the ref and whose acceptance boxes are all ticked.

  python3 -I tools/backlog/catchup_done.py [--repo DIR] [--ref origin/main] [--apply]

Dry run by default: prints one line per candidate and the count. A candidate is a drift.py finding D (a merge subject on the ref names
the task, the task is not Done) whose acceptance block has at least one box and every box ticked; cards with open boxes are listed
as "stays open" and never touched. --apply runs `backlog task edit N -s Done` per candidate (the same call the merge worker's
auto-Done uses; set BACKLOG_CMD to change it) and leaves the changes uncommitted: regenerate the shared files and commit as the
coordinator. Standard library only; why it exists: docs/process/pipeline-audit-2026-11-24.md, finding 2.
"""
import argparse
import os
import re
import shlex
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import drift  # noqa: E402


def candidates(repo, ref):
    """(ready, open_boxes): task dicts, D findings split by whether every acceptance box is ticked."""
    by_id = {t["id"]: t for t in drift.load(repo)}
    ready, held = [], []
    for ln in drift.report(repo, ref):
        m = re.match(r"D (TASK-\d+) ", ln)
        t = by_id.get(m.group(1)) if m else None
        if t is None:
            continue
        (ready if t["total"] and t["ticked"] == t["total"] else held).append(t)
    return ready, held


def main(argv):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--repo", default=".")
    ap.add_argument("--ref", default="origin/main")
    ap.add_argument("--apply", action="store_true")
    a = ap.parse_args(argv)
    ready, held = candidates(a.repo, a.ref)
    for t in ready:
        print("%s %s [%s]: all %d boxes ticked, merged: %s" % (t["id"], t["card"], t["status"], t["total"], "set Done" if a.apply else "would set Done"))
    for t in held:
        print("%s %s [%s]: %d of %d boxes ticked: stays open" % (t["id"], t["card"], t["status"], t["ticked"], t["total"]))
    print("catchup-done: %d to set Done, %d stay open (%s)" % (len(ready), len(held), "applied" if a.apply else "dry run, nothing changed"))
    if not a.apply:
        return 0
    bl = shlex.split(os.environ.get("BACKLOG_CMD", "backlog"))
    bad = 0
    for t in ready:
        n = re.sub(r"\D", "", t["id"])
        r = subprocess.run(bl + ["task", "edit", n, "-s", "Done"], cwd=a.repo, capture_output=True, text=True)
        if r.returncode:
            bad += 1
            print("catchup-done: FAILED %s: %s" % (t["id"], (r.stderr or r.stdout).strip()[-150:]))
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
