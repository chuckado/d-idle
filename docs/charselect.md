# Character Select: list, create, switch, delete (archive first), restore

> status: Phase A BUILT offline (card charselect, 2026-11-15): backend and routes with fakes and temp dirs; no UI, nothing run live. Phase B (the page) follows. Tags: V verified in this job (named test or file), S secondary, U unverified.

User request 2026-11-15: a Character Select screen that loads, creates and deletes characters. Decisions (popups): the audience is the owner AND friends (guard rails everywhere); delete = archive, then delete, with the typed name and a restore from the list; load = switch immediately (leave the game like Reset and start the chosen character).

## Feasibility: what d2d can and cannot do

| Need | d2d (private fork) | How we do it |
|---|---|---|
| list | no verb (V: the typed verbs in docs/d2dctl.md are Ping Info State Screenshot Quit Save Click RClick Key Text Hover Wheel Cmd*; none lists characters) | read the save folder: `<Name>.d2s` header fields only (name, class, level, status flags, last played). V: `TestListReadsHeadersOnTwoClasses` |
| create | start flags only: `--start-screen ingame --start-class N --start-name X` (V: internal/d2dinput/lifecycle.go Args; N is the d2d UI order); `--start-hardcore`, `--start-classic` exist | a new d2d process with those flags; the save file appears at the first save (U: not re-checked that d2d writes at Save and Exit for a created character, the saved-exp reader relies on it for loaded ones) |
| load | `--start-screen charselect --start-name X` (V: same file) | a new d2d process with `--d2d-class -1 --char X` |
| delete | no verb | file operation on the save folder, archive first |
| leave the old game | `Save` and `Quit` verbs (SaveExit in internal/d2denv/proc.go) | the Manager's `Leave` hook; NOT `Cycler.LeaveGame`, which re-enters the old character |

