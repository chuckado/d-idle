#!/usr/bin/env python3
"""Rewrite private worker paths in run logs to placeholders (deterministic, in place, stdlib only).

Usage: tools/ci/scrub_paths.py [--check] [ROOT...]     default root: docs/probes
  ~/d2d-lab            -> <lab>
  ~/labrun/<host>      -> <labrun>
  ~/d2lod-ops, ~/d2lod-wt -> <ops>, <wt>
  /home/<name>/        -> <home>/
Only the path text changes; every other byte stays, so measured values and log structure are untouched and files only
shrink. --check rewrites nothing and exits 1 when something would change. The same patterns are cmd/leakcheck's
`labpath` and `homepath` classes (docs/process/gamecode-guard.md).
"""
import os
import re
import sys

RULES = [
    (re.compile(r"~/labrun/[A-Za-z0-9_-]+"), "<labrun>"),
    (re.compile(r"~/labrun\b"), "<labrun>"),
    (re.compile(r"~/d2d-lab\b"), "<lab>"),
    (re.compile(r"~/d2lod-ops\b"), "<ops>"),
    (re.compile(r"~/d2lod-wt\b"), "<wt>"),
    (re.compile(r"/home/[A-Za-z_][\w.-]*/"), "<home>/"),
]
TEXT = {".txt", ".json", ".md", ".log", ".csv", ".jsonl", ".sh", ".tsv"}


def scrub(data):
    n = 0
    for rx, rep in RULES:
        data, k = rx.subn(rep, data)
        n += k
    return data, n


def main(argv):
    check = "--check" in argv
    roots = [a for a in argv if not a.startswith("--")] or ["docs/probes"]
    files = total = 0
    for root in roots:
        for dp, dn, fn in os.walk(root):
            dn.sort()
            for name in sorted(fn):
                if os.path.splitext(name)[1] not in TEXT:
                    continue
                p = os.path.join(dp, name)
                with open(p, "rb") as f:
                    raw = f.read()
                try:
                    old = raw.decode("utf-8")
                except UnicodeDecodeError:
                    continue
                new, n = scrub(old)
                if n == 0:
                    continue
                files += 1
                total += n
                print(f"{p}: {n}")
                if not check:
                    with open(p, "wb") as f:
                        f.write(new.encode("utf-8"))
    print(f"scrub_paths: {total} paths in {files} files" + (" (check, nothing written)" if check else ""))
    return 1 if check and total else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
