## Why

The user asked (2026-11-01) for a roadmap chunker that also grooms the backlog on later runs, checks that every open work item is still needed, and keeps the board and the GitHub issues current without a human closing them ("like an outside helper would").

## What Changes

- `tools/contrib/chunker.py sync`: deterministic, no model API, runs on every dispatcher tick. Reconciles docs/community/open-work.md, the `work-item` issues and the board: creates missing issues, closes done or withdrawn ones (one templated comment), fixes body drift, handles stale claims (one nudge after 14 days, claim released by comment 7 days later). Dry run by default, per-run write cap.
- `tools/contrib/chunker.py groom`: weekly or on demand, proposals only. Sends public repo text (privacy-scanned) and issue counts/dates to one model call, validates the JSON and every cited path deterministically, writes docs/community/groom-proposal.md. Never edits open-work.md.
- docs/process/chunker.md: how the coordinator uses it each tick, caps, safety rules, measured cost.

## Overlap (declared)

tools/contrib/board.py and docs/community/open-work.md / issues.tsv / work-board.md belong to the archived communitywork and boardlinks work. chunker imports board.py (parse, validate, issue_body, render) and does not edit it. `sync --apply` rewrites issues.tsv and work-board.md through board.py; the coordinator commits them. A `doc` item kind in a groom proposal needs a small board.py change first (not made here). tools/ci/scripts.sh already discovers every test in tools/contrib/tests, so it is not edited. New file docs/community/withdrawn.tsv (id and reason for items removed from open-work.md) is owned here.

## Impact

No game, no instance, no Go. Python stdlib only. The only network use is `gh` (read by default, writes only with `--apply`) and one Messages API call per groom run.
