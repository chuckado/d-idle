# Lockstep or speed-up for the ORIGINAL Game.exe 1.14d (agent LOCKSTEPRESEARCH, 2026-11-01)

> status: research, docs and web only. No game, no instance, no soak, nothing injected, nothing built. Target classic 1.14d (not D2R).
> Tags: **V** = fetched and read by me at the URL given (2026-11-01), **S** = search snippet or secondary, **U** = unverified / memory / inference.
> Machines are called "the worker host". D2R-only facts are flagged.

Question: can Game.exe (offline, under Wine, hidden Xvfb instance) get LOCKSTEP (logic advances only when the bot steps it, virtual
deterministic time, like the lockstep mode of our private d2d fork) or at least a safe SPEED-UP, so a 2 h gate or baseline run takes
minutes? Has anyone built it?

## 0. Answer in six lines

1. Nobody has published a time-control or frame-step tool for Diablo II Game.exe (no D2 TAS, no D2 "lockstep"; S, section 1). The generic tools exist but none targets Wine + a 32-bit Win32 game that is not SDL/OpenGL.
2. The cheapest honest route is an in-process proxy DLL (the `ijl11.dll` stand-in that the d2d maintainer's `whisper` tool already loads into Game.exe, V) that patches the time imports of Game.exe. `LD_PRELOAD` on the Wine process is the wrong layer (section 2.4).
3. The frame pacing is simple and well located: one function, `GAME_UpdateGamesProgress` (1.14d `0x0052FC20`, V in D2MOO), gates server frames on `timeGetTime() - last >= 1000/25`. Scale or replace that clock and the logic rate follows (section 2.1).
4. Tier A (speed-up by a factor k, 2 to 8): about 6 to 12 agent-hours, medium risk. Tier B (true stepping with a control socket): about 20 to 40 agent-hours, high risk (client thread barrier, DirectDraw/Glide present, loopback packet queue). Tier C (emulate the logic headless): this is d2d, which we already have.
5. Cheapest probe: a 30 line Win32 test program under Wine on the worker host, no game, that shows which clock APIs a preload can scale (section 4, probe P0). Then a proxy-DLL null hook, then k = 2 on a hidden instance.
6. Policy: a time hook is instrumentation only if it provably writes no game state and a null-hook control (k = 1) matches an unhooked run. Section 5 lists the evidence. Recommendation: P0, then Tier A with the control; Tier B only if A shows fidelity drift or the bot's reads need a paused game.

## 1. Prior art

| Item | What it does (source) | Licence, maintained | 1.14d? Wine? | Frame step / time control for Game.exe? |
|---|---|---|---|---|
| **libTAS** (clementgallet) | GNU/Linux tool giving TAS features (frame advance, fast forward, savestates) to native games; game starts paused, `V` advances a frame, `tab` fast-forwards (V https://github.com/clementgallet/libTAS). Time hooking and frame sync internals are NOT on the README (V, not stated); from memory it hooks libc time and sync calls plus SDL/GL swap as the frame boundary (U). | GPL-3.0 (V); release v1.4.8 published 2026-05-14 (V, GitHub API); 3,017 commits (V) | Not D2-specific. Wine: README says "Wine support is limited and not being actively developed" (V); TASVideos Wine page: not worked on since 2019/2020, Wine 4 or 5 likely best, Wine 9.0 shown to work, only DX-Ball 2 and I Wanna Be the Boshy listed as examples, Wine TASes are not accepted for submission (V https://tasvideos.org/EmulatorResources/LibTAS/Wine; S search snippet for the submission rule). Known Wine issues: no disk-write prevention, window focus, repaint (V README). | Possible in principle, in practice no: Game.exe is a 32-bit non-SDL Win32 game on DirectDraw/D3D/Glide behind Wine; libTAS needs its frame boundary (present call) and the wineserver and Wine's own thread timers escape LD_PRELOAD (U). Verdict: not a fit; maybe a reading source for ideas. |
| **libfaketime** | LD_PRELOAD that fakes `clock_gettime`, `gettimeofday` etc. (U, memory). | LGPL-2.0 (U) | Wine bug threads: a user "can't get libfaketime to work with Wine", limits with child processes (S https://list.winehq.org/mailman3/hyperkitty/list/the Wine bug tracker/thread/M37ZLJN5T645EPZUN7KRIRBAA4GDKI5H). No report found that FAKETIME scales a Wine game's `timeGetTime`/`GetTickCount` (S). | Wall-time faking only. Not a frame step. Verdict: no. |
| **clock_gettime_override** (hi-angel) | Rust LD_PRELOAD that overrides `clock_gettime` so the game sees time at 1/4 speed; author tested it with Wine games (Alien Shooter 2, Dungeon Nightmares, GTA IV); "only works if the game calls `clock_gettime`", check with strace; need both 32-bit and 64-bit builds (V https://github.com/hi-angel/clock_gettime_override). | no licence stated (V) | Wine: yes for those games (V, author's claim). | Evidence that a preload CAN scale Wine games whose Wine path reaches `clock_gettime`; says nothing about Game.exe's `timeGetTime`. Useful as the P0 comparison. Verdict: P0 candidate, speed-up at best (slow-down in their case). |
| **Speed hackers (Cheat Engine speedhack, `speedhack-rs`)** | Windows speed hack: hook time APIs inside the process. speedhack-rs ships as a proxy `version.dll` (or other name) next to the exe; README does not name the hooked functions, says some games crash if hooked instantly; no licence on the page, no Wine mention (V https://github.com/Hirtol/speedhack-rs). Cheat Engine's speedhack hooks `GetTickCount`, `timeGetTime`, `QueryPerformanceCounter` (U, memory). | unknown | Windows. Under Wine the DLL proxy idea works for any DLL Game.exe loads by name (U, see whisper). | Speed-up yes, the same design as our Tier A. No stepping. Verdict: the design pattern to copy; do not run third-party binaries. |
| **d2decomped / d2d** (psi29a) | GPL-3.0-or-later reimplementation of D2 LoD using 1.14d as reference, 1,552 commits, last commit 2026-11-01 (V https://github.com/psi29a/d2decomped, GitHub API). Public tree has a control channel (`--headless --devctl SOCK`) with `info` (frame counter), `cmd`, `state`, `monsters` and so on, but NO step/pause verb and no virtual clock in `docs/control_channel.md` (V https://raw.githubusercontent.com/psi29a/d2decomped/main/docs/control_channel.md; the tree listing has no path containing "lockstep", V). The lockstep mode we use is OUR private fork (`kTickMs = 40`, `step` adds one tick to a virtual clock, S read in docs/d2d-tick-rate.md). | GPL | Not Game.exe. Builds on Linux, no Wine. | It is the lockstep we already have but of a re-implementation. For Game.exe: reusable IDEAS only (section 3.4). |
| **whisper** (in d2decomped `apps/whisper`) | A remote control INSIDE the original Game.exe 1.14d, loaded as a stand-in `ijl11.dll` (Game.exe imports IJL11.DLL ordinals 2, 3, 5), forwards to `ijl11_real.dll`, starts a socket thread, runs commands on the game's own thread via `WM_APP`, `peek ADDR N`, hardcoded 1.14d addresses (V as read by the earlier agent in docs/research/d2d-whisper-lan-oracle.md section 1, source raw.githubusercontent.com/psi29a/d2decomped/main/apps/whisper). The maintainer ran Game.exe under Wine and joined it with d2d (commit eafcb5f2, 2026-10-27, S gh api in the same doc); whether whisper loaded in that Wine run is U. | GPL | 1.14d yes (V by doc). Wine: DLL search order should load a native proxy DLL placed beside the exe (U). | The loading mechanism we need. It has no time hook. Verdict: best starting point for the injection method; code is GPL, so we would write our own (docs/decisions.md: patches never go into this repo). |
| **D2MOO** (ThePhrozenKeep) | Decompiled D2 source with per-version addresses; `source/D2Game/src/GAME/Game.cpp` carries the 1.14d frame scheduler (V, section 2.1). | open source (U licence) | Source for reading, not a tool. | Gives exact function names and addresses; no stepping. |
| **D2BS / kolbot / D2Bot** (blizzhackers) | In-process JS bot frameworks, Game.exe injection by loader; forum threads show 1.13c and 1.14b/d efforts, uneven support, last D2BS thread activity June 2025 (S https://blizzhackers.discourse.group/ search snippets). | open source (U) | 1.14d: partial (S). Wine: U. | They automate play in real time; no time control found (S). |
| **PlugY, BH, d2hackmap, D2SE, D2 Loader, Median XL** | Single-player extension DLL (PlugY), HUD/maphack DLLs, launchers, a data mod (Median XL). None is known to touch frame pacing (U, memory; not fetched). A Median XL forum post says 25 is the engine's FPS maximum and that attack-speed "hacks" exploit fps caps (S https://forum.median-xl.com/viewtopic.php?p=66434). | various (U) | PlugY supports 1.14d (U). | Not time control. They are proof that the proxy-DLL / loader pattern is standard for D2 (S). |
| **Tool-assisted D2 / "D2 speed hack" search** | One search for D2 1.14d speedhack / TAS found no D2-specific tool, only generic speedhack utilities and the patch notes (S, section 0 line 1). | - | - | None found. Absence in search results is not proof. |
| **Emulator / replay approaches** | Our own harness `tools/emu` runs Game.exe functions natively under an emulator with Win32 imports stubbed (for example `emu.Emu.w_GetTickCount` replaced by the case's clock, tools/emu/wp_refuse.py, V read here). That is function-level replay, not a running game. TASVideos suggests PCem for old Windows games (V Wine page). | ours | works for logic functions | A frame-level step of the whole process in an emulator is not realistic (needs a full Win32 + graphics stack); function-level is what we already use. |
| **Speedrun community** | D2 speedruns are real-time (U, memory, not searched). | - | - | no tooling known. |

Dead ends already known: D2MOO cannot be built to a Linux headless server cheaply (docs/research/d2moo-fit.md, V repo read). D2R is a different executable; its tooling does not apply.

## 2. Technical feasibility against the real structure

### 2.1 Where the 25 fps logic lives (V from D2MOO source, 1.14d annotations; our docs for the rest)

Source: https://raw.githubusercontent.com/ThePhrozenKeep/D2MOO/master/source/D2Game/src/GAME/Game.cpp (fetched, 3,430 lines). Addresses are D2MOO's own annotations, not checked by me against our Game.exe, so they are V as a source claim and U as our binary until a probe reads them.

| Item | Detail |
|---|---|
| `GAME_UpdateGamesProgress(bLimitCatchUp)` | 1.14d `0x0052FC20` (1.10f ordinal 10004, 1.13c ordinal 10008 export of D2Game.dll). Calls `QueryPerformanceCounter` at start and end (returns the QPC delta, a load measure), `timeGetTime() & 0x7FFFFFFF`, keeps `dword_6FD45844` = last frame time. If `elapsed < gnTargetMsPerFrame` it returns 0 (no frame). Else it carries the surplus (`nRemainingDelayMs`), clamped to one frame when `bLimitCatchUp`, and runs `GAME_UpdateProgress` for every live game GUID. |
| `gnTargetMsPerFrame` | `1000 / gnTargetFrameRate` (`GAME_InitializeFrameRate`, line 2178) and set from a frame-rate argument at line 3240. Default rate 25, so 40 ms (U for the exact default: `DEFAULT_FRAMES_PER_SECOND` is used in the same file; value 25 is consistent with the 25 fps in our docs/d2-1.14d-memory.md line 322 and the fork `kTickMs = 40`). |
| `GAME_UpdateProgress(pGame)` | 1.14d `0x0052D870`. `++pGame->dwGameFrame`, environment update, then per-frame events keyed on `dwGameFrame % 20 / 12 / 11 / 1500 / 8192`. It reads `GetTickCount()` only for frame-rate statistics and a few timers. So **game logic time is the frame counter**, not wall time. |
| CPU ratio field | `nCreationTimeMs_Or_CPUTargetRatioFP10`: with global `dword_6FD2CA10 = 1` (initial value, V) it is forced to 1024, so no load-dependent logic (V source comment at line 2817: "always 1024"). |
| Wall-clock reads inside LOGIC | Yes, a few: waypoint refusal 10 s window `GetTickCount() < pcdata+0x160 + 10000` (V run + read, docs/research/waypoint-trace.md; never fires in single player), teleport landing stamp at `+0xa4` (docs/research/sk-teleport-trace.md, S read). Any time hook must cover `GetTickCount` consistently with `timeGetTime`, or these two disagree. |
| Linked time APIs | D2Game links `winmm` for `timeGetTime` (V docs/research/d2moo-fit.md). In 1.14d D2Game is merged into Game.exe, so one module's import table (IAT) covers server, client and Storm/Fog code. Whether some code reads time through Storm or Fog wrappers is U. |

Not yet found (U, must be read by probe P1): who calls `GAME_UpdateGamesProgress` in 1.14d single player (a server thread in the same process, or the client main loop between frames), its `Sleep` or wait, and the client frame cap (`FPS` pointer at `0x3BB390`, V D2Ptrs.h:183 in docs/d2-1.14d-memory.md line 68, is a display counter, not the limiter, U). My memory of 1.13 is that the single-player server is a separate thread inside the process (U).

### 2.2 Client side clocks (U, none of this was read in Game.exe)

- The client has its own render/interpolation timing (animations on `GetTickCount`/`timeGetTime`/QPC, U) plus the present call. In single player the client presents under Wine on Xvfb through DirectDraw or D3D (Wine's wined3d/ddraw on software GL, S from our container notes in docs/oog.md and fleet docs that Game.exe runs hidden under Xvfb, V from those docs; the wrapper, if any, is U).
- Client and server exchange packets through an in-process loopback queue (the TCP/IP-game path is the same packet layer, V docs/research/d2d-whisper-lan-oracle.md section 2). The client applies a server frame's packets on its next client iteration. So "server frame N done" does not mean "client memory shows frame N": the bot's memory reads (client unit tables) lag. That is the key difficulty for stepping.

### 2.3 Is Game.exe deterministic given inputs and seed?

- Per-unit and per-missile RNG are seeded, and logic is frame-based (V: Blizzard shard draw reseeds `{cx + age, 666}`, docs/research/sk-ground-trace.md; FUN_0045c3e0 missile seed). The game seed is picked at game creation (the host sends the map seed, V net-join notes in the whisper doc). How that seed is chosen in single player (clock-derived?) is U. If it comes from `GetTickCount`/QPC, a virtual clock would make it repeatable; probe P3 reads `Act.dwMapSeed` (+0x0C, V D2Structs.h:359-363 in docs/d2-1.14d-memory.md line 236) after two identical starts.
- Sources of non-determinism that a lockstep must remove: input arrival time relative to frames (clicks reach the client at wall time and get a packet to the server on the next client iteration), client/server thread interleaving, the wall-time reads above, save file contents. Our own d2d determinism work shows the same classes even with a perfect virtual clock: wall-time thread pool installs and a smoothed camera (docs/d2d-determinism.md sections 3 and 4, V own docs). Expect Game.exe to need the same "input only at frame boundaries" discipline.

### 2.4 Which layer to hook (Wine)

| Layer | Reaches | Problem |
|---|---|---|
| `LD_PRELOAD` on the wine process | whatever Wine's unix-side calls via libc: QPC (Wine's ntdll.so uses `clock_gettime`, S/U) and `nanosleep`-based waits | Wine's `timeGetTime`/`GetTickCount` are built in ntdll/winmm (the Wine tree's `dlls/winmm/time.c` calls `timeGetTime` from its timer thread, V https://raw.githubusercontent.com/wine-mirror/wine/master/dlls/winmm/time.c; how `timeGetTime` itself derives its value was not shown, U). Newer Wine reads tick count from a shared page (U). Different APIs could scale differently, and the wine bug search found that `timeGetTime`'s winmm entry gets special treatment (S). Also needs both 32 and 64 bit preloads and wineserver timing is untouched. |
| Proxy DLL beside Game.exe (whisper model) | patch Game.exe's IAT slots (and the winmm/kernel32 stubs) for `timeGetTime`, `GetTickCount`, `GetTickCount64`, `QueryPerformanceCounter/Frequency`, `Sleep`, `SleepEx`, `WaitForSingleObject(timeout)`, `GetSystemTimeAsFileTime`, `timeSetEvent` | Works on any Wine version, one consistent clock. Needs a mingw-w64 32-bit build (the whisper build does this, V), and IAT patching leaves data/code of the game untouched. |
| Inline detour of `GAME_UpdateGamesProgress` (0x52FC20) | the server gate only | Tier B. Writes code bytes of the game (a patch), so stricter policy treatment (section 5). |
| Patched Wine | everything | Costly to maintain; a Wine bug thread suggests the idea for faketime (S). Not recommended. |

### 2.5 What a hook must intercept

(a) Speed-up only (virtual = base + (real - base) * k, monotonic, same for all APIs):
`timeGetTime`, `GetTickCount`, `GetTickCount64`, `QueryPerformanceCounter` (and `QueryPerformanceFrequency` unchanged), `GetSystemTimeAsFileTime` if used, `Sleep`/`SleepEx`/waits with timeouts divided by k (otherwise a 40 ms Sleep stays 40 ms real and the loop is throttled), `timeSetEvent` periods divided by k. Not intercepted: the present call; if Xvfb software rendering cannot draw k * 25 frames a second the client lags behind the server (server frames come in bursts, `bLimitCatchUp` clamps to one frame per call, which then CAPS the achievable k at about (calls per second) / 25, U).

(b) True stepping with a socket (compare d2d: `step N` adds N * 40 ms to a virtual clock and runs N fixed ticks, answer after the world has done them, S docs/d2d-tick-rate.md):
- all of (a), but the virtual clock only advances on a `step N` verb (the hook returns the same time until stepped), so `GAME_UpdateGamesProgress` yields 0 frames when idle (a natural pause, no code patch needed when the clock is frozen);
- a barrier: after N server frames wait until the client consumed the loopback packets (a client-side frame counter or an empty queue; the hook cannot know this without a game read, so it must read one counter, U) and until the client presented the frame;
- every blocking wait inside Game.exe that uses real time must be released (otherwise a frozen clock deadlocks `Sleep` and timeouts: Sleep(n) must return after the virtual clock moves, or be shortened to a yield);
- Windows messages and mouse input: queue the bot's clicks and release them at a step boundary, as d2d's `Pacer` does (docs/d2d-determinism.md section 3.1, V);
- a control socket with `step N`, `now` (virtual ms, frame counter), `rate k`, `hash` (optional), the same shape as d2dctl (`internal/d2dctl/verbs.go`, S), run on a hook thread that is itself exempt from the virtual clock.

### 2.6 Bot side

Our memory backend reads asynchronously from another process (`/proc/PID/mem` style, docs/d2-1.14d-memory.md). For stepping the game must be PAUSED between steps so reads are consistent: that is the barrier above plus the frozen clock. For speed-up there is no pause; the bot sees a game running k times faster, its own timers (Clock, input holds of at least 40 ms, 50 to 100 ms between actions, docs/d2-1.14d-memory.md line 322) must be divided by k on the bot's clock (`env.Clock` is already virtual under d2d, `env.RealClock` is wall time on Game.exe, V docs/d2d-determinism.md section 3.1), so a speed-up needs a scaled `RealClock` and a rate read from the hook (a socket or a shared file).

## 3. Reusable ideas from d2d's lockstep (for a Game.exe hook)

3.1 (V own docs, S fork) Fixed step `kTickMs = 40`, the virtual clock is the only time source, `step` returns when the world ticked.
3.2 Everything else is a frame-boundary discipline: watchers run on the stepping goroutine after the chunk (`Pacer.Watch`), neighbour-level builds wait under lockstep (docs/d2d-determinism.md 4.1). A Game.exe hook needs the equivalent of "no wall-time asynchrony inside a step".
3.3 Wall speed: d2d idle town 250 ticks in 11 ms (V-d2d). Game.exe will not get close: it renders (even hidden) and the client must consume each frame; expect a few times real time, not hundreds (U).
3.4 The d2d control channel is a text socket with verbs (V, public docs); the same text protocol fits the hook thread.

## 4. Tiers, cost, and the probe runbook

| Tier | What | Agent-hours | Risk | Gain |
|---|---|---|---|---|
| P0 | Win32 test program under Wine on the worker host: print `timeGetTime`, `GetTickCount`, QPC, `Sleep(100)` elapsed per clock, run unhooked and with a `clock_gettime` preload; no game | 1 to 2 | none | tells whether preload is viable at all (probably not for `timeGetTime`, U) |
| A | proxy DLL `ijl11.dll` stand-in with IAT patch, scale k, rate file; scaled bot clock | 6 to 12 | medium: client/server pace mismatch, Xvfb render cost, third-party-like DLL under Wine, fidelity drift vs real time | gate runs at about 2x to 6x (U); a 2 h run in 20 to 60 min |
| B | stepping: frozen virtual clock, barrier, control socket, bot step API | 20 to 40 (U) | high: unknown caller of 0x52FC20, client threads, deadlocks, DirectDraw present | determinism and consistent reads; speed limited by render |
| C | headless logic emulation | 0 (exists) | medium | this is d2d; it remains the fidelity-checked fast path |

### Runbook: cheapest probes (run later on a hidden instance only, never on the user's desktop, never while a gate run owns the worker host)

**P0 (no game).** Build with the worker host's mingw-w64 a 32-bit exe `timeprobe.exe` that loops 10 times: read the four clocks, `Sleep(100)`, print deltas. Run `wine timeprobe.exe` (baseline), then with `LD_PRELOAD` of `clock_gettime_override`-style slow-down built locally (do not run third-party binaries; build from source after reading it). Success: ratio of delta(clock) to delta(real) per API; any API whose ratio is not 1/k breaks a preload approach. Output: a table under docs/. Stop there if `timeGetTime` does not move.

**P1 (read only, one hidden instance, throwaway char).** With the existing memory reader, read `dwGameFrame` (game struct field of `GAME_UpdateProgress`, offset to be derived from D2MOO's `D2GameStrc`, U) twice 10.000 s apart (wall, by the bot's wall clock): expect 250 frames. Also read the client's frame counter. No writes, no hook. Success: frames per wall second = 25.0 within 1 percent (confirms the clock model of 2.1) and the offset exists in our build.

**P2 (null hook, then k = 2).** Proxy DLL with a stand-in `ijl11.dll` forwarding the real exports and an IAT patch for the clocks, rate k from a file read at start (k = 1 first). Run 10 min in town and the first Blood Moor fight at k = 1 (identity), then at k = 2. Success: (virtual game seconds from `dwGameFrame` / 25) divided by (real seconds) = k within 5 percent; the bot's reads stay consistent (no unit position that jumps backwards, no read error rate above the k = 1 level); xp per game-hour and kills per game-hour at k = 2 within the k = 1 noise band (use the existing baselines, docs/soaks and `d2dsoak bands`). Then raise to 4, stop at the first k whose ratio drops below 0.9 k or whose death/stall rate changes.

**P3 (determinism, k = 1 virtual clock fixed start).** Two cold starts of one throwaway character with the virtual clock starting at the same value, same scripted inputs released at frame boundaries (needs Tier B for strict inputs; with Tier A only the SEED check is possible): compare `Act.dwMapSeed`, the first 5000 frames of a command/hash journal (same machinery as `tools/d2ddet`, docs/d2d-determinism.md). Success: equal seed and equal journal to the first divergence, which is then bisected. Failure with equal seeds means a thread race (expected without Tier B).

Measures (all runs): ratio virtual/real time, read error count, journal equality, hook log (section 5), process list empty afterwards (AGENTS.md cleanup rule).

## 5. Policy: instrumentation, not a crutch

(The user is rewriting the memory rules. I did not see the new text; the points below follow the stated rule: reading and editing memory is allowed any time but editing must not be a crutch, and time-control hooks count as INSTRUMENTATION that must not change how game state evolves. Existing rules today: memory is read-only except logged cheats, AGENTS.md; docs/cheats.md.)

A time hook is on the instrumentation side if all of this holds, and a report may claim it only with the evidence:

1. **It changes time, not state.** It writes only to its own memory and to the import slots (or detour bytes) it installs, and it does that BEFORE the first game frame. Evidence: a hook log line per installed patch (module, import name, old and new address, timestamp), count of patches, and "game writes after start = 0": after the first frame the hook performs no write into Game.exe's data, heap or code (self-check: the hook never calls `WriteProcessMemory`, never touches the game structs; the IAT/detour patch set is hashed at start and re-hashed at the end, equal).
2. **Monotonic, consistent clock.** One virtual clock for all APIs; never goes backwards; the same function of real time for every API; logged factor k and the virtual-minus-real offset at start and end (and each rate change).
3. **Null-hook control.** k = 1 with the hook loaded must give the same results as no hook (frame rate 25.0, xp/h, kills/h inside noise, same journal when seeds are equal). Without this control a k > 1 result is not comparable with the unhooked baselines.
4. **Tag every result.** Soak and baseline reports that ran under k state `time hook k=N, null control run <id>`, and the fidelity gate does not mix them with real-time baselines until the k = 2 and k = 4 runs sit inside the baseline band. Speed-up changes the bot's relation to the game (bursts, catch-up clamp), so a speed-up run is a measurement under a modified environment, not proof that a real-time run would pass.
5. **Stepping adds input ordering evidence.** Under Tier B the input (clicks, keys) is delivered at frame boundaries; log the frame number of each input. That is observation of a changed schedule, not a change of state, but it must be in the journal so a reader can see what the game was fed.
6. **Detour of `GAME_UpdateGamesProgress` is a code patch.** Prefer a frozen clock (no code patch) for pausing; if a detour is unavoidable, list its address and bytes in the ledger as a patch, with the reason, like a cheat ledger entry (docs/cheats.md), and keep it out of runs that claim the unmodified Game.exe baseline.
7. Anti-cheat: classic offline single player with no Battle.net (AGENTS.md); Warden-style checks do not apply (U for 1.14d offline). D2R is out of scope.

## 6. Recommendation

1. Do P0 (an hour or two, no game): it decides between a cheap preload and the proxy DLL and costs nothing. Expect the proxy DLL.
2. Do P1 to confirm the 25 fps model on our own build (read-only).
3. Build Tier A (P2) with a null-hook control; target k = 2 to 4. It turns a 2 h baseline into 30 to 60 min and keeps the gate's meaning if the k = 1, 2, 4 results agree. Keep every claim tagged with k.
4. Do NOT start Tier B unless P2 shows a fidelity drift or the bot needs a paused game for reads. Tier B is the answer to "lockstep" but is a research project (unknown thread structure, client barrier). Meanwhile d2d lockstep stays the determinism and speed path; Game.exe stays the fidelity oracle at real time or k-scaled time.
5. First steps after approval: a change card for P0 and P1 (docs and a tiny tool, `owns` limited), no game needed for P0.

Open questions: caller and thread of `0x52FC20` in single player; how the single-player game seed is chosen; whether the client reads time through the same IAT slots; whether Wine's `timeGetTime` moves under any preload; the real ceiling of k under Xvfb software rendering; the new memory-rule text.
