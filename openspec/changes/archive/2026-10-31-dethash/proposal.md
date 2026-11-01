## Why

The upstream PR standard requires a deterministic state-hash run on a candidate. Three candidates (chill, walk velocity v3, PR #155) have that step open. This card runs the scripted det scenarios (the lab's det6.py) on d2d builds of upstream main and each head and diffs the hashes. no observable behaviour change: research docs only in this repo.

## What Changes

- docs/research/det-hashes-2026-11-01.md: commands, baseline, per-candidate table, verdicts.
- docs/research/pr-bundle-queue.md rows (det step) and docs/BACKLOG.md (one line).

owns: docs/research/det-hashes-2026-11-01.md, docs/research/pr-bundle-queue.md (chill and walk rows), docs/BACKLOG.md (one line), openspec/changes/dethash/**
exclusive: none (offline; headless d2d slot on the worker host, no game.exe)
gates: none (research)
