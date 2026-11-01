## Why

Three traces (SKLIGHT D1/D2, SKNOVA D1, SKMISS) found the same d2d deviation, D-NEXTHIT: game.exe's NextHit rule puts state 0x56 `justhit` on a struck unit for NextDelay frames, so one nova / chain / pierce missile effect hits a unit once per window and one unit a frame; d2d tracks NextHit per missile. Effects were modelled only; no differential existed.

## What Changes

- On the worker host (not in this repo): local upstream candidate `fix/nexthit` from upstream main, claim + tooling commits, with a stateful differential (d2d `test_next_hit --next-hit`, game.exe replay `tools/emu/next_hit.py`).
- In this repo: the evidence pack `docs/research/pr-nexthit-proof.md`, the PR body draft `docs/research/pr-nexthit-draft.md`, copies of the tooling (`tools/emu/next_hit.py`, `next_hit_mode.inc`, `next_hit_summary.py`), a queue line in `docs/research/pr-bundle-queue.md`, this card. Shared-file lines go to the merger in the proof's last section.
- Nothing pushed or posted.

## Impact

Docs and tools only. Bot measurements: Nova / Frost Nova (and Chain Lightning) kill rates measured on d2d are inflated until D-NEXTHIT is fixed (proof section "Our bot's measurements").
