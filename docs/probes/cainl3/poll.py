#!/usr/bin/env python3 -I
"""CAIN-L3 driver: plays the part of the player's taps on the quest board (GET /api/quests, POST start / handin)
for search_for_cain, logs every change of the board view with a wall clock, and exits when the report chunk is done,
the board gives up, or the deadline passes. Token from env WTOKEN_FILE (never printed)."""
import json, os, sys, time, urllib.request, urllib.error

view, ans, quest = sys.argv[1], sys.argv[2], "search_for_cain"
deadline = time.time() + float(sys.argv[3]) * 60
tok = open(os.environ["WTOKEN_FILE"]).read().strip()
H = {"Authorization": "Bearer " + tok}

def get():
    r = urllib.request.Request(view + "/api/quests", headers=H)
    return json.load(urllib.request.urlopen(r, timeout=10))

def post(path, body="{}"):
    r = urllib.request.Request(ans + path, data=body.encode(), method="POST", headers=dict(H, **{"Content-Type": "application/json"}))
    try:
        return urllib.request.urlopen(r, timeout=10).read().decode()
    except urllib.error.HTTPError as e:
        return "HTTP %d %s" % (e.code, e.read().decode()[:200])

def log(*a):
    print(time.strftime("%H:%M:%S"), *a, flush=True)

last = None
taps = 0
while time.time() < deadline:
    try:
        v = get()
    except Exception as e:
        log("get failed", repr(e)); time.sleep(5); continue
    run, aw, bl = v.get("running"), v.get("awaiting"), v.get("blocked")
    sig = json.dumps([run and [run["chunk"], run["state"], run.get("step"), run.get("why")], aw and [aw["chunk"], aw["state"]], bl and [bl["chunk"], bl.get("why")],
                      v.get("handIn"), [h.get("chunk") for h in v.get("history", []) if h.get("quest") == quest]])
    if sig != last:
        log("board", sig, "meter=", (v.get("meter") or {}).get("line"))
        last = sig
    hist = [h for h in v.get("history", []) if h.get("quest") == quest]
    if any(h.get("chunk") == "report" for h in hist):
        log("DONE report chunk in history"); sys.exit(0)
    if bl and bl.get("quest") == quest:
        log("BLOCKED", json.dumps(bl)); sys.exit(3)
    if aw and aw.get("quest") == quest and not v.get("handIn"):
        taps += 1; log("tap handin", aw["chunk"], post("/api/quests/%s/handin" % quest))
    elif not run and not aw:
        for o in v.get("offers", []):
            if o.get("quest") == quest and o.get("startable"):
                taps += 1; log("tap start", o["chunk"], post("/api/quests/%s/start" % quest)); break
        else:
            if taps == 0 and int(time.time()) % 30 == 0:
                log("no startable offer yet", [(o["quest"], o["chunk"], o.get("startable"), o.get("why")) for o in v.get("offers", []) if o.get("quest") == quest])
    time.sleep(5)
log("DEADLINE"); sys.exit(4)
