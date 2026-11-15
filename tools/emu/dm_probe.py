# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp31 probe: roll speed and table headers."""
import time, random
import drops
o = drops.Oracle()
e = o.e
for f in ("Weapons", "Armor", "Misc", "MonStats", "SuperUniques", "Levels"):
    r = drops.txt_rows(e, f)
    print(f, len(r), list(r[0].keys())[:60])
rng = random.Random(1)
t = time.time()
for _ in range(2000):
    o.roll(rng.getrandbits(32), "Act 1 H2H A", 0, 5, 1, 0)
print("2000 rolls", time.time() - t)
