# Lockstep probe P0: time hook under Wine, no game

> status: result of probe P0 of docs/research/original-game-lockstep.md (section 4). Card `lockstepp0`. Verdict: PASS.

All measurements are V (run on the worker host, Wine 11.20 Staging, console programs only, no display, scratch Wine prefix, no game, no game prefix, no d2d-lab slot). Raw output: `tools/lockstep/results-p0.log` (the per-sample `S` lines are dropped there; the ratios below were computed from them). One run per row (single sample, not repeated): treat the last digit as noise.

## Toolchain

| Item | Result (V) |
|---|---|
| mingw-w64 (`i686-/x86_64-w64-mingw32-gcc`) | not installed |
| zig | not installed |
| clang 23.1.4 + lld | present. `clang --target=i686-windows-gnu -fuse-ld=lld -nostdlib` plus Wine's own headers (`/usr/include/wine/msvcrt`, `/usr/include/wine/windows`) and Wine's 32-bit import libraries (`/usr/lib/wine/i386-windows/lib*.a`) builds working 32-bit PE files |
| `winegcc` | present but needs mingw-style host libs; fails (`cannot find -lkernel32`), not used |
| compiler-rt | absent: no 64-bit division in the probes (they print raw integers, `run.sh` and awk do the arithmetic) |
| libfaketime | not present (no library, no package) so question 5 was NOT run |

No package was installed. Nothing system-wide changed.

## Recipe

```
labrun <worker-host> -- tools/lockstep/build.sh          # -> tools/lockstep/build/{probe.exe,fakeijl.dll,gamedll.dll}
labrun <worker-host> -- tools/lockstep/run.sh all        # about 3 min, one core or less
labrun <worker-host> -- tools/lockstep/run.sh one gate 400 0   # one case: MODE KPCT [SLEEP [MODS]]
```

(Start them through `tools/agentrun.sh <agent> --ttl N --`. The prefix defaults to a scratch directory, override with `LOCKSTEP_PREFIX`.)

- `fakeijl.dll` (hook.c): stand-in proxy DLL. `probe.exe` imports it like Game.exe imports IJL11.DLL, so the loader runs its `DllMain` before the exe entry point. `DllMain` reads `LOCKSTEP_KPCT` (k in percent; unset = no patch), saves the real function addresses, then overwrites the import-address-table slots of the listed modules (default: the exe and `gamedll.dll`) for `timeGetTime`, `GetTickCount`, `GetTickCount64`, `QueryPerformanceCounter`, `Sleep`. Virtual = base + (real - base) x k in 16.16 fixed point, base = value at hook start (so virtual time continues from the real value). `QueryPerformanceFrequency` is left alone (the counter is scaled, not the frequency). It logs one `HOOK module=... patched_slots=N` line per module.
- `gamedll.dll`: a second module with its own clock imports (stands for D2Game.dll). `gate_run` is the frame gate of the brief: `if (timeGetTime() - last >= 40) { frame++; last += 40; } Sleep(1);`.
- Reference clock inside the programs: `GetSystemTimePreciseAsFileTime` (Wine maps it to the host realtime clock), never patched.

## Q1: the clocks under Wine, unhooked (20 s, sample every 50 ms)

| API | Ratio to host clock | Resolution | Notes |
|---|---|---|---|
| `timeGetTime` | 1.0000 (both halves 1.0000 / 1.0001) | 1 ms steps, 1000 changes per second | no `timeBeginPeriod` needed in this Wine |
| `GetTickCount` | 0.9997 | about 15.6 ms steps (64 changes per second) | lags `timeGetTime` by up to 15 ms; the 0.9997 is that quantisation at the end of the run |
| `GetTickCount64` | identical to `GetTickCount` in every sample | same | |
| `QueryPerformanceCounter` | 1.0000 | frequency 10 MHz, step 0.4 us | |
| Monotonic | no backward step in any API | | |
| `Sleep(100)` | mean 100.11 ms, range 100.04 to 100.19 ms | | |
| `Sleep(1)` | mean 1.06 ms, range 0.08 to 1.34 ms | | gate loop runs about 9430 iterations per 10 s |
| `timeBeginPeriod(1)` | no measurable effect (`Sleep(1)` mean 1.05 ms before and after) | | |

## Q2/Q3: IAT hook, ratio virtual / real over 10 s (sleep not scaled)

| k | timeGetTime | GetTickCount | GetTickCount64 | QPC | halves (tgt) | backward steps | max abs(tgt - gtc) | max abs(tgt - qpc) |
|---|---|---|---|---|---|---|---|---|
| 1 (null) | 1.0000 | 1.0000 | 1.0000 | 1.0000 | 1.0001 / 0.9999 | 0 | 15 ms | 0.6 ms |
| 2 | 2.0001 | 1.9993 | 1.9993 | 2.0000 | 2.0000 / 2.0002 | 0 | 32 ms | 1.4 ms |
| 4 | 3.9999 | 3.9999 | 3.9999 | 4.0000 | 4.0000 / 3.9999 | 0 | 60 ms | 3.2 ms |

