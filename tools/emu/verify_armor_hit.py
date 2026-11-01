# SPDX-License-Identifier: GPL-3.0-or-later
"""VERIFYARMORHIT's own oracle: Frozen Armor's answer to a melee striker by game.exe's code, with end frames.

    python verify_armor_hit.py gen SEED N [edge] > cases.tsv
    python verify_armor_hit.py game cases.tsv > game.tsv
    python verify_armor_hit.py compare cases.tsv game.tsv d2d.tsv [...pairs]

Run for real: FUN_005cac50, FUN_0057c6c0, FUN_0057c1e0 (FUN_0057c060, FUN_0057c140, FUN_0057bf80,
FUN_0057be00, FUN_00483360, the table at 0x732980), FUN_0057af80, FUN_0057b230, FUN_0057af30,
FUN_0063e9f0, FUN_005a0180, FUN_0063ee90, and at the state-1 list's expiry the freeze-end callback
FUN_0057b170 (which reschedules the AI event type 2 at expiry + MonStats aidel).
Hooked (not the claim): stat get/set, DifficultyLevels row (from the MPQ), MonStats row (fake: boss
byte +0xc, aidel +0x4f..0x51, coldeffect +0x168..), the skill calc FUN_00646ca0 (calc1 from Skills.txt's
own text: ln56 * (100 + (Shiver + Chilling blvl) * par8) / 100, ln56 = par5 + (lvl - 1) par6), states,
stat lists, timers (recorded, then replayed: a type-2 event while state 1 is on is dropped, FUN_005a7f80),
event dispatch, logging, the other record appliers (fields all 0).
Case: id diff kind ce res pierce level shiver chilling phys aidel prior_freeze prior_chill
kind: 0 plain 1 champion 2 unique 3 superunique 4 minion 5 MonStats boss column.
Output: id chill_end freeze_end ai_resume (frames from now, '-' none).
"""
import os
import struct
import sys

import emu as emu_module

FRAME = 2000
SKILL = 40
FLAGS = {0: 0, 1: 0xD, 2: 0x8, 3: 0xA, 4: 0x10, 5: 0}   # champion | 8 | 5 (FUN_005a48c0), unique | 8, superunique | 2 | 8, minion | 0x10
CLASS = 5                                             # zombie1's row index slot (fake row: only our fields)


class Lcg:
    def __init__(self, seed):
        self.x = seed & 0xFFFFFFFF

    def next(self, n):
        self.x = (self.x * 1664525 + 1013904223) & 0xFFFFFFFF
        return (self.x >> 8) % n

    def pick(self, seq):
        return seq[self.next(len(seq))]


def gen(seed, n, edge):
    r = Lcg(seed)
    rows = []
    if edge:
        for diff in range(3):
            for kind in range(6):
                for ce in (-50, -1, 0, 1):
                    for res in (-150, -100, -99, -1, 0, 1, 50, 99, 100, 101, 150):
                        for pierce in (0, 1, 50, 99, 100, 150):
                            for level in (1, 20, 40):
                                rows.append((diff, kind, ce, res, pierce, level, 0, 0, 1, 15, 0, 0))
                for phys in (0, 1):
                    for prior in (0, 1, 5, 200):
                        rows.append((diff, kind, -50, 0, 0, 10, 3, 4, phys, 9, prior, prior))
    for _ in range(n):
        diff = r.next(3)
        kind = r.pick((0, 0, 0, 1, 2, 3, 4, 5))
        ce = r.pick((-50, -40, -33, -25, -100, 0, 0, r.next(256) - 128))
        res = r.pick((0, 0, r.next(251) - 100, r.next(251) - 100, r.next(21) + 90, r.next(101) - 200))
        pierce = r.pick((0, 0, 0, r.next(60) + 1, r.next(200)))
        level = r.next(40) + 1
        shiver = r.pick((0, 0, r.next(21), r.next(41)))
        chilling = r.pick((0, 0, r.next(21), r.next(41)))
        phys = r.pick((1, 1, 1, 0, r.next(500)))
        aidel = r.pick((15, 14, 13, 0, r.next(30)))
        prior_freeze = r.pick((0, 0, 0, r.next(80) + 1))
        prior_chill = r.pick((0, 0, 0, r.next(200) + 1))
        rows.append((diff, kind, ce, res, pierce, level, shiver, chilling, phys, aidel, prior_freeze, prior_chill))
    for i, row in enumerate(rows):
        print(i, *row, sep="\t")


