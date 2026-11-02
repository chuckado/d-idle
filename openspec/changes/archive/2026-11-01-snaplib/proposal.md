## Why

The five extra baseline runs start from one honest character restored five times (user approval 2026-11-02: file-level save copies, one Fresha restore per run). The repo has per-prefix backup and restore only (internal/saves, cmd/d2saves): no manifest, no placement into another prefix, no preflight, no restore log (docs/research/character-factory.md section 4). ADR 0014 makes a restore a logged SETUP action and a copy of an honest character copy-of-honest.

## What Changes

- internal/saves/snaplib.go: library snapshot with manifest v1, verify, list, rename of the .d2s name and checksum, place with refusals.
- internal/cheat/d2s/tail.go: `File.Tail()` for the kit hash.
- cmd/d2restore: snapshot, verify, list, place; preflight (no i0, writable ledger, no running game, instance locks), setup-copy ledger line.
- docs/snaplib.md: usage, rules, live preflight checklist.

## Impact

No spec delta. Offline only; synthetic fixtures; NOT run against the game or a real save. Overlap (declared): card ledgerkey (same agent) adds `cheat.KindSetupCopy` that d2restore writes; card memrules (agent/memrules, unmerged) adds `Entry.Run` and run.go: d2restore avoids `Entry.Run` and puts the run id in params, so either merge order builds. internal/saves and cmd/** are owned by main by default; this card adds new files only.
