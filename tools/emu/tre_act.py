# SPDX-License-Identifier: GPL-3.0-or-later
"""What a treasure class gives per kill: exact odds from game.exe's own runtime class records, and a native roll check.

    bash tools/emu/tre_run.sh tre_act.py classes [prefix...]   # TSV: one line per TreasureClassEx class (default: names starting Act 1 / Act 2)
    bash tools/emu/tre_run.sh tre_act.py monsters              # TSV: Act 1 / Act 2 monsters (Levels.txt lists) and their classes by difficulty
    bash tools/emu/tre_run.sh tre_act.py mc [N]                # N native rolls (default 6000) of 16 classes against the exact odds
    bash tools/emu/tre_run.sh tre_act.py quality [N]           # quality shares of N native rolls per class at magic find 0 / 100 / 300

Records (V run: read out of game.exe's memory after its own loader ran, class table DAT_0096c5ec stride 0x2c,
entries stride 0x1c, expansion running totals as docs/research/re/drops.md): picks +0x10, NoDrop +0x14, total +0xc,
entry probability = next running start - own (the last: total - start). One player, no magic find, no gold find.
A pick draws rand(NoDrop + total): below NoDrop nothing; a negative Picks value draws nothing and takes entry k for
pick k (Countess). Odds of the number of drops are exact (a polynomial per class, degree 7 = "7 or more"; the cap
of 6 is not applied: P(7 or more) is printed so its effect can be read). Coins: ilvl + rand(5 ilvl), at least 1, then
x mul >> 8 for a "gld,mul=N" entry (FUN_00557ab0, FUN_0055a6d0).
"""
import math
import random
import re
import struct
import sys
from collections import defaultdict

import drops

DEG = 7
CATS = ("gold", "hp", "mp", "rv", "otherpot", "scroll", "tome", "key", "gem", "rune", "gear", "ring", "amulet", "jewel", "charm", "ammo", "misc")


class Table:
    def __init__(self, o):
        self.o, e = o, o.e
        self.base = e.r32(0x96c5ec)
        self.n = e.r32(0x96c5f0)
        self.type = {}
        for f in ("Weapons", "Armor", "Misc"):
            for r in drops.txt_rows(e, f):
                if r.get("code"):
                    self.type[r["code"]] = (f, r.get("type", ""), r.get("name", ""))
        self.cache = {}
        self.poly = {}

    def rec(self, i):
        if i in self.cache:
            return self.cache[i]
        e, o = self.o.e, self.o
        r = self.base + i * 0x2c
        n, total, ents = e.s32(r + 4), e.s32(r + 0xc), e.r32(r + 0x28)
        items = []
        for k in range(n):
            ent = ents + k * 0x1c
            start = e.s32(ent + 4)
            end = e.s32(ent + 0x1c + 4) if k + 1 < n else total
            idx, mul, flags = struct.unpack("<HHB", e.read(ent + 8, 5))
            if idx == 0xFFFF:
                items.append(("none", None, end - start, 0))
            elif flags & 4:
                items.append(("class", idx, end - start, 0))
            elif flags & 3:                      # a unique / set item entry (Annihilus, ROP): one drop, no category
                items.append(("item", f"uniq{idx}", end - start, 0))
            else:
                items.append(("item", o.codes[idx], end - start, mul))
        out = {"name": o.names[i], "picks": e.s32(r + 0x10), "nodrop": e.s32(r + 0x14), "total": total, "items": items,
               "group": e.r16(r), "level": e.r16(r + 2)}
        self.cache[i] = out
        return out

    fine = False

    def cat(self, code):
        c = self.base_cat(code)
        return code if self.fine and c in ("hp", "mp", "rv", "otherpot", "scroll", "tome", "key", "gem", "rune") else c

    def base_cat(self, code):
        if code == "gld":
            return "gold"
        f, t, _ = self.type.get(code, ("", "", ""))
        if code in ("tsc", "isc"): return "scroll"
        if code in ("tbk", "ibk"): return "tome"
        if code == "key": return "key"
        if re.fullmatch(r"r\d\d", code): return "rune"
        if t == "hpot": return "hp"
        if t == "mpot": return "mp"
        if t == "rpot": return "rv"
        if t in ("spot", "apot", "wpot"): return "otherpot"
        if t.startswith("gem"): return "gem"
        if t == "ring": return "ring"
        if t == "amul": return "amulet"
        if t == "jewl": return "jewel"
        if t in ("scha", "mcha", "lcha"): return "charm"
        if t in ("bowq", "xboq"): return "ammo"
        if f in ("Weapons", "Armor"): return "gear"
        return "misc"

    def one(self, i):
        """(poly, exp, piles): the drop-count polynomial, expected drops per category, expected gold piles by mul."""
        if i in self.poly:
            return self.poly[i]
        c = self.rec(i)
        T = c["nodrop"] + c["total"] if c["picks"] >= 0 else c["total"]
        pick = [0.0] * (DEG + 1)
        pexp, ppile = defaultdict(float), defaultdict(float)
        outs = []          # (weight, poly, exp, piles) per entry
        for kind, what, prob, mul in c["items"]:
            if kind == "none":
                outs.append((prob, [1.0] + [0.0] * DEG, {}, {}))
            elif kind == "class":
                outs.append((prob,) + self.one(what))
            else:
                p = [0.0] * (DEG + 1); p[1] = 1.0
                cat = self.cat(what)
                outs.append((prob, p, {cat: 1.0}, {mul: 1.0} if what == "gld" else {}))
        if c["picks"] >= 0:
            k = max(c["picks"], 1)
            if T == 0:
                return self.poly.setdefault(i, ([1.0] + [0.0] * DEG, {}, {}))
            pick[0] += c["nodrop"] / T
            for w, p, ex, pl in outs:
                for d in range(DEG + 1): pick[d] += w / T * p[d]
                for kk, v in ex.items(): pexp[kk] += w / T * v
                for kk, v in pl.items(): ppile[kk] += w / T * v
            poly = [1.0] + [0.0] * DEG
            for _ in range(k): poly = mul_poly(poly, pick)
            exp = {kk: v * k for kk, v in pexp.items()}
            piles = {kk: v * k for kk, v in ppile.items()}
        else:                                   # negative picks: entry k holds position k, no draw
            poly, exp, piles = [1.0] + [0.0] * DEG, defaultdict(float), defaultdict(float)
            pos = 0
            for k in range(abs(c["picks"])):
                if k >= c["total"]:
                    break
                acc, hit = 0, None
                for entry, o_ in zip(c["items"], outs):
                    if k < acc + entry[2]:
                        hit = o_
                        break
                    acc += entry[2]
                if hit is None:
                    continue
                w, p, ex, pl = hit
                poly = mul_poly(poly, p)
                for kk, v in ex.items(): exp[kk] += v
                for kk, v in pl.items(): piles[kk] += v
        self.poly[i] = (poly, dict(exp), dict(piles))
        return self.poly[i]


