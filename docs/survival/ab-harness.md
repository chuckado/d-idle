# Survival A/B harness, S-D readiness audit and drill R3 plan (card survab, 2026-11-16)

> status: current (TASK-550). Offline: nothing here ran against a live game; the report tool ran on synthetic logs only. Tags: V verified (how), S secondary (a doc), I inferred, U unverified. Gate source: [design](design-2026-11-14.md) section 10 (S-D) and [decisions](decisions-2026-11-14.md) answers 10; knob defaults are the user's later answer "default on right away", the A/B stays the gate.

## 0. Decision memo (what the user must approve)

**Ask: a go to run the A/B on the worker host, in lockstep d2d slots, about 12 virtual hours (3 seeds x 2 h x 2 arms) = roughly 2 slot-hours of wall time (about 1 core each, niced), not 12 container-hours.** Nothing touches the play sandbox container or its data volume.

| Item | Proposal |
|---|---|
| Where | the lab slot pool of the worker host (`tools/d2dslot.sh`, `tools/ab/` family), same as the merc A/B. Lockstep runs about 6.7 times real time (V `docs/ab-merc.md`, 0.5 h in 4.5 min under load 12 to 21). 12 virtual hours = about 107 min on one slot, about 55 min on two, about 30 min on four. RAM per slot not measured here (U); the d2d container caps at 3 GB and 2 CPUs (S `docs/d2d-container/README.md`). |
| Heroes | fresh throwaway heroes (names `SaHero-<seed>`) made by `tools/ab/prep`: a declared SETUP write (level 12, 500 gold, free Rogue), then loaded by a d2d without lab verbs, so no write is possible after the start (V by design, `docs/ab-design.md` Variants). A clvl 1 start reaches only clvl 8 in 2 h and never visits the zones where the two arms differ, so the late start is the only start that tests the margin. Class: Paladin first (the sandbox hero), Sorceress second (answer 8). Both are tagged CHEATED for provenance (`docs/probes/test-characters.md`): the result is a relative A/B, not evidence of game behaviour. |
| Arms | `legacy`: `bot.zoneModel` 0, `margin`: `bot.zoneModel` 1; everything else identical, `bot.portalGuard` 0 in both (until R3). One binary. |
| BLOCKER | `bot.zoneModel` does not exist on main (section 1). A card must build slice S-D first (Plan.Pick asks the controller). `abrun.sh` refuses to start when the shipped binary lacks the knob, so two legacy arms can never be mistaken for an A/B. |
| Gate wording needs one decision | "No more deaths and chickens" is a coin flip at 3 seeds: at about 1 death per arm in 6 virtual hours, two equal engines fail the strict reading 35 percent of the time (Poisson, V computed: P(X<=Y) = 0.654). Pick one: (a) strict as worded and accept re-runs; (b) a slack of one death and one chicken pooled (`abreport -death-slack 1 -chicken-slack 1`); (c) 8 seeds, which cost 32 virtual hours = about 5 slot-hours (the standard of `docs/ab-design.md`, which also says 3 seeds cannot show an XP effect under about 30 percent: paired SD of 13 to 20 percent of the mean). Recommendation: run 3 seeds first as a regression screen (cheap, catches a gross failure), judge with (b), and run 8 seeds only if the screen is inconclusive or before the default is flipped for friends. |
| XP rule | "at least equal at 120 min": strict ratio >= 1.0 on 3 seeds is within noise, so `abreport` only FAILs on XP when the whole 95 percent interval is below the floor, and otherwise says INCONCLUSIVE. Say if you want a tolerance (`-xp-floor 0.95`). |
| Risks | the margin arm is a different run per seed (low pairing, `docs/ab-design.md`: correlation near 0 when the first field changes), so the paired SD is large; lockstep polls once per tick, so the intake (lower bound) reads higher there than in the realtime sandbox; a lockstep result is d2d-only (no game.exe); the legacy arm may itself die zero times in 6 h, which makes "no more deaths" trivially true for it. |
| Cost of the alternative | two realtime containers on fresh volumes on the worker host: 3 seeds x 2 h x 2 arms = 12 container-hours of wall time, 2 CPUs and 3 GB reserved per container (S), 4 in parallel = 3 h wall and 8 CPUs; realtime also has the sandbox's own poll rate, so it is the closer test of what the user sees. Recommend it only as a confirmation after the lockstep screen passes, one seed per arm. |

Approve: (1) the lockstep screen above, (2) the gate reading (a, b or c), (3) the class order, (4) a card for S-D itself (`bot.zoneModel`, S sized; it is the prerequisite). R3 (section 5) is a separate, smaller go: one slot, about 20 minutes, a lab binary with verbs.

## 1. Audit: what exists on main for S-A to S-D

