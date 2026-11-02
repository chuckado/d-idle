# Tasks

```yaml
change: fixci
status: in-progress
agent: fixci
branch: agent/fixci
worktree: d2lod-wt/fixci
skip_specs: true
owns:
  - docs/community/work-archive.md
  - tools/contrib/chunker.py
  - openspec/changes/fixci/**
exclusive: []
gates:
  - "tools/ci/scripts.sh passes; changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Paths fixed in work-archive.md; chunker flag has no BACKLOG excerpt. tools/ci/scripts.sh rc 0 locally. Not verified on GitHub CI."
live: "no game"
```

- [x] fix four research paths in the archive
- [x] chunker flag without free-text excerpt
- [x] run the Scripts job locally