def mul_poly(a, b):
    out = [0.0] * (DEG + 1)
    for i, x in enumerate(a):
        if x == 0: continue
        for j, y in enumerate(b):
            out[min(i + j, DEG)] += x * y
    return out


def coins(ilvl, mul):
    """Mean of max(ilvl + r, 1) (x mul >> 8) over r in 0..5 ilvl - 1 (FUN_00557ab0)."""
    ilvl = max(ilvl, 1)
    vals = [max(ilvl + r, 1) for r in range(5 * ilvl)]
    return sum(((v * mul) >> 8) if mul else v for v in vals) / len(vals)


def stats_line(t, i):
    poly, exp, piles = t.one(i)
    c = t.rec(i)
    en = sum(d * p for d, p in enumerate(poly))
    cols = [c["name"], c["picks"], c["nodrop"], c["total"], f"{c['nodrop'] / (c['nodrop'] + c['total']) * 100:.1f}" if c["nodrop"] + c["total"] else "-",
            f"{1 - poly[0]:.4f}", f"{en:.3f}", f"{poly[DEG]:.5f}"]
    cols += [f"{exp.get(k, 0):.4f}" for k in CATS]
    cols.append(" ".join(f"mul{m}:{v:.3f}" for m, v in sorted(piles.items())))
    cols += [c["group"], c["level"]]
    return "\t".join(str(x) for x in cols)


def parse_roll(line):
    got = line.split(":", 1)[1].rsplit(" ->", 1)[0].split()
    out = []
    for tok in got:
        code, rest = tok.split(":", 1)
        q, _, m = rest.partition("*")
        out.append((code, int(q), int(m or 0)))
    return out