Error against k is below 0.04 percent for every API at every k (worst: `GetTickCount` k = 2, -0.035 percent, tick quantisation). All patched clocks stay monotonic and agree within the real-time quantisation of `GetTickCount` (15.6 ms real, so up to 62 ms virtual at k = 4). A game that compares `GetTickCount` with `timeGetTime` at sub-100 ms virtual precision sees that skew; the waypoint and teleport stamps from the parent doc use 10 s windows, so they are not affected (U for any other use).

Frame gate (`gamedll.dll`, 10 s of host time, `Sleep(1)` loop, not scaled):

| Case | Frames | Frames per host second | Virtual ms per 10.0 host s | Process CPU (kernel + user) |
|---|---|---|---|---|
| no hook | 249 | 24.9 | 10000 | 0.17 s = 1.7 % of a core |
| k = 1 (null) | 250 | 25.0 | 10001 | 0.19 s = 1.9 % |
| k = 2 | 499 | 49.9 | 20000 | 0.17 s = 1.7 % |
| k = 4 | 1000 | 100.0 | 40004 | 0.18 s = 1.8 % |
| k = 4, only the exe patched, `gamedll.dll` not | 250 | 25.0 | 10001 | 0.18 s |
| k = 4, `Sleep(ms)` scaled to `Sleep(ms / k)` | 1000 | 100.0 | 40000 | 10.05 s = 100.5 % of a core (4.96 million loop iterations: `Sleep(0)` spins) |

Q4: the frame-gate loop costs about 1.8 percent of a core at k = 1, 2 and 4 alike (includes process start; the wineserver process is not counted, it was idle). That is the cost of the gate alone: Xvfb software rendering and the game's own work are not in this number.

## Q5: preload scaling

NOT run: libfaketime is not installed on the worker host and the brief forbids installing it. Not measured, so the answer is "unknown". Reasoning only (U): a `clock_gettime` preload can scale only what Wine routes through the host libc, and the parent doc's findings (a Wine `timeGetTime` path, wineserver timers) were not checked here. The in-process IAT hook does not depend on the answer.

## Verdict

PASS: a proxy-DLL import-table time hook can scale the frame gate of a Win32 program under Wine. Virtual time = k x real within 0.04 percent for k = 1, 2, 4 for all four clocks; the 25 fps gate gave 25.0, 49.9 and 100.0 frames per host second; the k = 1 null hook matched the unhooked run (250 vs 249 frames, one-frame start quantisation); clocks stayed monotonic; CPU cost is negligible and independent of k.

Findings that Tier A must respect (all V from this run unless marked):
1. Every module that reads the clocks must be patched. With only the exe patched the DLL's gate ran at 25 fps. For the game that means Game.exe, D2Game.dll, D2Client.dll and every other module that imports these functions, plus modules loaded later (hook `LoadLibrary`, or patch on a module list at start; not tested here, U) and callers that resolve the functions with `GetProcAddress` (not covered by an IAT patch, not tested, U).
2. Do not scale `Sleep` toward zero. `Sleep(1)` scaled by 1/4 became `Sleep(0)` and burned a whole core with no gain. Leave `Sleep` raw (or floor at 1 ms). The gate still produced the right frame rate.
3. Timeouts that never go through these five imports (`WaitForSingleObject` with a timeout, `SetTimer`, `timeSetEvent`, DirectDraw or window-message timing, Wine-internal timers) were not tested and are not scaled (U). The client-server pace of the real game depends on them.
4. The probe `ijl_stub` exports one dummy function. The real stand-in must forward `IJL11.DLL` ordinals 2, 3 and 5 to the real DLL (parent doc, whisper). Not tested here.
5. The gate loop here produces at most one frame per `Sleep(1)` (about 940 per host second, so about k = 37). The real game carries surplus time and clamps catch-up, so its ceiling is a P1/P2 question (U), as is the Xvfb rendering cost.

## Effect on the tier A estimate

The mechanism risk (can an IAT clock hook scale a 40 ms frame gate under Wine, stay monotonic and consistent, and cost nothing) is closed. `hook.c` is about 100 lines and is reusable as the core. The 6 to 12 agent-hour estimate STAYS, because the budget is dominated by what P0 cannot see: forwarding the real IJL11 exports, patching all game modules, the Game.exe-specific waits and timers, the hook log and hash rule of section 5 of the parent doc, and the P1/P2 live runs at k = 1, 2, 4 on a hidden instance. The hook itself now sits at the low end (roughly 2 to 3 agent-hours to adapt: rate file, module list, `LoadLibrary` hook, forwarding stub, evidence log). Risk stays "medium", with the unknowns moved to client-server pace and rendering, not to the clock mechanism.

## Shared-file lines for the merger

See `openspec/changes/lockstepp0/merge-lines.txt`.
