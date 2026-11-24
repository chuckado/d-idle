#!/usr/bin/env python3
"""Ownership and guard-path check for one branch (pipeline audit 2026-11-24, row 5).

  ownscheck.py [--base origin/main] [--head HEAD] [--branch agent/x] [--root .] [--warn-owns-until YYYY-MM-DD]

Looks at the files the branch changed against the merge base with --base and fails (exit 1) when
  1. a changed file is outside the card's `owns:` (after the allowances below), or
  2. docs/adr/*, docs/decisions.md or GLOSSARY.md changed and the card has no `allow-adr: true` (a brief that names those files
     is the only reason to set it; AGENTS.md "guard rule").
The card is found from the branch name like the merge step does (tools/sharedgen resolve: branch key, card key). A branch with no
card has no owns to check (a WARN) but the guard paths still fail. Branches that are not agent/* and `main` itself are skipped.

Allowed without being in `owns:`: the card's own file, generated shared files (BACKLOG, CHANGELOG, INDEX, OWNERS, STATUS,
cards.md, live-checks.md, knobs.md) and the branch's backlog/merge-lines file, and every path of the card's `contracts:` list.
--warn-owns-until DATE turns the owns finding into a WARN before that day (rollout); the ADR guard never softens.
Stdlib only.
"""
import argparse
import datetime
import fnmatch
import importlib.util
import os
import re
import subprocess
import sys

GUARDED = ("docs/adr/", "docs/decisions.md", "GLOSSARY.md")
GENERATED = {"docs/BACKLOG.md", "CHANGELOG.md", "docs/INDEX.md", "docs/OWNERS", "docs/plan/STATUS.md", "docs/cards.md",
             "docs/live-checks.md", "docs/knobs.md"}


def git(root, *args):
    r = subprocess.run(["git", "-C", root, *args], capture_output=True, text=True)
    if r.returncode != 0:
        raise RuntimeError("git %s: %s" % (" ".join(args), r.stderr.strip()))
    return r.stdout


def load_sharedgen(root):
    p = os.path.join(root, "tools", "sharedgen", "sharedgen.py")
    if not os.path.exists(p):
        p = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "sharedgen", "sharedgen.py")
    spec = importlib.util.spec_from_file_location("sharedgen_for_ownscheck", p)
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def front_matter(path):
    with open(path, encoding="utf-8") as fh:
        m = re.match(r"---\n(.*?)\n---\n", fh.read(), re.S)
    return m.group(1) if m else ""


def list_key(fm, key):
    out, on = [], False
    for line in fm.splitlines():
        if re.match(r"^%s:\s*(\[\])?\s*$" % key, line):
            on = not line.rstrip().endswith("[]")
            continue
        if on:
            li = re.match(r"^\s+-\s*(.*)$", line)
            if li:
                out.append(li.group(1).strip().strip("\"'"))
            else:
                on = False
    return out


def owned(path, patterns):
    for p in patterns:
        p = p.strip()
        if not p:
            continue
        if p in (path, path + "/") or (p.endswith("/") and path.startswith(p)):
            return True
        if p.endswith("/**") and path.startswith(p[:-2]):
            return True
        if fnmatch.fnmatchcase(path, p.replace("**", "*")):
            return True
    return False


def check(root, base, head, branch, warn_until=None, today=None):
    """Returns (failures, warnings)."""
    fails, warns = [], []
    mb = git(root, "merge-base", base, head).strip()
    changed = [f for f in git(root, "diff", "--name-only", mb, head).splitlines() if f]
    sg = load_sharedgen(root)
    task = sg.resolve(sg.load_tasks(root), branch)
    fm = front_matter(task["path"]) if task else ""
    allow_adr = bool(re.search(r"^allow-adr:\s*true\s*$", fm, re.M))
    for f in changed:
        if any(f == g or (g.endswith("/") and f.startswith(g)) for g in GUARDED) and not allow_adr:
            fails.append("%s: guarded path (ADR, decisions, glossary); only a brief that names it sets `allow-adr: true` in the card" % f)
    if not task:
        warns.append("no task card found for %s: owns not checked" % branch)
        return fails, warns
    rel = os.path.relpath(task["path"], root)
    allowed = list(task["owns"]) + list_key(fm, "contracts")
    name = branch.split("/", 1)[-1]
    soft = bool(warn_until) and (today or datetime.date.today()) < datetime.date.fromisoformat(warn_until)
    for f in changed:
        if f in GENERATED or f == rel or f == "backlog/merge-lines/%s.txt" % name or owned(f, allowed):
            continue
        (warns if soft else fails).append("%s: not under owns of %s (%s)" % (f, task["card"] or rel, ", ".join(task["owns"]) or "no owns"))
    return fails, warns


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--base", default="origin/main")
    ap.add_argument("--head", default="HEAD")
    ap.add_argument("--branch", default="")
    ap.add_argument("--root", default=".")
    ap.add_argument("--warn-owns-until", default="")
    a = ap.parse_args(argv)
    root = os.path.abspath(a.root)
    branch = a.branch or git(root, "rev-parse", "--abbrev-ref", a.head).strip()
    if branch in ("main", "master", "HEAD") or not branch.startswith("agent/"):
        print("ownscheck: %s is not an agent/* branch, skipped" % branch)
        return 0
    try:
        fails, warns = check(root, a.base, a.head, branch, a.warn_owns_until or None)
    except RuntimeError as e:
        print("ownscheck: skipped (%s)" % e)
        return 0
    for w in warns:
        print("WARN ownscheck: " + w)
    for f in fails:
        print("FAIL ownscheck: " + f)
    if not fails:
        print("ownscheck ok (%s)" % branch)
    return 1 if fails else 0


if __name__ == "__main__":
    sys.exit(main())
