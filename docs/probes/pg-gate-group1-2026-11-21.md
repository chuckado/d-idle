# Proving Ground gate, first group (2026-11-21)

> status: record of the first rollout group of docs/design/fable-review-2026-11-21/proving-ground-gate.md (owner approved the plan and its ten answers on 2026-11-21). Cards TASK-733 to TASK-756 hold the 16 cards; this group delivers pgprov, pgdigest, pgtaps step 1, pgjournal and pgledger. Nothing here ran on a played hero: every run is a CHEATED or SETUP scenario hero, mechanics evidence only.

## What was built (branch agent/pggroup1, stacked pg1 to pg5)

| card | task | what | verdict changes |
|---|---|---|---|
| pgprov | TASK-733 | result carries `prov` (bot commit and dirty flag, bot and d2d sha256, host role, load, pace); a bot run without `virtual run length reached` is BLOCKED | a bot that exits early without an error was judged, now BLOCKED |
| pgdigest | TASK-735 | D1 assertion digest (discrete facts by value, ranges by pass or fail bucket) written per run; `determinism` fails hard on verdict or D1, reports the whole-state digest as drift only | none for a scenario; a pair can now fail on D1 |
| pgtaps step 1 | TASK-738 | the board tapper polls on a 30 virtual-second grid read from the bot (`/api/thoughts` `now`), wall fallback after 90 s without a reading; `taps_at` in the result | board-taps scenarios tap at the first grid point after the offer, not at a wall instant. API calls stay wall-clock; scripted taps inside the bot are NOT done |
| pgjournal | TASK-737 | `run -journal`, `determinism-run -n N` (sequential repetitions, compared with the first, journals compared with d2ddet) | none |
| pgledger | TASK-743 | ledger line per run written by exec-one into `<out>/ledger.jsonl` on the host and shipped through `PG_LEDGER_CMD`; refused without provenance or setup tag; `d2scenario ledger status-doc\|check` | none |

Ops side (private repo, described by role): the board-host ingest command (validate, append, idempotent by run id, exclusive lock) and a script that regenerates the status page from the ledger, both with an offline test, merged to the ops master. The ledger file lives on the board host next to the status files.

Offline evidence for D1: recomputed from the stored asserts of the 2026-11-19 run records (`docs/probes/provingground-run-2026-11-19/run2`): the `cain.stones`, `cain.cage` and `cain.scroll` (Sorceress) pairs keep verdict and D1 and only the state digest drifts; `chicken.recovery` fails hard on D1 (same RED verdict, different facts) and `loot.rare-ring-classified` flips RED/GREEN.

## The journal pair (design 2.5)

Builder long lane, job 20261121-002134-7785, runner of branch pg4 (it already had pgprov, pgdigest and the virtual tapper), `tools/provingground/run.sh determinism-run -n 3 -id town.resupply,act1.burial.blood_raven -class Paladin -seed 11 -parallel 1`, 29 builder minutes. Hero setup: CHEATED/SETUP scenario hero, mechanics only.

| comparison (rep 1 against 2 and against 3) | result |
|---|---|
| verdict | equal, both scenarios, both pairs |
| assertion digest D1 | equal, both scenarios, both pairs |
| whole-state digest | equal, both scenarios, both pairs (zero drift; the 2026-11-19 `town.resupply` runs drifted on gold) |
| journal, `town.resupply` | identical for 28769 records (26 equal checkpoints, last at tick 2002), then a COMMAND differs at tick 2242 at equal state: one run sends `debug presets`, the other `debug missiles` |
| journal, `act1.burial.blood_raven` | identical for 32622 records (31 equal checkpoints, last at tick 7751), then at tick 7838: `debug players` against `debug missiles` |

Reading (the layer, V by d2ddet's own classification): no engine divergence (every checkpoint equal up to the first difference) and no divergence in any game-visible outcome; the first difference is a bot-side choice of which READ-ONLY probe to send, at equal observed state. Same position in both runs of the pair and in both repetition pairs, so it is systematic, not noise. The cause is NOT pinned to a line (U): a refresh of the presets or players view on a timer or an ordering by map iteration are the candidates; card pgclock (TASK-739) owns it. The wall-clock tapper (H1) is not what moved the journal: the Start tap happened in all three repetitions.

Caveats: results.json and journals were not pulled (the `--pull` path mapping of labrun on this checkout failed, a pull of `provingground-out` from the builder mirror finds nothing); the numbers above are from the job log, which prints the same comparison, and the artifacts stay on the builder under `provingground-out/20261121-002150/rep1..3`. The scenario verdict word (GREEN or RED) is not in that log, only that it is equal. In the three `blood_raven` repetitions only repetition 1 shows a Hand in tap in its taps log, which did not change verdict or digests. The ledger spool of this run does not exist (pg4 predates pgledger).

## Ledger status

Code, ingest command and status page generator exist and pass their offline tests; no scenario has run with `PG_LEDGER_CMD` set, so the ledger on the board host holds no line yet and `docs/process/proving-ground-status.md` is not generated yet. Turning it on: set `PG_LEDGER_CMD` for the builder runs (an ssh to the board host running the ingest command) and `PG_EVENT`, `PG_BRANCH`, `PG_CARD`, `PG_JOB` per job. Nothing is EARNED yet (no earned-save card), runs BLOCKED before exec-one are not ledgered.

## Next (cost-cut group, not started)

pgstability, pgclock, pghero, pgselfcheck, pgpage, mgonejob, mgcache, mglanes, mgwatchdog, pgnightly, pgearned, plus the scripted-taps half of pgtaps.
