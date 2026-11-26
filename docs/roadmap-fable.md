# Fable roadmap: every recommendation, its state, and what flips it

> status: current (card fableroadmap, 2026-11-26; written on user approval of the same day). A living tracker: the coordinator edits the STATE, evidence and next-action cells when a card merges or a knob flips; the dark-launch ledger (`data/darklaunch.json`, rule `docs/process/dark-launch.md`) is the machine-checked half for default-0 knobs. It holds no counts that change: tally by `grep -c '| built-dark |' docs/roadmap-fable.md` (and the other state words).
>
> Sources: the logic review of 2026-11-25 ([research/fable-logic-review-2026-11-25.md](research/fable-logic-review-2026-11-25.md), proposals P1 to P12, tests T1 to T7, cause classes A to K); the design review of 2026-11-19 ([design/fable-review-2026-11-19/](design/fable-review-2026-11-19/00-summary.md), roadmap R1 to R17, ADRs 0020 to 0024); the four reviews of 2026-11-21 (Proving Ground gate, combat engine, quest board, town planner; [design/fable-review-2026-11-21/](design/fable-review-2026-11-21/), ADRs 0025 to 0028); the map and UI reviews of 2026-11-22 ([design/fable-review-2026-11-22/](design/fable-review-2026-11-22/)).
>
> Evidence tags: **V** read in this tree at the commit named (grep, `git log`, `docs/knobs.md`, a task file); **S** another doc's claim; **U** not known from the repo. Card status in `backlog/tasks/` lags merges (several cards read Proposed or In Progress while their commits are on main); the evidence column therefore cites `git log origin/main --grep` and the knob table, not the card's status line.

## STATE words

`not started` (no code, maybe a card) | `in flight` (code partly merged or an agent running) | `built-dark` (merged, behind a knob at its default 0 or a mode that does nothing; has a ledger row) | `trial` (switched on in the sandboxes, not default) | `enabled` (default on, or a tool that is simply in use) | `retired` (decided against or deleted).

## Standard enable rule (user decision 2026-11-26)

Flip, watch 24 h, auto-revert: a dark feature is switched on the sandboxes as soon as its scenario, drill or replay passes; the invariant monitor watches for 24 h; a HARD invariant breach (a death, a stall, a potion not drunk, a refusal-as-failure storm) reverts the knob automatically; otherwise it stays on and the ledger row closes. No criterion needs more than 24 h of live observation; longer evidence comes from scenarios and replays. The owner of every flip is the coordinator, due next session. The tool is card `autorevert1`; until it exists the coordinator reverts by hand when the monitor reports a hard breach.

## Operating rule

