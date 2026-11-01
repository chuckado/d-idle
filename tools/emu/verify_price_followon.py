#!/usr/bin/env python3
"""Independent verifier for upstream candidate fix/vendor-prices-followon (d2lod-bot card verifyprs2:
D9 e297025a, D10 312272a7). Written without reading the author's oracles (price_edge.py,
price_claims.py, price_oracle.py).

game.exe side, re-derived from disasm.py reads of FUN_0062efb0:
- D9 (0x62faac..0x62fc53): three prices in parallel, the player's buy at [ebp-0x10] with npc.txt
  "sell mult" (record +4), the player's sale in edi with "buy mult" (+8), the repair at [ebp-8] with
  "rep mult" (+0xc) (column offsets from the loader's descriptors at 0x656650..). Each: price above
  0xffff (signed jle) and the mult tested non-zero -> (price / 1024, toward zero: cdq / and 0x3ff /
  sar 10) x mult; else mult x price / 1024 toward zero; 32-bit imul. The sale tests the BUY's mult
  ("sell mult", ecx from 0x62fab7 / 0x62fad1) for zero, not its own; the repair and the quest steps
  test their own. Then for each of the 3 quest flags (+0x10/14/18) set and done (FUN_0065c310 bit 0
  or bit 1): the same with questsellmult (+0x1c..), questbuymult (+0x28..), questrepmult (+0x34..).
- D10 FUN_0062e660: an item unit (type 4) with data, identified (flags & 0x10), not ethereal
  (0x400000); FUN_0062a840 (stat 204 list: some entry with value & 0xff < value >> 8) -> pass; else
  ItemTypes +8 (Repair) of its type must be set, then FUN_0062e5d0 (ItemTypes Repair, Throwable +0x10
  (FUN_0062ba80), Items +0x132 stackable (FUN_006289f0), not ethereal) or FUN_00629930 (Items +0x113
  nodurability 0, +0x112 durability > 0, FUN_00625e00 stat 73 != 0, stat 152 <= 0) -> pass.
  Other callers: 0x62efee (the repair price: 0 when refused), 0x62fc86 (a partly used stack's sale),
  0x62fea7 (Repair All's sum, FUN_0062fe60), 0x5761cb / 0x5781fa (the server's repair), 0x4b1f96.

game side: game.exe's own instructions under unicorn (upstream tools/emu/emu.py): D9 runs
0x62faac..0x62fc59 with the three locals, esi / [ebp+0x10] = a fake npc.txt record, FUN_0065c310
hooked (the case's quest flags); D10 calls FUN_0062e660 with FUN_0062e5d0, FUN_0062a840,
FUN_00629930, FUN_0062ba80, FUN_006289f0, FUN_0062a8d0, FUN_0062b400, FUN_00628700, FUN_006335f0,
FUN_00555c90 native against fake Items ([0x96ca5c]) and ItemTypes (dt +0xbf8) rows; hooked:
FUN_006261d0 (the stat 204 list), FUN_00625480 (stat 152), FUN_00625e00 (stat 73).
d2d side: the REAL rules::item_price / rules::repair_cost (`test_rules --verify-price`,
verify_price_followon_mode.inc pasted into a throwaway worktree).

    python verify_price_followon.py gen N SEED > cases.tsv
    python verify_price_followon.py game cases.tsv > game.tsv
    build/tests/test_rules --verify-price cases.tsv > d2d.tsv
    python verify_price_followon.py compare cases.tsv game.tsv d2d.tsv

cases kind 9: id 9 which P mult done q1 q2 q3 bits z (z: game only, the buy's "sell mult" set to 0 for
  a sale: the column quirk, outside 1.14d's npc.txt)
cases kind 10: id 10 ident eth repair throw stack kind2 nodur durcol max dur indes qty maxstack c1l c1m c2l c2m
game.tsv: kind 9 `id price overflow` (overflow 1 when an exact intermediate leaves int32);
          kind 10 `id gate expected` (expected = what d2d's repair_cost > 0 should be: the gate and
          something to repair in #152's model: charges missing, a restocking stack short of its max
          stack, or another item's durability below max; a whole item is 0 in d2d, 1 in game.exe,
          #152's ponytail, so the gate is compared, not the price).
"""
import random
import struct
import sys

