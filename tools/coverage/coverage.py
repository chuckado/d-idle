#!/usr/bin/env python3
"""Generate docs/coverage/rule-coverage.md: how much of game.exe 1.14d's rule tables d2d has been compared against.

  coverage.py                      print the table (dry run, writes nothing)
  coverage.py --write              write docs/coverage/rule-coverage.md
  coverage.py --check              exit 1 if the inputs are invalid or the output file is stale (CI)
  coverage.py --tables DIR         also check docs/coverage/denominators.tsv against a local data cache (DIR holds
                                   skills.tsv, monstats.tsv, ...; the bot's extracted cache). Exit 1 on a mismatch.
  coverage.py --tables DIR --write-denominators
                                   rewrite denominators.tsv from the cache (the hand-listed `rules` rows are kept)

Inputs (all under docs/coverage/): denominators.tsv (fixed row universes), ledger.tsv (hand-maintained status per row
range), gaps.tsv (hand-ranked top gaps). Stdlib only; reads no game data unless --tables is given.
Statuses: TESTED-0-DIFF, TESTED-DEVIATION (a V run exists), READ-ONLY (V read, no run), BOT-ONLY, UNTESTED (default).
"""
import argparse
import csv
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
DIR = "docs/coverage"
OUT = DIR + "/rule-coverage.md"
STATUSES = ("TESTED-0-DIFF", "TESTED-DEVIATION", "READ-ONLY", "BOT-ONLY", "UNTESTED")
RUN = ("TESTED-0-DIFF", "TESTED-DEVIATION")
KINDS = ("ints", "names", "count", "rules")
CLASSES = (("ama", "Amazon"), ("sor", "Sorceress"), ("nec", "Necromancer"), ("pal", "Paladin"),
           ("bar", "Barbarian"), ("dru", "Druid"), ("ass", "Assassin"))
DATE_RE = re.compile(r"^\d{4}-\d{2}-\d{2}$")
SEED_RE = re.compile(r"^seeds? \d+")
ASPECT = lambda name: name.startswith("skillfx.") or name == "monster.flags"  # single-aspect checks
DID_RE = re.compile(r"\b[A-Z]{1,2}\d+\b")


class CovError(Exception):
    pass


def read_tsv(path, cols, root=ROOT):
    """Rows of a tab separated file as dicts; '#' lines and blanks skipped; the header line is checked."""
    full = os.path.join(root, path)
    if not os.path.exists(full):
        raise CovError("missing %s" % path)
    rows, header = [], None
    with open(full, encoding="utf-8") as f:
        for n, ln in enumerate(f, 1):
            ln = ln.rstrip("\n")
            if not ln.strip() or ln.startswith("#"):
                continue
            parts = ln.split("\t")
            if header is None:
                header = parts
                if tuple(header[:len(cols)]) != tuple(cols):
                    raise CovError("%s: header must start with %s" % (path, ", ".join(cols)))
                continue
            if len(parts) < len(cols):
                parts += [""] * (len(cols) - len(parts))
            d = dict(zip(header, parts))
            d["_line"] = n
            d["_file"] = path
            rows.append(d)
    if header is None:
        raise CovError("%s: no header" % path)
    return rows


def expand_ints(spec, where):
    out = set()
    for tok in spec.split(","):
        tok = tok.strip()
        if not tok:
            continue
        m = re.fullmatch(r"(\d+)(?:-(\d+))?", tok)
        if not m:
            raise CovError("%s: bad id token %r" % (where, tok))
        a = int(m.group(1))
        b = int(m.group(2) or a)
        if b < a:
            raise CovError("%s: descending range %r" % (where, tok))
        out.update(range(a, b + 1))
    return out


def compact_ints(ids):
    ids = sorted(ids)
    parts, i = [], 0
    while i < len(ids):
        j = i
        while j + 1 < len(ids) and ids[j + 1] == ids[j] + 1:
            j += 1
        parts.append(str(ids[i]) if i == j else "%d-%d" % (ids[i], ids[j]))
        i = j + 1
    return ",".join(parts)


class Universe:
    def __init__(self, row):
        self.name = row["subsystem"]
        self.label = row["label"]
        self.kind = row["kind"]
        self.source = row["source"]
        where = "denominators.tsv:%d (%s)" % (row["_line"], self.name)
        if self.kind not in KINDS:
            raise CovError("%s: kind must be one of %s" % (where, ", ".join(KINDS)))
        try:
            self.total = int(row["total"])
        except ValueError:
            raise CovError("%s: total is not a number" % where)
        if self.kind == "ints":
            self.keys = expand_ints(row["keys"], where)
        elif self.kind in ("names", "rules"):
            self.keys = {k.strip() for k in row["keys"].split(",") if k.strip()}
        else:
            self.keys = None
        if self.keys is not None and len(self.keys) != self.total:
            raise CovError("%s: total %d but %d keys listed" % (where, self.total, len(self.keys)))
        if self.total <= 0:
            raise CovError("%s: total must be positive" % where)


