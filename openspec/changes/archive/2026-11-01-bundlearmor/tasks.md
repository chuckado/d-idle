# Tasks

```yaml
change: bundlearmor
status: archived
merge: 63a8fbb
agent: bundlearmor
branch: agent/bundlearmor
worktree: ~/code/d2lod-wt/bundlearmor
skip_specs: true
owns:
  - docs/research/pr-bundle-armor-resume-body.md
  - openspec/changes/bundlearmor/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Combined PR body and two-commit local branch ready, held for the user's go; not opened upstream. Not verified live (offline differentials). Combined local branch fix/frozen-armor-and-resume on the worker host (upstream main 72132b99; commits 6b5f0267 armor claim, a2a16a87 armor tooling, 0ab7b5e7 resume claim, a4ad2099 resume tooling; ai.cpp comment now cites callback 0x56e900, commit body callback 0x57b170). Re-run on the combined head: author armor 0 of 108,506 x2 (base 95,920 / 95,986), verifier armor end frames 0 of 134,400 (base 109,573), author resume 0 of 12,548 / 12,516 frozen (base 12,503 / 12,459), verifier resume 0 of 17,180 / 17,119 (base 17,130 / 17,070); pr_guard 0/0 on range and each commit, presubmit --quick 0 FAIL / 3 WARN, tests green, gcc and tidy 0 on touched lines, ruff clean. Det hashes for the resume claim and the combination NOT run (needs a d2d slot). Nothing pushed or posted. Combined PR body written: docs/research/pr-bundle-armor-resume-body.md (READY, held for the user go)."
live: "docs only here; offline differentials on the worker host, no game, no d2d slot"
```

- [x] combined local branch on the worker host from upstream main 72132b99
- [x] both differentials and the checks re-run on the combined head
- [x] combined PR body
