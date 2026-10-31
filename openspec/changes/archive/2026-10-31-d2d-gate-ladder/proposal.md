## Why

User decision 2026-10-31 ("retire g1", ADR 0013): the 2 h and 8 h gate soaks on the original Game.exe under Wine (G1, G2) are retired as release gates. The same gates run on d2d as a ladder: Gate 0 drills, Gate 1 2 h, Gate 2 8 h, Gate 3 24 h. ADR 0009 stands; Game.exe stays frozen and stays the reference.

## What Changes

- delta spec for soak-gates: ADDED Gate 0 to Gate 3 on d2d, translated G1 criteria and the fault injection; REMOVED the four Game.exe G1 requirements; MODIFIED G2 meaning and the start guard wording
- docs/review-strategy.md: status banner and a d2d gate ladder subsection; docs/g1-log.md banner
- ADR 0013 and its decisions.md line; a proposed dispatcher step 4 (needs the user's go)
- not changed: ADR 0009, the frozen Game.exe path, the decomp and fidelity lanes
