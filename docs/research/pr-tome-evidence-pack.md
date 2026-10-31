# Evidence pack: PR candidate "Town Portal from the pack opens at once" (agent DECOMPTOME, 2026-10-31)

> status: candidate, not ready. Proof: [pr-tome-proof.md](pr-tome-proof.md); body: [pr-tome-draft.md](pr-tome-draft.md); trace: [town-portal-trace.md](town-portal-trace.md). Nothing pushed, posted or filed.

## PROOF: incomplete

Reason: the game.exe side is proven by a native run, but there is no d2d patch, no after-run, no upstream test run and no independent verifier yet.

The claim in one sentence: a Town Portal scroll or a charged tome used from the inventory opens the portal on that tick in any live mode and spends the charge only on success; d2d casts SC instead, drops the use while busy and spends the charge at the click.

| Gate (upstream-pr-standard.md section 7) | Status |
|---|---|
| 1 Differential | game.exe native (`FUN_0054d750` and its chain), 108,000 cases (exhaustive grid + 100,000 random, seed 7); before 22,395 differ; after not run |
| 2 Addresses | every link V (disasm.py read and run); pr_guard not run (no commit) |
| 3 Statistics | soak lines only as motivation (town-portal-trace.md section 3), not proof |
| 4 Scope | one claim; their tests, presubmit, d2ddet not run |
| 5 Verifier | not done |
| 6 Limits | listed in the proof, section 4.5 |
| 7 Their rules | the draft follows them; no commit exists |
| 8 Spacing | waits behind the open candidates (README "Upstream PRs") |

## Recommendation

No-go as it stands; worth finishing: it changes how often a d2d town portal works under attack, which moves trips and deaths in our soaks. Next card: write the patch on `fix/tome-use`, add `test_game --tome-use`, run the after-differential to 0, then the verifier. The bot fix (FIXPORTAL) must not wait for it: it should handle d2d's current rule (click when idle, success = object 59).

Push and PR commands: none yet (no branch with commits).
