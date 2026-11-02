#!/usr/bin/env python3
"""Maintainer claims on board items. Stdlib only; needs `gh` for --send. Rule: docs/process/chunker.md, section "Claim before launch".

  claim.py claim ITEM-ID --job NAME --note TEXT [--send]
      post "Claimed by the maintainers: ..." on the item's issue (the comment IS the claim: the board defines no claim label)
  claim.py done  ITEM-ID --doc PATH --note TEXT [--close] [--send]
      post the result with the deliverable path, optionally close the issue as completed

  claim.py watch [--update]
      list work-item issues whose comment count or last-comment time changed since the last run (state: data-cache/claims-watch.json,
      gitignored; --update saves it) and classify the newest comment: CLAIM, SUBMISSION, QUESTION, WITHDRAWAL or OTHER. Read only.

Dry run is the default (prints what would be posted). The id maps to the issue through docs/community/issues.tsv. Refused when the
issue is closed (claim), is assigned, or already carries a human's `claiming` comment: a human's claim always wins and is never
answered here. Posted text passes the repo privacy scan. Every sent action is appended to docs/community/claims.tsv.
Only issues labelled `work-item` are touched; pull requests and upstream repos never.
"""
import argparse
import datetime
import json
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.dirname(os.path.dirname(HERE))
sys.path.insert(0, HERE)
import chunker  # noqa: E402

TSV = "docs/community/issues.tsv"
LOG = "docs/community/claims.tsv"
CLAIM_RE = re.compile(r"(?im)^\s*claiming\b")
MARK = "<!-- claim:%s -->"
ID_RE = re.compile(r"[A-Z][A-Z0-9-]*")


class Refused(Exception):
    pass


def issue_number(item_id, root=ROOT):
    with open(os.path.join(root, TSV), encoding="utf-8") as f:
        for ln in f:
            c = ln.rstrip("\n").split("\t")
            if len(c) == 2 and c[0] == item_id:
                return int(c[1].rsplit("/", 1)[1])
    raise Refused("%s has no issue in %s" % (item_id, TSV))


def check_issue(item_id, iss, action):
    if chunker.LABEL not in [l["name"] for l in iss.get("labels") or []]:
        raise Refused("issue is not labelled %s" % chunker.LABEL)
    if action == "claim" and iss.get("state") != "OPEN":
        raise Refused("issue is closed")
    if iss.get("assignees"):
        raise Refused("issue is assigned to someone")
    for c in iss.get("comments") or []:
        body = c.get("body") or ""
        if "<!-- claim:" not in body and CLAIM_RE.search(body):
            raise Refused("a human already commented `claiming` (%s)" % ((c.get("author") or {}).get("login") or "?"))


def already(item_id, iss, kind):
    tag = MARK % ("%s:%s" % (kind, item_id))
    return any(tag in (c.get("body") or "") for c in iss.get("comments") or [])


def clean(s, limit=300):
    return " ".join("".join(c for c in s if c.isprintable()).replace("@", "").replace("<", "").replace(">", "").split())[:limit]


def build(kind, item_id, a):
    if kind == "claim":
        txt = "Claimed by the maintainers: %s. %s Will update here when it lands or if we release it." % (clean(a.job, 80), clean(a.note))
    else:
        if not os.path.exists(os.path.join(ROOT, a.doc)):
            raise Refused("doc %s does not exist in the repository" % a.doc)
        txt = "Answered by the maintainers' own work: %s. %s%s" % (clean(a.doc, 120), clean(a.note), " Closing as completed." if a.close else "")
    txt = txt.strip() + "\n" + MARK % ("%s:%s" % (kind, item_id))
    try:
        return chunker.safe_text(txt)
    except ValueError as e:
        raise Refused(str(e))


def log_row(kind, item_id, num, detail, now):
    p = os.path.join(ROOT, LOG)
    new = not os.path.exists(p)
    with open(p, "a", encoding="utf-8") as f:
        if new:
            f.write("date\titem\tissue\taction\tdetail\n")
        f.write("%s\t%s\t%d\t%s\t%s\n" % (now, item_id, num, kind, clean(detail, 120)))


