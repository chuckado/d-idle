"""monsterai: dump game.exe's MonAI table (0x73ca18, 16 bytes a row: search type, think at +4 / +8, other think at +12 ...) with the MonAI.txt names.
usage (worker host, upstream tools/emu on the path): D2_MPQ_DIR=<data dir with bin/game.exe> python mai_table.py <monai.txt>"""
import sys
import emu

e = emu.Emu()
names = [l.split("\t")[0] for l in open(sys.argv[1], encoding="latin1").read().splitlines()[1:]]
for i, n in enumerate(names):
    row = 0x73ca18 + i * 16
    print(i, n, *(f"{e.r32(row + k):x}" for k in (0, 4, 8, 12)))
