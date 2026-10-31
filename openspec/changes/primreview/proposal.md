## Why

The long d2d soaks show an empty healing belt in about a third of the minutes, potion pickups skipped as "already in the pack", and a poor recovery economy. The coordinator asked for an independent, read-only bug hunt on the primitives layer (belt, pickup, consumables, NPC, input and state plumbing) that confirms or refutes the "pack-to-belt move fails" hypothesis from code and soak logs.

## What Changes

- docs/primitives-review.md: verdict on the hypothesis, a ranked findings table with code paths, log counts and fix sizes, cleared suspects, and the log lines that would settle the open questions.
- No code changes (internal/**, cmd/**): fixes are proposed only.

## Dependencies

None. Read-only on the soak dirs of the worker host (a soak may still be running).
