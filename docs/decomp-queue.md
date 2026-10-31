# Decomp lane: ranked queue and rules

> status: current, 2026-10-31 (agent DECOMPBELT created it; the user started the lane: "start decomp lane"). The coordinator keeps the ranks and the status column; workers add a result line to their item.

The decomp lane answers one kind of question: **"is d2d (or our bot's assumption) right about game.exe?"** It is driven by BOT NEEDS, not by coverage. A question becomes a trace in game.exe 1.14d (disassembly with `tools/emu/disasm.py`, the unicorn oracle in `tools/emu/*.py`), compared with d2d's code and with our bot's model.

## Rules

1. Entry: an item enters the queue when a drill, a review, a soak log or the SME raises "is d2d / our assumption right?" for a behaviour the bot depends on. The trigger is written in the row.
2. Rank by bot impact: lethal or item-loss primitives first (docs/primitives.md crit S, A), then the fidelity risk map (docs/research/d2d-fidelity-riskmap.md), then the rest.
3. One agent per item, Opus, offline (no game, no d2d slot), its own upstream read worktree under the PR checkout dir on the worker host. GitHub GET-only.
4. Results come back as one of: (a) **finding**: our bot differs from game.exe (and d2d): a row for the primitives fix queue (docs/primitives-review.md style); (b) **PR candidate**: d2d differs from game.exe, proven by a differential under docs/process/upstream-pr-standard.md (>= 100k cases or exhaustive, mismatches before > 0, after 0, verifier attached), local branch only, posted only on the user's go; (c) **clean**: no deviation (a valid result; the risk moves from U to V).
5. Every result: a trace doc `docs/research/<topic>-trace.md` (rules with addresses, V/D/S/I/U tags, d2d's version, the bot's model, findings and deviations tables, the answer the drills should expect), and one line per result in the coordinator's report.
6. Guard: workers never edit docs/adr, GLOSSARY.md or docs/decisions.md; proposed text goes in the report.

## Queue

| id | question | why the bot needs it (trigger) | evidence | size | status |
|---|---|---|---|---|---|
| 1 | belt and ground-potion pickup: where a picked potion goes, belt layout and column kind, shift/ctrl clicks | P7 (939 "potion already in the pack" skips), P1-P3; review F7/F8; SME said the game auto-belts | docs/research/belt-pickup-trace.md; PR candidate pr-belt-*.md | M | done 2026-10-31: findings B1-B5 (pickit and belt model); d2d D1 (isc/tsc) proven 3,111 -> 0, branch `trace/belt` local; D2-D4 client gaps |
| 2 | Tome of Town Portal: empty-tome click, cast requirements, where the portal opens | P10 (2,683 failed TP attempts, 710 "no portal opened") | agent DECOMPTOME (in progress) | M | in progress |
| 3 | waypoint menu: row choice, act tabs, the 10 s refusal after use | P23 (220 + 198 waypoint failures; SME unsure on the refusal) | - | M | queued |
| 4 | skill bind and cast: hotkey binding semantics (left/right, the bound skill vs the selected), cast with the cursor on a unit | P26 (O, UNVERIFIED) | - | M | queued |
| 5 | repair, sell and vendor price formulas | P15-P17 (36 "repair not confirmed"; gold floor tuning); SME unsure | mpq-recheck-pr-claims.md (price data, partial) | M | queued |
| 6 | Static Field (floor, life - 1, negative resistance) and cold length / chill on the hit | DECOMPSKILL follow-ons (read, not run) | skill-damage-trace.md "Not covered" | S-M | queued |
| 7 | monster ranged / melee / specials / walk speed | risk map ranks 6, 9, 7, 10 (deaths, kill time) | monster-attack-trace.md (partial) | L | queued |
| 8 | pets and monster foes beyond PR #134 (summons, merc targeting) | merc policy, deaths near pets | pr-get-hit-rule-proof.md (#134 scope) | M | queued |

## Result log (one line each)

- 1 belt/pickup (DECOMPBELT, 2026-10-31): ground potion -> belt first (matching column, else empty column for hp/mp/rv), inventory open -> cursor; bot findings B1-B5; d2d D1 PR candidate READY (verifier holds), pending the user's go.
