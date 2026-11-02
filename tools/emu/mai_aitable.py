#!/usr/bin/env python3
"""monsterai: the per-AI table of docs/research/monster-ai-trace.md (one row per distinct monstats AI, 146).
usage: mai_aitable.py <tables dir (monstats.txt monai.txt levels.txt, from mpq-cat)> <MonAI dump from mai_table.py> <traced names file> <upstream inc dir> <mine: stderr/stdout counts file 'mai AI cases differ'>
Tables are Blizzard data and are not committed; the output is the markdown rows."""
import csv, glob, re, sys, collections

tdir, tbl, traced_f, incdir, mine_f = sys.argv[1:6]
def rd(n): return list(csv.reader(open(f"{tdir}/{n}.txt", encoding="latin1"), delimiter="\t"))
ms = rd("monstats"); h = [c.lower() for c in ms[0]]; ix = {c: i for i, c in enumerate(h)}
lv = rd("levels"); lh = lv[0]
monai = {r[0]: i for i, r in enumerate(rd("monai")[1:])}
table = {}
for l in open(tbl):
    p = l.split()
    if len(p) >= 6 and p[0].isdigit(): table[p[1]] = p[2:]
traced = set(open(traced_f).read().split())
names = [re.match(r'\{ "([^"]+)", (\w+)', l).groups() for l in open(f"{incdir}/mon_ais_names.inc") if l.startswith("{")]
corpus = collections.Counter()
for f in glob.glob(f"{incdir}/mon_ais*_cases.inc"):
    for l in open(f):
        m = re.match(r'"(\d+) ', l)
        if m and not names[int(m.group(1))][1] == "true": corpus[names[int(m.group(1))][0]] += 1
mine = {}
for l in open(mine_f):
    p = l.split()
    if len(p) == 4 and p[0] == "mai" and p[1] != "TOTAL": mine[p[1]] = (int(p[2]), int(p[3]))
# monsters placed per act: levels mon1..mon10 / nmon / umon columns, Act column 0-based (0 = Act 1)
cols = [i for i, c in enumerate(lh) if re.fullmatch(r"(n|u)?mon(\d+)", c.lower())]
place = collections.defaultdict(set)
for r in lv[1:]:
    if len(r) <= max(cols): continue
    try: act = int(r[lh.index("Act")]) + 1
    except ValueError: continue
    for i in cols:
        if r[i].strip(): place[r[i].strip().lower()].add(act)
byai = collections.defaultdict(list)
for r in ms[1:]:
    if len(r) <= ix["ai"] or not r[ix["id"]] or r[ix["id"]] == "Expansion": continue
    byai[r[ix["ai"]]].append(r)
out = []
for ai in sorted(byai):
    rows = byai[ai]
    a12 = sorted({r[ix["id"]] for r in rows if place.get(r[ix["id"]].lower(), set()) & {1, 2}})
    t = table.get(ai, ["?", "?", "?", "?"])
    port = "yes" if ai in traced else "no"
    cs, df = mine.get(ai, (None, None))
    oracle = corpus.get(ai, 0)
    r0 = rows[0]
    aidel = "/".join(r0[ix[k]] for k in ("aidel", "aidel(n)", "aidel(h)") if k in ix)
    aip = " ".join(r0[ix[f"*aip{k}"]] if f"*aip{k}" in ix else r0[ix.get(f"aip{k}", 0)] for k in range(1, 6))
    out.append(f"| {ai} | {len(rows)} | {', '.join(a12[:6]) + ('...' if len(a12) > 6 else '') if a12 else '-'} | {t[0]} | {t[2]} | {aidel} | {port} | {oracle or '-'} | {('%d / %d' % (cs, df)) if cs else '-'} |")
print("| AI | monstats rows | Act 1-2 placed (Id) | search type | think (+8) | aidel N/NM/H | d2d mon_think port | upstream corpus cases | this run cases / differ |")
print("|---|---|---|---|---|---|---|---|---|")
print("\n".join(out))
print(f"\n{len(byai)} distinct AIs, {sum(len(v) for v in byai.values())} monstats rows", file=sys.stderr)
