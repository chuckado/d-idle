# Verifier: Crushing Blow r3 (2026-11-17)

> status: current, 2026-11-17 (card verifycrush3, TASK-628, Opus, singly; independent verifier before posting). Nothing posted. Subject: private fork branch `fid/crush-r3` (oracle `7476c90c`, claim `9bd2983f`, on upstream `bebfdbcf`), body `docs/upstream/bodies/crush-r3-body.md`, trace `docs/upstream/crush-r3-2026-11-17.md` (agent crush3). Checklist from `verifier-decomp35..39` and `verifier-gearabsorb3-2026-11-17.md`. The fixed `fid/crush-r4` is on the private fork remote only.

Tags: **V run** (run by this verifier today, the builder container), **V read** (disassembly or source read by this verifier), **I** (inferred), **S** (an earlier doc, named), **U** (unverified).

## Verdict: r3 FAIL (test gaps, RE finding missing, a wrong evidence row), r4 PASS WITH CHANGES

The rule is game.exe's and r3 implements it: 0 differ on three seeds (one fresh) in both oracles and on the hand cases at the edges (V run). r3 should not be posted as it is:

1. **ctest does not pin the rule (high).** 8 of 18 plausible mutants pass r3's `test_combat` (whole points of life, the mana leech cut before the crush, the resist not capped, a negative resist adding, the resist share by a plain multiply, rounding to nearest, the crush of the max life, a one-point floor). One of them (m11, a plain multiply instead of `mul_div`) is caught by NOTHING in r3: the oracle grid stops at 4,000 points of life, and the two part only on a crush past `0x100000` 256ths (a life above about 16,384 points at divisor 4: Hell act bosses). r4 pins game.exe's values for those cases; 18 of 18 abort.
2. **The RE finding is missing and contradicts the claim (medium).** Upstream's `docs/research/re/monsters.md` says the crush takes "current life (stat 6) / a divisor, after the hit", and `combat.md`'s approximation list still names the crush. The template's checklist asks for the finding in `docs/research/re/`; r3's body leaves that box out. r4 adds a docs commit.
3. **The divisor row overclaims (medium, claims).** The body says "a boss or unique 8 (FUN_0063e9f0, FUN_005a0180)". `FUN_005bffc0` passes mask 2 to `FUN_005a0180` (`lea edx, [eax+2]` at `0x5c0063`); bit 2 is the superunique bit (`FUN_005a03a0` reads the SuperUniques index only under it, V read). V run (`vcrush.py` below, `FUN_005a0180` native, not hooked): data word `+0x16` = 8 or 4 gives divisor 4, 2 or 0xa gives 8. So 8 is "a MonStats boss or a superunique". d2d's `Monster::crush_divisor` gives 8 to `Boss::unique` too (D5 below): a pre-existing deviation the PR does not touch; the body must not state "unique".

