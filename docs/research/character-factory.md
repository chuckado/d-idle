# Character factory: can we make test characters, levels and kits? (card charfactory)

> status: current research, 2026-11-02, docs only. Nothing here was run live: no game started, no character created or deleted, no save file touched. The one worker-host look was read-only (directory listings, a read of `.d2s` headers and of the cheat ledger, `timeout 30`, a scratch script removed afterwards). Tags: V = verified (a file or doc named, or the read-only look of 2026-11-02), S = another doc's statement or community knowledge, U = unverified. Backend: Game.exe 1.14d (original game) unless a row says d2d. Sources: `docs/cheats.md`, `docs/fresh-character.md`, `docs/oog.md`, `docs/saves.md`, `docs/d2d-backend.md`, `docs/forkcheats.md`, `docs/process/baseline-runs.md` (branch agent/baselineruns), ADR 0014 and the lab-rules delta (branch agent/memrules, commit edaaacf: NOT on origin/main when this was written, so cited as "the memory rule").

User question (2026-11-02): "Do we have the capabilities of creating characters and specific levels and such for testing?" Short answer: yes for a fresh character of any class and for level, XP, gold, stats, skill points and act, quest and waypoint state (all CHEATED setups, by tools that exist and are ledgered); yes for file-level copies and restores of a whole character (a tool exists but only inside one prefix); no for creating items on the original game, no proven hardcore creation, and no library of honest characters at set levels: exactly one honest in-band Sorceress exists. The missing piece is the snapshot library of section 4.

## 1. Inventory: what exists today

### 1.1 Creating a character

| piece | what it does | state |
|---|---|---|
| `tools/soak/fresh.sh <iN> <class> <Name>` | refuses unless the instance is idle and the name is on `tools/soak/fresh-names.txt` (Fresha..Freshg); starts that instance's game, runs `oog fresh`, verifies in game, Save and Exit, stops the game, checks the file with `d2savecheck` (name, class, level 1, exp 0, softcore, non-ladder, expansion) | V: Fresha, Freshb, Freshc, Freshd exist as Sorceress clvl 1 files on the fifth hidden instance (`.d2s` header read 2026-11-02). Live proof for classes other than Sorceress is not recorded in `docs/fresh-character.md` (checklist item 5 is open), U |
| `cmd/oog` + `internal/launch` `Flow.CreateCharacter`, `CreateVerified` | click-driven create screen: class (all 7 click points V live, header read back, 4 retries), name via WM_CHAR, expansion and hardcore checkboxes (code paths exist: `ch.Classic`, `ch.Hardcore`), difficulty button, yes/no popup | V live for class, name, softcore expansion Normal (`docs/oog.md`). Toggling the expansion and hardcore boxes was never tried live (oog.md: "UNVERIFIED"), the hardcore yes/no popup path is only offline-tested, U |
| `cmd/d2multi`, `cmd/oog host/join/party`, `tools/game.sh start/stop/helper` | multi-instance start/stop and LAN party flows | V live (`docs/oog.md` LAN section); not character factories |
| d2d fork | `--start-screen ingame --start-class N --start-name X [--start-hardcore] [--start-classic] [--seed N] [--no-save]` creates and enters at once; class N is UI order, not the `.d2s` class byte (N=1 gave a Necromancer) | V (`docs/d2d-backend.md`, `docs/d2d-spike.md`). Class, name, hardcore, classic and the map seed are all controllable; difficulty is Normal on a new character |
| d2d loading an existing character | the fork reads our `.d2s` from `<XDG_CONFIG_HOME>/d2d/save/<Name>.d2s` with `--start-screen charselect --start-name <Name>` then `key Return` | V: save, quit, restart, character loads (`docs/d2d-fork.md`); a copy of a Game.exe character joined a LAN game (`docs/d2d-lan-oracle.md`). The replay recipe records `save_sha256` (`docs/d2d-determinism.md`) |
| The bot | `d2bot` never creates a character (`docs/fresh-character.md`) | V |
| Agent rule | agents never create or delete characters (AGENTS.md); creation needs the main agent or a d2-live agent holding the instance and the user's or coordinator's go | V |

