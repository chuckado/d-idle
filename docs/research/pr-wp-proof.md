# PR candidate proof: the waypoint refusal window counts from hostility (agent DECOMPWP, 2026-10-31)

> status: proof for a LOCAL candidate; nothing pushed or posted. Trace: [waypoint-trace.md](waypoint-trace.md). Standard: docs/process/upstream-pr-standard.md.

## Claim

game.exe 1.14d refuses a waypoint travel (C->S 0x49) for 10 s after the player's last **hostile declaration**; d2d refuses it for 10 s after **every arrival on a level**. One claim, one commit.

## Addresses (strongest first)

| Link | Evidence | Tag |
|---|---|---|
| FUN_0054c5d0 | `call FUN_0055b6c0; add eax, 0x2710; call [GetTickCount]; cmp; jbe` -> refuse: voice event 0x13 (FUN_00553380), FUN_00554190, return 1 | V (read and run) |
| FUN_0055b6c0 | second unit 0: `mov eax, [pcdata + 0x160]` (pcdata = FUN_006221a0(player)) | V |
| pcdata +0x160 writers | every `mov [reg + 0x160], x` in game.exe's executable sections (upstream `disasm.py --all`): 0x421e77, 0x4b83d5, 0x4b83fb (client code, not player data) and 0x55b76b (FUN_0055b720, `= GetTickCount()` on FUN_006221a0's result) | V (exhaustive scan) |
| FUN_0055b720 callers | one: 0x5a5f51 in FUN_005a5e50, the hostile declaration (both players stat 0xc >= 9; per-pair timer now + 60,000 ms) | V (scan) |
| FUN_00584870 | the town portal reads the same field + 0x1388 (5 s): upstream objects.md already names it "the player's hostility time ... not in single player" | V, S |
| d2d | components/game/world/commands.cpp:267 `now - actor->arrived_at < 10000`; arrived_at set in World::swap_npcs (world/setup.cpp:363) on every arrival; world.hpp:209 calls it pcdata+0x160 | V-d2d |

## Differential

`tools/emu/wp_refuse.py` runs game.exe's dispatcher FUN_0054d750, the handler FUN_0054c5d0, FUN_0055b6c0, FUN_00554190 and the setter FUN_0055b720 natively (unicorn, upstream's emu.py); hooks: GetTickCount (the case clock), FUN_006221a0, FUN_00549570 (validation result), FUN_00584f60 (travel recorded), FUN_00553380 (voice recorded). Level changes run nothing in game.exe because nothing there writes the field (the scan above). Cases: exhaustive grid 201 (ms since the last level change 0..20,000 step 100) x 82 (hostile: never, or 0..20,000 step 250) x 2 (target active) = 32,964, plus 100,000 random timelines (start tick 70,000..2^31, 1-3 level changes, 0-2 hostile declarations, 10 % inactive targets): 132,964 cases.

| Run | Cases | Same | Differ |
|---|---|---|---|
| d2d today (`wp_refuse.py 100000 7`) | 132,964 | 73,655 | 59,309 (47,157 d2d refuses within 10 s of a level change where game.exe travels; 12,152 game.exe refuses within 10 s of a hostile declaration, d2d has no hostility) |
| proposed rule (`--after`) | 132,964 | 132,964 | 0 |

The `--after` model keeps game.exe's exact compare on a hostile timestamp; d2d has no hostile declaration, so the patch drops the test (the second class cannot occur in d2d). Limits: GetTickCount below 10,000 (first 10 s of machine uptime, zeroed player data, I) and the 2^32 wrap are outside the cases.
