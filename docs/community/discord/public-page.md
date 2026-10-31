# Linking the server from the README and the public status page

> status: PROPOSAL. Nothing is linked or published until the user supplies the invite and says so. No URL is in the repo.

## README
One line (exact), under Contributing; the user adds it:
`Chat and help: the d.idle Discord, <INVITE URL>. Rules and FAQ are pinned there.`

## Public status page (generator is in the private ops repo)
The generator (`tools/publish/`, entry `run_publish.sh`) is a stub in this tree; the code (`pages.py`, `publish.py`, `sanitize.py`)
lives in the private ops repo, so this change cannot be made here. Proposed change there (small, about 15 lines plus tests):
1. Config, not hard-coded: `D2IDLE_DISCORD_URL` env var (or `discord_url` in the publisher's local config). Empty by default.
2. `sanitize.py`: the status page allows only `ALLOWED_URL_HOSTS`. Add an exact pattern like the Kick one, accepting only
   `https://discord.gg/<code>` (code 2 to 32 letters or digits; no port, no query) and still refusing the `/api/` and webhook forms.
3. `pages.py` `render_html` footer (and `render_md` last line): if `discord_url` is set AND passes the sanitizer, add
   `<p class="mute"><a href="URL" rel="noopener">Join the d.idle Discord</a></p>` before the disclaimer; otherwise render nothing and
   print one line "discord link skipped" (never fail the page).
4. Tests (`tests/test_publish.py`): no config -> no link; valid invite -> link rendered once; query-string, webhook, other host,
   http, user-info URLs -> not rendered and the page still builds.
Also add the same exact-invite exception wherever the scene page or `stats.json` could carry it: they should NOT; keep it to the footer.
