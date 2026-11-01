# SPDX-License-Identifier: GPL-3.0-or-later
"""VERIFYCURSEIMMUNE's own oracle: a curse / enemy aura's resistance change on a defender, by game.exe's code.

    python verify_curse_immune.py gen SEED N > cases.tsv      # SEED masked to 32 bits
    python verify_curse_immune.py game cases.tsv > game.tsv   # self-checks on stderr
    python verify_curse_immune.py compare cases.tsv game.tsv d2d.tsv [vlist.txt]

Independent of the author's sk_resist.py / test_sk_resist.cpp (not read before the first run).
Run for real (unicorn, upstream's emu.py): FUN_005c3540 (the per-stat adjustment the curse callbacks
FUN_005c35c0 / FUN_005c3de0 and the aura callback FUN_005cedc0 call) and FUN_00645b20 (the dm
diminishing formula, the only `imul 0x6e; add 6; idiv` in the binary). Hooked: FUN_006253b0 (the
unit's own base stat: the case's base) and FUN_0063ee90 (hireling class: the case's defender kind).
Rules taken from my disassembly (not run): FUN_005c35c0 applies no curse when its FIRST aurastat comes
out 0 after FUN_005c3540 (0x5c3639); the other stats and the aura's stats are skipped when 0.
Calc texts (Skills.txt, read from the MPQ and asserted): Lower Resist -dm56 -> -FUN_00645b20(L, 25, 70);
Conviction -min(ln34,150) -> -min(30 + 5 (L - 1), 150) (ln: I); Amplify Damage -par5 = -100;
Decrepify par5 = -50.
Case: id class b0 b1 b2 b3 b4 b5 curse clvl aura alvl def   (b = damage magic fire light cold poison;
curse / aura Skills ids, -1 / 0 none; def 0 monster, 1 hireling, 2 player: game side only).
Output: id r0..r5 (the totals: base + adjusted changes).
"""
import os
import sys

STATS = (36, 37, 39, 41, 43, 45)
LR, AMP, DEC, CONV = 91, 66, 87, 123
CURSES = (LR, AMP, DEC)
B = (-150, -100, -50, -1, 0, 1, 20, 50, 75, 95, 99, 100, 101, 104, 105, 110, 115, 125, 150, 200, 1000)


class Lcg:
    def __init__(self, seed):
        self.x = seed & 0xFFFFFFFF

    def next(self, n):
        self.x = (self.x * 22695477 + 1) & 0xFFFFFFFF
        return (self.x >> 7) % n

    def pick(self, seq):
        return seq[self.next(len(seq))]


def tables():
    import emu
    e = emu.Emu()

    def tbl(name):
        t = e.mpq_file("data\\global\\excel\\%s.txt" % name).decode("latin-1").splitlines()
        h = t[0].split("\t")
        return [dict(zip(h, ln.split("\t"))) for ln in t[1:]]
    return e, tbl


