## Why

The user asked (2026-11-01) for an honest README section on what the current build can do and a section on how to run it, with examples, also in the docs and linked from the README.

## What Changes

- docs/capabilities.md: what the current build can and cannot do, a Capability / Status / Evidence table (status words limited to Proven live, Works offline (tests), Experimental, Not yet verified, Known broken), every row linked.
- docs/running.md: how to run d.idle, with each example marked as run or not run by the author; the offline dry run was run, everything that needs a game is marked not run.
- README.md: two short sections ("What it can do today (honestly)" and "Running it") before the backend detail; existing text, header and links unchanged.

## Impact

Docs only. No spec, ADR, glossary or decision change.

Overlap (declared): README.md is not listed in docs/OWNERS; no other open card owns it.

## Shared-file lines for the merger

- INDEX: `docs/capabilities.md` what the current build can and cannot do (table with evidence links); `docs/running.md` how to run d.idle (examples marked run or not run).
- CHANGELOG: "readmedocs: README sections 'What it can do today (honestly)' and 'Running it', docs/capabilities.md, docs/running.md".
- BACKLOG: none.
