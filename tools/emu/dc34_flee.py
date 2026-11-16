"""decomp34: the real Fallen think (FUN_005f02c0) on real spawned units, the flee scan against the rooms' death rings.

  dc34_flee.py flee CASES SEED TABLES_DIR     one line per case:  F case fx fy away_ok | deaths (dx dy room_index fallen) ... | fled pace away fell calls

Game side: mm_launch.G (real rooms of a Blood Moor game), FUN_00555230 spawns, FUN_005a6ff0 (the death handler, registers the unit in its
room's ring via FUN_0061afa0 at 0x5a704d) kills, FUN_005f02c0 runs on a real Fallen. Hooked (acts and the AI command calls): 5ddf90 5de080
5de0f0 5de130 5dead0 5de000 5dec80 5ded40 5def80 5defb0 5defe0 (the back-off: returns the case's away_ok) 5de190 (pace: recorded)
58ed10 (drop the command) 58ee80 5dd2b0 5de200. Not hooked (real): 620bb0 619790 552f60 5dc530 45adf0 45ae20.
fled = the dying branch's pace call (pace 0x32) was made; fell = a call after a failed back-off (the think went on)."""
import os, random, sys
sys.path.insert(0, os.environ.get("DC34", "/data/dc34"))
import mm_launch

def room_at(g, x, y):
    for (r, x0, y0, w, h, _) in g.rooms:
        if x0 <= x < x0 + w and y0 <= y < y0 + h:
            return r
    return 0

def main():
    cases, seed, tables = int(sys.argv[2]), int(sys.argv[3]), sys.argv[4]
    rng = random.Random(seed)
    mons = {m["Id"]: m for m in mm_launch.rows(tables, "monstats.txt")}
    fal, zom = int(mons["fallen1"]["hcIdx"]), int(mons["zombie1"]["hcIdx"])
    g = mm_launch.G(0)
    e = g.e
    log = []
    cur = {"away": 1}
    def rec(name, ret=0):
        def f(e):
            log.append((name, e.arg(0), e.arg(1)))
            return ret
        return f
    for a, n, ret in ((0x5ddf90, 2, 1), (0x5de080, 1, 0), (0x5de0f0, 1, 0), (0x5de130, 1, 0), (0x5dead0, 5, 1), (0x5de000, 4, 1),
                      (0x5dec80, 2, 1), (0x5ded40, 2, 1), (0x5def80, 2, 1), (0x5defb0, 2, 1), (0x58ee80, 0, 0), (0x5dd2b0, 0, 0), (0x5de200, 1, 1)):
        e.hook(a, rec("%x" % a, ret), n)
    e.hook(0x5defe0, lambda e: (log.append(("away", e.arg(1), e.arg(2))), cur["away"])[1], 3)
    e.hook(0x5de190, lambda e: (log.append(("pace", e.arg(0), e.arg(1))), 0)[1], 2)
    e.hook(0x58ed10, lambda e: (log.append(("dropcmd", 0, 0)), 0)[1], 0)
    base = (1, )
    r1, x0, y0, w, h, _ = g.rooms[1]
    sx, sy = x0 + 20, y0 + 20
    for c in range(cases):
        for (rr, *_x) in g.rooms:                      # a fresh scene: every room's death ring empty (units from earlier cases stay in DT)
            for i in range(4): e.w32(rr + 0x38 + 4 * i, 0)
            e.mu.mem_write(rr + 0x14, b"\x00")
        away_ok = rng.choice((0, 1))
        n = rng.choice((0, 1, 1, 2, 3, 4, 5, 6, 7))
        fx, fy = sx + rng.randint(-6, 6), sy + rng.randint(-6, 6)
        F = e.call(0x555230, fx, fy, g.g, r1, 1, 1, 0, ecx=1, edx=fal)
        deaths, kill = [], []
        for k in range(n):
            if rng.random() < 0.7:
                dx, dy = rng.randint(-22, 22), rng.randint(-22, 22)
            else:
                dx, dy = rng.randint(-16, 16), rng.randint(-16, 16)
            px, py = fx + dx, fy + dy
            rm = room_at(g, px, py)
            if not rm:
                continue
            isf = rng.random() < 0.5
            u = e.call(0x555230, px, py, g.g, rm, 1, 1, 0, ecx=1, edx=fal if isf else zom)
            if not u:
                continue
            die = rng.random() < 0.8
            deaths.append((dx, dy, [r for (r, *_x) in g.rooms].index(rm), int(isf), int(die), u))
        for (dx, dy, same, isf, die, u) in deaths:
            if die:
                p = e.alloc(0x40); e.w32(p + 4, u); e.call(0x5a6ff0, ecx=g.g, edx=p)
        ctrl, params, row = e.alloc(0x40), e.alloc(0x40), e.alloc(0x200)
        e.w32(ctrl + 0x14, 7)
        e.w32(params, ctrl); e.w32(params + 0x1c, row)
        cur["away"] = away_ok
        del log[:]
        err = ""
        try:
            e.call(0x5f02c0, params, ecx=g.g, edx=F)
        except Exception as ex:
            err = type(ex).__name__
        names = [l[0] for l in log]
        fled = int(("pace", 0x32, 0) in log)
        away = int("away" in names)
        fell = int(bool(away and not away_ok and names.index("away") < len(names) - 1)) if away else 0
        print("F %d %d %d %d | %s | fled %d pace %d away %d fell %d state %d err %s calls %s" % (
            c, fx, fy, away_ok, " ".join("%d,%d,%d,%d,%d" % d[:5] for d in deaths), fled, int(("pace", 0x32, 0) in log), away, fell, e.r32(ctrl + 0x14), err or "-", ",".join(names)))
main()