NPC_MULTS = (1024, 1126, 921, 819, 768, 614, 512, 1536, 2048)   # around npc.txt's
EDGE_P = (0, 1, 1023, 1024, 1025, 0xfffe, 0xffff, 0x10000, 0x10001, 0x103ff, 0x10400, 0x1fffff, 0x7fffffff,
          -1, -1024, -1025, -0x10000, -0x10001, -0x7fffffff - 1)
EDGE_M = (0, 1, 2, 1023, 1024, 1025, 2047, 65535, 0x8000, -1, -1024)


def gen(n, seed):
    rng = random.Random(seed)
    rows = []
    # D9 boundaries: every edge price x edge mult x which, quests off / one / all with edge mults
    for which in range(3):
        for p in EDGE_P:
            for m in EDGE_M + NPC_MULTS[:3]:
                rows.append((9, which, p, m, 0, 1024, 1024, 1024, 0, 0))
                rows.append((9, which, p, m, 7, rng.choice(EDGE_M), rng.choice(NPC_MULTS), rng.choice(EDGE_M), rng.randrange(8), 0))
                if which == 1:
                    rows.append((9, which, p, m, 0, 1024, 1024, 1024, 0, 1))
    # D10 boundaries: every combination of the gate's flags with each durability edge
    for bits in range(1 << 8):
        ident, eth, repair, throw, stack, nodur, chg, full = ((bits >> k) & 1 for k in range(8))
        for durcol, mx, dur, indes in ((0, 10, 5, 0), (1, 10, 5, 0), (255, 10, 5, 0), (10, 0, 0, 0), (10, 10, 10, 0),
                                       (10, 10, 9, 0), (10, 10, 5, 1), (10, 10, 5, -1), (10, 10, 5, 2), (10, -5, -6, 0)):
            qty, ms = (5, 5) if full else (3, 5)
            c = (3, 10, 0, 0) if chg else (10, 10, 0, 0)
            rows.append((10, ident, eth, repair, throw, stack, throw, nodur, durcol, mx, dur, indes, qty, ms, *c))
    n += len(rows)
    while len(rows) < n:
        r = rng.random()
        if r < 0.55:
            which = rng.randrange(3)
            p = rng.choice((rng.randint(0, 0xffff), rng.randint(0x10000, 0x200000), rng.randint(0x10000, 0x7fffffff),
                            rng.randint(1, 5000), rng.randint(-0x80000000, 0x7fffffff), rng.choice(EDGE_P)))
            m = rng.choice(NPC_MULTS + (rng.randint(0, 4096), rng.choice(EDGE_M)))
            done = rng.randrange(8) if rng.random() < 0.7 else 0
            qs = [rng.choice(NPC_MULTS + (rng.randint(0, 4096), rng.choice(EDGE_M))) for _ in range(3)]
            z = int(which == 1 and rng.random() < 0.05)
            rows.append((9, which, p, m, done, *qs, rng.randrange(8), z))
        else:
            throw = rng.random() < 0.3
            kind2 = throw if rng.random() < 0.95 else not throw          # 5 %: off 1.14d's data (d2d's kind 2 vs Throwable)
            ms = rng.choice((1, 20, 40, 60, 250))
            mx = rng.choice((0, 0, rng.randint(1, 250), rng.randint(-100, 3000)))
            dur = rng.choice((mx, rng.randint(0, max(mx, 0)), rng.randint(-50, 300)))
            cl, cm = rng.randint(0, 40), rng.choice((0, 0, rng.randint(1, 40), rng.randint(1, 255)))
            rows.append((10, int(rng.random() < 0.85), int(rng.random() < 0.15), int(rng.random() < 0.7), int(throw),
                         int(rng.random() < 0.3 or throw), int(kind2), int(rng.random() < 0.15), rng.choice((0, rng.randint(1, 255))),
                         mx, dur, rng.choice((0, 0, 0, 1, rng.randint(-3, 3))), rng.randint(0, max(ms, 1)), ms,
                         cl, cm, rng.randint(0, 40), rng.choice((0, rng.randint(1, 60)))))
    for i, r in enumerate(rows):
        print(i, *r, sep="\t")


