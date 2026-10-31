# PR candidate proof: the waypoint refusal window counts from hostility (agents DECOMPWP, WPPR, 2026-10-31)

> status: proof for a LOCAL candidate (card wppr); nothing pushed or posted. Trace: [waypoint-trace.md](waypoint-trace.md). Standard: docs/process/upstream-pr-standard.md. Draft: [pr-wp-draft.md](pr-wp-draft.md). Pack: [pr-wp-evidence-pack.md](pr-wp-evidence-pack.md).

Branch `fix/waypoint-hostile-window` in `~/d2d-lab/pr/wp` on the worker host, two commits on upstream/main d0ea14b7: `74199f4f` (tooling: `tools/emu/wp_refuse.py`, `tests/game/test_game.cpp --waypoint-refuse`, a `tools/emu/README.md` row) and `4724e138` (the claim). 9 files, +199 / -9. The earlier one-commit candidate (fbdd9f57 on 03f42254, Python after-run only) is kept as local branch `wp-pre-rebase-fbdd9f57`. Local only.

## Claim (one, one commit)

game.exe 1.14d refuses a waypoint travel (C->S 0x49) only while `GetTickCount() < player data +0x160 + 10000`, and the only code that writes player data +0x160 is the hostile (PvP) declaration; in a single-player game it never fires. d2d refused for 10 s after every arrival on a level. The patch removes, and changes nothing else:

- `components/game/world/commands.cpp`: the `now - actor->arrived_at < 10000` test of the 0x49 command (the other refusals stay: not a waypoint object, the current level, an inactive target, a joined game);
- `components/game/world.hpp`: the `arrived_at` field (its only reader was that test);
- `components/game/world/setup.cpp`: its stamp in `World::swap_npcs`;
- `tests/pad_use_d2d.py`: the 10 s sleep before the waypoint use; `docs/research/re/act1-end.md` and `waypoint.md`: what the field is.

`// ponytail:` at the change: hostile declarations are not modelled (so nothing refuses); an act change arrives at once (game.exe's load not traced).

## The chain, tagged

| Link | Evidence | Tag |
|---|---|---|
| FUN_0054c5d0 | length 9; `FUN_0055b6c0(player, edx=0)`, `add eax, 0x2710`, `GetTickCount`, `cmp; jbe` -> else voice 0x13 (FUN_00553380), FUN_00554190, return 1 | V (read; run in the oracle; re-read by the verifier) |
| FUN_0055b6c0 | edx 0: player data +0x160; edx another unit: the hostile list at +0x38, entry +8. Callers: 0x54c5fc (waypoint, 0), 0x5848bc (town portal, 0, +0x1388), 0x5a5f13 (hostile path) | V (verifier) |
| player data +0x160 writers | one: 0x55b76b in FUN_0055b720 (`= GetTickCount()`); every other `0x160]` write in the executable sections is a client struct (0x421e77, 0x4b83d5, 0x4b83fb); player data is allocated by FUN_00621f90, 0x16c bytes, memset 0 (0x621fec), +0x160 not set there | V (whole-section sweeps by the author and the verifier) |
| FUN_0055b720 callers | one rel32 call, 0x5a5f51 in FUN_005a5e50; no jump or pointer to it | V (verifier) |
| FUN_005a5e50 | the hostile declaration: slot 4 of the party-action table 0x6e2218 (FUN_005a6000), from the C->S 0x5d / 0x5e handlers; needs a second player, both clvl >= 9 | V (verifier) |
| FUN_00584870 | the town portal reads the same field + 5 s; upstream objects.md already calls it the hostility time | V, S |
| d2d | commands.cpp:267 (upstream d0ea14b7) refused on `arrived_at`, set in swap_npcs (setup.cpp:363) on every arrival | V-d2d |

No U link.

## The differential: d2d's C++ against game.exe (V)

`wp_refuse.py game` runs game.exe's FUN_0054d750 -> FUN_0054c5d0 -> FUN_0055b6c0 / FUN_00554190 natively and FUN_0055b720 for each hostile declaration; hooks (docstring): GetTickCount, FUN_006221a0, FUN_00549570 (active or not), FUN_00584f60 / FUN_00553380 (recorded), the error path's FUN_00554d00 / FUN_00552f60, FUN_00410a80. `test_game --waypoint-refuse` runs d2d's own `World::apply(cmd::Waypoint)` in the camp to the Cold Plains, each level change a real `World::arrive` at its time, the target's bit set or cleared per case; hostile declarations are not run (d2d has none). Joined by `diff`.

| Set | Cases | Before (74199f4f) | After (4724e138) |
|---|---|---|---|
| single player, `gen 100000 7 --solo`: grid 0..20,000 ms step 1 x active (40,002) + 100,000 random timelines (1-3 level changes) | 140,002 | 66,969 (all: d2d refuses < 10 s after a level change, game.exe travels) | **0**, md5 e38fb142... equal |
| DECOMPWP's set, `gen 100000 7` (with hostile declarations) | 132,964 | 59,309 (47,157 level-change class + 12,152 hostile class: the Python model's numbers, now from the C++) | 30,763, every one game.exe refusing < 10 s after a hostile declaration (outside d2d: no hostility) |

