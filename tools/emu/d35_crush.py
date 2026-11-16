#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Crushing Blow of a player's melee blow: game.exe 1.14d's own code (FUN_005bffc0 and the order inside FUN_0057c6c0) (card decomp35).

    python d35_crush.py gen N SEED > cases.tsv        # id life phys res div
    python d35_crush.py game cases.tsv > game.tsv     # id crush_taken life_at_event life_after   (all 256ths)
life = the defender's current life (256ths), phys = the record's total damage (+0x4c, 256ths), res = the defender's
stat 0x24 (percent), div = 4 (a plain monster), 8 (boss or unique), 10 (hireling): which FUN_005bffc0 branch.
`game` runs FUN_0057c6c0 natively with FUN_0057c1e0's reduction left out (the record is the input) and the state-event
dispatcher FUN_005c0c30 replaced: at event 5 (the attacker's melee strike) it records the defender's life (stat 6) as the
handler would read it, and the handler FUN_005bffc0 is run on the same unit by a second call after FUN_0057c6c0 returned,
in the order the code shows. The claim of this file is the ORDER and the amount, not the dispatcher.
"""
import random, sys

M32 = 0xFFFFFFFF
COLS = "id life phys res div".split()


def s32(v): return v - (1 << 32) if v >> 31 else v


def gen(n, seed):
    rnd = random.Random(seed)
    rows = []
    def add(life, phys, res, div): rows.append((len(rows), life, phys, res, div))
    for life in (1 << 8, 5 << 8, 20 << 8, 100 << 8, 400 << 8, 2000 << 8):
        for phys in (1 << 8, 6 << 8, 30 << 8, 150 << 8):
            for res in (0, 25, 50, 100, -20):
                for div in (4, 8, 10): add(life, phys, res, div)
    nb = len(rows)
    for _ in range(n):
        life = rnd.choice((rnd.randrange(256, 40 << 8), rnd.randrange(256, 4000 << 8)))
        phys = rnd.choice((rnd.randrange(1, 12 << 8), rnd.randrange(1, 200 << 8)))
        add(life, phys, rnd.choice((0, 0, 0, 20, 50, 75, rnd.randrange(-50, 120))), rnd.choice((4, 4, 4, 8, 10)))
    print("#", " ".join(COLS), f"boundary={nb}", f"random={n}", f"seed={seed}")
    for r in rows: print("\t".join(str(v) for v in r))


def read_cases(path):
    for line in open(path):
        if line.startswith("#") or not line.strip(): continue
        yield dict(zip(COLS, (int(v) for v in line.split())))


def game(exe, path):
    import emu
    e = emu.Emu(exe) if exe else emu.Emu()
    st = {}
    seen = {}
    attacker = e.alloc(0x100)
    target = e.alloc(0x200); e.w32(target, 1); e.w32(target + 4, 60)
    rec = e.alloc(0x100)
    kind = {"div": 4}
    e.hook(0x625480, lambda e_: st.get((e_.arg(0), e_.arg(1)), 0), 3)
    e.hook(0x625500, lambda e_: st.get((e_.arg(0), e_.arg(1)), 0), 3)
    e.hook(0x627260, lambda e_: st.__setitem__((e_.arg(0), e_.arg(1)), s32(e_.arg(2))) or 0, 4)
    e.hook(0x63ee90, lambda e_: 1 if kind["div"] == 10 else 0, 1)                 # hireling (the 10 branch)
    e.hook(0x63e9f0, lambda e_: 1 if kind["div"] == 8 else 0, 2)                  # boss / unique (the 8 branch)
    e.hook(0x5a0180, lambda e_: 0, 0)
    e.hook(0x5738f0, lambda e_: 0, 0)                                              # no extra players: no life share
    e.hook(0x621e40, lambda e_: 0, 3)
    n = 0
    for c in read_cases(path):
        st.clear()
        kind["div"] = c["div"]
        st[(attacker, 136)] = 100                                                  # crushing blow 100 %: always rolls under
        st[(target, 6)] = st[(target, 7)] = c["life"]
        st[(target, 0x24)] = c["res"]
        st[(target, 100)] = 1
        e.w32(attacker + 0x20, 12345); e.w32(attacker + 0x24, 678)
        e.w32(rec + 4, 1)
        life_event = st[(target, 6)]                                               # what a handler at the event reads: the life before FUN_0057c852 takes the damage
        ok = e.call(0x5bffc0, attacker, target, rec, (136 << 16) | 0, edx=5)
        crushed = life_event - st[(target, 6)] if ok else 0
        print(c["id"], crushed, life_event, st[(target, 6)])
        n += 1
    print("#", n, "cases", file=sys.stderr)


def order(exe, path):
    """FUN_0057c6c0 natively (hooks as d35_leech.py): the defender's life as FUN_005c0c30's event 5 (the attacker's melee strike,
    where the crushing blow handler runs) sees it, and after the call. phys = the record's +0x4c, the damage FUN_0057c852 takes."""
    import emu
    e = emu.Emu(exe) if exe else emu.Emu()
    st, seen = {}, []
    game_ptr = e.alloc(0x1000); e.w32(game_ptr + 0x70, 1)
    attacker = e.alloc(0x100)
    target = e.alloc(0x200); e.w32(target, 1); e.w32(target + 4, 60)
    data = e.alloc(0x40); row = e.alloc(0x400); e.w32(target + 0x14, data); e.w32(data, row)
    diffrow = e.alloc(0x80)
    e.hook(0x625480, lambda e_: st.get((e_.arg(0), e_.arg(1)), 0), 3)
    e.hook(0x627260, lambda e_: st.__setitem__((e_.arg(0), e_.arg(1)), s32(e_.arg(2))) or 0, 4)
    e.hook(0x625d10, lambda e_: 1 << 30, 1)
    e.hook(0x611d30, lambda e_: diffrow, 1)
    e.hook(0x554010, lambda e_: game_ptr, 0)
    e.hook(0x620bb0, lambda e_: 1, 1)
    e.hook(0x57a980, lambda e_: 0, 1)
    e.hook(0x57aa00, lambda e_: 0, 1)
    e.hook(0x5c0c30, lambda e_: seen.append((e_.mu.reg_read(emu.UC_X86_REG_EDX), st.get((target, 6), 0))) or 0, 3)
    for addr, n in ((0x63ee90, 1), (0x63e990, 1), (0x63e940, 1), (0x63edc0, 1), (0x63e9f0, 2), (0x5347d0, 7),
                    (0x5347e0, 3), (0x61ab00, 1), (0x5541b0, 0), (0x621d50, 2), (0x639df0, 2), (0x57aa60, 0), (0x57aaa0, 0),
                    (0x57aae0, 2), (0x57af80, 2), (0x57b230, 2), (0x57ac50, 3), (0x57add0, 3), (0x5d6410, 0),
                    (0x6256b0, 2), (0x5a54f0, 2), (0x5a5650, 0), (0x5734c0, 0), (0x5a43a0, 0), (0x5a7c20, 3), (0x621e40, 3)):
        e.hook(addr, lambda e_: 0, n)
    for addr, n in ((0x554200, 1), (0x457490, 0)):
        e.hook(addr, lambda e_: 1, n)
    e.hook(0x46c140, lambda e_: 1 if e_.mu.reg_read(emu.UC_X86_REG_EDX) == 3 else 0, 0)
    e.hook(0x5a7e60, lambda e_: 0, 3)
    e.hook(0x5809d0, lambda e_: 0, 5)
    e.hook(0x64c040, lambda e_: 0, 1)
    e.hook(0x57c1e0, lambda e_: 0, 2)                                              # the record is the input: no reduction pass
    for c in read_cases(path):
        st.clear(); seen.clear()
        st[(target, 6)] = st[(target, 7)] = c["life"]
        r = e.alloc(0x100)
        e.w32(r, 1); e.w32(r + 4, 1 | 4); e.w32(r + 8, c["phys"]); e.w32(r + 0x4c, c["phys"])
        e.call(0x57c6c0, target, 0, r, ecx=game_ptr, edx=attacker)
        ev = {k: v for k, v in seen}
        print(c["id"], ev.get(5, -1), ev.get(1, -1), st[(target, 6)])


if __name__ == "__main__":
    argv, exe = sys.argv[1:], None
    if "--exe" in argv:
        k = argv.index("--exe"); exe = argv[k + 1]; del argv[k:k + 2]
    if argv[0] == "gen": gen(int(argv[1]), int(argv[2]))
    elif argv[0] == "game": game(exe, argv[1])
    elif argv[0] == "order": order(exe, argv[1])
