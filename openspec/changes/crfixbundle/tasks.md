# Tasks

```yaml
change: crfixbundle
status: in-progress
agent: crfixbundle
branch: agent/crfixbundle
worktree: d2lod-wt/crfixbundle
skip_specs: true
owns:
  - docs/research/crfix-bundle-20261102.md
  - docs/research/crfix-bundle-20261102/**
  - docs/research/pr-bundle-1-body.md
  - openspec/changes/crfixbundle/**
exclusive:
  - fork-branch:local/bundle-3
  - fork-branch:stage/bundle-3
gates:
  - "every comment judged with evidence; the two Functional Correctness items proven or refuted by a build or the game.exe oracle"
  - "bundle 3 commits: author and committer the numeric noreply identity, at most 3 body lines, last line Agent: Claude Code"
result: "bundle 3 pushed to the private remote (local f68c3e0d, stage 053a0507); 6 comments judged (all valid, docstring warning declined); verified offline against game.exe under the emulator, no live game"
credit:
  component: Review fixes for bundle 2 (bundle 3)
  author: owner
  agents: [Claude Code]
  verified_by: []
  consent: none
```

- [x] 1 comments read and judged
- [x] 2 repair price and ground-roll comparison proven against game.exe / a real run
- [x] 3 local/bundle-3 and stage/bundle-3 built, suites and guards re-run, pushed to the private remote
- [x] 4 reply texts, PR body update, #176 merge probe in the doc
- [x] 5 changecheck, leakcheck, ci check, openspec validate