def main():
    cmd = sys.argv[1] if len(sys.argv) > 1 else "classes"
    o = drops.Oracle()
    t = Table(o)
    e = o.e
    if cmd == "classes":
        prefixes = tuple(sys.argv[2:]) or ("Act 1", "Act 2")
        print("\t".join(["class", "picks", "nodrop", "total", "nodrop%", "P(>=1)", "E[drops]", "P(7+)"] + [f"E[{k}]" for k in CATS] + ["gold piles by mul", "group", "level"]))
        for i, name in enumerate(o.names):
            if name and name.startswith(prefixes) and t.rec(i)["items"] is not None:
                print(stats_line(t, i))
        return
    if cmd == "codes":                          # expected drops per consumable code, for the named classes
        t.fine = True
        for name in sys.argv[2:]:
            poly, exp, piles = t.one(o.index[name])
            print(name + ":", " ".join(f"{k}={v:.4f}" for k, v in sorted(exp.items()) if v >= 0.0005))
        return
    if cmd == "bosses":                         # SuperUniques rows (TC by difficulty) and the MonStats rows with a quest class (TCQuestId)
        t.fine = False
        def show(label, tcs):
            parts = []
            for tc in tcs:
                if tc and tc in o.index:
                    poly, exp, piles = t.one(o.index[tc])
                    en = sum(d * p for d, p in enumerate(poly))
                    parts.append(f"{tc} [pk{t.rec(o.index[tc])['picks']} nd{t.rec(o.index[tc])['nodrop']}/{t.rec(o.index[tc])['total']} E{en:.2f} gear{exp.get('gear', 0):.2f} pot{exp.get('hp', 0) + exp.get('mp', 0) + exp.get('rv', 0):.2f} rune{exp.get('rune', 0):.3f}]")
                else:
                    parts.append(f"{tc or '-'}")
            print(label + ": " + " | ".join(parts))
        for r in drops.txt_rows(e, "SuperUniques"):
            if r.get("Superunique") and r["Superunique"] != "Expansion":
                show(f"super {r['hcIdx']} {r['Superunique']} ({r['Class']})", [r["TC"], r["TC(N)"], r["TC(H)"]])
        for r in drops.txt_rows(e, "MonStats"):
            if r.get("Id") and r.get("TCQuestId") not in ("", "0", None):
                show(f"quest {r['Id']} id{r['TCQuestId']}/cp{r['TCQuestCP']} boss{r.get('boss')} noRatio{r.get('noRatio')}",
                     [r["TreasureClass1"], r["TreasureClass4"], r["TreasureClass1(N)"], r["TreasureClass4(N)"], r["TreasureClass1(H)"], r["TreasureClass4(H)"]])
        return
    if cmd == "players":                        # NoDrop share of a plain monster class by the player count, native rolls against the formula
        n_rolls = int(sys.argv[2]) if len(sys.argv) > 2 else 20000
        c = t.rec(o.index["Act 1 H2H A"])
        print("players\tNoDrop (formula, truncated)\tnodrop%\tnative nodrop%\tnative drops/kill")
        for n in range(1, 9):
            p = c["nodrop"] / (c["nodrop"] + c["total"]); q = 1 - p ** n
            nd = 0 if q == 0 else int((1 - q) * c["total"] / q)
            rng, zero, drops_n = random.Random(5), 0, 0
            for _ in range(n_rolls):
                got = parse_roll(o.roll(rng.getrandbits(32), "Act 1 H2H A", 0, 10, n, 0))
                zero += not got; drops_n += len(got)
            print(f"{n}\t{nd}\t{nd / (nd + c['total']) * 100:.1f}\t{zero / n_rolls * 100:.1f}\t{drops_n / n_rolls:.3f}")
        return
    if cmd == "coins":                        # mean coins per gold pile at an item level (mul 0), and per kill with the plain monster class's pile rate
        print("ilvl\tmean coins\tmin\tmax\tcoins per Act 1 H2H kill")
        piles = t.one(o.index["Act 1 H2H A"])[2].get(0, 0)
        for il in (1, 2, 3, 4, 5, 6, 8, 10, 12, 15, 20, 25, 30, 40, 50, 60, 70, 85):
            print(f"{il}\t{coins(il, 0):.2f}\t{il}\t{6 * il - 1}\t{piles * coins(il, 0):.2f}")
        return
    if cmd == "monsters":
        mon = {r["Id"].lower(): r for r in drops.txt_rows(e, "MonStats") if r.get("Id") and r.get("hcIdx")}
        lev = [r for r in drops.txt_rows(e, "Levels") if r.get("Act") in ("0", "1") and r.get("mon1")]
        seen = defaultdict(set)
        for r in lev:
            for pre, key in (("mon", "n"), ("nmon", "nh"), ("umon", "u")):
                for k in range(1, 11):
                    m = r.get(f"{pre}{k}", "").lower()
                    if m:
                        seen[m].add((int(r["Id"]), r["LevelName"], key))
        print("\t".join(["monster", "hcIdx", "act", "Level(N)", "areas(N,NM/H,unique lists)", "TC1", "TC2", "TC3", "TC4", "TC1(N)", "TC2(N)", "TC3(N)", "TC4(N)", "TC1(H)", "TC2(H)", "TC3(H)", "TC4(H)", "TCQuest", "noRatio", "boss"]))
        lvl_acts = {int(r["Id"]): r["Act"] for r in lev}
        for m in sorted(seen):
            r = mon.get(m)
            if not r:
                print(m, "no MonStats row", sep="\t")
                continue
            acts = sorted({lvl_acts[a] for a, _, _ in seen[m]})
            areas = ",".join(sorted({n for _, n, _ in seen[m]}))[:80]
            print("\t".join([m, r["hcIdx"], "".join(acts), r["Level"], areas] + [r.get(f"TreasureClass{k}{s}", "") for s in ("", "(N)", "(H)") for k in range(1, 5)]
                            + [f"{r.get('TCQuestId', '')}/{r.get('TCQuestCP', '')}", r.get("noRatio", ""), r.get("boss", "")]))
        return
    n_rolls = int(sys.argv[2]) if len(sys.argv) > 2 else 6000
    picks = [n for n in ("Act 1 H2H A", "Act 1 H2H C", "Act 1 (N) H2H B", "Act 1 (H) Ranged A", "Act 1 Good", "Act 1 Champ A", "Act 1 (N) Unique B", "Act 1 (H) Super A",
                         "Act 2 H2H A", "Act 2 (N) Ranged B", "Act 2 (H) Champ C", "Act 2 Unique A", "Andariel", "Andarielq (N)", "Act 1 Chest A", "Act 2 Large Chest B")
             if n in o.index]
    if cmd == "mc":
        rng = random.Random(7)
        print(f"class\trolls\tnodrop% exact/native\tE[drops] exact/native\tgold piles exact/native\tpotions(hp+mp+rv) exact/native\tscroll+tome exact/native\tgear exact/native\tmax |z|")
        for name in picks:
            i = o.index[name]
            poly, exp, piles = t.one(i)
            sums, zs = defaultdict(list), []
            for _ in range(n_rolls):
                line = o.roll(rng.getrandbits(32), name, 0, 30, 1, 0)
                got = parse_roll(line)
                sums["n"].append(len(got))
                for k, grp in (("gold", ("gold",)), ("pot", ("hp", "mp", "rv")), ("scr", ("scroll", "tome")), ("gear", ("gear",))):
                    sums[k].append(sum(1 for c, _, _ in got if t.cat(c) in grp))
            def mean(v): return sum(v) / len(v)
            def z(v, exact):
                m = mean(v); sd = (sum((x - m) ** 2 for x in v) / len(v)) ** 0.5
                return (m - exact) / (sd / len(v) ** 0.5) if sd else (0 if abs(m - exact) < 1e-9 else 99)
            zero = sum(1 for x in sums["n"] if x == 0) / n_rolls
            exact = {"n": sum(d * p for d, p in enumerate(poly)), "gold": exp.get("gold", 0), "pot": sum(exp.get(k, 0) for k in ("hp", "mp", "rv")),
                     "scr": exp.get("scroll", 0) + exp.get("tome", 0), "gear": exp.get("gear", 0)}
            # the cap (6) trims only a count above 6: compare with the capped exact mean for the number of drops
            exact["n"] = sum(min(d, 6) * p for d, p in enumerate(poly))
            zz = [z(sums[k], exact[k]) for k in exact]
            zs.append(abs((zero - poly[0]) / math.sqrt(max(poly[0] * (1 - poly[0]), 1e-12) / n_rolls)) if 0 < poly[0] < 1 else 0)
            print(f"{name}\t{n_rolls}\t{poly[0] * 100:.2f}/{zero * 100:.2f}\t{exact['n']:.3f}/{mean(sums['n']):.3f}\t{exact['gold']:.3f}/{mean(sums['gold']):.3f}\t"
                  f"{exact['pot']:.3f}/{mean(sums['pot']):.3f}\t{exact['scr']:.3f}/{mean(sums['scr']):.3f}\t{exact['gear']:.3f}/{mean(sums['gear']):.3f}\t{max(zz + zs):.2f}")
        return
    if cmd == "quality":
        print("class\tilvl\tmf\titems\t" + "\t".join(["low%", "normal%", "superior%", "magic%", "set%", "rare%", "unique%"]))
        names = {"Act 1 (N) Champ A": 24, "Act 1 (H) Super A": 60, "Act 2 (H) Ranged B": 70, "Act 1 H2H B": 5, "Andariel": 12}
        for name, il in names.items():
            if name not in o.index:
                continue
            for mf in (0, 100, 300):
                rng, q = random.Random(11), defaultdict(int)
                tot = 0
                for _ in range(int(sys.argv[2]) if len(sys.argv) > 2 else 6000):
                    for c, qq, _m in parse_roll(o.roll(rng.getrandbits(32), name, 0, il, 1, mf)):
                        if t.cat(c) in ("gear", "ring", "amulet", "jewel", "charm"):
                            q[qq] += 1; tot += 1
                print(f"{name}\t{il}\t{mf}\t{tot}\t" + "\t".join(f"{q[k] / tot * 100:.2f}" if tot else "-" for k in (1, 2, 3, 4, 5, 6, 7)))
        return


if __name__ == "__main__":
    main()
