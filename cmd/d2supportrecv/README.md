# d2supportrecv: the support bundle receiver

Receives the zip the watch screen's "Send support bundle" button builds (docs/support-bundle.md has the threat model
and the client side). One small stdlib-only Go program; it stores zips and never unpacks, serves or lists them.

## What it enforces
- `POST /support/upload` only, `Authorization: Bearer <upload token>`, `Content-Type: application/zip`, at most 20 MB,
  a readable zip with at most 64 plain relative names (no `..`, absolute, backslash, directory or symlink entries) and at
  most 100 MB unpacked.
- Upload tokens live in a file as SHA-256 lines (`d2supportrecv -hash`); a token is never stored in clear. Delete a line
  to revoke it (the file is re-read when it changes; a missing file refuses everybody). The token is a separate, write-only
  secret: it cannot read anything, and it is not the watch token.
- Limits per hour (all in memory, none logged): `-token-rate` uploads per token (20), `-ip-rate` requests per client (10),
  `-fail-rate` failed logins per client (5). Behind the tunnel use `-trust-proxy` (client address from `CF-Connecting-IP`).
- Inbox `-dir`: files `<CODE>.zip` (8 characters from a 32-symbol alphabet, random, mode 0600), outside any web root;
  `-quota-mb` (2048) and `-max-files` (500) purge the oldest first. The answer is `{"code":"ABCD2345"}`.
- It logs only its start line: no address, token, code or size.

## Tower deploy steps (owner; nothing here is deployed by the repo)
1. Make a token and its hash line (the token never goes into git):
   `openssl rand -base64 33 | tee /root/support-token.txt | d2supportrecv -hash >> /mnt/user/appdata/didle-support/tokens`
   (build the binary first, or run the image with `--entrypoint /d2supportrecv ... -hash`). Put the token text into the
   friends pack `.env` as `D2BOT_SUPPORT_TOKEN=<token>` and the address as
   `D2BOT_SUPPORT_URL=https://<updates host>/support/upload` (docs/support-bundle.md).
2. `mkdir -p /mnt/user/appdata/didle-support/inbox`, `chown 99:100` it, and make `tokens` readable by 99:100 (mode 0640).
3. Build the image (Dockerfile here) and create the container through an Unraid template so it shows in the UI:
   name `didle-support-recv`, network `didle-updates` (the nginx one), no published port, read-only rootfs, user 99:100,
   cap-drop ALL, mounts `inbox` rw at `/data/inbox` and `tokens` ro at `/data/tokens`.
4. Add `nginx-support.conf` to the nginx server block of the updates site and reload nginx.
5. Check: `curl -s https://<updates host>/support/health` prints `ok`; a POST without a token answers 401.
6. Revoke: delete the token's line from `tokens` (no restart). Rotate: add a new line, ship the new pack, then delete the old.
7. Owner fetch: `tools/support-fetch.sh <CODE>` in the ops repo downloads and unpacks one bundle.
