## Why

CodeRabbit reviewed upstream PR #164 (monster missile record and poison units) and left two inline findings on tools/emu/mm_diff.py (lines 222 and 226) under the standing go for CodeRabbit follow-ups. This change records the evaluation, the one follow-up commit (cd01212f on the fork branch), the two replies, and mirrors the fix to the public copy of tools/emu/mm_diff.py.

## What Changes

- tools/emu/mm_diff.py: the MMTAB table is read only when a difference needs its row data; a case without a d2d line is counted as a failure; zip() gets strict=False.
- Doc lines in docs/research/openpr-log.md.
