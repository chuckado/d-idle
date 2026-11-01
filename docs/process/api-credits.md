# API credits: PR review and the optional Haiku tick

The owner has an Anthropic Console API key funded by a promo credit. That credit is for direct API calls only; it is not valid for Claude Code, so it never pays for agent sessions. It is used for two small things.

## What the key is for
- Automatic PR review (`.github/workflows/pr-review.yml`, `tools/ci/pr_review.py`): one advisory summary comment per PR, updated on each push.
- An optional Haiku tick writer (`tools/ci/haiku_tick.py`): turns tick facts into the text for `tools/discord/feed.py tick`. Wired into the dispatcher tick through `tools/discord/tick_post.py` (see below).
- Home Assistant use is the owner's own step and is out of scope here.

## Where the key lives and the never-print rules
- Local use: a file in the owner's home under `.local/share/d2idle-api/` named `anthropic-key`, mode 0600. Only `haiku_tick.read_key` reads it (called by `haiku_tick.py --send` and `tick_post.py`), and only to put it in the `x-api-key` header sent to api.anthropic.com.
- GitHub use: the repository secret `ANTHROPIC_API_KEY`. The owner sets it: `gh secret set ANTHROPIC_API_KEY -R shatteredeagle/d-idle`. Agents never set it.
- Never print, log, echo, commit or paste the key; never put a key-shaped literal in a test (tests build fake keys from fragments); `anthropic_http.py` raises errors that carry only the HTTP status and the API error type.
- Agents do not read the file. The tests and dry runs use a fake HTTP layer and make no call.

## The owner-only gate (PR review)
The job runs only when all of these hold: the event is `pull_request` (opened, synchronize, ready_for_review); the PR head repo equals this repo (no fork PRs); the actor is the repository owner; the PR is not a draft. Strangers and forks cannot spend the credit (fork PRs cannot read secrets anyway). The job uses `continue-on-error`, so it never blocks a merge, and the comment says when a review was skipped (no secret, too many files, nothing reviewable, an error). The workflow checks out the base commit, so the script that runs is the one already on main; the PR that introduces it gets no review.

## What the review sends and caps
- The PR title, the first 2000 characters of the description, and the diff from the GitHub API with the workflow token.
- Diff cap 120 KB (cut at a line boundary, files over the cap are listed as omitted); skipped files: `docs/probes/**`, `docs/research/**`, images, lock and sum files, generated Go, JSON, testdata, CHANGELOG; refused above 60 changed files; `max_tokens` 1500; default model `claude-sonnet-5-5`, override with `REVIEW_MODEL=claude-haiku-5-5`.
- Local dry run (no API call): `DRY_RUN=1 GITHUB_TOKEN=... GITHUB_REPOSITORY=... PR_NUMBER=... python3 tools/ci/pr_review.py` prints the request size.

## Cost guesses (estimates, not measured)
A capped review is at most about 30k input tokens and 1.5k output tokens. At published per-token prices that is on the order of cents to a few tens of cents per review on Sonnet, and several times less on Haiku. A tick is under 1k tokens in and 600 out: a fraction of a cent on Haiku. Check the Console usage page after the first week and adjust the caps. These numbers are guesses until then.

## Haiku tick: optional, and what it does and does not save
`echo '{"merged":["x"],"waiting":"y"}' | python3 tools/ci/haiku_tick.py` is a dry run (request size and the facts). `--send` makes the call. Honest limit: this saves Claude Code plan tokens only if idle ticks are handled by a script and a session is woken only when there is work. If a Claude Code session still runs every tick and merely calls this helper, it adds API spend and saves nothing.

## What stays on the plan
All agent sessions, coding, merges and the dispatcher's decisions stay on the Claude Code plan. The API credit is not used for them.

## Revoke
Console, API keys: delete the key (kills both uses at once). Then `gh secret delete ANTHROPIC_API_KEY -R shatteredeagle/d-idle` and remove the key file. Without the secret the review posts a "skipped" comment and nothing else.

## Haiku tick: wired (card tickwire)
`tools/discord/tick_post.py` reads the tick facts JSON on stdin (keys running, merged, launched, upstream, blocked, waiting), has Haiku (`claude-haiku-5-5`, max_tokens 600) write the field text, checks that every fact survived and nothing was invented, then validates with the `feed.py tick` path (idle refusal, 3800 cap, notify sanitizer with the privacy guard). Dry run is the default (it still makes the one API call and prints text, chars and tokens); `--send` posts one embed to #progress through `feed.send`; `--no-api` renders without a call.
- Fallbacks, never blocking the tick: if the guard would refuse the Haiku text (for example a run of five 4-to-6-letter words, the cdkey pattern), Haiku is asked once to reword the matched phrase; if that still fails, or on any API failure (no key file, HTTP error, timeout, text that drops a fact), the plain deterministic rendering of the facts is posted and stderr says so. If the plain text trips the guard too, nothing is sent (exit 2), same as `feed.py tick`.
- The old direct `feed.py tick` call stays the documented fallback in the dispatcher notes.
- Needs `tools/ci/haiku_tick.py` and `tools/ci/anthropic_http.py` next to `tools/discord` when extracted.
- Honest scope: this saves the coordinator session only the tokens of writing the tick text (about 600 to 900 output tokens). Decisions (merges, launches) still need a Claude Code session; the tick text is still produced inside that session's turn, so the saving is real only for the writing step.

Measured (V, the wiring test of 2026-11-01, real API usage field): dry run 292 input and 459 output tokens for a 156 character tick; the sent post 265 input and 331 output tokens for 74 characters. Output counts exceed the visible text, so the model apparently spends tokens beyond the text (U: not investigated). About 300 in and 300 to 500 out per tick is a fraction of a cent on Haiku; at several ticks a day it is negligible against the credit.

## Haiku tick: why it always fell back, and the fix (card tickfix)
- Cause (VERIFIED LIVE, one dry-run call): `claude-haiku-5-5` runs adaptive thinking by default and thinking tokens are billed inside `max_tokens`. A tiny tick used 263 output tokens, 222 of them `thinking_tokens` (usage.output_tokens_details). Real ticks reached the 600 cap, the answer was cut off and failed the check, so six of six runs used the plain rendering.
- Fix: the call sends `thinking: {"type":"disabled"}` (accepted on Haiku 5.5 at its default effort; non-default `temperature` is rejected there, so none is sent) and `max_tokens` 1200. A response with `stop_reason: max_tokens` is treated as a failure.
- The fact check is set based: numbers, every word of list items (card names), and capitalised or id-like tokens must survive; a number that was not in the facts fails. Rewording passes. When the API path fails, stderr gives the reason (stop_reason, which check, first 80 chars of a bad line); never the key.
- Measured per tick (real dry runs, 6 fields, 12 items): 350 input and 98 output tokens, one call, source=haiku, about 0.02 cent at $0.10 / $0.50 per MTok. A reword adds one similar call.
