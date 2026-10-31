#!/usr/bin/env python3
"""game.exe 1.14d's waypoint refusal window as the oracle (d2lod-bot, card decompwp).

The server's C->S 0x49 path runs natively under unicorn through upstream's harness
(psi29a/d2decomped tools/emu/emu.py): the dispatcher FUN_0054d750 (its busy test
FUN_0057eec0), the 0x49 handler FUN_0054c5d0 with its timer read FUN_0055b6c0
(player data +0x160 + 10000 ms > GetTickCount -> refuse: voice event 0x13,
FUN_00554190 ends the interaction) and FUN_00554190. A hostile declaration runs
game.exe's own setter FUN_0055b720 natively (the only writer of player data
+0x160 in .text: a whole-image scan, see docs/research/waypoint-trace.md).
Hooked: GetTickCount (the case's clock), FUN_006221a0 (the player data), the
validation FUN_00549570 (0 = object in range, same act, target active; 2 = not
active), FUN_00584f60 (recorded: the travel), FUN_00553380 (recorded: voice).

d2d's rule is a Python copy of components/game/world/commands.cpp:267
(upstream main 03f42254): refuse when now - arrived_at < 10000, arrived_at being
set by World::swap_npcs on every arrival (setup.cpp:363), the game entry included.
`--after` checks the proposed fix (a hostile timestamp that is set only by a
hostile declaration, compared the way game.exe does).

    cd <d2decomped>/tools/emu && source <venv>/bin/activate
    PYTHONPATH=. python <this repo>/tools/emu/wp_refuse.py --exe game.exe [random N] [seed] [--after]
"""
import random
import sys
from collections import Counter

import emu
from unicorn.x86_const import UC_X86_REG_EDX

T_MIN, T_MAX = 10_000, 1 << 31      # GetTickCount at the packet: past the first 10 s of uptime, before the wrap


def d2d_refuses(now, levels, hostiles, active):
    """commands.cpp:267: the 10 s test on the last arrival, then the active-waypoint test."""
    arrived = max(levels)
    return (now - arrived) < 10000 or not active


def fix_refuses(now, levels, hostiles, active):
    """Proposed: the window counts from the last hostile declaration only (never set in a single-player game)."""
    last = max(hostiles) if hostiles else 0
    return (last + 10000) & 0xffffffff > now or not active


def run(exe, n_random, seed, after):
    cur = {"now": 0}
    emu.Emu.w_GetTickCount = lambda self: (cur["now"], 0)
    e = emu.Emu(exe) if exe else emu.Emu()
    out = {}
    game = e.alloc(0x2000)
    player = e.alloc(0x200)
    other = e.alloc(0x200)
    pcdata = e.alloc(0x400)
    packet = e.alloc(16)
    e.w32(player, 0); e.w32(player + 0xc, 1); e.w32(player + 0x10, 1)    # a player, id 1, mode NU (walk modes pass the busy test)
    e.w32(other, 0); e.w32(other + 0xc, 2)

    e.hook(0x6221a0, lambda e_: pcdata, 1)
    e.hook(0x549570, lambda e_: 0 if cur["active"] else 2, 2)
    e.hook(0x554d00, lambda e_: 0, 0)
    e.hook(0x552f60, lambda e_: 0, 1)
    e.hook(0x584f60, lambda e_: out.__setitem__("travel", 1) or 0, 2)
    e.hook(0x553380, lambda e_: out.setdefault("voice", []).append(e_.mu.reg_read(UC_X86_REG_EDX)) or 0, 1)
    e.hook(0x410a80, lambda e_: 0, 0)
    e.w32(game + 0x1dc4, 0)

    def case(now, levels, hostiles, active):
        """levels, hostiles: the GetTickCount of each level change and hostile declaration before the packet."""
        cur["active"] = active
        out.clear()
        e.mu.mem_write(pcdata, b"\0" * 0x400)                            # fresh player data (I: the allocator zeroes it)
        for t in sorted(hostiles):
            cur["now"] = t
            e.call(0x55b720, (t + 60000) & 0xffffffff, ecx=player, edx=other)   # the hostile setter, as FUN_005a5e50 calls it
        # level changes: nothing in game.exe writes player data +0x160 (scan), so nothing runs here
        cur["now"] = now
        e.mu.mem_write(packet, bytes([0x49]) + (900).to_bytes(4, "little") + (3).to_bytes(2, "little") + b"\0\0")
        e.w32(player + 0x64, 0); e.w32(player + 0x68, 0)
        e.call(0x54d750, packet, 9, ecx=game, edx=player)
        return not out.get("travel")

    rng = random.Random(seed)
    cases = []
    for dl in range(0, 20001, 100):                                     # exhaustive grid: ms since the last level change
        for dh in [None] + list(range(0, 20001, 250)):                  # ms since the last hostile declaration (None: never)
            for active in (1, 0):
                now = 1_000_000
                cases.append((now, [now - dl], [] if dh is None else [now - dh], active))
    for _ in range(n_random):
        now = rng.randrange(T_MIN + 60000, T_MAX)
        levels = [now - rng.choice((rng.randrange(0, 12000), rng.randrange(0, 600000))) for _ in range(rng.randrange(1, 4))]
        hostiles = [now - rng.choice((rng.randrange(0, 12000), rng.randrange(0, 60000))) for _ in range(rng.choice((0, 0, 0, 1, 2)))]
        cases.append((now, levels, hostiles, int(rng.random() < 0.9)))
    model = fix_refuses if after else d2d_refuses
    classes = Counter()
    for c in cases:
        got, want = case(*c), model(*c)
        now, levels, hostiles, active = c
        near_level = now - max(levels) < 10000
        near_host = bool(hostiles) and now - max(hostiles) < 10000
        if got == want: cls = f"same: {'refused' if got else 'travel'}"
        elif want and not got: cls = "differ: d2d refuses, game.exe travels" + (" (within 10 s of a level change)" if near_level else "")
        else: cls = "differ: game.exe refuses, d2d travels" + (" (within 10 s of a hostile declaration)" if near_host else "")
        classes[cls] += 1
    for name, c in (("2 s after a level change", (1_000_000, [998_000], [], 1)),
                    ("2 s after a hostile declaration", (1_000_000, [500_000], [998_000], 1)),
                    ("10.0 s after a hostile declaration", (1_000_000, [500_000], [990_000], 1)),
                    ("inactive target", (1_000_000, [500_000], [], 0))):
        print(f"{name}: game.exe refuses={case(*c)} voice={out.get('voice')}; {'fix' if after else 'd2d'} refuses={model(*c)}")
    for k, v in sorted(classes.items()):
        print(v, k)
    print(f"{len(cases)} cases, {sum(v for k, v in classes.items() if k.startswith('differ'))} differ ({'after the fix' if after else 'd2d today'})")


def main():
    argv, exe = sys.argv[1:], None
    after = "--after" in argv
    if after: argv.remove("--after")
    if "--exe" in argv:
        at = argv.index("--exe"); exe = argv[at + 1]; del argv[at:at + 2]
    run(exe, int(argv[0]) if argv else 100000, int(argv[1]) if len(argv) > 1 else 7, after)


if __name__ == "__main__": main()
