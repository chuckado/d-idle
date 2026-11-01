# API credits: PR review and the optional Haiku tick

The owner has an Anthropic Console API key funded by a promo credit. That credit is for direct API calls only; it is not valid for Claude Code, so it never pays for agent sessions. It is used for two small things.

## What the key is for
- Automatic PR review (`.github/workflows/pr-review.yml`, `tools/ci/pr_review.py`): one advisory summary comment per PR, updated on each push.
- An optional Haiku tick writer (`tools/ci/haiku_tick.py`): turns tick facts into the text for `tools/discord/feed.py tick`. Not wired anywhere.
- Home Assistant use is the owner's own step and is out of scope here.

## Where the key lives and the never-print rules
- Local use: a file in the owner's home under `.local/share/d2idle-api/` named `anthropic-key`, mode 0600. Only `haiku_tick.py --send` reads it, and only to put it in the `x-api-key` header sent to api.anthropic.com.
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
