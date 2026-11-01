## Why

The user (2026-11-01) wants friends to help speed up the bot and d2d work by running their own agents on decomps and verifications. Chosen: claiming through GitHub issues (one issue per work item, a friend comments `claiming`, the PR links the issue). Scope: contributor guide and templates, a work board generator, a review gate for outside PRs, and a Discord draft (a read-only `#work-board` feed and an `#agents` channel).

## What Changes

- docs/contributing/agent-work.md: the guide for a friend and their agent (what is wanted, what is not open, setup, safety and legal rules, deliverable format, evidence checklist, prompt block, our injection rule).
- .github/ISSUE_TEMPLATE/work-item.yml and .github/pull_request_template.md.
- docs/community/open-work.md (the list of items open to outside agents; kept here, not in docs/decomp-queue.md, because the skmiss card owns that file), tools/contrib/board.py with tests (renders docs/community/work-board.md and issue bodies; never calls GitHub), a hook into tools/ci/scripts.sh.
- docs/process/outside-prs.md and .coderabbit.yaml: the proposed gate (nothing switched on).
- docs/community/discord/work-channels.md, server-work.json, tools/discord/board_feed.py with tests: a DRAFT; nothing created or sent.
- One link line in README.md.

## Impact

Docs, small stdlib scripts, templates. No Go code, no spec, ADR, glossary or decisions change. No repository setting, app, label, issue, webhook or channel was created.
