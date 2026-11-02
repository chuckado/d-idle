# Lockstep probes P1 and P2: preparation state

> status: PARTIAL. Probes P1 and P2 of docs/research/original-game-lockstep.md (section 4) are prepared but NOT run: no live numbers exist yet. Card `lockstepp1p2`. Previous step: docs/research/lockstep-p0.md.

## Authorisation and ledger

- User go, 2026-11-02 ("go for lockstep probes"), covering the hook set named `d2-timehook proxy ijl11 IAT clock scaling` as INSTRUMENTATION under ADR 0014 (docs/cheats.md: code injected into the game is instrumentation only, the user's go per hook set, documented, logged).
- Reason for the ledger: confirm the 25 fps model and test a clock-scaling proxy on a throwaway character. Run ids planned: `lockstep-p2-k1`, `lockstep-p2-k2`. Not yet written: `d2cheat runstart` needs a running instance with the character; no ledger entry exists for the hook yet (U: whether runstart is the right marker for a hook that writes nothing to game memory; proposal: record the hook load as a ledger note with the PATCH/REHASH lines of the hook log).
- No memory write is planned. The hook writes only its own memory and import slots before the first game frame (hook protocol: `PATCH`, `READY`, exit `REHASH match=1` are the evidence).

## What stopped the live part (V, 2026-11-02, devhealth)

- Only one hidden instance was running and it is owned by a baseline lane in progress; it was not touched.
- The other hidden instances were stopped. Agents must not start, restart or log in a game (AGENTS.md "The live games"); only the main agent path (`tools/game.sh start <id>`, then the OOG login) may. So the agent stopped at P1 preparation, as the brief allowed.

## Done offline (V unless marked)

| Item | Result |
|---|---|
| Hook repository tests (`test.sh`, run on the worker host under Wine) | all checks PASS: forwarding, clock ratio per API for k = 1, 2, 4, waits, timers, null control |
| Proxy built from the game's own `ijl11.dll` (`build.sh proxy <path>`) | built, 16 KB, 6 exports, all by name (the brief expected ordinals 2, 3 and 5; the export table here is by name, so forwarding by name covers them, U whether the ordinals are preserved: the generator keeps ordinal numbers) |
| Game directories are hardlink clones (link count 6 on every file) | installing the proxy must rename the original (`ijl11_real.dll`) and add a NEW file, never write through the link, or all instances change |

## Run plan (for the run after an instance is up)

P1 (read only): read `dwGameFrame` twice 10.000 s apart by wall clock. The field offset in the game struct comes from D2MOO's `D2GameStrc` (a 1.10f source: U for 1.14d) and the pointer to the game struct comes from the argument of `GAME_UpdateProgress` (U: no static global known in our docs). Also find the caller of the progress update and the client frame cadence. Report frames per wall second and jitter; confirm or refute 25 fps. Tag U until the field moves 250 +- 2.5 per 10 s.

P2: install the proxy in the chosen instance's game folder (instance stopped), `D2HOOK_OFFLINE=1`, k in `d2hook.cfg`. Control: k = 1, 10 min town plus the first Blood Moor fight, compare with the unhooked run (frame rate, kills per minute, deaths, desync, crash, hook log `REHASH match=1`). Then k = 2 for 10 minutes only if the control is clean. Every claim is tagged with k and V run / V read / U.

## What the hook needs next (U, from reading its documented limits)

- Not hooked: `GetSystemTimeAsFileTime`, DirectDraw timing, waitable timers, delay-loaded imports. Whether Game.exe uses any of these for pacing is unknown until P1.
- k is fixed at process start; the P2 plan needs a restart per k.
- The bot's own timers need a scaled clock for k > 1 (original-game-lockstep.md section 2.6); not written.
