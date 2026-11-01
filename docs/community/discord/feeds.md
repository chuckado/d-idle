# Feeds into the server

> status: LIVE since 2026-10-31 (see below). Webhook URLs are secrets: private 0600 file, keyring or environment only, never
> in chat, issues, the repo or screenshots. If one leaks, delete the webhook in Discord and create a new one.

## Live state (2026-10-31)
- Three webhooks named `d.idle feed` exist on `#progress`, `#announcements` and `#stream`; the URLs are only in private 0600 files (never in the repo). Test posts succeeded.
- A fourth webhook named `d.idle feed` exists on `#d2d-upstream` (feeds2, 2026-11-01); private 0600 file `discord-webhook-d2d-upstream` in the same directory.
- The GitHub repo webhook (id 705756116) now delivers `release` and `pull_request` events only, to `#progress`. `workflow_run` was removed (feeds2): GitHub cannot
  filter a webhook by conclusion, so it posted every CI run. Failures come from the CI job `notify-failure` instead (section (f)). Edit the events from a shell:
  `gh api -X PATCH repos/shatteredeagle/d-idle/hooks/705756116 -f 'events[]=release' -f 'events[]=pull_request'` (never print `config.url`).

Order of value:

## (a) GitHub to `#progress`
1. Discord: `#progress` settings, Integrations, Webhooks, New Webhook, Copy Webhook URL (store it as in (b)).
2. GitHub: repo Settings, Webhooks, Add webhook. Payload URL: the Discord webhook URL with `/github` appended. Content type:
   `application/json`. Secret: empty. SSL verification on.
3. Events (choose "Let me select"): Releases; Pull requests; Workflow runs (or Check runs, not both); Pushes only if wanted (noisy,
   skip at first); Issues optional. Leave everything else off. Public repo events are public anyway.
4. Redeliver one recent delivery from the webhook page to test; check the embed, then remove the test noise if any.

## (b) Soak, drill and release summaries via the tool
`python3 tools/discord/notify.py init --channel progress` prints the steps. Store each URL one of two ways (never in chat):
- Headless box (no keyring): `install -d -m 700 ~/.local/share/d2idle`, then per channel `(umask 077; printf %s 'URL' > ~/.local/share/d2idle/discord-webhook-progress)`
  (same for `-announcements`, `-stream`). The tool refuses a file that is a symlink, not yours, not exactly mode 600, in a
  group/world-writable directory, empty, with spaces, or not of the form `https://discord.com/api/webhooks/<id>/<token>`.
  `D2IDLE_DISCORD_WEBHOOK_<CHANNEL>_FILE` names another path; `D2IDLE_DISCORD_HOME` moves the whole directory (tests use it).
- Desktop: `secret-tool store --label="d.idle Discord progress webhook" service d2idle account discord-webhook-progress`.

Lookup order: env `D2IDLE_DISCORD_WEBHOOK_<CHANNEL>`, then the file, then the keyring. Then
`python3 tools/discord/notify.py post --channel progress --title "Soak" --text "2 h, 0 deaths (verified, d2dsoak)" --dry-run`.
The tool refuses text with home paths, private hostnames or addresses, tokens, webhook URLs, emails or lab names, sends no
mentions, honours Discord rate limits (3 tries) and never raises, so release scripts may call it without guarding. It is
deliberately generic: callers build the text; keep it to facts you can tag verified or unverified.

## (c) Stream live notice (later)
When the Kick stream is configured (`docs/stream/kick-title.md`), the stream service may call `notify.py post --channel stream` once
when going live, with the public channel link only. Not wired; U until the stream exists.

## (d) No bots that read messages
No moderation or chat bots at the start: they need the message-content intent, add attack surface and a data-handling duty. Revisit
only if solo moderation fails, and then choose one with open source and minimal permissions.

## (e) Summaries with `feed.py` (live since 2026-11-01)
`python3 tools/discord/feed.py <subcommand>` builds a short plain summary from facts in our docs (or from flags) and posts it through
the same webhook loader, sanitizer and 3-try rate-limit handling as `notify.py`. **Dry run is the default**: it prints the channel and
the text (no webhook URL is ever printed); `--send` posts. Run it by hand or from an agent after an event; it is a command, not a
daemon or a timer (nothing is self-hosted on the dev box).

