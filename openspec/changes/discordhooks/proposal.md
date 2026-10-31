## Why

The Discord feed webhooks now exist and their URLs sit in private 0600 files on the headless box (no keyring). The notifier only knew env and keyring, so it could not use them. no observable behaviour change for other tools: Discord tooling and community docs only.

## What Changes

- tools/discord/common.py: `load_webhook_url(channel)` with the same precedence and fail-closed file rules as the bot token (env, then file, then keyring); the value must be exactly a Discord webhook URL; channel names are validated; `D2IDLE_DISCORD_HOME` overrides the data directory (the bot token default path honours it too).
- tools/discord/notify.py uses the loader; tests set `D2IDLE_DISCORD_HOME` to a temp dir so no test reads the real home files.
- docs/community/discord/feeds.md and server.md: private-file recipe for headless boxes next to the keyring route.

## Impact

Docs and tooling only (skip_specs). A feed test message is posted to three owner-only channels after the code is committed.
