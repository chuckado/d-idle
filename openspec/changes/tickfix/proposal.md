## Why

Six of six real tick_post runs fell back to the plain rendering with output_tokens=600. Cause (verified live, one call): claude-haiku-5-5 thinks by default and thinking tokens count against max_tokens (usage: output_tokens 263 of which thinking_tokens 222, for a tiny tick); real ticks hit the 600 cap and were cut off.

## What Changes

- `tools/ci/anthropic_http.py`: optional `thinking` request parameter and a `meta` dict (stop_reason, usage); pr_review is unchanged.
- `tools/ci/haiku_tick.py`: `thinking: {"type":"disabled"}` (accepted on Haiku 5.5 at its default effort; no temperature sent), MAX_TOKENS 1200.
- `tools/discord/tick_post.py`: a cut-off answer (stop_reason max_tokens) is an error; the fact check is set based (numbers, names of list items, capitalised or id-like tokens; reworded text passes, a dropped or invented number fails); stderr says which check failed, the stop_reason and the first 80 chars of a bad line (never the key); prompt tells Haiku to keep numbers and names verbatim.
- Tests (fake HTTP) and `docs/process/api-credits.md`.

## Impact

Overlap (declared): the files belong to the archived tickwire card and the earlier haiku tick card; edits are small. Real API use: 1 diagnosis call, 4 calls in a first fixed run that tripped the instance-id guard on my invented sample fact (not a code defect), then 2 clean dry runs; no Discord send.

## Shared-file lines for the merger
- CHANGELOG: `The Haiku dispatcher tick no longer falls back every time: thinking is disabled and the fact check tolerates rewording (tools/discord/tick_post.py).`
