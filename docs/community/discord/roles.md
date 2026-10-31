# d.idle Discord: roles and safety settings

> status: DRAFT guideline for a small public server. Builds on [server.md](server.md) (layout, `server.json`). Nothing is created.
> Source quality: the Discord support pages could not be re-read by the fetch tool on 2026-10-31 (HTTP 403), so claims tagged
> [official*] come from the coordinator's reading of those pages and are unconfirmed here; [secondary] are SEO blogs (rough);
> [ours] is our decision. Re-check [official*] items in a browser before launch. No authoritative source gives moderator tier counts.

## 1. Principles
- Least privilege: grant a permission to a role only when a task needs it. Administrator only for `owner`, never for bots or helpers [official*, Raids 101; secondary].
- Few roles: most servers need 5 to 15, Discord allows 250 [secondary]. We use 4 (5 with the optional `mod`).
- Role order matters: a role can only manage roles below it; keep `owner` on top, then `mod`, `contributor`, `member` [official*].
- Separate three questions: who can see a channel (View Channel), who can speak (Send Messages, set per channel by overwrites), who can moderate (Manage Messages, Timeout, Kick, Ban). Roles carry moderation; channel overwrites carry seeing and speaking.
- `@everyone` is the floor; roles add; channel overwrites stack on top; where roles conflict the allowed permission wins [official*, secondary]. Test with "View Server As Role".
- Colour and hoist (show separately in the member list) only for roles people must find: `owner`, `contributor` (and `mod`). Never hoist `member`.
- No self-assignable role may carry permissions (Onboarding/reaction roles are cosmetic only). All four roles here are assigned by hand.
- Watch the audit log for permission and role changes [official*].

## 2. Role set and order (top to bottom)
1. `owner`: Administrator; the user; 2FA on. Hoisted.
2. `mod` (OPTIONAL, not created until a helper exists): see below. Hoisted, mentionable.
3. `contributor`: merged work or real review; flair, may post in `#dev`; no moderation. Hoisted.
4. `member`: trusted tier, assigned by hand to people seen behaving. [ours, user decision]: `@everyone` already gets normal chat; `member` only adds threads and external emoji, and is the raid fallback (section 5).
5. `@everyone`.

`mod`: recommend (as optional, disabled by default). Permissions: Manage Messages, Timeout Members, Kick Members, View Audit Log, View/Send as member. Not Manage Roles/Channels/Server/Webhooks, not Ban at first (owner bans), never Administrator. Reason: the plan already names a single moderator as the single point of failure; a helper needs delete-and-timeout without being able to change permissions. Create it only when a person is trusted; require 2FA for them (Discord requires 2FA for moderation actions on Community servers when "require 2FA for moderators" is on [official*]).

`muted`/`restricted` role: reject. Use Discord's built-in Timeout (Moderate Members permission): it needs no role, no per-channel overwrites to maintain (a muted role breaks every time a channel is added and can leak through a forgotten overwrite), expires on its own, is logged in the audit log with a reason, and cannot be applied to Administrators. Max timeout is 28 days [official*]. The action ladder in [plan.md](plan.md) already says 1-day timeout.

