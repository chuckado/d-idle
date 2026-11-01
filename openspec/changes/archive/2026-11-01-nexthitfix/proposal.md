## Why

The independent verifier (docs/research/pr-nexthit-verifier.md) found one real bug in the local upstream candidate `fix/nexthit`: game.exe ends justhit one frame later than the patch (blocked f+1..f+NextDelay, next hit f+NextDelay+1), because FUN_005414d0 runs the missile timers before the monster timers. The author's oracle had the same off-by-one built in.

## What Changes

- On the worker host (not in this repo): new upstream candidate branch `fix/nexthit-v2` (worktree `pr/nexthit2`, base upstream main 72132b99): `<` becomes `<=` in the justhit end test, the `ponytail` note states the traced rule, the oracle `next_hit.py` expires justhit after the frame's contacts, `docs/research/re/skills.md:928` corrected. The old `fix/nexthit` is untouched.
- Private fork: one new commit on `lab-nexthit` (fast-forward, pushed to the private remote), new binary `bin/d2d.lab-nexthit2`.
- In this repo: `docs/research/pr-nexthit-v2-addendum.md` (rule, re-run counts of both harnesses, det hashes, the verifier's five draft edits, fork update, shared-file lines), a `> status:` banner on the old proof, a queue line in `docs/research/pr-bundle-queue.md`, this card.
- Nothing pushed to the public repo or posted upstream.

## Impact

Docs only here. Nova / Frost Nova numbers measured on `d2d.lab-nexthit` are slightly high at moving or doubled ring contacts; soaks use `d2d.lab-nexthit2` from now on.