def calc1(level, shiver, chilling):
    ln56 = 30 + (level - 1) * 3
    return ln56 * (100 + (shiver + chilling) * 5) // 100


class Oracle:
    def __init__(self):
        e = self.e = emu_module.Emu()
        text = e.mpq_file("data\\global\\excel\\DifficultyLevels.txt").decode("latin-1").splitlines()
        head = text[0].split("\t")
        self.levels = [dict(zip(head, line.split("\t"))) for line in text[1:4]]
        self.boss_mask = e.r32(0x6CE280) & 0xFF
        self.game, self.wearer, self.striker = e.alloc(0x200), e.alloc(0x200), e.alloc(0x200)
        self.mdata, self.diffrec, self.hit = e.alloc(0x100), e.alloc(0x100), e.alloc(0x80)
        self.monstats = e.alloc(0x1A8 * 0x200)
        self.skills = e.alloc(0x23C * 0x80)
        tables = e.alloc(0x1000)
        e.w32(0x744304, tables)
        e.w32(tables + 0xA78, self.monstats)
        e.w32(tables + 0xA80, 0x200)
        e.w32(tables + 0xB98, self.skills)
        e.w32(tables + 0xBA0, 0x80)
        e.w32(tables + 0xBC0, 0)
        e.w32(self.skills + SKILL * 0x23C + 0x138, 0x77)
        e.mu.mem_write(self.skills + SKILL * 0x23C + 0x110, b"\xff\xff")
        e.w32(self.wearer, 0)
        e.w32(self.wearer + 0xC, 1)
        e.w32(self.striker, 1)
        e.w32(self.striker + 4, CLASS)
        e.w32(self.striker + 0xC, 77)
        e.w32(self.striker + 0x14, self.mdata)
        e.w32(self.mdata, self.monstats + 0x1A8 * CLASS)
        e.w32(self.game + 0x70, 1)
        e.mu.mem_write(self.game + 0x6A, b"\1")
        self.calc_value = 0
        self.calc_calls = []
        self._hooks()

    def _hooks(self):
        import unicorn.x86_const as x86
        e = self.e
        reg = lambda h, r: h.mu.reg_read(r)

        def stat(h):
            return self.stats.get((h.arg(0), h.arg(1)), 0) & 0xFFFFFFFF

        def new_list(h):
            a = h.alloc(0x40)
            self.lists[a] = {"state": None, "end": h.arg(2), "cb": 0, "unit": None}
            return a

        def set_state_list(h):
            self.lists[h.arg(0)]["state"] = h.arg(1)
            return 0

        def set_cb(h):
            self.lists[h.arg(0)]["cb"] = h.arg(1)
            return 0

        def attach(h):
            lst = self.lists[h.arg(1)]
            lst["unit"] = h.arg(0)
            self.by_unit[(h.arg(0), lst["state"])] = h.arg(1)
            return 0

        def set_end(h):
            self.lists[h.arg(0)]["end"] = h.arg(1)
            return 0

        def state_on(h):
            unit, state, on = h.arg(0), h.arg(1), h.arg(2)
            (self.states.add if on else self.states.discard)((unit, state))
            return 0

        def has_state(h):
            return int((h.arg(0), h.arg(1)) in self.states)

        def timer(h):                      # FUN_005417d0(ecx game, edx unit, type, frame, 0, 0)
            self.timers.append((h.arg(1), len(self.timers), h.arg(0), reg(h, x86.UC_X86_REG_EDX)))
            return 0

        def untimer(h):                    # FUN_00540e60(ecx game, edx unit, type, arg)
            unit, kind = reg(h, x86.UC_X86_REG_EDX), h.arg(0)
            self.timers = [t for t in self.timers if not (t[2] == kind and t[3] == unit)]
            return 0

        def calc(h):
            self.calc_calls.append((h.arg(1), h.arg(2), h.arg(3)))
            return self.calc_value

        for addr, n in ((0x621D50, 2), (0x57C420, 1), (0x57A980, 1), (0x57AA60, 0), (0x57AAA0, 0), (0x57AAE0, 2),
                        (0x57AC50, 3), (0x57ADD0, 3), (0x5D6410, 0), (0x5C0C30, 3), (0x5347D0, 7), (0x5347E0, 3),
                        (0x61AB00, 1), (0x5541B0, 0), (0x6270B0, 4), (0x623F50, 3), (0x627260, 4), (0x5A4390, 0),
                        (0x621E40, 3), (0x553570, 1), (0x63A4A0, 2), (0x4638A0, 0)):
            e.hook(addr, lambda h: 0, n)
        for addr, n in ((0x554200, 1), (0x620BB0, 1), (0x457490, 0)):
            e.hook(addr, lambda h: 1, n)
        for addr, fn, n in ((0x625480, stat, 3), (0x625500, stat, 3), (0x611D30, lambda h: self.diffrec, 1),
                            (0x451F80, lambda h: self.monstats + 0x1A8 * reg(h, x86.UC_X86_REG_ECX), 0),
                            (0x554010, lambda h: self.game, 0),
                            (0x6256B0, lambda h: self.by_unit.get((h.arg(0), h.arg(1)), 0), 2),
                            (0x6251F0, new_list, 5), (0x6252D0, set_state_list, 2), (0x625CE0, set_cb, 2),
                            (0x626E10, attach, 3), (0x625330, lambda h: self.lists[h.arg(0)]["end"], 1),
                            (0x625310, set_end, 2), (0x639DB0, state_on, 3), (0x639DF0, has_state, 2),
                            (0x5417D0, timer, 4), (0x540E60, untimer, 2), (0x646CA0, calc, 4)):
            e.hook(addr, fn, n)

    def run(self, diff, kind, ce, res, pierce, level, shiver, chilling, phys, aidel, prior_freeze, prior_chill):
        e = self.e
        lv = self.levels[diff]
        e.mu.mem_write(self.game + 0x6D, bytes([diff]))
        e.w32(self.game + 0xA8, FRAME)
        e.w32(self.diffrec, int(lv["ResistPenalty"]) & 0xFFFFFFFF)
        e.w32(self.diffrec + 0x14, int(lv["MonsterFreezeDivisor"]))
        e.w32(self.diffrec + 0x18, int(lv["MonsterColdDivisor"]))
        row = self.monstats + 0x1A8 * CLASS
        e.mu.mem_write(row, b"\0" * 0x1A8)
        e.mu.mem_write(row + 0xC, bytes([self.boss_mask if kind == 5 else 0]))
        e.mu.mem_write(row + 0x4F, bytes([aidel] * 3))
        e.mu.mem_write(row + 0x168, struct.pack("<bbb", ce, ce, ce))
        e.mu.mem_write(self.mdata + 0x16, struct.pack("<H", FLAGS[kind]))
        e.w32(self.striker + 0x20, 0x9E3779B9)
        e.w32(self.striker + 0x24, 0)
        self.stats = {(self.striker, 6): 1 << 28, (self.striker, 0x2B): res & 0xFFFFFFFF,
                      (self.wearer, 0x14F): pierce}
        self.lists, self.by_unit, self.states, self.timers = {}, {}, set(), []
        for state, prior in ((1, prior_freeze), (0xB, prior_chill)):
            if prior:
                a = e.alloc(0x40)
                self.lists[a] = {"state": state, "end": FRAME + prior, "cb": 0x57B170 if state == 1 else 0x57AD80, "unit": self.striker}
                self.by_unit[(self.striker, state)] = a
                self.states.add((self.striker, state))
                self.timers.append((FRAME + prior, len(self.timers), 0xC, self.striker))
        self.calc_value = calc1(level, shiver, chilling)
        e.mu.mem_write(self.hit, b"\0" * 0x80)
        e.w32(self.hit + 8, phys)
        e.call(0x5CAC50, self.wearer, self.striker, self.hit, SKILL, level, ecx=self.game)
        a1 = self.by_unit.get((self.striker, 0xB))
        a2 = self.by_unit.get((self.striker, 1))
        chill = self.lists[a1]["end"] - FRAME if a1 and self.lists[a1]["end"] > FRAME else None
        freeze = self.lists[a2]["end"] - FRAME if a2 and self.lists[a2]["end"] > FRAME else None
        zero = a2 is not None and self.lists[a2]["end"] <= FRAME
        resume = self.replay() if a2 is not None else None
        return chill, freeze, resume, zero

    def replay(self):
        """Frames after now: group-1 timers of each frame in insertion order (a timer filed for a past frame runs next frame: U)."""
        e = self.e
        for t in range(FRAME + 1, FRAME + 4000):
            due = sorted(x for x in self.timers if x[0] <= t)
            for item in due:
                if item not in self.timers:
                    continue
                self.timers.remove(item)
                _frame, _seq, kind, unit = item
                if kind == 0xC:
                    for (u, state), a in list(self.by_unit.items()):
                        lst = self.lists[a]
                        if u == unit and lst["end"] <= t:
                            del self.by_unit[(u, state)]
                            if lst["cb"] == 0x57B170:
                                e.w32(self.game + 0xA8, t)
                                e.call(0x57B170, a, ecx=unit, edx=state)
                            else:
                                self.states.discard((u, state))
                elif kind == 2:
                    if (unit, 1) in self.states:
                        continue                       # FUN_005a7f80: frozen and alive -> dropped
                    return t - FRAME
        return None


