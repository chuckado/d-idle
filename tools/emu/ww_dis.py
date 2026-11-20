# decomp23 helper: disassemble game.exe (capstone over the PE image) or dump dwords.
#   ww_dis.py dis ADDR LEN      ww_dis.py dw ADDR COUNT
import sys, os, struct
import capstone, pefile
pe = pefile.PE(os.environ["GAMEEXE"]); base = pe.OPTIONAL_HEADER.ImageBase
mode, s, n = sys.argv[1], int(sys.argv[2], 16), int(sys.argv[3], 0) if len(sys.argv) > 3 else 0
if mode == "callers":      # ww_dis.py callers ADDR : every E8 / E9 rel32 call or jump to ADDR in .text, with the containing function start guess
    text = pe.get_data(0x1000, 0x2f0000)
    import re
    for m in re.finditer(rb"[\xe8\xe9]", text):
        i = m.start()
        if i + 5 <= len(text):
            rel = struct.unpack_from("<i", text, i + 1)[0]
            if (0x401000 + i + 5 + rel) & 0xffffffff == s: print("%x %s" % (0x401000 + i, "call" if text[i] == 0xe8 else "jmp"))
    sys.exit(0)
if mode == "dis":
    md = capstone.Cs(capstone.CS_ARCH_X86, capstone.CS_MODE_32)
    for i in md.disasm(pe.get_data(s - base, n), s):
        print("%x  %s %s" % (i.address, i.mnemonic, i.op_str))
else:
    import struct
    d = pe.get_data(s - base, 4 * n)
    for k in range(n):
        print("%x: %x" % (s + 4 * k, struct.unpack_from("<I", d, 4 * k)[0]))
