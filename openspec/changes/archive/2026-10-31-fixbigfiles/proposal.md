## Why

Public CI on main (0b0bdb25) failed the hygiene gate (`tools/ci/check.sh`, limit 1 MB outside testdata): five DRILLS9 run logs under docs/probes/drills/ were 1.0 to 1.1 MB (full `-v` stderr pasted next to the result). The drills9 card is archived; no other open card owns docs/probes/drills/**.

## What Changes

- The five `2026-11-01-drills9-f*.txt` logs are capped in place (head, tail, every verdict/error line, one marker line).
- `d2drill -cap FILE` (internal/drill CapLog, tested) caps a saved log before commit; check.sh fails probe results over 512 KB; README rule.
- No product code changes.

Ownership: the open card m1-liveproof owns docs/probes/drills/**, cmd/d2drill/** and internal/drill/**. This card edits there only additively (five log files shrunk, one new file internal/drill/cap.go with its test, a flag in cmd/d2drill/main.go, a README paragraph) and owns just its own card and docs/fixbigfiles.md, so there is no overlap; tools/ci/check.sh gets one added check.
