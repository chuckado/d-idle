# Tasks

```yaml
change: armorfinish
status: in-progress
agent: armorfinish
branch: agent/armorfinish
worktree: ~/code/d2lod-wt/armorfinish
skip_specs: true
owns:
  - openspec/changes/armorfinish/**
  - docs/research/pr-armorhit-addendum.md
  - docs/research/pr-armorhit-body-final.md
gates:
  - "changecheck 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Docs only, offline, no game. Addendum, final body and ready checklist written; one status banner on pr-armorhit-draft.md (declared overlap with card armorhit, text of the banner only). Upstream branch fix/armor-hit checked on the worker host: head d5bd19c1, parent 36cbe015. Not opened: awaiting the user go."
```

## 1. Work

- [x] 1.1 Addendum with the verified limits and proof rows
- [x] 1.2 Final PR body and ready checklist
- [x] 1.3 Banner on the old draft (overlap with card armorhit, declared)
- [x] 1.4 Upstream branch check on the worker host
