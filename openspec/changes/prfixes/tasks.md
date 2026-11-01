# Tasks

```yaml
change: prfixes
status: in-progress
agent: prfixes
branch: agent/prfixes
worktree: ~/code/d2lod-wt/prfixes
skip_specs: true
owns:
  - docs/research/pr-walk-draft.md
  - docs/research/pr-walk-evidence-pack.md
  - docs/research/pr-price2-draft.md
  - docs/research/pr-price2-evidence-pack.md
  - docs/research/pr-bundle-queue.md
  - openspec/changes/prfixes/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge: ""
result: "offline (no game, no d2d slot, nothing pushed): fix/walk-velocity-v3 head 8db11206 (two ponytail notes, four comment lines vs v2 10338345; walk3 on the worker host) and fix/vendor-prices-followon-v2 head da80d0ab (strict=True at price_claims.py:41; price4); ruff clean, pr_guard 0/0 per commit, test binaries rc 0, gcc and tidy 0 on added lines, walk 122,646 cases 0 mismatches and price 220,000 cases clean 0/0/0 unchanged; open: det state hashes (needs a slot), price waits for #152"
```

## 1. Walk velocity v3

- [x] 1.1 Two ponytail notes (pets.cpp untraced pet, fight/tick.cpp pets in town) fixed up into the pet commit; tree diff only those lines
- [x] 1.2 CI-class checks and `test_game --walk-velocity` unchanged

## 2. Vendor prices follow-on v2

- [x] 2.1 `zip(..., strict=True)` fixed up into the tooling commit; ruff clean on every added Python file
- [x] 2.2 CI-class checks and the price follow-on mode unchanged

## 3. Docs and close

- [x] 3.1 pr-walk and pr-price2 drafts and evidence packs, the two queue rows
- [x] 3.2 changecheck, leakcheck, openspec validate
