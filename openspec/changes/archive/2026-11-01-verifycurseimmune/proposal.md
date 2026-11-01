## Why

docs/process/upstream-pr-standard.md section 5: every upstream candidate needs an independent adversarial verifier before the coordinator may ask the user for a go. Candidate: fix/curse-immune (a curse or enemy aura lowering a monster resistance counts a fifth when the monster's base is 100 or more; card skresistfix), claim 6a04cb3e, tooling 557d14e3, on upstream main 72132b99.

## What Changes

- Own reading of FUN_005c3540, the curse callback FUN_005c35c0 and the aura callback FUN_005cedc0, the boss-mod setter in FUN_005a1370, and the dm formula FUN_00645b20.
- Own oracle (tools/emu/verify_curse_immune.py) and d2d harness (tools/emu/test_verify_curse_immune.cpp) with the REAL curse and aura skills and levels (the value calcs included), run against base, claim alone and head.
- Side claims (D2 untouched and unreachable, no other caller), textual merges with the open PRs and local candidates in both orders, and the rebased copy on the new upstream main.
- Verifier record docs/research/pr-curseimmune-verifier.md with a verdict.

## Impact

Docs and verifier tooling only. No game, no d2d slot. Nothing pushed or posted; the author's branch and worktrees untouched. Overlaps: none (docs/research/pr-curseimmune-*.md and pr-bundle-queue.md are not edited; lines for them are in the verifier's closing section).
