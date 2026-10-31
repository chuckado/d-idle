## Why

User request 2026-10-31: the Discord ops bot should be long-lived so the assistant can help with the server without the token being handed over each time. no observable behaviour change: Discord tooling and community docs only, no bot code. The token needs a safe private home (headless box, no keyring) and the bot needs a least-privilege role with written limits.

## What Changes

- tools/discord/common.py: `load_bot_token()` (env, then private 0600 file, then keyring; fail-closed file checks); apply_server.py uses it; tests in tools/discord/tests/test_token.py
- docs/community/discord/server.md: long-lived ops bot setup, file creation without echo, rotate and revoke
- docs/community/discord/roles.md: section "Ops bot: least privilege" (role, permissions, allowed operations, always-ask list); plan.md aligned
- CHANGELOG and docs/INDEX.md entries. Webhook URLs stay on env/keyring.

## Impact

Docs and tooling only (skip_specs). Nothing is created or sent on Discord.