The second row is why the claim is single player: with a hostile declaration in the case, game.exe refuses and d2d (before and after) has nothing to refuse on; before, 18,611 of those cases agreed only because d2d also refused for the wrong reason (a recent level change).

## Checks on 4724e138

GCC 16, Release, `-Wno-error=maybe-uninitialized -Wno-error=array-bounds`, full build (the d2d app target builds). `test_game` rc 0 (19 OK, MPQs + LODPatch_114d.exe, incl. the new "a waypoint travels at once after a level change, not to its own level or an inactive one"), `test_combat`, `test_rules`, `test_protocol` rc 0. `ctest -j4`: 54 of 58 pass, failures test_mpq, smoke_d2d, quests_act1, quests_act4 (test_iso9660 skipped), the same four on an unmodified upstream build (d0ea14b7, same flags); pad_use_d2d passes on both (on the branch without its 10 s sleep). `pr_guard.py upstream/main` 0 FAIL 0 WARN; `presubmit.py --quick` 0 FAIL, ruff run (0.8.6, clean), 3 WARN answered in the draft (siblings, "never", build by hand). Test merges with the open #149 (0dc3e0d3) and #152 (70e54f57) heads: no conflict (`git merge-tree`).

## Callers (what moved)

`World::deal`'s 0x49 branch: travels at once after an arrival (the claim). `World::swap_npcs` (callers: arrive, the edge crossing in travel.cpp, respawn, setup): no longer stamps a time; nothing else read `arrived_at` (grep over components, apps, tests, tools). Siblings: only players use waypoints; mercs, pets and monsters do not.

## Verifier

Independent second agent (Opus, fresh brief, read-only on the branch), report verbatim:

