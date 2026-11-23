#!/usr/bin/env python3
"""Exploration run: prints verb formats. Usage: skeleton_speed_explore.py <d2d>"""
import os, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from skeleton_speed_lib import Session
s = Session(sys.argv[1], ["--lab-verbs", "--read-combat", "--lockstep", "--no-save"])
try:
    for v in ("state", "clock", "caps", "debug help", "debug level"):
        print(">>", v); print(s.cmd(v)[:3500])
    print(">> goto", s.cmd("debug goto 2"))
    print(s.cmd("step 100")); print(s.cmd("state")); print(s.cmd("debug minions"))
finally:
    s.close()
