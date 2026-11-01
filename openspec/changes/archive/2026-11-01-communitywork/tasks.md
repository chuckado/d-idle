# Tasks

```yaml
change: communitywork
status: archived
merge: 37fc3ad
agent: communitywork
branch: agent/communitywork
worktree: ~/code/d2lod-wt/communitywork
skip_specs: true
owns:
  - docs/contributing/**
  - docs/community/open-work.md
  - docs/community/work-board.md
  - docs/community/discord/work-channels.md
  - docs/community/discord/server-work.json
  - docs/process/outside-prs.md
  - .github/ISSUE_TEMPLATE/**
  - .github/pull_request_template.md
  - .coderabbit.yaml
  - tools/contrib/**
  - tools/discord/board_feed.py
  - tools/discord/tests/test_board_feed.py
  - tools/discord/tests/test_work_channels.py
  - tools/ci/scripts.sh
  - README.md
  - openspec/changes/communitywork/**
exclusive: []
gates:
  - "python3 -m unittest discover -s tools/contrib/tests -t tools/contrib; the two tools/discord tests; python3 tools/contrib/board.py --check"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Not verified live (offline-complete; board.py --check and python tests pass): docs and stdlib scripts only, nothing run live, nothing posted or created (no issue, label, webhook, channel, repo setting). Guide, templates, open-work list (8 items) with the board generator, outside-PR gate proposal with .coderabbit.yaml, Discord draft (#work-board feed script, #agents, server delta planned offline)."
```

## 1. Work

- [x] 1.1 Contributor guide
- [x] 1.2 Issue and PR templates
- [x] 1.3 Open-work list, board generator, tests, CI hook
- [x] 1.4 Outside-PR gate proposal and CodeRabbit config
- [x] 1.5 Discord draft: channels, server delta, board_feed.py, tests
- [x] 1.6 README link line; shared-file lines for the merger (below)

## 2. Shared-file lines for the merger (not edited here)

- docs/INDEX.md: `contributing/agent-work.md` (guide for friends and their agents), `community/work-board.md` (generated board), `community/open-work.md`, `process/outside-prs.md`, `community/discord/work-channels.md` (DRAFT).
- CHANGELOG.md: "Contributor guide, work board, issue and PR templates, outside-PR review gate proposal and a Discord work-channels draft (nothing switched on)".
- docs/BACKLOG.md: `COMMUNITYGO (user go needed): open the GitHub issues from the board (python3 tools/contrib/board.py --write-issues DIR), create the work-item label, branch protection and CodeRabbit per docs/process/outside-prs.md, the Discord delta and webhook per docs/community/discord/work-channels.md`.
- docs/decomp-queue.md (owned by skmiss while it is open): add after the queue table the line "Rows 19-22 also appear in docs/community/open-work.md for outside contributors; keep the two in step (python3 tools/contrib/board.py --check fails when a queue row is done and the item is not)".
- docs/OWNERS: the lines for this card (added here).
