## Why

Upstream PR #153 (waypoint) failed its Linux CI step "Names and includes (Linux)": clang-tidy readability-identifier-length flagged the test local `at` (3 characters minimum) in our tooling commit. Fixed with one fixup commit on the PR branch (user standing go, as the #152 fix); this card records it in the evidence pack.

## What Changes

- Evidence pack: one dated `posted:` line (fixup pushed, old and new head, reason).
- BACKLOG (REFOCUS): the lesson, run upstream's clang-tidy names-and-includes check and the MSVC shadowing sweep before every upstream push.

## Impact

Docs only. docs/process/upstream-pr-standard.md is not owned by this card; the lesson is in the report and BACKLOG.
