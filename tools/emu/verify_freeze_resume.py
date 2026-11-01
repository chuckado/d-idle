# SPDX-License-Identifier: GPL-3.0-or-later
"""VERIFYFREEZERESUME oracle: game.exe's freeze, its end and the AI event queue, all run for real.

    python verify_freeze_resume.py gen N SEED > cases.tsv     # own generator (LCG, seed masked to 32 bits)
    python verify_freeze_resume.py game cases.tsv > g.tsv      # game.exe under upstream's emulator
    python verify_freeze_resume.py compare cases.tsv g.tsv d.tsv [col]   # d.tsv: test_verify_freeze_resume

A case: id difficulty MonStats-row kind pending stun_at stun_len (at len)...; frames are relative to the
first freeze (frame 0). kind 0 plain, 1 champion, 2 unique, 3 superunique, 4 minion, 5 MonStats boss flag.
`len` is the freeze length in frames AFTER MonsterFreezeDivisor; len -1 asks for a raw length 1 (a
zero-length freeze where the divisor is > 1).

Run natively (no hook on any of them): the queue: FUN_005417d0 -> FUN_005416b0 (file), FUN_00540df0 /
FUN_00540c30 / FUN_00540bb0 (event pool), FUN_00540e60 / FUN_00540cd0 (remove by type), FUN_005414d0 (the
frame: groups 2, 0, 1, 3, 4) -> FUN_00541060 (the monster group: due when event +4 == game +0xa8) ->
FUN_005a7f80 (gate) -> the type handlers; FUN_0057b230 (freeze / chill), FUN_0057aae0 (stun), the
state-list code (FUN_006256b0, FUN_006251f0, FUN_006252d0, FUN_00625ce0, FUN_00626e10, FUN_00625330,
FUN_00625310, FUN_00639db0, FUN_00639df0), the expiry 0x5a7ef0 -> FUN_00627460 -> FUN_00626c00 ->
FUN_006269f0 -> the callback 0x57b170 (freeze); FUN_0057af30,
FUN_0063e9f0, FUN_005a0180, FUN_0063ee90, FUN_00451f80, FUN_005541b0.
Hooked: fog alloc / free (FUN_0040b430 / FUN_0040b480), unit -> game (FUN_00554010), the client update
FUN_00553570, stat getters FUN_00625480 / FUN_00625500 (0), unit-changed FUN_0064c040, the States.txt
record FUN_00452040, FUN_00623f50 (stat refresh after a state change), FUN_005c0b50 (frees the stun's AI
node), FUN_00646f20, the stun's end callback 0x56e900 and the chill FUN_0057af80 (see the hooks), and the AI driver 0x5b1740 (type 2's handler): its frame is the output.
Data: MonStats rows (aidel / aidel(N) / aidel(H) at +0x4f..+0x51, coldeffect -50 at +0x168.., Velocity
1 at +0x32, the boss flag 0x40 at +0xc for kind 5) and DifficultyLevels' MonsterFreezeDivisor from the
1.14d MPQ; game +0x6a = 3 (single player), +0x6d the difficulty.
"""
import struct
import sys

import emu
from drops import txt_rows

BASE = 6000
HORIZON = 1100
KIND_FLAGS = {0: 0, 1: 0xd, 2: 0x8, 3: 0xa, 4: 0x10, 5: 0}
KINDS = ("plain", "champion", "unique", "superunique", "minion", "bossrow")


def lcg(seed):
    x = seed & 0xFFFFFFFF
    while True:
        x = (x * 1664525 + 1013904223) & 0xFFFFFFFF
        yield x >> 8


def aidel_rows(rows):
    """One row id per distinct aidel triple (the first row carrying it)."""
    seen = {}
    for r in rows:
        if not r.get("Id") or r.get("Id") == "Expansion":
            continue
        t = (r.get("aidel", ""), r.get("aidel(N)", ""), r.get("aidel(H)", ""))
        seen.setdefault(t, r["Id"])
    return seen


