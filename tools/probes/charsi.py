#!/usr/bin/env python3
"""charsi: what Charsi's menu offers with the Malus in the pack and quest 3 started (card qdmatrix, Tools of the Trade reward
0x204e). Usage: charsi.py D2D DATA SEED. `debug goto`, `debug give`, `debug quest set` are SETUP verbs (cheated probe). Stdlib only."""
import os, re, sys, time
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gridwhole import session
p, cmd = session(sys.argv[1], sys.argv[2], int(sys.argv[3]))
def step(n):
    for _ in range(n): cmd("step 25")
try:
    print(cmd("lockstep on").strip()[:30])
    print("verbs", cmd("debug verbs").replace("\n", " ")[:600])
    print("goto", cmd("debug goto 1").strip()[:40]); step(4)
    print("quest", cmd("debug quest set 3 %s" % os.environ.get("QW", "12")).strip()[:40])
    print("give hdm", cmd("debug give hdm pack").strip()[:50])
    print("give", cmd("debug give wsd 2 pack").strip()[:50] if False else "")
    npcs = cmd("npcs")
    print(npcs[:300].replace("\n", " | "))
    idx = None
    for l in npcs.split("\n"):
        f = l.split("\t")
        if f and f[0] == "Charsi": idx = f[4] if len(f) > 4 else None
    print("charsi index", idx, [l for l in npcs.split("\n") if l.startswith("Charsi")])
    f = [l for l in npcs.split("\n") if l.startswith("Charsi")][0].split("\t")
    idx = f[4]
    print("warp", cmd("debug warp 7 18").strip()[:30]); step(6)
    print("kick", cmd("click 400 300").strip()[:20]); step(6)
    print("interact", cmd("cmd interact %s" % idx).strip()[:30])
    for i in range(80):
        step(4)
        m = cmd("menu")
        if m.count("\n") > 1 and "ok" in m and ("Talk" in m or "Imbue" in m):
            print("menu after", i * 4 * 25, "ticks:", m.replace("\n", " | ")[:400]); break
    print("menu raw:", repr(cmd("menu"))[:500])
    print("click talk", cmd("click 428 97").strip()[:20]); step(20)
    print("menu after talk:", repr(cmd("menu"))[:400])
    print("quest 3", cmd("debug quest words")[:400].replace("\n", " ") if False else cmd("debug quest set 3 12").strip()[:40])
    print("state", cmd("state")[:700].replace("\n", " "))
    print("items", [l[:80] for l in cmd("items").split("\n") if l.startswith("[")][:12])
finally:
    p.kill()
