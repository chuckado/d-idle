# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp28: list the call sites (E8 rel32) of game.exe functions, with the instructions leading to each.
    $PY dc28_callers.py 5809d0 [more addrs]"""
import sys, os, struct
sys.path.insert(0, os.environ["EMU"])
import capstone, pefile
pe = pefile.PE(os.environ["GAMEEXE"]); base = pe.OPTIONAL_HEADER.ImageBase
sec = [s for s in pe.sections if s.Name.startswith(b".text")][0]
code = sec.get_data(); lo = base + sec.VirtualAddress
md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_32)
for t in sys.argv[1:]:
    tgt = int(t, 16)
    print("== callers of %x" % tgt)
    for off in range(len(code) - 5):
        if code[off] == 0xE8 and lo + off + 5 + struct.unpack("<i", code[off+1:off+5])[0] == tgt:
            a = lo + off
            print("-- %x" % a)
            for j in md.disasm(code[max(0, off - 40):off + 5], lo + max(0, off - 40)):
                print("   %x %s %s" % (j.address, j.mnemonic, j.op_str))
