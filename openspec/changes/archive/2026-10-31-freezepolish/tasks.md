# Tasks

```yaml
change: freezepolish
status: archived
agent: freezepolish
branch: agent/freezepolish
worktree: ~/code/d2lod-wt/freezepolish
skip_specs: true
owns:
  - openspec/changes/freezepolish/**
  - docs/research/pr-freeze-draft.md
  - docs/research/pr-freeze-proof.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
exclusive: []
gates:
  - "pr_guard 0 FAIL / 0 WARN on the range and each commit; author and verifier differentials unchanged (0 and the 6 known 0x100001 cases)"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge: 1ee44ef
result: "local only, nothing pushed: fix/freeze-rules-v2 head 409903a6 on the fork's #155 head 8c56e96a, commit bodies two lines plus the Agent line; author differential 225,748 cases 0 mismatches, verifier differential 146,591 cases 6 (length 0x100001), pr_guard 0/0, test binaries ok, static-field and cold-length modes 0 differ, clang-tidy and gcc 0 on added lines; NOT verified: det state hashes (need a d2d slot); open: rebase after #155 merges, the user's go"
```

- [x] 1.1 Branch fix/freeze-rules-v2 from the fork's #155 head, cherry-pick, cut the commit bodies
- [x] 1.2 Re-run the gates on the new head
- [x] 1.3 Draft body and proof doc updated
- [x] 1.4 Queue row and BACKLOG line
- [x] 1.5 Remove the throwaway worktree on the worker host
