#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp31: dm_model.py output (dm.json) -> internal/plandata/dropmodel-normal.json (schema dropmodel/1).

  python3 tools/emu/dm_pack.py DM.json internal/plandata/areaxp.csv OUT.json
Derives per-1000-kill and per-minute summary rows (kills per minute = units_per_clear / clear_min of areaxp.csv, time model I).
"""
import csv
import json
import sys

dm = json.load(open(sys.argv[1]))
clear = {}
for r in csv.reader(l for l in open(sys.argv[2]) if not l.startswith("#")):
    if r and r[0].isdigit():
        clear[int(r[0])] = float(r[3])

HP = [f"potion_hp{i}" for i in range(1, 6)]
MP = [f"potion_mp{i}" for i in range(1, 6)]
Q = ("plain", "superior", "magic", "rare", "set", "unique")


def summary(v):
    g = lambda k: v.get(k, 0.0)
    s = {}
    for i in range(1, 6):
        if g(f"potion_hp{i}"): s[f"potion_hp{i}"] = g(f"potion_hp{i}")
        if g(f"potion_mp{i}"): s[f"potion_mp{i}"] = g(f"potion_mp{i}")
    s["potion_hp"], s["potion_mp"], s["potion_rejuv"] = g("potion_hp"), g("potion_mp"), g("potion_rejuv")
    s["potion_other"] = g("potion_other")
    s["scroll_town_portal"], s["scroll_identify"] = g("scroll_town_portal"), g("scroll_identify")
    s["tome"] = g("tome_town_portal") + g("tome_identify")
    s["key"] = g("key")
    for gr in ("chipped", "flawed", "standard", "flawless", "perfect"):
        if g(f"gem_{gr}"): s[f"gem_{gr}"] = g(f"gem_{gr}")
    s["gem"], s["rune"] = g("gem"), g("rune")
    for k in v:
        if k.startswith("rune_r"): s[k] = v[k]
    s["rune_low"] = sum(g(f"rune_r{i:02d}") for i in range(1, 11))
    s["rune_high"] = s["rune"] - s["rune_low"]
    # gear (weapons + armor), jewelry, jewels, charms: magic or better
    mag = lambda p: g(f"{p}_magic")
    s["gear_magic"], s["gear_rare"] = g("gear_magic"), g("gear_rare")
    s["gear_set"], s["gear_unique"] = g("gear_set"), g("gear_unique")
    s["gear_superior"], s["gear_plain"] = g("gear_superior"), g("gear_plain")
    s["jewelry_magic"], s["jewelry_rare"] = g("jewelry_magic"), g("jewelry_rare")
    s["jewelry_set"], s["jewelry_unique"] = g("jewelry_set"), g("jewelry_unique")
    s["jewel_any"] = sum(g(f"jewel_{q}") for q in ("normal", "magic", "rare", "unique"))
    s["charm"] = sum(g(k) for k in v if k.startswith("charm_") and k.count("_") == 1)
    s["magic_plus"] = sum(g(k) for k in ("gear_magic", "gear_rare", "gear_set", "gear_unique", "jewelry_magic", "jewelry_rare", "jewelry_set",
                                          "jewelry_unique", "jewel_magic", "jewel_rare", "jewel_unique", "charm_magic", "charm_rare", "charm_unique", "charm_set"))
    s["rare_plus"] = sum(g(k) for k in ("gear_rare", "gear_set", "gear_unique", "jewelry_rare", "jewelry_set", "jewelry_unique", "jewel_rare", "jewel_unique"))
    s["set_unique"] = sum(g(k) for k in ("gear_set", "gear_unique", "jewelry_set", "jewelry_unique", "jewel_unique", "charm_unique", "charm_set"))
    for b in ("early", "mid", "late"):
        for q in Q:
            if g(f"gear_{b}_{q}"): s[f"gear_{b}_{q}"] = g(f"gear_{b}_{q}")
    s["gear_usable25_magic_plus"] = sum(g(f"gear_{b}_{q}") for b in ("early", "mid") for q in ("magic", "rare", "set", "unique"))
    return {k: v for k, v in s.items() if v}


CORE = ("potion_hp", "potion_mp", "potion_rejuv", "potion_other", "scroll_town_portal", "scroll_identify", "tome", "key", "gem", "rune", "rune_low",
        "rune_high", "magic_plus", "rare_plus", "set_unique", "gear_usable25_magic_plus", "gear_plain", "gear_superior", "gear_magic", "gear_rare", "gear_set",
        "gear_unique") + tuple(f"potion_hp{i}" for i in range(1, 6)) + tuple(f"potion_mp{i}" for i in range(1, 6))


def core(d):
    return {k: v for k, v in d.items() if k in CORE}


def rd(d, nd=4):
    return {k: round(v, nd) for k, v in sorted(d.items())}


out = {"schema": "dropmodel/1", "version": 1, "difficulty": "normal", "doc": "docs/research/decomp31-drops.md",
       "conditions": "one player, magic find 0, gold find 0; item level = the unit's level (stat 12); rolls are game.exe's FUN_0055a6d0 natively under the emulator (V run)",
       "units": "every value is an expected COUNT of dropped items: per_clear = whole area once; per_1000_kills = per 1000 kills; per_min = per_clear / clear_min (clear_min is the TIME MODEL of areaxp.csv, I). Gold is excluded (docs/research/xp-gold-trace.md).",
       "keys": {
           "potion_hpN / potion_mpN": "healing / mana potion of tier N (1 minor .. 5 super); potion_hp and potion_mp are the sums; potion_rejuv rejuvenation (small and full); potion_other stamina, antidote, thawing",
           "scroll_town_portal / scroll_identify, tome, key": "scrolls, book items and keys",
           "gem_<grade>, gem, rune, rune_low, rune_high": "gems by grade, all runes, runes r01..r10 and r11+",
           "gear_<quality>": "weapons and armor by quality: plain (low and normal), superior, magic, rare, set, unique",
           "gear_<early|mid|late>_<quality>": "the same by base levelreq: early <= 10, mid 11..25, late > 25 (strength and dexterity not modelled, U)",
           "jewelry_*, jewel_any, charm": "rings and amulets by quality, jewels, charms",
           "magic_plus, rare_plus, set_unique": "all items (gear, jewelry, jewels, charms) of magic or better, rare or better, set or unique",
           "gear_usable25_magic_plus": "weapons and armor of magic or better quality on early or mid bases (a clvl 15..25 hero)"},
       "check": dm["check"], "areas": {}, "monsters": {}, "bosses": {}, "chests": {}}
for name, a in sorted(dm["areas"].items(), key=lambda x: x[1]["lid"]):
    lid = a["lid"]
    units = a["units_per_clear"]
    pc = summary(a["per_clear"])
    cm = clear.get(lid)
    ent = {"lid": lid, "seeds": a["seeds"], "units_per_clear": units, "kinds_per_clear": a["kinds_per_clear"], "clear_min": cm,
           "kills_per_min": round(units / cm, 2) if cm else None,
           "per_clear": rd(pc, 3), "per_1000_kills": rd({k: v / units * 1000 for k, v in pc.items()}, 2)}
    if cm:
        ent["per_min"] = rd({k: v / cm for k, v in pc.items()}, 4)
    ent["per_clear_by_kind"] = {k: rd(core(summary(v)), 3) for k, v in a["per_clear_by_kind"].items()}
    if a.get("units_without_treasure_class"):
        ent["units_without_treasure_class"] = a["units_without_treasure_class"]
    out["areas"][name] = ent
for k, m in sorted(dm["monsters"].items()):
    out["monsters"][k] = {"tc": m["tc"], "observed": m["observed"], "ilvl_mean": m["ilvl_mean"], "per_1000_kills": rd({a: b * 1000 for a, b in core(summary(m["per_kill"])).items()}, 2)}
for k, b in sorted(dm["bosses"].items()):
    e = {"area": b["area"], "class": b["class"], "tc": b["tc"], "ilvl": b["ilvl"], "drops": b["drops"], "per_kill": rd(summary(b["per_kill"]), 4)}
    if "first_kill" in b:
        e["first_kill"] = {"tc": b["first_kill"]["tc"], "drops": b["first_kill"]["drops"], "per_kill": rd(summary(b["first_kill"]["per_kill"]), 4)}
    out["bosses"][k] = e
for k, d in dm["chests"].items():
    out["chests"][k] = {il: rd(summary(v), 4) for il, v in d.items()}
json.dump(out, open(sys.argv[3], "w"), indent=1)
