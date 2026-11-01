# PR proof: a NextHit hit marks the unit `justhit` (D-NEXTHIT, Nova / Frost Nova)

> status: SUPERSEDED on the end frame and the counts by [pr-nexthit-v2-addendum.md](pr-nexthit-v2-addendum.md): justhit ends after the missiles of frame + NextDelay (struck at f, next hit at f+NextDelay+1), candidate `fix/nexthit-v2`; the "0 of 100,198" below was built into the oracle on that point.
> status: BUILT, local candidate `fix/nexthit` on the worker host (claim 8abab0ed, tooling 7901fda2, on upstream main 72132b99); differential run before and after on real d2d builds; **NOT READY**: no independent verifier, no det state hashes, Chain Lightning and Lightning claims not built (designs below). Nothing pushed or posted (card nexthit, 2026-11-01). Follows docs/research/sk-lightning-trace.md (D1, D2), docs/research/sk-nova-trace.md (D1) and docs/research/sk-missile-trace.md. Standard: docs/process/upstream-pr-standard.md. Draft body: [pr-nexthit-draft.md](pr-nexthit-draft.md).

Tags: **V run** game.exe's own code run under upstream's emulator; **V read** disassembled with upstream's `tools/emu/disasm.py`; **D** 1.14d Missiles.txt read through `emu.mpq_file`; **S** d2d source at 72132b99 (file:line) or an earlier trace of ours; **I** inferred; **U** unknown.

## Claim (C1)

game.exe: when a missile whose Missiles.txt row has NextHit strikes a unit, FUN_005adb00 puts state 0x56 (`justhit`) on that unit until game frame + NextDelay (4 for `nova` and `frostnova`). The collision filter (0x5a8850 -> FUN_005a8730) and FUN_005adf10 skip a unit in that state for every NextHit missile, of any cast, and a LastCollide row also skips the unit it struck last (FUN_0064a470). FUN_005ae1f0 hands one unit a frame to FUN_005adf10, then the missile's frame is over.

d2d (S, components/game/fight/missiles.cpp:587-591, tick.cpp:394-411): a per-missile `hit_at` list, no unit state; every ring missile that touches a monster strikes it, and a missile goes on to the next touching monster in the same frame.

The commit: `Monster::just_hit_until` set by a NextHit hit for NextDelay frames and tested by every NextHit missile; `Missile::last_hit` for LastCollide rows (Missiles.txt `LastCollide`, now loaded); the frame loop stops a NextHit missile after its hit. Nothing else changes for rows without NextHit.

## Rule (V read, addresses)

| Address | What it does | Tag |
|---|---|---|
| FUN_005adf10 0x5adf84..0x5adfa6 | row +0x188 (NextHit) set: FUN_00639df0(unit, 0x56) not 0 -> return; FUN_0064a470(missile, unit) not 0 -> return; FUN_0064a400 last unit := unit | V read, V run |
| FUN_005adf10 0x5ae0e6..0x5ae0fe | FUN_005adb00, the hit event FUN_005c0c30, the SrvHitFunc; then the roll FUN_005a89a0 and the apply FUN_005adcd0 | V read, V run |
| FUN_005adb00 | row NextHit and a unit: stat list FUN_006251f0 ending at game +0xa8 + row +0x189, timer FUN_005417d0 (event 0xc at that frame), state 0x56 on it (FUN_006252d0), expiry callback 0x5adaf0 (FUN_00639db0(unit, state, 0)), FUN_00639db0(unit, 0x56, 1) | V read, V run (hooks below) |
| 0x5a8850 -> FUN_005a8730 | the CollideType filter: unit +0xc4 bits 3 and 2, NextHit row: not in state 0x56, not the last unit (FUN_0064a470), hostile (FUN_00554200) or CollideFriend | V read, V run |
| FUN_0064a470 / FUN_0064a400 | one last-unit slot (missile data +0x18 type / +0x1c id), compared only when row +4 bit 0 (LastCollide) | V read, V run |
| FUN_005ae1f0 0x5ae3a4..0x5ae3dd | FUN_00641cb0 finds one unit through that filter in a subtile the frame entered; found: FUN_005adf10(missile, unit, 0) and return | V read (the harness restates this loop) |
| `nova`, `frostnova` | LastCollide 1, NextHit 1, NextDelay 4, CollideKill 0, no ToHit, no pSrvHitFunc / pSrvDmgFunc (the oracle exits if any is set) | D |

## Differential (real d2d builds both sides)