def s32(v):
    v &= 0xFFFFFFFF
    return v - (1 << 32) if v & 0x80000000 else v


def tdiv(a, b):
    q = abs(a) // abs(b)
    return q if (a >= 0) == (b > 0) else -q


def exact_overflows(which, p, m, done, qs):
    """Whether d2d's 64-bit arithmetic leaves int32 where game.exe's 32-bit would wrap."""
    vals, x = [], p
    for k, mm in [(None, m)] + [(k, qs[k]) for k in range(3) if done >> k & 1]:
        prod = (tdiv(x, 1024) * mm) if (x > 0xffff and mm) else (mm * x)
        vals.append(prod)
        x = prod if (x > 0xffff and mm) else tdiv(prod, 1024)
        vals.append(x)
    return any(v < -2**31 or v >= 2**31 for v in vals)


def game(path):
    sys.path.insert(0, ".")
    from emu import Emu
    from unicorn.x86_const import UC_X86_REG_EBP, UC_X86_REG_EDI, UC_X86_REG_ESI, UC_X86_REG_ESP, UC_X86_REG_EAX
    emu = Emu()
    npc = emu.alloc(0x100)
    frame = emu.alloc(0x400)
    ebp = frame + 0x200
    quest = {"done": 0}
    # FUN_0065c310(questrec, flag, bit): the case's quest flags 1..3, done on its bit
    emu.hook(0x65c310, lambda e: int((quest["done"] >> (e.arg(1) - 1) & 1) and (quest["bits"] >> (e.arg(1) - 1) & 1) == e.arg(2)
                                     if 1 <= e.arg(1) <= 3 else 0), 3)
    # D10 tables
    dt = emu.alloc(0x1000)
    emu.w32(0x744304, dt)
    types = emu.alloc(0xe4 * 2)
    emu.w32(dt + 0xbf8, types)
    emu.w32(dt + 0xbfc, 2)
    items = emu.alloc(0x1a8 * 2)
    emu.w32(0x96ca5c, items)
    emu.w32(0x96ca58, 2)
    unit = emu.alloc(0x100)
    data = emu.alloc(0x100)
    st = {}

    def statlist(e):
        buf, entries = e.arg(2), st.get("chg", [])
        for k, (left, most) in enumerate(entries):
            e.w32(buf + 8 * k, 1 << 6 | 5)
            e.w32(buf + 8 * k + 4, most << 8 | left)
        return len(entries)

    emu.hook(0x6261d0, statlist, 4)
    emu.hook(0x625480, lambda e: st.get(e.arg(1), 0) & 0xFFFFFFFF, 3)
    emu.hook(0x625e00, lambda e: st.get(73, 0) & 0xFFFFFFFF, 1)
    for line in open(path):
        f = line.split()
        if not f:
            continue
        cid, kind = f[0], int(f[1])
        if kind == 9:
            which, p, m, done, q1, q2, q3, bits, z = map(int, f[2:11])
            qs = (q1, q2, q3)
            cols = [1024, 1024, 1024]
            cols[which] = m                              # record +4 sell mult (our buy), +8 buy mult (our sale), +0xc rep
            if z:
                cols[0] = 0
            for k, v in enumerate(cols):
                emu.w32(npc + 4 + 4 * k, v)
            for k in range(3):
                emu.w32(npc + 0x10 + 4 * k, k + 1)
                for base in (0x1c, 0x28, 0x34):
                    emu.w32(npc + base + 4 * k, 1024)
                emu.w32(npc + (0x1c, 0x28, 0x34)[which] + 4 * k, qs[k])
            quest["done"], quest["bits"] = done, bits
            emu.w32(ebp - 0x10, p)
            emu.w32(ebp - 8, p)
            emu.w32(ebp - 4, 0)
            emu.w32(ebp + 0x10, npc)
            emu.w32(ebp + 0xc, 0x1234)
            mu = emu.mu
            mu.reg_write(UC_X86_REG_EBP, ebp)
            mu.reg_write(UC_X86_REG_ESP, frame + 0x20)
            mu.reg_write(UC_X86_REG_ESI, npc)
            mu.reg_write(UC_X86_REG_EDI, p & 0xFFFFFFFF)
            mu.emu_start(0x62faac, 0x62fc59)
            out = (emu.s32(ebp - 0x10), emu.s32(ebp - 4), s32(mu.reg_read(UC_X86_REG_EDI)))[which]   # buy, sale, repair at 0x62fc59
            # the d2d functions' own ends, not part of the claim, on both sides: at least 1 (a buy's
            # and a repair's price_reduced, a sale's floor), a sale under max buy (INT_MAX here)
            print(cid, max(out, 1), int(exact_overflows(which, p, m, done, qs)))
        else:
            (ident, eth, repair, throw, stack, kind2, nodur, durcol, mx, dur, indes, qty, ms,
             c1l, c1m, c2l, c2m) = map(int, f[2:19])
            emu.mu.mem_write(types + 8, bytes([repair]))
            emu.mu.mem_write(types + 0x10, bytes([throw]))
            emu.mu.mem_write(items + 0x112, bytes([durcol, nodur]))
            emu.mu.mem_write(items + 0x11e, struct.pack("<h", 0))
            emu.mu.mem_write(items + 0x132, bytes([stack]))
            emu.w32(unit, 4)
            emu.w32(unit + 4, 0)
            emu.w32(unit + 0x14, data)
            emu.w32(data + 0x18, (0x10 if ident else 0) | (0x400000 if eth else 0))
            st.clear()
            st["chg"] = [(c, m_) for c, m_ in ((c1l, c1m), (c2l, c2m)) if m_ > 0]
            st[73], st[152], st[0x48] = mx, indes, dur
            gate = emu.call(0x62e660, unit)
            charges = any(c < m_ for c, m_ in st["chg"])
            restock = repair and throw and stack and not eth
            durable = not nodur and durcol > 0 and mx != 0 and indes <= 0
            # what d2d (#152's model) repairs once the gate passes: a restocking stack its quantity (its
            # durability free), anything else its durability; both: the missing charges
            missing = charges or (repair and (qty < ms if restock else durable and dur < mx))
            print(cid, gate, int(bool(gate) and bool(missing)))


