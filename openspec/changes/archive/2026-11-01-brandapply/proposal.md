## Why

The cursor brand (card `brandfinal`) is decided but not applied anywhere. This card prepares the in-repo application so it is ready the moment the user gives the go. THIS BRANCH IS NOT MERGED: it awaits the user's go (and `brandfinal` merged first; this branch is built on its tip).

## What Changes

- README.md: the header image (`assets/branding/final/header-1280x640.png`, alt text `d.idle`) at the top, and the Discord link repeated under the title; all existing text is intact.
- cmd/d2dview/web/: `icon.svg`, `favicon.ico`, `apple-touch-icon.png` (copies of the final assets) test `cmd/d2dview/brand_test.go`. The head links for `index.html` and `stream.html` are NOT edited in place: card `mapitem` (active) owns those two pages, so they ship as `openspec/changes/brandapply/head-links.patch` (`git apply` it after mapitem merges; the test skips the link check until then).
- cmd/d2bot: the banner (`banner.go`, `banner_test.go`, one line in `main.go`) printed ONLY at the head of the usage text (`-h` or a flag error, on stderr; plain ASCII unless the output is a UTF-8 terminal). It is NOT printed on `--version` (packaging/release_test.sh and tools/soak/soak24.sh parse it) and never on the log (soak and log scripts parse it).

## Relation to other cards

Built on `brandfinal` (same assets). Overlaps nothing that is active: it owns README.md, cmd/d2bot/banner*.go and main.go, the three new icon files in cmd/d2dview/web/ and brand_test.go; the two html pages stay with mapitem. main.go is a one-line edit; if another card touches cmd/d2bot/main.go the merge is a trivial re-apply.

## Follow-ups (not in this repo or not here)

- The status page and the pages repo live in the private ops repo: favicon links and header there need the same files (icon.svg, favicon.ico, apple-touch-icon.png) and the user's go.
- web/ (the D2 Idle app pages, `web/embed.go` lists the embedded files): favicon links need the embed list extended; not done here because those pages are the local dashboard, not a public surface.
- GitHub social preview, avatar, Discord icon and banner: manual uploads by the user.
