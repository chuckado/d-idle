import os, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from skeleton_speed_lib import Session
s = Session(sys.argv[1], ["--lab-verbs", "--read-combat", "--lockstep", "--no-save"], cls=1, seed=int(sys.argv[2]))
try:
    s.cmd("debug goto 2"); s.cmd("step 100")
    print(s.cmd("debug blocked 44.2 24.2"), s.cmd("debug blocked 5 5"))
    g = s.cmd("debug grid 1")
    open(sys.argv[3], "w").write(g)
finally:
    s.close()
