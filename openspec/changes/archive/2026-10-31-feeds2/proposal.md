## Why

Two feeds steps were held back by the first feeds card: failures-only CI notices (GitHub cannot filter a webhook by conclusion, so `workflow_run` posted every run) and a `#d2d-upstream` webhook. The user approved exactly two secret writes on 2026-11-01.

## What Changes

- .github/workflows/ci.yml: final job `notify-failure` (posts one line to `#progress` through the repository secret `DISCORD_PROGRESS_WEBHOOK`) and a `workflow_dispatch` input `feed_test`.
- The GitHub repo webhook drops `workflow_run` (release and pull_request stay).
- A webhook `d.idle feed` on `#d2d-upstream`, stored only as a private 0600 file; `feed.py` fail-closed text made generic; tests updated.
- Docs: feeds.md sections (f) and (g), server.md, plan.md, BACKLOG, CHANGELOG.
- No numeric gate changes (no delta).

Overlap (declared): docs/BACKLOG.md and CHANGELOG.md are standing `main` docs, one added line each; tools/discord/feed.py, its test and the two discord docs were owned by the archived card `feeds`; no open card owns them or ci.yml.
