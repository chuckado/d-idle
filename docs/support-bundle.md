# Support bundle: the "Send support bundle" button

> status: new (card `support`, task 542). Code: `internal/support` (build, redact, scan, send), `internal/watch/supportapi.go` and `web/supportui.js` (routes and sheet), `cmd/d2bot/supportwire.go` (what is collected), `cmd/d2supportrecv` (the receiver, its README has the deploy steps). Nothing here is deployed; the live checks at the end are open.

A friend (or the owner) opens the "..." menu, presses "Send support bundle...", sees exactly what would be sent, and presses Send. The bundle goes to the updates site and the sheet shows a short support code to give the owner. A second choice, "Save to file instead", downloads the same zip (offline fallback). There is also the older `./didle report` (a local tarball the friend sends by hand, docs/friend-guide.md); the bundle replaces the hand-sending, it does not replace `didle report` when the bot is down. `cmd/d2doctor` has a bundle of its own (config tail, doctor report); this one is built inside the running bot so it can include the live status, and it takes the same approach (drop secret-named keys, tail the log, never read token files).

## Threat model (read first)

What we protect: the friend's identity and machine (home path, user and host name, addresses, e-mail), every secret on the box (Diablo II CD keys and `.key` files, the watch token, the pack's upload token, update and API keys, cookies), the friend's character save, and the owner's infrastructure (host names, private paths). What we defend against:

1. **A secret or personal string in a log or status file ends up on the server.** Answer: an allowlist of collected files, redaction of every text file, a second independent scan that fails closed (below), and a preview the person confirms.
2. **The button becomes a silent exfiltration channel.** Answer: no upload without the press of Send on the preview of that very bundle (the server keeps the zip it showed; Send takes its id, not new content). No timer, no auto-send, no upload on a crash.
3. **The upload token leaks** (it sits in the pack `.env` and in the container environment). Answer: it is a separate, write-only token, hashed on the server, revocable by deleting one line, rate-limited per token and per address; it cannot read or list anything and is not the watch token. The environment is never put into the bundle (no env dump; the token is only ever read to send the request).
4. **A hostile client or scanner hits the endpoint** (zip bombs, path traversal, disk fill, brute force). Answer: size cap 20 MB, zip magic and readable central directory, at most 64 plain relative names, 100 MB unpacked cap, nothing is unpacked by the receiver, per-IP and per-token and failed-login limits, quota with oldest-first purge, random file names, mode 0600, inbox outside any web root.
5. **The server is read by someone else.** Answer: the bundle holds no secret by construction; codes are random but are only a locator; nothing under `/support/` is served; `access_log off` for the path; the receiver logs nothing about senders.
6. **A fix breaks redaction and nobody notices.** Answer: tests seed one secret of every class (`internal/support/support_test.go`), the collector test seeds a key file, a token file, a home path and an address (`cmd/d2bot/supportwire_test.go`).

Out of scope: a malicious bot build (it could send anything), a friend who ticks "include my save" and sends it to the wrong person, traffic analysis (the sender address is visible to the receiver; only the owner reads it, and it is not logged).

## Contents (allowlist, not blocklist)

`support.Allowed` in `internal/support/bundle.go` is the only list of names; a collector offering any other name is dropped and the preview says so. Today:

