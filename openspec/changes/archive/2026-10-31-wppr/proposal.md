## Why

The waypoint PR candidate (decomp-queue item 3, card decompwp) proved that d2d refuses a waypoint for 10 s after every arrival where game.exe refuses only after a hostile declaration, but its after-run checked a Python copy of the patched rule and no independent verifier had run. This card finishes the LOCAL candidate to the upstream PR standard: replay on the current upstream main, an oracle harness that runs d2d's real C++ against game.exe's own code, d2d unit tests, upstream checks, and an independent Opus verifier. no observable behaviour change: research docs and an emulator script in this repo; the d2d patch lives on a local upstream branch on the worker host, nothing pushed or posted.

## What Changes

- `tools/emu/wp_refuse.py`: mirror of the upstream-style oracle (`gen` / `game` modes, diffed against `test_game --waypoint-refuse`).
- `docs/research/pr-wp-{proof,draft,evidence-pack}.md`: proof with the C++ differential before / after, upstream test results, the verifier's verbatim report, limits; PR body draft; one-page evidence pack.
- Index lines: docs/decomp-queue.md item 3, docs/research/pr-bundle-queue.md waypoint row, docs/BACKLOG.md follow-up (the bot's `WpCooldownSec` becomes unnecessary once the PR merges; the bot is not changed here).

owns: docs/research/pr-wp-*.md, tools/emu/wp_*.py, openspec/changes/wppr/**, docs/decomp-queue.md (item 3 row), docs/research/pr-bundle-queue.md (waypoint row), docs/BACKLOG.md (one line)
exclusive: none (offline; no game instance, no d2d slot, no Go changes; the upstream branch is local on the worker host)
gates: none (research)