Limits: the char-select screen holds 8 slots, more needs the scrollbar, not implemented (`docs/oog.md`); V the busiest prefix today has 6 characters.

### 1.2 Setting level, XP, gold, stats, skills, progress (`internal/cheat`, `cmd/d2cheat`)

Rules the code enforces (V, `docs/cheats.md`): one entry point `cheat.Do`; a reason is required (a call without one is refused and still ledgered); character allow-list (`internal/cheat/policy.go` `Allowed`: Testy, Botty, Navi, Combi, Towni, Booti, Multi, Palli, Palli2, Amazi, Necki, Druki, Assi, Barby; V in code, so `docs/liveruns-checks.md` saying Druki and Assi are missing is stale); anything else needs `D2CHEAT_ALLOW=<Name>` for one task, and the Fresh* names are deliberately not on it; offline only (GameInfo ip/account/realm empty); save edits only with no game running, backup first, parse-back before atomic rename; the ledger (`cheats.jsonl` in the data dir on the worker host) is written before anything is applied, append-only, refusals included.

| backend | what can be set | state |
|---|---|---|
| Game.exe, memory (`d2cheat`) | `heal`, `god`, `gold`, `stashgold`, `xp` (only if the stat already exists), `level N` (also XP for that level, life/mana/stamina by class gains, +5 stat and +1 skill point per level), `stats str= dex= vit= ene= points= skills=`, `setstat` (movement speed, an insert experiment for magic find) | V live on the fifth hidden instance, persists through Save and Exit. `xp` level-up at the next gain, magic find effect, skill HARD levels: U / not written in memory |
| Game.exe, save file (`d2cheat save`, `internal/cheat/d2s`) | everything above plus `skills=ID:LVL,...`, `waypoints=MASK` (bit per difficulty, 39 bits), `difficulty= act=`, `progression=`, quest words (`quest set 1.den_of_evil:done ...`, `quest clear`, `quest reset`), `unlock-difficulty`, `act [nightmare/]N` | V live: level 25 + gold + stash gold loaded correctly; act start 1..5 loaded the right town; waypoint bits 0, 1, 3, 8; Nightmare unlock via progression byte (`docs/saves.md`); quest bits partly V (word indexes for Den, Sisters, Cain, acts II..V first words). Skill levels as hard points and the Hell unlock (progression 10): U |
| Game.exe, items | nothing: the `.d2s` item section ("JM") is copied byte for byte and never edited or created (V `docs/cheats.md`, d2s editor) | no item creation or editing on the original game |
| d2d (`debug` verbs, lab binary, ledger kind `d2d-debug`) | `debug stat` (life, mana, level, gold), `debug give <code> [pack\|belt] [q<quality>] [l<ilvl>] [unid]`, `debug equip`, `debug clearinv`, `debug unequip`, `debug durability`, `debug act`, `debug wp`, `debug goto`, `debug difficulty`, `debug quest list/set/bit/join`, `debug merc hire/kill/revive/hurt`, `debug spawn`, `debug states`, `debug mon`, `debug mana full` | V live per verb (`docs/forkcheats.md`, `docs/cheats.md` table); wrapped by `lab.Cheat` which ledgers BEFORE sending; offline `--no-save` throwaway only. Results do not transfer to Game.exe (`docs/d2d-backend.md`) |

Ledger policy for gates (V `docs/cheats.md`, restated by the memory rule): a character with ANY ledger line (also refused or failed) never counts toward a gate or baseline; `soak24.sh preflight` warns, the soak report writes "cheated, not a gate run". The memory rule (ADR 0014, branch agent/memrules): setup writes are allowed before a measured run if declared (`d2cheat runstart <run>`, `d2cheat ledger --since-run <run>` prints `writes after start: N`, exit 1 when N > 0 or the marker is missing); a run with a write after start is CHEATED and counts nowhere. The delta spec says a run with setup writes "may still count where the spec allows cheated setup": for baselines that need an uncheated character (`docs/cheats.md`: "NO cheats at all, ever") the stricter existing rule wins, so a cheated character is disqualified for them.

