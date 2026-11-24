# Vault page load time, 2026-11-23

> status: findings + fixes (branch agent/vaultperf). Sandbox read-only: file sizes and logs only; no credentials, no POST. V = measured here (command named), S = read from code, U = not measured.

Sandbox vault (hero Doofer): `log.jsonl` 2.39 MB / 7743 events (4521 adds, 801 instant sales), `archive.jsonl` 1.87 MB (not loaded per request), base capacity 400, max 2400 (`data/vault/capacity.json`), so a live vault is 400..2400 records (V: file sizes via `docker exec ls -l`; the live count itself U, "vault full" lines in the log say it sits at capacity).

| # | Cause | Evidence | Number | Tag |
|---|---|---|---|---|
| 1 | Every Vault file is sent raw with `Cache-Control: no-store`: no gzip, no ETag, so each visit re-downloads everything | `pactHdr` / `asset` in http_view.go; 12 files | about 250 KB raw per visit (64 KB gzip, what the budget test counts) | V (wc -c, `TestAssetBudget`), a proxy may compress (U) |
| 2 | Serial start-up waterfall: html, then vault.js, then vault-logic.js (module import), then `/api/config`, then `vault.json`, then `overlay.js` (dynamic import), then the first `/api/vault` page, then `inventory-tip.js` | `boot()` in vault.js (awaits in a row) | 5 to 6 dependent round trips before the first row can be asked for; at 150 ms RTT about 0.8 s of pure waiting | S (code), not timed in a browser (U) |
| 3 | List page weight | `BenchmarkVaultPage` | 200 rows = 52 KB plain JSON (about 260 B/row), 7.7 KB gzip | V |
| 4 | Server work per request (poll `limit=1` every 6 s and each page) | `BenchmarkVaultPoll`, `BenchmarkVaultPage` | poll 0.29 ms at 400 records, 1.04 ms at 2400; page 0.8 / 1.7 ms. Not a cause; the real Keeper adds three `Dump()` copies and game-side gates per call (U, not measured) | V (bench), U (real Keeper) |
| 5 | Client render | `vault.js` windows the grid (`windowRange`), loads 60 rows then 200 per page | only rows near the view are in the DOM | S |

Fixes (this branch): (1) `serveAsset` in compress.go: ETag from the bytes, `no-cache` (a 304 with no body when unchanged), a gzip copy made once; the Vault, the shared viewer assets and the inventory files use it; HTML pages stay `no-store` (the update banner edits them). `/api/vault` is gzipped when the client accepts it. (2) `boot()` starts config, texts, overlay module and the first list page together (the page is kept by its URL and used once only if the page asks for exactly that URL), and `vault.html` modulepreloads vault-logic.js. Guards: `TestVaultWireWeight` (validators, gzip, 304, 200-row ceilings 60 KB plain / 8 KB gzip, poll under 4 KB), `TestVaultPollCost` (poll under 10 ms at 2400 records), `TestAssetBudget` (Vault 64,178 gzip bytes, ceiling 69,000).

Expected gain: repeat visit moves about 250 KB to a few 304 headers; first visit about 64 KB instead of 250 KB if no proxy compressed; two to three round trips fewer before the list is asked for. Not measured in a browser (U): a fit/uishots timing run on the worker host is the check.

Left: the real Keeper's per-poll `Dump()` copies (cache by store version), `inventory-tip.js` still loads after the grid, `/api/vault` view is rebuilt per request.