def game(path):
    o = Oracle()
    for line in open(path):
        f = line.split()
        if not f:
            continue
        cid, *args = map(int, f)
        chill, freeze, resume, zero = o.run(*args)
        s = lambda v: "-" if v is None else str(v)
        print(cid, s(chill), s(freeze), s(resume), "Z" if zero else "")
    bad = [c for c in o.calc_calls if c[0] != 0x77 or c[1] != SKILL]
    print(f"# calc calls {len(o.calc_calls)} bad {len(bad)}", file=sys.stderr)


def compare(case_path, pairs):
    rows = {int(f[0]): list(map(int, f[1:])) for f in (l.split() for l in open(case_path)) if f}
    for game_path, d2d_path in pairs:
        g = {int(f[0]): f[1:] for f in (l.split() for l in open(game_path)) if f and f[0].isdigit()}
        d = {int(f[0]): f[1:] for f in (l.split() for l in open(d2d_path)) if f and f[0].isdigit()}
        n = len(rows)
        cnt = {"len": 0, "endframe": 0, "resume": 0, "zero_len": 0, "calc": 0, "missing": 0, "phys0_any": 0}
        by_kind = {}
        deltas = {}
        examples = []
        for k, r in rows.items():
            if k not in g or k not in d:
                cnt["missing"] += 1
                continue
            gc, gf, gr = g[k][0], g[k][1], g[k][2]
            dc, df, dcalc = d[k][0], d[k][1], int(d[k][2])
            if dcalc != calc1(r[5], r[6], r[7]):
                cnt["calc"] += 1
            if len(g[k]) > 3 and g[k][3] == "Z":
                cnt["zero_len"] += 1
            if r[8] == 0 and (dc != "-" or df != "-"):
                cnt["phys0_any"] += 1
            if (gc, gf) != (dc, df):
                cnt["len"] += 1
                by_kind[r[1]] = by_kind.get(r[1], 0) + 1
                if len(examples) < 4:
                    examples.append((k, r, (gc, gf), (dc, df)))
            # resume: game AI event frame vs d2d's first unstunned frame (stun end, ai.cpp:1831)
            if gr != "-" or df != "-":
                dres = df
                if gr != dres:
                    cnt["resume"] += 1
                    key = "-" if "-" in (gr, dres) else int(gr) - int(dres)
                    deltas[key] = deltas.get(key, 0) + 1
        print(f"{game_path} vs {d2d_path}: cases {n}", cnt)
        print("  length/end differs by kind (0 plain 1 champ 2 uniq 3 su 4 minion 5 bosscol):", dict(sorted(by_kind.items())))
        print("  resume delta (game - d2d stun end) top:", sorted(deltas.items(), key=lambda kv: -kv[1])[:8])
        for ex in examples:
            print("  e.g.", ex)


if __name__ == "__main__":
    sys.path.insert(0, os.getcwd())
    cmd = sys.argv[1] if len(sys.argv) > 1 else ""
    if cmd == "gen":
        gen(int(sys.argv[2], 0), int(sys.argv[3]), len(sys.argv) > 4)
    elif cmd == "game":
        game(sys.argv[2])
    elif cmd == "compare":
        a = sys.argv[3:]
        compare(sys.argv[2], list(zip(a[0::2], a[1::2])))
    else:
        sys.exit(__doc__)