### 1.3 Save files (V, read-only look 2026-11-02, `docs/saves.md`)

A Game.exe character is a file set `<Name>.*` in each prefix's `Saved Games/Diablo II`:
- `<Name>.d2s`: the character (v96, 0x60): header, stats, skills, quests, waypoints, NPC, items including stash and belt. Size on the worker host today 940 to 2267 bytes (a clvl 1 Sorceress 958, Fresha 1498, Soaky 2267).
- `<Name>.key` (1142 bytes everywhere: key bindings, S), `<Name>.ma0` .. `.ma3` (automap data, S, varying sizes up to 74 KB; Soaky.ma0 is 0 bytes), `<Name>.map` (24 bytes, `0c000000 01000000 <seed> 00..`, the seed equals the header map id at 0xAB; V Botty and the whole table below). No `.d2x` or shared stash file exists in the listing: the stash lives inside the `.d2s` (V: no such file in any of the 6 directories).
- Per instance: one prefix per instance, own saves: `<data>/prefixes/<iN>` (or the dev-era `~/Games/d2lod-inst/<iN>`), five hidden instances i1..i5 plus i0 (the user's desktop game, agents never start it). Instances are bound by `tools/instances.conf`; one prefix has its own char list, so a character exists in the prefix where it was created. V d2d: saves live under the slot's own XDG dir; slots 0..11 ceiling, default pool 7 (`tools/d2dslot.sh`).
- Existing tools: `saves.Snapshot` (copies `.d2s .key .map .ma0-9` to `<backupDir>/<name>/<UTC stamp>/`, refuses in game, refuses a corrupt source, skips unchanged, retention 10 + 14 days), `saves.List/Prune/Restore` (Restore verifies, keeps the replaced files in `replaced-<stamp>/`, same in-game gate), `saves.Verify/Decode/ProgressFile`, `cmd/d2saves snapshot|list|restore|verify NAME [ID]`, `d2cheat qdump FILE`, `d2cheat restore`, `cmd/d2savecheck NAME CLASS`. V live: snapshot at the menu, edit, restore gave an identical hash and the game loaded it (`docs/saves.md` LIVEQ). Not wired into the sentinel or the bot (docs/saves.md API list). `baseline-runs.md` says "there is no restore tool in the repo": inaccurate, `saves.Restore` and `d2saves` exist; what is missing is placement into ANOTHER prefix and a manifest.
- The d2fleet container path already seeds `<Name>.*` from a base prefix into a slot's save dir once, never overwriting (`docs/container.md`, V by unit test): the same file-level copy idea.
- 41 save-edit backups from cheat edits sit under the ledger backups directory (V count).

Map seed (matters for restores): V live that the header map id is the same across 15 loads on one difficulty, through act, quest and waypoint edits, and changes when the game is entered on a different difficulty than the last saved one (`docs/saves.md`). So a copied `.d2s`/`.map` carries the same map id into every prefix. That the id determines the Game.exe dungeon layout is S (community knowledge, and the d2mapapi "map id = Act.dwMapSeed" note in `docs/d2-1.14d-memory.md`), never compared by us: U for our setup. Consequence: five runs from copies of one character probably replay the same Act I layout; monster placement and rolls are U (d2d: layout follows `--seed`, monster rolls differ per run, V d2d only).

### 1.4 Honest kit: how did Fresha get its kit?

By play alone (V `docs/process/baseline-runs.md` section 1, `docs/g1-log.md`, ledger count 0): `fresh.sh` made the clvl 1 file, `soak24.sh start i5 --for 4h --profile m24 --char Fresha` ran the managed bot (restarted three times on newer builds); every item came from drops, vendor buys and quest rewards. There is no scripted kit and no recorded gear list (U: the first checklist step is meant to read it). Soaky was played from clvl 8 to 15 over attempts 1 to 18 the same way.

### 1.5 Class coverage and the characters that exist (V, `.d2s` headers and ledger lines by NAME, read 2026-11-02)

The ledger is keyed by name, not by prefix; Palli exists in two prefixes (shared count 9).

| instance | character | class | clvl | ledger lines | note |
|---|---|---|---|---|---|
| i5 | Fresha | Sorceress | 11 | 0 | HONEST; the one honest in-band kit; died flag set (status bit 0x08, S: the character died at least once, as the baseline note says) |
| i5 | Freshb, Freshc, Freshd | Sorceress | 1 | 0 | HONEST clvl 1, unused |
| i0 | Botty | Sorceress | 1 | 0 | clvl 1, i0 is the user's desktop game: not for agents |
| i4 | Soaky | Sorceress | 15 | 5 (all refused or failed) | honest in play, out of the late band |
| i4 | Booti | Sorceress | 12 | 24 | cheated |
| i1, i2, i3, i5, i0 | Navi 24, Combi 30, Towni 20, Multi 30, Testy 9 | Sorceress | | 385, 60, 52, 65, 25 | cheated |
| i1, i2 | Palli | Paladin | 13 | 9 | cheated (probe char) |
| i2 | Pally | Paladin | 4 | 0 | honest, clvl 4, NOT on the allow-list |
| i2 | Necki 15 / Necci 20 | Necromancer | | 5 / 21 | cheated; Necci not on the allow-list |
| i2 | Druki | Druid | 4 | 0 | honest, clvl 4 |
| i3 | Amazi 3 / Amma 20 | Amazon | | 5 / 1 | cheated (Amazi), Amma has 1 ledger line |
| i3 | Assi | Assassin | 4 | 0 | honest, clvl 4 |
| i3 | Barby | Barbarian | 5 | 0 | honest, clvl 5 |

All 22 files: softcore, expansion, last played Normal, progression byte 0 (Nightmare never unlocked). So every class has a character (class coverage by creation: all 7), but only a Sorceress has an unattended managed run (`docs/fresh-character.md`: the other classes are smoke-test material); honest non-Sorceress characters stop at clvl 3 to 5.

### 1.6 Difficulty, quest state, waypoints, instances

Difficulty and act: V live via save edit (`difficulty=`, `act=`, `unlock-difficulty`) and by playing; quest words and waypoint bits: V live for the listed bits, S for the rest; d2d has `debug act|wp|quest|difficulty|goto` (V). Hardcore: d2d V (`--start-hardcore`), Game.exe U (flag exists in code, never toggled live). Instance slots: Game.exe i1..i5 (hidden, private displays, about 1.05 cores and 265 MB each, V `docs/multi.md` via baseline-runs) shared with other agents; d2d slots 0..11 (default 7) native processes. A restore is a file placement, so slots are not the limit; the char-select 8-slot cap and the exclusive instance lock are.

## 2. Capability table

"Honest" = no ledger line for the character, so the result may count toward gates and baselines. Cost: wall time unless stated.

| need | how it can be done today | cost | honest? | what is missing |
|---|---|---|---|---|
| fresh clvl 1 character of class X (Game.exe) | `fresh.sh iN <class> <Name>` (Sorceress V; other classes created by hand earlier: Pally, Druki, Assi, Barby exist V), name from the list | about 1 min (`docs/oog.md` timings) plus preflight | YES (never cheated) | live proof of `fresh.sh` for non-Sorceress classes; names list only Fresha..Freshg (4 used), the creating agent needs a go |
| fresh clvl 1 (d2d) | `--start-class N --start-name X [--seed N] --no-save` | seconds | n/a for Game.exe baselines (different backend); no ledger needed | none |
| honest character at clvl N with a plausible kit | play it: managed bot via `soak24.sh`, or restore a snapshot of such a character | Fresha: about 2 h to clvl 7, about 4 h to clvl 11 (V); 14 and 18: not measured | YES when played; a copy of a played character is "copy-of-honest" (section 4) | only ONE such character exists (Fresha, clvl 11); no snapshot library, no manifest, cross-prefix placement not built; the bot only plays a Sorceress |
| character at clvl N quickly | `d2cheat level N` (memory) or `save level=N` (file) | about 25 s per cycle | NO: any ledger line disqualifies (setup declared with `runstart`, still not a baseline character) | kit is not built (no gear, empty belt), XP/gold differ from honest play |
| character at a specific act, quest or waypoint state | `d2cheat quest/waypoints/act/unlock-difficulty` (save edits, game stopped); d2d `debug quest/wp/act/goto` | seconds to a minute | NO (ledgered) | Hell unlock U, many quest bits S, no named presets ("act 3 with Cain done"), Game.exe hard skill points U |
| character with specific items or skill allocation | skills: `d2cheat save skills=ID:LVL` (hard points U in game); items: Game.exe NONE; d2d `debug give/equip` | skills seconds; items impossible on Game.exe | NO | item creation or transplant on Game.exe; skill-point live check |
| hardcore character | d2d `--start-hardcore` V; Game.exe flag in code, never toggled live | minutes if it works | would be honest (no cheat) | live proof of the hardcore box and popup; the bot's death policy for hardcore is only offline |
| restore a character to an earlier state | `d2saves snapshot/restore` inside one prefix (V live, byte identical); `d2cheat restore` | seconds | restored copy is honest if the source was (provenance rule below) | cross-prefix placement, manifest, preflight, restore log, name-keyed ledger problem, the `.map` seed pin |
| copy Fresha's honest save into other prefixes (user approval 2026-11-02) | file copy of `Fresha.d2s .key .ma* .map` into another prefix's save dir, game stopped | seconds | YES as "copy-of-honest" | the tool (today `cp -p` by hand or the baseline launcher's snapshot step, prepared only), a restore log, and the same-name ledger rule (4.5) |

Summary: CAN fresh clvl 1 Sorceress and, by hand, the other six classes; CAN cheated level, gold, stats, act, waypoints, difficulty unlock and some quest state; CAN file-level snapshot and restore within one prefix. PARTIALLY honest mid-level characters (one exists), restores across prefixes, skills, hardcore. CANNOT items on Game.exe, an honest character at a chosen level on demand, an honest non-Sorceress beyond clvl 5, a verified hardcore character.

## 3. What the user's two approvals change (2026-11-02)

"approve save copies" and "go on the first late run" (baseline runs, `docs/process/baseline-runs.md` sections 3 and 9): file-level copies of Fresha's honest save into other prefixes are allowed, and the first late run may start. Placing a character into a prefix that did not have it is a file-level creation: it is covered by this approval only for Fresha and for the baseline runs; any other copy needs the same kind of go. This document treats such a copy as SETUP, never as a cheat (no gameplay value is edited), and records it (4.4).

## 4. Plan: an honest snapshot library

Idea: produce honest characters by playing and snapshot the whole save set at fixed levels; tests restore a snapshot by file copy. Honesty comes from the source run, not from the copy.

### 4.1 What to snapshot and when

- Levels: clvl 1, 5, 8, 11, 14, 18 (bands of the early and late baselines: early file ends at clvl 7, late file is clvl 11 to 14, `docs/process/baseline-runs.md` section 2). The bot reaches 18 only when the campaign gets deep into Act II or III: U, do not plan on it before the act-chain milestones; clvl 1 snapshots already exist (Freshb, c, d).
- Per class later: Sorceress first. Paladin is the recommended second class (`docs/classes/ranking.md`) but has no unattended run yet; the honest Pally, Druki, Assi and Barby (clvl 3 to 5, ledger 0) are clvl 1 to 5 class fixtures as they stand.
- When: at the main menu after Save and Exit, game stopped (the `saves.Snapshot` gate). A soak run that ends with a graceful save-and-exit (V, attempt 18: save at the 2 h mark) leaves the next snapshot for free. Loading a save always starts in the last act's town (S, community behaviour), so where the bot stood at save time does not matter, only the kit does.
- Pick the crossing, not the middle: snapshot at the first save with clvl >= band and the belt stocked, then record what is in the pack (an honest bot may sit at 0 gold and no potions, as a14 did: clvl 11 and 0 gold, V baseline-runs).

### 4.2 Manifest (one JSON per snapshot, next to the files; the files themselves are never committed)

```
{ "id": "sor-c11-20261030-a920fef",          // class, clvl, date, bot commit (short)
  "name": "Fresha", "class": "sorceress", "clvl": 11, "xp": 0,   // header 0x28, 0x2B; XP from the stat block
  "hardcore": false, "expansion": true, "difficulty": "normal", "act": 1,   // 0xA8..0xAA, status bits
  "died_flag": true, "gold": 0, "stash_gold": 0,
  "kit_sha256": "<sha256 of the item section 'JM'..end>",   // what she wears and carries
  "kit_summary": "weapon, armor codes, belt, potion count",   // from d2savecheck or a reader
  "map_id": "e7db542b", "map_file_sha256": "...",   // see 4.6
  "files": { "Fresha.d2s": "<sha256>", ".key": "...", ".map": "...", ".ma0": "...", ... },
  "source": { "run": "<soak run id>", "bot_commit": "a920fef", "profile": "m24", "instance": "i5", "log": "<path on the worker host>" },
  "ledger": { "lines_at_snapshot": 0, "status": "honest" },   // ledgercheck.sh result
  "created": "<UTC>", "verified": { "verify": "ok", "smoke": "<id of the load-and-continue run or none>" } }
```

Values come from existing readers (`saves.Decode`, `d2cheat qdump`, `d2savecheck`); XP and kit hash need small additions. "ledger.status" is `honest` only when `ledgercheck.sh` finds no line for the name at snapshot time; otherwise the snapshot is stored as `cheated` (kept for debugging fixtures, never offered to gates).

### 4.3 Integrity checks

1. `saves.Verify` on the `.d2s` (signature, version, size, checksum, markers, stat block) at snapshot and at every restore.
2. sha256 of every file recorded and re-checked before a restore copy and after it (byte compare of the placed `.d2s`).
3. Smoke ("loads and continues"), once per snapshot and again whenever the game helper or bot build changes: place the snapshot into a free hidden instance, start the game, read name, level and gold through the normal readers, Save and Exit, `d2savecheck` shows the same name/class/level. Cost about 1 minute of game time plus the lock. UNVERIFIED so far for a cross-prefix copy; V only for a same-prefix restore of Multi.
4. Ledger check: no entry for the name at snapshot; the `ledger.status` field.

### 4.4 Restore tool design (a new `d2saves place`, or a small `cmd/d2restore`; Go, standard library, offline unit tests with a fake directory)

```
d2restore place --snapshot ID --prefix iN [--as NAME] --reason "baseline late run 3" --run RUNID
```
Preflight, all must pass, otherwise nothing is touched and the exit code says why:
1. The target instance is one of i1..i5 (never i0), its lock is free, there is no `Game.exe` carrying the instance marker, no bot process, no soak unit, no restart notice (same probes as `fresh.sh`).
2. NEVER OVER A LIVE CHARACTER: if a `<Name>.d2s` exists in the target, it is replaced only when its sha256 equals a recorded snapshot or placement (so no unsnapshotted progress is lost); otherwise refuse. When it is replaced the old files go to `replaced-<UTC>/` (never pruned), as `saves.Restore` does.
3. The snapshot's own files re-hash to the manifest and `saves.Verify` passes.
4. The target char-select has a free slot (fewer than 8 characters, V today at most 6).
5. The name does not collide with a character that has ledger lines (see 4.5).
Then copy with `cp -p` semantics via a staging dir and rename, compare bytes, run `saves.Verify`, append a line to a restore log (`<data>/ledger/restores.jsonl`, same append-only rule as the cheat ledger): time, actor, snapshot id, source name, target prefix, `kind: "setup-copy"`, reason, run id, sha256 before and after. The log line is the SETUP declaration the memory rule asks for; nothing in `cheats.jsonl` is written because no value was edited. Print the line to paste in the run record: `setup: copy-of-honest <snapshot id>`.

### 4.5 Rules

- A restored copy is "copy-of-honest" provenance, not "cheated": the state equals one reached by honest play, the only difference is that the play was not repeated. Reports must say so and carry the snapshot id; a baseline note says "start kit: copy of <id>".
- Never restore over a live character (4.4 item 2); never into i0; never while the instance has a game, bot or soak unit.
- Log every restore as a SETUP action before the run starts (4.4); the run marker (`d2cheat runstart`) comes after it. A copy made after the marker, or any cheat on a restored copy, makes the copy CHEATED.
- Same-name problem (open): the cheat ledger, `soak24.sh preflight` and the baseline preflight key by NAME. Five copies named Fresha in five prefixes share one ledger identity, so one cheat on one copy poisons all copies and the gate cannot tell them apart. Options: (a) key the ledger and ledgercheck by `(instance, name)`; (b) place copies under distinct names with a header rename (the d2s editor can rewrite the name and checksum, the file names change, U for the `.ma*` and `.key` files, and the edit is itself a save edit that would need to be declared); (c) accept it and write "copy of Fresha, i4" in each record. (a) is the smallest and keeps the file bytes unchanged.
- Fresha is used up when a run advances it: snapshot BEFORE the first late run (the prepared launcher does, with `baseline-kits/<set>`), and also snapshot Freshb/c/d clvl 1 before their runs (a fresh character is used up after its run).
- Snapshots are only taken from characters whose ledger status is honest; a cheated character can be snapshotted as `cheated` for fixtures but never offered to gates.

### 4.6 The map seed pin (flag)

The header map id (0xAB) and the `.map` seed are the same in every copy (V: stable across loads on one difficulty). If it also pins the Game.exe dungeon layout, all five late baseline runs from one kit replay the same Act I layout and are less independent than three runs of one character that kept changing its state (the old a14, a16 and a18 runs were also one character with one id, V 15 loads stable on one difficulty), so this is not a new loss compared with the old baselines, but it does weaken the "independent live runs" claim. Cheap tests, no cheating: (1) the first late run's checklist item (compare the `.map` before and after a game: the id stays); (2) enter Act I on two copies in two prefixes and compare a level dump (`d2mapdump`) of the Blood Moor or the Den: same rectangles means the layout is pinned (U). If pinned, a per-run id change is available by entering a game on another difficulty first (V: changes the id) or by using different source characters; both alter the character, so decide only after the test.

### 4.7 Cost

| item | wall time | agent time |
|---|---|---|
| clvl 1 snapshots (Freshb, c, d) | minutes (copy + manifest) | about 0.5 h with the tool |
| Fresha clvl 11 snapshot | minutes | included |
| clvl 5 and 8 Sorceress | about 1 to 2.5 h of bot time, interpolated from Fresha (clvl 7 about 1.9 h, clvl 11 about 4 h, V), U | the baseline fresh runs already produce them: a run ends at clvl 7 or later, snapshot at the end |
| clvl 14 | Soaky needed several attempts of 2 h from clvl 11 to 14 (V `docs/g1-log.md`), call it 4 to 6 h more, U; each late baseline run that ends at 12 to 14 is a candidate | free from the late runs |
| clvl 18 | not estimated: XP per hour is about 9 to 11 thousand at clvl 11 to 14 (V g1-log) and needs content beyond Act I | later |
| each restore | seconds to place, about 1 minute with the smoke | automatic |
| per class | about 10 h of unattended bot time for 1 to 14 if the class has a managed run | a class run card each |

So the library is mostly a by-product of runs that must happen anyway: the cost is the tool, the manifest and the discipline of snapshotting at each end-of-run, not extra play.

### 4.8 Interaction with the baseline runs (`docs/process/baseline-runs.md`, branch agent/baselineruns)

- The prepared launcher already does the file placement (its step 2, `cp -p` into a lane's prefix, `.d2s` compared, replaced files moved aside) and a kit snapshot into `baseline-kits/<set>`: the snapshot library replaces that snapshot step and adds the manifest, preflight, restore log and ledger rule; the launcher should call the restore tool once it exists. Until then the launcher's own copy is acceptable under the user's approval.
- The late baseline set wants five runs from one kit (Fresha): with approved copies this is possible, but all five share the ledger name (4.5) and probably the layout (4.6). Both are visible before starting.
- Run order that keeps the user's two approvals useful: (1) snapshot Fresha and Freshb/c/d; (2) place Fresha into one other prefix and run the first late run alone (the existing checklist item 6); (3) compare `.map` and, if cheap, a level dump of the two copies; (4) then the full set.
- A d2d run can start from the same file: d2d loads our `.d2s` from its slot's XDG save dir (V), so the snapshot id is the common start state for both backends, and `save_sha256` in the replay recipe records it. That is the cleanest fidelity comparison (same class, level, kit), better than a `-setup` cheat script (`docs/d2d-soak.md`).

## 5. Ranked backlog cards (one line each) and order

1. SNAPLIB-TOOL: `d2restore place` (cross-prefix, preflight, never over live, sha compare, restore log) plus manifest v1 writer and reader, offline tests; wire the baseline launcher to it. Smallest step that makes the user's approved copies safe.
2. LEDGERKEY: key the cheat ledger, `ledgercheck.sh` and the baseline preflight by (instance, name), or decide the rename option, so copies are not one identity.
3. SNAPSEED-TEST: verify live whether copies replay the Act I layout (compare `.map`, level dumps of two copies); decides whether late baseline runs are independent.
4. SNAPSMOKE: the loads-and-continues smoke for a placed snapshot (name, level, gold read, Save and Exit, `d2savecheck`), run once for Fresha in another prefix; first live proof of a cross-prefix copy.
5. HONESTLADDER: end-of-run snapshot hook in `soak24.sh` (at the menu after the graceful save) writing a manifest from the run; takes clvl 5, 8, 14 snapshots from the baseline and later runs.
6. D2DSNAP: place a snapshot into a d2d slot's XDG save dir and record `save_sha256`; first d2d run from the honest clvl 11 file (fidelity on equal start).
7. FRESHCLASS: live proof of `fresh.sh` for a non-Sorceress class and of the expansion and hardcore boxes (Game.exe), so a hardcore or classic character can be made honestly; add names to `fresh-names.txt` for the other classes.
8. PRESETS: named declared-setup recipes by save edit (act and difficulty, waypoint sets, quest state) for drills only, each with `runstart` and a reason, never baselines; verify the Hell unlock, skill hard points and the quest bits still tagged S.
9. ITEMKIT: research whether items can be placed on Game.exe (transplant the item section from an honest snapshot, or a `.d2s` item writer); until then d2d `debug give` is the only item creator.
10. CLASSLADDER: per-class honest runs and snapshots after Sorceress (Paladin first), each needing a managed run for that class.

Recommended order: 1, 2 (together), 3 and 4 in the first late run's window (they need only a quiet instance), then 5, then 6; 7 and 8 whenever an instance is free; 9 and 10 later.

## 6. Open questions and dead ends

- Open: does the map id pin the Game.exe layout (U); do `.key` and `.ma*` files need to travel with the `.d2s` (U: `saves.Snapshot` copies them, a load without `.ma*` was never tried); does a hardcore character create through the checkbox (U); does the managed bot behave the same on a restored copy (smoke, U).
- Dead ends and rejected: `d2cheat level N` as the route to an honest mid-level character (ledger disqualifies, no kit); running the late baselines on d2d (the baseline must come from the original game); leveling five fresh characters honestly first (about 4 h of bot time each, `baseline-runs.md` section 8); creating or deleting characters by an agent on its own (AGENTS.md).
- Hygiene: no save file, name list with history or host name is committed here; manifests and snapshots stay on the worker host.
