# Rule coverage table: how to read it and how to add a row

> status: current, 2026-11-01 (card coveragetable). The table is `rule-coverage.md` (generated; never edit by hand).

The question it answers: how many of the rows in game.exe 1.14d's own data tables have we compared against d2d, and how. Method 1 of the completeness measure; it counts rows of the tables, not code paths (limits at the bottom of `rule-coverage.md`).

## Files

| file | who writes it | what |
|---|---|---|
| `denominators.tsv` | generated for table-backed rows, hand-listed for `formulas` | the fixed row universe per subsystem (ids, names or a count) |
| `ledger.tsv` | by hand, after a differential or a read | status per row range with its evidence |
| `gaps.tsv` | by hand | the ranked top gaps (rows must still be untested) |
| `rule-coverage.md` | `python3 tools/coverage/coverage.py --write` | the table |

## Statuses and tag rules

- `TESTED-0-DIFF`: a differential run (evidence tag V run: game.exe's own code under the emulator against d2d's real rules) on those rows showed 0 differences. Needs `cases` (a number of at least 1) and `seed_or_date`; the number must appear in the evidence doc (plain or with thousands commas), so the doc and the ledger cannot drift apart silently.
- `TESTED-DEVIATION`: a V run found a d2d deviation. Same requirements, plus the deviation id in `note` (D1, S2, C3, ...) or a `docs/...md` path in `note` that exists.
- `READ-ONLY`: V read in game.exe (disassembly), no run against d2d. Evidence doc and date, no case count needed. Never offered as proof.
- `BOT-ONLY`: our bot models it, no game.exe comparison. Evidence doc and date.
- `UNTESTED`: the default for every row without a ledger line; an explicit line is allowed with evidence `-`.
- Secondary (S), inference (I) and unknown (U) evidence never earns a status above `UNTESTED`.

A row is classified by the best evidence on one aspect of it; say the aspect in `note`. Aspects that cut across rows have their own subsystems (`skillfx.mana`, `skillfx.damage`, `monster.flags`) so they do not make a skill look done.

## Adding a row after a differential

1. Finish the trace doc with the case count and seed (`docs/research/<topic>-trace.md`).
2. Append a line to `ledger.tsv` (tab separated, columns `subsystem rows status evidence cases seed_or_date note`). `rows` is ids and ranges (`36,38-40`), names for `monster.ai` and `itemtypes`, `F01`-style keys for `formulas` (names are in the comment line of `denominators.tsv`), `all`, or `#N` for the count-only subsystems (`items.*`, `treasure`).
3. If the row was a listed gap, remove it from `gaps.tsv` (the generator refuses a gap that already has a status).
4. `python3 tools/coverage/coverage.py --write`, commit the four files. `--check` runs in `tools/ci/scripts.sh` (and so in `tools/ci/check.sh`) next to the unit tests (`python3 -m unittest discover -s tools/coverage/tests -t tools/coverage`).

The generator refuses: an unknown subsystem or row, a row claimed by two lines, more rows than the universe, a missing evidence doc, a V run claim with no case count or a count missing from the doc, a bad date.

## Denominators

Row counts come from the patched 1.14d tables as the bot extracts them (`internal/data`, docs/data.md): `skills.txt` by `charclass`, `missiles.txt` Id, `monstats.txt` hcIdx and distinct AI, `states.txt` id, `itemtypes.txt` Code, `armor.txt` and `weapons.txt` rows, `treasureclassex.txt` rows, `levels.txt` Id. The tables are not committed (Blizzard data), so the counts and id lists live in `denominators.tsv`. To re-check them on a machine that has the cache: `python3 tools/coverage/coverage.py --tables <cache dir>`; to regenerate them: add `--write-denominators` (keep the `formulas` comment line by hand). `skillfx.damage` is the skills with an element type, the set the skill-damage differential covered. `formulas` is a hand-listed rule set (F01..F18), not a table.
