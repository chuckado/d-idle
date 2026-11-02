# SPDX-License-Identifier: GPL-3.0-or-later
"""Find the code that reads a stat id in game.exe 1.14d (agent monsterstats): every `push imm8 STAT` and
`mov edx, STAT` in .text, with the call that follows (to name the getter).

    tools/emu/ms_run.sh ms_scan.py 0x4a        # who reads stat 0x4a (life regen)
"""
import sys

import drlg

stat = int(sys.argv[1], 0)
e = drlg.boot()
base, size = 0x401000, 0x2d0000
img = e.read(base, size)
if len(sys.argv) > 2 and sys.argv[2] == "callers":    # every `call stat` (stat = a code address)
    i = img.find(b"\xe8")
    while i >= 0:
        if base + i + 5 + int.from_bytes(img[i + 1:i + 5], "little", signed=True) == stat: print(f"{base + i:#x} call {stat:#x}")
        i = img.find(b"\xe8", i + 1)
    raise SystemExit
pats = {bytes([0x6a, stat]): "push", bytes([0xba, stat, 0, 0, 0]): "mov edx"}
for pat, name in pats.items():
    i = img.find(pat)
    while i >= 0:
        nxt = img[i + len(pat):i + len(pat) + 12]
        call = ""
        for k in range(len(nxt) - 4):
            if nxt[k] == 0xe8:
                rel = int.from_bytes(nxt[k + 1:k + 5], "little", signed=True)
                call = f" call {base + i + len(pat) + k + 5 + rel:#x}"
                break
        print(f"{base + i:#x} {name} {stat:#x}{call}")
        i = img.find(pat, i + 1)
