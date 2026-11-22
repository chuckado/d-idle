# Bot CPU fixes after the live calm-window profile (2026-11-16)

> status: current. Follows docs/probes/bot-cpu-live-2026-11-16.md (Ledger.load 30 percent, DB.IsA 14.5 percent). All numbers are offline Go benchmarks on the 4-core coordinator box (V: `go test -bench`); NOT yet measured live (U).

## What changed
| Fix | Where | Before | After |
|---|---|---|---|
| Ledger keeps the parsed state; a read stats the file and parses only appended bytes (full re-read on shrink, new inode, same size with new mtime, and in `Verify`) | `internal/currency/ledger.go` | `Entries` 312 KB: 9.9 ms, 4.9 MB alloc; 5 MB: 156 ms, 85 MB | 0.02 ms, 1.8 KB, both sizes |
| One viewer poll of the tally strip (3 summaries + levels) | `BenchmarkTallyPoll` | 312 KB: 63 ms, 34 MB alloc (5 MB about 0.6 s, derived) | 312 KB: 3.1 ms; 5 MB: 12 ms |
| Tally strip view cached 1 s, dropped on every strip event, credit and order | `internal/watch/tallystrip.go` | one recompute per poll | one per second at most |
| `DB.IsA` reads a per-type ancestor list built at load | `internal/data/db.go` | 117 ns | 27 ns, 0 alloc (synthetic 3-deep table; the real table is deeper) |
| `dataCatalog.Item` memoises `ItemMeta` per class (about 20 IsA calls per item before) | `internal/bot/catalog.go` | 448 ns | 6.7 ns, 0 alloc |

Equality: `TestLedgerCacheEqualsFullLoad` (two writers on one directory, two characters and runs, compared with a fresh full load after every step), `TestLedgerCacheDetectsTruncationAndTamper`, `TestLedgerCacheTornTail`, `TestIsAMatchesWalkOverWholeTable` (all code pairs incl. cycle and dangling Equiv against the old walk), `TestCatalogItemMemoEqualsUncached`.

## Known limit
A same-size, same-mtime in-place edit of an old ledger line is not seen by a poll read (the old code re-hashed every time); `Verify` and the next process start still re-read in full. Appends are fsynced and chained, so a rewrite that keeps the size and the mtime needs a deliberate attack on a local file.

## Remaining remarks (not changed)
- `SummaryFor` still runs `time.Parse` per entry three times per uncached strip build (the 3 ms left); `Hub.Currency` and the hall view scan the cached entries without reloading. `gear.IsA` still walks per call (a few calls per poll, equipped items only).
- The burst itself (60 to 100 percent of a core for about 70 s) is not explained by these two hot spots alone; capture it with the endpoint below.

## Capturing a burst without a recreate
Start the bot with `D2BOT_PPROF_HTTP=1` (off by default; there is no knob, a knob would be writable from the API). The routes sit on the answer listener behind the same token or passcode and Origin rule as the other writes. Files go to `<state dir>/pprof/cpu-<time>.pprof`, the newest 5 stay.

```
A=http://127.0.0.1:<answer port>; H="Authorization: Bearer $(cat <watch token file>)"
curl -X POST -H "$H" "$A/api/debug/cpuprofile?secs=60"     # 1..60, 409 while one runs or D2DCTL_PPROF holds the profiler
curl -H "$H" $A/api/debug/cpuprofile                       # {running,file,leftSecs}
curl -X POST -H "$H" $A/api/debug/cpuprofile/stop          # end early, keep the file
curl -H "$H" -o cpu.pprof $A/api/debug/cpuprofile/file     # the newest finished profile
go tool pprof -top cpu.pprof
```
