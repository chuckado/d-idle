#!/usr/bin/env python3
"""strips: for every border link of the given levels, is there an approach cell behind the map edge? (card toolstrade5)

Usage: strips.py D2D DATA LEVEL[,LEVEL..] SEED     (lab d2d, --headless --lab-verbs; `debug goto` is SETUP: cheated mechanics probe)
Per level: `debug goto`, `debug level` (size in tiles), `debug warps` (tile exits), `debug nearby` (levels linked past the edge,
with their origin in tiles relative to ours), `debug grid 5` (one char per subtile). A neighbour that is no tile exit is a BORDER
link; its side is taken from the sign of its offset (U: heuristic, the link table has no side). The edge band is the first or last
BAND grid rows or columns on that side. Verdict per link, one JSON line:
  open     an open cell in the band joins an open region of at least MIN cells (the walk can reach it)
  pocket   open cells in the band, but only a small region: a door or object shuts it (nav gatedoor.go handles doors; U)
  closed   no open cell in the band at all: the edge is two or more rows short of the strip (U: the reader must walk the gap)
  noside   the offset gives no clear side
Exit code 1 when any link is pocket or closed (a listing, not a failure of the bot: the nav code decides what it does). Stdlib only."""
import json, sys, time
sys.path.insert(0, __file__.rsplit("/", 1)[0])

BAND, MIN = 3, 400

def flood(g, sx, sy):
    h, w = len(g), len(g[0])
    seen, stack = {(sx, sy)}, [(sx, sy)]
    while stack:
        x, y = stack.pop()
        for nx, ny in ((x + 1, y), (x - 1, y), (x, y + 1), (x, y - 1)):
            if 0 <= nx < w and 0 <= ny < h and (nx, ny) not in seen and g[ny][nx] == ".":
                seen.add((nx, ny))
                stack.append((nx, ny))
    return len(seen)

def verdict(g, side, band=BAND, need=MIN):
    """side N, S, W or E: the band of rows or columns next to that edge of grid g (rows of '.' and '#')."""
    h, w = len(g), len(g[0])
    if side == "N": cells = [(x, y) for y in range(band) for x in range(w)]
    elif side == "S": cells = [(x, y) for y in range(h - band, h) for x in range(w)]
    elif side == "W": cells = [(x, y) for x in range(band) for y in range(h)]
    elif side == "E": cells = [(x, y) for x in range(w - band, w) for y in range(h)]
    else: return "noside"
    open_ = [(x, y) for x, y in cells if g[y][x] == "."]
    if not open_: return "closed"
    best = 0
    seen = set()
    for x, y in open_:
        if (x, y) in seen: continue
        n = flood(g, x, y)
        best = max(best, n)
        if best >= need: return "open"
        seen.add((x, y))
    return "pocket"

def side_of(ox, oy, w, h):
    """side of a neighbour whose origin is (ox, oy) tiles from ours, own size w x h tiles (U: heuristic)."""
    if oy <= -1 and -w < ox < w and abs(oy) >= abs(ox) / 4: return "N"
    if oy >= h - 1 and -w < ox < w: return "S"
    if ox <= -1 and abs(ox) > abs(oy): return "W"
    if ox >= w - 1: return "E"
    return None

def main():
    from gridwhole import session, grid
    d2d, data, levels, seed = sys.argv[1], sys.argv[2], [int(x) for x in sys.argv[3].split(",")], int(sys.argv[4])
    p, cmd = session(d2d, data, seed)
    bad = 0
    try:
        for lv in levels:
            cmd("debug goto %d" % lv)
            time.sleep(1.5)
            L = cmd("debug level").split()
            w, h = int(L[4]), int(L[5])
            warps = {int(l.split()[1]) for l in cmd("debug warps").split("\n") if len(l.split()) >= 2 and l.split()[1].isdigit()}
            g = grid(cmd)
            near = {}
            t0 = time.time()
            while time.time() - t0 < 6:   # d2d builds neighbours on a worker thread: poll until the list stops growing
                cur = {}
                for l in cmd("debug nearby").split("\n"):
                    f = l.split()
                    if len(f) == 3 and all(x.lstrip("-").isdigit() for x in f):
                        cur[int(f[0])] = (int(f[1]), int(f[2]))
                if cur == near: break
                near = cur
                time.sleep(1)
            for no, (ox, oy) in sorted(near.items()):
                if no in warps: continue
                side = side_of(ox, oy, w, h)
                v = verdict(g, side) if g and side else "noside"
                if v in ("pocket", "closed"): bad += 1
                print(json.dumps({"level": lv, "to": no, "side": side, "verdict": v, "offset": [ox, oy], "size": [w, h]}), flush=True)
    finally:
        p.kill()
    sys.exit(1 if bad else 0)

if __name__ == "__main__":
    main()