def load_universes(root=ROOT):
    rows = read_tsv(DIR + "/denominators.tsv", ("subsystem", "label", "kind", "total", "keys", "source"), root)
    uni = {}
    for r in rows:
        u = Universe(r)
        if u.name in uni:
            raise CovError("denominators.tsv: duplicate subsystem %s" % u.name)
        uni[u.name] = u
    return uni


def resolve_tokens(u, spec, where):
    """Return (set of keys, count). `all` is the whole universe; count universes take `all` or `#N`."""
    spec = spec.strip()
    if spec == "all":
        return (set(u.keys) if u.keys is not None else None), u.total
    if u.kind == "count":
        m = re.fullmatch(r"#(\d+)", spec)
        if not m or int(m.group(1)) < 1:
            raise CovError("%s: %s is count-only; rows must be `all` or `#N`" % (where, u.name))
        return None, int(m.group(1))
    keys = expand_ints(spec, where) if u.kind == "ints" else {k.strip() for k in spec.split(",") if k.strip()}
    bad = sorted(keys - u.keys, key=str)
    if bad:
        raise CovError("%s: not rows of %s: %s" % (where, u.name, ", ".join(map(str, bad[:5]))))
    return keys, len(keys)


def note_doc(note, root):
    return any(os.path.isfile(os.path.join(root, m)) for m in re.findall(r"docs/[\w./-]+\.md", note))


def doc_has_number(text, n):
    return str(n) in text or "{:,}".format(n) in text


def load_ledger(uni, root=ROOT):
    rows = read_tsv(DIR + "/ledger.tsv", ("subsystem", "rows", "status", "evidence", "cases", "seed_or_date", "note"), root)
    per = {name: [] for name in uni}
    claimed = {name: set() for name in uni}
    counted = {name: 0 for name in uni}
    doc_cache = {}
    for r in rows:
        where = "ledger.tsv:%d" % r["_line"]
        name, status = r["subsystem"], r["status"]
        if name not in uni:
            raise CovError("%s: unknown subsystem %r" % (where, name))
        if status not in STATUSES:
            raise CovError("%s: status %r not in %s" % (where, status, ", ".join(STATUSES)))
        u = uni[name]
        ev = r["evidence"].strip()
        if status == "UNTESTED" and ev in ("", "-"):
            pass
        else:
            if not ev or not os.path.isfile(os.path.join(root, ev)):
                raise CovError("%s: evidence doc %r does not exist" % (where, ev))
            if not (r["seed_or_date"] and (DATE_RE.match(r["seed_or_date"]) or SEED_RE.match(r["seed_or_date"]))):
                raise CovError("%s: seed_or_date must be YYYY-MM-DD or 'seed N'" % where)
        cases = 0
        if status in RUN:
            try:
                cases = int(r["cases"].replace(",", ""))
            except ValueError:
                cases = 0
            if cases < 1:
                raise CovError("%s: %s claims a V run but gives no case count" % (where, status))
            if ev not in doc_cache:
                with open(os.path.join(root, ev), encoding="utf-8") as f:
                    doc_cache[ev] = f.read()
            if not doc_has_number(doc_cache[ev], cases):
                raise CovError("%s: the case count %d does not appear in %s" % (where, cases, ev))
        if status == "TESTED-DEVIATION" and not (DID_RE.search(r["note"]) or note_doc(r["note"], root)):
            raise CovError("%s: TESTED-DEVIATION needs the deviation id (D1, S2, ...) or an existing docs path in the note" % where)
        keys, n = resolve_tokens(u, r["rows"], where)
        if keys is not None:
            dup = claimed[name] & keys
            if dup:
                raise CovError("%s: row(s) already in another ledger line: %s" % (where, ", ".join(map(str, sorted(dup, key=str)[:5]))))
            claimed[name] |= keys
        counted[name] += n
        if counted[name] > u.total:
            raise CovError("%s: %s ledger rows exceed the %d rows of the universe" % (where, name, u.total))
        r["_n"], r["_cases"], r["_keys"] = n, cases, keys
        per[name].append(r)
    return per


