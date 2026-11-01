# SPDX-License-Identifier: GPL-3.0-or-later
"""VERIFYNEXTHIT oracle: replays test_verify_nexthit contact lines through game.exe.

    python verify_nexthit.py d2d.txt [I|P] [post|pre] > g.txt

I / P: which contact order to replay (monster index order / crossed-subtile order).
post (default, the traced order): a struck unit's type-12 timer (group 1, monsters)
runs in FUN_005414d0 AFTER the missile group (group 2) of its frame; pre: before.
Run for real: the filter 0x5a8850 -> FUN_005a8730, FUN_00639df0 / FUN_00625b70
(state bits of a StatListEx at unit +0x5c), FUN_0064a470 / FUN_0064a400, FUN_005adf10,
FUN_005ada40, FUN_005adb00, FUN_006251f0 / FUN_006252d0 / FUN_00625ce0 / FUN_00626e10,
FUN_00639db0 / FUN_00625a70, and at expiry FUN_005a7ef0 -> FUN_00627460 -> FUN_00626c00
-> FUN_006269f0 -> the callback 0x5adaf0.
Hooked: fog alloc / free (FUN_0040b430 / FUN_0040b480), the game (FUN_00554010), the
owner (FUN_00552fd0, a player), hostility (FUN_00554200: yes), stat getters FUN_00625480 /
FUN_00625500 (0), unit-changed FUN_0064c040, the States.txt record FUN_00452040 (none), the timer insert FUN_005417d0 (recorded),
hit event FUN_005c0c30, damage roll FUN_005a89a0 (counted), apply FUN_005adcd0.
"""
import sys

import emu
from drops import txt_rows
from unicorn.x86_const import UC_X86_REG_EDX

FRAME0 = 5000