def compare(cases, a, b):
    from collections import Counter
    rows = {ln.split()[0]: ln.split() for ln in open(cases) if ln.strip()}
    ga = {ln.split()[0]: ln.split() for ln in open(a) if ln.strip()}
    gb = {ln.split()[0]: ln.split() for ln in open(b) if ln.strip()}
    seen, bad, ex = Counter(), Counter(), {}
    for k, r in rows.items():
        if k not in ga or k not in gb:
            seen["missing"] += 1
            continue
        if r[1] == "9":
            which = ("buy", "sell", "repair")[int(r[2])]
            if r[10] == "1":
                key = "D9 sale, buy column 0 (quirk, off-data)"
            elif ga[k][2] == "1":
                key = f"D9 {which} int32 overflow (off-data)"
            else:
                key = f"D9 {which}" + (" quest" if r[5] != "0" else "")
            ok = int(ga[k][1]) == int(gb[k][1])
        else:
            off = r[5] != r[7]                       # Throwable vs d2d's kind 2
            key = "D10 repair" + (" (kind 2 != Throwable, off-data)" if off else "")
            ok = int(ga[k][2]) == int(int(gb[k][1]) > 0)
        seen[key] += 1
        if not ok:
            bad[key] += 1
            ex.setdefault(key, []).append(k)
    print(f"cases {len(rows)}")
    for key in sorted(seen):
        print(f"  {key}: {seen[key]} compared, {bad[key]} differ")
    for key, ks in ex.items():
        for k in ks[:4]:
            print("   ", key, "|", " ".join(rows[k]), "| game", " ".join(ga[k][1:]), "d2d", gb[k][1])


if __name__ == "__main__":
    cmd = sys.argv[1] if len(sys.argv) > 1 else ""
    if cmd == "gen":
        gen(int(sys.argv[2]), int(sys.argv[3]))
    elif cmd == "game":
        game(sys.argv[2])
    elif cmd == "compare":
        compare(*sys.argv[2:5])
    else:
        sys.exit(__doc__)
