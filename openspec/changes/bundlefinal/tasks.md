# Tasks

```yaml
change: bundlefinal
status: in-progress
agent: bundlefinal
branch: agent/bundlefinal
worktree: d2lod-wt/bundlefinal
skip_specs: true
owns:
  - docs/research/pr-bundle-1-body.md
  - openspec/changes/bundlefinal/**
exclusive:
  - fork-branch:local/bundle-2
  - fork-branch:stage/bundle-2
gates:
  - "10 commits on current upstream main, bodies of at most three lines, author and committer the numeric noreply identity"
  - "credit.py check passes for the bundle's cards"
result: "offline on the worker host, not verified live. local/bundle-2 4ae66eaa (upstream main 82ea6db4) and stage/bundle-2 dab01645 (on stage/base-2 c4fcb0a6), pushed to the private remote only; 8 suites rc 0, freeze oracle 155,748 cases 0 differ, pr_guard 0/0 per commit, presubmit 0 FAIL; PR body in docs/research/pr-bundle-1-body.md"
credit:
  component: Final PR body for bundle 1
  author: owner
  agents: [Claude Code]
  verified_by: ["Claude Code (card verifybundle, independent run)"]
  consent: none
```

- [x] 1 local/bundle-2 and stage/bundle-2 on the private fork, suites and guards re-run
- [x] 2 credit blocks on the bundle's cards, `credit.py check` ok
- [x] 3 PR body written, numbers from the verifier and the author's counted runs only
- [x] 4 changecheck, leakcheck, ci check, openspec validate