def gen(seed, n):
    r = Lcg(seed)
    rows = []
    for sk in CURSES + (CONV,):                                   # class 0: one effect, levels 1..60, every base
        for lvl in range(1, 61):
            for b in B:
                rows.append((0, (b,) * 6, sk if sk != CONV else -1, lvl if sk != CONV else 0, CONV if sk == CONV else 0, lvl if sk == CONV else 0, 0))
    for sk in CURSES:                                             # class 1: curse + Conviction, every base
        for cl in (1, 5, 10, 20, 30, 60):
            for al in (1, 5, 10, 20, 24, 25, 26, 60):
                for b in B:
                    rows.append((1, (b,) * 6, sk, cl, CONV, al, 0))
    _e, tbl = tables()                                            # class 2: real MonStats columns of 3 difficulties
    cols = ("ResDm", "ResMa", "ResFi", "ResLi", "ResCo", "ResPo")
    real = set()
    for m in tbl("MonStats"):
        for suf in ("", "(N)", "(H)"):
            try:
                real.add(tuple(int(m.get(c + suf) or 0) for c in cols))
            except ValueError:
                continue
    real = sorted(real)
    adds = ((0,) * 6, (0, 0, 40, 40, 40, 0), (0, 0, 75, 0, 0, 0), (0, 0, 0, 75, 0, 0), (0, 0, 0, 0, 75, 0), (0, 0, 40, 115, 40, 0))
    while sum(1 for x in rows if x[0] == 2) < n:
        for base in real:
            add = r.pick(adds)
            b = tuple(x + y for x, y in zip(base, add))
            curse = r.pick((-1,) + CURSES)
            aura = r.pick((0, CONV))
            if curse < 0 and not aura:
                curse = r.pick(CURSES)
            rows.append((2, b, curse, r.next(60) + 1 if curse >= 0 else 0, aura, r.next(60) + 1 if aura else 0, 0))
    def rbase():
        return r.pick((r.next(1151) - 150, r.next(41) + 90, r.pick(B), r.next(201) - 100))
    for cls in (3, 4):                                            # 3: one random effect; 4: curse + aura
        for _ in range(n):
            b = tuple(rbase() for _ in range(6))
            if cls == 3:
                if r.next(4):
                    curse, aura = r.pick(CURSES), 0
                else:
                    curse, aura = -1, CONV
            else:
                curse, aura = r.pick(CURSES), CONV
            rows.append((cls, b, curse, r.next(99) + 1 if curse >= 0 else 0, aura, r.next(99) + 1 if aura else 0, 0))
    for _ in range(20000):                                        # 5: hireling / player defenders (game side only)
        b = tuple(rbase() for _ in range(6))
        rows.append((5, b, r.pick(CURSES), r.next(60) + 1, CONV, r.next(60) + 1, r.pick((1, 2))))
    for i, (cls, b, curse, cl, aura, al, d) in enumerate(rows):
        print(i, cls, *b, curse, cl, aura, al, d, sep="\t")


