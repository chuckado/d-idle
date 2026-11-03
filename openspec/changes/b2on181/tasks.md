# Tasks

```yaml
change: b2on181
status: in-progress
agent: b2on181
branch: agent/b2on181
worktree: d2lod-wt/b2on181
skip_specs: true
owns:
  - openspec/changes/b2on181/**
  - docs/research/b2-on181.md
exclusive:
  - fork-branch:local/batch2-on181-r2-r2
gates:
  - "local/batch2-on181 = PR 181 head + the four batch 2 commits, pushed to the private fork remote"
  - "CI flags g++ / clang++ and the suites pass on that tree; counted differentials re-run, results compared with the r3 head"
  - "changecheck 0 warnings; leakcheck -mode fail 0"
result: "branch local/batch2-on181-r2 (b6d1be24) on PR 181 head, one comment conflict, pr_guard fixed with two oracle scripts, suites and CI flags rc 0 on the lab Windows box (WSL); counted differentials and MSVC lane NOT run (the lab Windows box (WSL) VM wedged by concurrent jobs); not verified live"
credit:
  component: Batch 2 on PR 181
  author: owner
  traced_by: []
  agents: [Claude Code]
  verified_by: []
  consent: none
merge:
```

## 1. Work

- [x] 1.1 Fetch pull/181/head read-only, local/batch2-on181 = it + four commits, conflict resolved and explained
- [x] 1.2 CI flags, extra warnings, suites, pr_guard per commit on the lab Windows box (WSL); NOT done: clang-tidy result, MSVC 14.51
- [ ] 1.3 Counted differentials against the real d2d build of that tree (the lab Windows box (WSL)): NOT done, see docs/research/b2-on181.md section 4
- [x] 1.4 docs/research/b2-on181.md, merge-lines.txt, OWNERS lines