class Oracle:
    def __init__(self, post):
        self.post = post
        e = self.e = emu.Emu()
        rows = {r["Missile"]: r for r in txt_rows(e, "Missiles") if r.get("Missile")}
        tables = e.alloc(0x1000)
        e.w32(0x744304, tables)
        e.w32(tables + 0xc4, 0x100)            # state count
        self.row_at = {}
        block = e.alloc(0x1a4 * 2)
        e.w32(tables + 0xb64, block); e.w32(tables + 0xb6c, 2)
        for k, name in enumerate(("nova", "frostnova")):
            r = rows[name]
            v = lambda c: int(r.get(c) or 0)
            at = block + 0x1a4 * k
            e.w32(at + 4, v("LastCollide") | v("Explosion") << 1 | v("Pierce") << 2)
            for off, col in ((0xc, "pSrvDoFunc"), (0xe, "pSrvHitFunc"), (0x10, "pSrvDmgFunc")):
                e.mu.mem_write(at + off, v(col).to_bytes(2, "little"))
            for off, col in ((0x183, "CollideType"), (0x186, "CollideKill"), (0x187, "CollideFriend"), (0x188, "NextHit"),
                             (0x189, "NextDelay"), (0x18a, "Size"), (0x18b, "ToHit"), (0x18c, "AlwaysExplode")):
                e.mu.mem_write(at + off, bytes([v(col)]))
            assert v("pSrvHitFunc") == 0 and v("pSrvDmgFunc") == 0 and v("ToHit") == 0, name
            self.row_at[name] = (k, at)
            print(f"# {name} NextHit {v('NextHit')} NextDelay {v('NextDelay')} LastCollide {v('LastCollide')}", file=sys.stderr)
        self.game = e.alloc(0x400)
        e.w32(self.game + 0x1c, e.alloc(0x40))  # memory pool handle (opaque)
        self.player = e.alloc(0x200); e.w32(self.player, 0); e.w32(self.player + 0xc, 1)
        self.timers = []
        self.count = {}
        self.heap = None
        e.hook(0x40b430, lambda e: e.alloc(max(e.mu.reg_read(UC_X86_REG_EDX), 0x40)), 3)
        e.hook(0x40b480, lambda e: 0, 3)
        e.hook(0x554010, lambda e: self.game, 0)
        e.hook(0x552fd0, lambda e: self.player, 0)
        e.hook(0x554200, lambda e: 1, 1)
        e.hook(0x625480, lambda e: 0, 3)
        e.hook(0x625500, lambda e: 0, 3)
        e.hook(0x64c040, lambda e: 0, 1)
        e.hook(0x452040, lambda e: 0, 0)               # States.txt record: none (only a +0xc8 display flag hangs on it)
        e.hook(0x5417d0, self.timer, 4)
        e.hook(0x5c0c30, lambda e: 0, 3)
        e.hook(0x5a89a0, self.roll, 3)
        e.hook(0x5adcd0, lambda e: 0, 0)

    def timer(self, e):
        self.timers.append((e.arg(1), e.mu.reg_read(UC_X86_REG_EDX), e.arg(0)))
        return 0

    def roll(self, e):
        unit = e.arg(1)
        self.count[unit] = self.count.get(unit, 0) + 1
        return 0

    def new_case(self, skill, n):
        e = self.e
        if self.heap is None: self.heap = e.brk
        e.brk = self.heap                     # every case's units, lists and contexts from the same heap
        self.timers = []
        self.kind, self.row = self.row_at[skill]
        self.mons = []
        for i in range(n):
            u = e.alloc(0x200)
            e.w32(u, 1); e.w32(u + 0xc, 10 + i); e.w32(u + 0xc4, 0xc)
            ex = e.alloc(0x80)
            e.w32(ex + 0x10, 0x80000000); e.w32(ex + 0x58, e.alloc(0x40)); e.w32(ex + 4, u)
            e.w32(u + 0x5c, ex)
            self.mons.append(u)
        self.missiles = {}
        self.total = [0] * n

    def missile(self, key):
        if key not in self.missiles:
            e = self.e
            u = e.alloc(0x200); d = e.alloc(0x40)
            e.w32(u, 3); e.w32(u + 4, self.kind); e.w32(u + 0xc, 90000 + key); e.w32(u + 0x10, 3); e.w32(u + 0x14, d)
            e.w32(d + 0x18, 0); e.w32(d + 0x1c, 1)          # last unit: the owner (FUN_0059fa30)
            self.missiles[key] = u
        return self.missiles[key]

    def expire(self, frame):
        due = [t for t in self.timers if t[0] == frame]
        self.timers = [t for t in self.timers if t[0] != frame]
        for _, unit, kind in due:
            assert kind == 0xc
            self.e.call(0x5a7ef0, kind, 0, ecx=self.game, edx=unit)

    def frame(self, frame, contacts):
        e = self.e
        now = FRAME0 + frame
        e.w32(self.game + 0xa8, now)
        assert all(t[0] >= now for t in self.timers), "missed timer"
        if not self.post: self.expire(now)
        self.count = {}
        ctx = e.alloc(0x10)
        for key, units in contacts:
            m = self.missile(key)
            e.w32(ctx, self.game); e.w32(ctx + 4, m); e.w32(ctx + 8, self.player); e.w32(ctx + 0xc, self.row)
            for i in units:
                if e.call(0x5a8850, ecx=self.mons[i], edx=ctx):
                    e.call(0x5adf10, self.mons[i], 0, ecx=self.game, edx=m)
                    break
        if self.post: self.expire(now)
        out = []
        for i, u in enumerate(self.mons):
            if self.count.get(u):
                self.total[i] += self.count[u]
                out.append(f"{i}x{self.count[u]}")
        return out

    def flush(self, last):
        # timers due after the last contact frame: run them so nothing leaks
        self.timers = []


def main():
    path, order = sys.argv[1], (sys.argv[2] if len(sys.argv) > 2 else "I")
    post = (sys.argv[3] if len(sys.argv) > 3 else "post") == "post"
    o = Oracle(post)
    tag = None
    last = 0

    def finish():
        if tag is not None: print("S" + tag + "".join(f" {t}" for t in o.total))

    for line in open(path):
        if line[0] == "K":
            finish()
            f = line.split()
            tag = f[0][1:]
            o.new_case(f[1], int(f[3]))
            last = 0
        elif line[0] == order:
            f = line.split()
            fr = int(f[1])
            for skipped in range(last + 1, fr):     # frames without contacts still run their timers
                o.frame(skipped, [])
            last = fr
            contacts = [(int(a), [int(x) for x in b.split(".")]) for a, b in (c.split(":") for c in f[2:])]
            hits = o.frame(fr, contacts)
            if hits: print(f"D{tag} {fr} " + " ".join(hits))
    finish()


if __name__ == "__main__":
    main()
