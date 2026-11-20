#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""A monster's block chance against a landed blow, by game.exe's own code: the oracle for Monster::target(...).block.

    cd <d2decomped>/tools/emu
    uv run python monster_block.py gen > cases.tsv
    uv run python monster_block.py game cases.tsv > game.tsv              # id chance   (-2: a row the emulator cannot spawn)
    <build>/tests/test_monster_block --monster-block cases.tsv > d2d.tsv     # id chance   (-1: no such variant)
    uv run python monster_block.py compare cases.tsv game.tsv d2d.tsv      # exit 1 on a mismatch, or an empty or bad file
    uv run python monster_block.py codes                                   # the item-code rule behind the shield test (game.exe against the tables)

A case line is `id name difficulty variant`: a MonStats row (its Id), the game's difficulty (0 Normal, 1 Nightmare, 2 Hell) and
the index of the entry of the row's MonStats2 shield-layer list (SHv, layer 7) its look holds: monster data +0xb. One case for
every row, difficulty and entry (a row with no list: entry 0).

`game` boots a game of each difficulty (the real spawner FUN_00555230 makes each row in a Blood Moor room), writes the variant
into the unit's monster data (+0xb) and calls FUN_00622720 natively, the function FUN_0057dfb0 asks for a monster's chance to
block. FUN_00622720 reads, for a monster: gate A, the MonStats record's bit 26 of the dword at +0xc (MonStats.txt NoShldBlock,
the mask at 0x6ce270 is 4); or gate B, FUN_006225f0: FUN_00664860 returns the unit's variant code of layer 7 (the MonStats2
list entry, or the Act 2 table of FUN_00664860 for a base-0 skeleton row, which for layer 7 lists the same shields), refused when
it is in the list at 0x744580 (the torch), else FUN_00633640 looks the code up in the item table and the item's type word (+0x11e)
must be 2 (ItemTypes row `shie`). A gate open: stat 20 (MonStats ToBlock of the difficulty), at most 75. Closed: 0.
"""
import csv
import io
import sys

COLS = "id name difficulty variant".split()


def table(name):
    import emu
    raw = emu.mpq_cache().read('data\\global\\excel\\%s.txt' % name).decode('latin-1')
    r = list(csv.reader(io.StringIO(raw), delimiter='\t'))
    return [dict(zip(r[0], x)) for x in r[1:] if x and x[0] and x[0] != 'Expansion']


def variants(row, ms2):
    r = ms2.get(row['MonStatsEx'])
    return [v for v in (r.get('SHv', '') if r else '').split(',') if v]


def gen():
    mons, ms2 = table('monstats'), {r['Id']: r for r in table('monstats2')}
    n = 0
    print('#', *COLS)
    for diff in range(3):
        for row in mons:
            for idx in range(max(1, len(variants(row, ms2)))):
                print(n, row['Id'], diff, idx)
                n += 1


def read_cases(path):
    try:
        lines = open(path).read().splitlines()
    except OSError as err:
        raise SystemExit(f"{path}: {err}") from None
    for k, line in enumerate(lines, 1):
        if line.startswith("#") or not line.strip():
            continue
        f = line.split()
        try:
            c = {'id': int(f[0]), 'name': f[1], 'difficulty': int(f[2]), 'variant': int(f[3])}
            ok = len(f) == 4 and 0 <= c['difficulty'] <= 2 and c['variant'] >= 0
        except (ValueError, IndexError):
            ok = False
        if not ok:
            raise SystemExit(f"{path}:{k}: malformed case line: {line!r}")
        yield c


def game(path, exe=None):
    import mm_launch
    cases = list(read_cases(path))
    if not cases:
        raise SystemExit(f"{path}: no cases: an empty file is not a pass")
    index = {r['Id']: int(r['hcIdx']) for r in table('monstats')}
    games, units, bad = {}, {}, 0
    for c in cases:
        if c['name'] not in index:
            raise SystemExit(f"{path}: unknown monster {c['name']}")
        d = c['difficulty']
        if d not in games:
            games[d] = mm_launch.G(d)
        g = games[d]
        key = (d, c['name'])
        if key not in units:
            try:
                u, _, _ = g.unit(index[c['name']], index[c['name']])
            except Exception:
                u = 0
            units[key] = u
        u = units[key]
        if not u:
            print(c['id'], -2)
            continue
        e = g.e
        e.mu.mem_write(e.r32(u + 0x14) + 0xb, bytes([c['variant']]))
        v = e.call(0x622720, u, 0)
        v -= (1 << 32) if v >> 31 else 0
        print(c['id'], v)
    print("#", len(cases), "cases", file=sys.stderr)


def load(path, cases, name):
    try:
        lines = open(path).read().splitlines()
    except OSError as err:
        print(f"{path}: {err}", file=sys.stderr)
        return None
    rows = []
    for k, line in enumerate(lines, 1):
        if line.startswith("#") or not line.strip():
            continue
        try:
            row = [int(v) for v in line.split()]
        except ValueError:
            row = []
        if len(row) != 2:
            print(f"{path}:{k}: malformed line: {line!r}", file=sys.stderr)
            return None
        rows.append(row)
    if len(rows) != len(cases):
        print(f"{name}: {len(rows)} lines against {len(cases)} cases", file=sys.stderr)
        return None
    for c, r in zip(cases, rows):
        if r[0] != c['id']:
            print(f"{name}: id {r[0]} where case {c['id']} was expected", file=sys.stderr)
            return None
    return rows


def compare(cpath, gpath, dpath):
    cases = list(read_cases(cpath))
    if not cases:
        print("no cases: an empty file is not a pass", file=sys.stderr)
        return 1
    g, d = load(gpath, cases, "game"), load(dpath, cases, "d2d")
    if g is None or d is None:
        return 1
    bad, skipped, blocking, shown = 0, 0, 0, 0
    kinds = {}
    for c, gl, dl in zip(cases, g, d):
        if gl[1] == -2:
            skipped += 1
            continue
        blocking += gl[1] > 0
        if gl[1] != dl[1]:
            bad += 1
            k = ('d2d blocks where game.exe does not' if gl[1] == 0 else 'game.exe blocks where d2d does not' if dl[1] == 0 else 'both block, chances differ')
            kinds[k] = kinds.get(k, 0) + 1
            if shown < 12:
                shown += 1
                print(f"  {c['name']} difficulty {c['difficulty']} variant {c['variant']}: game.exe {gl[1]}, d2d {dl[1]}")
    for k, v in sorted(kinds.items()):
        print(f"  {k}: {v}")
    print(f"compare: {len(cases)} cases ({skipped} rows the emulator cannot spawn skipped, {blocking} where game.exe lets the monster block), d2d differs from game.exe in {bad}")
    return 0 if bad == 0 and skipped * 20 < len(cases) else 1


def codes():
    """The shield test over every item code of the tables: FUN_00633640's record word +0x11e == 2 (the game) against the table's
    `type` column equal to `shie` (the rule d2d uses)."""
    import mm_launch
    g = mm_launch.G(0)
    e = g.e
    tmp = e.alloc(16)
    seen, bad, shields = 0, 0, 0
    for name in ('armor', 'weapons', 'misc'):
        for r in table(name):
            code = r['code']
            if not code or len(code) > 4:
                continue
            word = int.from_bytes(code.ljust(4).encode('latin-1'), 'little')
            rec = e.call(0x633640, word, tmp)
            if not rec:
                print('no record for', name, code)
                continue
            seen += 1
            game_shield = e.r16(rec + 0x11e) == 2
            rule = r['type'] == 'shie'
            shields += game_shield
            if game_shield != rule:
                bad += 1
                print('differs', name, code, 'type', r['type'], 'word', e.r16(rec + 0x11e))
    print(f"codes: {seen} item codes, {shields} shields in game.exe, {bad} where the type column disagrees")
    return 0 if bad == 0 and seen else 1


def cap():
    """The cap of 75 (0x62280e): FUN_00622720 on a gate-A monster (Mephisto) with its stat 20 set to each value with game.exe's own
    stat setter (FUN_00627260); no MonStats row has a ToBlock above 75, so the cases never reach it."""
    import mm_launch
    g = mm_launch.G(0)
    e = g.e
    index = {r['Id']: int(r['hcIdx']) for r in table('monstats')}
    u, _, _ = g.unit(index['mephisto'], 1)
    for value in (0, 1, 74, 75, 76, 90, 200):
        e.call(0x627260, u, 20, value, 0)
        v = e.call(0x622720, u, 0)
        print('cap', value, v - (1 << 32) if v >> 31 else v)
    return 0


USAGE = "usage: monster_block.py gen | game CASES | compare CASES GAME D2D | codes | cap"

if __name__ == "__main__":
    argv = sys.argv[1:]
    arity = {"gen": 1, "game": 2, "compare": 4, "codes": 1, "cap": 1}
    if not argv or arity.get(argv[0]) != len(argv):
        raise SystemExit(USAGE)
    if argv[0] == "gen":
        gen()
    elif argv[0] == "game":
        game(argv[1])
    elif argv[0] == "cap":
        sys.exit(cap())
    elif argv[0] == "codes":
        sys.exit(codes())
    else:
        sys.exit(compare(argv[1], argv[2], argv[3]))
