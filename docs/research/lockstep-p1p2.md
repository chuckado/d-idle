# Lockstep probes P1 and P2: preparation state

> status: PARTIAL (P1, k=1 control and k=2 done; REHASH not captured). Probes P1 and P2 of docs/research/original-game-lockstep.md (section 4) are prepared but NOT run: no live numbers exist yet. Card `lockstepp1p2`. Previous step: docs/research/lockstep-p0.md.

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


## P1 results (V read, hidden instance i4, 2026-11-02, read only, tool `cmd/lockstepscan`)

Method: snapshot all rw regions below 2 GB twice 2 s apart, keep 32-bit words that grew by 45 to 56, re-filter twice, then sample the survivors once a second for 10 s by the tool's own wall clock. No writes, no hook (first session) and again with the hook at k=1 (second session).

| Fact | Value |
|---|---|
| Candidates after the 2 s pass / after filters | 36 / 13 (first session), 34 / 8 (second) |
| Counters at exactly 25 fps | 11 words (static data of the exe image, plus 3 heap words), all 24.975 per second over 10.010 s = 250 frames; per-second minimum 24.97, maximum 24.98 (the 0.025 is the 10.01 s sampling span, not jitter) |
| Same static addresses after a restart | yes (two game starts, one unhooked, one with the proxy) |
| Other near-25 counters | one at 23.976 per second (U what it counts), one client-like counter with per-second jitter 23.97 to 27.97 (mean 25.175) |
| Jitter of the main counter | none measurable at 1 s resolution (U below that) |

Which word is `dwGameFrame` of the D2MOO game struct is U (the heap words are the candidates for fields of a game struct; the static ones may be copies or client-side counters). Not derived: the caller of the progress update and the client frame cadence (U, not done). The 25 fps model holds on our build (V read): 250 frames per 10.01 s, 7506 frames in 300.229 s = 25.0009 fps with the proxy at k=1.

## P2 k=1 control (V run, hook with k=1 from `d2hook.cfg`, run id lockstep-p2-k1)

- Hook log: `CLOCK k=1.0000 source=cfg`, `READY patched_slots=20` before the first frame; modules patched: the exe, the video and smack libraries and the JPEG stand-in's real DLL (APIs: tick counts, QPC, `timeGetTime`, `Sleep`, waits, `timeSetEvent`, `LoadLibrary*`, `GetProcAddress`). The exit `REHASH` line is pending the stop.
- Town idle 300.2 s: 7506 frames, 25.0009 fps, windows 24.98 to 25.18, 0 read errors (unhooked: 24.975 per 10 s, same counter).
- Blood Moor, bot `farm:Blood Moor`, 5 min (character level 12): 36 kills in 4 min and 0 deaths in the log minutes; no crash, no stall; frame counter after the run 25.03 fps over 20 s. No unhooked bot run on the same character was made today, so kills per minute has no same-day control (U); the null control claim rests on frame rate and on the run finishing like a normal run.

## What failed on the way (V)

- First start with the proxy: the hook refused (`D2HOOK_OFFLINE` is an environment variable and does not cross the remote start). The launcher process of Game.exe stayed alive after the refusal and left a hung process tree (proton launcher, steam.exe, wineserver, Xvfb) in the instance prefix; every later start, even unhooked, printed "Game.exe did not start" until `game.sh stop` cleared it. Rule: always stop before a retry. Fix in the hook (private repo): `offline=1` in `d2hook.cfg` also confirms offline use.
- Game logs of an extra instance are in the shared logs folder as `game-<id>.out`, not in the instance folder.

## P2 k=2 (V run, hook `k=2` from `d2hook.cfg`, same start path, run id lockstep-p2-k2)

| Measure | k=1 control | k=2 |
|---|---|---|
| Hook log | `CLOCK k=1.0000`, `READY patched_slots=20 hash_start=ca48e140e919615a` | `CLOCK k=2.0000`, `READY patched_slots=20`, same hash_start |
| Town idle 300.2 s, frames per real second | 25.0009 (windows 24.98 to 25.18) | 50.0001 (windows 49.958 to 50.160), 15012 frames |
| Ratio to k | 1.0000 | 1.0000 (error 0.002 percent; limit was 5 percent) |
| 10 s read, main counters | 24.975 | 50.055 (per-second min/max 49.95/50.95) |
| After the bot run, 20 s | 25.03 | 50.01 |
| Read errors | 0 | 0 |
| Blood Moor bot run, 5 min real (character level 12) | 36 kills by minute 4, 0 deaths | 76 kills by minute 3 (09:27 of a run started 09:24), 0 deaths in the minute log, no crash, no stall, ended in town alive |
| Game-time reading | 25 fps | the game ran 2.00 times faster per real second |

Kills per real minute at k=2 was about 25 against about 9 at k=1; the two runs differ in start state (the k=1 run began with vendor chores and a full pack), so this is not a clean xp-per-game-hour comparison (U); the original success test (xp and kills per game hour inside the k=1 noise band, `d2dsoak bands`) was NOT run. Frame ratio is V run; fidelity of fights at k=2 is only a first look (no death, no stall in 5 minutes). k=4 not tried (not in the brief).

Not captured: the hook's exit `REHASH`/`EXIT` lines. `game.sh stop` kills the process, so the exit handler never ran (V: no REHASH line in the k=1 log after a stop). The start `READY` hash is the only evidence of the slot set; an end-of-run re-hash needs a graceful exit or a periodic hook log line (U, next step).

## What the hook needs next

- A periodic `STAT` log line (virtual and real time, offset) so the evidence survives a kill.
- A scaled bot clock for k > 1 (the bot's own timers, input holds, are real time; section 2.6 of original-game-lockstep.md). The k=2 bot run worked without it, so the bot tolerates 2x at least in town and Blood Moor (V run, one 5 min sample).
- Same-day unhooked bot control on the same character and a k=4 step, then the `d2dsoak bands` comparison.
