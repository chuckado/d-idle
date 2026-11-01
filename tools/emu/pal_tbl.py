# SPDX-License-Identifier: GPL-3.0-or-later
"""palskill: read dword tables out of game.exe's image (no emulator).
    python3 pal_tbl.py 73c960 12      # 12 dwords from 0x73c960 (the missile pSrvDmgFunc table)
"""
import sys

import pefile

pe = pefile.PE("/tmp/ug-emu/bin/game.exe")
base = pe.OPTIONAL_HEADER.ImageBase
addr, n = int(sys.argv[1], 16), int(sys.argv[2])
for i in range(n):
    a = addr + 4 * i
    data = pe.get_data(a - base, 4)
    print(f"{a:x}[{i}] = {int.from_bytes(data, 'little'):#x}")
