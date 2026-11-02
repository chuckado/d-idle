## Why
The user decided on 2026-11-02 that memory may be read and edited freely as long as it is never a crutch. The rules in AGENTS.md, docs/cheats.md and lab-rules must say so, and cheap enforcement is needed so a measured run can state how many writes happened after its start. The ledger change adds a run id and a start marker.

## What Changes
- AGENTS.md memory bullet replaced by the memory rule and the crutch rule; docs/cheats.md section; ADR 0014 and one decisions.md line.
- Delta spec lab-rules: Memory access, Crutch rule.
- internal/cheat: Entry.Run ($D2BOT_RUN), marker kind runstart, Ledger.MarkRun and WritesAfterStart; d2cheat `runstart ID` and `ledger --since-run ID`. Unit-tested.
- Two stale statements of the old rule updated (docs/d2-1.14d-memory.md, docs/research/d2d-whisper-lan-oracle.md).