def gen(n, seed):
    e = emu.Emu()
    triples = aidel_rows(txt_rows(e, "MonStats"))
    ids = sorted(triples.values())
    r = lcg(int(seed, 0))
    def d(k): return next(r) % k
    out = []
    cid = 0
    # grid: every aidel triple x difficulty x plain / minion x shapes
    for row in ids:
        for diff in range(3):
            for kind in (0, 4):
                a = 15
                shapes = [
                    (5, 0, 0, [(0, 30)]),                 # one freeze
                    (0, 0, 0, [(0, 30)]),                 # think due on the freeze frame
                    (31, 0, 0, [(0, 30)]),                # think due one after the end
                    (100, 0, 0, [(0, 30)]),               # pending later than end + aidel
                    (5, 0, 0, [(0, 1)]), (5, 0, 0, [(0, 250)]),
                    (5, 0, 0, [(-20, 40), (0, 10)]),      # prior freeze, longer
                    (5, 0, 0, [(-20, 25), (0, 30)]),      # prior freeze, shorter
                    (5, 0, 0, [(0, 30), (10, 30)]),       # refreeze inside
                    (5, 0, 0, [(0, 30), (30, 20)]),       # refreeze on the end frame
                    (5, 0, 0, [(0, 30), (31, 20)]),       # refreeze in the delay window
                    (5, 0, 0, [(0, 30), (30 + a - 1, 20)]),
                    (5, 0, 0, [(0, 30), (31, 1), (40, 5)]),
                    (5, -10, 60, [(0, 30)]),              # prior stun, ends after the freeze
                    (5, -10, 20, [(0, 30)]),              # prior stun, ends inside the freeze
                ]
                if diff: shapes.append((5, 0, 0, [(0, -1)]))   # zero-length freeze (raw length 1 under divisor 2 / 4)
                for pending, sa, sl, fz in shapes:
                    out.append((cid, diff, row, kind, pending, sa, sl, fz)); cid += 1
    # specials: one freeze, each kind and difficulty, on every row
    for row in ids:
        for diff in range(3):
            for kind in (1, 2, 3, 5):
                out.append((cid, diff, row, kind, 5, 0, 0, [(0, 30)])); cid += 1
    # random
    for _ in range(n):
        diff = d(3); row = ids[d(len(ids))]
        kind = (0, 4, 0, 4, 0, 4, 1, 2, 3, 5)[d(10)]
        pending = d(81)
        fz = [(0, 1 + d(250))]
        if d(4) == 0: fz.insert(0, (-1 - d(300), 1 + d(300)))
        if d(3) == 0: fz.append((1 + d(fz[-1][1] + 60), 1 + d(120)))
        if d(8) == 0: fz.append((fz[-1][0] + d(40), 1 + d(60)))
        sa, sl = 0, 0
        if d(10) == 0: sa, sl = -d(30), 1 + d(120)
        if diff and d(50) == 0: fz = [(0, -1)]
        out.append((cid, diff, row, kind, pending, sa, sl, fz)); cid += 1
    for c in out:
        cid, diff, row, kind, pending, sa, sl, fz = c
        print(cid, diff, row, kind, pending, sa, sl, *[v for f in fz for v in f], sep="\t")


