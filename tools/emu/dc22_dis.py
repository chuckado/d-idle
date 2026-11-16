# SPDX-License-Identifier: GPL-3.0-or-later
"""dc22_dis.py START LEN: disassemble game.exe 1.14d bytes (capstone over the PE image), read only.
Run: source $D2_EMU_DIR/env.sh; $PY tools/emu/dc22_dis.py 0x57edd0 0x1c0"""
import os
import sys

import capstone
import pefile

pe = pefile.PE(os.environ["GAMEEXE"])
base = pe.OPTIONAL_HEADER.ImageBase
s = int(sys.argv[1], 16)
n = int(sys.argv[2], 0)
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_32)
for i in md.disasm(pe.get_data(s - base, n), s):
    print("%x %s %s" % (i.address, i.mnemonic, i.op_str))
