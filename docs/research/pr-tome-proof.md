# Proof: a Town Portal read from the pack opens at once, in any live mode (agent DECOMPTOME, 2026-10-31)

> status: candidate, PROOF incomplete (section 4). Trace and tags: [town-portal-trace.md](town-portal-trace.md). Standard: [../process/upstream-pr-standard.md](../process/upstream-pr-standard.md). Nothing pushed, posted or filed; the local upstream branch `trace/tome` (worker host) has no commits.

## 1. The claim

Using a Scroll of Town Portal (tsc) or a Tome of Town Portal (tbk) with charges from the inventory (C->S 0x20, `cmd::UseItem`) opens the portal pair on that tick, with no SC mode, whatever mode a live player is in; the charge (stat 70) or the scroll is spent only when the portal opened. d2d instead starts an SC cast (`World::read_portal` -> `Fight::cast_scroll`), drops the use while `pmode >= 0`, spends the charge at the click and opens at the action frame.

## 2. game.exe chain (all V: disasm.py read and the native run below)

FUN_0054d750 (0x20 passes unless FUN_0057eec0 says dead / state 0x36) -> FUN_0054b1e0 (owner, range) -> FUN_0055e170 (cursor empty, item mode 0, book: stat 0x46 > 0) -> FUN_005bf240 (books.txt pSpell 2, table 0x741790 entry 2 server = FUN_005be290) -> FUN_005be290 (not town, not level 0x88; FUN_00535430 closes the old pair; FUN_0056d130 opens) -> back in FUN_0055e170: stat 0x46 -1 (`0x55e2a2`) or FUN_0055e000 removes the scroll. The player's mode word (+0x10) is read only by FUN_0057eec0's dead test (V: the oracle varies it 0..0x13).

## 3. Differential (game.exe native vs d2d line copy)

`tools/emu/tome_use.py` runs FUN_0054d750 natively (and every function above except the hooks in its docstring: item lookup, stat getter, type tests, room / level / town, object creation result, packets, voice). d2d side: `d2d_use`, a copy of world.cpp:2159-2167, fight.cpp:4438-4446 and town.cpp:1596 at f8aab0e8 (the standard's allowed "before").

    cd <d2decomped>/tools/emu && source <venv>/bin/activate
    D2_MPQ_DIR=<mpq dir> python tome_use.py --exe game.exe 100000 7

| Class | Cases |
|---|---|
| same: nothing happens (incl. empty tome 5,743; town; cursor; dead) | 85,605 |
| game.exe opens now, d2d at the SC action frame | 3,109 |
| game.exe opens now, d2d nothing (pmode >= 0) | 7,930 |
| no free spot: game.exe keeps the charge and closes the old pair, d2d spends it | 3,116 |
| busy and no free spot: game.exe closes the old pair, d2d nothing | 7,962 |
| state 54: game.exe drops the packet, d2d has no test | 278 |
| total, seed 7 | 108,000 |

Mismatches before: 22,395. After: not run (no patch).

## 4. Missing for PROOF: complete

1. The d2d patch: `read_portal` without `cast_scroll`: the dead / state 54 / town / empty tests, `open_portal_at` returning whether it opened (closing the old pair first, as FUN_00535430 does), the charge or the scroll spent only on success; the cast cue kept. Every caller of `read_portal` (world.cpp:2598 only, V-d2d grep) and the joined-game path (town.cpp:1693, the host opens it) listed.
2. A d2d test mode printing the same lines (`test_game --tome-use SEED N`) and the after-run at 0 differ.
3. Upstream tests (`ninja test_combat test_game`, `ctest`), `tools/presubmit.py --quick`, `pr_guard.py` 0/0, `tools/d2ddet` hashes.
4. An independent adversarial verifier (not the author).
5. Limits to state: the client side of game.exe (what the 1.14d client sends on a right-click) not read; the skill-bar path (srvdofunc 113 -> FUN_005bf3d0) not compared with d2d's `portal_due` (d2d's do-func does not spend a charge, V-d2d read, U for its callers); FUN_00628c20 (usable) and FUN_0056d130's spot search hooked.
