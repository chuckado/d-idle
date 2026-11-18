import sys,os
sys.path.insert(0,os.environ["EMU"])
import capstone, pefile
pe=pefile.PE(os.environ["GAMEEXE"]); base=pe.OPTIONAL_HEADER.ImageBase
s=int(sys.argv[1],16); n=int(sys.argv[2],0)
data=pe.get_data(s-base,n)
md=capstone.Cs(capstone.CS_ARCH_X86,capstone.CS_MODE_32)
for i in md.disasm(data,s): print("%x %s %s"%(i.address,i.mnemonic,i.op_str))
