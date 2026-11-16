# Friends game-data bundle (2026-11-15)

What: one private zip with the nine data archives the d.idle server reads, so a friend does not have to copy them out of an install by hand. It is hosted on the updates site behind an unguessable path, not linked, not listed, `X-Robots-Tag: noindex, nofollow`. The friends pack, feed, update list and release notes contain no game files and no link to it; the URL is kept only in the private ops notes (ask the maintainer). Friends still bring their own legal copy and keys for the real game.

## Packaging decision
One zip, `didle-d2-data-1.14d.zip` (1466501700 bytes, stored not deflated: MPQ data does not compress), unpacking to `Diablo II/` plus `README.txt`. Why: the installer (`tools/friendpack/install.sh`) wants a folder with `d2data`, `d2exp`, `d2char`, `Patch_D2` and, for the container mounts, `d2sfx`, `d2speech`, `d2xtalk`, `d2music`, `d2xmusic`; given the unzip root it finds the single `Diablo II` folder one level down (V, run below). So the friend runs unzip then `./install.sh --game-dir <unzip dir>`. Nothing else is inside: no exe or DLL, no key file, no saves, no logs, no videos.

Members (names only): README.txt, Diablo II/{d2data,d2exp,d2char,Patch_D2,d2sfx,d2speech,d2xtalk,d2music,d2xmusic}.mpq.
sha256 of the zip: `8bd010e89389ee7080f1baefc766f130b342c9ae9d30c2ad701857a702921f8f` (a `.sha256` file and the README sit beside it).

## How it was built (V)
Source: the maintainer's read-only install backup zip. Only the nine named members were streamed out with `unzip -p` (sizes equal the zip's listing) into a scratch dir on the NAS builder, packed with Python `zipfile`, scratch deleted afterwards. No key file or save file was read.

## Verification (V)
- Installer check logic (install.sh lines for the four required files, the one-level-down search, and the nine-archive mount loop) run against the unzipped bundle: root says the four missing, the search finds `Diablo II`, all nine mounts found.
- Through the public URL: HTTP 200, content-length equals the file size, `application/zip`, `x-robots-tag: noindex, nofollow`; the token directory and `/p/` answer 404 (no listing). One full download sha256 equals the value above. The existing 0.1.3 pack still answers 200.
- nginx: added one `location ^~ /p/` block (cache, nosniff, noindex) before the generic route; `nginx -t` ok, reloaded; backup `nginx.conf.bak-20261115-205501` on the host.

## Rollback
Remove the bundle directory under `www/p/` of the update appdata on the Tower (the maintainer's note names the exact path). To undo the nginx block: `cat nginx.conf.bak-20261115-205501 > nginx.conf; docker exec didle-updates nginx -t && docker exec didle-updates nginx -s reload`. Feed, release list, signed sets and the pack were not touched.
