# Tasks

```yaml
change: verifymods
status: archived
agent: verifymods
branch: agent/verifymods
worktree: ~/code/d2lod-wt/verifymods
skip_specs: true
owns:
  - tools/emu/verify_mods.py
  - tools/emu/verify_mods_mode.inc
  - tools/emu/verify_mods_compare.py
  - docs/research/pr-mods-proof.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
exclusive: []
gates:
  - "verify_mods.py vs test_verify_mods on >= 100k cases: before > 0, head 0 per claim, residuals classified"
merge: "77e5bca"
result: "see docs/research/pr-mods-proof.md Independent verifier (VERIFYMODS); offline only, not verified live"
```

- [x] re-derive the spawn / per-hit level rules from the disassembly before reading the author's docs
- [x] own generator and harness (seed masked to 32 bits on both sides), >= 100k cases, before / each commit / head
- [x] sweep d2d for other boss-stat and XP paths and readers of the changed fields
- [x] CI-class checks (pr_guard, presubmit --quick with ruff, upstream tests, gcc sweep, clang-tidy, file sizes, replay)
- [x] recompute the impact tables, review the draft
- [x] verifier section, bundle-queue verdict, BACKLOG line
- [x] det state hashes (done by modsprep)