def run(a, runner=chunker.run_gh, now=None):
    kind, item_id = a.cmd, a.item
    if not ID_RE.fullmatch(item_id):
        raise Refused("bad item id")
    num = issue_number(item_id)
    gh = chunker.Gh(runner, apply=a.send, cap=2)
    iss = gh.read(["issue", "view", str(num), "--json", "number,state,labels,assignees,comments"])
    check_issue(item_id, iss, kind)
    if kind == "claim" and already(item_id, iss, kind):
        print("%s: already claimed by us on #%d, nothing to do" % (item_id, num))
        return 0
    body = build(kind, item_id, a)
    gh.write("comment on #%d (%s %s)" % (num, kind, item_id), ["issue", "comment", str(num), "--body-file", "-"], body)
    if kind == "done" and a.close:
        gh.write("close #%d as completed" % num, ["issue", "close", str(num), "--reason", "completed"])
    print(body.split("\n")[0])
    if a.send:
        log_row(kind, item_id, num, a.job if kind == "claim" else a.doc, now or datetime.date.today().isoformat())
    else:
        print("(dry run: nothing posted; add --send)")
    return 0


CACHE = "data-cache/claims-watch.json"
WATCH_RULES = (("CLAIM", re.compile(r"(?i)\bclaim(ing|ed)?\b|\btake (this|it)\b|\bi.ll (take|do|work)")),
               ("WITHDRAWAL", re.compile(r"(?i)\b(releas(e|ing)|withdraw|giving up|unclaim|dropping)\b")),
               ("SUBMISSION", re.compile(r"(?i)https?://|\bPR\b|\bsubmitted\b|\bpull request\b")),
               ("QUESTION", re.compile(r"\?")))


def classify(body):
    """Keyword guess for the maintainer's eyes only; never steers a write."""
    for name, rx in WATCH_RULES:
        if rx.search(body or ""):
            return name
    return "OTHER"


def watch(runner=chunker.run_gh, cache=None, update=False):
    cache = cache or os.path.join(ROOT, CACHE)
    try:
        seen = json.load(open(cache, encoding="utf-8"))
    except (OSError, ValueError):
        seen = {}
    gh = chunker.Gh(runner)
    issues = gh.read(["issue", "list", "--label", chunker.LABEL, "--state", "all", "--limit", "200",
                      "--json", "number,state,title,comments"])
    now, rows = {}, []
    for i in issues:
        cs = i.get("comments") or []
        last = cs[-1] if cs else {}
        key = [len(cs), last.get("createdAt") or ""]
        now[str(i["number"])] = key
        if str(i["number"]) in seen and seen[str(i["number"])] == key:
            continue
        if not cs:
            continue
        who = (last.get("author") or {}).get("login") or "?"
        ours = "<!-- claim:" in (last.get("body") or "") or "<!-- chunker:" in (last.get("body") or "")
        rows.append((i["number"], i.get("state"), len(cs), "MAINTAINERS" if ours else classify(last.get("body")), who))
    for r in rows:
        print("#%d %s comments=%d newest=%s by %s" % r)
    if update:
        os.makedirs(os.path.dirname(cache), exist_ok=True)
        json.dump(now, open(cache, "w", encoding="utf-8"))
    return rows


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    sub = ap.add_subparsers(dest="cmd", required=True)
    c = sub.add_parser("claim")
    c.add_argument("item")
    c.add_argument("--job", required=True)
    c.add_argument("--note", required=True)
    d = sub.add_parser("done")
    d.add_argument("item")
    d.add_argument("--doc", required=True)
    d.add_argument("--note", required=True)
    d.add_argument("--close", action="store_true")
    w = sub.add_parser("watch")
    w.add_argument("--update", action="store_true", help="save the seen state")
    for s in (c, d):
        s.add_argument("--send", action="store_true", help="post (default: dry run)")
    a = ap.parse_args(argv)
    if a.cmd == "watch":
        watch(update=a.update)
        return 0
    try:
        return run(a)
    except Refused as e:
        print("refused: %s" % e, file=sys.stderr)
        return 2


if __name__ == "__main__":
    sys.exit(main())
