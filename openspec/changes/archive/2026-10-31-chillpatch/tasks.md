# Tasks

```yaml
change: chillpatch
status: archived
agent: chillpatch
branch: agent/chillpatch
worktree: ~/code/d2lod-wt/chillpatch
skip_specs: true
owns:
  - docs/research/pr-static-chill-proof.md
  - docs/research/pr-static-chill-draft.md
  - tools/emu/chill_length.py
  - tools/emu/chill_length_mode.inc
  - docs/decomp-queue.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
  - openspec/changes/chillpatch/**
exclusive:
  - fork-branch:fix/static-field-chill
gates:
  - "differential on real d2d builds: before > 0, after 0; pr_guard 0/0; presubmit --quick 0 FAIL; upstream tests rc 0"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge: "merged to main (agent/chillpatch)"
result: "Local upstream candidate fix/static-field-chill (claim 46ef6c75, tooling 79c9a496) on #155 head: C1 differential 14,239 -> 0 of 121,312 on real d2d builds; C2 not reachable (no patch); CI-class checks clean. Offline, nothing verified live, nothing pushed; verifier and det hashes open."
```

## 1. Work

- [x] 1.1 Re-derive the chill rule from the disassembly; decide C1 / C2 and the boundary cases
- [x] 1.2 Claim commit (C1) and tooling commit on fix/static-field-chill (worker host)
- [x] 1.3 Differential before / after on real builds; CI-class checks
- [x] 1.4 Proof, draft, queue rows, backlog line
- [ ] 1.5 Independent verifier (not this card)
- [ ] 1.6 det state hashes (needs a d2d slot; not run)