Conclusion: no fork verb is required. A fork verb `chars` (list names with class and level, and say whether a name is free) would only replace the header read and would make the name rule the game's own; it is optional hardening, not built, private fork only, never upstream (card text below). Where d2d keeps the folder is d2d's choice under its XDG data dir (U: the exact subfolder; `saveFolder` in cmd/d2bot/charwire.go searches five levels under `<instance>/d2d` and `$XDG_DATA_HOME`, preferring the folder with the active hero's file). The Game.exe backend is not wired (its folder is in the Wine prefix and create needs the OOG tooling, cmd/oog); the Manager works the same there once a backend supplies `SaveDir` and `Leave`.

Follow-up card (not created): "d2d `chars` verb": list of {name, class, level, flags, free names check}, verb gated like the other private verbs, contract in docs/d2dctl.md first. Only if the header read proves fragile live.

## What is per character today (V: read from cmd/d2bot)

The bot keeps state per INSTANCE, not per character: `goals.json`, `deaths.json`, `routes.json`, `ledger.json` (survival) next to the instance, and in its `state/` dir `tuning.json` (dashboard settings), `merc.json`, `playerpause.json` (pause), `orderloop.json` (orders, queue), `pacts-state.json`, `hardcore.json`, `away-spool.jsonl`, the journey file. Profile level (shared by every hero of the profile, never parked): `onboarding.json` (keyed by hero already), `visit.json`, `currency`, `ladder` (Hall), `sessions`, `layout-*`, `proposals-*`, `bestiary-*`, `heapguard`, `playerreset.json`. Reset (docs/reset.md) keeps all of it because it rejoins the same hero.

So a switch must swap the per-character files. The list is data: `internal/charsel/perchar.json` (U: the journey file name is a glob guess `state/journey*`; a file not listed stays shared, so the miss shows as a new hero inheriting a journey, never as data loss). Parked state lives in `<instance>/characters/<Name>/`.

## Design

States of the screen (Phase B reads `GET /api/characters`): list (rows with class, level, last played, a badge for the active one), a row's actions (Switch, Delete), New hero (name, class), Deleted (restorable archives), Switching (strip until the new hero is in game).

### API (additive, `watch.characters/1`; goldens in internal/watch/testdata/characters)
- `GET /api/characters` (viewer): `{schema, active, caps{create,switch,delete,restore,why}, switch?{from,to,creating,since}, characters[{name,class,className,level,hardcore,died?,expansion,lastPlayed?,active,damaged?,archived[]}], deleted[{id,name,class,level,archivedAt}]}`. `archived` are earlier archives of a character that exists now, `deleted` are archives whose character is gone. `caps.why` says in words why create and switch are off (measured run, no supervisor, backend cannot leave).
- `POST /api/characters {name, class, hardcore?}` create; `POST /api/characters/switch {name}`; `POST /api/characters/delete {name, confirm}`; `POST /api/characters/restore {archive}`. Answer-listener gates as pause and reset (origin, token, JSON, unknown or foreign fields 400). Replies `watch.characters.result/1`: `{action, name, archive?, switch?}`. Refusals are `{ok:false, code, error}`: 400 `bad_name bad_class confirm_mismatch bad_archive body`; 404 `not_found`; 409 `measured_run switch_in_progress no_supervisor already_active active damaged corrupt dead_hardcore exists hardcore_unbuilt save_failed`; 501 `unsupported`; 503 `off no_save_dir`.
- Code: `internal/charsel` (Store, Manager, SwapState, Intent), `internal/watch/charactersapi.go` (desk interface, routes), `cmd/d2bot/charwire.go` (d2d desk, supervisor side). The hook is `ladder.characters` (the `Deps.Characters` of the brief; both backends or a fake implement `watch.CharactersDesk`).

### Rules
- Names (S, D2 rule): 2 to 15 letters, at most one hyphen or underscore, never first or last; no duplicates, compared case-insensitively (a friend's PC folds case). Classes come from `internal/charsel/classes.json` (d2s byte, d2d index); an unknown class is an error naming it, never a Sorceress. V: `TestNameRules`, `TestCreateValidation`, `TestListReadsHeadersOnTwoClasses` (Paladin and Sorceress; the d2d indexes 2 and 4 are asserted).
- Hardcore: the flag is listed from the header; create with `hardcore:true` is 409 `hardcore_unbuilt` (docs/decisions.md Current truth: hardcore stays unbuilt until the T4 acceptance); switching to a dead hardcore hero is `dead_hardcore`.
- Never delete the active character (409 `active`: switch first). Delete needs `confirm` equal to the exact name (case-sensitive, 400 otherwise).
- Every changing verb is refused while `D2BOT_RUN` is set (`measured_run`) and while a switch is under way (`switch_in_progress`).
- Delete = archive, verify, remove. The archive is `<instance>/characters-archive/<Name>/<UTC stamp>/` holding every `<Name>.d2s .key .map .ma0-9`, a copy of the parked state, and `meta.json`; each copy is read back byte for byte before anything is removed (V: `TestArchiveDeleteRestoreRoundTrip`). Archive id `<Name>@<stamp>`; it becomes a path only through the listing (a path-like id is 400). Archives are never pruned by the bot.
- Restore copies the files and the state back; 409 `exists` if any character of that name exists (the archive stays); a failed copy removes what it copied.
- Switch refuses a save that fails the d2s check (`saves.Verify`: signature, checksum, sections): `corrupt`. The list still shows a damaged file (`damaged` says why) so it can be archived.

### Create
The new hero is a level 1 character of the chosen class with default plan, ledger and tuning: its state dir is empty, the bot makes defaults, and the supervisor passes `--goals` again (the same campaign the first hero got) because the parked-goals file does not exist. The save appears at the first save; until then the list does not show it (U), the `switch` field does. After the save appears, the supervisor changes the next start from create to load, so a crash restart can never create over a saved hero (V: `TestSwitcherCreateStartsFreshWithGoalsThenLoads`).

### Switch, step by step (`POST switch` or `create`)
1. Gates above; `no_supervisor` unless the bot runs as a child of its own supervisor (`D2BOT_SUPERVISE=1` or `--supervise`; the d2d container passes no flag today, set the env in the pack; under d2fleet the fleet restarts d2bot itself and cannot change `--char`, so switch is off there, U to extend).
2. Write `<instance>/charswitch.json` (`watch.charswitch/1`: from, to, create, classIndex).
3. `Begin` (Runner.BeginHardReset, so the runner does not rejoin) then `Leave` = Save and Exit of the owned d2d. If that fails: intent removed, 409 `save_failed`, nothing changed.
4. After 750 ms the process exits with code 6 (`exitSwitch`). The supervisor reads the intent, `SwapState` parks the old hero's per-character files and brings the new one's back (all or nothing: on a conflict everything is put back and the old hero restarts), rewrites `--char` and `--d2d-class` (-1 loads), clears the intent and starts the child at once; exit 6 is no crash, no backoff, no restart count (V: `TestSuperviseSwitchIsNotACrash`, `TestSwitcherLoadsAnotherCharacterAndParksState`).
5. The old hero's files stay on disk (save plus parked state); orders, pause, ledger, tuning and journey come back when you switch to it again.

### Friends' machines
The d2d container keeps data under `/data` (`XDG_DATA_HOME=/data/share`); the save folder is found under it, the archive and parked state sit in the instance dir on the same volume, so they survive `docker stop` with the character. Pack follow-up (Phase C, deploy owner): export `D2BOT_SUPERVISE=1` in `entrypoint.sh` and let the entrypoint prefer `charswitch`-written `--char` over its env on a container restart (today a container restart starts the env's name). The installer pack paths need the same check (U).

### Failure modes
Save missing or damaged: 404 or `damaged`/`corrupt`, nothing changes. Name exists on create or restore: 409 `exists`. Delete of the active hero: `active`. Measured run: `measured_run` for all four verbs. Two presses: `switch_in_progress`. d2d does not answer the save: `save_failed`, the hero keeps playing (the runner state is not reset: U if BeginHardReset has already marked the runner; see checklist 5). The bot dies between the intent and its exit: an intent older than 5 minutes is dropped by the next verb or list (V: `TestStaleIntentIsDropped`); the supervisor only reads it on exit 6.

## Live checklist (U, not run; needs the sandbox, one hidden instance, supervised bot, a Paladin and a Sorceress save)
1. Start `d2bot --managed --backend d2d --supervise --char Holy ...`; `GET /api/characters` lists both saves with the right class and level and `active` on Holy.
2. `POST switch {"name":"Zapp"}` from town: game leaves in about 15 s, log `supervise: switching character`, Zapp loads in town, `goals.json` is Zapp's, Holy's files sit in `characters/Holy/`. Switch back: Holy's orders, pause and ledger are back.
3. `POST` create `Newb` as Paladin: new level 1 hero in town, list shows it after the first save, then a kill of d2bot (not a switch) restarts and LOADS Newb.
4. Delete Zapp while Holy plays: archive appears under `characters-archive/Zapp/`, files gone, list shows it under `deleted`; restore brings it back identical; delete of Holy: 409 `active`.
5. Switch while a fight runs: Save and Exit works mid-fight; the runner does not rejoin Holy between exit and restart.
6. With `D2BOT_RUN` set: all four verbs 409 `measured_run`.
7. Whether the created hero's `.d2s` lands in the folder `saveFolder` finds; record the real path here.

## Phase B plan (UI, not in this card)
A Characters page reachable from the "..." menu and the Orders More sheet (not a cockpit slot), phone-first: a list of cards (class icon, name, level, last played, "playing now" badge), Switch with a confirm sheet ("Leave the game and play Zapp?"), New hero sheet (name field with the live rule hint, class chips from `classes`), Delete in a row menu with a type-the-name confirm and a line "It is archived first and can be restored", a Deleted section with Restore, a "Switching..." strip while `switch` is set, and the `caps.why` text in place of disabled buttons. Texts in the d.idle voice from a data file; fit fixtures at the eight sizes plus 412x780 and 360x640 like reset.
