# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp31 DROPMODEL: what an Act 1 Normal area / monster / boss drops, from game.exe's own roll function.

Input: the room population of docs/research/areamon-trace.md (tools/emu/ms_spawn.py output, `U lid seed n class flags statlvl f7 stat:value ..`,
game.exe's FUN_00573cb0 writes under the emulator). Each unit gets the treasure class game.exe gives it (FUN_005a6600, rule of
docs/research/treasure-trace.md section 1: superunique -> SuperUniques TC; flag 4 -> TreasureClass2; flag 8 -> TreasureClass3; else
TreasureClass1) and item level = its stat 12. For every distinct (class, ilvl) the native roll FUN_0055a6d0 is run N times (one player,
magic find 0, gold find 0) through upstream's drops.Oracle, and every made item is tallied by (code, quality). The exact odds of the class
records (tre_act.Table) are compared with the native tally (V run check of the table math).

  bash tools/emu/tre_run.sh dm_model.py POP.txt OUT.json [N]
Needs the upstream emu modules and the image of docs/emu-image.md (worker host only).
"""
import json
import math
import random
import re
import sys
import zlib
from collections import defaultdict

import drops
import tre_act

POP, OUT = sys.argv[1], sys.argv[2]
N = int(sys.argv[3]) if len(sys.argv) > 3 else 10000

QNAME = {1: "low", 2: "normal", 3: "superior", 4: "magic", 5: "set", 6: "rare", 7: "unique", 8: "crafted"}
GRADE = {"chipped": "chipped", "flawed": "flawed", "flawless": "flawless", "perfect": "perfect"}


def lvl_band(lr):
    return "early" if lr <= 10 else "mid" if lr <= 25 else "late"


class Items:
    """item code -> (file, type, name, levelreq)"""
    def __init__(self, e):
        self.m = {}
        for f in ("Weapons", "Armor", "Misc"):
            for r in drops.txt_rows(e, f):
                if r.get("code"):
                    self.m[r["code"]] = (f, r.get("type", ""), r.get("name", ""), int(r.get("levelreq") or 0))

    def key(self, code, q):
        """the flat model keys one made item adds to (a list)."""
        f, t, name, lr = self.m.get(code, ("", "", "", 0))
        if code == "gld":
            return []
        if re.fullmatch(r"r\d\d", code):
            return ["rune", f"rune_{code}"]
        if t == "hpot": return [f"potion_{code}", "potion_hp"]
        if t == "mpot": return [f"potion_{code}", "potion_mp"]
        if t == "rpot": return ["potion_rejuv"]
        if t in ("spot", "apot", "wpot"): return ["potion_other"]
        if code == "tsc": return ["scroll_town_portal"]
        if code == "isc": return ["scroll_identify"]
        if code == "tbk": return ["tome_town_portal"]
        if code == "ibk": return ["tome_identify"]
        if code == "key": return ["key"]
        if t.startswith("gem") or t == "gem":
            g = name.split(" ")[0].lower()
            return ["gem", f"gem_{GRADE.get(g, 'standard')}"]
        if t in ("ring", "amul"):
            return [f"{'ring' if t == 'ring' else 'amulet'}_{QNAME[q]}", "jewelry_" + ("plain" if q in (1, 2, 3) else QNAME[q])]
        if t == "jewl": return [f"jewel_{QNAME[q]}"]
        if t in ("scha", "mcha", "lcha"): return [f"charm_{t}_{QNAME[q]}", f"charm_{QNAME[q]}"]
        if t in ("bowq", "xboq"): return ["ammo"]
        if f in ("Weapons", "Armor"):
            qq = "plain" if q in (1, 2) else QNAME[q]
            kind = "weapon" if f == "Weapons" else "armor"
            return [f"gear_{qq}", f"gear_{kind}_{qq}", f"gear_{lvl_band(lr)}_{qq}"]
        return ["misc"]


def mult(code):  # amount made by one entry: always 1 item
    return 1


def main():
    o = drops.Oracle()
    e = o.e
    tab = tre_act.Table(o)
    im = Items(e)
    ms = {r["hcIdx"]: r for r in drops.txt_rows(e, "MonStats") if r.get("Id") and r.get("hcIdx", "").isdigit()}
    sup = defaultdict(list)
    for r in drops.txt_rows(e, "SuperUniques"):
        if r.get("Superunique") and r["Superunique"] != "Expansion":
            sup[r["Class"].lower()].append(r)
    levels = {int(r["Id"]): r["LevelName"] if "LevelName" in r else r["Name"] for r in drops.txt_rows(e, "Levels") if r.get("Id", "").isdigit()}
    # population
    area = defaultdict(lambda: defaultdict(list))        # lid -> seed -> [(class, flags, ilvl)]
    for line in open(POP):
        f = line.split()
        if not f or f[0] != "U" or f[6] == "None":
            continue
        lid, seed, cls, flags, statlvl = int(f[1]), int(f[2]), int(f[4]), int(f[5], 16), int(f[6])
        st = {}
        for tok in f[8:]:
            k, _, v = tok.partition(":")
            try: st[int(k, 16)] = int(v)
            except ValueError: pass
        if st.get(0xd, 0) <= 0:
            continue
        area[lid][seed].append((cls, flags, st.get(0xc, statlvl)))
    unknown, notc = defaultdict(int), defaultdict(float)

    def unit_tc(cls, flags):
        r = ms[str(cls)]
        mid = r["Id"].lower()
        if flags & 2:
            rows = sup.get(mid, [])
            if len(rows) == 1:
                return rows[0]["TC"], "boss", rows[0]["Name"]
            unknown[mid] += 1
            return r["TreasureClass3"], "boss", mid
        if flags & 4: return r["TreasureClass2"], "champion", r["Id"]
        if flags & 8: return r["TreasureClass3"], "unique", r["Id"]
        return r["TreasureClass1"], ("minion" if flags & 0x10 else "plain"), r["Id"]

    # per (tc, ilvl): n made per roll, by key
    cache = {}
    zrows = []

    def roll_key(tc, il):
        k = (tc, il)
        if k in cache:
            return cache[k]
        rng = random.Random(zlib.crc32(f"{tc}|{il}".encode()))
        tally = defaultdict(float)
        cats = defaultdict(list)
        per_roll = []
        for _ in range(N):
            got = tre_act.parse_roll(o.roll(rng.getrandbits(32), tc, 0, il, 1, 0))
            c = defaultdict(int)
            for code, q, m in got:
                for kk in im.key(code, q):
                    tally[kk] += 1
                cc = tab.cat(code)
                c[cc] += 1
                if code != "gld":
                    c["items"] += 1
                c["all"] += 1
            per_roll.append(c)
        vec = {kk: v / N for kk, v in tally.items()}
        # z against the exact odds of the class record
        poly, exp, piles = tab.one(o.index[tc])
        for cat in ("hp", "mp", "scroll", "gear", "rune", "gem", "ring", "amulet", "jewel", "charm", "gold", "key", "rv"):
            xs = [pr.get(cat, 0) for pr in per_roll]
            m = sum(xs) / N
            sd = (sum((x - m) ** 2 for x in xs) / N) ** 0.5
            ex = exp.get(cat, 0.0)
            sd = sd or math.sqrt(max(ex, 1.0 / N))      # a cell that never dropped: Poisson spread of the exact mean
            z = (m - ex) / (sd / math.sqrt(N))
            zrows.append((tc, il, cat, ex, m, z))
        vec["_drops"] = sum(len(p) for p in [per_roll]) and sum(pr.get("all", 0) for pr in per_roll) / N
        vec["_nodrop"] = sum(1 for pr in per_roll if not pr.get("all")) / N
        vec["_piles"] = sum(pr.get("gold", 0) for pr in per_roll) / N
        cache[k] = vec
        return vec

    out = {"schema": "dropmodel/1", "difficulty": "normal", "n_rolls_per_class_ilvl": N, "areas": {}, "monsters": {}, "bosses": {}}
    mon_acc = defaultdict(lambda: defaultdict(float))     # (name, kind) -> weights
    mon_meta = {}
    for lid in sorted(area):
        seeds = area[lid]
        ns = len(seeds)
        per_kind = defaultdict(lambda: defaultdict(float))
        count = defaultdict(float)
        tot = defaultdict(float)
        for seed, us in seeds.items():
            for cls, flags, il in us:
                tc, kind, name = unit_tc(cls, flags)
                il = max(il, 1)
                if not tc or tc not in o.index:
                    notc[(lid, f"{name}:{kind}")] += 1 / ns
                    count[kind] += 1 / ns
                    continue
                vec = roll_key(tc, il)
                count[kind] += 1 / ns
                for kk, v in vec.items():
                    if kk.startswith("_"): continue
                    per_kind[kind][kk] += v / ns
                    tot[kk] += v / ns
                if kind != "boss":
                    ma = mon_acc[(name.lower(), kind)]
                    ma["n"] += 1
                    ma["ilvl"] += il
                    mon_meta[(name.lower(), kind)] = tc
                    for kk, v in vec.items():
                        ma[kk] += v
                else:
                    pass
        units = sum(count.values())
        out["areas"][levels.get(lid, str(lid))] = {
            "lid": lid, "seeds": ns, "units_per_clear": round(units, 2),
            "kinds_per_clear": {k: round(v, 2) for k, v in sorted(count.items())},
            "per_clear": {k: round(v, 5) for k, v in sorted(tot.items())},
            "per_clear_by_kind": {k: {kk: round(v, 5) for kk, v in sorted(d.items())} for k, d in sorted(per_kind.items())},
        }
    for (name, kind), ma in sorted(mon_acc.items()):
        n = ma["n"]
        out["monsters"][f"{name}:{kind}"] = {"tc": mon_meta[(name, kind)], "observed": int(n), "ilvl_mean": round(ma["ilvl"] / n, 2),
                                             "per_kill": {k: round(v / n, 5) for k, v in sorted(ma.items()) if k not in ("n", "ilvl")}}
    # bosses: one entry per superunique / named boss seen in the population, at its observed ilvl
    seen = {}
    for lid in area:
        for seed, us in area[lid].items():
            for cls, flags, il in us:
                r = ms[str(cls)]
                mid = r["Id"].lower()
                if flags & 2 or r.get("boss") == "1" or mid in ("bloodraven",):
                    tc, kind, name = unit_tc(cls, flags)
                    seen.setdefault((name, levels.get(lid, str(lid))), {})[il] = (tc, cls, flags)
    pretty = {"andariel": "Andariel", "bloodraven": "Blood Raven"}
    for (name, where), d in sorted(seen.items()):
        name = pretty.get(name, name)
        il = max(d)
        tc, cls, flags = d[il]
        vec = roll_key(tc, il)
        ent = {"area": where, "class": ms[str(cls)]["Id"], "name": name, "tc": tc, "ilvl": il, "flags": hex(flags),
               "per_kill": {k: round(v, 5) for k, v in sorted(vec.items()) if not k.startswith("_")},
               "drops": round(vec["_drops"], 3), "nodrop": round(vec["_nodrop"], 3), "gold_piles": round(vec["_piles"], 3)}
        r = ms[str(cls)]
        if r.get("TCQuestId") not in ("", "0", None) and r.get("TreasureClass4"):
            v4 = roll_key(r["TreasureClass4"], il)
            ent["first_kill"] = {"tc": r["TreasureClass4"], "per_kill": {k: round(v, 5) for k, v in sorted(v4.items()) if not k.startswith("_")},
                                 "drops": round(v4["_drops"], 3), "gold_piles": round(v4["_piles"], 3)}
        out["bosses"][name] = ent
    out["chests"] = {}
    for cname in sorted(o.index):
        if re.fullmatch(r"Act 1 Chest [ABC]", cname):
            out["chests"][cname] = {str(il): {k: round(v, 5) for k, v in sorted(roll_key(cname, il).items()) if not k.startswith("_")} for il in (1, 3, 5, 8, 11)}
    out["unknown_superunique_class"] = dict(unknown)
    for (lid, nm), v in notc.items():
        out["areas"][levels.get(lid, str(lid))].setdefault("units_without_treasure_class", {})[nm] = round(v, 3)
    zs = [abs(z) for *_, z in zrows]
    out["check"] = {"class_ilvl_keys": len(cache), "cells": len(zrows), "max_abs_z": round(max(zs), 2),
                    "cells_over_3p5": sum(1 for z in zs if z > 3.5)}
    out["check"]["worst"] = [list(map(lambda x: round(x, 4) if isinstance(x, float) else x, r)) for r in sorted(zrows, key=lambda r: -abs(r[5]))[:6]]
    json.dump(out, open(OUT, "w"), indent=1, sort_keys=True)
    print("keys", len(cache), "max |z|", out["check"]["max_abs_z"], ">3.5:", out["check"]["cells_over_3p5"], "of", len(zrows))


if __name__ == "__main__":
    main()
