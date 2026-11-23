import sys,time
sys.path.insert(0,'/home/jesse/code/d2lod-wt/qd18/tools/probes')
from gridwhole import session
p,cmd=session(sys.argv[1],sys.argv[2],11)
try:
    print(cmd("debug goto 80").strip()[:40]); time.sleep(3)
    for c in ("debug warps","debug warptriggers"):
        print(c); print(cmd(c)[:6000])
finally:
    p.kill()
