## Why

An audit of the public tree for copied game code and data (2026-11-01) found nothing, but listed gaps: no README line saying the repo holds no game files, five recorded fixtures without a provenance note, no CI rule that keeps pasted listings and tables out, and worker paths (`~/d2d-lab/...`) in 105 drill and soak logs that leakcheck did not flag.

## What Changes

- README: one line that the repo contains no game files. docs/fixtures.md: what the five .snap/.d2mi recordings are, what they contain (checked), how to regenerate.
- tools/ci/gamecode_guard.sh (+ allow list, counter-test), run by check.sh and scripts.sh; docs/process/gamecode-guard.md.
- cmd/leakcheck: class `labpath` (docs/probes only) with tests; tools/ci/scrub_paths.py; the scrub of docs/probes (801 paths, 105 files, text replaced only).
- Not done, by the user's choice: paraphrasing quoted upstream comments in docs/sme/d2decomped-findings*.md.

## Impact

Offline only, no spec delta.

Overlap (declared): docs/probes/drills/** is owned by `drills` (card m1-liveproof) and `drills6`: this card only replaces path text in existing logs, no value or line changes (same line counts), and merges additively; docs/INDEX.md and docs/BACKLOG.md are owned by `skmiss` (one added line each) and CHANGELOG.md, README.md, tools/ci/check.sh, tools/ci/scripts.sh, tools/ci/README.md are standing `main` files (one added block each).
