#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp29: build internal/power/skilldamage.json and the doc tables from the worker-host runs.

    python3 tools/emu/dc29_build.py SCRATCH_DIR OUT.json [DOC_TABLES.md]

SCRATCH_DIR holds: skills.txt, missiles.txt (tools/emu/mpq-table.sh skills|missiles), all.jsonl + low.jsonl + mana.jsonl
(dc29_skill.py all|mana), rate.tsv (dc29_rate.py), animdata.d2 (d9_animdata.sh). Pure arithmetic and reshaping: no game data
is invented; every number is a game.exe oracle value or a patched-table cell (docs/research/decomp29-skilldamage.md)."""
import csv
import json
import os
import struct
import sys

d = sys.argv[1]
out_json = sys.argv[2]
out_doc = sys.argv[3] if len(sys.argv) > 3 else None


def table(name):
    rows = list(csv.reader(open(os.path.join(d, name), encoding="latin-1"), delimiter="\t"))
    h = rows[0]
    return h, [r for r in rows[1:] if len(r) > 10]


sh, srows = table("skills.txt")
si = {n: i for i, n in enumerate(sh)}
mh, mrows = table("missiles.txt")
mi = {n: i for i, n in enumerate(mh)}
missiles = {r[0]: r for r in mrows if r[0]}

oracle = {}
for f in ("all.jsonl", "low.jsonl"):
    for line in open(os.path.join(d, f)):
        j = json.loads(line)
        oracle.setdefault(j["id"], {})[j["lvl"]] = j
mana = {}
for line in open(os.path.join(d, "mana.jsonl")):
    j = json.loads(line)
    mana.setdefault(j["id"], {})[j["lvl"]] = j

# AnimData frames per (class code, mode, weapon class)
raw = open(os.path.join(d, "animdata.d2"), "rb").read()
p = 0
anim = {}
for _ in range(256):
    n = struct.unpack_from("<I", raw, p)[0]
    p += 4
    for _ in range(n):
        name = raw[p:p + 8].split(b"\0")[0].decode()
        fr, sp = struct.unpack_from("<II", raw, p + 8)
        p += 160
        anim[name] = (fr, sp)
assert p == len(raw)
CLASSCODE = {"ama": "AM", "sor": "SO", "nec": "NE", "pal": "PA", "bar": "BA", "dru": "DZ", "ass": "AI"}
MODES = ["A1", "A2", "SC", "TH", "KK", "SQ", "S1", "S2", "S3", "S4"]
frames = {}
for cls, code in CLASSCODE.items():
    for k, (fr, sp) in anim.items():
        if len(k) == 7 and k[:2] == code and k[2:4] in MODES:
            frames.setdefault(cls, {}).setdefault(k[2:4], {})[k[4:]] = {"frames": fr, "speed": sp}

rate = {"attack_wsm0": [0] * 201, "attack_by_wsm": {}, "cast": [0] * 201}
for line in open(os.path.join(d, "rate.tsv")):
    f = line.split()
    if not f:
        continue
    if f[0] == "A":
        ias, wsm, r = int(f[1]), int(f[2]), int(f[3])
        rate["attack_by_wsm"].setdefault(str(wsm), [0] * 201)[ias] = r
    elif f[0] == "C":
        rate["cast"][int(f[1])] = int(f[2])
rate["attack_wsm0"] = rate["attack_by_wsm"].pop("0")

profiles = json.load(open("internal/power/profiles.json")) if os.path.exists("internal/power/profiles.json") else {}
kinds = profiles.get("skillKinds", {})


def g(r, n):
    return r[si[n]].strip() if n in si and si[n] < len(r) else ""


def num(s):
    try:
        return int(s)
    except ValueError:
        return None


EXTRA = {"Smite", "Static Field", "Corpse Explosion", "Dragon Talon", "Dragon Tail", "Kick"}
skills = {}
for r in srows:
    sid = num(r[si["Id"]])
    if sid is None or sid not in oracle:
        continue
    o = oracle[sid]
    lv = sorted(o)

    def ser(tag, key):
        return [o[k][tag][key] for k in lv]
    emax = max(max(o[k][t]["e"][1] for t in ("none", "all20")) for k in lv)
    pmax = max(max(o[k][t]["p"][1] for t in ("none", "all20")) for k in lv)
    srcdam = num(g(r, "SrcDam"))
    anim_mode = g(r, "anim")
    name = g(r, "skill")
    # damage-free rows kept on purpose: Smite (shield damage), Static Field (percent of life), Corpse Explosion, Dragon Talon / Tail
    if not (emax or pmax or srcdam) and name not in EXTRA:
        continue
    seqtrans = g(r, "seqtrans")
    if anim_mode == "SC":
        rk = "cast"
    elif anim_mode == "SQ":
        rk = "cast" if seqtrans == "SC" else "attack_sq"
    else:
        rk = "attack"
    ms = {}
    for k in lv:
        if sid in mana and k in mana[sid]:
            ms[k] = mana[sid][k]
    ent = {
        "id": sid, "class": g(r, "charclass") or None, "anim": anim_mode, "seqtrans": seqtrans or None, "rate_kind": rk,
        "kind": kinds.get(name),
        "srcdam": srcdam, "uses_attack_rate": g(r, "UseAttackRate") == "1",
        "tohit_pct": num(g(r, "ToHit")), "tohit_per_level": num(g(r, "LevToHit")),
        "etype": g(r, "EType") or None, "hitshift": num(g(r, "HitShift")),
        "req_level": num(g(r, "reqlevel")), "delay_frames": num(g(r, "delay")),
        "mana_cols": {"mana": num(g(r, "mana")), "lvlmana": num(g(r, "lvlmana")), "minmana": num(g(r, "minmana")),
                      "manashift": num(g(r, "manashift")), "startmana": num(g(r, "startmana"))},
        "mana": [round(ms[k]["paid"] / 256, 4) if k in ms else None for k in lv],
        "calc_src": [g(r, "calc%d" % i) or None for i in range(1, 5)],
        "calc_desc": [g(r, "*calc%d desc" % i) or None for i in range(1, 5)],
        "params": {"Param%d" % i: g(r, "Param%d" % i) for i in range(1, 9) if g(r, "Param%d" % i)},
        "elem": {t: {"min": [v[0] for v in ser(t, "e")], "max": [v[1] for v in ser(t, "e")], "len": [v[2] for v in ser(t, "e")]}
                 for t in ("none", "all1", "all20")},
        "phys": {t: {"min": [v[0] for v in ser(t, "p")], "max": [v[1] for v in ser(t, "p")]} for t in ("none", "all20")},
        "calc": {t: [list(v) for v in ser(t, "c")] for t in ("none", "all1", "all20")},
        "levels": lv,
    }
    mn = g(r, "srvmissile")
    if mn and mn in missiles:
        mr = missiles[mn]
        ent["missile"] = {"name": mn, "pierce": mr[mi["Pierce"]], "vel": mr[mi["Vel"]], "range": mr[mi["Range"]],
                          "explosion": mr[mi["Explosion"]]}
    skills[name] = ent

doc = {
    "schema": "power.skilldamage/1",
    "version": 1,
    "tags": "V run unless a field says otherwise: values are game.exe 1.14d's own functions run in the emulator (docs/research/decomp29-skilldamage.md)",
    "units": {"damage": "256ths of a hit point (divide by 256)", "mana": "points", "length": "ticks (elem.len)", "tick_seconds": 0.04,
              "synergy_sets": "none = the skill alone at its level; all1 / all20 = every other skill at 1 / 20 points (only skills the calc names count)"},
    "assumptions": {"clvl_for_ulvl_terms": 30, "masteries": 0, "levels": "skill level with items = base level"},
    "formulas": {
        "elem_dps_one_target": "avg(elem.min, elem.max)/256 * (1 - resist) / (cast_ticks * tick_seconds)",
        "cast_ticks_plain": "ceil(256 * frames / rate_cast[fcr]) - 1",
        "cast_ticks_sequence": "ceil(256 * frames / rate_cast[fcr])",
        "attack_ticks": "ceil(256 * frames / rate_attack[ias])",
        "melee_hit": "weapon_avg * (srcdam / 128) * (100 + ED + stat bonuses) / 100 * hit_chance; ED from calc (see calc_desc), phys from phys.*",
        "mana_per_second": "mana[level] / (ticks * tick_seconds)",
    },
    "rate": rate,
    "frames": frames,
    "skills": skills,
}
json.dump(doc, open(out_json, "w"), separators=(",", ":"), sort_keys=True)
print("skills", len(skills), "bytes", os.path.getsize(out_json))
