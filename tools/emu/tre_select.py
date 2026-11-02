# SPDX-License-Identifier: GPL-3.0-or-later
"""Which treasure class game.exe picks for a dying monster (FUN_005a6600 -> FUN_0055afa0 -> FUN_00654e00 -> FUN_0055a6d0),
against a transcription of d2d's rule (components/game/loot.hpp Loot::drop, upstream main 72132b99).

    bash tools/emu/tre_run.sh tre_select.py            # all cases, summary
    bash tools/emu/tre_run.sh tre_select.py -v 20      # plus the first 20 mismatches of each kind

Game side (V run): the real FUN_005a6600 on a fake monster unit (unit type 1, class id, MonStats row native,
monster data +0x16 flags 4 champion / 8 unique / 2 superunique, +0x26 the SuperUniques index), a player killer,
game +0x6d difficulty, +0x70 expansion; the real FUN_00558200 (item level), FUN_005a03a0, FUN_005a0180,
FUN_006556e0, FUN_0055afa0 and FUN_00654e00 run natively. Hooked (not the claim): the unit's room and level id
(FUN_00620bb0, FUN_0061a1b0, FUN_006427f0), the stat getter FUN_00625480 (stat 12 = the case's level), the
killer's player data (FUN_00543520) and the quest flag test FUN_0065c310 (the case's three flags); the roller
FUN_0055a6d0 is a recorder (class record, item level, noNoDrop flag).

d2d side (S, transcription): the selection expression, the quest class, tc_fixed (MonStats noRatio or boss column),
and tc_upgrade over the TreasureClassEx group / level columns.

Cases: every MonStats row x 3 difficulties x {normal, champion, unique} x 11 levels, every SuperUniques row x 3 x 11,
every row with a TCQuestId x its 8 quest-flag states, and every TreasureClassEx class x levels 0..99 for the upgrade.
"""
import sys

import drops


LEVELS = (1, 2, 5, 9, 15, 24, 33, 45, 67, 80, 99)
KINDS = (("normal", 0), ("champion", 4), ("unique", 8))