| subcommand | input | default channel | notes |
|---|---|---|---|
| `soak` | `--file docs/probes/postfix-soak*.md` or `--verdict/--duration/--deaths/--restarts/--xph/--kph` | `#progress` | 3 to 5 lines; REFUSED without a verdict line (`Verdict: ...` or the `d2dsoak compare` `OVERALL <verdict>` line, or `--verdict`) |
| `drill` | newest `## DRILLS<n>` section of `docs/probes/drills/README.md`, or `--pass/--fail/--blocked/--finding` | `#progress` | PASS/FAIL/BLOCKED counts from the section's verdict table, open findings |
| `upstream` | `--pr N --event opened\|merged\|reviewed\|ci-green\|changes-requested --title T --url U` | `#d2d-upstream` | `U` must be exactly `https://github.com/psi29a/d2decomped/pull/<digits>` and match `--pr` |
| `pr-status` | `--pr 'N\|title\|CI\|reviews\|state'` (repeat) or the same lines on stdin | `#progress`; `--channel d2d-upstream` for the open upstream PRs | compact code-block table |
| `release` | `--tag T --notes TEXT\|--notes-file F --url U` | `#announcements` | REFUSED unless the git tag exists; `U` must be `https://github.com/shatteredeagle/d-idle/releases/tag/<tag>` |

`--channel progress|announcements|d2d-upstream|stream` overrides the default. A channel without a webhook fails closed with
`no webhook for <channel>: store its private file ...`; `--dry-run` (the default) still works for it.

```
python3 tools/discord/feed.py soak --file docs/probes/postfix-soak1.md            # look first
python3 tools/discord/feed.py soak --file docs/probes/postfix-soak1.md --send     # then post
python3 tools/discord/feed.py drill --file docs/probes/drills/README.md
python3 tools/discord/feed.py upstream --pr 153 --event ci-green --title "Waypoint hostile window" --url https://github.com/psi29a/d2decomped/pull/153
```

Mentions are disabled (and every `@` is broken in the text). **Do not post**: machine names, home paths, IPs, emails, tokens,
webhook-shaped URLs (the sanitizer refuses them), unverified numbers presented as fact, private chat or maintainer messages,
per-run noise (one summary per soak or drill report, not per attempt), anything about the user's lab. The rule for when to run
it is [process/feed-after-events.md](../../process/feed-after-events.md).

### (f) CI failures to `#progress` (live, feeds2)
The job `notify-failure` in `.github/workflows/ci.yml` (`needs: [checks, test]`, `permissions: contents: read`) posts ONE line, `CI failed on main: <short sha> <run URL>`,
to `#progress` when a non-PR run fails; it is skipped on a green run and on pull requests. Why a job: a GitHub webhook cannot filter `workflow_run` by conclusion.
- Secret: GitHub Actions repository secret `DISCORD_PROGRESS_WEBHOOK` (the `#progress` webhook URL); it is in the env of that single step only. Nothing user-controlled
  (commit message, branch name) is interpolated into `run:`; the body is built with `jq -n --arg`, `allowed_mentions` is empty, `curl --fail`.
- Rotate: in Discord reset or recreate the `#progress` webhook, rewrite `~/.local/share/d2idle/discord-webhook-progress` (umask 077, mode 0600), then
  `gh secret set DISCORD_PROGRESS_WEBHOOK -R shatteredeagle/d-idle < ~/.local/share/d2idle/discord-webhook-progress`. Check with `gh secret list` (names only).
- Test without a failure: `gh workflow run ci -R shatteredeagle/d-idle -f feed_test=true` (a `workflow_dispatch` input); the job then posts `feed test: CI failure notices are wired`.
  The `if` uses `always()` so the dispatch test runs even though the needed jobs succeeded.

### (g) `#d2d-upstream`
The webhook file exists, so `upstream` and `pr-status --channel d2d-upstream` post there. Build `pr-status` rows from the live GET state of our open PRs
(CI, CodeRabbit review count, whether the maintainer reviewed); do not guess. If the webhook leaks: delete it in Discord and create a new one, rewrite the file.
