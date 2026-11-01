# Tasks

```yaml
change: boardgrow
status: in-progress
agent: boardgrow
branch: agent/boardgrow
worktree: ~/code/d2lod-wt/boardgrow
skip_specs: true
owns:
  - tools/contrib/board.py
  - tools/contrib/tests/test_board.py
  - docs/community/open-work.md
  - docs/community/work-board.md
  - docs/community/work-archive.md
  - docs/community/issues.tsv
  - docs/contributing/agent-work.md
  - openspec/changes/boardgrow/**
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/contrib/tests passes; python3 tools/contrib/board.py --check"
  - "go run ./cmd/changecheck -cli reports 0 warnings; go run ./cmd/leakcheck -mode fail 0 hits"
merge: ""
result: ""
```

Overlap (declared): the open card `chunker` owns tools/contrib/chunker.py, tools/contrib/tests/test_chunker.py and docs/process/chunker.md. This card makes small edits there (sync and groom read the archive through `board.load_all`; one test; the archive rule in the doc). They are not in `owns:` here; the merger applies them on top of the chunker card.

## 1. Archive
- [ ] 1.1 board.py: archive file, --archive, --check, render only non-done on the board
- [ ] 1.2 chunker.py reads the archive; tests (archive keeps issue closed, nothing re-created, done closes once)
- [ ] 1.3 SKMELEE and SKRESIST moved

## 2. New items
- [ ] 2.1 12 items in open-work.md; sync dry run lists a create for each

## 3. Docs
- [ ] 3.1 docs/process/chunker.md, docs/contributing/agent-work.md

## Shared-file lines for the merger
- none for BACKLOG, INDEX or CHANGELOG (no edits there). OWNERS lines are in this branch.
