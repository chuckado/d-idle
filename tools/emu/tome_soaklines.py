#!/usr/bin/env python3
"""Classify the bot's failed town portals in soak logs (d2lod-bot, card decomptome).

Read-only, streaming: one pass per log, no file written. For every
"d2d town: no portal opened, falling back" line (combat's d2dPortal gave up
after its three tries, cmd/d2bot/d2dwire.go) it records the error and the
context the log holds before it:

  tp, tome     the last "shop: visit" line (town's scrollCount: tome charges +
               loose scrolls, and whether a tome is carried)
  trips        "town trip" lines since that shop visit (each one a portal
               the kit had to pay for, an upper bound on charges spent)
  hostiles     hostileNear of the last "status" line
  emergency    the "town trip" line that started this trip

and buckets it:

  ctx-dead     the error is a cancelled context: the three tries fail in
               milliseconds, nothing is clicked (bot-side)
  empty-shop   tp == 0 at the last shop visit (the click hits an empty tome)
  empty-est    tp - trips <= 0 (the tome ran out since the shop, estimate)
  tome-empty-scroll-left  the tome's own charges (tp - loose tsc of the
               last "inventory contents") minus trips <= 0 but loose scrolls
               remain: the pick takes the first Portal item, the tome (F4)
  charged      tp - trips > 0: the tome should still hold charges; the cast
               was refused or cut (d2d's busy / get-hit rule) or the check
               is wrong
  no-shop      no shop visit logged before it

    python3 tome_soaklines.py LOG...
"""
import json
import sys
from collections import Counter


def scan(path):
    tp = tome = None
    trips = 0
    loose = 0
    hostiles = None
    emergency = None
    per = []
    with open(path, errors="replace") as f:
        for line in f:
            if '"msg":"' not in line:
                continue
            try:
                r = json.loads(line)
            except ValueError:
                continue
            m = r.get("msg", "")
            if m == "inventory contents":
                loose = 0
                for part in r.get("pack", "").split(","):
                    code, _, n = part.strip().partition(" x")
                    if code == "tsc":
                        loose = int(n or 1)
            elif m == "shop: visit" and "tp" in r:
                tp, tome, trips = r["tp"], r["tome"], 0
            elif m == "town trip":
                trips += 1
                emergency = r.get("emergency")
            elif m == "status":
                hostiles = r.get("hostileNear")
            elif m == "d2d town: no portal opened, falling back":
                err = r.get("err", "")
                if "context canceled" in err or "deadline" in err:
                    kind = "ctx-dead"
                elif tp is None:
                    kind = "no-shop"
                elif tp == 0:
                    kind = "empty-shop"
                elif tp - (trips - 1) <= 0:
                    kind = "empty-est"
                elif tome and tp - loose - (trips - 1) <= 0:
                    kind = "tome-empty-scroll-left"
                else:
                    kind = "charged"
                per.append((kind, hostiles, emergency, tome))
    return per


def main():
    total = Counter()
    hostile = Counter()
    for path in sys.argv[1:]:
        per = scan(path)
        c = Counter(k for k, *_ in per)
        total.update(c)
        for kind, h, _, tome in per:
            hostile[(kind, "hostiles>0" if h else "hostiles=0")] += 1
            if kind == "charged":
                hostile[("charged", "tome" if tome else "scrolls only")] += 1
        print(path.split("/")[-3] if path.count("/") >= 2 else path, dict(c))
    print("total", dict(total), sum(total.values()))
    print("by hostiles near", dict(sorted(hostile.items())))


if __name__ == "__main__":
    main()
