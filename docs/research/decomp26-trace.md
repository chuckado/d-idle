# decomp26: a hit interrupts a swing (P04), a monster missile is blocked only with physical damage (P06), and the P09 / P10 / P03c notes

> status: finished 2026-11-15 (card decomp26, Sonnet). Offline: the emulator image on the worker host (game.exe's own functions) and the NAS builder for d2d; no game, no live d2d, nothing posted, no Go change. Tags: V run (run here against game.exe's real code, compared with d2d's real build), V read (disassembly or source read), D (MPQ tables), S (an earlier doc), I (inferred), U (unverified). Consumers: the two verifier briefs ([verify-hitrec-interrupt](../upstream/briefs/verify-hitrec-interrupt.md), [verify-missile-block](../upstream/briefs/verify-missile-block.md)), `docs/upstream/pipeline-2026-11-14.md` rows P04 / P06 / P09 / P10, the matrices (section 8). Data: `docs/research/data26/`. Fork branches (private remote): `fid/hitrec-interrupt-r1` (oracle `308e96f8`, claim `a78f66a3`) and `fid/missile-block-r3` (oracle `13783a6d`, claim `cd314f2c`), both on upstream main `fb6d799d` (the older `fid/missile-block-r1` and `-r2` on the remote are stale heads of the same work: never force-pushed, not to be used).

## 1. Questions

### P04, a hit interrupts a swing or cast

| Q | Question | Answer | Tag |
|---|---|---|---|
| Q1 | What decides whether a hit starts a get-hit while a skill is in use? | The server's player mode switch `FUN_005809d0` (called by the hit at `0x57d383`; its last argument 0) calls `FUN_0057edd0` (current-mode table) and, only if that agrees, `FUN_0057eec0(unit, new mode, 1, 0)` (call at `0x580a1c`). | V read |
| Q2 | What does `FUN_0057eec0` do? | State `0x36` (uninterruptable), a dead unit (modes 0, 0x11) and a missing unit refuse. With a used skill (`FUN_00620250` non-null) whose row's flags byte `+7` has the mask bit of `0x6ce284` clear (Skills.txt `interrupt` blank) it refuses mode 4 and 9 unless the unit is idle (NU); with the bit set it answers yes. No used skill: yes. | V read (`0x57eec0..0x57efb8`) and V run (below) |
| Q3 | Which rows refuse? | 22 of the 357 rows have `interrupt` blank: Kick, Jab, Strafe, Fend, Smite, Zeal, Charge, Concentrate, Frenzy, Whirlwind, the Identify and Town Portal scrolls and books (ids 217 to 220), Werewolf, Werebear, and six monster rows (QueenDeath, Cry Help, BearSmite, Delerium Change, SerpentCharge, Trap Nova). The `+7` bit equals the column in 357 of 357 rows (S, decomp20). | V run (`interrupt_swing.py game`, 335 of 357 allowed in each of 8 modes) + D (names from Skills.txt) |
| Q4 | Which current modes does the combined test allow? | For a row with the bit set, a Paladin is interrupted from A1, A2, SC, TH, KK (mode-end event queued), S1, S3, S4: 335 of 357 rows in each of the 8 modes, 2,856 cells. Class gates only: class 0 (Amazon) in S1 and class 5 (Druid) in S3 refuse every row (670 `C` lines). | V run |
| Q5 | Does the used-skill pointer outlive a swing? | No: idle, walk, run and the swing's end clear it (2,856 swing ends, 714 walk starts, 0 failures). | S (decomp23, V run; `warp-walk-game-trace.md` Part B) |
| Q6 | What does d2d do (before)? | `tick.cpp:715`: `!attack_mode(pmode) && pmode != BL`, so A1, KK, S1, S3, S4 are never interrupted and SC always is (for the 22 rows, wrongly). | V read |
| Q7 | d2d against game.exe before and after the claim | Oracle commit alone (`test_hit_recovery --interrupt`): 1,741 of 2,856 cells differ, 210 of the 357 own-mode cells (the mode d2d plays the skill in); 1,005 cells are interrupted with repeats, sequence or movement still pending. Claim: 0 of 2,856, 0 of 357, no pending state. | V run (NAS builder, `data26/interrupt-summary.txt`) |
| Q8 | What does a refused big hit still do in d2d? | It falls to the soft-hit branch and plays the struck cue; in game.exe the refusal path (`0x57d383` returns at once) sets no flag `0x8000` (only the too-small path `0x57d3b3` does). Unchanged by the claim. | V read |

### P06, a monster missile is blocked only with physical damage

| Q | Question | Answer | Tag |
|---|---|---|---|
| Q9 | Is the gate still open upstream? | Yes on `fb6d799d`: `rules::monster_blow` (`components/rules/combat.hpp:962`) rolls `defender.block` for every missile. (S-1 re-check; the earlier read was `4e054f43`.) | V read |
| Q10 | What does game.exe pass as can-block? | `FUN_005ad730` (the missile's hit entry) calls `FUN_0057dfb0(defender, 1, [record + 8] != 0)` at `0x5ad7f9..0x5ad806`; with 0 the block roll and its draw on the defender's seed are skipped (`0x57dfb4`) and `FUN_0057dd60` runs (dodge, avoid, evade, the claws' weapon block). | V read |
| Q11 | What is `record + 8`? | The record's physical field, rolled by `FUN_005a89a0` from the missile's stats 0x15 / 0x16 (`FUN_005a8910`: min and max BOTH over 0, else 0). Run through `FUN_005a89a0` and `FUN_005ad730` with `FUN_0057dfb0` hooked to read its third argument: can-block equals `lo > 0 && hi > 0` on every grid cell, including `min` 0 with `max` over 0 (the record is 0 and the missile is unblockable). | V run (`missile_block.py game`, 10 synthetic pairs; 29,421 real cases below) |
| Q12 | How many real shots are unblockable? | Of 29,421 cases of the existing `mm_diff.py` (30,000 cases, seed 20261115, 36 MissA / MissS1 / MissC rows, game.exe's own launch `FUN_0059fa30`; 579 Blood Raven cases excluded: her record carries her weapon's damage): 4,550 have a zero physical record. Rows with zero in EVERY case: `mummy1..4` (the poison shots), `pantherpotgreen`; 25 other rows are zero only for owner damage pairs of 0 (I). | V run + D |
| Q13 | The skill-launched missiles (`mm_launch.py`: Fallen Shaman fire, Radament's bolt, the Countess's and the vampires' fire, Frost Nova, Leap)? | 54 of 57 game-side launch records have no physical (`game256 0 0`): shafire1..4, unholybolt1..3, countessfirewallmaker, frostnova, leapknockback, vampirefirewallmaker, vampiremeteorcenter; only `andarielspray` (3 cases) has one. So game.exe never lets a shield block a Fallen Shaman's fire bolt. | V run |
| Q14 | d2d against game.exe before and after | mm_diff cases: parent 4,550 of 29,421 differ (d2d blocks ~75 percent of 400 trials where game.exe's can-block is 0), claim 0. Skill rows: parent 42 of 45 differ, claim 0; **12 further cases have no d2d row at all** (`shafire2`, `shafire4`, `unholybolt2`, `unholybolt3`: d2d loads only the Missiles.txt rows its monsters and skills name; `andariel_missile` returns without a missile when the row is absent, `ai.cpp:471`): listed by `compare` as NOROW, not judged. | V run (NAS builder, `data26/block-summary.txt`); the effect of NOROW is I |
| Q15 | Fixed where? | `MonStats::unblockable`, set by `monster_shot_stats` (`!(phys.low > 0 && phys.high > 0)`) and `missile_blow` (`!(info.min > 0 && info.max > 0)`), read by `monster_blow` for a missile. | V read (claim diff) |

### The P09, P10 and P03c notes

| Q | Result | Tag |
|---|---|---|
| Q16 | P09 (raised monster mode): `git grep -i ResurrectMode` on `fb6d799d` is empty and `ai.cpp:1963` still ends the raise with `set_mode(.., "NU")`: open. Issue draft [drafts/revive-mode-issue.md](../upstream/drafts/revive-mode-issue.md). | V read |
| Q17 | P10 (merc auras and regeneration): friendly auras on the merc and the Act 2 merc's aura are merged upstream (`50f7f610`, `fight/merc.cpp:81`); only the merc's own life regeneration (stat 0x4a, life / 2000 a frame) is open (no `regen` read in `merc.cpp`). Issue draft [drafts/merc-aura-issue.md](../upstream/drafts/merc-aura-issue.md). The pipeline row said "check what #504 and aura_pulse cover (U)": answered. | V read |
| Q18 | P03c (a knocking hit inside a get-hit, `FUN_00580a70` -> `FUN_0057edd0` with mode 0x13): `tick.cpp:660` still has no get-hit test, so d2d slides and restarts the hold where game.exe refuses. Candidate note with its oracle plan [drafts/hitrec-knock-in-hit.md](../upstream/drafts/hitrec-knock-in-hit.md); build it after P04 (same lines). | V read; oracle not run (U) |

## 2. Deviations (d2d against game.exe)

| id | Rule in game.exe | Rule in d2d (`fb6d799d`) | Tag | Differential |
|---|---|---|---|---|
| D1 | A hit interrupts Attack, Sacrifice, Vengeance, Conversion and every cast whose row has `interrupt` set (335 of 357 rows) | no melee swing is interrupted | V run (1,741 of 2,856 cells, own-mode 210 of 357) | yes: `tools/emu/interrupt_swing.py` (fork branch); fixed on `fid/hitrec-interrupt-r1` |
| D2 | The 22 rows with `interrupt` blank (Zeal, Smite, Charge, scrolls and tomes, Werewolf, Werebear, ...) are never interrupted | every cast is interrupted (SC), including the TP and ID scrolls and books and the shapeshifts | V run | same differential; same fix |
| D3 | A monster missile with a zero physical record cannot be blocked | blocked like any other | V run (4,550 of 29,421; 42 of 45 skill cases) | yes: `tools/emu/missile_block.py` (fork branch); fixed on `fid/missile-block-r3` |
| D4 | A refused big hit sets no struck flag | the soft-hit branch plays the struck cue | V read | not built: needs the client-side cue (`FUN_004cc5b0` path); U |
| D5 | `shafire2..5`, `unholybolt2..3` exist (Fallen Shaman levels, Radament's variants fire them) | d2d loads only named rows: those rows are absent, so `andariel_missile` makes no missile | V run (rows absent) + V read; effect I | the compare lists NOROW cases; a `test_game` scenario (a Fallen Shaman 2 with `trans_lvl` 1 and a target in range, count missiles) would prove the effect |
| D6 | A skill missile with SrcDamage over 0 carries the owner's damage in its physical record | `missile_blow` takes the row's min and max only | U | the oracle's skill compare would show it; only `andarielspray` has a record, and it matches the can-block |
| D7 | A knocking hit inside a get-hit is refused (mode 0x13 from GH) | slides and restarts the hold (`tick.cpp:660`) | V read; game column S (P03b run covered modes 4 and 9) | buildable from `mode_switch.py` plus a 0x13 column (note in drafts/hitrec-knock-in-hit.md) |
| D8 | A raised monster plays its `ResurrectMode` | NU | V run (game side, S: decomp9), V read (d2d) | none (the raise is inline) |
| D9 | The merc regenerates life (stat 0x4a) | no steady regeneration | V run (stat), I (effect) | buildable (draft) |

## 3. Bot findings (described, NOT changed)

| id | Where | Finding |
|---|---|---|
| B1 | `internal/combat/resource.go:103`, `internal/resource/types.go:69` (`CastAdjacentMax`, `combat.crowdAdjacent`), `docs/fixcrowd.md` | The crowd guard assumes a melee hit cuts a Town Portal cast. In game.exe it cannot (the scroll and tome rows are two of the 22 refused rows, V run). Before this claim d2d cut it on a big hit (the `debug hit` cheat did, fixcrowd.md); after it d2d agrees with game.exe. The guard stays safe but is stricter than the rule; tune only after drill R3 on the fork binary with the claim. |
| B2 | `internal/town/trip.go:210` (`openPortal`, 3 s wait for the portal object) | A hit during the cast no longer makes the wait fail on the fork binary with the claim; the "town portal did not appear" warning should drop. U until the drill. |
| B3 | `internal/combat/fight.go:310` (armor upkeep, recast after a cut) and the Zeal / Smite chains | With the claim, a big hit now cuts an Attack / Sacrifice swing (the bot sees mode 4 and a 10-tick hold), never a Zeal or Smite chain. The bot's "recast after a cut" logic is therefore live for Attack, Sacrifice and casts, dead inside Zeal and Smite, as in game.exe. No change needed. |
| B4 | the Fallen Shaman threat in the bot's danger model (`internal/combat`) | In game.exe a shield never blocks the shaman's fire; the fork binary now agrees. Whether the bot's threat model reads block against fire bolts was not checked (U); no change. |

## 4. PR candidates

Both are proven with the EXISTING differential style (oracle commit builds alone and prints the parent's mismatches; claim 0). Neither is posted; the next gate is the independent Opus verifier (`d2-reviewer`, singly), then the user's go.

| PR | Branch (private fork) | Before | After | Tests | Brief |
|---|---|---|---|---|---|
| P04 | `fid/hitrec-interrupt-r1` oracle `308e96f8`, claim `a78f66a3` | 1,741 of 2,856 cells differ | 0 | `test_hit_recovery`, `test_game`, `test_combat`, `test_rules`, `test_swing_start`, `test_swing_players`, `test_classic_swing`, `test_game_rules`, `test_monster_ai`, `test_monsters` rc 0 | [verify-hitrec-interrupt](../upstream/briefs/verify-hitrec-interrupt.md); post only after #548 merges (same condition) |
| P06 | `fid/missile-block-r3` oracle `13783a6d`, claim `cd314f2c` | 4,550 of 29,421 (+ 42 of 45 skill cases) | 0 (+ 0) | `test_monster_missile` (unit and `--missile-block`), `test_game`, `test_combat`, `test_rules`, `test_monster_ai`, `test_monsters`, `test_game_rules`, `test_monster_thinks`, `test_missile_blow_poison`, `test_energy_shield` rc 0; `test_missile_hit` rc 77 (skipped, needs its own input) | [verify-missile-block](../upstream/briefs/verify-missile-block.md) |

`pr_guard.py upstream/main` 0 FAIL 0 WARN on both heads; `presubmit.py --quick` 0 FAIL (warnings: siblings, data-gating of an existing read, build skipped by `--quick`; to answer in the body: P04 touches melee, skill but not missile (monster missiles are not interrupted by a hit), P06 touches missile only). Drafts: [hitrec-interrupt](../upstream/drafts/hitrec-interrupt.md), [missile-block](../upstream/drafts/missile-block.md); bodies in `docs/upstream/bodies/`. The heads were built and tested at `b4d81433` (P04) and `c439ac7e` (P06); the pushed heads differ only in README / combat.md wording and the oracle script's NOROW handling (V read: `git diff` of `components/` and `tests/` between them is empty).

## 5. Dead ends and limits

- The struck cue of a refused hit (D4) was not run: the client cue path is not in the server mode switch.
- `FUN_0057eec0`'s third argument (a skill id compared with the used skill, ids 0x43 and 0x4c) is 0 for the get-hit call, so that branch is not reachable from a hit (V read).
- `FUN_005ad730` was run with its callers' side effects stubbed (`FUN_0057c6c0`, `FUN_0057cee0`, `FUN_0064a5a0`, `FUN_0064a190`, `FUN_005a5560`) and a zeroed Missiles row (flags 0, ToHit 0): it runs `FUN_0057dfb0`'s argument for real, but the later users of the record are not exercised.
- The skill-missile record when the owner has damage (D6) and `missile_blow` against `FUN_0064b860` for rows other than the 14 in `mm_launch.py` are U.
- KK (kick) needs a queued end event in game.exe (`FUN_0057edd0`); the oracle sets one (a swing has it) and d2d sets `pmode_until`; the no-event case is S (decomp22).

## 6. Drill recipe (lab fork binary, docs/forkdbg.md and docs/forkcheats.md; NOT run)

| id | Confirms | Steps |
|---|---|---|
| R2 | D1 | lab binary with the P04 claim, fresh throwaway Paladin, a dummy from `debug spawn` (Attack is skill 0 and needs no points), hold Attack, `debug hit 20` at tick 3 of the swing: expect mode 4 (GH) at once and no strike at tick 7 (d2d parent: mode stays A1, the strike lands). Repeat with Sacrifice (expect GH) and with `debug points 106 1` (Zeal) and Smite (expect no flinch). `debug hit` enters the get-hit "unless it kills" through the game's own hurt path: it is the right probe here but the fixcrowd note says it also cut a cast on the parent. |
| R3 | D2, B1, B2 | read a Town Portal scroll (the real inventory click, not `debug cast`), `debug hit 20` at steps 2 and 3: expect mode 10 to continue, the portal after 4 steps, the scroll spent once (parent: the cast is cut). Repeat with Werewolf. |
| R6 | D3, D5 | a Paladin with a shield and Block over 0 (`debug stat` to set Block 75 only on a throwaway), `debug spawn` a Fallen Shaman (fallenshaman1), stand in range, 200 casts: count blocks via `debug combat log`: expect 0 (parent: about 150). With fallenshaman2 and 4: count the `shafire` missiles in `debug missiles`: game.exe fires, d2d expected none (D5, I). |
| R7 | D7 | hold a get-hit (`debug hit`), then a knocking hit (a Fallen Shaman... or a Barbarian's Leap Attack from a spawned unit): game.exe expects no slide and an unchanged hold end; d2d slides. |

## 7. Reproduce (the worker host and the NAS builder; the scripts of this card are in the scratch directory, the fork scripts in the fork branches)

```
# game side, emulator image (docs/emu-image.md), one core, nice 10
$PY tools/emu/interrupt_swing.py game > game.tsv                       # 3,526 lines: 2,856 G, 670 C
$PY tools/emu/mm_diff.py game TABLES 30000 20261115 OUT                 # the existing oracle: OUT/cases.txt, OUT/game.tsv
$PY tools/emu/missile_block.py game OUT/game.tsv > block.tsv            # 29,421 B, 579 E
$PY tools/emu/mm_launch.py TABLES > launch.out; missile_block.py skill-cases launch.out > skill_cases.txt; missile_block.py game-skill launch.out > skill_block.tsv
# d2d side, NAS builder (Release, tests and tools on, nice, -j3)
test_hit_recovery --interrupt | grep '^I '                                # oracle commit: parent; claim: after
test_monster_missile --missile-block OUT/cases.txt | grep '^K '
test_monster_missile --missile-block-skill skill_cases.txt | grep '^[KN] '
$PY tools/emu/interrupt_swing.py compare game.tsv d2d.tsv               # exit 1 before, 0 after
$PY tools/emu/missile_block.py compare block.tsv d2d_block.tsv          # same, and for the skill files
```

## 8. Matrix-cell updates (proposed text; not applied)

- `pal-skill-matrix.md`, Attack / Sacrifice rows, defence column: "a big hit interrupts the swing (GH, 10 ticks at FHR 0); in d2d before fid/hitrec-interrupt it did not (decomp26 D1)". Zeal, Smite, Charge: "never interrupted by a hit (V run, 22 rows with `interrupt` blank); d2d agreed by accident for melee". Deviation list: D7 "a get-hit never interrupts an Attack / Sacrifice swing in d2d" is now a proven candidate (P04); D8 / D9 stay (block length merged by #541, restart #548).
- `sorc-skill-matrix.md`: every Sorceress cast is interruptible (V run, 335 of 357 rows); the TP scroll and tome are not (Q3); d2d agrees after P04.
- `skill-matrix` MA10 (monster missiles, `fid-missile-monster.md`): "UPDATE 2026-11-15 (decomp26): block gate proven on game.exe's own code, V run: can-block = record physical over 0 (FUN_005ad730 -> FUN_0057dfb0), 4,550 of 29,421 cases differ in d2d, 0 after the claim (`fid/missile-block-r3`); the Fallen Shaman's fire has no physical (54 of 57 skill-launch records), D5: d2d does not load `shafire2/4`".
- `docs/research/hitrec-trace.md` (branch decomp20) D1 and D5: V run closed, deviations proven by differential; R1 answered (decomp23).
- `docs/probes/fid/fid-revive-mode.md` D1 and `merc2-trace.md` D1, D3: re-checked on `fb6d799d` (Q16, Q17): D1 and D3 of the merc trace are merged upstream (`50f7f610`); D2 (regeneration) stays open.

## 9. Glossary

No new term needed. "Used skill" (the unit's current skill pointer, `FUN_00620250`) appears in `hitrec-trace.md`; `GLOSSARY.md` has no entry for it: proposed line for the coordinator: "Used skill: the Skills.txt row a unit is swinging or casting; game.exe keeps a pointer from the skill's start to the end of its swing or cast and reads its `interrupt` column when a hit lands."
