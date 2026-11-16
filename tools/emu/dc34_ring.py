"""decomp34: the Fallen's flee scan (FUN_005f02c0 0x5f032e..0x5f03df) reads the NEAR ROOMS' death rings (room +0x38, four unit ids, head byte +0x14).
Probe: boot a Blood Moor game, spawn monsters in real rooms, report the room list (FUN_00619790) and what the real death handler
(FUN_005a6ff0, caller of FUN_0061afa0 at 0x5a704d) puts into the ring.   usage: dc34_ring.py probe"""
import os, sys
sys.path.insert(0, os.environ.get("DC34", "/data/dc34"))
import mm_launch

def near(e, room):
    p = e.alloc(8)
    e.call(0x619790, room, p, p + 4)
    lst, cnt = e.r32(p), e.r32(p + 4)
    return [e.r32(lst + 4 * i) for i in range(cnt)]

def ring(e, room):
    return [e.r32(room + 0x38 + 4 * i) for i in range(4)], e.r32(room + 0x14) & 0xff

def main():
    g = mm_launch.G(0)
    e = g.e
    print("rooms", len(g.rooms))
    for k, (r, x0, y0, w, h, _) in enumerate(g.rooms[:6]):
        nr = near(e, r)
        print("room %d %#x at %d,%d size %d x %d near=%d self_in_near=%s near=%s" % (k, r, x0, y0, w, h, len(nr), r in nr, [hex(a) for a in nr]))
        print("  ring", ring(e, r))
    # a Fallen class: hcIdx from monstats
    mons = {m["Id"]: m for m in mm_launch.rows(sys.argv[-1], "monstats.txt")} if os.path.isdir(sys.argv[-1]) else {}
    print("have monstats", len(mons))

def die(e, g, u):
    p = e.alloc(0x40)
    e.w32(p + 4, u)
    return e.call(0x5a6ff0, ecx=g.g, edx=p)

def probe2(g, mons):
    e = g.e
    fal = int(mons["fallen1"]["hcIdx"]); zom = int(mons["zombie1"]["hcIdx"])
    r, x0, y0, w, h, _ = g.rooms[1]
    us = []
    for k, cls in enumerate([fal, fal, zom, fal, fal, fal]):
        u = e.call(0x555230, x0 + 5 + 4 * k, y0 + 20, g.g, r, 1, 1, 0, ecx=1, edx=cls)
        us.append(u)
        print("spawn", k, cls, hex(u), "id", hex(e.r32(u + 0xc)) if u else None, "mode", e.r32(u + 0x10) if u else None, "room", hex(e.call(0x620bb0, u)) if u else None)
    for k, u in enumerate(us[:5]):
        try:
            die(e, g, u)
            print("die", k, "mode", e.r32(u + 0x10), "ring", [hex(a) for a in ring(e, r)[0]], ring(e, r)[1])
        except Exception as ex:
            print("die", k, "ERR", type(ex).__name__, str(ex)[:120]); break

if sys.argv[1] == "probe2":
    g = mm_launch.G(0)
    mons = {m["Id"]: m for m in mm_launch.rows(sys.argv[-1], "monstats.txt")}
    probe2(g, mons)
else:
    main()
