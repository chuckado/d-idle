## Why

The owner has Console API credit that is not valid for Claude Code. It can pay for an automatic PR review on the public repo and for a cheap Haiku tick writer.

## What Changes

- `.github/workflows/pr-review.yml` and `tools/ci/pr_review.py` (+ `anthropic_http.py`): owner-only, same-repo-only advisory review comment, size and file caps, fail soft.
- `tools/ci/haiku_tick.py`: optional, not wired, dry run by default.
- `tools/ci/tests/test_pr_review.py`: fake-HTTP unit tests (`python3 -m unittest discover -s tools/ci/tests -t tools/ci`).
- `docs/process/api-credits.md`.

## Impact

Offline only; no API call was made and no secret was read or set. The owner must run `gh secret set ANTHROPIC_API_KEY -R shatteredeagle/d-idle`. Overlap (declared): `.github/workflows/ci.yml` is main's; this card adds a separate workflow file and does not edit it.

## Shared-file lines for the merger
- docs/INDEX (process list): `docs/process/api-credits.md: API credit use, PR review gate, optional Haiku tick, revoke`
- CHANGELOG: `Automatic advisory PR review (owner PRs from this repo only, needs the ANTHROPIC_API_KEY secret) and an optional Haiku tick helper; see docs/process/api-credits.md.`
- tools/ci/scripts.sh or ci.yml: optionally add `python3 -m unittest discover -s tools/ci/tests -t tools/ci` to the checks job.