Before the coordinator approves a new patch card (a `bug` or `bot-fix` card for the bot), the coordinator names its cause class (A to K of review 2026-11-25 section 2.2: A two deciders undo each other, B refusal counted as failure, C supply and economy, D navigation geometry, E stale state after portal, death or new game, F a fact the backend lacks, G threshold moved by one incident, H per-quest special case, I class-specific combat, J tests and flakes, K not visible) and checks that class against this tracker and `data/darklaunch.json`. If a row covers the class (see the "kills" column of section "Cause class to row" below), the fix card is attached to that row (its id goes into the row's card cell and the card says so in its description) and the question asked is "does the row's flip remove this cause?", not "what number do we move?". A fix card that repeats a class whose row is `built-dark` needs the coordinator to state why the row has not been flipped, or it waits. A fix card may not add a knob without removing one or the owner's word (P10). Class D (navigation geometry) is the only class this rule does not redirect: it is the game's own complexity and keeps being fixed case by case with a fixture each time (review 7.4).

### Cause class to row

| class | what | rows that kill it |
|---|---|---|
| A | deciders undo each other | ARB, P6, P2, TEMP (below) |
| B | refusal counted as a failure | P2, P4, BSM |
| C | supply and economy | TOWN, P5 |
| D | navigation geometry | MAP rows (partial), none structural |
| E | stale state after portal, death, new game | P6, P2 |
| F | a fact the backend lacks | ADR22 rows, P9, INPUT rows |
| G | threshold moved by one incident | P4, P10, ZONE |
| H | per-quest special case | P9, BSM, EPIC rows |
| I | class-specific combat | DLV rows |
| J | tests, races, flakes | PG rows, T3 |
| K | not visible | P1, P7, T4 |

## Built but dark

The ledger rows (`data/darklaunch.json`, V) with the proposed flip order. Dates are EST, from the ledger; `criteria_met` is U for every row until a card records evidence (V `python3 -I tools/darklaunch/report.py`).

| order | feature | knob (default) | why this position | criteria (measurable) | blocker | cards |
|---|---|---|---|---|---|---|
| 0 | invariant monitor (not dark: log only) | none | it is the measuring stick for every flip below | none: it is the watcher (hard-breach counters, `inv=` in the minute line and ops status) | none | invmon |
| 1 | `bot.checkpointEnable` | 0 | smallest scope, an independent safety add; the Dark Wood deaths are its target | live drill green (portal cast healthy, retreat at low life, return after death), then flip on the sandboxes, 24 h invariant-monitor watch, automatic revert on a hard breach (death, stall, potion not drunk, refusal-as-failure storm) | the live drill has not run (U) | ckptlive |
| 2 | `town.plan` stage 2 | `town.plan` 0 (stages 1 to 3 merged) | largest cause class (C, 49 cards); stage code exists; needs only the flip and counts | stage 2 scenario set green, then flip on the sandboxes, 24 h invariant-monitor watch, automatic revert on a hard breach (death, stall, potion not drunk, refusal-as-failure storm); stages 3, 4, 5 the same, one at a time | P1 counts do not exist yet; owner looks before stages 3 to 5 (ADR 0028) | townflip, invmon, town5 to town8, town11 |
| 3 | `quests.boardSM` | 0 | removes the per-quest patch treadmill for the board; the machine, store, exec, holds, API and panel are merged (bsm1 to bsm9) | restart drill green with boardSM 1 (a park survives), then flip on the sandboxes, 24 h invariant-monitor watch, automatic revert on a hard breach (death, stall, potion not drunk, refusal-as-failure storm) | P1 counts; bsm checker (rules 12 to 14) not built | bsmdef, bsmchk, invmon |
| 4 | `bot.arbiterShadow` then the arbiter itself | 0 (shadow); acts: no knob yet | the core; do it after the counters (P1) and the outcome type (P2) exist | outcome rule on replays and incident scenarios (fewer monitor hits than the chain, no new ones), then the first family on one sandbox with a 24 h watch and automatic revert on a hard breach | P2, ticklog (T4), incident1 (P8) | arbmig, ticklog, outcome1, incident1 |
| 5 | `bot.zoneModel` | 0 | one margin for zone choice; the early-road thresholds of 2026-11-25 are the same question answered by hand | A/B seed sets per class (offline), then flip on the sandboxes, 24 h invariant-monitor watch, automatic revert on a hard breach (death, stall, potion not drunk, refusal-as-failure storm) | A/B sets per class; P4 provides the entry point | admit1, thrgov |
| 6 | `agency.planner` mode 2, `agency.autoAlloc` | 0 | the owner's central ask (path D); lower risk than 1 to 5 because the player decides by default | shadow lines replayed and reviewed with no wrong pick, owner go for mode 2, then flip on the sandboxes, 24 h invariant-monitor watch, automatic revert on a hard breach (death, stall, potion not drunk, refusal-as-failure storm) | dlv-data tier source (merged); review of shadow lines (U) | agencyp1b |
| 7 | combat dodge and stand knobs (`combat.missileDodge`, `combat.threatDodge`, `combat.standOK`, `combat.deliveryExpect`) | 0 | each has a live drill in its ledger row; none is blocked by structure | per row in the ledger | live drills (U) | none new; ledger rows own them |

Ledger rows to ADD or ADJUST: `data/darklaunch.json` is on main (mergeq #1323), so this branch adjusts the existing rows rather than listing them here: `bot.arbiterShadow` carries the outcome rule as its enable criterion (owner approved 2026-11-26); `town.plan` names stage 2 and the P1 dependency; `quests.boardSM` names the card that flips it. No new knob is added by this branch, so no new row is needed. Each card of the table below that adds a knob adds its ledger row in the same PR (rule 1 of `docs/process/dark-launch.md`): `outcome1`, `admit1` (one row per provider), `arbmig` (one row per family), `stepexec1` (per chunk), `ticklog`.

## Review of 2026-11-25 (logic review, card fablereview2)

| id | recommendation | review | card | STATE | knob (default) | evidence | blocker | next action | owner | enable criteria |
|---|---|---|---|---|---|---|---|---|---|---|
| P1 | invariant monitor, log only (T1) | 11-25 | invmon (TASK-955, agent/invariants) | enabled | none (log only) | V `Merge agent/invariants` (e201e0878) on main; U whether both sandboxes run it and the row set is complete | none | read the `inv=` counts (24 h watch rule) | coordinator | n/a (it is the watcher) |
| P2 | typed step outcome, Refused{need,wake} | 11-25 | outcome1 (TASK-956 to 961, agent/outcome1) | in flight | planned, default 0 | V branch agent/outcome1 with six task cards, queued (mergeq #1326); four partial exemptions listed in review 3.1 | merge | merge, then read counts | coordinator | scenario green, flip, 24 h monitor, auto-revert on a hard breach |
| P3 | reflex lane: latency, cancellation test, in-town check, drink by the safety check | 11-25 | reflex1 (TASK-962 is step a and b, agent/reflex1) | in flight | none for a, b, d | V branch agent/reflex1, queued (mergeq #1328); review 3.1 rows 3 and 4 | none: the owner approved the safe drink by the safety check (`docs/bot.md` rule change) on 2026-11-26 | steps a, b, d, then c | coordinator | scenario green, flip, 24 h monitor, auto-revert on a hard breach (safety latency over 2 s counts as one) |
| P4 | admission function MayEnter / MayLeaveTown | 11-25 | admit1 | not started | per provider, default 0 | V eleven judges named in review 1.2 | P2 | design doc, first provider in shadow | coordinator | per provider: shadow replay equals the old judge or is right in a reviewed sample, then flip, 24 h monitor, auto-revert on a hard breach |
| P5 | finish and flip town.plan, delete old counters | 11-25 | townflip, town5 to town8, town11 | in flight | `town.plan` 0 | V stages 1, 2, 3 merged (`git log --grep '^town[345]'`); knob default 0 in `docs/knobs.md`; V town6 to town12 have no merged commit | P1 counts before the flip | set stage 2 on both sandboxes, read counts | coordinator | scenario set green, flip, 24 h monitor, auto-revert; see Built but dark row 2 |
| P6 | arbiter migration with the outcome exit rule, TEMPORARY intents, lost-control hook | 11-25 | arbmig | not started | `bot.arbiterShadow` 0; acting knobs per family planned | V `internal/arbiter` last commit 2026-11-19 (`a8751bd8d`); `docs/arbiter.md` migration table still has the agreement rule (to be rewritten by arbmig) | P2, ticklog; owner approved the rule change 2026-11-26 (ADR text: agent/decisions1) | rewrite the table, build TEMPORARY | owner for each family go | Built but dark row 4 |
| P7 | explain: decision ring, /api/explain, operator status --why | 11-25 | explain1 (TASK-965, agent/explain1) | in flight | none | V branch agent/explain1 (ring, /api/explain, operator why); the ADR 0030 allowlist read verb was approved by the owner 2026-11-26 (ADR text: agent/decisions1) | P2 for the refusal records | finish | coordinator | endpoint answers on both sandboxes |
| P8 | incident to scenario (T2) plus report-only scenario tier (T7) | 11-25 | incident1 (TASK-971, agent/incident1) | in flight | none | V branch agent/incident1; pipeline audit row 6: the merge worker runs no scenario tier; report-only tier approved by the owner 2026-11-26 (ADR 0020 item 1 practice) | ops half (the report-only tier) is in the ops repo | finish; first three Questy deaths | coordinator | report-only tier 7 clean days, then owner decides (ADR 0027 item 2) |
| P9 | generic step executor with postconditions | 11-25 | stepexec1 | not started | per chunk, default 0 | V `internal/questspec` holds the declaration side only (review 2.4) | P2 | skeleton plus Cain scroll chunk | coordinator | chunk scenario green 2 of 3 runs with it on |
| P10 | threshold governance | 11-25 | thrgov | not started | none | V knob table 560 rows (review 1.3) | none | rule doc plus report-only CI check | coordinator | check in CI report-only for one session, then failing (the 24 h rule: evidence from scenarios, not weeks) |
| P11 | operator verbs hold, home, go, park | 11-25 | explain1 (TASK-965; hold, home, go, park verbs in the same branch) | in flight | none | V branch agent/explain1 subject lists "why hold home go park"; allowlist amendment approved by the owner 2026-11-26 | none | finish | coordinator | audited and rate-limited verbs on a sandbox |
| P12 | fault-injection soak (T5) | 11-25 | faultsoak | not started | none | V review 5.3 | invmon, outcome1; night window ADR 0027 | injection list | coordinator | nightly report-only; each hit becomes an incident scenario |
| T3 | property tests on the pure deciders | 11-25 | proptest1 | not started | none | U | none | combat, supply, arbiter first | coordinator | tests in CI |
| T4 | recorded sessions, tick line, `d2arbiter -follow` | 11-25 | ticklog | not started | planned, default 0 | V arbiter fixture facts are assumed, not logged (`docs/arbiter.md`) | none | tick line behind a knob | coordinator | one real night replayed |
| T6 | cancellation-latency test | 11-25 | reflex1 (step b) | not started | none | U | none | with P3 | coordinator | in CI |
| DL | no structural work parked at 0 without a flip date and a measure (7.4) | 11-25 | darklaunch, this tracker | enabled | ledger check in `tools/ci/check.sh` | V `tools/darklaunch/check.py` on main (merge of agent/darklaunch) | none | weekly `report.py` decisions | coordinator | every row past due gets a decision |
| DEC | four decision changes (7.5): ADR 0030 allowlist (why, hold, home, go, park); `docs/bot.md` safe drink by the safety check; arbiter exit rule by outcome; ADR 0020 incident-to-scenario report-only tier | 11-25 | explain1, reflex1, arbmig, incident1 | in flight | n/a | V all four APPROVED by the owner (popup 2026-11-26); the ADR and doc text is written by agent/decisions1 (cite `docs/decisions.md` once it lands) | none | decisions1 lands | coordinator | n/a |

## Design review of 2026-11-19 (card fablereview, ADRs 0020 to 0024)

| id | recommendation | review | card | STATE | knob (default) | evidence | blocker | next action | owner | enable criteria |
|---|---|---|---|---|---|---|---|---|---|---|
| R1 | Proving Ground scenario runner as the merge gate | 11-19, ADR 0020 item 1 | provingground (done), pgexpand, pgcain | trial | none | V `internal/scenario`, `tools/provingground/scenarios`; review 2.5 says the merge worker runs no scenario tier | merge-gate wiring (mgonejob; incident1 for the report-only tier) | see PG rows below | coordinator | T0 and T1 block every merge; T2 report-only, owner decides when to block (ADR 0027) |
| R2 | Search for Cain end to end as the first scenario set | 11-19 | pgcain, cainfix2, cainfix3, cainchain, cainscroll | enabled | none | V pgcain, cainfix2, cainchain, cainscroll Done in `backlog/tasks/`; cainfix3 Proposed | live verdict of a fresh hero (U) | none | coordinator | scenarios green 2 of 3 |
| R3 | agency P0: points by verb, sheet, autopilot switch, legality | 11-19, ADR 0020 item 2 | agencyp0 | enabled | `resource.pointsDecisions` 0 by design (the player decides) | V agencyp0 Done; ledger row says stays 0 by design | none | none | owner (product) | owner go only |
| R4 | intent arbiter | 11-19, ADR 0020 item 3 | arbitershadow (done), arbmig | built-dark | `bot.arbiterShadow` 0 | V `docs/knobs.md:342`; commits `a8751bd8d`, `f9aee01f9` (2026-11-19) | see P6 | see P6 | coordinator | Built but dark row 4 |
| R5 | quest chains as checkable data (questspec) | 11-19 | questspec | enabled | none | V `internal/questspec`, `TestShippedChains`; questspec Done | rules 12 to 14 not built (bsmchk) | bsmchk | coordinator | rules in CI |
| R6 | survival model as the one judge | 11-19 | survsd, survab8, admit1 | built-dark | `bot.zoneModel` 0 | V `docs/knobs.md:254`; ledger row | A/B sets per class; P4 | see Built but dark row 5 | coordinator | A/B per class, then default per archetype |
| R7 | agency P1: capability table, planner in shadow, then driving, viability line | 11-19 | agencyp1a (done), agencyp1b | built-dark | `agency.planner` 0 (mode 1 shadow, 2 drives) | V `docs/knobs.md:582`; ledger row | shadow review; U | agencyp1b | coordinator, owner go | see Built but dark row 6 |
| R8 | Act 2 through the quest board | 11-19 | epic-act2 | in flight | none | V `docs/epics/epic-act2.md`; blockers listed in the card (Duriel place_staff verb) | P9 (generic executor); Proving Ground rows for Act 2 | link the epic blockers to stepexec1 | coordinator | epic-act2 criteria |
| R9 | difficulty transition (Nightmare, Hell) | 11-19 | epic-normal, epic-nightmare, epic-hell | not started | none | V cards Proposed (`backlog/tasks/`) | Acts 2 to 5 | none now | coordinator | epic cards |
| R10 | Necromancer Summoner (upkeep, minions, curses) | 11-19 | epic-class-necromancer, dlv-upkeep-state, dlv-upkeep-minions, dlv-mark | not started | `combat.minionRead` 0 (ledger) | V the dlv cards are Proposed; no merged commit for them | dlv wave 0 done; live drill (U) | after dlv wave 1 | coordinator | class drill green on 3 seeds |
| R11 | Hammerdin and second Sorceress build | 11-19 | epic-class-paladin, epic-class-sorceress, dlv-stance-pack | not started | none | V cards Proposed | dlv-aim-point, dlv-cast-verbs | after dlv wave 2 | coordinator | per the epic cards |
| R12 | Barbarian | 11-19 | epic-class-barbarian | not started | none | V card Proposed | R10 or R11 | none now | coordinator | per the epic card |
| R13 | Acts 3 to 5 | 11-19 | epic-act3, epic-act4, epic-act5 | not started | none | V cards Proposed | R8 | none now | coordinator | per the epic cards |
| R14 | runeword and cube advisor, shrines, status reflexes | 11-19 | advisor1 | not started | none | V no card existed before this branch | agencyp1b | build after agencyp1b | coordinator | advisor lines on the Return screen reviewed by the SME agent |
| R15 | agency P3: share, import, NIP importer | 11-19 | agencyp3 | not started | none | V no card existed before this branch | agencyp1b | none now | coordinator | round trip of the eight shipped tables stays exact |
| R16 | Blizzard, Druid, Assassin, Amazon (ground target, traps, bow) | 11-19 | epic-class-druid, epic-class-assassin, epic-class-amazon, dlv-aim-point, dlv-aim-corpse | not started | none | V cards Proposed | dlv wave 2; fidelity A5, A7 to A9 | none now | coordinator | per the epic cards |
| R17 | input layer I1 to I7 | 11-19, ADR 0021 | inputlayer (I1, I2 done), inputl3, inputl4, inputl5 | in flight | `input.planner` 0, `input.hotkeyMap` 0 | V `docs/knobs.md:574-575`; inputlayer Done | I3 to I7 had no card before this branch; fork branch for I4 | inputl3 | coordinator | drill passes per verb; CI pointer-call count 0 on the d2d path |
| F | fidelity lane re-pointed by class and act | 11-19 | fid-* cards, fid-matrix-refresh | in flight | none | V cards fid-matrix-refresh (task-100) and the fid-* series exist | none | continuous | coordinator | per rule: a zero-difference differential or a listed shim (ADR 0023) |
| Q4 | ladder-only runewords enabled in d.idle mode (ADR 0024) | 11-19 | q4runes (merged) | built-dark | `items.ladderRunewords` 0 | V `docs/knobs.md:588`; `Merge agent/q4runes` on main | owner go to turn the rule on | owner go | owner | gates stay vanilla; flag shown to the player |
| Q5 | engine facts as reads never make a run CHEATED; level layout read (ADR 0022) | 11-19 | k1layout, knownedge, warptrigger | built-dark | `quests.layoutRead` 0, `nav.fullGrid` 0, `nav.warpRead` 0 | V `docs/knobs.md:51-52, 555`; commits `Merge agent/k1layout`, knownedge, warptrigger on main | live drills (U) | drills per ledger rows | coordinator | quest layout drill: objects found by read in 3 areas; no stuck regressions on 5 areas; warp drill across 5 areas |
| Q7 | fidelity acceptance per rule (ADR 0023) | 11-19 | fid-* | enabled | none | V ADR 0023 accepted; header amends 0015 | none | none | coordinator | n/a |

## Reviews of 2026-11-21

### Proving Ground gate (card fablepg, ADR 0027; 16 cards)

| id | recommendation | review | card | STATE | knob (default) | evidence | blocker | next action | owner | enable criteria |
|---|---|---|---|---|---|---|---|---|---|---|
| PG1 | pgprov: result records bot commit, d2d hash, host role, load | 11-21 | pgprov | in flight | none | V `git log --grep '^pgprov'` on main; card In Progress | none | finish and merge | coordinator | records present in a run |
| PG2 | pgdigest: assertion digest D1, determinism compare | 11-21 | pgdigest | in flight | none | V commit on main; card In Progress | none | finish | coordinator | determinism pair run |
| PG3 | pgjournal: journal variable, `determinism-run -n N` | 11-21 | pgjournal | in flight | none | V two commits on main; card In Progress | none | finish | coordinator | pair proof written up |
| PG4 | pgtaps: scripted taps on the virtual clock | 11-21 | pgtaps | in flight | none | V one commit on main | none | finish | coordinator | scenario run id in the result |
| PG5 | pgclock: goals clock, farm timers on the env clock | 11-21 | pgclock | in flight | none | V `Merge agent/pgc5` on main | none | finish | coordinator | two-build test |
| PG6 | pgstability: stable, quarantined, tier retag | 11-21 | pgstability | not started | none | V card Proposed, no merged commit | PG2, PG3 | build | coordinator | auto-quarantine on a verdict flip |
| PG7 | pghero: hero library with played and banked profiles | 11-21 | pghero | not started | none | V card Proposed | none | build | coordinator | survival prep uses it |
| PG8 | pgselfcheck | 11-21 | pgselfcheck | not started | none | V card Proposed | PG7 | build | coordinator | self-check passes on the prep d2d |
| PG9 | pgledger: JSONL ledger and ingest | 11-21 | pgledger, pgledgeron | in flight | none | V `pgledgeron` commit on main | none | finish | coordinator | results written on the host |
| PG10 | pgpage: board page and status doc | 11-21 | pgpage | not started | none | V card Proposed | PG9 | build | coordinator | page shows the tiers |
| PG11 | mgonejob: merge gate as one locked builder job | 11-21 | mgonejob | not started | none | V card Proposed, no commit | ops repo | build in the ops repo | coordinator | positive ran-marker |
| PG12 | mgcache: cached whole-repo gate, nightly uncached race | 11-21 | mgcache | in flight | none | V two commits on main; card In Progress | 20-merge measurement | measure | coordinator | measured over 20 merges, then default |
| PG13 | mglanes | 11-21 | mglanes | enabled | none | V card Done, `Merge agent/pgc1` | none | none | coordinator | n/a |
| PG14 | mgwatchdog | 11-21 | mgwatchdog | enabled | none | V card Done, `Merge agent/pgc2` | none | none | coordinator | n/a |
| PG15 | pgnightly: one nightly job (T2, T3, T4) | 11-21 | pgnightly | not started | none | V card Proposed | night window ADR 0027 item 4 | build | coordinator | one night run, at most 3 runs |
| PG16 | pgearned: earned save chain, `pre.from_save` | 11-21 | pgearned | not started | none | V card Proposed | save portability U | last | coordinator | one earned act run |
| PGT | tiers T0 to T4; T0 and T1 block every merge, T2 report-only then blocking after 7 clean days | 11-21, ADR 0027 | incident1 (TASK-971, report-only tier), mgonejob | in flight | none | V review 2.5: no scenario tier in the merge worker | PG11 | wire T0 and T1; T2 report-only | owner (after 7 clean days) | owner decides when to block |

### Combat engine (card fablecombat, ADR 0026; 16 cards plus one parked)

| id | recommendation | review | card | STATE | knob (default) | evidence | blocker | next action | owner | enable criteria |
|---|---|---|---|---|---|---|---|---|---|---|
| DLV0a | one capability table, generated views, tiers | 11-21 | dlv-data | enabled | none | V `data/capabilities/capabilities.json`, `internal/delivery/gen.go`; commits `80c5939bf`, `0bbdc37d3` | none | none | coordinator | generated views equal checked-in files |
| DLV0b | Rotation type, FromProfile adapter, golden parity | 11-21 | dlv-rotation | in flight | U (see the card) | V commit `dd605af3b` (WIP in the subject); card result says done, NOT verified live | live look (U) | confirm merge state | coordinator | golden parity over the corpus |
| DLV0c | expectation ledger | 11-21 | dlv-expect | built-dark | `combat.deliveryExpect` 0 (mode 2 max) | V `docs/knobs.md:127`; commits `a82801ab5`, `ea683eb38` | live look (U) | review mode 1 failure log | coordinator | mode 1 failure log replayed, no false failures; flip, 24 h monitor, auto-revert on a hard breach |
| DLV0d | read verbs minions, skillstate, combat events | 11-21 | dlv-reads | in flight | none | V two commits on main; fork branch named in the card | live (U) | confirm | coordinator | scenario per assertion kind |
| DLV1 | upkeep (state, minions), mark, stance behindArmy | 11-21 | dlv-upkeep-state, dlv-upkeep-minions, dlv-mark, dlv-stance-army | not started | none | V cards Proposed, no commits | wave 0 | after wave 0 | coordinator | per card scenario |
| DLV2 | cast verbs, point and corpse aim, stance insidePack | 11-21 | dlv-cast-verbs, dlv-aim-point, dlv-aim-corpse, dlv-stance-pack | not started | none | V cards Proposed | wave 1; fidelity A5, A8; R6 survival judge before `insidePack` acts | after wave 1 | coordinator | per card scenario |
| DLV3 | press, Teleport driver, Charge and Leap (Whirlwind parked) | 11-21 | dlv-press, dlv-move-tele, dlv-move-melee, dlv-whirl | not started | none | V cards Proposed | wave 2; fidelity A6 for Whirlwind | after wave 2 | coordinator | per card scenario |
| DLVc | cleanup: delete skills.json, skillKinds, class switches | 11-21 | dlv-cleanup | not started | none | V card Proposed | every driver knob at 2 for the agreed number of nights | last | coordinator | CI literal check at 0 |

### Quest board state machine (card fableboard, ADR 0025; 10 cards)

| id | recommendation | review | card | STATE | knob (default) | evidence | blocker | next action | owner | enable criteria |
|---|---|---|---|---|---|---|---|---|---|---|
| BSM | stored runs and latches in the plan file (cards 1 to 7 of the review) | 11-21 | bsm1 to bsm9 | built-dark | `quests.boardSM` 0 | V `docs/knobs.md:548`; `Merge agent/bsm1` to `bsm9` on main (bsm7 card still reads Proposed); `docs/probes/boardsm-progress-2026-11-21.md`; ledger row | P1 counts; restart drill (U) | bsmdef | coordinator | Built but dark row 3 |
| BSMc | questspec rules 12 to 14 in the merge gate | 11-21 | bsmchk | not started | none | V no such rules in `internal/questspec` | none | build | coordinator | rules in CI |
| BSMa | the board as an arbiter intent source | 11-21 | arbmig (family quest) | not started | none | V no card existed; folded into P6 | P6 | with P6 | coordinator | see P6 |
| BSMd | default 1 after the scenario set is green; later remove old paths and dual-write | 11-21, ADR 0025 item 11 | bsmdef | not started | `quests.boardSM` 0 | V ledger row | P1 counts | build | coordinator | Built but dark row 3 |

### Town planner (card fabletown, ADR 0028; 11 cards)

| id | recommendation | review | card | STATE | knob (default) | evidence | blocker | next action | owner | enable criteria |
|---|---|---|---|---|---|---|---|---|---|---|
| TOWN1 | planner scenarios | 11-21 | town1 | enabled | none | V Done, `Merge agent/town1` | none | none | coordinator | n/a |
| TOWN2 | one counter (`supply.CountOf`) | 11-21 | town2 | enabled | none | V Done; `internal/supply/count.go`; at stage 0 it returns the old numbers | none | none | coordinator | n/a |
| TOWN3 | NeedFor, Solve, profiles, stage 1 shadow trace | 11-21 | town3 | built-dark | `town.plan` 0 | V `docs/knobs.md:144`; `Merge agent/town3` | flip (townflip) | see P5 | coordinator | stage 1: shadow lines present |
| TOWN4 | stage 2: guards, chunk preconditions and shop read the plan | 11-21 | town4 | built-dark | `town.plan` 0 | V `Merge agent/town4`; `docs/probes/town-plan-stage2-2026-11-21.md` | P1 counts | set stage 2 on sandboxes | coordinator | Built but dark row 2 |
| TOWN5 | stage 3: purse settled once per visit, recovery reads wealth | 11-21 | town5 | built-dark | `town.plan` 0 | V `a559c72fe`, `Merge agent/townstage3`; `docs/probes/town-plan-stage3-2026-11-22.md` (CHEATED setups); card still In Progress | stage 2 first | after stage 2 | coordinator | stage drill green, flip, 24 h monitor, auto-revert on a hard breach |
| TOWN6 | stage 4: sell, keep and pickup | 11-21 | town6 | not started | none | V card Proposed, no commit | stage 3 | build | coordinator | per stage look |
| TOWN7 | short-gold priority table as data | 11-21 | town7 | not started | none | V card Proposed | none | build | coordinator | table reviewed |
| TOWN8 | stage 5: route stops by nav cost | 11-21 | town8 | not started | none | V card Proposed | stage 4 | build | coordinator | per stage look |
| TOWN9 | viewer GET /api/townplan and a card | 11-21 | town9 | not started | none | V card Proposed | none | build | coordinator | endpoint answers |
| TOWN10 | town.plan as arbiter source and visitor | 11-21 | town10 | not started | none | V card Proposed | P6 | with P6 | coordinator | see P6 |
| TOWN11 | cleanup: delete stage 0 and retired knobs | 11-21 | town11 | not started | none | V card Proposed | stage 5 stable | last | coordinator | old counters deleted, knob rows removed |
| TOWN12 | SME review of the supply profile numbers | 11-21 | town12 | not started | none | V card Proposed | none | run the SME agent once | coordinator | advice recorded |

## Reviews of 2026-11-22 (map knowledge, UI)

| id | recommendation | review | card | STATE | knob (default) | evidence | blocker | next action | owner | enable criteria |
|---|---|---|---|---|---|---|---|---|---|---|
| MAP1 | mapverify: Countess three ways with the map knobs, Monastery Gate border strip memo | 11-22 | mapverify | enabled | none | V Done; 23 commits matching the card on main | none | none | coordinator | n/a |
| MAP2 | FullyKnown measured; a whole grid is planned on alone | 11-22 | knownedge | built-dark | `nav.fullGrid` 0 | V `docs/knobs.md:51`; ledger row | A/B nav drill (U) | drill | coordinator | no stuck regressions on 5 areas |
| MAP3 | read warp trigger areas from the engine fork | 11-22 | warptrigger | built-dark | `nav.warpRead` 0 | V `docs/knobs.md:52`; fork branch pushed per the card result | drill (U) | drill | coordinator | warp drill across 5 areas |
| MAP4 | engine link table against the bot's exit data | 11-22 | exitsread | enabled | none | V one commit on main; card result: verified on the lab d2d (CHEATED, link facts only) | fix card for the findings (U) | read the result | coordinator | n/a |
| UI | UI unification (navigation, sheets, toasts, build cards, pre-cockpit deletion E11) | 11-22 | uinav, uisheet, uitoast, uisettings, uifit1 to uifit4 and others | in flight | none | V most listed ui* cards Done in `backlog/tasks/`; uifit1 Proposed | E11 deletion after the Settings sheet | see `docs/design/fable-review-2026-11-22/e11-deletion-plan.md` | coordinator | per the plan |

## Other built-dark rows of the ledger (not from a Fable recommendation)

These are in `data/darklaunch.json` and owned by the ledger, not by this tracker: `resource.gearDecisions`, `resource.skillCapPerLevel`, `agency.autoAlloc`, `combat.missileDodge`, `combat.threatDodge`, `combat.standOK`, `combat.minionRead`, `quests.layoutRead`, `loot.vault`, `loot.rulesEngine`, `bot.portalGuard`, `bot.farmReset`, `bot.mercReviveUnverified`, `bot.cubeJobs`, `town.identifyTome`, `ladder.enabled`, `invdesk.verbs`, `invdesk.instant`, `ui.managementFast`, `watch.paletteFeed`, `input.planner`, `input.hotkeyMap`, `items.ladderRunewords`. The weekly `python3 -I tools/darklaunch/report.py` is their review; do not repeat them here.

## Order of work (the coordinator's reading)

1. `invmon` (P1, merged: read its counts) and `ticklog` (T4): no behaviour, they give every later flip its measure.
2. `ckptlive` (live drill, small) and `townflip` stage 2 on the sandboxes: both are already built.
3. `outcome1` (P2): the ground for P4, P6, P7, P9, P12.
4. `bsmdef` (boardSM flip) once P1 counts exist.
5. `incident1` (P8), `reflex1` (P3), `thrgov` (P10) beside them.
6. `arbmig` (P6) and `admit1` (P4) after P2; `stepexec1` (P9) chunk by chunk.
7. Everything else (dlv waves, classes, acts, input I3 to I7) follows the epic cards; the blockers named in the rows map to the epics `epic-act2` (Duriel needs the generic executor), `epic-class-necromancer` (dlv wave 1) and `epic-class-paladin` (dlv wave 2).

## How to keep this file true

- On each merge that touches a row's card: update STATE, evidence (commit or doc) and next action in the same PR as the coordinator's merge bookkeeping, not in the feature PR (the feature agent does not own this file).
- A row moves to `trial` when the knob is set on a sandbox; to `enabled` only when the default changes and the ledger row says `default-on`.
- A row with `built-dark` past its `review_by` appears in `report.py` as due; the decision is enable, extend (with a changed why) or retire.

## Cards created by this branch

Cards for work that has no other card. The pieces already carded by other agents are NOT recreated: P1 invmon is TASK-955, P2 is TASK-956 to 961, P3 is TASK-962 (agent/reflex1), P7 and P11 are TASK-965 (agent/explain1), P8 is TASK-971 (agent/incident1), the ledger is TASK-954. All status Proposed, size and order in each card; none ticks a box. Task ids are assigned on this branch and can shift at merge (the card id in the first column is the stable name).

| card | task | tracker row |
|---|---|---|
| admit1 | TASK-970 | P4 |
| townflip | TASK-972 | P5 |
| arbmig | TASK-973 | P6 (also BSMa, TOWN10 as sources) |
| stepexec1 | TASK-976 | P9 |
| thrgov | TASK-977 | P10 |
| faultsoak | TASK-979 | P12, T5 |
| proptest1 | TASK-980 | T3 |
| ticklog | TASK-981 | T4 |
| bsmchk | TASK-982 | BSMc |
| bsmdef | TASK-983 | BSMd, BSM |
| agencyp1b | TASK-984 | R7 |
| agencyp3 | TASK-985 | R15 |
| advisor1 | TASK-991 | R14 |
| inputl3 | TASK-986 | R17 (I3) |
| inputl4 | TASK-988 | R17 (I4) |
| inputl5 | TASK-989 | R17 (I5 to I7) |
| autorevert1 | TASK-994 | standard enable rule (flip, 24 h watch, auto-revert) |
| ckptlive | TASK-990 | checkpoint row of Built but dark |

Existing cards that already cover a recommendation (dlv-*, pg*, mg*, town1 to town12, bsm1 to bsm9, epic-act*, epic-class*) are not recreated; only the epic blockers that map to a row are linked in the rows above.
