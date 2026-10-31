## Why

fixstall2 (map-aware NPC approach, walking portal entry) was proven offline only. DRILLS7 runs H.pause x3 and D6.d x3 live, the D-sync arrival variants DRILLS6 skipped, and repairs the drill harness so the full D2 group passes in one process.

## What Changes

- Harness only: cmd/d2drill and internal/drill (portal state reset between D2 cases, revive after a lethal debug hit, per-case setup). No product code changes: failures are findings.
- Results in docs/probes/drills (drills7 files, README section DRILLS7), primitives rows, `live:` lines on the fixstall and fixstall2 cards.
- docs/OWNERS: the drill paths are already covered by the m1-liveproof globs, so only the card dir is mirrored.
- Edits to cmd/d2drill, internal/drill and docs/probes/drills/*drills7* are under the m1-liveproof ownership globs (docs/OWNERS); this card does not claim them, to avoid the overlap.
