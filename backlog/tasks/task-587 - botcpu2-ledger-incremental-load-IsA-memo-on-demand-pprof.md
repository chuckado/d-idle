---
id: TASK-587
title: 'botcpu2: ledger incremental load, IsA memo, on-demand pprof'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T03:44:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 302000
card: "botcpu2"
owns:
  - "internal/currency"
  - "internal/data/db.go"
  - "internal/bot/catalog.go"
  - "internal/watch/tallystrip.go"
  - "internal/watch/pprofapi.go"
  - "docs/probes/bot-cpu-fix-2026-11-16.md"
  - "internal/watch/serve.go"
  - "internal/watch/tallystrip_test.go"
  - "cmd/d2bot/ladder.go"
  - "docs/knobs.md"
exclusive: []
contracts: []
result: "Ledger reads only the appended tail (Entries 312 KB 9.9 ms to 0.02 ms; tally poll 63 ms to 3.1 ms), IsA 117 to 27 ns, ItemMeta memo 448 to 6.7 ns, 1 s tally view cache, on-demand CPU profile endpoint behind D2BOT_PPROF_HTTP=1. Offline benchmarks only, not verified live. docs/probes/bot-cpu-fix-2026-11-16.md"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 currency.Ledger reads only appended bytes (equality tests vs full load, benchmarks)
- [x] #2 IsA ancestor table and ItemMeta memo (equality over a whole synthetic table, benchmarks)
- [x] #3 authenticated on-demand CPU profile endpoint behind D2BOT_PPROF_HTTP=1
<!-- AC:END -->
