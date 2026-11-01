# Tasks

```yaml
change: tickfix
status: archived
merge: b051a79
agent: tickfix
branch: agent/tickfix
worktree: ~/code/d2lod-wt/tickfix
skip_specs: true
owns:
  - tools/discord/tick_post.py
  - tools/ci/haiku_tick.py
  - tools/ci/anthropic_http.py
  - tools/discord/tests/test_tick_post.py
  - tools/ci/tests/test_pr_review.py
  - docs/process/api-credits.md
  - openspec/changes/tickfix/**
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/discord/tests -t tools/discord -p test_tick_post.py passes; tools/ci/check.sh passes"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "offline tests pass; verified live: 1 diagnosis call and 2 dry runs on the fixed code gave source=haiku, 350 in / 98 out tokens"
```

- [x] diagnose with one real call
- [x] disable thinking, 1200 tokens, set-based fact check, diagnostics
- [x] tests with fake HTTP
- [x] two real dry runs
- [x] docs
