# Hydra, Fire Wall, the shard hit area and the real-build Blizzard test (decomp row 22, SKGROUND2)

> status: card skground2, 2026-11-01, offline (no game, no d2d slot, nothing pushed or posted). Finishes what [sk-ground-trace.md](sk-ground-trace.md) left: **Fire Wall** read and run (cast, makers, hits), **Hydra** read to the spawn (AI and missile not run), **D4** measured on a real d2d build against game.exe's own hit test (V run both sides), **D1** proven on a real d2d build and a local PR candidate built ([pr-blizzard-draft.md](pr-blizzard-draft.md), verifier next). The board item SKGROUND can be marked done for Blizzard, Meteor, Fire Wall and the hit areas; Hydra's AI and Meteor's blast shape stay as U lines (section 9).

Tags: **V read** disassembled by this agent with upstream's `tools/emu/disasm.py` (addresses given); **V run** game.exe's own code run natively under the emulator and compared with a REAL d2d build (upstream main 7d5b3a5c, a scratch checkout on the worker host, tests built Release, run with the patched 1.14d data); **D** the patched tables; **S** d2d source or an earlier trace; **I** inference or arithmetic; **U** not verified. Model note: this run was Sonnet; every PR-bearing number below is a V run with both sides running real code (D1, D4), and the independent Opus verifier is still to come.

## Answer in ten lines

