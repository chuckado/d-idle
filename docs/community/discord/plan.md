# d.idle Discord: plan

> status: DRAFT for the user's review. Nothing is created, posted or linked yet. Layout: [server.md](server.md); feeds: [feeds.md](feeds.md).

## Purpose and audience
A small, separate server named exactly `d.idle` (the project is styled d.idle everywhere; repo and directory names stay). Three audiences:
contributors (code, specs, evidence), followers of the stream and progress, and people who want to run the bot themselves.
It links to the d2d maintainer's community; it does not use their channels or name ([maintainer-message.md](maintainer-message.md)).
The user moderates alone at first, so the design favours few channels, read-only feeds and tight defaults.

## Launch order
1. Create the empty server by hand (name `d.idle`), enable Community, set verification and filter ([server.md](server.md)).
2. Build the layout: by hand from server.md, or `tools/discord/apply_server.py` (dry run first).
3. Paste `welcome.md`, `rules.md`, `faq.md`, `templates.md`; pin them. Set Rules Screening and Onboarding.
4. Create webhooks, store them in the keyring, test the feeds privately (server still invite-only, only the owner inside).
5. Day-one readiness checklist below, then create the invite (member cap, see below).
6. Put the invite on the README and the public status page only when happy ([public-page.md](public-page.md)).

## What the user does vs what is automated
- User only: create the server, accept the bot invite (if used), Community and Onboarding toggles, invites, role assignment, moderation.
- Automated (optional): channels, categories, roles ([apply_server.py](../../../tools/discord/apply_server.py)); GitHub feed (webhook setting in the repo); soak and release posts ([notify.py](../../../tools/discord/notify.py)).
- Never automated: reading members' messages (no moderation bots at the start), DMs, announcements about people.

## Day-one readiness checklist (public arrivals soon after setup)
- [ ] Verification level Medium (account older than 5 min) at least; High if arrivals spike. Explicit media filter: all members.
- [ ] Rules Screening on; `#rules` and `#welcome` content (with the FAQ) posted and pinned; `#announcements` and `#progress` read-only.
- [ ] Default notifications: mentions only. `@everyone` mention not granted to anyone but the owner.
- [ ] Slowmode set per channel; AutoMod keyword filter for invites, mass mentions and links from new accounts.
- [ ] Invite: one link, no expiry, member cap (start 100, raise on purpose), temporary membership OFF. It will be public, so assume
      it leaks. Rotate or revoke: Server Settings, Invites, delete the link, create a new one, update README and page (one commit).
- [ ] Moderator contact path stated in `#rules`; Developer Mode on (copy ids for reports); audit log checked after the first day.
- [ ] Raid switch known: Server Settings, Safety Setup, pause invites; raise verification to Highest.

## Risks
- Blizzard IP: no game files, keys, cracks, piracy talk, no extracted tables. Rules say so; delete and warn on sight.
- Battle.net: the tool is offline single-player; botting or cheating on realms is off topic and removed.
- Support scope: we support the d2d container path and the documented quickstart; frozen Game.exe path is best effort; no help sourcing a game copy.
- Agent disclosure: AI agents write most of the code; welcome and faq say so plainly.
- Unverified claims: members and we tag verified or unverified; do not repeat a number that has an owner (README, STATUS).
- Spam and raids: verification level, AutoMod, slowmode, invite cap, read-only feeds; webhooks are the only bots.
- Privacy: ask for logs scrubbed of paths and names; no hostnames, tokens or webhook URLs in the server (the notify tool refuses them).
- Single point of failure: one moderator; add a trusted second owner or moderator before the server grows.

## Moderation checklist (solo moderator)
- Daily (5 min): scan `#help`, `#bug-reports`, new members; check the audit log and AutoMod alerts.
- Weekly: check invite uses and the member count against the cap; review pinned templates; confirm feeds still post.
- Action ladder: delete, private warning, 1-day timeout, ban. Note the reason in the audit log reason field.
- Remove at once: game files, keys, links to pirated copies, secrets and webhook URLs (then rotate them), doxxing, hate speech.
- Bug reports: move to the template, open a GitHub issue only with the reporter's consent; keep evidence scrubbed.
- Never DM members first; never run bots that read messages; keep the owner account on 2FA.

## Proposed README line (the user decides when; not applied)
Under Contributing, after the issue/PR bullets: `Chat and help: the d.idle Discord, <INVITE URL>. Rules and FAQ are pinned there.`
Page change: [public-page.md](public-page.md).