| File | What | Source |
|---|---|---|
| `MANIFEST.txt` | time, the two option states, what was done to the text | made by `Build` |
| `version.json` | bot version, image tag (`D2BOT_IMAGE_TAG`, U: only if the image sets it), instance id | `buildinfo.Current()` |
| `system.json` | OS, distro name, CPU count, Go version, "in a container" | runtime, `/etc/os-release` |
| `knobs.json` | the tuner's values (numbers only, no names that say token, key, secret) | `tuner.Snapshot().Values` |
| `glance.json`, `status.json`, `readiness.json`, `danger.json` | the viewer's own `/api/glance`, `/api/hero`, `/api/readiness`, `/api/danger` | in-process GET on the viewer mux |
| `log.txt` | the end of the instance log (`instances/<id>/logs/d2bot.log`), the last 6 hours when lines carry a time, at most 1 MB after redaction | file tail |
| `errors.txt` | the lines of that log with ERROR, WARN or PANIC | filtered from the same tail |
| `decisions.json`, `deaths.json` | the last 200 valid lines of `decisions.jsonl` and `deaths.jsonl` of the state directory | file tail |
| `screenshot.png` | optional, one frame of the d2d view tap; only with the checkbox, default off | tap `/frame` (U: the tap's frame format, JPEG or PNG both accepted) |
| `save/character1.d2s` ... | optional, only with "Include my character save file", default off; the sheet says plainly: your character, items and name; the file name never carries the character name | `Instance.SaveDir` (U: a placeholder for fresh setups, a real install keeps saves in the Wine prefix) |

Open slots, named in the brief and not collected yet because their source files are not named in the repo (U): the minute, area and ledger summaries, the d2d trace summary and the stream stats. To add one: add its name to `Allowed`, add one `add(...)` line in `supportwire.go`, add it to this table.

Never collected: `.key` files and CD keys, anything under the tokens directory, passphrases, the watch token, the upload token, cookies, Battle.net data (the bot has none), the environment, other characters' saves, `.d2s` files without the checkbox.

## Redaction and the fail-closed scan

`Redact` (`internal/support/redact.go`) replaces, in every text file: private-key blocks, `Authorization`/Bearer values, `name=value` and `"name": "value"` for names that say token, secret, password, passphrase, cookie, credential, auth, `api_key`, `cd_key`, `private_key`, `key`; URL query tokens and `/p/<token>/` path tokens; UUIDs; CD-key shaped groups with a digit; `*.key` and `*.d2s` file names; e-mail addresses; home paths on Linux, macOS and Windows; IPv4 and IPv6 addresses; host names ending in lan, local, internal, home, ts.net and any name containing the owner's domain word; hex runs of 32 or more characters (a 40-hex build id stays); long mixed-case tokens; and the literal host name, user name and data directory of this machine (`<local>`). JSON files first lose every key whose name says token, secret, password, passphrase, cookie or key, then go through the same text pass. It over-redacts on purpose (a dotted quad that is a version number becomes `<ip>`).

`Scan` runs the same rule table plus the machine literals over the redacted text and finds what is left; a character-save signature (the `0x55 0xAA 0x55 0xAA` header) in any file outside `save/` also counts. Any finding sets `Findings`, the zip is not even built, the sheet shows each finding (file, line, class, the line with the match blanked and everything else re-redacted: never the secret), and `/api/support/send` answers 409. There is no override. The scan is a safety net, not an independent oracle: it shares the rules with `Redact`, so a new secret shape is added to the table once and covers both.

Not covered (U): a secret in a shape no rule knows (say a password in free text without a name before it) passes both. That is why the preview shows the log excerpt and why the sheet exists. `internal/support` does not import `cmd/leakcheck` (a `main` package); `leakcheck -mode fail` still scans the repo for the same classes of text.

## The flow

1. `POST /api/support/bundle {includeScreenshot, includeSave}` (answer mux: token, origin list, JSON only, small body): collects, builds, scans, remembers the bundle (the last two, 15 minutes, in memory) and answers a preview: `id`, `files` (name, size), `logPreview` (the last 40 lines as they will be sent), `findings`, `dropped`, `zipSize`, `canSend`, `canSave`, `configured`.
2. `POST /api/support/send {id}` uploads exactly that zip and answers `{"code":"ABCD2345"}`. Errors are words ("too many uploads just now", "the upload key was refused"), never the URL or token. Unknown or expired id: 404; blocked bundle: 409.
3. `GET /api/support/download?id=` (viewer mux, behind the same viewer token) is "Save to file instead"; `GET /api/support` says whether upload is set up. Without an endpoint the sheet offers only Save and says "Uploading is not set up on this install".

Client settings (environment; there is no tuner knob, so nothing for `docs/knobs.md`): `D2BOT_SUPPORT_URL` (https only, or http on loopback) and `D2BOT_SUPPORT_TOKEN`. The friends pack passes both through `compose.yml`; `install.sh` writes them into `.env` from the files `SUPPORT_URL` and `SUPPORT_TOKEN` next to the installer when the owner puts them in the pack (not built by `build.sh`, never committed; the pack's own secret scan does not see them). Redirects are never followed (the token would go along).

## Build flexibility

Nothing here is class or build specific; the collector reads files and the viewer's own routes. No skill ids or class checks.

## Owner side

Deploy and revoke: `cmd/d2supportrecv/README.md` (token creation, container, nginx snippet `nginx-support.conf`). Fetch: `tools/support-fetch.sh <CODE>` in the private ops repo downloads `<CODE>.zip` from the inbox, checks the entry names again and unpacks to a fresh directory.

## Live checklist (open; needs the deploy, the owner does it)

- [ ] Deploy the receiver and the nginx snippet (README steps); `curl -s https://<updates host>/support/health` prints `ok`; a POST without a token answers 401; a GET of `/support/upload` answers 405 from nginx or the receiver.
- [ ] Generate a token, put URL and token in a test pack, install it; the "..." menu shows "Send support bundle...".
- [ ] The preview lists the files above, the log excerpt has no home path or address, screenshot and save are unticked.
- [ ] Send: a code appears; `tools/support-fetch.sh <CODE>` unpacks it; grep the unpacked files for the friend's user name, host name, an address and the watch token: no hit.
- [ ] Revoke the token (delete its line): the next Send says "the upload key was refused" and Save still works.
- [ ] Rate limit: send 21 times in an hour (or lower `-token-rate` for the test): the answer is "too many uploads".
- [ ] The browser fit run on the worker host: `tools/uishots.sh --fit --only 'support-*'` (sheet, preview, blocked, no-upload pages at all sizes and the phone sizes) is clean. Not run in this branch (no browser).
- [ ] Phone: the sheet opens from the "..." menu, scrolls inside itself, Send and Save are reachable with the keyboard up.