def tally(uni, per):
    out = {}
    for name, u in uni.items():
        t = {s: 0 for s in STATUSES}
        cases = 0
        for r in per[name]:
            t[r["status"]] += r["_n"]
            cases += r["_cases"]
        t["UNTESTED"] = u.total - sum(v for k, v in t.items() if k != "UNTESTED")
        t["cases"] = cases
        out[name] = t
    return out


def untested_keys(uni, per, name):
    u = uni[name]
    if u.keys is None:
        return None
    done = set()
    for r in per[name]:
        if r["status"] != "UNTESTED" and r["_keys"] is not None:
            done |= r["_keys"]
    return u.keys - done


def load_gaps(uni, per, root=ROOT):
    rows = read_tsv(DIR + "/gaps.tsv", ("rank", "subsystem", "rows", "impact", "source"), root)
    out = []
    for r in rows:
        where = "gaps.tsv:%d" % r["_line"]
        name = r["subsystem"]
        if name not in uni:
            raise CovError("%s: unknown subsystem %r" % (where, name))
        if not os.path.isfile(os.path.join(root, r["source"])):
            raise CovError("%s: source doc %r does not exist" % (where, r["source"]))
        try:
            r["_rank"] = int(r["rank"])
        except ValueError:
            raise CovError("%s: rank is not a number" % where)
        u = uni[name]
        r["_shown"] = r["rows"].strip()
        if r["_shown"] == "untested":
            t = tally(uni, per)[name]["UNTESTED"]
            if t < 1:
                raise CovError("%s: %s has no untested rows left; drop the gap" % (where, name))
            r["_shown"] = "untested (%d of %d)" % (t, u.total)
        elif u.keys is not None and r["rows"].strip() != "all":
            keys, _ = resolve_tokens(u, r["rows"], where)
            stale = keys - untested_keys(uni, per, name)
            if stale:
                raise CovError("%s: gap rows already have a status in the ledger (drop or edit the gap): %s"
                               % (where, ", ".join(map(str, sorted(stale, key=str)[:5]))))
        out.append(r)
    out.sort(key=lambda r: r["_rank"])
    return out


def pct(n, d):
    return "%d%%" % int(100.0 * n / d + 0.5) if d else "-"


def ledger_date(per):
    dates = [r["seed_or_date"] for rs in per.values() for r in rs if DATE_RE.match(r["seed_or_date"])]
    return max(dates) if dates else "unknown"


