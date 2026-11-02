## Why

CodeRabbit left a third inline finding on upstream PR #164 (tools/emu/mm_diff.py line 182): cmp returns 0 when game.tsv has no cases, so an empty comparison passes. Valid; fixed in 11bc2c4d on the fork branch (plain push, reply 4439895484 to comment 4439678010). This change mirrors the fix to the public copy.

## What Changes

- tools/emu/mm_diff.py: cmp prints "no game cases in game.tsv" and returns 1 when game.tsv is empty. Negative check on the worker host: empty game.tsv rc 1 (before: rc 0). Not verified live (tooling only).
