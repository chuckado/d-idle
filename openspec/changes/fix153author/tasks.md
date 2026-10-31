# Tasks

```yaml
change: fix153author
status: in-progress
agent: fix153author
branch: agent/fix153author
worktree: ~/code/d2lod-wt/fix153author
skip_specs: true
owns:
  - docs/research/pr-wp-evidence-pack.md
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0 hits"
merge:
result: "PR branch rewritten and force-pushed (tree identical); local checks rerun on the worker host; not verified live (no game involved)"
live: ""
```

- [x] rewrite the PR branch, tree identical, force-push with pinned lease
- [x] append the posted line to the evidence pack
