#!/usr/bin/env python3
"""Read-only plan view of the backlog: what is now, next, waiting on the owner, blocked, overdue, stale, dark.

  python3 -I tools/backlog/plan.py [--lane X] [--epic CARD] [--json] [--days 14] [--ref origin/main] [--today YYYY-MM-DD]
  python3 -I tools/backlog/plan.py --write docs/plan/QUEUE.md     # generated page, do not edit by hand

Lane of a card: its `lane:` key (now next later blocked waiting-owner parked); without one an In Progress card is
`now`, a Deferred card `parked` and anything else open `next`. A card in lane next/now whose native `dependencies` are not all Done is
shown as BLOCKED (with the open ids). Done cards never appear. STALE = In Progress or To Do whose `owns:` paths
have no commit on --ref in --days days (same test as drift.py code C). Never writes a task. Standard library only.
"""
import argparse
import json
import re
import subprocess
import sys
from datetime import date
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from check import LANES, parse  # noqa: E402
from drift import git  # noqa: E402

ROOT = Path(__file__).resolve().parents[2]
PRIO = {"high": 0, "medium": 1, "low": 2}


def tid_num(t):
    return int(re.sub(r"\D", "", t["id"]) or 0)


def load(repo):
    out = []
    for f in sorted((Path(repo) / "backlog" / "tasks").glob("task-*.md")):
        fm = parse(f)
        if not fm or fm.get("status") == "Done":
            continue
        L = fm["_lists"]
        lane = fm.get("lane") or {"In Progress": "now", "Deferred": "parked"}.get(fm.get("status"), "next")
        out.append(dict(
            id=fm.get("id", "").upper(), card=fm.get("card") or fm.get("title", "").split(":", 1)[0].strip(),
            title=fm.get("title", "").split(":", 1)[-1].strip(), status=fm.get("status", ""), lane=lane,
            explicit="lane" in fm, prio=fm.get("priority", "medium"), owner=",".join(L.get("assignee", [])) or "-",
            due=fm.get("due", ""), epic=fm.get("epic", ""), note=fm.get("lane_note", ""), deps=[d.upper() for d in L.get("dependencies", [])],
            owns=L.get("owns", []), updated=str(fm.get("updated_date") or fm.get("created_date", ""))[:10]))
    return out


def done_ids(repo):
    ids = set()
    for f in (Path(repo) / "backlog" / "tasks").glob("task-*.md"):
        fm = parse(f)
        if fm and fm.get("status") == "Done":
            ids.add(fm.get("id", "").upper())
    return ids


def key(t):
    return (PRIO.get(t["prio"], 1), t["due"] or "9999-12-31", tid_num(t))


def build(repo, today, days=14, ref="origin/main", lane=None, epic=None):
    repo = Path(repo)
    if not git(repo, "rev-parse", "--verify", "-q", ref).strip():
        ref = "HEAD"
    done = done_ids(repo)
    tasks = load(repo)
    if epic:
        tasks = [t for t in tasks if t["epic"] == epic or t["card"] == epic]
    for t in tasks:
        t["open_deps"] = [d for d in t["deps"] if d not in done]
        eff = t["lane"]
        if eff in ("now", "next") and t["open_deps"]:
            eff = "blocked"
        t["eff"] = eff
        t["age"] = (today - date.fromisoformat(t["updated"])).days if t["updated"] else None
    sec = {k: [] for k in ("now", "next", "waiting-owner", "blocked", "later", "parked", "overdue", "stale")}
    for t in sorted(tasks, key=key):
        sec[t["eff"]].append(t)
        if t["due"] and t["due"] < today.isoformat():
            sec["overdue"].append(t)
        if t["status"] in ("In Progress", "To Do") and t["owns"] and not git(repo, "log", "-1", f"--since={days}.days", "--format=%h", ref, "--", *t["owns"]).strip():
            sec["stale"].append(t)
    if lane:
        sec = {k: (v if k == lane or lane == "all" else []) for k, v in sec.items()}
    return sec


def dark_lines(repo):
    rep = Path(repo) / "tools" / "darklaunch" / "report.py"
    if not rep.exists():
        return []
    r = subprocess.run([sys.executable, "-I", str(rep)], capture_output=True, text=True, cwd=repo)
    return [x for x in r.stdout.splitlines() if x.strip()] if r.returncode == 0 else ["(darklaunch report failed)"]


def fmt(t, extra=""):
    due = f" due {t['due']}" if t["due"] else ""
    age = f" {t['age']}d" if t["age"] is not None and t["status"] == "In Progress" else ""
    return f"  {t['id']:<9}{t['card'][:22]:<23}{t['prio'][:3]:<4}{t['owner'][:13]:<14}{t['status'][:11]:<12}{age}{due}{extra}"


def render(sec, dark, limit=12):
    out = []

    def block(title, items, extra=lambda t: "", cap=None):
        if not items:
            return
        out.append(f"{title} ({len(items)})")
        for t in items[:cap]:
            out.append(fmt(t, extra(t)))
        if cap and len(items) > cap:
            out.append(f"  ... {len(items) - cap} more (plan.py --lane {title.split()[0].lower()})")

    block("NOW", sec["now"])
    block("NEXT", sec["next"], cap=limit)
    block("WAITING ON OWNER", sec["waiting-owner"], lambda t: f"  {t['note']}"[:60])
    block("BLOCKED", sec["blocked"], lambda t: "  waits for " + (" ".join(t["open_deps"]) or t["note"] or "?"))
    block("OVERDUE", sec["overdue"])
    block("STALE (no commit on owns)", sec["stale"], lambda t: f"  lane {t['lane']}")
    if sec["later"] or sec["parked"]:
        out.append(f"LATER {len(sec['later'])}, PARKED {len(sec['parked'])} (plan.py --lane later|parked)")
    if dark:
        out.append("DARK-LAUNCH")
        out += ["  " + x for x in dark[:12]]
    return out


def main(argv):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--repo", default=str(ROOT))
    ap.add_argument("--ref", default="origin/main")
    ap.add_argument("--days", type=int, default=14)
    ap.add_argument("--today", default=date.today().isoformat())
    ap.add_argument("--lane", choices=sorted(LANES) + ["overdue", "stale"], help="show one section in full")
    ap.add_argument("--epic")
    ap.add_argument("--json", action="store_true")
    ap.add_argument("--write", metavar="FILE", help="write the page (generated, do not edit)")
    a = ap.parse_args(argv)
    sec = build(a.repo, date.fromisoformat(a.today), a.days, a.ref, a.lane if a.lane in LANES else None, a.epic)
    if a.lane in ("overdue", "stale"):
        sec = {k: (v if k == a.lane else []) for k, v in sec.items()}
    if a.json:
        slim = {k: [{x: t[x] for x in ("id", "card", "status", "lane", "eff", "prio", "owner", "due", "epic", "note", "open_deps", "age")} for t in v] for k, v in sec.items()}
        print(json.dumps(slim, indent=1))
        return 0
    limit = 1000 if a.lane or a.write else 12
    lines = render(sec, [] if a.lane else dark_lines(a.repo), limit)
    if a.write:
        head = ["# Queue (generated: do not edit)", "", "Generated by `python3 -I tools/backlog/plan.py --write docs/plan/QUEUE.md`; lanes and rules: docs/backlog-howto.md, \"Planning lanes\".", "", "```"]
        Path(a.write).write_text("\n".join(head + lines + ["```", ""]))
        return 0
    print(f"PLAN {a.today}" + (f"  lane={a.lane}" if a.lane else "") + (f"  epic={a.epic}" if a.epic else ""))
    print("\n".join(lines))
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