Webhooks and bots: webhooks have no role and post regardless of channel permissions; Manage Webhooks stays with `owner`. The one bot is the long-lived ops bot, with its own role `d.idle-ops` (see [Ops bot: least privilege](#ops-bot-least-privilege)), placed below `mod` and `owner`.

## 3. Permission table
Y allowed, - not granted, ro = denied in the read-only channels by overwrite. Server-level base permissions.

| permission | @everyone | member | contributor | mod (opt.) | owner |
|---|---|---|---|---|---|
| View Channel, Read History | Y | Y | Y | Y | Y |
| Send Messages, Send in Threads | Y (ro) | Y (ro) | Y (ro) | Y | Y |
| Embed Links | Y | Y | Y | Y | Y |
| Attach Files | only `#show-your-runs`, `#bug-reports` | same | same | Y | Y |
| Add Reactions | Y | Y | Y | Y | Y |
| Create Public Threads | - | Y | Y | Y | Y |
| External Emoji | - | Y | Y | Y | Y |
| Mention @everyone/@here/all roles | - | - | - | - | Y |
| Manage Messages | - | - | - | Y | Y |
| Manage Threads | - | - | Y | Y | Y |
| Timeout Members | - | - | - | Y | Y |
| Kick Members | - | - | - | Y | Y |
| Ban Members | - | - | - | - (owner bans) | Y |
| Create Invite | - | - | - | - | Y |
| Manage Roles/Channels/Server/Webhooks | - | - | - | - | Y |
| View Audit Log | - | - | - | Y | Y |
| Administrator | - | - | - | - | Y |

@everyone decisions [ours]: normal chat; embed links stay on (link previews are useful and AutoMod plus Medium verification handle link spam); attach files only where evidence is wanted (run screenshots, bug logs), because arbitrary uploads in `#general` are the cheapest way to post unwanted media and the explicit media filter is only a backstop; no mention everyone, no Create Invite (the owner controls the one invite), no Manage Webhooks. Note: role base permissions are overridden by an `@everyone` deny in the same channel, so `member` cannot post in read-only channels even with Send granted.

## 4. Server-level safety settings
- Verification level: Medium (account registered over 5 minutes) at launch for a public invite; High (also member over 10 minutes) during a raid [official*: Verification Levels; Discord recommends some verification for any public-invite server]. Re-check the level definitions.
- Explicit media content filter: scan all members [official*, Verification Levels / Community requirements].
- Rules Screening: on, text from [rules.md](rules.md); Discord recommends it over bot-based gates [official*, Rules Screening FAQ]. Needs Community enabled.
- AutoMod: block severe slurs (keyword preset), block mass mentions, block invite links, spam filter; alert, not block, on context-dependent terms [official*, Auto Moderation in Discord; Raids 101]. No alerts channel at first: use the audit log.
- 2FA requirement for moderation: turn on once a `mod` exists [official*, Community requirements].
- Invites: one link, no expiry, member cap (start 100), temporary membership off. Revoke: Server Settings, Invites, delete the link, create a new one, update README and public page. Pause invites: Server Settings, Safety Setup [official*, Raids 101].
- Onboarding: default channels only, no role questions; Discord limited it by member count, re-check the current threshold.

## 5. Raid runbook (after Raids 101 [official*])
1. Raise verification to High (Highest if needed) and slowmode to 60 s in open channels.
2. Pause invites (Safety Setup) and delete the public invite link.
3. Move chat from `@everyone` to `member`: remove Send Messages, Send in Threads, Create Public Threads from `@everyone` at server level (`member` already has them). Newcomers can read only.
4. Review the audit log and member list; ban or kick raiders, delete their messages, check no role, webhook or invite changed. Rotate webhooks if any URL leaked.
5. Restore: wait out the wave, give `@everyone` its permissions back, verification to Medium, create a new invite and update README and page; note what happened in the log.

## 6. Check after setup
- "View Server As Role" for `@everyone`, `member`, `contributor` (and `mod`): read-only channels show no input box; `#dev` and others do.
- Send test messages from a plain account (not the owner: Administrator bypasses everything): chat works in open channels, fails in read-only ones, files upload only in `#show-your-runs` and `#bug-reports`, cannot mention everyone, cannot create an invite.
- Post a test via each webhook into `#progress`; it must work although `@everyone` cannot send there.
- Flip the raid switch on a throwaway test and back; confirm `member` can still talk.
- Audit log shows the setup actions and nothing unexpected.


## Ops bot: least privilege

The ops bot (token in a private file, see [server.md](server.md)) has one role, `d.idle-ops`, placed below `owner` and `mod`. Manage Roles only reaches roles below it, so it can never touch `owner` or `mod`. Settings: "Public Bot" OFF, no privileged intents (no message-content intent: it never reads message text), never DMs members.

Permissions, only these:
- View Channels, Send Messages, Embed Links, Read Message History
- Manage Channels (create/edit channels, topics, overwrites)
- Manage Roles (create roles and set their permissions from `server.json`)
- Manage Messages (pin/unpin the welcome, rules, faq and template texts; tidy its own posts)
- Manage Webhooks, only while creating feed webhooks; remove it afterwards
- View Audit Log, optional, read only
- Manage Server: only temporarily for one `apply_server.py --settings` run, then remove

Never granted: Administrator, Kick Members, Ban Members, Moderate Members, Mention Everyone, Create Invite (unless the owner decides otherwise).

Allowed operations (no further question needed):
- create and edit channels, categories and topics from `server.json`
- pin the welcome, rules, faq and templates texts; post feed test messages
- create roles and set their permissions from `server.json`
- read the audit log (if granted)

Always ask first:
- deleting channels or roles
- kicking, banning or timing out anyone
- changing verification level or `@everyone` permissions after launch
- creating or revoking the public invite
- anything touching members or DMs

## 7. Sources (retrieved 2026-10-31; support pages returned 403 to the fetch tool, content via coordinator)
- Verification Levels, official: https://support.discord.com/hc/en-us/articles/216679607-Verification-Levels
- Rules Screening FAQ, official: https://support.discord.com/hc/en-us/articles/1500000466882-Rules-Screening-FAQ
- How to Protect Your Server from Raids 101, official: https://support.discord.com/hc/en-us/articles/10989121220631-How-to-Protect-Your-Server-from-Raids-101
- Enabling Your Community Server, official: https://support.discord.com/hc/en-us/articles/360047132851-Enabling-Your-Community-Server
- Auto Moderation in Discord, official: https://discord.com/safety/auto-moderation-in-discord
- Safety Library, official: https://discord.com/safety-library
- Game developer playbook, official: https://discord.com/blog/the-game-developer-playbook-part-one-getting-started-on-discord
- Discord server roles, secondary: https://www.mava.app/blog/discord-server-roles
- Roles and permissions guide, secondary: https://blog.communityone.io/discord-roles-permissions-guide/
- Server best practices, secondary: https://www.mava.app/blog/discord-server-best-practices
