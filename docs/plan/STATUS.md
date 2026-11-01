# Where we are (hand-written part last edited 2026-10-30 evening by REVIEWDOC; the generated block at the bottom is current)

One page for the user. Decisions: [../decisions.md](../decisions.md) and [../adr/](../adr/README.md); plan: [plan-v2.md](plan-v2.md); live history: [../g1-log.md](../g1-log.md); work queue: [../BACKLOG.md](../BACKLOG.md); doc map: [../INDEX.md](../INDEX.md). Counts that change (merges, attempts) are not copied here: `git log`, g1-log and the generated block own them. Tags: V = verified (how, in the linked doc), V-d2d = measured on the reimplementation, U = unverified.

## Headline
- **d2d is the primary backend** ([ADR 0009](../adr/0009-d2d-primary-backend.md), grill 0001): v1 product is the d2d container, Game.exe is frozen, stays the acceptance reference and retires at the d2d container's 24 h soak; leaving d2d is break glass only. Only the four gated metrics of [fidelity-gate](../../openspec/specs/fidelity-gate/spec.md) must match the original within 20 %.
- **Game.exe soaks are done and judged** (g1-log): G1 attempt 18 PASS (2 h, 0 deaths, 9.1k XP/h); G2 8 h on Soaky (clvl 14 -> 15, 4 deaths, 7.1k kills, 4.1k XP/h, kill injection PASS with the user's one-off permission, criteria 6/7 closed); Fresha 4 h fresh Sorceress (clvl 6 -> 11, 3 deaths, 13.4k XP/h). Both found the same bug class (healing potions in the belt, none drunk; portal cast starvation; vendor dialog ignoring clicks): fixed offline by potionbug, portalfail, npctrade, manabackoff, rejuvuse, equipconfirm (live effect U, Game.exe frozen).
- **The whole bot runs on the lockstep fork** (`d2d.lab5`, 22-33x real time, V-d2d): navigation, combat, loot, town (Akara, stash, repair, portal), merc behind a knob. Replicated soak 4 seeds x 2 virtual h: XP/h 3013 then 3869 after ECONOMY (baseline 6600: -54 % then -41 %), kills/h 132 -> 167 (baseline 500), deaths 0 (the bot chickens out instead). **The fidelity gate is NOT met**; verdict INCONCLUSIVE below the sample minimums ([../d2d-loop.md](../d2d-loop.md), [../d2d-economy.md](../d2d-economy.md)).
- **The gap is mostly bot-side, not d2d** (research, V): seconds per kill 23.6 on d2d vs 9.1 original, of which search/stall 5.5 vs 0.3 and town/recovery 5.2 vs 3.0 ([research/per-kill-time-split.md](../research/per-kill-time-split.md)); casts per kill was a probe artifact; XP per monster identical clvl 1-7 ([research/xp-award-trace.md](../research/xp-award-trace.md)); monster placement equals game.exe per room ([research/spawn-density-trace.md](../research/spawn-density-trace.md)). One real d2d gap: spell damage truncated to whole points (-17 % at skill 1), PR branch built, not pushed (prmeasure measures first).

## Merged today (by area; details in the archived cards, `openspec/changes/archive/2026-10-30-*`)
- d2d stack: adapter, control client, pacer/virtual clock, world source, combat, town, loot, exits, roam, economy, speed (Prefetch, immediate verbs), merc wire, tap viewer (`cmd/d2dview`, `/stream`, sanitized `/api/pub`), replicated-soak tool (`cmd/d2dsoak`, gate v2 with exact Poisson bounds), per-kill time split, 4-slot lab pool (`tools/d2dslot.sh`), floating the worker host capacity rule (`tools/capacity.sh`), a Windows worker as a WSL2 d2d worker.
- Game.exe bot (frozen, acceptance): route/tower guards, death autopsy, recovery XP, early-game tables, inventory/stash, waypoints, nav stuck memory, potion and portal fixes, NPC visit/hover/skip/trade, two soaks at once, minute-line proxy fields (`dmgTaken`, `hpMinPct`).
- Upstream (psi29a/d2decomped): PR #127 prices posted, PR #128 cast fixes and replies posted (both user-approved); cast trace, XP trace, spawn trace, upstream gap items and three PR drafts (nothing else posted); MPQ reading method; Redivivus, D2MOO, landscape, whisper/LAN-oracle research.
- Process: change cards + 7 specs + `changecheck`, ADRs 0001-0009, grill workflow and agent, needs-you panel, board `/changes`, forge runners on the NAS/the worker host/a Windows worker/Mac worker A (Mac worker B offline), CI on master/PR, pre-push hook, public page with live soak status and a private runs cache, plugin pin.

## Open now
- Cards: d2d-container-build, d2d-container-bot, d2d-container-soak (the retirement soak), d2d-default-backend, d2d-replicated-soak, d2d-lan-oracle-demo, stream-grid; deferred: gameexe-retire-soak, orig-decision-probes. Running agents: killfix, forkrebase, prmeasure, decompmon, decomprec, containerbuild, stridegrid, lanoracle, licensedecide.
- Unmerged, reviewed in [research/code-review-2026-10-30.md](../research/code-review-2026-10-30.md): agent/killgap (per-kill time buckets), agent/detloop (determinism: the guard runs on the stepping goroutine; the full bot loop was not deterministic before it), agent/replay (record/replay of d2d sessions; two High findings to fix before or with the merge).
- Review findings to card (same doc, section 9): d2dsoak exit codes and empty/crashed seeds counted as clean exposure (H), replay tap cannot be stopped with a client attached and panics on crafted files (H), d2dview watch speed can stall the bot (M), public page: character names and a stale cache shown as live (M).
- Not started: Act 2+ on either backend, Andariel, the free Rogue claim on Game.exe, the quit-to-town shortcut, the idle-game layer (parked by the user).

## Decisions awaiting you
1. Licence and GPL review before any d2d container ships (ADR 0003; agent licensedecide is drafting the options; the fork stays in private the forge `<owner>/d2d-fork`).
2. Which character names may appear on the public page (today the sanitizer blocks machines and people, not the agents' characters; Soaky/Fresha/Loopy are shown on purpose).
3. Grill 0003 (merc and Act 2) when the coordinator asks; the SME contradicts the plan on ROUTEGUARD clvl 20 (suggests 14) and the Andariel floor (needs-you `grill-0003`).
4. Upstream follow-ups: push the missile-damage PR after prmeasure's numbers; the cast follow-up PRs wait for the maintainer's response to #127/#128 (needs-you `pr-cast-followups`).
5. Stream and ports: the Kick stream stays off; the d2dview `/stream` URL, its deploy and the container viewer port are yours ([../streamctl-d2d.md](../streamctl-d2d.md)).

## Top risks
1. The fidelity gate (-41 % XP/h after ECONOMY) is a bot-on-d2d problem (search/stall, poverty recovery, chicken restarts) and the d2d-replicated-soak card inherits it; a d2dsoak verdict can also exit 0 on INCONCLUSIVE/FAIL today (review H1).
2. Determinism of the full bot loop depends on the unmerged detloop change; replay and the determinism finder are built on it.
3. The container is unbuilt (build, bot, soak cards all proposed); licence review blocks shipping it; Windows/WSL2 (a Windows worker) runs d2d but not the container.
4. The public page is PUBLIC: the sanitizer is fail-closed on hosts, IPs, paths and tokens, but names and the stale-cache "live" label are the gaps above.
5. Game.exe is frozen with known live bugs fixed only offline (potions not drunk, trade dialog, `restarts` counter counts 1 of 17 recycles); it is still the acceptance reference for the 20 % gate.

<!-- statusgen:begin (generated by `go run ./cmd/changecheck status -write`; do not edit by hand) -->
### Active changes
| change | status | agent | owns | exclusive |
|---|---|---|---|---|
| [d2d-container-soak](../../openspec/changes/d2d-container-soak/) | proposed | d2d-container-soak | `docs/d2d-container-soak.md` | `soak-slot:1` |
| [d2d-default-backend](../../openspec/changes/d2d-default-backend/) | proposed | d2d-default-backend | `cmd/d2bot/d2dbackend_default_test.go` | - |
| [d2d-replicated-soak](../../openspec/changes/d2d-replicated-soak/) | proposed | d2d-replicated-soak | `docs/d2d-replicated-soak.md` | `soak-slot:2` |
| [drills6](../../openspec/changes/drills6/) | in-progress | drills6 | `openspec/changes/drills6/**` | - |
| [fixbuff](../../openspec/changes/fixbuff/) | in-progress | fixbuff | `docs/fixbuff.md`<br>`internal/combat/fight.go`<br>`internal/data/extract.go`<br>`internal/combat/fixbuff_test.go`<br>`internal/d2/states.go`<br>`internal/d2dctl/verbs.go`<br>`internal/d2dctl/parse.go`<br>`internal/d2dctl/parse_test.go`<br>`internal/d2dsrc/source.go`<br>`internal/d2dsrc/source_test.go`<br>`internal/d2denv/env.go` | - |
| [fixcastgap](../../openspec/changes/fixcastgap/) | in-progress | fixcastgap | `docs/fixcastgap.md`<br>`internal/combat/castgap.go`<br>`internal/combat/castgap_test.go`<br>`docs/knobs.md` | - |
| [fixmouth](../../openspec/changes/fixmouth/) | in-progress | fixmouth | `internal/nav/walk.go`<br>`internal/nav/nav.go`<br>`internal/nav/fixmouth_test.go`<br>`internal/combat/combatspin_test.go`<br>`docs/fixmouth.md` | - |
| [m1-liveproof](../../openspec/changes/m1-liveproof/) | in-progress | drills | `docs/basics-audit.md`<br>`docs/basics-live.md`<br>`docs/primitives.md`<br>`docs/probes/basics/**`<br>`cmd/d2drill/**`<br>`internal/drill/**`<br>`docs/probes/drills/**` | `soak-slot:0` |
| [m3-merc](../../openspec/changes/m3-merc/) | proposed | m3-merc | `internal/merc/**`<br>`internal/bot/merc_test.go`<br>`docs/merc-d2d.md`<br>`docs/probes/merc/**` | `soak-slot:3` |
| [mapitem](../../openspec/changes/mapitem/) | in-progress | mapitem | `docs/stream/map-contiguity.md`<br>`cmd/d2dview/stitch.go`<br>`cmd/d2dview/stitch_test.go`<br>`cmd/d2dview/feed.go`<br>`cmd/d2dview/pub.go`<br>`cmd/d2dview/main.go`<br>`cmd/d2dview/view_test.go`<br>`cmd/d2dview/web/stream.html`<br>`cmd/d2dview/web/index.html` | - |
| [skmiss](../../openspec/changes/skmiss/) | in-progress | skmiss | `docs/research/sk-missile-trace.md`<br>`tools/emu/sk_missile_hit.py`<br>`tools/emu/sk_missile_hit_test.cpp`<br>`docs/decomp-queue.md`<br>`docs/BACKLOG.md`<br>`docs/INDEX.md`<br>`openspec/changes/skmiss/**` | `fork-branch:fix/missile-roll-top` |
| [soak-minutes-virtual](../../openspec/changes/soak-minutes-virtual/) | in-progress | specmin | `openspec/changes/soak-minutes-virtual/**`<br>`internal/d2dsoakstat/invariants.go`<br>`internal/d2dsoakstat/invariants_test.go`<br>`internal/d2dsoakstat/invreport.go`<br>`internal/d2dsoakstat/minute.go`<br>`internal/d2dsoakstat/testdata/invreport.golden`<br>`docs/probes/postfix-soak2.md`<br>`docs/review-strategy.md`<br>`docs/d2d-soak.md`<br>`docs/geninv.md` | - |

### Gates
- spec [backend-contract](../../openspec/specs/backend-contract/spec.md): 34 requirements
- spec [bot-loop](../../openspec/specs/bot-loop/spec.md): 6 requirements
- spec [combat-loop](../../openspec/specs/combat-loop/spec.md): 9 requirements
- spec [d2d-farm](../../openspec/specs/d2d-farm/spec.md): 8 requirements
- spec [dashboard-pub](../../openspec/specs/dashboard-pub/spec.md): 3 requirements
- spec [death-rule](../../openspec/specs/death-rule/spec.md): 8 requirements
- spec [early-game-policy](../../openspec/specs/early-game-policy/spec.md): 17 requirements
- spec [fidelity-gate](../../openspec/specs/fidelity-gate/spec.md): 12 requirements
- spec [lab-rules](../../openspec/specs/lab-rules/spec.md): 9 requirements
- spec [licence-policy](../../openspec/specs/licence-policy/spec.md): 7 requirements
- spec [merc-policy](../../openspec/specs/merc-policy/spec.md): 7 requirements
- spec [process-hygiene](../../openspec/specs/process-hygiene/spec.md): 9 requirements
- spec [public-split](../../openspec/specs/public-split/spec.md): 10 requirements
- spec [replay](../../openspec/specs/replay/spec.md): 7 requirements
- spec [soak-gates](../../openspec/specs/soak-gates/spec.md): 19 requirements
- spec [town-basics](../../openspec/specs/town-basics/spec.md): 17 requirements
- spec [upstream-pr](../../openspec/specs/upstream-pr/spec.md): 14 requirements
- change d2d-container-soak: 24 h soak report.txt says PASS with exit 0
- change d2d-container-soak: memory curve and periodic d2d restart interval recorded
- change d2d-default-backend: go test ./cmd/d2bot -run Backend passes
- change d2d-default-backend: bot started with no flag uses d2d
- change d2d-replicated-soak: d2dsoak verdict printed for all four gated metrics, 5 or more seeds
- change d2d-replicated-soak: gaps listed as explained or as fidelity bugs
- change drills6: changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; whole-repo vet and test -short on the worker host
- change fixbuff: go test -p 1 ./internal/combat ./internal/d2 ./internal/d2dctl ./internal/d2dsrc ./internal/d2denv passes including TestFightBuffRecastsInLongFight and TestBuffStates
- change fixbuff: go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes
- change fixcastgap: go test -p 1 ./internal/combat ./internal/env ./tools/knobdump passes including TestFightLoopLightningNeverReissuedEarly
- change fixcastgap: go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes
- change fixmouth: live: DW.m3 x3 and DW.m4 rerun, no bounce (docs/fixmouth.md)
- change m1-liveproof: docs/primitives.md: every primitive with criticality S or A has a drill verdict PASS (or FAIL with a named fix card) in its evidence column, checked by the coordinator by eye
- change m1-liveproof: docs/basics-audit.md coverage table has 0 G rows and every S1 or S2 row is C with a test name in the offline-test column and a live proof line (`grep -c '| C ' docs/basics-audit.md` is at least 15)
- change m1-liveproof: starting gold and unidentified sell prices recorded V in docs/basics-live.md with cheat ledger ids
- change m1-liveproof: tools/agentreap.sh m1-liveproof leaves nothing; changecheck -cli 0 warnings; leakcheck 0
- change m3-merc: 2 seeds x 4 virtual hours from clvl 9 after Blood Raven on hidden d2d: the Rogue claimed exactly once per seed, 0 paid Act 1 hires, merc alive at least 70 percent of field time, every revive logged with price and gold delta within the revive rule (raw in docs/probes/merc)
- change m3-merc: go test -p 1 ./internal/merc ./internal/bot -run 'Merc|Claim' passes; go test ./internal/env -run EveryKnob passes with the d2d default
- change m3-merc: go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes; tools/agentreap.sh m3-merc leaves nothing
- change mapitem: go test -p 1 ./cmd/d2dview passes including TestStitchTransitionBloodMoorColdPlains, TestStitchFieldAndTown, TestStitchFieldToCaveIsNotStitched
- change mapitem: go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes
- change skmiss: differential on real d2d builds: before > 0, after 0 for any claim; pr_guard 0/0; presubmit --quick 0 FAIL; upstream tests rc 0
- change skmiss: changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict
- change soak-minutes-virtual: go test -p 1 ./internal/d2dsoakstat ./cmd/d2dsoak passes, incl. TestInvariantsGate1PostfixSoak2 (27 and 26 fail) and the clean synthetic log (passes)
- change soak-minutes-virtual: tools/openspec.sh validate --all --strict passes; changecheck -cli 0 warnings; leakcheck -mode fail 0 hits

### Recent merges
- 4b86423 11-01 01:52 `Merge agent/fixcastgap: per-skill cast gap (Lightning 890 ms on game.exe, 650 ms on d2d)` (fixcastgap)
- c53dc1d 11-01 01:48 `Merge agent/drills15: DRILLS15 live drill results (13 drills on d2d vs game.exe traces)` (drills15)
- ea3e136 11-01 01:41 `Merge agent/sklight: SKLIGHT trace of Lightning/Chain Lightning and piercing re-hit` (sklight)
- d83ceb0 11-01 01:36 `Merge agent/fixmouth: mouthArrivalR for the first walk after a warp arrival (den bounce fix A)` (fixmouth)
- 27049dd 11-01 01:27 `Merge agent/fixbuff: re-cast buffs in long fights, buff states, stale skills cache check` (fixbuff)
- 8e783b1 11-01 01:21 `Archive fixrecast (1a4b228); OWNERS; STATUS` (fixrecast)
- 0c88fdc 11-01 01:16 `Merge agent/crfix157: CodeRabbit follow-up records for upstream PR 157 (docs only)` (crfix157)
- 89d6a22 11-01 01:13 `Merge agent/drills13: DRILLS13 live results (Frozen Armor recast pass, mana timing M1)` (drills13)
- 3f62842 11-01 01:11 `Merge agent/fixdenbounce: soak 2 den bounce evidence record (docs only)` (fixdenbounce)
- 598d80e 11-01 01:06 `Merge agent/skarmor: sk-armor trace (decomp row 15), game-side oracle sk_armor.py` (skarmor)
<!-- statusgen:end -->
