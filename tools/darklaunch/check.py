#!/usr/bin/env python3
"""Dark-launch ledger check (docs/process/dark-launch.md). Stdlib only.

  python3 -I tools/darklaunch/check.py [--ledger F] [--knobs F] [--today YYYY-MM-DD]

Fails when
  - a ledger row misses a field, has an unknown status, or a bad date;
  - a knob with default 0 whose description carries the marker [dark-launch] has no ledger row;
  - a row of status dark or trial has a review-by date in the past (decide: enable, extend with a new date and
    reason, or retire);
  - a row names a knob that is not in docs/knobs.md (renamed or deleted) while its status is dark or trial.
The marker is the declared contract: new default-off features put [dark-launch] in the knob description.
Existing ones are listed in the ledger by name. This is not a heuristic over all default-0 knobs.
"""
import argparse
import os
import sys
from datetime import date

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import ledger as L  # noqa: E402

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))


def problems(items, knobs, today):
    out = []
    names = set()
    for i, it in enumerate(items):
        n = it.get("name", f"#{i}")
        miss = [f for f in L.FIELDS if not str(it.get(f, "")).strip()]
        if miss:
            out.append(f"{n}: missing field(s) {','.join(miss)}")
            continue
        names.add(n)
        if it["status"] not in L.STATUSES:
            out.append(f"{n}: unknown status {it['status']!r}")
        try:
            due = L.parse_date(it["review_by"])
            L.parse_date(it["dark_since"])
        except ValueError:
            out.append(f"{n}: bad date")
            continue
        if it["status"] in L.OPEN:
            if due < today:
                out.append(f"{n}: review-by {it['review_by']} is past ({it['status']}): enable, extend with a new date and reason, or retire")
            if n not in knobs:
                out.append(f"{n}: not a knob in docs/knobs.md (renamed or deleted?): retire the row")
    for k, (default, row) in knobs.items():
        if L.MARKER in row and default in ("0", "0.0") and k not in names:
            out.append(f"{k}: default 0 with {L.MARKER} but no row in data/darklaunch.json")
    return out


def main(argv=None):
    ap = argparse.ArgumentParser()
    ap.add_argument("--ledger", default=os.path.join(ROOT, "data", "darklaunch.json"))
    ap.add_argument("--knobs", default=os.path.join(ROOT, "docs", "knobs.md"))
    ap.add_argument("--today", default=None)
    a = ap.parse_args(argv)
    today = L.parse_date(a.today) if a.today else date.today()
    ps = problems(L.load_ledger(a.ledger), L.knob_rows(a.knobs), today)
    for p in ps:
        print("FAIL darklaunch:", p)
    if not ps:
        print("darklaunch ok")
    return 1 if ps else 0


if __name__ == "__main__":
    sys.exit(main())
