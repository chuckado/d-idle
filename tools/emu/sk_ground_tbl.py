# SPDX-License-Identifier: GPL-3.0-or-later
"""SKGROUND: dump the patched 1.14d Skills.txt rows of the ground / area skills
and every Missiles.txt row they reach (srvmissile*, SubMissile*, HitSubMissile*,
recursively), as game.exe's MPQ stack serves them (emu.Emu().mpq_file).

    python sk_ground_tbl.py skills [ID...]     # non-empty columns (default: the six skills)
    python sk_ground_tbl.py missiles NAME...   # non-empty columns of the named rows + their subs
    python sk_ground_tbl.py tree [ID...]       # skills, then the missile closure
    python sk_ground_tbl.py dw ADDR N          # N dwords of game.exe's image at ADDR
    python sk_ground_tbl.py find HEX...        # image addresses holding these dwords (tables)

Runs next to upstream's tools/emu (emu.py), like sk_pass_tbl.py.
"""
import os
import sys

sys.path.insert(0, os.getcwd())
import emu  # noqa: E402

SIX = ("59", "56", "64", "57", "62", "51")   # Blizzard, Meteor, Frozen Orb, Thunder Storm, Hydra, Fire Wall
MISSILE_REFS = ("srvmissile", "srvmissilea", "srvmissileb", "srvmissilec", "cltmissile",
                "cltmissilea", "cltmissileb", "cltmissilec", "cltmissiled")
SUB_REFS = ("SubMissile1", "SubMissile2", "SubMissile3", "HitSubMissile1", "HitSubMissile2",
            "HitSubMissile3", "HitSubMissile4", "Explosion")


def table(e, name):
    t = e.mpq_file("data\\global\\excel\\%s.txt" % name).decode("latin-1").splitlines()
    head = t[0].split("\t")
    return head, [dict(zip(head, ln.split("\t"))) for ln in t[1:]]


def show(title, r):
    print("----", title)
    for k, v in r.items():
        if v not in ("", "0") and k:
            print("  %s = %s" % (k, v))


def find(values):
    """Image addresses whose dword equals one of `values` (function tables)."""
    import struct
    import pefile
    pe = pefile.PE(os.path.join(emu.data_path(), "bin", "game.exe"), fast_load=True)
    img, base = pe.get_memory_mapped_image(), pe.OPTIONAL_HEADER.ImageBase
    for v in values:
        pat = struct.pack("<I", v)
        at = img.find(pat)
        while at >= 0:
            print("%08x at %08x" % (v, base + at))
            at = img.find(pat, at + 1)


def main():
    what = sys.argv[1] if len(sys.argv) > 1 else "tree"
    args = sys.argv[2:]
    if what == "find":
        find([int(a, 16) for a in args])
        return
    e = emu.Emu()
    if what == "dw":
        a, n = int(args[0], 16), int(args[1])
        print(" ".join("%08x" % e.r32(a + 4 * i) for i in range(n)))
        return
    _, mrows = table(e, "Missiles")
    by_name = {r.get("Missile"): (i, r) for i, r in enumerate(mrows)}
    want = []
    if what in ("skills", "tree"):
        _, srows = table(e, "Skills")
        ids = args or SIX
        for r in srows:
            if r.get("Id") in ids:
                show("skill %s %s" % (r.get("Id"), r.get("skill")), r)
                want += [r.get(k, "") for k in MISSILE_REFS if r.get(k, "")]
        if what == "skills":
            return
    else:
        want = list(args)
    seen = set()
    while want:
        n = want.pop(0)
        if n in seen or n not in by_name:
            continue
        seen.add(n)
        i, r = by_name[n]
        show("missile %d %s" % (i, n), r)
        want += [r.get(k, "") for k in SUB_REFS if r.get(k, "")]


if __name__ == "__main__":
    main()
