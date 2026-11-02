# SPDX-License-Identifier: GPL-3.0-or-later
"""What game.exe 1.14d gives EVERY monster unit of a populated level (agent monsterstats).

A copy of umod_spawn.py's trace that keeps the plain monsters too: upstream's room
population (monsters.py, objgroups.quest_stubbed) runs for real and every stat write
through the setter FUN_00627260 and the adder FUN_006272b0 is recorded per unit, plus
the level FUN_00573cb0 hands FUN_006538a0 (0x573f7e: stack +0xc).

  python ms_spawn.py LO-HI LEVELIDS [difficulty]   # seeds, comma list of Levels.txt ids
  line: `unit lid seed class flags statlvl stat:value stat:value ...`   (stats in hex, final values;
        7 and 8 / 9 in 256ths; `f7` = the first life written)
Never writes game data.
"""
import os
import sys

sys.path.insert(0, os.getcwd())
from unicorn import UC_HOOK_CODE                              # noqa: E402
from unicorn.x86_const import UC_X86_REG_EBP, UC_X86_REG_ESI, UC_X86_REG_ESP   # noqa: E402
import drlg                                                   # noqa: E402
import monsters                                               # noqa: E402
import objgroups                                              # noqa: E402


SEQ = {int(x, 16) for x in (sys.argv[4] if len(sys.argv) > 4 else "").split(",") if x}   # 4th arg: stats whose write sequence is printed (hex, comma list)


def s32(v): return v - (1 << 32) if v >> 31 else v


def trace(e, seed, lid, difficulty):
    writes, statlvl, units = {}, {}, []

    def on_set(mu, addr, size, _):
        sp = mu.reg_read(UC_X86_REG_ESP)
        unit, stat, value = e.r32(sp + 4), e.r32(sp + 8), s32(e.r32(sp + 0xc))
        if e.r32(sp + 0x10) != 0: return
        cur = writes.setdefault(unit, {})
        if SEQ and stat in SEQ:
            cur.setdefault("seq", []).append(f"{'add' if addr == 0x6272b0 else 'set'}{stat:x}={value}@{e.r32(sp):x}")
        if addr == 0x627260:
            cur.setdefault(("first", stat), value)
            cur[stat] = value
        else:
            cur[stat] = cur.get(stat, 0) + value

    def on_rows(mu, addr, size, _):
        unit = mu.reg_read(UC_X86_REG_ESI)
        statlvl[unit] = e.s32(mu.reg_read(UC_X86_REG_ESP) + 0xc)
        if unit not in units: units.append(unit)

    def on_game(mu, addr, size, _):
        # FUN_00573cb0 tests game byte +0x6a (expansion): the harness game has it 0 (classic NM / Hell columns).
        # A real LoD game has it 1: set it (arg [ebp + 8] = the game).
        game = e.r32(mu.reg_read(UC_X86_REG_EBP) + 8)
        mu.mem_write(game + 0x6a, b"\x01")

    hooks = [e.mu.hook_add(UC_HOOK_CODE, on_game, begin=0x573f41, end=0x573f41)]
    hooks += [e.mu.hook_add(UC_HOOK_CODE, on_set, begin=a, end=a) for a in (0x627260, 0x6272b0)]
    hooks.append(e.mu.hook_add(UC_HOOK_CODE, on_rows, begin=0x573f7e, end=0x573f7e))
    try:
        objgroups.quest_stubbed(e, monsters.dump, seed, lid, difficulty)
    finally:
        for h in hooks:
            e.mu.hook_del(h)
    out = []
    for n, unit in enumerate(units):
        data = e.r32(unit + 0x14)
        if not data: continue
        flags = e.r16(data + 0x16)
        w = writes.get(unit, {})
        stats = " ".join(f"{k:x}:{v}" for k, v in sorted((k, v) for k, v in w.items() if isinstance(k, int)))
        if SEQ: stats += " seq " + ",".join(w.get("seq", []))
        first = w.get(("first", 7), 0)
        out.append(f"{n} {e.r32(unit + 4)} {flags:x} {statlvl.get(unit)} f7:{first} {stats}")
    return out


def main():
    lo, _, hi = sys.argv[1].partition("-")
    lids = [int(x) for x in sys.argv[2].split(",")]
    difficulty = int(sys.argv[3]) if len(sys.argv) > 3 else 0
    e = drlg.boot()
    for lid in lids:
        for seed in range(int(lo), int(hi or lo) + 1):
            try:
                lines = trace(e, seed, lid, difficulty)
            except Exception as err:
                print(f"# lid {lid} seed {seed} failed: {err}")
                continue
            for line in lines:
                print(f"U {lid} {seed} {line}")


if __name__ == "__main__":
    main()
