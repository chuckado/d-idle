# SPDX-License-Identifier: GPL-3.0-or-later
"""SKGROUND2: dump a Missiles.txt row's record (0x1a4 bytes) as game.exe holds it.
    python sk_gnd2_row.py ID [ID...]   # prints the non-zero bytes/words by offset
"""
import os
import sys

sys.path.insert(0, os.getcwd())
import drlg  # noqa: E402

e = drlg.boot()
rows = e.r32(e.r32(0x744304) + 0xb64)
for mid in (int(a) for a in sys.argv[1:]):
    d = e.read(rows + mid * 0x1a4, 0x1a4)
    print(f"row {mid}: " + " ".join(f"{i:#x}={d[i]}" for i in range(0x1a4) if d[i]))