class Oracle:
    def __init__(self):
        e = self.e = emu.Emu()
        self.rows = {r["Id"]: r for r in txt_rows(e, "MonStats") if r.get("Id")}
        self.divisor = [int(r["MonsterFreezeDivisor"]) for r in txt_rows(e, "DifficultyLevels") if r.get("Name")]
        tables = e.alloc(0x1000)
        e.w32(0x744304, tables)
        e.w32(tables + 0xc4, 0x200)                      # state count
        self.mon_table = e.alloc(0x1a8 * 2)
        e.w32(tables + 0xa78, self.mon_table); e.w32(tables + 0xa80, 2)
        dl = e.alloc(0x58 * 3)
        for k in range(3):
            e.w32(dl + 0x58 * k + 0x14, self.divisor[k])
        e.w32(0x96c8ac, dl); e.w32(0x96c8b0, 3)
        self.game = e.alloc(0x2000)
        e.w32(self.game + 0x1c, e.alloc(0x40))
        e.mu.mem_write(self.game + 0x6a, bytes([3]))
        self.thinks = []
        self.chills = 0
        self.heap = None
        from unicorn.x86_const import UC_X86_REG_EDX
        e.hook(0x40b430, lambda e: e.alloc(max(e.mu.reg_read(UC_X86_REG_EDX), 0x40)), 3)
        e.hook(0x40b480, lambda e: 0, 3)
        e.hook(0x554010, lambda e: self.game, 0)
        e.hook(0x553570, lambda e: 0, 1)
        e.hook(0x625480, lambda e: 0, 3)
        e.hook(0x625500, lambda e: 0, 3)
        e.hook(0x64c040, lambda e: 0, 1)
        e.hook(0x452040, lambda e: 0, 0)
        e.hook(0x623f50, lambda e: 0, 3)
        e.hook(0x5c0b50, lambda e: 0, 0)
        e.hook(0x646f20, lambda e: 0, 1)
        e.hook(0x5b1740, self.think, 2)
        # The stun's end callback 0x56e900 (V read: state 21 off, FUN_005c0b50 frees the stun node, a stat refresh,
        # FUN_00575900 for players only; no event filed) faults under the emulator in its stat code: hooked to do
        # nothing. Nothing on this path tests state 21 (FUN_005a7f80 tests state 1, FUN_005416b0 state 0x36).
        e.hook(0x56e900, lambda e: 0, 1)
        # The chill FUN_0057af80 (V read: state 0xb, a type-0xc timer, the velocity stats 0x43..0x45, callback
        # 0x57ad80; no type-2 event filed or removed) faults in its stat code under the emulator: hooked, counted.
        e.hook(0x57af80, self.chill, 2)

    def chill(self, e):
        self.chills += 1
        return 0

    def think(self, e):
        self.thinks.append(e.s32(self.game + 0xa8))
        return 0

    def new_case(self, diff, row, kind):
        e = self.e
        if self.heap is None: self.heap = e.brk
        e.brk = self.heap
        r = self.rows[row]
        mrow = self.mon_table + 0x1a8                   # class id 1
        e.mu.mem_write(mrow, bytes(0x1a8))
        e.mu.mem_write(mrow + 0xc, bytes([0x40 if kind == 5 else 0]))
        e.mu.mem_write(mrow + 0x32, struct.pack("<H", 1))
        for k, col in enumerate(("aidel", "aidel(N)", "aidel(H)")):
            e.mu.mem_write(mrow + 0x4f + k, bytes([int(r.get(col) or 0) & 0xff]))
            e.mu.mem_write(mrow + 0x168 + k, struct.pack("<b", -50))
        e.mu.mem_write(self.game + 0x6d, bytes([diff]))
        q = e.alloc(0xa40)                              # the queue (game +0xb8), fresh each case
        e.w32(self.game + 0xb8, q)
        e.w32(q + 0xa1c, e.call(0x540bb0, regs={"eax": self.game}))   # the queue's first event block (FUN_00540bb0)
        u = self.unit = e.alloc(0x200)
        e.w32(u, 1); e.w32(u + 4, 1); e.w32(u + 0xc, 77); e.w32(u + 0x10, 1)
        md = e.alloc(0x80); e.w32(md, mrow); e.mu.mem_write(md + 0x16, struct.pack("<H", KIND_FLAGS[kind]))
        e.w32(u + 0x14, md)
        ex = e.alloc(0x80)
        e.w32(ex + 0x10, 0x80000000); e.w32(ex + 0x58, e.alloc(0x80)); e.w32(ex + 4, u)
        e.w32(u + 0x5c, ex)
        self.chills = 0
        a = self.attacker = e.alloc(0x200)
        e.w32(a, 0); e.w32(a + 0xc, 1); e.w32(a + 0x20, 0x1234567); e.w32(a + 0x24, 0x89abcdef)
        self.thinks = []

    def run(self, diff, row, kind, pending, sa, sl, fz):
        e = self.e
        self.new_case(diff, row, kind)
        start = min([pending] + ([sa] if sl else []) + [a for a, _ in fz]) - 2
        e.w32(self.game + 0xa8, BASE + start - 1)
        # the pending think: FUN_005417d0(ECX game, EDX unit; type 2, frame, 0, 0)
        e.call(0x5417d0, 2, BASE + pending, 0, 0, ecx=self.game, edx=self.unit)
        for frame in range(start, HORIZON + 1):
            e.w32(self.game + 0xa8, BASE + frame)
            # the missile group (2) runs first in FUN_005414d0: a missile's freeze lands before the monsters' events
            if sl and frame == sa:
                e.call(0x57aae0, self.attacker, sl, regs={"edi": self.unit})
            for at, ln in fz:
                if at == frame:
                    raw = 1 if ln < 0 else ln * self.divisor[diff]
                    e.call(0x57b230, self.attacker, raw, regs={"eax": self.unit})
            e.call(0x5414d0, ecx=self.game)
            if self.thinks:
                return self.thinks[0] - BASE, self.chills
        return None, self.chills