def render(uni, per, gaps):
    t = tally(uni, per)
    L = []
    L.append("# Rule coverage of d2d against game.exe 1.14d")
    L.append("")
    L.append("> status: GENERATED by `python3 tools/coverage/coverage.py --write` from `denominators.tsv`, `ledger.tsv` and "
             "`gaps.tsv` in this directory. Do not edit by hand; `--check` fails when this file is stale. How to add a row: "
             "[README.md](README.md). Ledger as of %s (the newest date in `ledger.tsv`)." % ledger_date(per))
    L.append("")
    L.append("Each subsystem is a fixed set of rows taken from the game's own data tables (`denominators.tsv`). A row is "
             "**run-tested** when a differential run (evidence tag V run, with a case count) compared game.exe's own code with "
             "d2d's real rules on it; it is **read-only** when the game.exe code was read (V read) but nothing was run; "
             "**bot-only** when only our bot models it; everything else is **untested**, which means unknown, not wrong.")
    L.append("")
    L.append("| subsystem | rows | run 0 diff | run deviation | read-only | bot-only | untested | run-tested | compared (run + read) | cases run |")
    L.append("|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|")
    sums = {s: 0 for s in STATUSES}
    tot = cases = 0
    for name, u in uni.items():
        x = t[name]
        tot += u.total
        cases += x["cases"]
        for s in STATUSES:
            sums[s] += x[s]
        run = x["TESTED-0-DIFF"] + x["TESTED-DEVIATION"]
        L.append("| %s (%s) | %d | %d | %d | %d | %d | %d | %s | %s | %s |" % (
            name, u.label, u.total, x["TESTED-0-DIFF"], x["TESTED-DEVIATION"], x["READ-ONLY"], x["BOT-ONLY"],
            x["UNTESTED"], pct(run, u.total), pct(run + x["READ-ONLY"], u.total), "{:,}".format(x["cases"])))
    run = sums["TESTED-0-DIFF"] + sums["TESTED-DEVIATION"]
    L.append("| **total** | **%d** | **%d** | **%d** | **%d** | **%d** | **%d** | **%s** | **%s** | **%s** |" % (
        tot, sums["TESTED-0-DIFF"], sums["TESTED-DEVIATION"], sums["READ-ONLY"], sums["BOT-ONLY"], sums["UNTESTED"],
        pct(run, tot), pct(run + sums["READ-ONLY"], tot), "{:,}".format(cases)))
    core = {s: sum(t[n][s] for n in uni if not ASPECT(n)) for s in STATUSES}
    core_tot = sum(u.total for n, u in uni.items() if not ASPECT(n))
    core_run = core["TESTED-0-DIFF"] + core["TESTED-DEVIATION"]
    L.append("")
    L.append("Excluding the single-aspect subsystems (`skillfx.*`, `monster.flags`), which test one property across many rows "
             "and so reach 100%% easily: %d rows, run-tested %d (%s), compared %d (%s), read-only %d, untested %d." % (
                 core_tot, core_run, pct(core_run, core_tot), core_run + core["READ-ONLY"],
                 pct(core_run + core["READ-ONLY"], core_tot), core["READ-ONLY"], core["UNTESTED"]))
    L.append("")
    L.append("Case counts add the cases of every ledger line; one differential that covers many rows counts once per line "
             "that cites it, so the column is an effort figure, not a unique-case total.")
    L.append("")
    L.append("## Evidence behind the numbers")
    L.append("")
    L.append("| subsystem | rows | status | evidence | cases | seed or date | note |")
    L.append("|---|---|---|---|---:|---|---|")
    for name in uni:
        for r in per[name]:
            L.append("| %s | %s | %s | `%s` | %s | %s | %s |" % (
                name, r["rows"].replace("|", "/"), r["status"], r["evidence"], "{:,}".format(r["_cases"]) if r["_cases"] else "-",
                r["seed_or_date"], r["note"].replace("|", "/")))
    L.append("")
    L.append("## Top gaps by likely play impact")
    L.append("")
    L.append("Hand-ranked in `gaps.tsv` from the class plans (`docs/classes`) and the work board; the generator refuses a gap "
             "whose rows already have a status in the ledger.")
    L.append("")
    L.append("| rank | subsystem | rows | why it matters | source |")
    L.append("|---:|---|---|---|---|")
    for r in gaps:
        L.append("| %d | %s | %s | %s | `%s` |" % (r["_rank"], r["subsystem"], r["_shown"].replace("|", "/"),
                                                   r["impact"].replace("|", "/"), r["source"]))
    L.append("")
    L.append("## Limits of this measure")
    L.append("")
    L.append("- The denominators are only what the data tables list (rows, plus the hand-listed rule set under `formulas`). "
             "Code paths outside the tables (movement, pathing, packets, timing, the server loop) are not counted at all; "
             "the number says nothing about them.")
    L.append("- A row's status is the best evidence on any one aspect of it, named in the ledger note. A skill marked "
             "run-tested for its missile pipeline can still be untested for its mana or its states; the cross-cutting "
             "`skillfx.*` subsystems exist to show those aspects separately.")
    L.append("- Untested means unknown, not wrong. Run-tested with 0 diff means 0 differences on the cases generated, not "
             "a proof; the case count and seed are in the ledger.")
    L.append("- READ-ONLY rows were read in game.exe by an agent and are not offered as proof of anything.")
    L.append("- The ledger is hand-maintained. The generator checks that every cited document exists, that a V run claim "
             "has a case count that appears in the cited document, and that rows are not claimed twice; it cannot check "
             "that the document is right.")
    L.append("- Denominators were read from the patched 1.14d tables of the worker host's install; `--tables DIR` "
             "re-checks them against any local extracted cache.")
    L.append("")
    return "\n".join(L)


# --- denominators from a local data cache ---------------------------------------------------------------------------

def read_table(dirpath, name):
    p = os.path.join(dirpath, name + ".tsv")
    with open(p, encoding="utf-8", newline="") as f:
        rd = csv.reader(f, delimiter="\t", quoting=csv.QUOTE_NONE)
        rows = list(rd)
    head = rows[0]
    return [dict(zip(head, r + [""] * (len(head) - len(r)))) for r in rows[1:]]


def isint(s):
    return bool(re.fullmatch(r"\d+", s.strip()))


