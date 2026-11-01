# SPDX-License-Identifier: GPL-3.0-or-later
"""What game.exe 1.14d writes into a champion's, unique's and minion's stats as a level populates (agent DECOMPMODS).

Runs upstream's room population (monsters.py, objgroups.quest_stubbed) for real
and records, per monster unit, every stat write through the setter FUN_00627260
and the adder FUN_006272b0 (stdcall unit, stat, value, layer), plus the level
FUN_00573cb0 hands FUN_006538a0 when the unit's life / damage / defence / AR
rows are read (0x573f7e: stack +0xc = level). Prints, for every unit that ends
with flag 4 (champion), 8 (unique) or 0x10 (minion) in MonsterData +0x16:

  <class> f<flags> mods <ids> statlvl <level at 6538a0> -> lvl <stat 12> exp <13> maxhp <7>/256 hpcreate <first 7>/256
      dmg% <0x19> th% <0x77> vel% <0x43> dr <0x24> mr <0x25> fr/lr/cr/pr <0x27/0x29/0x2b/0x2d> armor% <0x10> regen <0x4a>

Usage, from upstream d2decomped tools/emu (emu.py, drlg.py, monsters.py, objgroups.py; D2_MPQ_DIR set):
  python umod_spawn.py 1-20 3 [difficulty]      # seeds, Levels.txt id
Never writes game data.
"""
import os
import sys

sys.path.insert(0, os.getcwd())
from unicorn import UC_HOOK_CODE                              # noqa: E402
from unicorn.x86_const import UC_X86_REG_ESI, UC_X86_REG_ESP   # noqa: E402
import drlg                                                   # noqa: E402
import monsters                                               # noqa: E402
import objgroups                                              # noqa: E402

STATS = (0x0c, 0x0d, 0x07, 0x19, 0x77, 0x43, 0x24, 0x25, 0x27, 0x29, 0x2b, 0x2d, 0x10, 0x4a, 0x36, 0x37, 0x30, 0x31, 0x32, 0x33, 0x3e, 0x3f, 0x1f)


def s32(v): return v - (1 << 32) if v >> 31 else v


def trace(e, seed, lid, difficulty):
    writes, statlvl, units = {}, {}, []

    def on_set(mu, addr, size, _):
        sp = mu.reg_read(UC_X86_REG_ESP)
        unit, stat, value = e.r32(sp + 4), e.r32(sp + 8), s32(e.r32(sp + 0xc))
        if e.r32(sp + 0x10) != 0: return
        cur = writes.setdefault(unit, {})
        if addr == 0x627260:
            cur.setdefault(("first", stat), value)
            cur[stat] = value
        else:
            cur[stat] = cur.get(stat, 0) + value

    def on_rows(mu, addr, size, _):
        unit = mu.reg_read(UC_X86_REG_ESI)
        statlvl[unit] = e.s32(mu.reg_read(UC_X86_REG_ESP) + 0xc)
        if unit not in units: units.append(unit)

    hooks = [e.mu.hook_add(UC_HOOK_CODE, on_set, begin=a, end=a) for a in (0x627260, 0x6272b0)]
    hooks.append(e.mu.hook_add(UC_HOOK_CODE, on_rows, begin=0x573f7e, end=0x573f7e))
    try:
        objgroups.quest_stubbed(e, monsters.dump, seed, lid, difficulty)
    finally:
        for h in hooks:
            e.mu.hook_del(h)
    out = []
    for unit in units:
        data = e.r32(unit + 0x14)
        if not data: continue
        flags = e.r16(data + 0x16)
        if not flags & 0x1c: continue
        mods = [b for b in e.read(data + 0x1c, 9) if b]
        w = writes.get(unit, {})
        g = lambda s: w.get(s, 0)
        out.append(f"{e.r32(unit + 4)} f{flags:x} mods {'.'.join(map(str, mods)) or '-'} statlvl {statlvl.get(unit)} -> lvl {g(0xc)} exp {g(0xd)}"
                   f" maxhp {g(7) / 256:.2f} hpcreate {w.get(('first', 7), 0) / 256:.2f} dmg% {g(0x19)} th% {g(0x77)} vel% {g(0x43)}"
                   f" dr {g(0x24)} mr {g(0x25)} fr/lr/cr/pr {g(0x27)}/{g(0x29)}/{g(0x2b)}/{g(0x2d)} armor% {g(0x10)} regen {g(0x4a)}"
                   f" cold {g(0x36)}-{g(0x37)} fire {g(0x30)}-{g(0x31)} ltng {g(0x32)}-{g(0x33)} mana {g(0x3e)}-{g(0x3f)} def {g(0x1f)}")
    return out


def main():
    lo, _, hi = sys.argv[1].partition("-")
    lid = int(sys.argv[2]) if len(sys.argv) > 2 else 3
    difficulty = int(sys.argv[3]) if len(sys.argv) > 3 else 0
    e = drlg.boot()
    for seed in range(int(lo), int(hi or lo) + 1):
        for line in trace(e, seed, lid, difficulty):
            print(f"seed {seed} lvl {lid}: {line}")


if __name__ == "__main__":
    main()