| Slice | Status | Where (V: read on this tree; tests named in the docs, not re-run here) |
|---|---|---|
| S-A danger profiles | DONE | `internal/danger/` (generate.go, model.go, params.go, bosses.go + bosses.json, areaunits.json, overrides.go + overrides.json with the evidence check, gap.go, unblockable.json, http.go for `/api/danger`), `docs/survival/danger.md`, `real_test.go` anchors. Behaviour change: none. |
| S-B power estimate, margin, shadow | DONE | `internal/power/` (archetype.go, hero.go, margin.go, profiles.json, skilldamage.json, zone.go), `internal/bot/zoneshadow.go` (`zoneShadow` logs next to every `planner` line), `cmd/d2bot/survivalwire.go`, `tools/levelsim` (`-shadow`, `-controller`), `docs/survival/power.md`, calibration `calibration-2026-11-14.md` and `units-fit-2026-11-15.md`. Open: the replay pass rule still FAILS 6 of 14 bad stays (calibration doc), defence/block/spell damage reads U, `k_hero` capped at 2 while the weapon is unread. |
| S-C ledger and shadow controller | DONE | `internal/survival/` (ledger.go incl. `IntakePerMin`, controller.go `Choose`, revive.go, economy.go, glance.go, why.go, params.json all U), `internal/watch/survivalapi.go` (`/api/survival`, words only), `ledger.json` per instance. Observed intake (`lifeLost`, `hitsTaken` on the minute line, `area summary lifeLost`, glance `intake`) is merged. The sandbox hour in shadow and the live check of `ledger.json` are listed in `ledger.md` as not done (U). |
| S-D switch zone choice | BUILT (card survsd, [zonemodel.md](zonemodel.md)); the row below is the audit before it |  `grep -rn zoneModel` over `*.go` and `docs/knobs.md` finds nothing. `Deps.ZoneChoice` (`internal/bot/runner.go`) only returns log lines (`zoneShadow` splits them into `r.e.Log.Info`); `Plan.Pick` (`internal/bot/plan.go`) never asks the controller. Needed: knob `bot.zoneModel` 0..1 in `internal/env/tuning_bot.go`, a `ZoneModelPick` hook used by `Plan.Pick` with the plan tables as candidates and fallback, empty candidate set falls back to the plan, knobs.md regenerated, tests on two builds. |
| S-E, S-F | not started | by design after S-D. |
| Portal guard | knob built, default 0 | `bot.portalGuard` (`internal/resource/castguard.go`, `internal/combat/safety.go`); flip waits for drill R3 (section 5). |

## 2. The A/B harness: design

Options considered:

