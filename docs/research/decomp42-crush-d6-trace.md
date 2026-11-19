# decomp42: crush D6 (the get-hit size), P08 re-run, player magic resistance, the next deviations (2026-11-19)

> status: current, 2026-11-19 (agent decomp42, Sonnet; docs, `tools/emu` copy of the oracle and a branch in the private fork; nothing posted, no live game). Follows [decomp41-merc-crush-trace.md](decomp41-merc-crush-trace.md) (D6 was "scoped, not built" there). The independent Opus verifier (singly, scheduled by the coordinator) comes before anything is posted. Body: `docs/upstream/bodies/crush-d6-body.md`.

Tags: **V run** (game.exe's own code, or d2d's real code, run and compared), **V read** (disassembly or source read), **D** (data), **S** (an earlier doc), **I** (inferred), **U** (unverified).

## 0. Answers in short

| Item | Status | Branch (private fork) | Proof |
|---|---|---|---|
| 1. Crush D6: the get-hit size test counts the crush in d2d, not in game.exe | **built** | `fid/crush-d6-r1` (oracle `abdd63a7`, claim `18961f04`) on upstream main `da98b7f8` | section 1 |
| 2. P08 timed-state put rule | **no candidate**: the curse slot is merged upstream; `curse_put.py` against upstream main: 0 of 20,000 and 0 of 100,000 puts differ (V run) | none | section 2 |
| 3. Player magic resistance (decomp37 T3) | **not worth building now**: no Normal Act 1 to 3 monster deals magic damage to the player (S, I); stays scoped | none | section 3 |
| 4. The next three player-visible deviations | list, ranked by Normal Act 1-2 impact | none | section 4 |