class Oracle:
    def __init__(self):
        e, tbl = tables()
        self.e = e
        isc = {x["Stat"]: int(x["ID"]) for x in tbl("ItemStatCost") if x.get("ID", "").isdigit()}
        self.direct = {int(x["ID"]): x.get("direct") for x in tbl("ItemStatCost") if x.get("ID", "").isdigit()}
        self.skills = {}
        for s in tbl("Skills"):
            if s.get("Id", "").isdigit() and int(s["Id"]) in CURSES + (CONV,):
                st = [(isc.get(s["aurastat%d" % k], -1), s["aurastatcalc%d" % k].strip('"')) for k in range(1, 7) if s["aurastat%d" % k]]
                self.skills[int(s["Id"])] = (s, st)
        expect = {LR: ["-dm56"] * 4, AMP: ["-par5"], DEC: ["par5"] * 4, CONV: ["-dm56"] + ["-min(ln34,150)"] * 3}
        for k, (s, st) in self.skills.items():
            assert [c for _, c in st] == expect[k], (k, st)
        assert self.skills[LR][0]["Param5"] == "25" and self.skills[LR][0]["Param6"] == "70"
        assert self.skills[CONV][0]["Param3"] == "30" and self.skills[CONV][0]["Param4"] == "5"
        assert self.skills[AMP][0]["Param5"] == "100" and self.skills[DEC][0]["Param5"] == "-50"
        self.unit = e.alloc(0x100)
        self.cur_base = {}
        self.hire = 0
        e.hook(0x6253B0, lambda h: self.cur_base.get(h.arg(1), 0) & 0xFFFFFFFF, 3)
        e.hook(0x63EE90, lambda h: self.hire, 1)
        self.memo = {}
        self.native_calls = 0

    def dm(self, lvl, lo, hi):
        return self.e.call(0x645B20, lvl, lo, hi) & 0xFFFFFFFF

    def value(self, skill, k, lvl):
        s, st = self.skills[skill]
        stat, text = st[k]
        if text == "-dm56":
            v = self.dm(lvl, int(s["Param5"]), int(s["Param6"]))
            v = -(v - (1 << 32) if v & 0x80000000 else v)
        elif text == "-min(ln34,150)":
            v = -min(int(s["Param3"]) + (lvl - 1) * int(s["Param4"]), 150)
        elif text == "-par5":
            v = -int(s["Param5"])
        else:
            v = int(s["Param5"])
        return stat, v

    def adjust(self, stat, value, base, kind):
        key = (stat, value, base.get(stat, 0) if stat in STATS else 0, kind)
        if key not in self.memo:
            self.cur_base = base
            self.hire = 1 if kind == 1 else 0
            self.e.w32(self.unit, 0 if kind == 2 else 1)
            got = self.e.call(0x5C3540, value & 0xFFFFFFFF, ecx=self.unit, edx=stat) & 0xFFFFFFFF
            self.native_calls += 1
            self.memo[key] = got - (1 << 32) if got & 0x80000000 else got
        return self.memo[key]

    def run(self, b, curse, cl, aura, al, kind):
        base = dict(zip(STATS, b))
        tot = dict(base)
        for skill, lvl, is_curse in ((curse, cl, True), (aura, al, False)):
            if skill is None or skill <= 0 or lvl <= 0:
                continue
            changes = []
            for k in range(len(self.skills[skill][1])):
                stat, v = self.value(skill, k, lvl)
                if is_curse or v != 0:                       # the aura skips a 0 before the call
                    a = self.adjust(stat, v, base, kind)
                    if is_curse and k == 0 and a == 0:       # FUN_005c35c0 0x5c3639: no curse at all
                        changes = []
                        break
                    changes.append((stat, a))
            for stat, a in changes:
                if stat in tot:
                    tot[stat] += a
        return [tot[s] for s in STATS]

    def selfcheck(self):
        hit = [s for s in range(0, 0x170) if self.adjust(s, -50, {x: 150 for x in range(0x170)}, 0) != -50]
        print("# stats FUN_005c3540 changes (value -50, base 150):", hit, file=sys.stderr)
        bad = 0
        for st in STATS:
            for b in (99, 100, 101):
                for v in range(-260, 41):
                    want = int(v / 5) if (v < 0 and b >= 100) else v
                    bad += self.adjust(st, v, {st: b}, 0) != want
        print("# FUN_005c3540 vs trunc(v/5) rule, 6 stats x bases 99/100/101 x v -260..40: mismatches", bad, file=sys.stderr)
        print("# ItemStatCost direct for 36..45:", {k: self.direct.get(k) for k in range(36, 46)}, file=sys.stderr)
        dm = [self.dm(lvl, 25, 70) for lvl in (1, 2, 5, 10, 20, 60, 99)]
        print("# FUN_00645b20(L,25,70) L=1,2,5,10,20,60,99:", dm, file=sys.stderr)
        self.memo.clear()
        self.native_calls = 0


def game(path):
    o = Oracle()
    o.selfcheck()
    for line in open(path):
        f = line.split()
        if not f:
            continue
        cid, _cls, *rest = map(int, f)
        b, (curse, cl, aura, al, kind) = rest[:6], rest[6:]
        print(cid, *o.run(b, curse, cl, aura, al, kind))
    print("# native FUN_005c3540 calls (memoized on stat, value, base, kind):", o.native_calls, file=sys.stderr)
    # game-side values for the d2d --vlist comparison
    with open(path + ".gvals", "w") as out:
        for skill in (LR, AMP, DEC, CONV):
            for lvl in range(1, 100):
                print(skill, lvl, *("%d:%d" % o.value(skill, k, lvl) for k in range(len(o.skills[skill][1]))), file=out)


