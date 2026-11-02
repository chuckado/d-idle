# Verify: the waypoint refusal window (VERIFY-WAYPOINT)

> status: research, current. Independent re-check of docs/research/waypoint-trace.md. Verifier: Sonnet, card verifywaypoint. Method: the scan and context were done with my own scripts BEFORE reading the trace's conclusions; the game-side outcomes come from game.exe's own code run under the emulator (the existing oracle `tools/emu/wp_refuse.py`, which I reuse for the run but not for the rule); the rule comparisons and the second case set are mine. Tags: V = read in the disassembly or run (how stated), S = secondary source, I = inference, U = unverified. No game files or code are pasted. Offline: no game, no d2d slot.

## Verdicts

| claim | verdict |
|---|---|
| C1 pcdata+0x160 has exactly one writer, FUN_0055b720 | **HOLDS WITH LIMITS** |
| C2 game.exe refuses only within 10 s of the last hostile declaration | **HOLDS** (one correction, below) |
| C3 never in single player | **HOLDS WITH LIMITS** (I, not V) |
| C4 d2d refuses within 10 s of every arrival | **HOLDS** (S: read in upstream main) |
| C5 59,309 of 132,964 cases differ | **HOLDS** (reproduced exactly) |

## C1 writers and readers (V read)

`tools/emu/wpv_scan.py` sweeps game.exe's whole .text (capstone, linear) for every instruction with a `[reg + 0x160]` operand: 24 hits (`wpv_ctx.py` prints the context of each).

- Writers of that offset: 00421e77 (a constructor zeroing a 0x14-based object, not pcdata), 004b83d5 and 004b83fb (a table of 0x110-byte records, not pcdata), and **0055b76b inside FUN_0055b720**, the only one that stores a tick: it calls GetTickCount and stores the result into the block returned by FUN_006221a0 (the player-data getter). The stored value is the current tick, not the argument (the argument `now + 60000` goes to a list entry, not the field).
- Readers: FUN_0055b6c0 (00055b6d4; returns the field), reading structs of other types at 00421da9, 00423afa..00423d42, 005c8be4/006435b2/00646525 (a 0x23c-stride table, a base+stride pair), 005ce43c (a different struct).
- Callers of the setter: exactly one, 005a5f51 in FUN_005a5e50. Callers of the reader FUN_0055b6c0: the waypoint handler FUN_0054c5d0, FUN_005a5e50 (005a5f13) and **FUN_00584870 (005848bc)**.

Limits (U): a linear sweep misses writes through `rep movs`, memcpy or memset of the whole block, and indexed forms whose displacement is not 0x160 (the field could be written through a base already offset). I did not exclude those. Also the sweep keyed on displacement only; "which struct" is my reading of each context (I).

**New finding (not in the trace):** FUN_00584870 also reads the field, but its window is **5000 ms** (add 0x1388), not 10000, and refuses with the same voice event 0x13. It is a different handler (not the 0x49 waypoint path; what it serves is U). The "only reader is the waypoint refusal" idea would be wrong; the trace's claim is about the writer, which holds.

## C2 / C3 the window and single player

- FUN_005a5e50 is the hostile declaration: both parties must pass player/level gates (a stat-0xc test of at least 9 on each, FUN_006253b0, V read), and with the declare flag set it first compares the stored tick against the clock (voice 0x13, refused) and otherwise calls the setter. So the field is the tick of the last accepted declaration (V read).
- 0x49 handler: refuses while field + 10000 > GetTickCount, else travels. V run: 132,964 + 100,000 (below) cases through game.exe's dispatcher FUN_0054d750; the rule "refuse iff inactive, or now - last hostile < 10000" matches the oracle in **0 of 232,964 cases** (both case sets).
- Correction: the setter's tick argument is a 60 s value, so the refusal window after a declaration is 10 s, not 70 s (the stored value is the tick, V read at 0055b765..0055b76b).
- Single player (C3, I): a declaration needs two player units that pass the level-9 gates, so with one player the setter is unreachable and the field stays 0 (zeroed block: the refusal window ends at tick 10 s of uptime). The oracle's `--solo` mode only covers "no declaration". I did not trace the packet that reaches FUN_005a5e50 to prove no single-player caller exists (U).

## C4 d2d (S)

Upstream main (read-only checkout on the worker host): `components/game/world.cpp:2263` refuses when `now - actor->arrived_at < 10000`; `arrived_at = now` at world.cpp:1189 on arrival; world.hpp:204 comments it "pcdata+0x160". So d2d keys the same 10 s on every level change (S/read; I did not run d2d).

## C5 the counts (V run)

Reproduce (worker host, one command each, about 17 s):

    tools/emu/run-on-worker.sh --ttl 2400 wpv_cmp.py author 7 100000
    tools/emu/run-on-worker.sh --ttl 2400 wpv_cmp.py own 11 100000

| set | cases | game refuses | d2d rule refuses, game travels | game refuses, d2d rule travels | d2d rule differs | hostile rule differs |
|---|---|---|---|---|---|---|
| trace's generator, seed 7, 100000 random (grid + random; casehash d4a06ec0296eea11) | 132,964 | 57,297 | 47,157 | 12,152 | **59,309** | 0 |
| my generator, seed 11, 100000 (casehash 45644311c5653e57) | 100,000 | 52,246 | 29,056 | 14,592 | 43,648 | 0 |

The d2d rule in `wpv_cmp.py` is mine (inactive, or now minus the latest level change below 10000, unsigned 32-bit); the game outcome is game.exe's real FUN_0054d750 / FUN_0054c5d0 / FUN_0055b6c0 / FUN_0055b720 under unicorn. The 59,309 reproduces exactly (47,157 + 12,152). My second set agrees on direction and zero mismatches for the hostile rule; its 43,648 is a different mix (more late arrivals, more declarations), not a rate.

## Limits not covered

1. The oracle hooks the target check (FUN_00549570: in range and active), the unit lookups of the error path, the player-data getter (a zeroed block) and the timing log; the claim covers only the refusal window, not distance or activity.
2. I reused the author's oracle harness for the game run (same hooks); an independent harness was not written. A wrong hook there would be shared.
3. GetTickCount is restricted to 70,000..2^31 (outside the first 10 s of uptime and the wrap); behaviour across the wrap is U for both sides.
4. d2d was not run; its rule is read (S).
5. Single player (C3) is inferred; see above.

## Deviations (d2d vs game.exe)

- **D1** d2d refuses 10 s after every arrival; game.exe only after a hostile declaration. Tag V run (game) + S (d2d). A differential exists: `wpv_cmp.py` / `wp_refuse.py` with the d2d test binary (`test_game --waypoint-refuse`) on the same case file; before > 0 (59,309), after not run here.
- **D2** d2d has no hostile declaration, so it never refuses for it (12,152 cases in the set). Tag V run.

## Bot findings

- **B1** Our bot waits a 10 s cooldown plus 1 s after first seeing a level (internal/nav/waypoint.go, `waitCooldown`; the trace names the line); against real game.exe that wait is unnecessary in single player, but needed on d2d (D1). Not changed.

## Drill recipe (U to confirm live later)

On the lab binary: arrive by waypoint, immediately open the panel and select a row within 1 s, then at 5 s and at 11 s; expect refusal with d2d under 10 s. For the hostile rule: needs two level-9 players (U whether the lab can do it).

## Matrix cells

No matrix doc row changes proposed beyond docs/research/waypoint-trace.md's own table; add the FUN_00584870 5000 ms reader as an open question there.

## Shared-file lines for the merger

See `openspec/changes/verifywaypoint/merge-lines.txt`.