- d2d side: `test_next_hit --next-hit SEED N` (new binary tests/game/test_next_hit.cpp; copy in tools/emu/next_hit_mode.inc): one Nova or Frost Nova through the REAL `Fight::fire` and `Fight::world` (flight, the frame's unit loop in tick.cpp, `skill_missile_hits`), on an open copy of the Blood Moor, 1..6 standing monsters (next_act held off), MonStats rows of size 1 / 2 / 3 (flyingscimitar, fallen1, bighead1), resistances cancelled, a million life, each missile a fixed 10 points (`Missile::fixed`), so life lost over a frame / 10 = that frame's hits. It prints the contacts (per frame, per live missile, the monsters in the subtiles it entered: `missile_struck`, monster order) and its own hits.
- game side: `tools/emu/next_hit.py` replays those contacts through game.exe's filter and FUN_005adf10 / FUN_005adb00 natively (hooks in its docstring: owner, hostility, stat getters, the state bit store, stat list and timer calls that record the end frame, the hit event, the damage roll that counts the hit, apply). The timer runs at the start of its frame.
- cases (seed 1): 198 boundary (2 skills x 3 sizes x distances 0..14 subtiles x 3 headings, skill level 1) and 100,000 random (skill, level 1..20, 1..6 monsters of random size within 12 subtiles). 32-bit xorshift in the C++ generator (`std::uint32_t`); the replay reads d2d's lines, so there is one generator.

| Build | per-monster hits differ (cases) | H lines differ | d2d hits per touched monster per cast | game.exe |
|---|---|---|---|---|
| before: upstream 72132b99 + the mode | **84,779 of 100,198** | 248,513 | Nova 4.43, Frost Nova 4.16 | 1.03, 1.04 |
| after: fix/nexthit 7901fda2 | **0 of 100,198** | 0 | Nova 1.03, Frost Nova 1.04 | same |

Point blank (boundary, distance 0, both skills): d2d 64 hits on a size-3 monster, 36 on size 2 (size 1 is not touched there: the ring starts past it); at distance 1 sizes 1 / 2 / 3 take 8.3 / 23.7 / 35.7; at 3 subtiles 3.0 / 8.3 / 12.0; game.exe 1.00 in every boundary bucket. Over the random packs by size (both skills): d2d 1.7-1.8 / 4.4-4.7 / 5.6-6.0, game.exe 1.00 / 1.00 / 1.09-1.11 (a size-3 monster can be reached again after its 4 frames). Summary: `python3 tools/emu/next_hit_summary.py d2d.txt game.txt`.

Contacts: the two builds print the same contact lines on the 2,198-case run; on the 100,198-case run 1,645 F lines differ (the strike count changes the frame's RNG draws; which d2d draw moves the contacts is not traced, I). Each build is compared with game.exe on its own contacts, so the after = 0 holds per build.

## Commands for the independent verifier (on the worker host, PR checkout `nexthit`)

```sh
cd <PR checkout>/nexthit                      # fix/nexthit, head 7901fda2
# build test_next_hit as for the other PRs (CMAKE_CXX_FLAGS -Wno-error=maybe-uninitialized -Wno-error=array-bounds)
D2_MPQ_DIR=<MPQ dir> D2_PATCH_INSTALLER=<patch> <build>/tests/test_next_hit --next-hit 1 100000 > d2d.txt
cd tools/emu && D2_MPQ_DIR=<MPQ dir> uv run python next_hit.py ../../d2d.txt > g.txt
grep -E '^[HT]' ../../d2d.txt | diff g.txt -        # empty after; before: a detached worktree at 72132b99 with the tooling commit's two files
```

## CI-class checks (fix/nexthit 7901fda2)

- pr_guard against upstream/main (72132b99, re-fetched 2026-11-01: still the tip): 0 FAIL, 0 WARN on the range and on each commit.
- presubmit --quick: 0 FAIL, 5 WARN (siblings unit kinds: players / mercs / pets are not struck through `skill_missile_hits`, the monster missiles' path does not read NextHit; siblings hit paths: melee and skills without missiles have no NextHit; docs-sync: no upstream .md states this rule except re/skills.md:928 "strike a unit again NextDelay frames on", which the PR should correct; build skipped by --quick; the amended run shows a fifth WARN not yet read), ruff with upstream rules clean (B023 fixed before the commit).
- gcc -Wshadow -Wunused -Wsign-compare build of test_game, test_rules, test_combat, test_protocol, test_next_hit: 0 warnings on added lines (MSVC-class shadowing sweep).
- clang-tidy (repo .clang-tidy, names and includes) on test_next_hit.cpp, missiles.cpp, tick.cpp, gamedata_load.cpp: 0 on touched lines after dropping an unused `<cmath>`.
- test_rules, test_combat, test_protocol, test_game (with the MPQ dir): rc 0. Tooling unit tests (tools/test_*.py, tools/emu/test_tooling.py): OK.
- game.exe md5 prefix 2b3600c10ee0; game-side output (after) md5 prefix 53bbb1d6cfc0.
- PR size: 7 files, 2 commits (claim 8abab0ed, tooling 7901fda2). test_game.cpp untouched (the mode is its own binary).

## Callers and what changes

`skill_missile_hits` (missiles.cpp:540) is the only reader of NextHit; its callers are the frame loop (tick.cpp:394-411) and nothing else. Rows with NextHit in 1.14d Missiles.txt (D): the novas, `chainlightning`, Twister / Tornado, Shock Field, the war cries, Blade Sentinel's row (d2d's stand-in). All now use the unit state and stop after one hit a frame; only `nova` and `frostnova` are in the differential. Blade Sentinel's turnaround clears `last_hit` as it cleared `hit_at`. Monster missiles (`missiles_update`'s foe path) do not read NextHit: unchanged.

## Limits

1. Covered: Nova and Frost Nova on standing monsters, one cast at a time. Not covered: two casts within 4 frames of each other, moving monsters, players / mercs / pets as targets, the other NextHit rows (Chain Lightning's hop missiles now cannot strike the just-struck monster for 4 frames, which is the game.exe rule, but its hop SEARCH still picks it: claim C2, not built).
2. Contacts come from d2d's flight (both sides see the same contacts); game.exe's own flight and FUN_00641cb0's order of units within a frame (subtile order) are not compared. The ring's flights are compared in docs/research/sk-nova-trace.md (896 real flights).
3. The frame a `justhit` ends is U: the harness runs the timer at the start of its frame (a hit at frame f can be repeated at f + 4), d2d does the same; the other reading (f + 5) is not run.
4. FUN_005ae1f0's loop (one unit a frame) is restated in the harness from the disassembly, not run.
5. Hooked: see the oracle's docstring; row +4 bits 1 / 2 are set from Missiles.txt `Explosion` / `Pierce` (S, the bin layout; both 0 for the novas).
6. Not run: independent verifier, det state hashes (need a d2d slot), Windows / MSVC build.

## Designs not built (C2 chain hop, C3 piercing Lightning)

- **C2 Chain Lightning hop search skips `justhit`** (FUN_0056b3e0 flag 0x80000, V read in sk-lightning-trace.md Q2). d2d change: one more test in the hop search (`now_ms < other.just_hit_until`: skip). Differential: the same replay with `chainlightning` rows plus a native run of FUN_005aa730 -> FUN_0056bd10 with stubbed rooms (FUN_00620bb0 / FUN_00619790 / FUN_0056b770 returning one room holding the case's units, as sk_tele_spot.py stubs rooms), printing the next target per hop; the d2d mode prints its hop targets from the same contacts. Blocker: the hop changes the contacts (a new missile), so the replay must run per hop with d2d's flight of the new missile; about half a day.
- **C3 piercing Lightning, one unit a frame, last unit skipped** (FUN_005ae1f0, FUN_0064a470, V read). d2d change: the same frame stop and `last_hit` for LastCollide rows without NextHit (`lightningbolt`), replacing `struck` for them. Differential: this harness with `lightningbolt` (CollideKill 0, LastCollide 1, no NextHit). Blocker: none technical; it changes every LastCollide row without NextHit (many), so it needs its own caller survey and det run; not started in this time box.

## Our bot's measurements

Every kill rate measured on d2d with Nova or Frost Nova is optimistic: on upstream main d2d lands 4.2-4.4 ring hits per touched monster per cast where game.exe lands about 1.04 (8 to 64 point blank). Soak and drill kill numbers for the Frost Nova crowd rule (`novaTick`, BUILDV2) and any Nova build are inflated until upstream or a fork patch fixes D-NEXTHIT; Chain Lightning numbers are inflated too (hop re-hits, sk-lightning-trace.md D1: 73 % of hops in the model), Lightning numbers slightly pessimistic (D2: game.exe about 10 % more hits in packs). Mark such numbers "d2d, D-NEXTHIT" until the fix runs.

## Shared-file lines for the merger

docs/BACKLOG.md, add: `- NEXTHIT: independent verifier + det state hashes for fix/nexthit (C1 novas: 84,779 -> 0 of 100,198); build C2 (Chain Lightning hop search skips justhit) and C3 (Lightning one unit a frame); propose the fork carry C1 (docs/research/pr-nexthit-proof.md)`

docs/INDEX.md, research section, add: `- docs/research/pr-nexthit-proof.md / pr-nexthit-draft.md: D-NEXTHIT (NextHit justhit unit state) local upstream candidate, Nova / Frost Nova differential (NEXTHIT)`

docs/decomp-queue.md, rows 17 and 18, append to the status cell: `; D1 built as local candidate fix/nexthit (NEXTHIT, docs/research/pr-nexthit-proof.md): novas 84,779 -> 0 of 100,198 cases`

docs/OWNERS (card nexthit): `docs/research/pr-nexthit-proof.md`, `docs/research/pr-nexthit-draft.md`, `tools/emu/next_hit.py`, `tools/emu/next_hit_mode.inc`, `tools/emu/next_hit_summary.py`, `openspec/changes/nexthit/**`; then `changecheck status -write`.

CHANGELOG.md: none (docs and tools only).

Glossary: `justhit` (state 0x56) is still not in GLOSSARY.md (proposed by SKLIGHT); not added.
