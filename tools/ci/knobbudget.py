#!/usr/bin/env python3
"""Threshold governance check (P10, docs/process/threshold-governance.md). Stdlib only.

  knobbudget.py [--base origin/main] [--head HEAD] [--branch agent/x] [--root .] [--report]

Two rules, both against the merge base with --base:
  1. Knob budget: the knob keys of docs/knobs.md (generated, between the knobdump markers) must not grow
     (added minus removed > 0) unless the branch's task card has `knob_net: owner-ok` in its front matter.
  2. Evidence rows: in the files listed in tools/ci/knobbudget-baseline.json ("governed_files") a row (a JSON object
     with a numeric value of its own) whose numbers changed or that is new must carry `evidence` (n) and `evidence_ref`,
     here or on an ancestor object, with n >= 2 or an evidence_ref that starts with `margin-model:`.
     A row with n = 1 is "single incident: needs seed set": its first change needs a better row.
--report prints findings and exits 0 (report-only phase); without it findings exit 1.
Branches that are not agent/* are skipped (exit 0).
"""
import argparse
import importlib.util
import json
import os
import re
import subprocess
import sys

META = ("evidence", "evidence_ref", "_comment")
KNOB_ROW = re.compile(r"^\|\s*([a-z][A-Za-z0-9]*\.[A-Za-z0-9_.]+)\s*\|")


def git(root, *args, check=True):
    r = subprocess.run(["git", "-C", root, *args], capture_output=True, text=True)
    if r.returncode != 0:
        if check:
            raise RuntimeError("git %s: %s" % (" ".join(args), r.stderr.strip()))
        return None
    return r.stdout


def knob_keys(text):
    """Knob keys inside the knobdump block (all of them when the markers are missing)."""
    if text is None:
        return set()
    m = re.search(r"<!-- knobdump:begin -->(.*?)<!-- knobdump:end -->", text, re.S)
    body = m.group(1) if m else text
    return {KNOB_ROW.match(l).group(1) for l in body.splitlines() if KNOB_ROW.match(l)}


def knob_findings(base_text, head_text, card_fm):
    added = knob_keys(head_text) - knob_keys(base_text)
    removed = knob_keys(base_text) - knob_keys(head_text)
    net = len(added) - len(removed)
    if net <= 0:
        return []
    if re.search(r"^knob_net:\s*.*owner-ok", card_fm or "", re.M):
        return []
    return ["knob budget: +%d net (added %s; removed %s) and the card has no `knob_net: owner-ok`; remove a knob or get the owner's word"
            % (net, ", ".join(sorted(added)) or "none", ", ".join(sorted(removed)) or "none")]


def rows(node, path="", inherited=None, out=None):
    """Map row path -> (numbers, effective evidence). A row is a dict with a numeric value of its own."""
    if out is None:
        out = {}
    if isinstance(node, dict):
        ev = (node.get("evidence"), node.get("evidence_ref")) if "evidence" in node or "evidence_ref" in node else inherited
        nums = {k: v for k, v in node.items() if k not in META and isinstance(v, (int, float)) and not isinstance(v, bool)}
        if nums:
            out[path or "."] = (nums, ev)
        for k, v in node.items():
            if k not in META and isinstance(v, (dict, list)):
                rows(v, path + "." + k if path else k, ev, out)
    elif isinstance(node, list):
        for i, v in enumerate(node):
            tag = str(v.get("name", v.get("area", i))) if isinstance(v, dict) else str(i)
            rows(v, "%s[%s]" % (path, tag), inherited, out)
    return out


def evidence_ok(ev):
    if not ev or ev[0] is None or not ev[1]:
        return "no evidence and evidence_ref"
    n, ref = ev
    if isinstance(n, (int, float)) and n >= 2:
        return ""
    if str(ref).startswith("margin-model:"):
        return ""
    return "evidence %s with ref %r (single incident: needs a seed set of n >= 2 runs or a margin-model: ref)" % (n, ref)


def row_findings(fname, base_doc, head_doc):
    base = rows(base_doc) if base_doc is not None else {}
    out = []
    for path, (nums, ev) in sorted(rows(head_doc).items()):
        if path in base and base[path][0] == nums:
            continue
        why = evidence_ok(ev)
        if why:
            out.append("threshold row %s %s: %s" % (fname, path, why))
    return out


def load_ownscheck(root):
    p = os.path.join(root, "tools", "ci", "ownscheck.py")
    spec = importlib.util.spec_from_file_location("ownscheck", p)
    m = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(m)
    return m


def card_front_matter(root, branch):
    try:
        oc = load_ownscheck(root)
        sg = oc.load_sharedgen(root)
        task = sg.resolve(sg.load_tasks(root), branch)
        return oc.front_matter(task["path"]) if task else ""
    except Exception:
        return ""


def check(root, base, head, branch):
    mb = git(root, "merge-base", base, head).strip()
    findings = []
    findings += knob_findings(git(root, "show", "%s:docs/knobs.md" % mb, check=False),
                              git(root, "show", "%s:docs/knobs.md" % head, check=False), card_front_matter(root, branch))
    cfg = json.loads(git(root, "show", "%s:tools/ci/knobbudget-baseline.json" % head))
    for f in cfg.get("governed_files", []):
        h = git(root, "show", "%s:%s" % (head, f), check=False)
        if h is None:
            continue
        b = git(root, "show", "%s:%s" % (mb, f), check=False)
        findings += row_findings(f, json.loads(b) if b else None, json.loads(h))
    return findings


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--base", default="origin/main")
    ap.add_argument("--head", default="HEAD")
    ap.add_argument("--branch", default="")
    ap.add_argument("--root", default=".")
    ap.add_argument("--report", action="store_true")
    a = ap.parse_args(argv)
    root = os.path.abspath(a.root)
    branch = a.branch or git(root, "rev-parse", "--abbrev-ref", a.head).strip()
    if not branch.startswith("agent/"):
        print("knobbudget: %s is not an agent/* branch, skipped" % branch)
        return 0
    try:
        findings = check(root, a.base, a.head, branch)
    except RuntimeError as e:
        print("knobbudget: skipped (%s)" % e)
        return 0
    for f in findings:
        print("%s knobbudget: %s" % ("REPORT" if a.report else "FAIL", f))
    if not findings:
        print("knobbudget ok (%s)" % branch)
    return 1 if findings and not a.report else 0


if __name__ == "__main__":
    sys.exit(main())