def compare(case_path, game_path, d2d_path, vlist=None):
    cases = {int(f[0]): list(map(int, f[1:])) for f in (ln.split() for ln in open(case_path)) if f}
    g = {int(f[0]): list(map(int, f[1:7])) for f in (ln.split() for ln in open(game_path)) if f and f[0].isdigit()}
    d, flags = {}, {"AURAOUT": 0, "BASEFAIL": 0}
    for f in (ln.split() for ln in open(d2d_path)):
        if f and f[0].isdigit() and len(f) >= 7:
            d[int(f[0])] = list(map(int, f[1:7]))
            for x in f[7:]:
                flags[x] = flags.get(x, 0) + 1
    per = {}
    cats = {}
    ex = []
    for cid, c in cases.items():
        cls, b = c[0], c[1:7]
        p = per.setdefault(cls, [0, 0, 0, 0])
        p[0] += 1
        if cls == 5:
            continue
        if cid not in d or cid not in g:
            p[3] += 1
            continue
        diff = [i for i in range(6) if g[cid][i] != d[cid][i]]
        if diff:
            p[1] += 1
            if any((g[cid][i] >= 100) != (d[cid][i] >= 100) for i in range(6)):
                p[2] += 1
            for i in diff:
                key = ("base>=100" if b[i] >= 100 else "base<100") + (" lowered" if d[cid][i] < b[i] else " raised/same")
                cats[key] = cats.get(key, 0) + 1
            if len(ex) < 4:
                ex.append((cid, c, g[cid], d[cid]))
    print(f"{d2d_path}: flags {flags}")
    for cls in sorted(per):
        n, nd, imm, miss = per[cls]
        print(f"  class {cls}: cases {n} rows differ {nd} immunity verdict differs {imm} missing {miss}" + (" (game only)" if cls == 5 else ""))
    print("  per-type differences by category:", cats)
    for x in ex:
        print("  e.g.", x)
    # class 5: game.exe gives hirelings / players the full value
    full_bad = 0
    o = None
    for cid, c in cases.items():
        if c[0] != 5:
            continue
        if o is None:
            o = Oracle()
        b, (curse, cl, aura, al, kind) = c[1:7], c[7:]
        want = list(b)
        for skill, lvl in ((curse, cl), (aura, al)):
            if skill > 0 and lvl > 0:
                for k in range(len(o.skills[skill][1])):
                    stat, v = o.value(skill, k, lvl)
                    if stat in STATS:
                        want[STATS.index(stat)] += v
        full_bad += g[cid] != want
    if per.get(5):
        print(f"  class 5 (hireling / player defenders, game.exe): totals != base + full values: {full_bad}")
    if vlist:
        gv = {}
        for ln in open(case_path + ".gvals"):
            f = ln.split()
            gv[(int(f[0]), int(f[1]))] = f[2:]
        bad = tot = 0
        zero_first = 0
        for ln in open(vlist):
            f = ln.split()
            if not f or f[0] != "L":
                continue
            sk, srv, lvl, vals = int(f[1]), int(f[2]), int(f[3]), f[4:]
            if srv == 30 and vals:
                st, v = map(int, vals[0].split(":"))
                if st in STATS and v < 0 and int(v / 5) == 0:
                    zero_first += 1
                if v == 0:
                    zero_first += 1
            if (sk, lvl) in gv:
                tot += 1
                bad += vals != gv[(sk, lvl)]
        print(f"  calc values d2d eval_calc vs game (FUN_00645b20 / ln / par), 4 skills x levels 1..99: {tot} rows, {bad} differ")
        print(f"  D1b reach: curse (srvdofunc 30) rows whose first aurastat is 0 after the 1/5 rule, levels 1..99: {zero_first}")


if __name__ == "__main__":
    sys.path.insert(0, os.getcwd())
    cmd = sys.argv[1] if len(sys.argv) > 1 else ""
    if cmd == "gen":
        gen(int(sys.argv[2], 0) & 0xFFFFFFFF, int(sys.argv[3]))
    elif cmd == "game":
        game(sys.argv[2])
    elif cmd == "compare":
        compare(*sys.argv[2:6])
    else:
        sys.exit(__doc__)
