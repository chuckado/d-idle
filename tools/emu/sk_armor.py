# SPDX-License-Identifier: GPL-3.0-or-later
"""Frozen Armor's striker effect (aura event 2, damagedinmelee) by game.exe's own code.

    python sk_armor.py gen > cases.tsv
    python sk_armor.py game cases.tsv > game.tsv       # FUN_005cac50 natively
    python sk_armor.py direct cases.tsv > direct.tsv   # FUN_0057c6c0 on a record with +0x34 = frames (freeze_rules.py)
    python sk_armor.py model cases.tsv > model.tsv     # d2d's buff_events rule as read (NOT run: a model of the source)
    python sk_armor.py compare cases.tsv game.tsv other.tsv

Runs the real FUN_005cac50 (ECX game; wearer, striker, the striker's hit
record, skill 40, level): it zeroes a 0x70 record, sets +4 |= 0x20 (no
on-hit events: the reflected hit does not fire the striker's own events),
puts calc1 (Skills +0x138) in +0x34 (the freeze length) and calls
FUN_0057c6c0(game, wearer -> striker, flag 1, record): FUN_0057c1e0 cuts
+0x34 by the striker's cold resistance less the wearer's cold pierce (row 6
of 0x732980), then FUN_0057af80 (chill, +0x30 = 0 here) and FUN_0057b230
(freeze: plain monster only at coldeffect < 0, length / MonsterFreezeDivisor;
MonStats boss, monster-data flag 8 or a hireling chilled instead).

Hooked (not the claim): everything freeze_rules.Game hooks, plus
FUN_00646ca0 (the calc: returns the case's frames, i.e. calc1 at the
case's level and synergies, computed from Skills.txt by `frames`) and
FUN_00621e40 (the overlay). The Skills row is a fake one (only +0x138 and
+0x110 = -1 matter to FUN_005cac50).

A case: id difficulty coldeffect res pierce frames kind bcol damage
(kind as freeze_rules: none, champion, unique, superunique, minion; damage
is the striker's hit record +8: 0 = the blow did no damage). Output
`id chill freeze` (ticks to the end of state 0xb / state 1, or `-`).
"""
import os
import random
import sys

sys.path.insert(0, os.getcwd())
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import freeze_rules  # noqa: E402

SKILL = 40
RECORD = 0x23C


def frames(level, synergy_points):
    """Frozen Armor calc1: ln56 x (100 + (Shiver + Chilling blvl) x par8) / 100, par5 30, par6 3, par8 5 (D)."""
    return (30 + 3 * (level - 1)) * (100 + synergy_points * 5) // 100


def gen():
    rows = []
    lens = sorted({frames(lvl, syn) for lvl in range(1, 21) for syn in (0, 1, 5, 20, 40)} | {1, 2, 3, 4})
    reses = (-100, -50, -1, 0, 1, 20, 33, 50, 75, 99, 100, 101, 150)
    for diff in range(3):
        for kind in range(5):
            for bcol in (0, 1):
                for ce in (-100, -50, -1, 0, 1, 25):
                    for res in reses:
                        for n in lens:
                            rows.append((diff, ce, res, 0, n, kind, bcol, 1))
        for res in reses:
            for pierce in (1, 20, 50, 100, 101):
                for n in (30, 33, 87):
                    rows.append((diff, -50, res, pierce, n, 0, 0, 1))
        for n in (30, 87):
            rows.append((diff, -50, 0, 0, n, 0, 0, 0))
    rng = random.Random(7)
    for _ in range(4000):
        rows.append((rng.randrange(3), rng.choice((-50, -33, -25, -100, 0, rng.randint(-128, 127))),
                     rng.choice((0, rng.randint(-100, 150))), rng.choice((0, 0, rng.randint(0, 120))),
                     rng.choice(lens), rng.choice((0, 0, 0, 1, 2, 3, 4)), int(rng.random() < 0.1), 1))
    for i, r in enumerate(rows):
        print(i, *r, sep="\t")


