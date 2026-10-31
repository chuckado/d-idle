#!/usr/bin/env python3
"""Per-run goldfloor metrics from bot logs (bot-A<seed>.log = levers on, bot-B<seed>.log = off), plus paired.py input.
usage: an_gold.py LOGDIR OUTDIR   writes OUTDIR/metrics.txt and OUTDIR/p_<metric>.txt (one an.py-shaped line per run, arm on=A off=B,
value in the xp/h field) for tools/ab/paired.py. Virtual minutes are the `minute` lines (hours = lines / 60)."""
import json, os, re, sys, datetime as dt

def ts(l): return dt.datetime.fromisoformat(l["time"][:26]).timestamp()

def run(path):
    mins, buys, trips, fresh, ex = [], [], [], 0, None
    for raw in open(path, errors="replace"):
        if raw.startswith("exit "): ex = raw.split()[1]; continue
        if "starting a fresh game" in raw: fresh += 1
        if '"msg":"minute"' in raw or '"msg":"shop: bought"' in raw or '"msg":"town trip"' in raw:
            try: l = json.loads(raw)
            except ValueError: continue
            if l["msg"] == "minute": mins.append(l)
            elif l["msg"] == "town trip": trips.append(l)
            else: buys.append(l)
    return mins, buys, trips, fresh, ex

def kind(item):
    i = item.lower()
    if "town portal" in i: return "tp"
    if "healing" in i: return "hp"
    if "mana" in i: return "mp"
    return "other"

def metrics(path):
    mins, buys, trips, fresh, ex = run(path)
    if not mins: return None
    n = len(mins); H = n / 60.0; last = mins[-1]
    mt = [ts(m) for m in mins]
    def clvl_at(t):
        c = mins[0]["clvl"]
        for m, x in zip(mins, mt):
            if x <= t: c = m["clvl"]
            else: break
        return c
    spend = {"hp": 0, "mp": 0, "tp": 0, "other": 0}; nbuy = dict(spend); tp_band = 0
    for b in buys:
        k = kind(b["item"]); spend[k] += b["price"]; nbuy[k] += 1
        if k == "tp" and 5 <= clvl_at(ts(b)) <= 10: tp_band += b["price"]
    band_min = sum(1 for m in mins if 5 <= m["clvl"] <= 10)
    # trip spend: purchases between this trip line and the next one
    tt = [ts(t) for t in trips] + [1e18]; bt = [(ts(b), b["price"]) for b in buys]
    tspend = [sum(p for x, p in bt if tt[i] <= x < tt[i + 1]) for i in range(len(trips))]
    tot = sum(spend.values())
    gold = [m["gold"] for m in mins]
    income = last["gold"] - mins[0]["gold"] + tot  # piles + sales (death losses count against it)
    return dict(log=os.path.basename(path), hours=H, minutes=n, clvl=last["clvl"], kills=last["kills"], deaths=last["deaths"],
        kills_h=last["kills"] / H, xp_h=last["xp"] / H, trips=len(trips), trips_h=len(trips) / H,
        trips_lt30=sum(1 for s in tspend if s < 30), trips_lt30_pct=100.0 * sum(1 for s in tspend if s < 30) / max(len(trips), 1),
        spend_h=tot / H, income_h=income / H, gold_mean=sum(gold) / n, tp_h=spend["tp"] / H, hp_h=spend["hp"] / H, mp_h=spend["mp"] / H,
        tp_band_h=(tp_band / (band_min / 60.0)) if band_min >= 10 else float("nan"), band_min=band_min,
        potions_bought=nbuy["hp"], mana_bought=nbuy["mp"], tp_bought=nbuy["tp"],
        deaths_1000k=1000.0 * last["deaths"] / max(last["kills"], 1),
        belt_empty_pct=100.0 * sum(1 for m in mins if m.get("potions", 0) == 0) / n,
        restarts=last.get("restarts", 0), fresh=fresh, exit=ex)

if __name__ == "__main__":
    ld, od = sys.argv[1:3]; os.makedirs(od, exist_ok=True)
    rows = {}
    for f in sorted(os.listdir(ld)):
        m = re.match(r"bot-([AB])(\d+)\.log$", f)
        if m and (r := metrics(os.path.join(ld, f))): rows[(m[1], int(m[2]))] = r
    keys = ["xp_h", "kills_h", "trips_h", "trips_lt30_pct", "spend_h", "income_h", "gold_mean", "tp_h", "tp_band_h", "hp_h", "mp_h", "potions_bought",
            "deaths_1000k", "belt_empty_pct"]
    with open(od + "/metrics.txt", "w") as o:
        for (a, s), r in sorted(rows.items(), key=lambda x: (x[0][1], x[0][0])):
            o.write(("%s%d " % (a, s)) + " ".join("%s=%s" % (k, ("%.1f" % v) if isinstance(v, float) else v) for k, v in r.items() if k != "log") + "\n")
    for k in keys:
        with open(od + "/p_%s.txt" % k, "w") as o:
            for (a, s), r in sorted(rows.items(), key=lambda x: (x[0][1], x[0][0])):
                v = r[k]
                if v != v: continue
                o.write("logs/bot-%s%d.log %s kills=%d (%.1f/h) xp/h=%.3f\n" % ("on" if a == "A" else "off", s, "x", r["kills"], r["kills_h"], v))
