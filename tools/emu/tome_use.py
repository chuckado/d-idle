#!/usr/bin/env python3
"""game.exe 1.14d's town portal item use as the oracle (d2lod-bot, card decomptome).

The server's C->S 0x20 path runs natively: the dispatcher FUN_0054d750 (its busy
test FUN_0057eec0: dead, mode 0 / 0x11, or state 0x36 refuses), the handler
FUN_0054b1e0 (FUN_00549150 owner, FUN_00548ef0 range), FUN_0055e170 (use an
inventory item), all under unicorn through upstream's harness
(psi29a/d2decomped tools/emu/emu.py) with FUN_005678a0, FUN_005bf240 (the pSpell
dispatch, table 0x741790), FUN_005be290 (pSpell 2: town portal), FUN_0055e050 and
FUN_005be1c0 native too. Hooked to the case: the item lookup (FUN_00552f60), the
cursor item (FUN_0063c1e0), the item type (FUN_0062b400 / FUN_00629bb0: 0x12 book,
0x16 scroll), the quantity stat 0x46 (FUN_00625480), the town test (FUN_0061ab00),
the portal object's free spot (FUN_0056d130's result). Recorded: the old pair closed
(FUN_00535430), the portal opened (FUN_0056d130), the stat change (FUN_006272b0),
the scroll removed (FUN_0055e000), the voice events (FUN_00553380).

d2d's rule is a Python copy of components/game/world.cpp `World::read_portal` +
fight.cpp `Fight::cast_scroll` + `World::open_portal_at` (upstream main f8aab0e8)
and apps/d2d/town.cpp:1596 (a right-click with a held item sends nothing).

    cd <d2decomped>/tools/emu && source /tmp/casttrace-venv/bin/activate
    python tome_use.py --exe game.exe [random N] [seed]

Outcome per case: (now, later, spent, closed): a portal opened at the click, a
portal due at the SC action frame, the charge (book) or the scroll (scroll) spent
at the click, the old pair closed. Prints one line per class of case.
"""
import random
import sys
from collections import Counter

import emu
from unicorn.x86_const import UC_X86_REG_EDX

BOOK, SCROLL = 0x12, 0x16
WALK_MODES = (1, 2, 3, 5, 6)                    # NU WL RN TN TW: d2d's pmode < 0 (no attack, cast, GH, BL, KB, SQ)
MODES = list(range(20))                         # player modes 0 DT .. 0x11 DD .. 0x13


def d2d_use(kind, qty, town, cursor, mode, spot, state54=0):
    """apps/d2d/town.cpp:1596 `!held`; world.cpp read_portal: in_town / dead / pmode >= 0 -> nothing;
    tbk with quantity <= 0 -> nothing; cast_scroll (SC); --quantity or erase at once; the portal at
    the SC action frame (open_portal_at: no free spot -> nothing, the old pair kept)."""
    if cursor or town or mode not in WALK_MODES: return (0, 0, 0, 0)  # dead (DT, DD) included: fight.dead()
    if kind == BOOK and qty <= 0: return (0, 0, 0, 0)
    return (0, 1 if spot else 0, 1, 1 if spot else 0)


