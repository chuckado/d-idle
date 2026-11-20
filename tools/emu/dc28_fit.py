#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp28: units-per-clear fit of the danger generator against game.exe's room population (stdlib only, no emulator needed).

Input  docs/research/data28/areas-input.csv (one row per Normal area: Levels.txt DrlgType / MonDen, mean pack size of the level's species,
       the mean tile count of the rooms game.exe's level generator built over 12 seeds, units per clear from game.exe's room population).
    python3 tools/emu/dc28_fit.py report            # the tables of docs/survival/units-fit-2026-11-15.md
    python3 tools/emu/dc28_fit.py json > internal/danger/areaunits.json
Model (V run inputs, I model): units = tiles x exp(a + b ln MonDen)   for mazes (DrlgType 1)
                               units = tiles x exp(a)                for presets (2) and outdoors (3)
Checks: in-sample, leave-one-area-out, leave-one-family-out (Tal Rasha's seven tombs, the Catacombs, ... held out together), and the
fit on Act 1 alone against Act 2.  The error is |model - game| / game; the median is the upper median (as TestAreamonDiff)."""
import csv, json, math, os, re, sys

HERE = os.path.dirname(os.path.abspath(__file__))
INP = os.path.join(HERE, "..", "..", "docs", "research", "data28", "areas-input.csv")
SHIPPED = {1: 0.0453, 2: 0.0272, 3: 0.1735}          # Params.UnitsPerDensityGroup of internal/danger/params.go
EXCLUDED = {45: "game.exe's level build is a 16-room stub there and the population is 1 or 2 units (7 seeds): not a population (U)",
            73: "Duriel's lair holds the boss only (1 unit in 12 seeds)"}
SCREEN_TILES = 280.0


def load():
    rows = []
    for r in csv.DictReader(open(INP)):
        d = {k: (float(v) if re.fullmatch(r"-?\d+(\.\d+)?", v) else v) for k, v in r.items()}
        for k in ("lid", "act", "drlg", "den", "seeds", "units_min", "units_max", "levels_size"):
            d[k] = int(d[k])
        rows.append(d)
    return rows


def ols(X, y):
    n = len(X[0])
    A = [[sum(r[i] * r[j] for r in X) for j in range(n)] for i in range(n)]
    b = [sum(r[i] * v for r, v in zip(X, y)) for i in range(n)]
    for i in range(n):
        p = max(range(i, n), key=lambda k: abs(A[k][i]))
        A[i], A[p], b[i], b[p] = A[p], A[i], b[p], b[i]
        for k in range(i + 1, n):
            f = A[k][i] / A[i][i]
            for j in range(i, n):
                A[k][j] -= f * A[i][j]
            b[k] -= f * b[i]
    w = [0.0] * n
    for i in range(n - 1, -1, -1):
        w[i] = (b[i] - sum(A[i][j] * w[j] for j in range(i + 1, n))) / A[i][i]
    return w


def feats(d):
    return [1.0, math.log(d["den"])] if d["drlg"] == 1 else [1.0]


def fit(rows):
    out = {}
    for t in (1, 2, 3):
        tr = [d for d in rows if d["drlg"] == t]
        if tr:
            out[t] = ols([feats(d) for d in tr], [math.log(d["units_mean"] / d["drlg_tiles"]) for d in tr])
    return out


def predict(W, d):
    return math.exp(sum(a * b for a, b in zip(W[d["drlg"]], feats(d)))) * d["drlg_tiles"]


def upper_median(v):
    v = sorted(v)
    return v[len(v) // 2]


def family(d):
    n = re.sub(r"\s*(Level\s*\d.*|\d+.*)$", "", d["area"]).strip()
    return (d["act"], n, d["drlg"])


def cv(rows, key):
    err = {}
    for k in {key(d) for d in rows}:
        W = fit([d for d in rows if key(d) != k])
        for d in rows:
            if key(d) == k:
                err[d["lid"]] = abs(predict(W, d) - d["units_mean"]) / d["units_mean"]
    return err


def usable(rows):
    return [d for d in rows if d["lid"] not in EXCLUDED]


def gen_units(d):
    return d["den"] * d["group_mean"] * SHIPPED[d["drlg"]]


def summary(name, err, rows):
    v = list(err.values())
    s = "%-26s n=%2d median %3.0f%% max %4.0f%%" % (name, len(v), upper_median(v) * 100, max(v) * 100)
    for t, tn in ((1, "maze"), (2, "preset"), (3, "outdoor")):
        vv = [err[d["lid"]] for d in rows if d["drlg"] == t]
        if vv:
            s += " | %s n=%d %3.0f%%" % (tn, len(vv), upper_median(vv) * 100)
    return s


def report():
    rows = usable(load())
    W = fit(rows)
    print("fit on all %d areas: maze ln(units/tile) = %.4f + %.4f ln(den); preset %.4f; outdoor %.4f" % (len(rows), W[1][0], W[1][1], W[2][0], W[3][0]))
    ins = {d["lid"]: abs(predict(W, d) - d["units_mean"]) / d["units_mean"] for d in rows}
    shipped = {d["lid"]: abs(gen_units(d) - d["units_mean"]) / d["units_mean"] for d in rows}
    print(summary("shipped generator", shipped, rows))
    print(summary("tiles model in-sample", ins, rows))
    print(summary("leave-one-area-out", cv(rows, lambda d: d["lid"]), rows))
    print(summary("leave-one-family-out", cv(rows, family), rows))
    a1 = [d for d in rows if d["act"] == 1]
    a2 = [d for d in rows if d["act"] == 2]
    W1 = fit(a1)
    e = {d["lid"]: abs(predict(W1, d) - d["units_mean"]) / d["units_mean"] for d in a2}
    print(summary("fit Act 1 -> Act 2", e, a2))
    print(summary("shipped on Act 2", {d["lid"]: shipped[d["lid"]] for d in a2}, a2))
    print()
    print("| lid | area | type | MonDen | tiles | game units | shipped | shipped err | tiles model (LOO) | LOO err | per screen game / shipped |")
    print("|---|---|---|---|---|---|---|---|---|---|---|")
    loo = {}
    for d in rows:
        Wd = fit([x for x in rows if x["lid"] != d["lid"]])
        loo[d["lid"]] = predict(Wd, d)
    for d in rows:
        g = gen_units(d)
        ps_game = d["units_mean"] / d["drlg_tiles"] * SCREEN_TILES
        ps_ship = g / d["levels_size"] * SCREEN_TILES      # Zone.Density.PerScreen: units / (Levels.txt SizeX x SizeY) x ScreenTiles
        print("| %d | %s | %d | %d | %d | %.0f | %.0f | %+.0f%% | %.0f | %+.0f%% | %.1f / %.1f |" % (
            d["lid"], d["area"], d["drlg"], d["den"], d["drlg_tiles"], d["units_mean"], g, (g - d["units_mean"]) / d["units_mean"] * 100,
            loo[d["lid"]], (loo[d["lid"]] - d["units_mean"]) / d["units_mean"] * 100, ps_game, ps_ship))


def emit_json():
    allrows = load()
    rows = usable(allrows)
    W = fit(rows)
    doc = {"schema": "danger.areaunits/1", "diff": "normal", "tag": "V run (game.exe room population under the emulator, 12 seeds) + I (the fit)",
           "evidence": "docs/research/decomp28-trace.md, docs/survival/units-fit-2026-11-15.md, docs/research/data28/areas-input.csv",
           "note": "Not read by the generator yet (a wiring card is proposed in the trace). Prefer rows[].unitsMean for a measured area; use fit.* only for an unmeasured area (other acts, other difficulties: not measured).",
           "fit": {"maze": {"lnUnitsPerTile": round(W[1][0], 4), "lnDenCoef": round(W[1][1], 4)}, "preset": {"lnUnitsPerTile": round(W[2][0], 4)},
                   "outdoor": {"lnUnitsPerTile": round(W[3][0], 4)}, "formula": "units = tiles * exp(a [+ b * ln MonDen])", "screenTiles": SCREEN_TILES},
           "rows": [], "excluded": [{"area": d["lid"], "why": EXCLUDED[d["lid"]]} for d in allrows if d["lid"] in EXCLUDED]}
    for d in rows:
        doc["rows"].append({"area": d["lid"], "name": d["area"], "act": d["act"], "drlgType": d["drlg"], "monDen": d["den"], "tiles": int(d["drlg_tiles"]),
                            "unitsMean": d["units_mean"], "unitsMin": d["units_min"], "unitsMax": d["units_max"], "seeds": d["seeds"],
                            "perScreen": round(d["units_mean"] / d["drlg_tiles"] * SCREEN_TILES, 2)})
    print(json.dumps(doc, indent=1))


if __name__ == "__main__":
    {"report": report, "json": emit_json}.get(sys.argv[1] if len(sys.argv) > 1 else "", lambda: sys.exit(__doc__))()