def parse(line):
    f = line.rstrip("\n").split("\t")
    cid, diff, row, kind, pending, sa, sl = int(f[0]), int(f[1]), f[2], int(f[3]), int(f[4]), int(f[5]), int(f[6])
    rest = list(map(int, f[7:]))
    return cid, diff, row, kind, pending, sa, sl, list(zip(rest[::2], rest[1::2]))


def game(path):
    o = Oracle()
    for line in open(path):
        if not line.strip() or line.startswith("#"): continue
        cid, diff, row, kind, pending, sa, sl, fz = parse(line)
        t, chills = o.run(diff, row, kind, pending, sa, sl, fz)
        print(cid, "-" if t is None else t, "chill" if chills else "freeze")


def classify(c, aidel):
    cid, diff, row, kind, pending, sa, sl, fz = c
    if kind not in (0, 4): return "special(" + KINDS[kind] + ")"
    if any(ln < 0 for _, ln in fz): return "zero-length"
    if sl: return "prior-stun"
    first = fz[0]
    if first[0] < 0: return "prior-freeze"
    if pending == 0: return "think-on-freeze-frame"
    end = first[0] + first[1]
    if len(fz) > 1:
        at = fz[1][0]
        if at < end: return "refreeze-inside"
        if at == end: return "refreeze-on-end"
        if at < end + aidel: return "refreeze-delay-window"
        if at == end + aidel: return "refreeze-on-think-frame"
        return "refreeze-after"
    return "single"


def compare(case_path, g_path, d_path, col=1):
    e = emu.Emu()
    rows = {r["Id"]: r for r in txt_rows(e, "MonStats") if r.get("Id")}
    cases = {}
    for line in open(case_path):
        if line.strip() and not line.startswith("#"):
            c = parse(line); cases[c[0]] = c
    g = {int(p[0]): p[1] for p in (l.split() for l in open(g_path) if l.strip())}
    d = {}
    for l in open(d_path):
        p = l.split()
        if len(p) >= 3 and p[0].isdigit(): d[int(p[0])] = p[col]
    stats = {}
    for cid, c in cases.items():
        r = rows[c[2]]
        aidel = int(r.get(("aidel", "aidel(N)", "aidel(H)")[c[1]]) or 0) or 15
        cls = classify(c, aidel)
        s = stats.setdefault(cls, [0, 0, {}])
        s[0] += 1
        if g.get(cid) != d.get(cid):
            s[1] += 1
            try: delta = int(d[cid]) - int(g[cid])
            except (KeyError, ValueError): delta = "?"
            s[2][delta] = s[2].get(delta, 0) + 1
    tot = [0, 0]
    for cls in sorted(stats):
        n, bad, deltas = stats[cls]
        tot[0] += n; tot[1] += bad
        top = ", ".join(f"{k}:{v}" for k, v in sorted(deltas.items(), key=lambda kv: -kv[1])[:4])
        print(f"{cls}\t{bad} of {n} differ\t(d2d - game.exe: {top})")
    print(f"TOTAL\t{tot[1]} of {tot[0]} differ")


if __name__ == "__main__":
    cmd = sys.argv[1]
    if cmd == "gen": gen(int(sys.argv[2]), sys.argv[3])
    elif cmd == "game": game(sys.argv[2])
    elif cmd == "compare": compare(sys.argv[2], sys.argv[3], sys.argv[4], int(sys.argv[5]) if len(sys.argv) > 5 else 1)