def main():
    verbose = int(sys.argv[sys.argv.index("-v") + 1]) if "-v" in sys.argv else 0
    o = drops.Oracle()
    e = o.e
    names = o.names
    base = e.r32(0x96c5ec)
    mon = [r for r in drops.txt_rows(e, "MonStats") if r.get("Id") and r.get("hcIdx")]
    sup = [r for r in drops.txt_rows(e, "SuperUniques") if r.get("Superunique")]
    idx_of = {r["Id"]: int(r["hcIdx"]) for r in mon}
    tcx = [r for r in drops.txt_rows(e, "TreasureClassEx") if r["Treasure Class"]]

    game, unit, data, player, kill = o.game, e.alloc(0x200), e.alloc(0x100), e.alloc(0x200), e.alloc(0x20)
    e.w32(game + 0x70, 1)
    e.w32(unit, 1)
    e.w32(unit + 0x14, data)
    e.w32(kill + 8, player)
    fake, state, rec = e.alloc(0x40), {"level": 1, "flags": {}}, {}
    for addr, n, fn in ((0x620bb0, 1, lambda e: fake), (0x61a1b0, 1, lambda e: 2), (0x6427f0, 1, lambda e: 0),
                        (0x543520, 0, lambda e: fake),
                        (0x625480, 3, lambda e: state["level"] if e.arg(1) == 0xc else 0),
                        (0x65c310, 3, lambda e: state["flags"].get((e.arg(1), e.arg(2)), 0))):
        e.hook(addr, fn, n)

    def roller(e):
        rec.update(rec=e.arg(1), ilvl=e.arg(3), nonodrop=e.arg(4), killer=e.arg(0))
        return 0
    e.hook(0x55a6d0, roller, 8)

    def game_pick(cls, diff, flags, superidx, level, quest):
        e.mu.mem_write(game + 0x6d, bytes([diff]))
        e.w32(unit + 4, cls)
        e.mu.mem_write(data, b"\0" * 0x40)
        e.mu.mem_write(data + 0x16, flags.to_bytes(2, "little"))
        e.mu.mem_write(data + 0x26, (superidx & 0xFFFF).to_bytes(2, "little"))
        state["level"], state["flags"] = level, quest
        rec.clear()
        e.call(0x5a6600, game, unit, 0, kill)
        if "rec" not in rec:
            return None
        return names[(rec["rec"] - base) // 0x2c], rec["ilvl"], rec["killer"] == player

    # d2d's upgrade (rules::tc_upgrade): the next row of the group, while lvl >= its level
    group = {r["Treasure Class"]: (int(r["group"] or 0), int(r["level"] or 0)) for r in tcx}
    order = [r["Treasure Class"] for r in tcx]
    nxt = {a: b for a, b in zip(order, order[1:]) if group[a][0] and group[a][0] == group[b][0]}

    def d2d_upgrade(name, lvl):
        for _ in range(len(order)):
            n = nxt.get(name)
            if lvl <= 0 or n is None or lvl < group[n][1]:
                break
            name = n
        return name

    def d2d_pick(row, srow, kind, diff, level, quest):
        sfx = ("", "(N)", "(H)")[diff]
        t = lambda k: row.get(f"TreasureClass{k}{sfx}", "")
        stc = srow.get(("TC", "TC(N)", "TC(H)")[diff], "") if srow else ""
        c2, c3 = (t(3), t(2)) if "-m" in sys.argv else (t(2), t(3))     # -m: swapped champion / unique, the sensitivity check
        tc = stc if stc else c2 if kind == "champion" and c2 else c3 if kind == "unique" and c3 else t(1)
        qid, cp = int(row.get("TCQuestId") or 0), int(row.get("TCQuestCP") or 0)
        if qid and t(4) and not quest.get((qid, 15)) and not quest.get((qid, 1)) and not quest.get((qid, cp)):
            tc = t(4)
        fixed = row.get("noRatio") == "1" or row.get("boss") == "1"
        return d2d_upgrade(tc, level if diff > 0 and not fixed else 0), max(level, 1)

    stats, bad = {"cases": 0, "equal": 0}, {}

    def compare(label, g, d, args):
        stats["cases"] += 1
        if g is None and d is not None and d[0] == "":      # no class on both sides: nothing rolls
            stats["equal"] += 1
            stats["empty"] = stats.get("empty", 0) + 1
            return
        if g is not None and d is not None and g[0] == d[0] and g[1] == d[1] and g[2]:
            stats["equal"] += 1
            return
        bad.setdefault(label, []).append((args, g, d))

    for row in mon:
        cls = int(row["hcIdx"])
        qid, cp = int(row.get("TCQuestId") or 0), int(row.get("TCQuestCP") or 0)
        qstates = [{}]
        if qid:
            qstates = [{(qid, f): (bits >> i) & 1 for i, f in enumerate((15, 1, cp))} for bits in range(8)]
        for diff in range(3):
            for kind, flags in KINDS:
                for level in LEVELS:
                    for quest in qstates:
                        g = game_pick(cls, diff, flags, 0, level, quest)
                        d = d2d_pick(row, None, kind, diff, level, quest)
                        compare(f"{kind}/d{diff}" + ("/quest" if qid else ""), g, d, (row["Id"], diff, kind, level, quest))
    by_id = {r["Id"].lower(): r for r in mon}
    idx_of = {k: int(r["hcIdx"]) for k, r in by_id.items()}
    for si, srow in enumerate(sup):
        mrow = by_id.get(srow["Class"].lower())
        if mrow is None:
            print("superunique without a MonStats row:", srow["Superunique"], repr(srow["Class"]))
            continue
        for diff in range(3):
            for level in LEVELS:
                g = game_pick(idx_of[srow["Class"].lower()], diff, 2 | 8, int(srow["hcIdx"]), level, {})
                d = d2d_pick(mrow, srow, "unique", diff, level, {})
                compare(f"super/d{diff}", g, d, (srow["Superunique"], diff, "super", level))
    # the upgrade alone: game.exe's FUN_00654e00 against the transcription
    up_bad = []
    for name in order:
        for lvl in range(0, 100):
            rec_ = e.call(0x654e00, o.index[name], lvl)
            gname = names[(rec_ - base) // 0x2c]
            if gname != d2d_upgrade(name, lvl):
                up_bad.append((name, lvl, gname, d2d_upgrade(name, lvl)))
    print(f"selection: {stats['cases']} cases, {stats['equal']} equal, {stats['cases'] - stats['equal']} differ (of the equal, {stats.get('empty', 0)} have no class on either side)")
    for label in sorted(bad):
        print(f"  {label}: {len(bad[label])}")
        for args, g, d in bad[label][:verbose]:
            print(f"     {args}: game {g} d2d {d}")
    print(f"upgrade: {len(order) * 100} cases (every TreasureClassEx class x lvl 0..99), {len(up_bad)} differ")
    for x in up_bad[:max(verbose, 3)]:
        print("    ", x)


if __name__ == "__main__":
    main()
