"""compare dc34_telemana.py game output (`id did x y pays ...`) with test_tele_mana --tele-mana output (`id did paid`).  usage: dc34_tmcmp.py game.txt d2d.txt [cost]
exit 1 on empty or mismatched-length input."""
import sys
g = [l.split() for l in open(sys.argv[1]) if l[:1].isdigit()]
d = [l.split() for l in open(sys.argv[2]) if l[:1].isdigit()]
if not g or len(g) != len(d): sys.exit("empty or unequal: %d game, %d d2d" % (len(g), len(d)))
cost = int(sys.argv[3]) if len(sys.argv) > 3 else 6144
k = {}
for a, b in zip(g, d):
    assert a[0] == b[0]
    gdid, gpaid = int(a[1]), cost * int(a[4])
    ddid, dpaid = int(b[1]), int(b[2]); dpaid = cost if abs(dpaid - cost) < 1024 else (0 if abs(dpaid) < 1024 else dpaid)   # regeneration noise: under 400
    key = ("same" if (gdid, gpaid) == (ddid, dpaid) else "DIFF") + " game(did=%d paid=%d) d2d(did=%d paid=%d)" % (gdid, gpaid, ddid, dpaid)
    k[key] = k.get(key, 0) + 1
for key, n in sorted(k.items()): print(n, key)
sys.exit(1 if any(x.startswith("DIFF") for x in k) else 0)
