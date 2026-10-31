# Feeds into the server

> status: LIVE since 2026-10-31 (see below). Webhook URLs are secrets: private 0600 file, keyring or environment only, never
> in chat, issues, the repo or screenshots. If one leaks, delete the webhook in Discord and create a new one.

## Live state (2026-10-31)
- Three webhooks named `d.idle feed` exist on `#progress`, `#announcements` and `#stream`; the URLs are only in private 0600 files (never in the repo). Test posts succeeded.
- The GitHub repo webhook (id 705756116) delivers `release`, `pull_request` and `workflow_run` events to `#progress` (ping delivered, 204). The
  user asked for workflow runs. `workflow_run` posts on every CI run: if `#progress` gets noisy, remove the event or keep only failures. Edit the events in
  GitHub (repo Settings, Webhooks), or from a shell: `gh api -X PATCH repos/shatteredeagle/d-idle/hooks/705756116 -f events[]=release -f events[]=pull_request`.

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
