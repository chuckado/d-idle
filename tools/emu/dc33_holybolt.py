# SPDX-License-Identifier: GPL-3.0-or-later
"""game.exe's Holy Bolt heal as the oracle: FUN_005a9fb0 (hit function 7) on a pet of the caster.

    set -o pipefail; ../../build/tests/test_holy_bolt_heal --heal | grep -E '^(H|# skill) ' > d2d.txt
    uv run python holybolt_heal.py compare d2d.txt         # last line: N cells  same, or N cells  M differ (exit 1)

A Holy Bolt that strikes a unit with sHitPar1 set: FUN_005542c0 (TargetPet: the struck unit is a monster whose owner is the bolt's owner) or FUN_00554d20
(TargetAlly: a party member) says yes, and the unit gets `FUN_00646ca0(calc1) << 8 + roll(FUN_00646ca0(calc2) << 8 - calc1 << 8)` (FUN_0045c3e0 on
the missile's seed at +0x20), capped at its maximum life (FUN_00625d10), written back as stat 6 (FUN_00627260); a roll of 0 or less takes
FUN_005a8910's missile stat instead (0 here). `d2d.txt` is test_holy_bolt_heal's grid: `H level d0 d1 life max | life_after` (256ths; the seed's two
words), d2d's Fight::skill_missile_hits on a pet of the caster. The game's unit gets the same seed and the same life and maximum.

Runs natively: FUN_005a9fb0, the calc machine FUN_00646ca0, FUN_00645ca0's table lookups it needs, FUN_0045c4b0 (the skill record), FUN_0045c3e0 (the roll).
Hooked (not the claim): FUN_00552fd0 (the missile's owner), FUN_0064a280 / FUN_0064a210 (the missile's skill and level), FUN_005542c0 (the pet test: yes; its
rule is read, not run), FUN_00625480 (stat 6: the case's life; the owner's stats: none), FUN_00625d10 (max life), FUN_00627260 (the write: recorded),
FUN_00621e40 (the overlay, a cue).
"""
import sys

import skill_damage

M32 = 0xFFFFFFFF


def s32(v):
    v &= M32
    return v - (1 << 32) if v >> 31 else v


class Game:
    def __init__(self, skill):
        self.o = o = skill_damage.Oracle()
        e = self.e = o.e
        self.skill = skill
        self.level = 1
        self.life = self.maxlife = 0
        self.written = []
        tables = e.r32(0x744304)
        self.missiles, self.mcount = e.r32(tables + 0xb64), e.r32(tables + 0xb6c)
        self.missile_row = self.expected_row()
        if e.r32(self.missiles + 0x1a4 * self.missile_row + 0x4c) == 0:
            raise SystemExit("the holybolt row has no sHitPar1: the heal is not reached")
        self.owner, self.target, self.missile = (e.alloc(0x300) for _ in range(3))
        e.hook(0x625480, lambda e_: (self.life if (e_.arg(0) == self.target and e_.arg(1) == 6) else o.stats.get(e_.arg(1), 0)) & M32, 3)
        e.hook(0x552fd0, lambda e_: self.owner, 0)
        e.hook(0x64a280, lambda e_: self.skill, 1)
        e.hook(0x64a210, lambda e_: self.level, 1)
        e.hook(0x5542c0, lambda e_: 1, 2)
        e.hook(0x625d10, lambda e_: self.maxlife, 1)
        e.hook(0x627260, lambda e_: self.written.append((e_.arg(1), s32(e_.arg(2)))) or 0, 4)
        e.hook(0x621e40, lambda e_: 0, 3)

    def expected_row(self):
        import emu
        text = emu.mpq_cache().read(r"data\global\excel\missiles.txt").decode("latin1").split("\n")
        head = text[0].rstrip("\r").split("\t")
        col = head.index("Missile")
        for i, line in enumerate(text[1:]):
            f = line.rstrip("\r").split("\t")
            if len(f) > col and f[col] == "holybolt":
                return i
        raise SystemExit("holybolt is not in Missiles.txt")

    def heal(self, level, d0, d1, life, maxlife):
        e, o = self.e, self.o
        self.level, self.life, self.maxlife = level, life, maxlife
        self.written.clear()
        o.levels = {self.skill: level}
        o.stats = {}
        e.mu.mem_write(self.missile, bytes(0x300))
        e.w32(self.missile + 4, self.missile_row)
        e.w32(self.missile + 0x20, d0)
        e.w32(self.missile + 0x24, d1)
        e.mu.mem_write(self.owner, bytes(0x300))
        e.mu.mem_write(self.target, bytes(0x300))
        ret = e.call(0x5a9fb0, self.target, ecx=0, edx=self.missile)
        if ret != 1:
            raise SystemExit("FUN_005a9fb0 returned %d, not 1" % ret)
        writes = [v for stat, v in self.written if stat == 6]
        return writes[0] if writes else life


def compare(path):
    skill = None
    rows = []
    for lineno, line in enumerate(open(path, encoding="latin1"), 1):
        p = line.split()
        if p[:2] == ["#", "skill"]:
            skill = int(p[2])
        if not p or p[0] != "H":
            continue
        if len(p) != 8 or p[6] != "|":
            raise SystemExit("%s:%d: malformed H line (%d fields)" % (path, lineno, len(p)))
        try:
            rows.append((lineno, [int(x) for x in p[1:6]], int(p[7])))
        except ValueError:
            raise SystemExit("%s:%d: malformed H line (a number is not one)" % (path, lineno)) from None
    if skill is None:
        raise SystemExit("%s: no '# skill N' line (keep it: grep -E '^(H|# skill) ')" % path)
    g = Game(skill)
    cells = bad = 0
    shown = 0
    for lineno, (level, d0, d1, life, maxlife), after in rows:
        got = g.heal(level, d0, d1, life, maxlife)
        cells += 1
        if got != after:
            bad += 1
            if shown < 12:
                shown += 1
                print("C level %d seed %d %d life %d max %d | d2d %d | game %d" % (level, d0, d1, life, maxlife, after, got))
    print("%d cells  %s" % (cells, "same" if not bad else "%d differ" % bad))
    if cells == 0:
        raise SystemExit("no comparable cells in %s: an empty file is not a pass" % path)
    return 1 if bad else 0


def main():
    if len(sys.argv) > 2 and sys.argv[1] == "compare":
        raise SystemExit(compare(sys.argv[2]))
    raise SystemExit(__doc__)


if __name__ == "__main__":
    main()
