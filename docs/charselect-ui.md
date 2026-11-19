# Character Select, Phase B: the Characters page

> status: BUILT offline (card charui, 2026-11-15). Backend and routes: docs/charselect.md (card charselect). Tags: V verified in this job (named test or fit run), S secondary, U unverified. Nothing here ran against a live bot.

User decisions 2026-11-15: the audience is the owner AND friends; delete = archive then delete with a typed name and a Restore; load = switch immediately.

## What it is

A self-contained page `/watch/characters` (also `/watch?view=characters`, which redirects there), not a cockpit slot. Files in `internal/watch/web/`: `characters.html` (shell), `characters.js` (module, pure parts exported for tests), `characters.css`, `characters.json` (every text in the d.idle voice and one sentence per refusal code). Tokens only, both themes (`?theme=` or the stored theme), no new dependency, textContent only.

Entry points (additive; old layouts load unchanged): the "..." menu item `Characters` (shell.js `setupMenu`, a link, every layout including Classic because the menu is shared), the Orders More sheet item `Characters (switch, create, delete)` (panel.js `moreItems`, `go:` to the page), and the `?view=characters` redirect at the top of shell.js.

## States

| State | What shows |
|---|---|
| loading | one "Loading the characters..." panel |
| list | header (back link, title, "Playing now: X", New hero), then Heroes: a card per character (class mark coloured by the class key, name, level and class, Expansion/Classic, Hardcore, Died, Save damaged and earlier-archive chips, last played in words, ACTIVE badge, Switch and Delete) and a dashed New hero tile; then Deleted with a Restore per archive |
| switching | a strip under the header ("Switching to Zapp... Holy is saved and the game is leaving", or "Creating ..."), polled every 2 s for up to 4 min; while the bot restarts the polls fail and the strip says it is waiting; all action buttons are off meanwhile |
| refused | the sheet stays open and shows the sentence for the route's code; the list is fetched again so the screen shows what really is |
| off (503 `off`) | "Character management needs the bot under its own supervisor", names docs/charselect.md |
| no save folder (503) | "No saves yet" with Try again |
| no answer | "No answer from the bot" with Try again |
| empty | "No saved heroes" and the New hero tile |
| caps off | `caps.why` in words under Heroes ("Switching is off: ..."), Switch and New hero disabled |

Sheets (one modal at a time, focus moves in and returns to the opener, Escape and the scrim close it, Tab stays inside, bottom sheet on a phone):

- Switch: "Leave the game now and start Zapp? The current hero is saved."
- Delete: what is archived and where (`characters-archive/<Name>/`), never cleaned up, restorable; the type-the-name field turns the button on only for the exact name (case-sensitive, as the server). The active hero gets a sheet that only gives the reason and a Close. The result note names the archive id from the reply.
- Restore: the archive's level, class and age; says it is not started.
- Create: name with a live hint from the API rule, class chips from the API's class table, "Hardcore is not available yet." (from `rules.hardcore`), the button reads "Create Newb" once valid. Off with the reason when `caps.create` is false or the rules did not arrive.

Read-back: after every accepted press the list is fetched again. A delete that is still listed, or a restore that is not listed, shows a warning instead of the success sentence (V: node test).

## API additions (additive)

`GET /api/characters/rules` (`watch.characters.rules/1`, viewer, no desk needed): the class table (`id`, `name`) from `internal/charsel/classes.json`, the name rule (`min`, `max`, `separators`, `foldCase`) and `hardcore` (false until the T4 acceptance). Golden `internal/watch/testdata/characters/rules.json`. `charsel.ValidName` now reads the same constants. The list route and its goldens are unchanged.

## Tests

- `go test ./internal/watch/ -run TestCharacters` : routes, the page assets, every refusal code of `internal/charsel` has a text, entry points present, and the node DOM test of `characters.js` in a stub DOM (list, switch sheet and strip, refused press is no success, typed-name delete, read-back catches a delete that did not happen, restore, create rules and class chips from the table, Escape and focus, caps.why, off, no-save and empty states). The node test skips without node.
- `tools/uishots.sh --fit --only 'chars*'` (worker host): fixtures `chars-1 -3 -8 -long -deleted -switching -measured -empty -off -nosave`, sheets `chars-new -switch-sheet -delete -delete-light -delete-active -restore`, the two entry points, light variants, at the eight sizes plus 412x780 and 360x640. fit.js `charsPage` judges: clipped text, controls under 44 px, covered controls, boxes outside the window, the list under the header, an open sheet that scrolls, dead space inside the list (largest empty square above the end of the content: 300 px, 240 px on a phone; what follows the last card is the end of a list). Must-fail fixtures `charsdefect-clip -wide -under -small -hole -sheet -covered` put one defect back through `defectCSS` and must fail with their verdict.

## Not verified

U: against a live supervised bot (docs/charselect.md checklist 1 to 7 still apply); U: how long a real switch takes, so the 4 min strip limit is a guess; U: contrast was taken from the token pairs (dark and light), not measured per pixel; U: screen readers (roles and labels are set, nothing was listened to). Earlier archives of a character that still exists are counted on its card but cannot be restored from the page (the name is taken); they stay in `characters-archive/`.