| Option | Fresh hero | Same seed in both arms | Touches the user's hero | Cost for 12 h | Verdict |
|---|---|---|---|---|---|
| A. Lockstep d2d slots (`tools/d2dslot.sh`, as `tools/ab/mercrun.sh`) | yes (prep) | yes, map seed plus lockstep determinism | no | about 2 slot-hours | **chosen** |
| B. Second realtime sandbox container, own volume | yes (reset path of `docs/dev-reset.md`, but on a NEW container and volume) | no, real time is not deterministic | no if the volume is separate | 12 container-hours | confirmation only |
| C. The play sandbox itself | no (it is the user's hero) | no | yes | n/a | rejected |
| D. Replay (`levelsim -controller`) | n/a | n/a | no | seconds | cannot say whether following the shadow changes deaths; stays as the unit-level check |

Why A: the repo already has the paired protocol (`docs/ab-design.md`: one binary, one knob, same seeds, arms interleaved so host load drift hits both) and a late-start variant with a declared SETUP and the CRUTCH rule enforced by the d2d (no lab verbs after the start). Lockstep is about 7 times faster, deterministic per seed, and needs no container build.

Files (new, all untested live):
- `tools/survival/abrun.sh ARM SEED HOURS [SLOT]`: one run; writes `$W/ab/<arm>/<seed>.log` and the line `writes after start: N`; the only arm difference is `bot.zoneModel` in the instance `tuning.json` (the same override file `mercrun.sh` writes). Refuses a binary without the knob.
- `tools/survival/ab.sh HOURS SEED...`: arms per seed started back to back, at most `AB_MAX` (2) runs at once.
- `tools/survival/abreport DIR`: the report; `go run ./tools/survival/abreport -base legacy -arm margin [-horizon 120 -xp-floor 1.0 -min-seeds 3 -death-slack N -chicken-slack N] DIR`; exit 0 PASS, 1 FAIL, 3 INCONCLUSIVE, 2 bad input.

Run on the worker host (after the S-D card merged and `d2bot.survab` built from it):
```
tools/agentrun.sh survab --ttl 14400 -- tools/survival/ab.sh 2 101 103 107     # 3 seeds, 2 virtual hours, class via AB_CLASS=2
go run ./tools/survival/abreport ~/d2d-lab/survab/ab                            # then AB_CLASS=4 for the Sorceress
```
Before it: `labrun <worker> -- top -bn2 -d1` (idle + nice at least 25 percent), slots 0 to 3 free of other gates, `tools/agentreap.sh survab` afterwards.

Report metrics (per seed and arm, plus the pooled verdict): deaths (counter delta), chickens (`safety decision action=chicken`), merc deaths (`merc: died`), potions used, gold change (pack plus stash), clvl start and end, XP gained at 30, 60, 90 and 120 virtual minutes (indexed by minute line, never by the wall clock), intake per zone (`lifeLost` per minute of the minute lines, grouped by the area of the minute; a lower bound, `ledger.md`), paired XP difference at the horizon with its 95 percent t interval. A seed that is missing or short (under 95 percent of the horizon) in either arm is dropped from both. Tests: `go test -p 1 ./tools/survival/...` on synthetic minute logs (metrics of one run, seven gate verdicts, strict versus tolerant XP floor, dropped seeds, an old log without intake, the t interval, death slack).

Gate rule as implemented: FAIL when pooled deaths or chickens of the margin arm exceed the base (plus slack) or the XP interval is wholly below the floor; PASS only with no failure, at least `-min-seeds` pairs and an XP ratio at or above the floor; otherwise INCONCLUSIVE. Not covered (U): per-zone kill rate, time in town, the `zone choice:` agreement with the planner (grep the log), merc-alive share (`tools/ab/mercan.py` has it).

## 3. Runtime and resource use (summary)

Lockstep, 3 seeds x 2 h x 2 arms: 12 virtual hours, about 107 slot-minutes plus about 1 minute of prep per run (S/V from `docs/ab-merc.md`), about 1 core per slot at nice 10. On four slots about 30 minutes, on two about an hour. 8 seeds: 32 virtual hours, about 5 slot-hours. Disk per run: a log of a few MB and a d2d home (U, small; `data-<arm><seed>` is removed by the next run of the same name only, clean by hand). Memory per slot: not measured (U).

## 4. Flip rule for `bot.zoneModel`

Default on only after: the S-D card is merged with the knob at 0, the lockstep screen is PASS under the gate reading the user chose, no run was CHEATED (every log ends `writes after start: 0`), and the user's go. The user's answer "default on right away" then is one line in the tuning default plus the contract and knobs.md in the same branch. A FAIL keeps the default at 0 and the report names the zone and the minute (intake table, `zone choice:` lines).

## 5. Drill R3: portal exposure and the `bot.portalGuard` flip

Purpose: measure the time a Town Portal read takes and the damage taken during it, so `bot.portalCastMs`, `bot.portalWalkMs`, `bot.portalHitMs` become measured and `bot.portalGuard` can default to 1 (`docs/fixportalmelee.md` has the rule; decomp28 section 8 has R3).

Safe on a throwaway hero: a new character (name `R3-<n>`, Paladin; repeat with the Sorceress if time), one lab slot, the lab binary WITH `--lab-verbs`, writes declared SETUP or DIAGNOSIS through the ledger (`d2cheat`, reason given), nothing counted toward any gate, hero logged in `docs/probes/test-characters.md` as CHEATED. No game instance of the users; the lab d2d only. Needs the user's go because it uses the lab verbs.

Steps (per repetition, 20 each at 1 and at 5 zombies adjacent):
1. Town scroll in the pack, life full, hero in the field (not town), no other monster near. Log the start (virtual ms).
2. `debug spawn` the pack (1 or 5 zombies) adjacent to the hero; wait until they attack.
3. Read the scroll with the real inventory click (the bot's `d2input` path, not a verb). The zombies' own hits are the damage; the recipe of decomp28 also names `debug hit 20` during the cast as a fixed-damage control, so run 10 of the 20 repetitions of each group with it (U: whether it is needed once the zombies' own damage is logged).
4. Record: seconds from the click to the portal object (cast), seconds from the portal object to the area change (walk), life at each step, whether the cast was interrupted (expected: no, mode 10 held, V run decomp28 Q5), scroll count (spent once).
5. Return to town through the portal, restore life (SETUP), next repetition.

What to measure and the output table: per repetition `cast_s`, `walk_s`, `life_start`, `life_end`, `lost`, `adjacent`, `interrupted`; per group of 20 the mean, p95 and max of `cast_s + walk_s` and the p95 of `lost`.

Flip rule (from `docs/fixportalmelee.md`, unchanged): E = p95 of cast+walk. Set `bot.portalCastMs` and `bot.portalWalkMs` to the measured p95 parts, and set the default `bot.portalGuard` to 1 only if E is at most 6 s and the measured life lost with 5 zombies adjacent is at most the guard's prediction for the same pack; otherwise lower `bot.portalHitMs` until the prediction covers the measurement, and if it still cannot, leave the default 0. Exit criteria: 40 runs logged, none interrupted or the interruption explained, the three knob values and the decision written into `docs/fixportalmelee.md` by the card that owns it (this card does not edit it).

Checklist for the live agent (not run here): book a lab slot (`tools/d2dslot.sh`), start the lab d2d with `--lab-verbs`, create `R3-1`, fill the pack with scrolls and a belt, run the 40 repetitions through a small drill in `internal/drill` (not written here: the card that runs R3 owns it), copy the log to `docs/probes/`, stop the slot, `tools/agentreap.sh`, log the hero.
