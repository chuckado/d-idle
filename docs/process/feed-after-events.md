# Feed after events

> status: current rule (card feeds, 2026-11-01). The helper is `tools/discord/feed.py`; details in docs/community/discord/feeds.md section (e).

After one of these events, the session that caused it (coordinator or agent) runs the helper once: dry run first, read the text, then `--send`.

| event | command |
|---|---|
| a soak report is merged or written (`docs/probes/postfix-soak*.md`) | `python3 tools/discord/feed.py soak --file <report>` |
| a drill report is merged (newest section of `docs/probes/drills/README.md`) | `python3 tools/discord/feed.py drill --file docs/probes/drills/README.md` |
| an upstream PR is opened, merged, reviewed, goes CI-green or gets changes requested | `python3 tools/discord/feed.py upstream --pr N --event E --title T --url U` |
| a dispatcher tick with something new (merged, launched, blocked, upstream change or waiting item; an idle tick is refused; the whole report goes in one post, capped at 3800 chars and cut at a line with a final `(cut: N chars)` line, never silently) | `python3 tools/discord/feed.py tick --running A,B --merged x --launched y --upstream T --blocked T --waiting T` |
| a release exists (git tag plus CHANGELOG notes) | `python3 tools/discord/feed.py release --tag T --notes-file F --url U` |

Rules: a command, never a daemon or timer; no secrets written anywhere (the helper reads the private webhook files itself and never prints them); a soak without a verdict is not posted; `upstream` fails closed until `#d2d-upstream` has a webhook file (needs the user go); one post per event, never a burst. Do not post what the sanitizer would refuse, and never paste a webhook URL.
