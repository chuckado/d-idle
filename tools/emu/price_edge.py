# SPDX-License-Identifier: GPL-3.0-or-later
"""Edge cases for the vendor price oracle: big prices and repair gates.

    uv run python price_edge.py 100000 11 > cases.tsv
    uv run python price_oracle.py game cases.tsv > game.tsv
    ../../build/tests/test_game --price cases.tsv | grep -E '^[0-9]+ ' > d2d.tsv
    uv run python price_claims.py cases.tsv game.tsv d2d.tsv

Writes price_oracle.py cases (same columns) aimed at two terms of
FUN_0062efb0 that its random generator rarely reaches:

- big: an armor whose defence puts its buy, sell or repair base around
  0xffff / 0x10000 (exact neighbours where the defence allows it) or far
  above, where the npc.txt and quest mults divide by 1024 first;
- gate: items whose ItemTypes row has no Repair, whose items row has
  nodurability or no durability column (bows, crossbows, throwing potions,
  potions, rings), with saved durability 0, 1, max - 1, max and 255;
- the rest: any priced item, every quality, durability at 0 / 1 / max /
  255, stack sizes 1 / max / random, ethereal, unidentified.

Vendors cycle through all 17, difficulties through 0..2. Bases are kept
below 0x7fffffff / 0x800 so game.exe's 32-bit products (cost x defence,
(base / 1024) x mult, (max - dur) x base) cannot wrap: the wrap is not
modelled by either side.
"""
import random
import sys

import price_oracle

SAFE = 0x7FFFFFFF // 0x800   # the biggest base whose 32-bit products cannot wrap
TARGETS = (0xFFFF, 0x10000, 4 * 0xFFFF, 4 * 0x10000, 16 * 0xFFFF, 16 * 0x10000)


def defence_for(base, cost, maxac):
    """The defence whose cost x defence / maxac is closest to base from below."""
    return max(base * maxac // cost, 0)


def main():
    n, seed = int(sys.argv[1]), int(sys.argv[2])
    o = price_oracle.Oracle()
    e = o.e
    rnd = random.Random(seed)
    lines = o.gen(n, seed).copy()          # the oracle's own mix as the template for the affix columns
    probe = e.alloc(0x20)
    e.w32(probe, 4)
    types = e.r32(e.r32(0x744304) + 0xbf8)

    def row(code):
        e.w32(probe + 4, o.cls[code])
        rec = o.items[o.cls[code]][1]
        ty = types + e.call(0x628700, o.cls[code]) * 0xe4
        return dict(rec=rec, cost=e.s32(rec + 0xe0), minac=e.s32(rec + 0xcc), maxac=e.s32(rec + 0xd0),
                    armor=bool(e.call(0x629bb0, probe, 0x32)), repair=e.read(ty + 8, 1)[0], nodur=e.read(rec + 0x113, 1)[0],
                    durcol=e.read(rec + 0x112, 1)[0], stack=e.read(rec + 0x132, 1)[0],
                    maxstack=e.call(0x629570, o.cls[code]) if e.read(rec + 0x132, 1)[0] else 0)

    codes = sorted({line.split()[1] for line in lines})
    rows = {c: row(c) for c in codes}
    armors = [c for c in codes if rows[c]["armor"] and rows[c]["maxac"] > 0 and rows[c]["cost"] > 0]
    gated = [c for c in codes if not rows[c]["repair"] or rows[c]["nodur"] or not rows[c]["durcol"]]
    out = []
    for i, line in enumerate(lines):
        f = line.split()
        kind = i % 10
        npc, diff = price_oracle.VENDORS[i % 17].lower(), (i // 17) % 3
        if kind < 4:                                      # big
            code = rnd.choice(armors)
            r = rows[code]
            if kind < 2:
                base = rnd.choice(TARGETS) + rnd.randint(-8, 8)
            else:
                base = rnd.randint(0x8000, SAFE)
            ac = min(defence_for(base, r["cost"], r["maxac"]), 0x7FFFFFFF // r["cost"])
            f[1], f[15] = code, str(ac)
            maxdur = r["durcol"] if not r["nodur"] else 0
            f[13] = str(rnd.choice([maxdur, 255, 1, 0]) if rnd.random() < 0.3 else maxdur)
            f[12] = str(rnd.choice([0, 1, int(f[13]) - 1, int(f[13])]) if int(f[13]) > 0 else 0)
            f[14] = "0"
        elif kind < 7:                                    # gate
            code = rnd.choice(gated)
            r = rows[code]
            f[1] = code
            f[13] = str(rnd.choice([0, 1, 2, r["durcol"], 255, rnd.randint(1, 255)]))
            md = int(f[13])
            f[12] = str(rnd.choice([0, 1, max(md - 1, 0), md, rnd.randint(0, md)]))
            f[14] = str(rnd.choice([1, r["maxstack"], rnd.randint(1, max(r["maxstack"], 1))])) if r["stack"] else "0"
            lo, hi = r["minac"], r["maxac"]
            f[15] = str(rnd.randint(min(lo, hi), max(lo, hi)) if hi > 0 else 0)
        else:                                             # the rest, with edge durability and stacks
            r = rows[f[1]]
            md = int(f[13])
            if rnd.random() < 0.5:
                md = rnd.choice([0, 1, r["durcol"], 255])
                f[13] = str(md)
            f[12] = str(rnd.choice([0, 1, max(md - 1, 0), md, int(f[12]) if int(f[12]) <= md else md]))
            if r["stack"]:
                f[14] = str(rnd.choice([1, r["maxstack"], max(r["maxstack"] - 1, 1), int(f[14]) or 1]))
        f[0], f[16], f[17] = str(i), npc, str(diff)
        out.append(" ".join(f))
    print("\n".join(out))


if __name__ == "__main__":
    main()
