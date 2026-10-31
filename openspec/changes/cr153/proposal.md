## Why

CodeRabbit reviewed the upstream waypoint PR (psi29a/d2decomped#153) with two inline findings and one nitpick. The standing go (memory d2-coderabbit-standing-go) is to answer them and fix the valid ones in a small push; the human maintainer is never answered.

## What Changes

- The upstream PR branch got one fast-forward fixup: `--waypoint-refuse` rejects zero-arrival rows, the documented `gen` example uses `--solo`, the new case variables are renamed.
- This card only records the posted actions in docs/research/pr-wp-evidence-pack.md. The claim and the 140,002-case differential (66,969 before, 0 after) are unchanged.

## Impact

Owns docs/research/pr-wp-evidence-pack.md and one BACKLOG line. No code, no spec change.