def universes_from_tables(d):
    """subsystem -> (kind, sorted keys or count). Mirrors docs/coverage/README.md section 'Denominators'."""
    out = {}
    sk = [r for r in read_table(d, "skills") if isint(r["Id"])]
    for code, _ in CLASSES:
        out["skills." + code] = ("ints", {int(r["Id"]) for r in sk if r["charclass"] == code})
    out["skills.monster"] = ("ints", {int(r["Id"]) for r in sk if r["charclass"] == ""})
    out["skillfx.mana"] = ("ints", {int(r["Id"]) for r in sk})
    out["skillfx.damage"] = ("ints", {int(r["Id"]) for r in sk if r.get("EType", "").strip()})
    out["missiles"] = ("ints", {int(r["Id"]) for r in read_table(d, "missiles") if isint(r["Id"])})
    ms = [r for r in read_table(d, "monstats") if isint(r["hcIdx"])]
    out["monster.stats"] = ("ints", {int(r["hcIdx"]) for r in ms})
    out["monster.flags"] = ("ints", {int(r["hcIdx"]) for r in ms})
    out["monster.ai"] = ("names", {r["AI"] for r in ms if r["AI"].strip()})
    out["states"] = ("ints", {int(r["id"]) for r in read_table(d, "states") if isint(r["id"])})
    out["itemtypes"] = ("names", {r["Code"] for r in read_table(d, "itemtypes") if r["Code"].strip()})
    out["items.armor"] = ("count", len([r for r in read_table(d, "armor") if r["code"].strip()]))
    out["items.weapons"] = ("count", len([r for r in read_table(d, "weapons") if r["code"].strip()]))
    out["treasure"] = ("count", len([r for r in read_table(d, "treasureclassex") if r["Treasure Class"].strip()]))
    out["levels"] = ("ints", {int(r["Id"]) for r in read_table(d, "levels") if isint(r["Id"])})
    return out


def keys_text(kind, v):
    if kind == "ints":
        return compact_ints(v)
    if kind == "names":
        return ",".join(sorted(v))
    return ""


def check_tables(uni, d):
    got = universes_from_tables(d)
    errs = []
    for name, (kind, v) in got.items():
        u = uni.get(name)
        if u is None:
            errs.append("%s: in the tables but not in denominators.tsv" % name)
            continue
        total = len(v) if kind != "count" else v
        if total != u.total:
            errs.append("%s: denominators.tsv says %d, the tables have %d" % (name, u.total, total))
        elif kind in ("ints", "names") and v != u.keys:
            errs.append("%s: same count (%d) but different keys" % (name, total))
    return errs


def write_denominators(uni, d, root=ROOT):
    got = universes_from_tables(d)
    path = os.path.join(root, DIR, "denominators.tsv")
    lines = ["#\tFixed row universes. Regenerate the table-backed rows with: coverage.py --tables DIR --write-denominators",
             "#\tkind: ints = numeric row ids (ranges), names = row keys, count = rows counted only (ledger uses `all` or `#N`), rules = hand-listed rule set.",
             "\t".join(("subsystem", "label", "kind", "total", "keys", "source"))]
    for name, u in uni.items():
        if name in got:
            kind, v = got[name]
            total = len(v) if kind != "count" else v
            lines.append("\t".join((name, u.label, kind, str(total), keys_text(kind, v), u.source)))
        else:
            lines.append("\t".join((name, u.label, u.kind, str(u.total), ",".join(sorted(u.keys)) if u.keys else "", u.source)))
    with open(path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")


def build(root=ROOT):
    uni = load_universes(root)
    per = load_ledger(uni, root)
    gaps = load_gaps(uni, per, root)
    return uni, per, gaps, render(uni, per, gaps)


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--write", action="store_true")
    ap.add_argument("--check", action="store_true")
    ap.add_argument("--tables", metavar="DIR")
    ap.add_argument("--write-denominators", action="store_true")
    a = ap.parse_args(argv)
    try:
        uni = load_universes()
        if a.tables:
            if a.write_denominators:
                write_denominators(uni, a.tables)
                print("wrote %s/denominators.tsv" % DIR)
                uni = load_universes()
            errs = check_tables(uni, a.tables)
            if errs:
                for e in errs:
                    print("coverage: " + e, file=sys.stderr)
                return 1
            print("coverage: denominators match the tables in %s" % a.tables)
        _, _, _, text = build()
    except CovError as e:
        print("coverage: %s" % e, file=sys.stderr)
        return 1
    path = os.path.join(ROOT, OUT)
    if a.check:
        cur = None
        if os.path.exists(path):
            with open(path, encoding="utf-8") as f:
                cur = f.read()
        if cur != text:
            print("coverage: %s is stale; run python3 tools/coverage/coverage.py --write" % OUT, file=sys.stderr)
            return 1
        return 0
    if a.write:
        with open(path, "w", encoding="utf-8") as f:
            f.write(text)
        print("wrote %s" % OUT)
        return 0
    print(text)
    return 0


if __name__ == "__main__":
    sys.exit(main())
