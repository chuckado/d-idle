# Det state hashes, second run: coverage gaps closed, five builds (agent DETHASH2, card dethash2)

> status: current research record; follows docs/research/det-hashes-2026-11-01.md (read its "What the harness is" first). Tags: V = run by us (how), S = read in a doc or source, U = unverified.

## What changed against the first run (V)

The first run left four gaps: no cold missile visibly hit a monster, no merc or pet existed, the player position was not hashed, and the freeze-rules candidate was not run. This run adds 9 scenarios to the 5 old ones (tools/emu/det_scenarios_lab.py, driver and compare tool tools/emu/det_scenarios_run.py), all on a headless d2d with the lab patch set, one hidden slot, no game.exe. Ruff 0.17 clean under its defaults.

- Every snapshot after a block of steps is five read-only replies: `monsters`, `debug monstate` (new, see below), `debug players` (the player's cell), `dump merc`, `debug pets`. Two hashes per run (md5, first 12 hex): `mon` is the `monsters` replies only (the first run's hash, so the old numbers are comparable) and `state` is every line of all five replies (monsters, chill and stun deadlines, the player's position, the merc, the pets).
- `debug monstate` is a read-only lab verb added for this run (id, name, hp, boss kind, chill deadline, stun deadline, mode, position of every live monster; 8 lines in `debug_fight.cpp`). Without it chill and freeze are invisible: `monsters` shows neither. It is lab patch set material, so it stays out of this repo (docs/decisions.md: patches never go into this repo); it is on the worker host with the build trees. U: it is not committed to the private fork (the PR checkout's remotes are public); say so if the next run needs it.
- The lab set is the first run's ported commit (094ec6cc) plus the four later cheat commits (`debug spawn`, `merc hire|kill|revive|hurt`, `debug missiles`; the first run did not carry them) plus `monstate` plus one fix: `give <code> belt` needs an explicit box (`belt_slot` is not upstream; the shift-click belt resolution of the first run). Every build carries the identical lab commits.

## Builds (V, worker host, trees under `pr/dethash2/`; `cmake -DCMAKE_BUILD_TYPE=Release -DD2D_WERROR=OFF`, target `d2d`)

Baseline upstream main is still 72132b99 (V: `git fetch upstream`, 2026-11-01, not moved since the first run). Each candidate is `git cherry-pick 72132b99..HEAD` of the candidate on top of the lab commits (no conflicts).

| tag | tree | d2d sha256 (first 12) |
|---|---|---|
| base, base2 | upstream main 72132b99 + lab (run twice) | 70c7b3694173 |
| p155 | #155 head 8c56e96a (Static Field + cold chill) + lab | 97b890153150 |
| freeze2 | `fix/freeze-rules-v2` head 409903a6 (C5, C3, C4 on top of #155's 8c56e96a) + lab | bfa26babd126 |
| walk | `fix/walk-velocity` as opened in PR #156, head e4a13181 (V: `gh api repos/psi29a/d2decomped/pulls/156`, base 72132b99, five claim commits + tooling) + lab | ea417b2896ee |

## Commands (V)

Inside one slot (`D2D_SLOT`, hidden headless d2d; one slot used at a time), on the worker host:

    export D2D_DATA=GAMEDIR
    S=passive,fight,cold,static,cold2,hit_plain,hit_champion,hit_unique,hit_boss,hit_immune,hit_minion,merc,pet,follow
    tools/d2dslot.sh run -n 1 -- python3 tools/emu/det_scenarios_run.py run BINDIR OUTDIR $S base,base2,p155,freeze2,walk
    python3 tools/emu/det_scenarios_run.py compare OUTDIR $S base base2,p155,freeze2,walk
    python3 tools/emu/det_scenarios_run.py table OUTDIR hit_plain freeze2      # the target's hp / chill / stun per block

`BINDIR/d2d.TAG` is the binary of each tag. Seed 3, lockstep, a fresh d2d per run, 70 runs of about 30 to 60 s.

## Scenarios (V)

| scenario | what | new coverage |
|---|---|---|
| passive, fight, cold, static, cold2 | the five of the first run (zombies and quill rats in Blood Moor; Ice Bolt 39, Static Field 42, Frozen Orb 64) | none (same inputs; the hashes of upstream main reproduce the first run's: d1ca6c46, 5f906303, bba8362c, 34a1b9ef) |
| hit_plain | a plain zombie (Drowned Carcass, 125 to 149 life) | cold hits that land, hp, chill and stun printed |
| hit_champion | the same row as a champion | champion class |
| hit_unique | the same row as a unique (no mods asked: random mods from the seed) | unique class |
| hit_boss | Blood Raven spawned by row (a MonStats boss-column monster, not a superunique) | boss class |
| hit_immune | a Hell zombie (MonStats cold resist 100 in Hell), `debug difficulty 2` | cold immune target |
| hit_minion | Hell Blood Moor's own unique pack: its minions are the targets (no spawn verb makes a minion) | minion class |
| merc | a hired merc (type 0) and three zombies | merc position and life hashed |
| pet | druid with a Spirit Wolf and a Raven, two zombies | pet position hashed |
| follow | druid with a wolf and a merc, the player walks to four goals, zombies chase | player position hashed (it moves), merc and pet follow |

Each hit scenario fires four skills one after another, each at a freshly spawned target, point blank (the player is warped 1.6 cells from the target, the cast aimed at its position): Ice Bolt 39, Ice Blast 45, Frost Nova 44, Glacial Spike 55, one skill point each so that the target survives; 8 blocks of 6 ticks after the cast, a snapshot per block. On the baseline (V) Ice Bolt, Ice Blast and Frost Nova leave a chill deadline (`chill=` in ms of game time) and no stun; only Glacial Spike carries a freeze: chill AND stun on every class.

Baseline target state at block 7 (V, base; hp now/max, chill and stun deadlines):

| class | Ice Bolt | Ice Blast | Frost Nova | Glacial Spike |
|---|---|---|---|---|
| plain | 120/125, chill 7560 | 140/149, chill 7080 | 106/133, chill 14000 | 95/119, chill 10240, stun 10240 |
| champion | 438/444, chill 7560 | 405/414, chill 7080 | 303/312, chill 13960 | 367/387, chill 10240, stun 10240 |
| unique | 694/700, chill 7560 | 711/720, chill 7080 | 494/520, chill 14000 | 543/560, chill 10240, stun 10240 |
| boss (Blood Raven) | 111/113, chill 4840 | 109/113, chill 5720 | 109/113, chill 10160 | 103/113, chill 10440, stun 10440 |
| immune (Hell zombie) | 5919/5919, none | 4360/4360, none | 5075/5075, none | 4495/4495, chill 9040, stun 10560 (the base freeze ignores the resist) |
| minion (Hell) | a neighbour minion: 3085/3088, chill 2240 | none reached | none reached | three minions: chill 8440, stun 9960 |

## Hashes (V; `mon` / `state`, first 12 hex, base; other columns say `same` or give the first 8 hex)

| scenario | base mon | base state | base2 | p155 | freeze2 | walk |
|---|---|---|---|---|---|---|
| passive | d1ca6c463f45 | 84942b4a0e26 | same | same | same | same |
| fight | 5f90630302ee | 70fa7405daed | same | same | same | mon 1124443e / state 9f5bc769 |
| cold | bba8362ccde9 | c8197bb82c7c | same | same | same | mon 5ef0cf79 / state 885a26fc |
| static | 34a1b9ef9f44 | cad5671002e0 | same | same | same | mon ef7c7c56 / state 7b5b5d1a |
| cold2 | bba8362ccde9 | c8197bb82c7c | same | same | same | mon 5ef0cf79 / state 885a26fc |
| hit_plain | 81b48ba4bdb0 | d75c22798784 | same | same | state 484c4acf (mon same) | mon 73e105ac / state 3d967f04 |
| hit_champion | 9b7c25f25f6a | 26c01e11198b | same | same | mon deb932a6 / state 71ed70d1 | mon bf38702e / state 20fe2c99 |
| hit_unique | 04f91fcf36ff | 865168e8ee37 | same | same | mon 5ae2cb7c / state dc53e794 | mon 3da3e322 / state aaa17e8a |
| hit_boss | d49d11125ccc | 0bcc4196a10e | same | same | mon 76af819b / state 6ea7eb8c | same |
| hit_immune | 2cf37f51eea2 | 65c6bbceea5d | same | same | mon b2ea5746 / state 3afaadd0 | same |
| hit_minion | 27714ca3de68 | 598e277f04e4 | same | same | mon d95c4585 / state e7885868 | mon 8a8d7fd4 / state 20a7d3a0 |
| merc | 1ecd7abc1cd7 | f993d75658e4 | same | same | same | mon e8e148ea / state 65cf8fad |
| pet | 99d912dcc2e9 | ac7ac7a0a49f | same | same | same | mon d9b0329f / state 6e1c877c |
| follow | 9f51307f5f85 | a583fd8df5d8 | same | same | same | mon 1beb4d60 / state c754ceaa |

The baseline run twice (base, base2) is identical in all 14 scenarios (V). The walk build reproduces the first run's walk3 hashes in fight, cold and static (1124443e, 5ef0cf79, ef7c7c56; V), as expected: e4a13181 and 8db11206 differ in comments only.

## Per candidate

### #155 head 8c56e96a (Static Field + cold chill): DET-CLEAN

Identical to upstream main in all 14 scenarios, both hashes. Coverage now (V): cold hits land on plain, champion, unique and boss targets (hp drops, a chill deadline is set, see the baseline table) and the chill commits leave all of it unchanged; the Static Field hit lands as before (zombie hp 6/8). Still NOT covered (U): the changed branch of the chill commit, a cold hit whose damage is 0 while the chill length is above 0. The cold-immune Hell zombie does not reach it (cold resist 100 cuts the chill length to 0 too: no chill in either build, V), and no spawnable Normal monster has a resist high enough to round a one-point Ice Bolt to 0 (the highest ordinary rows are 75 to 80, U: not tried each). The claim "no regression" is shown; "the changed branch behaves" is not by det.

### freeze rules v2, head 409903a6 (C5, C3, C4 on #155): DET-CHANGES-EXPLAINED

Only the Glacial Spike phases change, in every class: Ice Bolt, Ice Blast and Frost Nova (no freeze in them) and the old five scenarios are identical (V: compare lists the differing snapshots, all with `skill 55`; the first differing field is the chill or stun deadline, mode and position follow from a frozen or unfrozen monster). The state at block 7, Glacial Spike, base against freeze2 (V):

| class | base | freeze2 | the claim |
|---|---|---|---|
| plain (Normal) | chill 10240, stun 10240 | chill 0, stun 10240 | C5: plain frozen, not chilled. Holds |
| champion | chill 10240, stun 10240 | chill 10240, stun 0 | C3: chilled for the length, never frozen. Holds |
| unique | chill 10240, stun 10240 | chill 10240, stun 0 | C3. Holds |
| boss column (Blood Raven) | chill 10440, stun 10440 | chill 9440, stun 0 | C3, and C5's resist: the length is cut by her cold resist (1000 ms = 25 ticks shorter). Holds |
| cold immune Hell zombie | chill 9040, stun 10560 | none | C5: the freeze length is resisted like the chill length, 100 resist leaves 0. Holds |
| Hell minion pack | three minions chill 8440, stun 9960 | none | minions are plain (verifier: minions 0x10 freeze like plain), see note 1 |

Notes (not det failures; for the candidate's owner):

1. In Hell, plain monsters (the Hell zombie row and the minions) get neither freeze nor chill under v2, while a Normal plain monster is frozen (256 ticks, divisor 1). The claim text says a plain monster is frozen only at coldeffect < 0, so v2 follows the claim, but the Normal result freezes although the MonStats ColdEffect of Drowned Carcass reads 50 in the file we extracted, so the sign or column that `cold_effect` carries per difficulty is not what the claim text suggests. U (not traced here): whether `diff[Hell].cold_effect` is simply 0 because the Hell columns are blank in this data, which would make v2 never freeze a Nightmare or Hell plain monster. One print of `cold_effect` per difficulty settles it; it should be done before the PR text is final.
2. C4 (Blood Raven's and the Countess's freezes not halved) is NOT exercised: `debug spawn` makes plain monsters; `half_freeze` is set only by the superunique path of `make_boss` (no verb). The boss scenario row is Blood Raven's MonStats row without that flag.

### walk velocity, PR #156 head e4a13181: DET-CHANGES-EXPLAINED (now with merc and pet)

passive unchanged. fight, cold, cold2, static, the three zombie hit scenarios, hit_minion, merc, pet and follow change; hit_boss and hit_immune do not (U, not traced: Blood Raven stands and shoots, and the Hell zombie is watched for only 48 ticks per skill). V: a per-field classification of every differing line (scratch script, not kept) puts the FIRST difference in every scenario on a walking or running monster's, merc's or pet's position (fight/cold/cold2 first at block 3, static block 0, hit_plain/champion/unique at the `before` snapshot of the first skill, before any cast, merc at block 4 (position, then its fight: hp and mode), pet at block 1, follow at block 0). hp, mode, chill and stun differ later in the hit, merc, pet and follow scenarios as timing consequences (a zombie arrives at the bolt at a different tick: the chill deadline moves with it). In hit_minion the player's own cell also differs, only because the harness warps next to a minion that stood elsewhere (V: the warp targets differ). The player's position is identical in every other scenario, including follow (the player walks about with a merc and a wolf).

Explanation (S: claim commits d8e63d4a, a5cf6415, 455e366f, 8e5ff4cd): walking and running monsters move at MonStats Velocity with the walk percent a sum, a chilled one adds coldeffect, the merc at its velocity (at least 25 percent) and a pet likewise: the monsters, the merc (V: `dump merc` x/y differs) and the pet (V: `debug pets` `at=` differs). Not attributed per commit (U).

Not covered (U): the player `velocitypercent` commit (1c3fa4e5): no scenario gives the player a velocity source (the player's walk is identical in follow, as it should be without one). Two scenarios with a paladin and the Vigor and Holy Freeze auras were run and dropped: they were identical on all builds, but the aura was never confirmed active (`debug states` showed no buff, no monster chilled), so they would have shown "same" without testing anything (dead end).

## Verdicts

| candidate | head | verdict |
|---|---|---|
| #155 (Static Field + cold chill) | 8c56e96a | DET-CLEAN (changed zero-damage branch not reached, U) |
| freeze rules v2 | 409903a6 | DET-CHANGES-EXPLAINED (Glacial Spike only; open check in note 1; C4 not exercised) |
| walk velocity | e4a13181 | DET-CHANGES-EXPLAINED (walkers, merc, pet; player velocity commit not exercised) |

No unexplained change: no candidate is held back by this run.

## Still NOT covered

- A cold hit with damage 0 and a chill length above 0 (the chill commit's own branch); the player as the target of a cold missile.
- A superunique (Blood Raven's and the Countess's `half_freeze`): no spawn verb.
- A player with a velocity source (Vigor, Burst of Speed, FRW gear); the aura scenarios could not be confirmed active.
- Frozen Armor, Shiver Armor and Chilling Armor (their strikers' freeze path, `states.cpp`), Frozen Orb's own freeze, a merc or pet that is chilled, Nightmare difficulty, a second seed, other areas than Blood Moor, items.
- Per-commit attribution of the walk differences (the run is per candidate head, not per commit).

## Limits

- Baseline is upstream main plus the lab set, not pure upstream (upstream has no scripted-scenario harness); the lab set here carries the extra read-only `monstate` verb.
- 70 runs on one seed (3). The worker host was loaded (load average 20 to 30): runs took 30 to 90 s each; nothing depends on wall time (lockstep).
- Nothing was posted or pushed upstream. Scratch on the worker host: `pr/dethash2/` (trees, bin/, out-base/ with every log and results.txt).