class Armor(freeze_rules.Game):
    def __init__(self):
        super().__init__()
        e = self.e
        self.frames = 0
        tables = e.r32(0x744304)
        self.skills = e.alloc(RECORD * 0x80)
        e.w32(tables + 0xB98, self.skills)
        e.w32(tables + 0xBA0, 0x80)
        e.w32(tables + 0xBC0, 0)
        row = self.skills + SKILL * RECORD
        e.w32(row + 0x138, 0x1234)                    # calc1 index: anything, the calc is hooked
        e.mu.mem_write(row + 0x110, b"\xff\xff")       # no overlay
        self.calc_args = []
        e.hook(0x646CA0, self.calc, 4)
        e.hook(0x621E40, lambda h: 0, 3)
        self.hit = e.alloc(0x80)

    def calc(self, h):
        self.calc_args.append((h.arg(1), h.arg(2), h.arg(3)))
        return self.frames

    def armor(self, cid, diff, ce, res, pierce, n, kind, bcol, damage):
        """As freeze_rules.Game.run, but the record comes from FUN_005cac50."""
        e = self.e
        self.frames = n
        orig = e.call

        def call_cac50(addr, defender, flag, record, ecx, edx):
            e.mu.mem_write(self.hit, b"\0" * 0x80)
            e.w32(self.hit + 8, damage << 8)
            return orig(0x5CAC50, edx, defender, self.hit, SKILL, 1, ecx=ecx)
        e.call = call_cac50
        try:
            return self.run(cid, diff, ce, res, pierce, 0, n, 0, 0, freeze_rules.KIND_FLAGS[kind], bcol)
        finally:
            e.call = orig


def cases(path):
    for line in open(path):
        f = line.split()
        if f:
            yield list(map(int, f))


def game(path):
    g = Armor()
    for cid, diff, ce, res, pierce, n, kind, bcol, damage in cases(path):
        print(cid, *g.armor(cid, diff, ce, res, pierce, n, kind, bcol, damage))
    bad = [a for a in g.calc_args if a[0] != 0x1234 or a[1] != SKILL]
    print(f"# calc calls {len(g.calc_args)}, not (calc1 of skill 40): {len(bad)}", file=sys.stderr)


def direct(path):
    g = freeze_rules.Game()
    for cid, diff, ce, res, pierce, n, kind, bcol, damage in cases(path):
        if damage <= 0:
            print(cid, "-", "-")
            continue
        print(cid, *g.run(cid, diff, ce, res, pierce, 0, n, 0, 0, freeze_rules.KIND_FLAGS[kind], bcol))


def model(path):
    """d2d components/game/fight/states.cpp buff_events (upstream 72132b99), as read: a striker whose blow did
    damage gets stun_until and chill_until = now + max(calc1, 1) frames, any monster kind, any coldeffect, any
    resistance. The chill end is printed as d2d sets it (chill_until); `land`'s coldeffect / divisor rules are
    not applied on this path."""
    for cid, _diff, _ce, _res, _pierce, n, _kind, _bcol, damage in cases(path):
        if damage <= 0:
            print(cid, "-", "-")
            continue
        t = max(n, 1)
        print(cid, t, t)


def compare(case_path, a_path, b_path):
    rows = {r[0]: r for r in cases(case_path)}
    a = {f[0]: f[1:3] for f in (line.split() for line in open(a_path)) if f and f[0].isdigit()}
    b = {f[0]: f[1:3] for f in (line.split() for line in open(b_path)) if f and f[0].isdigit()}
    bad = [k for k in rows if a.get(str(k)) != b.get(str(k))]
    per = {}
    for k in bad:
        per.setdefault(classify(rows[k], a.get(str(k))), []).append(k)
    print(f"cases {len(rows)} a {len(a)} b {len(b)} mismatches {len(bad)}")
    for why, keys in sorted(per.items()):
        print(f"  {len(keys):7d}  {why}")
        for k in keys[:2]:
            print("           ", *rows[k], "a", a.get(str(k)), "b", b.get(str(k)))


def classify(r, got):
    _cid, _diff, ce, res, pierce, _n, kind, bcol, _damage = r
    if freeze_rules.KIND_FLAGS[kind] & 8 or bcol:
        return "special (flag 8 / MonStats boss): chilled, never frozen"
    if res - pierce >= 100 or (res >= 100):
        return "cold immune: nothing"
    if ce >= 0:
        return "coldeffect >= 0: nothing"
    return "plain monster: resisted / MonsterFreezeDivisor, no chill"


if __name__ == "__main__":
    cmd = sys.argv[1] if len(sys.argv) > 1 else ""
    if cmd == "gen" and len(sys.argv) == 2:
        gen()
    elif cmd in ("game", "direct", "model") and len(sys.argv) == 3:
        {"game": game, "direct": direct, "model": model}[cmd](sys.argv[2])
    elif cmd == "compare" and len(sys.argv) == 5:
        compare(*sys.argv[2:5])
    else:
        sys.exit(__doc__)