def run(exe, n_random, seed):
    e = emu.Emu(exe) if exe else emu.Emu()
    cur, out = {}, {}

    game = e.alloc(0x1000)
    player = e.alloc(0x200)
    inv = e.alloc(0x100)
    item = e.alloc(0x200)
    books = e.alloc(0x40)
    items_row = e.alloc(0x200)
    room = e.alloc(0x100)
    portal = e.alloc(0x100)
    client = e.alloc(0x100)
    packet = e.alloc(16)
    path = e.alloc(0x100)
    e.mu.mem_write(packet, bytes([0x20]) + (77).to_bytes(4, "little") + (100).to_bytes(4, "little") + (200).to_bytes(4, "little"))
    e.w32(player + 0x2c, path)
    e.w32(player, 0); e.w32(player + 0xc, 1); e.w32(player + 0x60, inv)
    e.w32(item, 4); e.w32(item + 4, 0x200); e.w32(item + 0xc, 77); e.w32(item + 0x10, 0)
    e.w32(books + 4, 2); e.w32(books + 8, 219); e.w32(books + 0xc, 220)   # books.txt: pSpell 2, ScrollSkill, BookSkill
    e.w32(items_row + 0x94, 2)
    e.w32(portal, 2); e.w32(portal + 4, 0x3b); e.w32(portal + 0xc, 900)

    def ev(name):
        return lambda e_: out.__setitem__(name, out.get(name, 0) + 1) or 0

    e.hook(0x55bf50, lambda e_: 0, 1)                                    # trace line
    e.hook(0x639df0, lambda e_: cur["state54"] if e_.arg(1) == 0x36 else 0, 2)   # the player's states
    e.hook(0x63ad50, lambda e_: 1, 2)                                    # the item is in the player's inventory
    e.hook(0x6488c0, lambda e_: 100, 1)                                  # the player's x, y (the packet's)
    e.hook(0x648900, lambda e_: 200, 1)
    e.hook(0x552f60, lambda e_: item if e_.mu.reg_read(UC_X86_REG_EDX) == 4 and e_.arg(0) == 77 else 0, 1)
    e.hook(0x63c1e0, lambda e_: cur["cursor"], 1)                        # the item on the cursor
    e.hook(0x628c20, lambda e_: 1, 1)                                    # usable (requirements)
    e.hook(0x62b400, lambda e_: cur["kind"], 1)                          # the item's type
    e.hook(0x629bb0, lambda e_: 1 if e_.arg(1) == cur["kind"] else 0, 2)
    e.hook(0x625480, lambda e_: cur["qty"] & 0xffffffff if e_.arg(1) == 0x46 else 0, 3)
    e.hook(0x554100, lambda e_: 0, 3)                                    # no trade / interact partner
    e.hook(0x627f80, lambda e_: 0x200, 2)                                # the item's class
    e.hook(0x6374b0, lambda e_: books, 1)
    e.hook(0x6335f0, lambda e_: items_row, 1)
    e.hook(0x6280a0, lambda e_: 0, 4)                                    # item flag 4 clear
    e.hook(0x6280d0, lambda e_: 0, 3)
    e.hook(0x620bb0, lambda e_: room, 1)
    e.hook(0x61a1b0, lambda e_: 2, 1)                                    # level id (Blood Moor; 0x88 refused)
    e.hook(0x6427f0, lambda e_: 0, 1)
    e.hook(0x61ab70, lambda e_: 1, 1)
    e.hook(0x61ab00, lambda e_: cur["town"], 1)
    e.hook(0x535430, ev("closed"), 0)
    e.hook(0x45ae20, lambda e_: 100, 4)
    e.hook(0x45adf0, lambda e_: 200, 1)

    def open_(e_):
        out["opened"] = out.get("opened", 0) + 1
        return portal if cur["spot"] else 0
    e.hook(0x56d130, open_, 7)
    e.hook(0x553380, lambda e_: out.setdefault("voice", []).append(e_.mu.reg_read(UC_X86_REG_EDX)) or 0, 1)
    for a in (0x5353b0, 0x552af0, 0x58cf00): e.hook(a, lambda e_: 0, 0)
    e.hook(0x553720, lambda e_: 0, 0)
    e.hook(0x5531c0, lambda e_: client, 1)
    e.hook(0x53b3d0, lambda e_: 0, 2)
    e.hook(0x53d130, lambda e_: 0, 3)
    e.hook(0x53c520, lambda e_: 0, 3)

    def stat(e_):
        if e_.arg(1) == 0x46: out["delta"] = out.get("delta", 0) + (e_.arg(2) ^ 0x80000000) - 0x80000000
        return 0
    e.hook(0x6272b0, stat, 4)
    e.hook(0x55e0d0, lambda e_: 0, 1)
    e.hook(0x6439b0, lambda e_: 0, 1)                                    # no scroll skill entry (belt count path skipped)
    e.hook(0x55e000, ev("removed"), 1)
    e.hook(0x63b2c0, lambda e_: 0, 1)                                    # FUN_005be1c0's inventory walk: empty
    e.hook(0x543520, lambda e_: 0, 0)

    def case(kind, qty, town, cursor, mode, spot, state54=0):
        cur.update(kind=kind, qty=qty, town=town, cursor=cursor, spot=spot, state54=state54)
        out.clear()
        e.w32(player + 0x10, mode)                                       # the player's mode: the dispatcher's dead test only
        e.w32(player + 0xc4, 0)
        e.w32(game + 0x1dc4, 0)
        e.call(0x54d750, packet, 13, ecx=game, edx=player)
        now = 1 if out.get("opened") and spot and "closed" in out else 0
        spent = 1 if (kind == BOOK and out.get("delta") == -1) or (kind == SCROLL and out.get("removed")) else 0
        return (now, 0, spent, 1 if out.get("closed") else 0)

    rng = random.Random(seed)
    cases = [(k, q, t, c, m, s) for k in (BOOK, SCROLL) for q in list(range(-2, 21)) + [255, 1 << 20]
             for t in (0, 1) for c in (0, 1) for m in MODES for s in (0, 1)]
    cases += [(rng.choice((BOOK, SCROLL)), rng.randrange(-5, 1 << 24) if rng.random() < 0.2 else rng.randrange(-2, 21),
               rng.randrange(2), rng.randrange(2), rng.choice(MODES), rng.randrange(2), int(rng.random() < 0.05)) for _ in range(n_random)]
    classes = Counter()
    shown = set()
    for args in cases:
        kind, qty = args[0], args[1]
        got, want = case(*args), d2d_use(*args)
        if len(args) > 6 and args[6] and got == (0, 0, 0, 0) and want != got: cls = "differ: state 0x36: game.exe drops the packet, d2d has no such test"
        elif got == (0, 0, 0, 0) and want == got: cls = "same: nothing"
        elif got == want: cls = "same"
        elif got[0] and want[1]: cls = "differ: game.exe now, d2d at the SC action frame"
        elif got[0] and want == (0, 0, 0, 0): cls = "differ: game.exe now, d2d nothing (player busy)"
        elif got == (0, 0, 0, 1) and want == (0, 0, 0, 0): cls = "differ: player busy, no free spot: game.exe closes the old pair only, d2d nothing"
        elif not got[0] and not got[2] and want[2]: cls = "differ: no free spot: game.exe keeps the charge, closes the old pair; d2d spends it, keeps the old pair"
        else: cls = f"differ: other game={got} d2d={want}"
        classes[cls] += 1
        empty = kind == BOOK and qty <= 0
        classes[("empty tome", "same" if got == want else "differ")] += empty
        if cls.startswith("differ: other") and cls not in shown:
            shown.add(cls); print(args, cls)
    for name, args in (("empty tome", (BOOK, 0, 0, 0, 1, 1)), ("tome 5, idle", (BOOK, 5, 0, 0, 1, 1)),
                       ("tome 5, in GH", (BOOK, 5, 0, 0, 4, 1)), ("tome 5, in town", (BOOK, 5, 1, 0, 1, 1)),
                       ("tome 5, cursor item", (BOOK, 5, 0, 1, 1, 1)), ("scroll, idle", (SCROLL, 1, 0, 0, 1, 1))):
        got = case(*args)
        print(f"{name}: game.exe {got} {dict(out)}; d2d {d2d_use(*args)}")
    for k, v in sorted(classes.items(), key=str):
        if v: print(v, k)
    print(f"{len(cases)} cases")


def main():
    argv, exe = sys.argv[1:], None
    if "--exe" in argv:
        at = argv.index("--exe"); exe = argv[at + 1]; del argv[at:at + 2]
    run(exe, int(argv[0]) if argv else 100000, int(argv[1]) if len(argv) > 1 else 7)


if __name__ == "__main__": main()
