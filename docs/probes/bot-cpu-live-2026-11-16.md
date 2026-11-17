# Bot CPU on the sandbox, live profile (card botcpulive, 2026-11-16)

> status: partial. Calm-window profile captured (V); NO profile inside a burst (the built-in profiler is one-shot, see 3). Follows
> docs/probes/bot-cpu-2026-11-16.md (offline, could not reproduce) and docs/video-feed.md section 12.

## 1. What was done (V)
Sandbox container on the worker host, image `devp9`, user go 2026-11-16. Backup first (`d2play-data-pre-main1116f.tgz`, 1578 entries).
One recreate with `D2DCTL_PPROF=/data/pprof/cpu.prof` (env overlay `pprof.yml`) and the CPU quota 2 -> 3 (`cpu2.yml`); the tuning
overrides live in the data volume and were unchanged (invdesk.verbs, invdesk.instant, bot.farmReset, watch.paletteFeed checked after).
Second recreate without the overlay. No bot code changed. The profile holds no secrets (function names, timings only).

## 2. Container numbers (V, cgroup `cpu.stat` and `/proc/<pid>/stat`, 5 s samples)
| | quota 2 (28 min before) | quota 3 (6 min after) |
|---|---|---|
| periods throttled | 4691 of 17197 (27 percent) | 92 of 3902 (2.4 percent) |
Burst seen live at quota 3: the bot alone at 60 to 100 percent of ONE core for 70 s (22:23:45 to 22:25:42), calm stretches before at
20 to 35 percent; d2d 25 to 45 percent throughout, container total up to 137 percent. At quota 1 that is throttling by construction.

## 3. Profile (V): calm window only
`D2DCTL_PPROF` writes ONE 30 s CPU profile starting 20 s after d2bot starts (`internal/d2denv/env.go` `cpuProfile`; no signal, no
debug verb, no HTTP pprof in d2bot). After a recreate the 20 to 50 s window falls on the login phase: bot at 8.3 percent of a core
(2.48 s CPU in 30 s) while the container ran at 45 to 65 percent. The burst that began 2 minutes later could not be profiled: a
second window needs a bot restart (SIGHUP to the boot shim, refused by the permission layer; another recreate not requested).
Top, cumulative (percent of the 2.48 s):

| Share | Function | Path |
|---|---|---|
| 29.8 | `currency.(*Ledger).load` (reads the whole `ledger.jsonl`, json Unmarshal then Marshal per entry for the hash check; the file is 312 KB and grew from 52 KB in days) | called by `Ledger.Entries` (17 %) and `Ledger.Balance` (13 %), from `watch.(*TallyStrip).View` -> `Tally.SummaryFor` and `Ledger.Levels`, i.e. a viewer polling the tally strip |
| 14.5 | `data.(*DB).IsA` (+`IsA.func1`, map lookups and a fresh allocation per item) | `bot.dataCatalog.Item` <- `Runner.kitOf` <- `trackAlive` <- `observe` <- `pollWith` <- `guardOnce` (bot guard, every poll, per item); also `beltAndInv` from `OrderGuardWhy` (web orders map) |
| 15 | `d2denv.wallStacks` (the profiler's own `runtime.Stack` sampler) | the instrument, ignore |
| 15 | GC (`gcBgMarkWorker`) | follows the allocations above |
| 47 | `watch.(*Access).viewGate` / HTTP handlers in total | viewer-driven work in the bot process |

## 4. Reading (S/U)
Not proven for the burst (U): no profile of it. What the calm profile shows is the two per-call costs that scale with state and
that a town/repair trip multiplies: the guard classifies every item on every poll (`kitOf` -> `IsA`, grows with items carried,
largest in town with a full inventory and during stash/sell/repair passes), and the ledger is fully re-parsed and re-hashed on every
balance or tally read (grows with every transaction; a repair trip writes several). Suspected fixes: cache `IsA(code, class)` results
per code (the data is static), and keep the ledger state in memory with an append-only tail (verify the chain once at open, not per
read). Next live step (needs a go): recreate with the overlay again and time it so the town loop lands in 20 to 50 s, or add a
bot-side trigger (signal or knob) for the profiler.

## 5. State left
Profiling off, `/data/pprof` removed, quota kept at 3 (27 percent throttled at 2 is above 10). Rollback: `cp cpu2.yml.orig cpu2.yml`
in the deploy directory and recreate.
