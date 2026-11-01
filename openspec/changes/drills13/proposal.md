## Why

Two live questions the offline cards left open: does the Frozen Armor recast of card fixrecast hold on a real d2d instance (state switch, real duration, the recast time the bot chooses), and when does d2d pay a spell's mana relative to the cast command and the action frame (drill M1 of docs/research/sk-mana-trace.md, D1 and D3).

## What Changes

- New drill group `Q` in cmd/d2drill and internal/drill (new files d13.go and d13_lab.go; a three-line hook in cmd/d2drill/main.go): D-fro1, D-fro2, D-m1.
- Result logs `docs/probes/drills/2026-11-01-drills13-*.txt` and a DRILLS13 section in docs/probes/drills/README.md.
- No product code changes (findings are reported, not fixed).