> VERDICT: holds, with limits
>
> 1. V: FUN_0054c5d0. It returns 3 when the length is not 9. It calls FUN_0055b6c0(player, edx=0), adds 0x2710 and compares the sum with GetTickCount using an unsigned test. While the sum is greater than the tick it plays voice 0x13, calls FUN_00554190 and returns 1. With edx=0, FUN_0055b6c0 returns player data +0x160. With edx non-zero it walks the hostile list at player data +0x38 and returns the matching entry's +8, or 0 when edx is the player itself. Its callers: 0x54c5fc (waypoint, edx=0), 0x5848bc (town portal, edx=0, 5 s window +0x1388, read only), 0x5a5f13 (inside the hostile path, edx=other player).
> 2. V: I swept every executable section for any instruction touching `0x160]`. Besides 0x55b76b, the only non-stack writes are 0x421e77 (zeroes fields in a client C++ constructor) and 0x4b83d5/0x4b83fb (a client table with 0x110 stride). Neither is player data. Indexed writes near the field (`[ecx+eax*4+0x114]` at 0x588xxx) work on unit+0x18, an object's data with a counted array, not player data. There is no `lea reg,[reg+0x140..0x16f]`. Player data is allocated at FUN_00621f90: 0x16c bytes, memset to 0 at 0x621fec, then only +0, +0x10..+0x24, +0x2c and +0x4c are set. That is the only 0x16c-sized memset or bulk op.
> 3. V: The only rel32 call or jmp to FUN_0055b720 is 0x5a5f51. No jcc reaches it and no absolute pointer to it exists. FUN_005a5e50 is reached only through the party-action table at 0x6e2218 (slot 4), which FUN_005a6000 dispatches. Its callers are the C->S 0x5d and 0x5e handlers (0x54c2a0 and 0x54c300). Those handlers first refuse when game-info +8 bit 16 is set (FUN_0053fd10). FUN_005a5e50 also requires both units at clvl>=9 and a room check (FUN_0061ab00).
> 4. V: FUN_00549570 checks that the object exists, then the same act, then range (10, or 0x16 when unit +4 == 1), then the level id, then the waypoint bit. The dispatcher's busy test FUN_0057eec0 drops the packet in mode 0 or 0x11 or with state 0x36. FUN_00584f60 has no tick read. None of these involves +0x160 or a level change.
> 5. V: Two differential runs, both with no hostile declarations.
>    - My generator `adv_gen.py 20000 314159` (20,372 cases: up to 50 level changes, last arrival 0/1/9998/9999/10000/10001 ms before, ticks near 2^31 and 2^32-1, inactive targets). After binary: 0 mismatches for tick >= 10000 (19,231 cases). All 951 mismatches have tick < 10000. Before binary: 9,696 mismatches with tick >= 10000. md5s: cases 6756024f…, game be5807bf…, after 413ccaef…, before a27e932a….
>    - The author's `gen 50000 271828 --solo` (90,002 cases): after binary 0 mismatches, md5 20ddfb7f… identical to game. Before binary 38,549 mismatches.
> 6. V: The client waypoint-row click (0x49d0f1 → FUN_004786a0) has no timer. FUN_00584f60 has none either.
>
> Limits to disclose:
> - game.exe refuses an active travel while GetTickCount < 10000, because the field is zeroed (the first 10 s of machine uptime). The patched d2d does not refuse there.
> - A crafted 0x5d naming the player's own unit id, with the hostile flag set and clvl >= 9, reaches the store unless the FUN_0053fd10 game-flag check (bit 0x10000, meaning not identified) refuses it. The stock client never sends this, and I did not run it.
> - d2d lacks the target check's range and same-act tests and the dead/state-0x36 drop. These are existing gaps, not part of this claim.
> - Hostile declarations are not modelled in d2d.
>
> Not checked:
> - A write through a pointer to player data at a base other than +0x160 that is computed and then offset (the sweep is linear, so data islands could hide code).
> - The client's object-operate path that opens the panel.
> - Anything run live in the game.
>
> I left nothing running. Scratch on the worker host in ~/d2d-lab/pr/wp-verify/: adv_gen.py, adv*.txt and solo*.txt. I deleted the full disassembly dump.

## Limits

- GetTickCount below 10,000 (the first 10 s of machine uptime): game.exe's zeroed field refuses there, patched d2d does not (the verifier's 951 mismatches, all there); outside the generator (ticks 70,000 .. 2^31).
- Hostile (PvP) declarations and the joined path are not modelled in d2d; a crafted self-targeted 0x5d (verifier, not run) could stamp the field in game.exe.
- Existing d2d gaps unchanged by this PR: the target check's range and same-act tests, the busy / dead / state 0x36 drop of the dispatcher.
- Cross-act load time inferred, not traced (ponytail). Writes through a computed base, and code hidden in data islands, are outside the linear sweeps.

**PROOF: complete** (C++ differential 66,969 -> 0 over 140,002 single-player cases, md5 equal; verifier holds over 110,374 more cases outside the author's generator, with the limits above; chain V/S; checks green).
