# SPDX-License-Identifier: GPL-3.0-or-later
"""Names the game.exe term behind each d2d price difference.

    uv run python price_terms.py cases.tsv game.tsv d2d.tsv

Re-prices each case with price_model.py with one game.exe term switched to
d2d's behaviour at a time; a difference that one switch closes is that
term's. Terms: armor (defence scaling), class (sell / 4), unid (no quality
extras unidentified), unidrep (no repair unidentified), whole (repair at
least 1), stack (stack quantity), ethzero (ethereal at 0 durability sells 0).
"""
import collections
import sys

import price_model
import price_oracle


def main():
    o = price_oracle.Oracle()
    m = price_model.Model(o)
    game = {l.split()[0]: l.split()[1:] for l in open(sys.argv[2]) if l.strip()}
    d2d = {l.split()[0]: l.split()[1:] for l in open(sys.argv[3]) if l.strip()}
    count = collections.Counter()
    total = collections.Counter()
    examples = {}
    for line in open(sys.argv[1]):
        if not line.strip(): continue
        case = price_oracle.parse(line)
        g, d = game[case[0]], d2d[case[0]]
        for i, name in enumerate(("buy", "sell", "repair")):
            total[name] += 1
            if g[i] == d[i]: continue
            hit = []
            for term in ("armor", "class", "unid", "unidrep", "whole", "stack", "ethzero"):
                m.off = {term}
                if str(m.price(case)[i]) == d[i]: hit.append(term)
            m.off = set()
            key = (name, "+".join(hit) or "several")
            count[key] += 1
            examples.setdefault(key, line.strip())
    for name in ("buy", "sell", "repair"):
        bad = sum(v for k, v in count.items() if k[0] == name)
        print(f"{name}: {bad}/{total[name]} differ")
        for k, v in sorted(count.items(), key=lambda kv: -kv[1]):
            if k[0] == name: print(f"  {v:6d} {k[1]:12s} e.g. {examples[k]}")


if __name__ == "__main__":
    main()