Upstream at this run: `psi29a/d2decomped` main `da98b7f8` (V read, `git fetch upstream`; the maintainer's #625 curse slot, `fe072fad` states and #626 Cairn are in). The new branch is two commits on it, author and committer the public noreply identity, `Agent: Claude Code` the only trailer after a blank line.

## 1. Crush D6: a Crushing Blow's life is not in the get-hit's size

### Questions

- **Q1. What does game.exe test for the size of a blow?** V read: `FUN_0057cb00` (`0x57cb22..0x57cb35`) tests the record's total `+0x4c` (under `0x100`: no get-hit), against the monster's max life over a divisor by hit class (S: `docs/research/re/combat.md`, hit_rule.py's 100,000-case oracle).
- **Q2. When is `+0x4c` written, and does the crush enter it?** V read: `FUN_0057c1e0` sums `[rec+8] + [rec+0x10] + [rec+0x1c] + [rec+0x20] + [rec+0x24] + [rec+0x28]` (physical, fire, lightning, magic, cold, poison, `0x57c3a5..0x57c3bd`) plus `[rec+0x38]` for a flagged unit into `+0x4c`. `FUN_0057c6c0` reads it at `0x57c852` and takes it off stat 6, AFTER the strike events (`0x57c7b7..0x57c7f3`, where `FUN_005bffc0` writes the crush to stat 6 directly) and after the cut of the record's physical to the life the crush left (`0x57c7f8..0x57c826`). `FUN_0057c6c0`'s second argument is a missile flag: nonzero it calls `FUN_0057c1e0` itself (`0x57c793..0x57c7a0`) and fires events 6 and 2; zero (a melee blow) it does not, so the caller has already filled `+0x4c` (`0x57c6c0` has two callers, `0x56ba73` with 1 and `0x57d5c7` with 0; which function fills the melee total was not read: U, harmless: the sum is the same function and the crush handler never writes `+0x4c`).
- **Q3. Is that what the machine does, run?** V run (`crush_hit.py game`, builder, emulator): per case `FUN_005bffc0` natively (the crush off stat 6), `FUN_0057c1e0` natively (the total), `FUN_0057c6c0` natively with the dispatcher `FUN_005c0c30` setting stat 6 to the crush's life at event 5 (as `crush.py pair`), then `FUN_0057cee0` natively on the SAME record (the get-hit, the seed's rolls, as `hit_rule.py`). The record total after the handler equals the case's physical in every case (the crush is never in it; the `total` column of `game.tsv`), and the life ends at `life - crush - total`. The harness choice that the melee total comes from calling `FUN_0057c1e0` first is I (Q2).
- **Q4. What does d2d do?** V read (`da98b7f8`): `Fight::land` (`fight/strike.cpp:553..554`) hands `hurt` `taken = soak(blow.phys_fine, (blow.damage << 8) + blow.frac)`; `player_blow` adds the crush to `blow.damage / frac` (`rules/combat.hpp:1050..1052`), so `hurt`'s `blow_fine` (`ai.cpp:2927`), the size for `rules::get_hit`, includes the crush. A crush proc therefore staggers a monster as if it were a bigger blow.

### 1.1 Differential (V run)

`crush_hit.py gen N SEED` writes `id maxlife life phys hc div seedlo seedhi` (256ths; boundary cases around `max / divisor-by-class`, the crush share and survival: the monster is alive after the blow, so the get-hit test is reached); `game` is Q3; `test_crush_get_hit --crush-hit` prints d2d's real `player_blow` (Crushing Blow 100 percent, a fixed physical, always hitting, the seed advanced until it crushes), `soak` and `hurt` on a zombie made plain or a superunique (divisor 8) with the case's lives and seed; `compare` prints how many get-hits and seeds differ and how many times d2d staggers where game.exe does not, and exits 1 on any difference, an empty, short or malformed file. Control: the same blows with no crush on both sides (`--no-crush`): the harness is exact without the crush.

| | seed 7 | seed 5001 | seed 9091 |
|---|---|---|---|
| cases | 100,800 | 100,800 | 100,800 |
| parent (oracle commit `abdd63a7`) | 72,108 differ (get-hit 46,345, seed 54,280), exit 1 | 72,163 (46,555 / 54,504) | 72,111 (46,412 / 54,459) |
| claim (`18961f04`) | 0, exit 0 | 0 | 0 |
| control, no crush, parent and claim | 0 | 0 | 0 |

The grid sits on the thresholds on purpose, so the parent's 72 percent is not a play frequency. A play-shaped grid (`playgrid`, 4,890 surviving crush procs: plain monsters of 12 to 1,024 points of max life at full and half life, a physical of 2 to 20 points, hit class 0, 30 seeds each; V run on the same oracle): game.exe staggers 1,921 (39 percent), the parent 4,513 (92 percent), the claim 1,921 (0 differ). A Crushing Blow weapon staggers a monster on nearly every proc in the parent whatever the weapon's damage, because the crush alone is a quarter of its life (more than max / 16).

Edge exits (V run, `crush_hit.py`): an empty game file, an empty d2d file, an empty case file, a short d2d file, a malformed d2d line, a missing file, no argument, `game` on an empty case file, `gen` without arguments: every one exits 1; the equal 8-case set exits 0.

### 1.2 The claim (`18961f04`), 8 files in 2 commits

`Blow::crush_fine` (the crush's share of `damage`/`frac`, set where the crush is added), `hurt(..., crush)` takes it off `blow_fine` for the get-hit only (the life still loses it), `blow_hurt(game_data, monster, blow, now_ms)` is the soak and the `hurt` call `Fight::land` made inline (now `land` calls it, so the test drives the same line), and `test_crush_get_hit` pins eight game.exe decisions. `docs/research/re/combat.md` says where the crush stands in the total. Bone Armor still takes the physical only (`soak`), and the total game.exe reads is after that event (`FUN_0057c1e0` raises event 0xb before its sum, S: `combat.md`), so `taken - crush` is the record total.

### 1.3 Mutants of the claim (V run; builder, local, never pushed)

Each mutant is one text change on the claim (`mutd6.py`, builder, long lane, rebuilding `test_crush_get_hit`); `test_crush_get_hit` (pins) and the compare (seed 7, 100,800 cases) run on each.

| Mutant | pins | compare (of 100,800) |
|---|---|---|
| m1 the parent: crush not taken off the size | abort | 72,108 differ |
| m2 `crush_fine` never set | abort | 72,108 |
| m3 `blow_hurt` passes 0 | abort | 72,108 |
| m4 the crush taken off twice | abort | 23,035 |
| m5 the crush cut to whole points | passes the first eight pins; the ninth pin (a crush of 12.5 points under the threshold) was added for it and the mutant was not rebuilt against it (U, a re-run for the verifier) | 9,300 |
| m6 no floor at 0 on the size | passes | 0: an **equivalent mutant**: the size is `taken - crush` and `taken` includes the crush (soak takes at most the physical), so it is never negative (I) |
| m7 `crush_fine` halved | abort | 50,307 |
| m8 a missile blow carries no crush | passes | 0: **escape**: the oracle runs the melee path only (D4) |
| m9 `marked` dropped in `blow_hurt` | passes | 0: **escape**: the oracle's records carry no element (D3) |

Built locally, never pushed. 7 of 9 are caught by the compare (5 also by the pins); m6 is equivalent; m8 and m9 are the two disclosed limits (a missile record and an element are not in the cases). The first mutant pass built m3 to m6 and m9 on a stale object (an edited header restored with its old mtime), so they equalled the parent; the table is the second pass with the mtime touched (V run: the m4 and m5 binaries differ from the parent's in play).

### 1.4 Build and tests

BUILD_PLACEHOLDER

### 1.5 Deviations (d2d against game.exe)

| D | Rule on both sides | Tag | Differential |
|---|---|---|---|
| D1 | game.exe: the get-hit size is the record total (+0x4c), the crush is not in it. d2d (parent): the crush is in it. Fixed by the candidate | V run | yes: `crush_hit.py` + `test_crush_get_hit` (built) |
| D2 | A crush by the MonProp on a monster blow (monster hits the player, merc or a pet): the crush is applied after the hits in d2d (`ai.hpp:683`, `tick.cpp:634/762`), before them in game.exe (decomp41 D7); the pet's get-hit has its own path | V read d2d, S game | the merc / pet get-hit sizes after a crush are not covered; an oracle would hook `FUN_0057cee0` for a unit type 0 with the same chain (the harness already runs it for type 1; change the unit type and the mode hooks) |
| D3 | Hit class and elemental marks with a crush: `crush_hit.py` leaves the record's elements at 0, so `marked` / class-by-element with a crush is not run; the claim does not touch `marked` | U | extend `gen` with element fields (the `hit_rule.py` columns) and print the d2d `marked` |
| D4 | A missile-borne crush (event 6, divisor doubled, `FUN_0057c6c0` second argument nonzero, `0x57c793`): the chain was run for the melee path only | U | run the chain with argument 1 (event 6 and the missile record); d2d's `swing.missile` already doubles the divisor in `player_blow` (pinned upstream by `test_combat`) |

### 1.6 Bot findings (described, not changed)

| B | Finding | Where | Tag |
|---|---|---|---|
| B1 | No bot code reads stat 136 (Crushing Blow) or models how it staggers a monster; weapon choice and kill-time estimates see only damage | `internal/` (grep `crush`: only party scaling and unrelated words) | V read |
| B2 | `party/scaling.go` still says "Crushing Blow is divided by 0.5 + 0.5 n": 2 and 3 players match game.exe (+50, +100 percent, V run decomp41); it does not say bosses divide by 8 and a plain unique by 4 | `internal/party/scaling.go:29` | V read, repeated from decomp41 B4 |
| B3 | A bot that uses a Crushing Blow weapon in d2d sees monsters stagger (lose their attack) on nearly every proc; in game.exe they stagger on 39 percent of procs in the play grid. Any soak or A/B number measured on d2d with a Crushing Blow weapon is flattered until the claim lands | measured runs with weapon `hax` in the survival A/B logs do not carry stat 136 (S: the card text gives no property) | U |

### 1.7 PR candidate

**Yes, `fid/crush-d6-r1`, local, not pushed to GitHub, awaiting the independent verifier.** Mismatches before: 72,108, 72,163, 72,111 of 100,800 (three seeds); after: 0; the control is 0 on both trees; MUTANT_SUMMARY. Standard (docs/upstream/verifier-*.md): the oracle commit comes first and builds alone; the oracle runs game.exe's functions natively; `compare` exits 1 on parent, empty, short, malformed; the claim's `test_crush_get_hit` pins game.exe's values; no force-push; the body is `docs/upstream/bodies/crush-d6-body.md`. The call site `Fight::land -> blow_hurt` is read, not run (a mutant that puts the old inline `soak` + `hurt` back in `land` would pass the oracle: listed below as an escape, tagged U): verifier, please grind it.

### 1.8 Drill recipe (lab binary, fork verbs; run later, not now)

R1 (D1): `debug spawn zombie1 1` or a Fallen, `debug mon set <id> life <n>` (docs/forkcheats.md: the verb names; a weapon with Crushing Blow: `debug give` of a base item plus a load of a `.d2s` with a crush property as a SETUP, logged in `docs/probes/test-characters.md`; U whether a verb sets properties), swing at it from a fixed lockstep seed over 40 swings, `monsters` after each swing (GH or not) and the seed. Expect (claim): a swing under max / 16 never staggers even when it crushes; (parent): the crush alone staggers. R2 (D4): the same with a missile (a javelin with the property), which doubles the divisor.

### 1.9 Matrix-cell updates (proposed text)

- `docs/research/decomp37-combat.md` section 6 / `docs/research/decomp41-merc-crush-trace.md` section 3: "D6 get-hit size counts the crush: **V run (decomp42), built `fid/crush-d6-r1`**".
- `docs/upstream/pr-queue.md` (shared; the merger): a row "Crush D6 get-hit size, `fid/crush-d6-r1`, ready for the verifier".

## 2. P08: the timed-state put rule, re-run on current upstream main

- **Q5. Does a provable curse-slot deviation remain on `da98b7f8`?** No. V run: `curse_put.py gen 100000 5001` / `game` / `test_curse_put --curse-put` / `compare` on the builder, upstream main built (`test_curse_put` from `da98b7f8`): 100,000 puts, **0 differ**; seed 7, 20,000 puts, 0 differ. Upstream took what we proved: #603 (Attract refuses every curse), #625 (`a8f06245`: one curse slot for Taunt, Battle Cry, Cloak, the Overseer's whip: covstates2 D1 and D5) and the maintainer's `fe072fad` (one node per non-curse state: D2). So covstates2 D1, D2, D3, D5 are fixed upstream.
- **What is left of covstates2 / P08, and why no candidate.**

| Left | Rule | Tag | Reach in Normal Act 1-2 | Provable now? |
|---|---|---|---|---|
| D4 Taunt's length floor | d2d floors monster state lengths at 25 frames, Taunt's data length is 0; game.exe keeps 0 = no end; where a Taunt really ends is not read | V run (put), U (end path) | Barbarian's Taunt on monsters: nil for Paladin and Sorceress heroes | no: the end path is U |
| D6 curse resistance (stat 109 on the player) | game.exe cuts a monster's curse on the player by the percent; d2d has no `curse_resistance` | V run game side (640 cases), d2d S | only gear with the stat; none in Normal Act 1-2 | d2d side needs a player with the stat: no harness |
| D7 state groups (Frozen / Chilling / Shiver armor, Quickness / Fade) | game.exe removes group members | V run game side (370 cases), S d2d | Sorceress armor swaps: real but cosmetic | needs a `Fight` object |
| poison replacement (tie, near rates, under one point) | `fid-poison-stun.md` D1 to D3: 2,629 of 19,997 on the old hand-built d2d side | V run | the poisoned monster must be poisoned by the player: Paladin and Sorceress heroes have no poison; a poison merc is not in Act 1-2 | the d2d side was hand-built in a scratch checkout, not in-tree: an effort L rebuild for no early reach |
| stun seed draw (boss column / Velocity 0) | game.exe draws one LCG step and then does not stun; d2d draws nothing: the attacker's seed stream differs, not the outcome of the blow | V run (1,165 cases) | stun gear only | the effect is a random-stream shift (I); no player-visible outcome |

Verdict: **no PR candidate from P08.** The hero builds in the lab never poison or stun monsters, so the remaining rows have no Normal Act 1-2 reach. Re-open P08 only for a Necromancer or Amazon build or Barbarian Taunt.

## 3. Player magic resistance (decomp37 T3): decision

- **Rule (V run, decomp37):** a monster's magic field goes through the player's magic resistance (stat 37, row 4 of the table at `0x732980`, `FUN_0057be00`); d2d: `Fighter::res` has four entries and `monster_blow` skips type 4 (decomp41 section 5; scope: five-entry `Fighter::res`, three callers, an oracle column in `gear_absorb.py`, about 400 lines).
- **Player-visible impact:** which monsters deal magic damage to the player at all? S (decomp41, `monsterhit-trace` D2): 21 MonStats rows with a magic El row, all Act 5 and later (Hell Act 5 melee builds). Which sources give the player stat 37? A handful of uniques and charms (decomp41, S), none a Normal Act 1-2 hero wears. I: for a Normal Act 1 to 3 hero both sides are zero, so the deviation cannot change a fight. Not re-measured here (no data run in this brief): a `grep` of the 21 rows is the first step if a later build reaches Act 5.
- **Decision: do not build now.** Effort M for impact zero until the bot reaches Hell Act 5. Keep the scope in decomp41 section 5 as is.

## 4. The next three player-visible deviations (scan; not re-run)

Method: read of `docs/upstream/pipeline-2026-11-14.md` section 1b against upstream main's log (`git log upstream/main --since=2026-11-10`, V read: what merged since), `docs/probes/fid/*`, the survival A/B probe docs (`docs/probes/survival-ab-*`, `survival-holdout*`: bot-side zone and chicken findings, no game.exe deviation among them), and `docs/research/re/deviations.md`. No new emulator run. Merged since the pipeline file: hit recovery and its follow-ups (#541 to #559), missile block (#553), merc life (#604, #615), curse slot (#603, #625), summon stats and Summon Resist (#627), get-hit size on 256ths. What is left in a Normal Act 1-2 fight, in order:

1. **P07, the monster block gate (V run game side, d2d S read): the highest.** game.exe lets a monster block only when its MonStats row has the bit-4 gate (Mephisto, Diablo, Baal's clone) or it was spawned with a shield component (skeletons always; Fallen and corrupt rogues on some spawns: 4 of 24 Fallen spawns), chance = ToBlock; d2d lets every `mBL = 1` row block with its ToBlock: a Fallen blocks about 9 percent of landed blows in d2d against about 1.5 percent in game.exe at Normal (`fid-monster-block.md`). Player-visible for every melee build on Act 1 (a blocked blow does no damage and plays the block animation). It makes d2d harder than game.exe (the safe side for the bot's numbers). **Blocked, not provable cheaply:** the shield is a per-spawn variant choice (`FUN_00664860`, layer 7), d2d needs the variant pick at spawn first (a per-spawn shield flag) and the d2d side of the differential is not run. A first step that needs no variant: the loaded-data comparison (60 of 733 rows differ at Normal) is already a V run.
2. **Crush D2/D4 (section 1.5): the merc and pet get-hit size after a crush and the missile-borne crush.** A merc (Rogue, Desert Guard) with a Crushing Blow item staggers monsters too often in d2d as the hero does; the pet / merc get-hit path calls the same `land`. Provable with the same chain (D4 is one argument change in `crush_hit.py`, D2 a unit-type hook); effort S once D6 is verified; reach: merc gear with the stat only.
3. **P09, a raised monster stands in NU where game.exe plays MonStats2 `ResurrectMode` (S1 for Fallen and Mummy, SQ for skeletons): negligible.** V run game side (703 rows, 0 disagree; fid-revive-mode), d2d V read (`git grep ResurrectMode` in `components` is still empty on `da98b7f8`, V read). About 24 ticks of the raised monster's own animation; reach: Fallen shaman raises in Act 1 (a shaman resurrecting Fallen). Third place only because nothing else with Normal Act 1-2 reach is open: the hit-recovery rows (#541, #548, #557, #559), the missile-block row (#553) and the curse-slot rows are merged (V read of the log above; `combat.md`'s approximation line "FBR acts as an animation-rate bonus on the plain frame count" predates #541 and looks stale: U, a docs-only check).

Nothing here is a PR candidate in this run: 1 needs the variant pick, 2 waits for the D6 verifier, 3 needs a d2d harness for the raise. I did not scan live runs; the survival A/B anomalies (zone choice after a chicken, `plan/model mismatch after a chicken`, `survival-holdout2` section 1.4) are bot-side (model and re-plan), not d2d versus game.exe: no monster-behaviour deviation was found in them (I: read of the probe text, no log analysis).

## 5. Limits and dead ends

- D6 is the melee path only (D4): the missile record and event 6 are not run. The harness sets elements to 0 (D3). Hit classes 0 to 13 and random classes up to `0x5f` are in the grid; the claim does not change class or `marked`.
- Dead end: the first harness passed argument 0 to `FUN_0057c6c0` and read `+0x4c` = 0 (the melee path does not fill the total itself); fixed by running `FUN_0057c1e0` first. A first case file with lethal blows aborted the d2d test (the monster dies before its get-hit): the generator keeps the monster alive.
- No live game, no lab drill, no Windows or macOS build; ctest as in 1.4.

## 6. Shared-file lines for the merger

See `backlog/merge-lines/agent-decomp42.txt`.
