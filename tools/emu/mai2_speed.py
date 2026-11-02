"""monsterai2 (D7, D8): a monster's path speed, game.exe's own code from FUN_00623f50's moving-mode branch (0x624128..0x6241ce) run on
a fake unit and a fake MonStats row, for Velocity / Run pairs, stat 0x43 values and both moving modes.

    PYTHONPATH=<upstream tools/emu>:tools/emu D2_MPQ_DIR=<dir> python mai2_speed.py

The stat reads (FUN_00621740 for table entry 4, FUN_00625480 for stat 0x43) are hooks (the answers are the scene's); the rate arithmetic,
FUN_00621360 (path velocity: MonStats Velocity << 8) and FUN_006213d0 (animation rate) are game.exe's code. The path speed is what
FUN_00648690 receives. Prints the 12-byte stat table 0x6e8e24 (entry: flag, cap, stat id) and `speed vel run mode stat43 frw path_speed
expected anim_rate`. Never runs a game."""
import struct
import sys

import emu

TABLE = 0x6e8e24


def main():
    e = emu.Emu()
    for i in range(0, 9):
        flag, cap, stat = struct.unpack("<3I", e.read(TABLE + 12 * i, 12))
        print(f"table[{i}] flag {flag} cap {cap} stat {stat:#x}")
    data, rows = e.alloc(0x1000), e.alloc(0x1a8 * 8)
    e.w32(0x744304, data); e.w32(data + 0xa80, 8); e.w32(data + 0xa78, rows)
    unit, path, frame = e.alloc(0x200), e.alloc(0x80), e.alloc(0x400)
    e.w32(unit, 1); e.w32(unit + 0x2c, path)
    cur = {}
    e.hook(0x621740, lambda e: cur["frw"], 0)
    e.hook(0x625480, lambda e: cur["stat43"], 3)
    e.hook(0x648690, lambda e: cur.__setitem__("speed", e.arg(1)) or 0, 2)
    n = bad = 0
    for vel, run in ((1, 3), (3, 3), (5, 5), (6, 8), (8, 8), (4, 4), (2, 2)):
        e.mu.mem_write(rows + 0x1a8 * 5 + 0x32, struct.pack("<hh", vel, run))
        e.mu.mem_write(rows + 0x1a8 * 5 + 0x36, struct.pack("<hh", 90, 110))        # the animation words: not what is asked
        for mode in (2, 0xf):
            for stat43 in (0, 25, 75, 100, 125, 175, 300):
                for frw in (0, 30):
                    cur.update(stat43=stat43, frw=frw)
                    ebp = frame + 0x100
                    e.w32(ebp - 4, 5); e.w32(ebp - 8, mode); e.w32(ebp + 8, 1)
                    e.mu.reg_write(emu.UC_X86_REG_EBP, ebp); e.mu.reg_write(emu.UC_X86_REG_ESP, frame + 0x80)
                    e.mu.reg_write(emu.UC_X86_REG_EDI, unit)
                    e.mu.emu_start(0x624128, 0x6241ce)
                    esi = max(25, frw + stat43)
                    want = (vel << 8) * esi // 100
                    n += 1
                    if cur["speed"] != want:
                        bad += 1
                        print("MISMATCH", vel, run, mode, stat43, frw, cur["speed"], want)
                    if mode == 2 and frw == 0 and stat43 in (75, 175): print("speed", vel, run, hex(mode), stat43, frw, cur["speed"], want)
    print(f"speed cases {n} mismatch {bad}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