**Post `fid/crush-r4`**: oracle `7476c90c` (unchanged), claim `760df1fa`, docs `1da9632a`, on upstream main `bebfdbcf` (still upstream's head at this run, V run `git fetch`); author and committer `shatteredeagle <26069301+shatteredeagle@users.noreply.github.com>`, `Agent: Claude Code` the only trailer after a blank line in all three (V run, `git interpret-trailers --parse`). Against r3: the claim is r3's code byte for byte plus seven new `test_combat` asserts (game.exe values); the docs commit corrects `monsters.md` (and says the MonProp crush still reads the life after the hits) and drops the stale `combat.md` bullet. The body then needs C1 to C9 below.

## Game side, re-derived (V read by this verifier)

- **`FUN_005bffc0`** (`0x5bffc0..0x5c0127`): stat 136 of the attacker (`FUN_00625500`) at most 0 returns; `seed % 100` on the attacker's seed (`[esi+0x20]`, `[esi+0x24]`), not below the stat returns. Divisor: a type-0 defender (player) 10, straight to the missile test (`0x5c003c`: no player-count share); type 1: a hireling (`FUN_0063ee90`) 10, also with no share (`0x5c0055`); else 4, 8 if `FUN_0063e9f0(0, unit)` (MonStats boss) or `FUN_005a0180(unit, 2)` (superunique), then plus `MulDiv(divisor, FUN_005738f0(max(stat 0x64, 1)), 100)`; any other type 4. Doubled for event 6 (`0x5c00a2`). `crush = stat 6 / divisor` (`cdq; idiv`, 256ths), less `FUN_00483360(crush, min(stat 0x24, 100), 100)` when stat 0x24 is above 0; stat 6 = `life - crush`, 0 when not above 0 (record `+4 |= 2`); state 0x93 when crush > 0. No immunity, no ancient or boss cap, no difficulty term.
- **`FUN_00483360`** is not a plain `x * y / z`: above `0x100000` it divides first (`0x483376..0x48338c`); `rules::mul_div` ports it (V read, `rules.hpp:1675`). The claim uses it.
- **Ordering in `FUN_0057c6c0`**: the strike events (5 then 1 melee, 6 then 2 missile, `0x57c7b7..0x57c7f3`) before the cut of the physical to stat 6 and the leech (`0x57c7f8..0x57c826`), then the absorb heal (`0x57c844`) and the record total off stat 6 (`LAB_0057c852`) (S: `verifier-gearabsorb3-2026-11-17.md`, `verifier-decomp35-2026-11-16.md`; re-run here through `crush.py order`: 2,360 of 2,360 events see the case's life).
- **The get-hit** (`FUN_0057cb00`, `0x57cb22..0x57cb35`) reads the record total `+0x4c` (under `0x100`: no get-hit), which the crush never enters (it writes stat 6 directly). See D6.

## Scope answers (the brief's a to f)

- **(a) Who.** Every attacker with stat 136 runs the same handler; d2d calls `player_blow` for the player, missiles, the merc (`deaths.cpp:429`) and pets (`pets.cpp`), all through `Monster::target`, which passes `life_fine()` (V read, `ai.hpp:527`), so the 256ths reach play, not only the harness. Defender classes: the claim does not change `crush_divisor`; d2d's hireling 10 without share matches; its `Boss::unique` 8 does not (D5). No boss or champion immunity in game.exe and none in d2d. The player as defender (divisor 10) is the MonProp path, unchanged (D3).
- **(b) Ordering.** Base `bebfdbcf` holds #587 (leech) and #606 (gear absorb); upstream main has not moved, so no rebase. The gear absorb is the defender's (a player), and this path hits monsters, which have no absorb rows: no interaction. Crush, then the leech cut to what it left (pinned; m3, m4, m18 abort), then the hit: d2d adds the crush to the blow and `hurt` takes it with the hit, the same life at the end. Bone Armor's `soak` (`ai.cpp:2888`) takes at most the physical, as `FUN_005c6370` on the record `+8`; the crush is not in game.exe's record either. Life Tap heals from `blow.damage`, which includes the crush (as on the parent; U for game.exe).
- **(c) 256ths and floors.** Whole division in 256ths, a life under the divisor gives 0, the fraction carried in `blow.frac` (pins below; m2, m8, m12, m14 abort). No floor but 0: the crush stays under the life for any divisor of 4 or more.
- **(d) Missile.** The body says read, not run, for `FUN_0057c6c0`'s dispatch of event 6: honest, and it claims no measurement there. This verifier ran the handler's own event-6 branch natively (hand cases: divisor doubled, V run); the dispatch stays read. The body may say so (C5).
- **(e) `test_stand_attack`.** See Build.
- **(f) Checks.** Below.

## Differentials (V run)

Game side `crush.py game` / `pair` / `order` on the builder, seeds 7 and 5001 (the author's) and 9091 (fresh); d2d side `crush_harness` compiled against each tree.

| Tree | seed 7 crush / pair | seed 5001 | seed 9091 (fresh) |
|---|---|---|---|
| parent (oracle `7476c90c`) | 2,223 / 2,745, exit 1 | 2,243 / 2,746, exit 1 | 2,234 / 2,745, exit 1 |
| r3 claim `9bd2983f` | 0 / 0, exit 0 | 0 / 0, exit 0 | 0 / 0, exit 0 |
| r4 claim `760df1fa` (same `combat.hpp`) | 0 / 0 | | 0 / 0 |

`compare` and `compare-pair` exit 1 on an empty file, a 100-line file and a malformed line; `crush.py` with no argument or `gen` without N exits 1 (V run).

Hand cases. `crush.py game` / `pair` on case files, and `vcrush.py` (verifier-only: `FUN_005bffc0` native with `FUN_005a0180` NOT hooked and the monster data word `+0x16` set, or event 6); lives and crushes in 256ths:

| Case | game.exe crush | d2d r3 / r4 |
|---|---|---|
| 100.5 points, res 0 | 6,432 | 6,432 |
| 400 points, res 120 | 0 | 0 |
| 400 points, res -20 | 25,600 | 25,600 |
| 40,001 points, res 33 (`MulDiv` past `0x100000`) | 1,715,264 (a plain multiply: 1,715,243) | 1,715,264 (parent 1,714,432) |
| 3/256 of a point | 0 | 0 |
| 400 points, data flag 8 (unique) or 4 (champion) | 25,600 (divisor 4) | not in the claim (d2d `Boss::unique`: 8) |
| 400 points, flag 2 (superunique), 0xa, or MonStats boss | 12,800 (divisor 8) | |
| 400 points, res 50, event 6 / event 6 and boss | 6,400 / 3,200 | `test_combat` pins 3,200 |
| pair: Nightmare, 20 % mana steal, 100 points of physical on 100.5 points, divisor 8 | crush 3,216, mana 2,251, life 0 | equal |
| pair: Hell, 10 % / 10 %, 90 on 100, divisor 4 | 6,400, 639, 639 | equal |

## Mutants (V run, builder, local, never pushed)

Each one text change in r4's `combat.hpp`; per mutant `test_combat` (r4's, and r3's) and `crush_harness` compiled alone, the oracle on seeds 7 and 9091 (seed 9091 within a few percent of seed 7's counts, same zeros).

| Mutant | oracle seed 7 crush / pair | r4 `test_combat` | r3 `test_combat` |
|---|---|---|---|
| m1 crush from the life after the hit (parent) | 2,200 / 2,674 | abort | abort |
| m2 whole points of life | 2,079 / 2,427 | abort | **passes** |
| m3 life leech cut before the crush | 0 / 868 | abort | abort |
| m4 mana leech cut before the crush | 0 / 684 | abort | **passes** |
| m5 no resist share | 1,258 / 0 | abort | abort |
| m6 resist not capped at 100 | 40 / 0 | abort | **passes** |
| m7 negative resist adds | 142 / 0 | abort | **passes** |
| m8 crush fraction dropped | 2,111 / 2,427 | abort | abort |
| m9 divisor plus one | 2,247 / 2,805 | abort | abort |
| m10 missile keeps the divisor | 0 / 0 | abort | abort |
| m11 resist share by a plain multiply | **0 / 0** | abort | **passes** (caught by nothing in r3) |
| m12 crush rounded to nearest | 732 / 1,074 | abort | **passes** |
| m13 crush of the max life | 1,920 / 1,984 | abort | **passes** |
| m14 crush at least one point | 439 / 426 | abort | **passes** |
| m15 crush capped at the life the hit leaves | 612 / 1,101 | abort | abort |
| m16 resist share in whole points | 1,170 / 0 | abort | abort |
| m17 melee doubles the divisor too | 2,247 / 2,809 | abort | abort |
| m18 leech cut to the life the hit and the crush leave | 0 / 1,590 | abort | abort |

18 of 18 abort r4's `test_combat`; m10 and m11 only through it (the oracle runs melee and lives up to 4,000 points).

## Build, ctest, tidy (V run)

BUILD_PENDING

## Hygiene (V run)

- `pr_guard.py upstream/main HEAD` on r4 (all three commits): 0 FAIL, 0 WARN. `presubmit.py --quick --base upstream/main`: 0 FAIL, 3 WARN (siblings hit paths; build; gcc), answered by C5 and C6. `attribution` ok (no co-author lines).
- Identity as above. Diff r4 against upstream: 7 files (`combat.hpp`, `test_combat.cpp`, `crush.py`, `crush_harness.cpp`, `tools/emu/README.md`, `monsters.md`, `combat.md`); no private code, host names or paths (read).

## Deviations left (for the queue, not this PR)

- **D3 (U, from the trace):** the MonProp crush on the player, merc and pets (`ai.hpp:683`, applied after the hits) reads the life the hits left; game.exe's handler reads the life before. r4's `monsters.md` now says so.
- **D5 (V run mask, I class):** `Monster::crush_divisor` gives 8 to `Boss::unique`; game.exe's 8 is a MonStats boss or data bit 2. Upstream's own `drops.md` ("Possible disagreement, not established") leaves open whether a plain unique ever carries bit 2; `FUN_005a03a0` reads the SuperUniques index under bit 2 only (V read), so a plain unique almost surely does not (I, strong). Follow-up: the writer of bit 2, an oracle running `FUN_005a0180` natively (as `vcrush.py`), a one-line fix in `crush_divisor` and its comment.
- **D6 (V read both sides, not run):** game.exe's get-hit roll takes the record total (`+0x4c`, `0x57cb22..0x57cb35`) without the crush; d2d's `hurt` passes the blow's 256ths with the crush to `rules::get_hit` (`ai.cpp:2900`, `2915`). As on the parent; r3 makes the crush a little larger. Needs an oracle before a claim.

## Required changes to the body (`crush-r3-body.md`, becomes the r4 body)

- **C1 (high).** Commits: "Three commits on upstream `bebfdbcf`: the oracle (`7476c90c`), the claim (`760df1fa`), the RE finding (`1da9632a`)". Counted differentials: name the seeds run; this run's fresh seed is 9091 (2,234 / 2,745 on the parent, 0 / 0 on the claim).
- **C2 (high, claims).** Evidence row `0x5c0031..0x5c00a8`: "4; a MonStats boss (FUN_0063e9f0) or a superunique (FUN_005a0180, mask 2) 8; a hireling or a player defender 10, both without the player-count share". In play: "A MonStats boss or a superunique (divisor 8)". Limits: "d2d gives a plain unique divisor 8 (`Monster::crush_divisor`); game.exe's mask 2 is the superunique bit. Not changed here."
- **C3 (high, template).** Checklist: tick "The finding in `docs/research/re/`" (`monsters.md`, `combat.md`). Tests: name the new pins.
- **C4 (medium).** Mutants: this file's 18 rows (r4 column) instead of the 10-row table; "18 of 18 abort `test_combat`; m10 and m11 only through it". Drop "whole points, the mana leech cut, the uncapped resist and a negative resist adding pass `test_combat`".
- **C5 (medium).** Limits, the missile line: "the handler's own event-6 branch (divisor doubled) was run natively on hand cases; the dispatch of event 6 by `FUN_0057c6c0` is read, not run".
- **C6 (medium).** Build and tests: r4's numbers from this file; `test_stand_attack` as found below.
- **C7 (low).** "What changes": "`test_combat` now pins game.exe's crush for a life with a fraction, a resist above 100 and below 0, a crush past MulDiv's `0x100000` branch, a life under the divisor, and the mana leech after a crush".
- **C8 (low).** Re-check block: `crush.py game` on the pins' cases (`id life phys res div`): `0 25728 7680 0 4`, `1 102400 7680 120 4`, `2 102400 7680 -20 4`, `3 10240256 7680 33 4`, `4 3 7680 0 4` gives 6,432 / 0 / 25,600 / 1,715,264 / 0; `crush.py pair` on `0 1 0 20 25728 25600 8` gives 3,216 / 0 / 2,251.
- **C9 (low).** Limits: "The get-hit roll (FUN_0057cb00) takes the record total without the crush in game.exe; d2d's counts it (as before this PR)."

## For the coordinator (not body changes)

- The trace (`crush-r3-2026-11-17.md`): D1's "boss or unique" wording; the gap list of section 6 is closed by r4.
- Follow-up cards: D5 (crush divisor of a plain unique) and D6 (get-hit without the crush), each needing an oracle; D3 stays.
- `vcrush.py` (verifier scratch, in no branch) is in the lab dir `verifycrush3`; fold it into `crush.py` in the D5 PR.

## Not checked

- No Windows, macOS or Clang build; no live game, no lab drill; nothing on the worker host.
- The dispatch of event 6 by a missile, a player count above one (`FUN_005738f0`), the MonProp crush (D3), Life Tap's base in game.exe (U).