1. **D1 on a real d2d build (V run):** `test_sk_ground --blizz` casts Blizzard through the real `Fight::fire` and `Fight::missile_tick`; 100,000 shards per seed: **14,926 (seed 1) and 14,957 (seed 7) fall on an offset game.exe never makes (the +6 row and column); after the one-line fix 0 and 0**; both builds make 25 shards on 25 frames a cast.
2. **PR candidate built (local):** `fix/blizzard-shard-offsets` = 9a1a90eb (tooling) + 681d52a9 (claim, `rng(2 (radius - 1)) - (radius - 1)`), pr_guard 0 FAIL / 0 WARN on the range and each commit, test_game and test_rules rc 0. Draft: docs/research/pr-blizzard-draft.md. Nothing pushed; an independent verifier is next.
3. **D4 measured (V run, game.exe's own FUN_005ae480 / FUN_005ae1f0 unit search against the real d2d `missile_tick`):** a standing Blizzard shard (Size 2) hits a size-1 monster on **5** subtiles (own + 4 neighbours) in game.exe and **21** in d2d (4.2x); size 2: 13 vs 21 (1.6x); size 3: 21 vs 21 (equal). The Fire Wall (Size 1) is thinner still: **1 / 5 / 9** subtiles (sizes 1 / 2 / 3) in game.exe against **21** in d2d for every size (21x, 4.2x, 2.3x).
4. **Fire Wall cast (V read, FUN_005c9ea0):** from the target point T, two `firewallmaker` rows (Vel 12, Range 7 + 2 lvl) aimed at T + (-dy, +dx) and T + (+dy, -dx) with (dx, dy) = caster - T, plus ONE `firewall` row (srvmissileb) at T itself. d2d's `kMissileWall` does the same.
5. **The wall (V run, 80 cases on the real build):** the maker drops a `firewall` (Range 90 frames, fixed) on every NEW subtile it enters, about 2 frames in 3 (the maker covers about half a subtile a frame) (game.exe: 2,360 drops; d2d one a frame: 4,296, the extra ones on the same subtile). **Every subtile game.exe's maker drops on gets a d2d firewall on the same frame (80 / 80 cases, one frame clock shift)**; d2d's maker is one tick short (the last drop missing).
6. **Fire Wall's damage (V run + V read):** a firewall hits every unit in its footprint **every frame** (30 of 30 frames at each hit offset; the run covered 30 of its 90 frames, I for the rest), no LastCollide skip (its row lacks the flag); `do 5` (FUN_005ae520) only animates the flame frame (+0x44), upstream's comment "do 5 burns whoever stands in it" is right in effect (the generic step hits) but the do is not the burner. Two overlapping firewalls on a size-2 or size-3 monster each hit it every frame (n missiles, n hits a frame, V run); d2d's `burned` list allows one burn a frame per monster per skill (S).
7. **DamageRate 41 (V read):** copied into stat 0x147 of the missile, into the hit record (+0x54) and used only to scale the DEFENDER's flat damage reduction (stats 0x22 / 0x23: x rate / 1024, 0x57c26a..0x57c29f): it is not a hit period. Fire damage ignores it (I); d2d does not read the column (S: no hit in components/).
8. **Hydra (V read, FUN_005c9220 start, FUN_005ca910 do 144):** the start check needs the target point inside a room with FUN_0061ab00 == 0; the do makes THREE hydras at the point + (-1, -1), (0, 0), (+1, -1) subtiles (x table 0x6e312c = -1, 0, 1; y table 0x6e3120 = -1, 0, -1, dumped), each through FUN_0056d940 with the monster class of Skills +0xbe (`hydra1`), lifetime FUN_004e6ca0 (Param1 250 frames, D) from the game frame, state / timer set after. d2d makes three at the cast subtile (one subtile apart is not modelled, I).
9. **Bot:** Fire Wall and Hydra have no row in `internal/combat/classdata/skills.json`: the bot cannot cast either (B6); the Blizzard / Meteor ground-aim findings of SKGROUND stand.
10. **Not done:** the hydra's own AI (FUN_005e9e60) and `HydraMissile` run, Meteor's blast shape (FUN_0056b7e0), where the player's `skilldelay` is refused (G3); the real-build FIX for D4 (needs a footprint call for standing missiles and per-missile burns, not one line).

## 1. Questions answered (Q1..)

| id | question | answer | where | tag |
|---|---|---|---|---|
| Q17 | What does Fire Wall's do 24 make? | FUN_005c9ea0(game ecx, unit edx, skill, level): Skills row valid, +0x48 srvmissilea valid, FUN_0056e450 must pass; the point: the target unit's (FUN_00553540) else the path's (FUN_00648a00 / FUN_00648a10); the point's room (FUN_00620bb0, FUN_00463740) and FUN_0061ab00 must be 0 or nothing is made (returns 0). Params {flags 0x21, owner, [4] row, [5] / [6] start = T, [7] / [8] aim, [0xb] skill, [0xc] level}: maker 1 aim (Tx - dy, Ty + dx), maker 2 aim (Tx + dy, Ty - dx), (dx, dy) = caster - T; then, when Skills +0x4a (srvmissileb = `firewall`) is valid, {flags 1, row, start T, no aim}. Returns 1 | 0x5c9ea0..0x5ca02b, calls 0x59fa30 x3 | V read |
| Q18 | What does the maker do each frame? | FUN_005ae680 (missile do 6, row 68 +0xc): owner and row valid, SubMissile1 (+0x18) >= 0, the path valid; when FUN_006505c0(path) says the maker has entered a new subtile it makes {flags 0x21, owner, row SubMissile1 = `firewall`, start = the maker's subtile, skill, level} through FUN_0059fa30, then FUN_005ae1f0 (the step). `firewallmaker`: Vel 12, MaxVel 12, Range 7, LevRange 2, Size 1, CollideType 8, LastCollide, SubMissile1 `firewall` | 0x5ae680..0x5ae77b; D (sk_ground_tbl.py tree 51) | V read + **V run** (Q19) |
| Q19 | Where and when do the firewalls land (real build)? | `sk_gnd2_fw.py` makes the two makers by the real FUN_0059fa30 with the Q17 params and steps them with row 68's own do, FUN_0059fa30 hooked after the maker exists to record each sub-missile's params; `test_sk_ground --fw` runs the real `Fight::fire` (`kMissileWall`) and records every `firewall` made. 80 cases (target and caster random, level 1..20): game.exe drops on 2,360 (trimmed) subtile entries; d2d drops one every frame (4,296); after trimming game.exe's frames to d2d's last, **every game.exe subtile is made by d2d on the same frame in 80 / 80 cases** (game.exe frame + 1: the harness counts the maker's first do as frame 1, d2d's first tick of a new missile is a frame after the cast); only-d2d subtiles 150 = T's own subtile (80, the srvmissileb one is not in the game-side harness) and subtiles past game.exe's walls (I: d2d's test level is open, the game's Blood Moor has walls) | `tools/emu/sk_gnd2_fw.py`, `sk_gnd2_fw_cmp.py`, `sk_gnd2_test.cpp --fw` | **V run** (both sides) |
| Q20 | Does a firewall hit every frame, and how wide? | Row 69: do 5, pSrvDmgFunc 3, Range 90, Size 1, CollideType 3, Collision 1, no LastCollide / NextHit. Standing missile, the unit search per frame: size-1 unit 1 subtile, size-2 unit the plus (5), size-3 unit 3 x 3 (9); a hit on every one of 30 frames; two or three overlapping missiles on a size-2 / size-3 unit give 2 / 3 hits every frame | `sk_gnd2_hit.py 7 firewall 30`, `sk_gnd2_fw2.py` | **V run** (game side) |
| Q21 | What is do 5 (FUN_005ae520)? | A per-frame animation choice for the flame: frame counter +0x44 >> 8 with row +0x181 (SubStart 12) and +0x182 (SubStop 36): at SubStart - 1 it jumps to a random frame in the loop (rand(SubStop - SubStart) + frame, the missile's seed FUN_0045c3e0); otherwise steps back 3 / 2 frames; no unit test of its own: FUN_005ae1f0 (the generic step) does the hits | 0x5ae520..0x5ae67c | V read |
| Q22 | What are pSrvDmgFunc 3 and DamageRate? | The dmg table (0x73c960 + 4 x n) entry 3 = FUN_005ad220: draws the missile's seed (&0x7f); when it is under row +0x78 (= dParam1 19 for `firewall`) the hit record gets flag 0x4000 (19 in 128, 14.8 % of the hits), which the post-hit FUN_0057cee0 tests at 0x57d04c to skip a branch (meaning U: a hit-reaction branch, not the damage, I). DamageRate: FUN_0059fa30 stores row +0x19c into stat 0x147 (0x59feb4), the hit builder FUN_005ad8a8 copies it to the record +0x54, FUN_0057c1f0..0x57c2a4 scales the defender's stats 0x22 / 0x23 (flat damage reduction) by record+0x54 / 1024 (FUN_00483360 with 0x400) | 0x5ad220..0x5ad27f, 0x57c26a..0x57c29f | V read; effect of 0x4000 U |
| Q23 | Where does a standing shard hit (real build)? | `sk_gnd2_hit.py` makes `blizzard1` (row 159) with the real FUN_0059fa30 beside a monster unit of the class and runs its do for the Range (9 frames), FUN_005adf10 hooked to record the unit; `test_sk_ground --hit` puts the same MonStats rows beside a real `Fight::launch`ed shard and counts life lost. Grid -7..7 x -7..7 subtiles. game.exe: size-1 class 190 (and 191) 5 offsets {own, 4 neighbours}; size-2 class 19 (63) the 13-subtile diamond |dx| + |dy| <= 2; size-3 class 10 (156) 21 (5 x 5 without the corners). d2d: 21 subtiles (hypot <= 0.5 cell = 2.5 subtiles) for every size. Classes 190 / 19 / 10 are MonStats rows maggotegg1 / fallen1 / bighead1 on both sides | `tools/emu/sk_gnd2_hit.py`, `sk_gnd2_hit_cmp.py`, `sk_gnd2_test.cpp --hit` | **V run** (both sides) |
| Q24 | The Hydra cast | FUN_005c9220 (srvstfunc 14): the point (FUN_0056d2c0) in a room (FUN_00463740) and FUN_0061ab00 == 0 (returns 1) else 0. FUN_005ca910 (do 144): Skills row, summon class +0xbe (`hydra1`) valid, FUN_0056e620(unit, skill, level) >= 0 (the pet-count check, not read), the point and room as above, owner redirect for a pet (FUN_0058f0d0), end frame = FUN_004e6ca0(skill, level) + game frame (+0xa8), then FOR i in 0..2: start = (x + [0x6e312c + 4 i], y + [0x6e3120 + 4 i]) = (-1, -1), (0, 0), (+1, -1) subtiles, FUN_00646ca0 (the monster level) and FUN_0056d940 make the hydra, FUN_0056deb0 and FUN_005c4470 set it up, a timer / end event at the end frame (FUN_005b0d70), the loop does not stop at the first success; returns 1 when any was made | 0x5c9220..0x5c927e, 0x5ca910..0x5cab30; tables dumped (`sk_ground_tbl.py dw 6e3120 8`) | V read + D |
| Q25 | Hydra and Fire Wall mana, delay | D: Fire Wall 22 / 1 / 8 shift, delay 35; Hydra 40 / 1 / 7, delay 40, Param1 250 frames; the pay and the delay are SKMISS / SKGROUND Q14 / Q15 (FUN_0056f7f0, FUN_0056ef90) | | D + S |

Dead ends: FUN_0056e620's role (pet count?) was not read; FUN_0056d940's use of the block's start field (the three offsets above) is read from the caller only. Frame numbering of the first tick of a new missile (game.exe vs d2d) is U: the Q19 shift of 1 is the harness's.

## 2. Formulas

```
Fire Wall cast (do 24):  T = target point; (dx, dy) = caster - T
   maker 1: row 68 at T aimed T + (-dy, +dx);  maker 2: row 68 at T aimed T + (+dy, -dx)
   firewall (row 69) at T                      (all three when the point is in a room and FUN_0061ab00 == 0)
maker (do 6, 7 + 2 lvl frames, Vel 12 x 75 % = 9):  on every new subtile: firewall at that subtile
firewall (do 5 + the generic step, Range 90, Size 1): every frame, every unit in the footprint
   (size 1 unit: its subtile; size 2: plus; size 3: 3 x 3), a hit each; DamageRate scales the
   defender's flat damage reduction only; 14.8 % of hits carry flag 0x4000 (reaction branch, U)
Hydra (do 144): three hydras at the point + (-1,-1), (0,0), (+1,-1) subtiles, Param1 = 250 frames
```

## 3. d2d (S, upstream main 7d5b3a5c)

- Fire Wall: `Fight::fire` kMissileWall (components/game/fight/missiles.cpp:359): both makers from the target spot along (T_y - C_y, C_x - T_x) and its negation, plus srvmissileb at the spot; `missile_tick` do 6 launches SubMissile1 at the maker's position EVERY frame; do 5: `burned` (monster, skill) once per frame for every monster within 0.5 cell (the same hypot test as the standing shard).
- Standing missiles' hit area: fight/tick.cpp:392-394 and missiles.cpp do 5: `hypot <= 0.5f` cells for every unit size; moving missiles use the real footprint (`missile_struck`, from the Nova port).
- Hydra: fight/summons.cpp:123-131: three through `spawn_spot` at radius -1 on one crowd snapshot (`others` is built before the loop, so the three heads can share the cast subtile), Param1 frames.

## 4. Deviations (d2d vs game.exe), new ones D9..

(D1..D8 are in sk-ground-trace.md section 4; D1 and D4 below get their real-build counts.)

| id | rule | game.exe | d2d | tag | reach | differential |
|---|---|---|---|---|---|---|
| D1 (proven) | Blizzard shard offsets | -6..5 | -6..6 | **V run both sides** (14,926 / 14,957 of 100,000 shards, 0 / 0 after the fix) | every Blizzard; 14.9 % of d2d's shards on a row / column game.exe never uses | built: `sk_ground_blizz.py`, `sk_gnd2_test.cpp --blizz`; PR candidate |
| D4 (proven) | standing missile hit area | size-2 missile (shard): units of size 1 / 2 / 3 hit on 5 / 13 / 21 subtiles; Size-1 missile (Fire Wall): 1 / 5 / 9 | 21 subtiles for every size | **V run both sides** (Q23) | Blizzard damage per cast on small monsters about 4x game.exe; Fire Wall coverage 21x (size 1) to 2.3x (size 3) the area | built: `sk_gnd2_hit.py`, `sk_gnd2_hit_cmp.py`, `--hit`; fix not built (a footprint call for standing missiles in tick.cpp and do 5) |
| D9 | Fire Wall burns per missile | each overlapping firewall hits the unit every frame (n hits) | one burn a frame per monster and skill (`burned`) | V run (game) + S (d2d) | big monsters (size 2 / 3) standing on the wall take 2 to 3x the hits per frame in game.exe; with D4 the d2d wall is a wide band that hits once a frame | game side built (`sk_gnd2_fw2.py`); d2d side: a test with two firewalls on one monster (not built) |
| D10 | Fire Wall maker drops | one firewall per NEW subtile | one per frame (a duplicate on the same subtile about every third frame), the maker one tick short | V run both sides (Q19) | the duplicates are harmless while `burned` dedups (D9 aside); the missing last drop is one subtile at the wall's far ends | built (`sk_gnd2_fw_cmp.py`) |
| D11 | Hydra heads | three at the point + (-1,-1), (0,0), (+1,-1) | three at the cast subtile (one crowd snapshot) | V read (game) + S | heads stand 1 subtile apart in game.exe, same spot on d2d: no range effect (I) | not built |
| D12 | DamageRate | scales the defender's flat damage reduction (stats 0x22 / 0x23) by rate / 1024 per hit | column not read | V read + S | none for fire; Fire Wall on a monster with flat physical / magic reduction only (none of Fire Wall's) | not built |
| D13 | do 5 comment | animation only; hits come from the generic step | comment says do 5 burns | V read | doc only: upstream's fight.hpp:906 and docs/research/re/skills.md | none |

## 5. The differentials (V run)

Commands (the worker host; the card's helper copies the script into the scratch upstream checkout, `D2_MPQ_DIR` = the patched 1.14d data folder):

```sh
# game side
tools/emu/sk_ground_run.sh <upstream> sk_ground_blizz.py gen 40000 7 > b7.cases
tools/emu/sk_ground_run.sh <upstream> sk_ground_blizz.py game b7.cases > b7.game
tools/emu/sk_ground_run.sh <upstream> sk_gnd2_hit.py 7 blizzard1 9 > hit_bz.game       # D4, shard
tools/emu/sk_ground_run.sh <upstream> sk_gnd2_hit.py 7 firewall 30 > hit_fw.game       # D4, fire wall
tools/emu/sk_ground_run.sh <upstream> sk_gnd2_fw.py 80 1 > fw.game; grep '^C' fw.game > fw.cases
tools/emu/sk_ground_run.sh <upstream> sk_gnd2_fw2.py                                    # D9
# d2d side: the real build of tools/emu/sk_gnd2_test.cpp (tests/game/test_sk_ground.cpp upstream-side)
test_sk_ground --blizz 7 4000 | grep '^S' > d2d.s
test_sk_ground --hit blizzard1 9 | grep '^H' > hit_bz.d2d
test_sk_ground --hit firewall 30 | grep '^H' > hit_fw.d2d
test_sk_ground --fw fw.cases | grep '^D' > fw.d2d
# compare
python sk_gnd2_blizz_d2d.py b7.cases b7.game d2d.s       # (the PR tree's sk_ground_blizz.py carries the same `d2d` mode)
python sk_gnd2_hit_cmp.py hit_bz.game hit_bz.d2d ; python sk_gnd2_hit_cmp.py hit_fw.game hit_fw.d2d
python sk_gnd2_fw_cmp.py fw.game fw.d2d 1 trim
```

| differential | cases | result |
|---|---|---|
| D1 before (main + test) seed 1 / 7 | 100,000 shards each | 14,926 / 14,957 shards outside -6..5; d2d offsets -6..6 |
| D1 after (candidate) seed 1 / 7 | 100,000 shards each | 0 / 0; d2d offsets -6..5 |
| D4 shard, sizes 1 / 2 / 3 | 225 offsets x 3 classes | offsets only d2d hits: 16 / 8 / 0 (game.exe 5 / 13 / 21 of d2d's 21); the 21 / 13 / 5 shared offsets hit once in both |
| D4 fire wall, sizes 1 / 2 / 3 | 225 offsets x 3 classes, 30 frames | only d2d 20 / 16 / 12 (game.exe 1 / 5 / 9 of 21); shared offsets: game.exe 30 of 30 frames, d2d 29 (the placing frame) |
| Fire Wall makers | 80 cases | every game.exe subtile on the same frame in 80 / 80; game-only 0; d2d-only 150 (see Q19) |
| overlapping firewalls (game) | 3 classes x 1..3 missiles | size 1: 1 hit a frame whatever n; size 2 / 3: n |

Md5 of the case files are not recorded (the cases come from the seeds above). Not covered: the collision check of a shard spot (hooked free), the maker's wall stops, a d2d run with two firewalls on one monster (D9's d2d half is read, S), a live game.

## 6. The bot (described, not changed)

| id | finding | file:line | tag | impact |
|---|---|---|---|---|
| B6 | Fire Wall and Hydra have no row in the bot's skill table, so no skill kind (`ground`, `summon`) is known for them: the bot never casts either; `Hydra` appears only in the fire option list | internal/combat/classdata/skills.json (no entry), internal/combat/rules.go:229, internal/combat/castgap.go:22 (Fire Wall only in the cast-time set) | S (our code) | nothing today; a build that wants them needs rows first |
| B7 | if Fire Wall is ever added: game.exe's wall is a 1-subtile line across the cast line at the target point; a size-1 monster is hit only when it stands on a wall subtile, so aiming at the monster's spot (as `KindGround` does) puts one firewall under it and the maker lines beside it; d2d's wall is about 5 subtiles thick, so d2d soaks over-rate it | internal/combat/skills.go:29 | I (from D4 / Q20) | design note |
| B8 | the d2d backend's Blizzard has the D1 pattern (centred 13 x 13) and the D4 area (about 4x on size-1 monsters): read END-build soak kills on d2d as an upper bound until D1 / D4 are fixed upstream | internal/d2denv | I | as B5 of SKGROUND |

## 7. PR candidate (D1): built, local, verifier next

`fix/blizzard-shard-offsets` in the scratch checkout of upstream main 7d5b3a5c on the worker host (local, not pushed, author identity the numeric-ID noreply, trailer `Agent: Claude Code`): **9a1a90eb** tooling (`tests/game/test_sk_ground.cpp --blizz`, `tools/emu/sk_ground_blizz.py` with the `d2d` mode) and **681d52a9** the claim (`missile_tick`: `reach = radius - 1; rng(2 * reach) - reach`; a header comment; one line of docs/research/re/skills.md). Differential before 14,926 / 14,957 of 100,000, after 0 / 0 (section 5); pr_guard 0 FAIL / 0 WARN on the range and on each commit; test_game and test_rules rc 0; the whole upstream test suite and clang-tidy were not run. Draft text: docs/research/pr-blizzard-draft.md. Next: an independent verifier (d2-reviewer, own harness, not these scripts), then the user's go; nothing is posted. D4 is NOT a candidate here: the difference is proven but the fix is not one claim (a footprint call for standing missiles, and for Fire Wall one burn per missile instead of `burned`); a follow-up card could take D4 + D9 as one claim with `sk_gnd2_hit.py` as its differential.

## 8. Drill recipe (lab binary, fork verbs; not run here)

Verbs and limits as sk-ground-trace.md section 8 (`debug missiles`, `debug spawn`, `step N`, `cmd skill <id> X Y`).

- G1b (D1 after the fix reaches the lab binary): Blizzard at an empty point, `debug missiles` each frame: offsets -6..5 only.
- G6 (D4, Fire Wall): one zombie (size 1) parked on and beside the wall's line after a Fire Wall cast; count its hp drops per frame: game.exe only when it stands on a wall subtile (1 subtile, every frame); d2d within 2.5 subtiles of any firewall.
- G7 (D9): a size-2 monster (fallen) standing over the wall's diagonal: hits per frame 2 to 3 in game.exe (per missile), 1 on d2d.
- G8 (Q24): Hydra cast at an empty point: three `hydra1` units; `debug monsters` positions: the point + (-1,-1), (0,0), (+1,-1) subtiles in game.exe, one subtile on d2d.
- G9 (Q22): count the fraction of Fire Wall hits that carry no stagger on a monster that staggers (flag 0x4000, 14.8 %): a long Fire Wall soak on a stagger-prone monster with hp pinned (needs `debug mon set`, missing verb).

## 9. Matrix cell updates (docs/research/sorc-skill-matrix.md; proposed text for the merger)

| Row | Cell | New value |
|---|---|---|
| 51 Fire Wall (line 43) | behaviour | V read + V run (sk-ground2-trace.md): two makers across the cast line + one firewall at the point; a firewall per new maker subtile, each hits every unit in its 1 / 5 / 9-subtile footprint every frame for 90 frames; d2d: wall placement equal, hit area 21 subtiles (D4), one burn a frame (D9) |
| 62 Hydra (line 54) | behaviour | V read: three hydras at the point + (-1,-1), (0,0), (+1,-1), 250 frames; AI and HydraMissile not run (U) |
| 59 Blizzard (line 49) and section 3.13 | shard offsets, hit area | real build: d2d offsets -6..6 (D1, fixed on a local branch), hit area 5 / 13 / 21 vs d2d 21 subtiles (D4) |
| rank 9 SKGROUND (line 255) | status | done for Blizzard, Meteor, Fire Wall, the hit areas; open: Hydra AI, Meteor blast shape (FUN_0056b7e0), delay refusal (G3) |

## 10. Not traced (U)

- The hydra's AI (FUN_005e9e60) and `HydraMissile` (rate, damage) under game.exe's code; FUN_0056e620 and FUN_0056d940's use of the start block.
- The meaning of hit flag 0x4000 (Fire Wall's 14.8 %) beyond "tested in FUN_0057cee0".
- Fire Wall's later frames (the run covered 30 of the 90) and the maker stopping at walls.
- Meteor's blast shape (FUN_0056b7e0), the skill-delay refusal (G3), Frozen Orb's y tables (sk-ground-trace.md section 10).

## Shared-file lines for the merger

docs/decomp-queue.md, row 22 (replace the status cell): `done for Blizzard, Meteor, Fire Wall (SKGROUND 2026-11-01 and SKGROUND2, docs/research/sk-ground-trace.md, sk-ground2-trace.md); Hydra AI, Meteor blast shape open`; and append to the finished list:

```
- 22b ground skills 2 (SKGROUND2, 2026-11-01): Fire Wall do 24 (two makers across the cast line + one firewall at the point), maker drops on every new subtile (V run vs the real d2d build, 80 / 80 cases), firewall hits every unit in its Size-1 footprint every frame (1 / 5 / 9 subtiles for size 1 / 2 / 3, V run), overlapping firewalls each hit (D9); Hydra do 144 makes three at the point + (-1,-1), (0,0), (+1,-1) (V read); D4 standing-missile hit area proven on a real d2d build (shard 5 / 13 / 21 vs d2d 21 subtiles); D1 proven on a real build (14,926 / 14,957 of 100,000 shards off game.exe's grid, 0 / 0 after one line); LOCAL PR candidate fix/blizzard-shard-offsets (9a1a90eb tooling, 681d52a9 claim), verifier next; docs/research/sk-ground2-trace.md, docs/research/pr-blizzard-draft.md
```

docs/BACKLOG.md (fix queue):

```
- SKGROUND2 D1: PR candidate fix/blizzard-shard-offsets (local, upstream 7d5b3a5c: 9a1a90eb tooling, 681d52a9 claim; before 14,926 / 14,957 of 100,000, after 0 / 0; pr_guard 0/0); needs the independent verifier, then the user's go (docs/research/pr-blizzard-draft.md)
- SKGROUND2 D4 + D9: standing missiles hit every monster within 0.5 cell on d2d (shard 5 / 13 / 21 subtiles in game.exe, Fire Wall 1 / 5 / 9, d2d 21); Fire Wall burns once a frame per monster on d2d, once per overlapping missile in game.exe; differentials tools/emu/sk_gnd2_hit.py + sk_gnd2_fw2.py; fix = one footprint call plus per-missile burns (a follow-up card)
- SKGROUND2 open: hydra AI and HydraMissile, Meteor blast shape, hit flag 0x4000 (14.8 % of Fire Wall hits), delay refusal (G3)
- SKGROUND2 B6: Fire Wall and Hydra have no row in internal/combat/classdata/skills.json (the bot cannot cast them)
```

docs/INDEX.md (research list):

```
- docs/research/sk-ground2-trace.md: Fire Wall (cast, makers, hits V run), Hydra spawn, standing-missile hit area D4 and Blizzard offsets D1 on a real d2d build; docs/research/pr-blizzard-draft.md: the local PR candidate
```

docs/research/README.md (if it lists traces): `sk-ground2-trace.md - SKGROUND2, decomp row 22b (Fire Wall, Hydra spawn, D1 / D4 real-build); pr-blizzard-draft.md - draft for the Blizzard offset candidate`.

Community board item SKGROUND (issue 4): `done 2026-11-01 (Blizzard, Meteor, Fire Wall; D1 and D4 proven on a real d2d build; local PR candidate fix/blizzard-shard-offsets awaiting an independent verifier); open U lines: hydra AI, Meteor blast shape, skill-delay refusal (G3). It can be marked done; the follow-ups are in BACKLOG`.

CHANGELOG.md: none (docs and tooling only).

Glossary (proposed, not written): none new beyond sk-ground-trace.md's "skill delay (skilldelay)"; "firewall maker": the `firewallmaker` missile, the two projectiles Fire Wall launches across the cast line, each dropping a `firewall` on every new subtile.
