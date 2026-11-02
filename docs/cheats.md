# Cheats: logged test cheats (agent CHEAT)

User rule (2026-10-28): we control the game's files and memory, so for testing we may cheat on the throwaway
characters (Testy, Botty, Navi, Combi, Towni, Booti, Multi, Palli, Palli2, Necki, Amazi, Druki, Assi, Barby), offline only, and EVERYTHING is recorded. Later the same
mechanism may back cheats as an earned meta reward (ideas at the end, nothing built).

Code: `internal/cheat` (policy, ledger, memory and save cheats, overlay), `internal/cheat/d2s` (.d2s editor),
`cmd/d2cheat` (CLI), `mem.Writer` + `Proc.WriteAt` in `internal/mem`. Tags: VERIFIED LIVE = observed on instance i5
("Multi", GE-Proton11-11, 1.14d, 2026-10-28); UNVERIFIED = not observed.

## Memory rule (ADR 0014, user decision 2026-11-02)
> "We can read/edit memory whenever we want, but we have to make sure we aren't using memory editing as a crutch."

Reading is always allowed. Writing is allowed for SETUP, INSTRUMENTATION, DIAGNOSIS and FIXTURE capture (wording and the CRUTCH RULE: AGENTS.md "The live games"), always through this tool or a declared instrumentation module, with a reason, in the ledger. This replaces the old "memory is READ-ONLY except test cheats" rule; everything below (allow-list, offline only, ledger) still applies to the tool.
- Measured runs (gate, soak, baseline, fidelity): setup writes before the start only, declared in the run record; no gameplay-affecting write after the start. Otherwise the run is CHEATED and counts toward nothing.
- Mark the start with `d2cheat --reason R runstart <run>`; set `D2BOT_RUN=<run>` for the writes of that run so entries carry the run id. `d2cheat ledger --since-run <run>` prints the writes after the marker and the line `writes after start: N` (exit 1 when N > 0 or the marker is missing: "no marker" never reads as 0). The record of a soak or gate pastes that line.
- A feature is not done until its live drill passes with writes disabled. Code injected into the game (DLL, hook) is INSTRUMENTATION only: the user's go per hook set, documented, logged.

## Rules the code enforces
- One entry point: `cheat.Do(ctx, Request{Kind, Params, Reason, Actor})` (`Env.Do` for tests). **Reason is required**; a call
  without one is refused and still written to the ledger (reason `(none)`).
- Actor = `Request.Actor`, else `$D2BOT_AGENT`, else `unknown`. `tools/live.sh` does not forward env to the remote run, so pass
  `--actor NAME` to d2cheat.
- Character allowlist (Testy, Botty, Navi, Combi, Towni, Booti, Multi; case-insensitive). Memory cheats read the name from the
  running game, save cheats from the file header (and the file name must match it). Anything else is refused.
- Offline only: memory cheats read `GameInfo` (RVA 0x3A0438: server ip +0x37, account +0x8D, realm +0xD5) and refuse if any is
  non-empty. VERIFIED LIVE that all three are empty in single player; what Battle.net fills in is UNVERIFIED (we never go online).
- Save edits: the character must not be in a game (no Game.exe for the instance, or the game sits at a menu, VERIFIED LIVE that
  editing while the game is at the main menu works), a backup copy goes to `<data>/ledger/backups/` (`$D2CHEAT_BACKUPS`; older ones stay in `/mnt/cold/d2lod/backups/`, ledger entries keep their absolute path) as `<Name>-<UTC>.d2s` first
  (if the backup fails the save is not touched), the new file is parsed back before an atomic rename.
- The ledger must be writable before anything is applied. Memory edits are made per instance: `tools/live.sh multi d2cheat ...`
  (i5 only for this agent; other agents use their own agent name).

## Policy for fresh test characters (M1, 2026-10-30)
Applies to characters made by `tools/soak/fresh.sh` (names in `tools/soak/fresh-names.txt`, see [fresh-character.md](fresh-character.md)).
- Those names are NOT on the built-in allow-list (`internal/cheat.Allowed`), deliberately unchanged: no cheat can touch them unless someone
  names them with `D2CHEAT_ALLOW=Freshb` (or the allow file) for one task. Adding a name there is a decision to note in the report.
- Baseline for any gate soak (G1, G2, 24 h): NO cheats at all, ever, on that character. Not before the run, not during.
- The only planned exception is the merc probe (plan-v2 M4): gold only (`d2cheat gold`), with a reason, on a character that is then
  retired from gates. Level, XP, stats, god and save edits on a fresh character make it a debugging character, not a gate character.
- Rule: a character that appears in the ledger (any kind, also refused or failed lines) never counts toward a gate. `soak24.sh preflight <iN> NAME`
  and `soak24.sh start ... --char NAME` WARN when the ledger on the worker host has an entry for NAME (`tools/soak/ledgercheck.sh`); the warning is
  advisory, the verdict is made by whoever reads the soak report: write "cheated, not a gate run" into it.

## Ledger
`<data>/ledger/cheats.jsonl` (default `~/.local/share/d2bot`; the old `~/.local/state/d2lod/cheats.jsonl` is moved there on first use) on the machine that runs the tool (the worker host; override `$D2CHEAT_LEDGER` or `--ledger`). Append-only
JSONL, one `Entry` per attempt, refusals and failures included:

    {"time":"2026-10-28T18:23:02Z","actor":"multi","instance":"i5","character":"Multi","kind":"level",
     "params":{"value":30},"before":{"server.level":25,...},"after":{"server.level":30,...},
     "reason":"...","success":true,"error":"","backup":"/mnt/cold/d2lod/backups/Multi-....d2s"}

Kinds: runstart (marker, no write) heal god gold stashgold level xp stats setstat saveedit dataoverlay. God writes a `phase:start` line when it begins
(a killed process still leaves a trace) and the normal line when it ends. Reader: `Ledger.List(Filter{Since, Until, Actor,
Instance, Character, Kind, OnlySuccess})`, `Summarize(entries)`, `Ledger.Cheated(since, character)` and `cheat.Cheated(since)`:
telemetry/dashboard call these with a run's start time to flag or exclude cheated sessions. CLI: `d2cheat ledger [--since 2h|RFC3339]
[--char N] [--kind K] [--actor A] [--json]`. The ledger is on the worker host: a dashboard elsewhere must read it over ssh or have it synced.

## d2d debug verbs (lab binary, ledger kind `d2d-debug`)
The d2d lab build has its own test verbs: `debug <sub-verb> ...` on the control socket (not game memory). They are cheats: allowed only on the
throwaway characters, offline (the d2d lab is always offline, `--no-save`), and every call is ledgered BEFORE it runs (kind `d2d-debug`,
params `verb` and `id`, an actor, a REQUIRED reason, then a second line with the reply) by the harness wrapper `lab.Cheat` (cmd/d2drill). Nothing
else may send a `debug` verb. A character that appears in the ledger never counts toward a gate (rules above). Which verbs exist in the lab
binary as of 2026-10-31 (the patches live in the private fork, never in this repo; results and drills they unblock: docs/forkcheats.md):

| Verb | Changes | Notes |
|---|---|---|
| `debug stat <id> <v>` | a character stat (6 life, 8 mana: 24.8 fixed point; 12 level; 14 purse gold; 15 stash gold) | read one tick later |
| `debug give <code> [quality] [unid]` | a new item on the ground at the feet; quality 0 (by item level), 1 low, 2 normal (default), 3 superior, 4 magic, 6 rare, 7 unique; `unid` leaves it unidentified whatever the quality (a magic or better ground item is unidentified already) | pickup rules then apply (belt first); `err quality 0..8` |
| `debug give <code> pack [q<quality>] [l<ilvl>] [unid]` | a new item straight into the pack's first free spot; `q4` makes it magic (item level 30 unless `l<n>`), `unid` clears its identified flag | reply `ok pack <col> <row> quality=<q> identified=<0/1>`; `err no room`; `err quality 1..8`; `err unknown token` |
| `debug give <code> belt [box] [q<quality>] [l<ilvl>] [unid]` | a new item into the box `rules::belt_slot` would pick (the pickup's choice), or the box asked for | reply `ok belt box <n> ...`; `err no belt box` when none fits or the box is taken |
| `debug equip <code> <slot>` | a normal item worn in body slot (8 = belt) | replaces what is worn |
| `debug clearinv` | the pack grid emptied | belt and worn items stay |
| `debug belt clear` | every belt box emptied | the belt itself stays worn |
| `debug unequip <slot\|belt>` | what is worn there is removed (gone) | belt boxes stay: clear them first |
| `debug durability <slot\|all> <v>` | worn items' durability set to v, capped at the item's maximum | `debug wear` halves instead |
| `debug act <1-5>` | the quest flags the waypoint panel reads: tabs for acts 1..n shown, the others hidden | no other quest state or level gating changes; reply `ok tabs=12` |
| `debug hit <dmg>` | one monster-style blow of `<dmg>` points on the player through the game's own hurt path (life, armour wear, death at 0) | always enters the get-hit mode unless it kills; works in town |
| `debug merc hire <type>` | the character's merc becomes a hireling.txt Id `<type>` (no gold), spawned beside the player | `err no such type; known: ...` |
| `debug merc kill` | a blow that kills the live merc, its death plays out, then it counts as dead | |
| `debug merc revive` | the dead merc is alive again beside the player (no gold) | `err no dead merc` |
| `debug merc hurt <dmg>` | one blow of `<dmg>` points on the live merc through the game's own hurt path (life, death at 0: it dies and plays its death) | reply `ok merc life=<n> mode=<m>`; `err damage > 0`; `err no live merc` (none hired, dying or dead) |
| `debug spawn <monster> [n] [near] [champion\|unique [mod ids]]` | `n` (1..20, default 1) monsters of a MonStats Id (`zombie1`, `fallen1`) or row number, made by the level population's own constructor (`make_monster`, then `make_boss` for the variants) and added to the level's monster list on a ring 3 cells round the player; `champion` (mod 16 unless mods are given) and `unique` (mods optional, e.g. `unique 5 6`) make the boss variants; `near` is accepted and is the only placement | field levels only (the monster list of the level the player stands in; not town); reply `ok spawned=<n> ids=#a,#b <name> hp=<h> lvl=<l>`; `err no such monster`, `err n 1..20`, `err unknown token <t>`, `err no monster level here` |
| `debug quest list` | prints every non-zero quest record of the active difficulty: `<record>\t<flags>\tact<n>\tbits=<set bits>` | read only |
| `debug quest set <record> <flags>` | the 16 flag bits of quest record 0..47 written whole (`0x0002` = bit 1; decimal or hex), then the act's join re-runs so the quest machines (Den, Blood Raven, Andariel, ...) take the state; the quest NPCs' hidden flags are refreshed | `err quest 0..47`, `err flags 0..0xffff`, usage line on a missing or non numeric argument |
| `debug quest bit <record> <bit> on\|off` | one flag bit of a record, then the same re-join | `err bit 0..15` |
| `debug quest join <act 1-5>` | re-runs one act's join from the current flags (what a new game does) | `err act 1..5` |
| `debug pin on [r]` | a ring of blocked collision subtiles (the walk grid's 0x01 bit) at Chebyshev distance `r` subtiles (3..8, default 4) round the player: a pocket nothing walks out of (units, click walks, the unstick) | `err radius 3..8 subtiles`, `err already pinned (pin off first)`, `err too near the level edge`; reply `ok pinned cells=<n>` |
| `debug pin off` / `debug pin status` | the ring's cells put back to their values; status prints `pinned=<0/1> cells=<n>` | `err not pinned` |
| `debug missiles` | prints one tab-separated row per missile in flight (the list `state missiles=` counts): `unit, missile id, x, y, dx, dy, speed, owner type, owner unit, ttl, friendly, name`. unit: place in the list (not stable between calls); missile id: Missiles.txt `Id` (-1 unknown); x y: cells, the frame of `monsters`; dx dy: unit vector of the flight (0 0 when it does not move); speed: cells per second; owner type: 0 player, 1 monster or merc, -1 none (a hit effect, a shrine); owner unit: the shooter's unit id (the player's, the `monsters` `#id` of a monster; a merc's missile gives the merc's unit); ttl: ticks of 40 ms until it ends (-1 unlimited); friendly: 1 for the player's or merc's missiles (they hit monsters, not the player); name: Missiles.txt name. Hit effects (`shamanexp`) are listed too, with speed 0 and owner -1 | read only (reads `Fight::missiles` and the world clock, writes nothing; a missile made in a tick joins the list after that tick, as it joins the count) | no rows then `ok` when none; `err usage: debug missiles` with an argument |
| `debug mana full` | the lead player's mana set to max mana | reason required (ledgered by `d2dctl.SkillDebug` or `d2drill -probe` before sending); offline --no-save only, throwaway character |
| `debug states set <skill> [level]` / `debug states clear <skill>\|all` | the player's self state started as a cast would (duration, Bone Armor pool) or removed | reason required (ledgered first); offline --no-save only, throwaway character; bare `debug states` stays read only |
| `debug mon set <id> [life\|frac\|maxlife\|dmg\|res ...]` | a live monster's life, damage range or absolute resistances | reason required (ledgered first); offline --no-save only, throwaway character |
| `debug cast <skill> <x> <y>\|unit <id>\|near` | queues a real skill cast (test action) | reason required (ledgered first, test_action true); offline --no-save only, throwaway character |
| `debug help` | one line per sub-verb: the lab verbs with their usage, the others by name | read only |
| `debug wp <i>\|all\|level <id>` | waypoint bits | |
| `debug goto <level>`, `debug quest <q>`, `debug die`, `debug difficulty`, `debug warp` | position, quest bits, death, difficulty | as the fork's usage line says |

## What works, memory cheats (VERIFIED LIVE unless noted)
Mechanism: `/proc/pid/mem` write (needs the ptrace right `tools/live.sh` grants). In single player the process holds TWO copies of
the player UnitAny: the client's (global `PlayerUnit`, RVA 0x3A6A70, also in the "server table" 0x3A5E70 that internal/d2 reads;
that table is client side) and the embedded game server's authoritative one. There is no static pointer to the server unit (only
heap structures reference it), so `serverPlayer` scans the private rw mappings for a second UnitAny with type 0, same class and id,
an extended stat list and the same name (about a second). Each unit has a stat list with a base vector (+0x24) and a merged vector
(+0x48, what GetUnitStat reads for players); entries are `{WORD sub, WORD id, int32 value}`; life/mana/stamina are 24.8 fixed point.

| What | Result |
|---|---|
| Server life write | Reaches the client within a second (client showed 5/40 after writing 5 to the server). |
| Server attribute write (str) | Does NOT reach the client (client stayed 10). Attribute-like stats are therefore written to BOTH units, both vectors. |
| `heal` | Server life/mana/stamina := max (both vectors) plus client copy. Works, instant. |
| `god on --for D` | No invincibility flag found. Mechanism: max life and life raised to the 21-bit limit (8191), refilled every 5 ms (about 180 refills/s) for life/mana/stamina, original max life restored at the end (a Save and Exit afterwards keeps the real value). Loop, start/stop and restore VERIFIED with simulated 150-damage hits. Real monster test NOT done: `cmd/combatlive` finds the game with `input.FindGamePID`, which ignores `D2BOT_INSTANCE` and picked Testy (i0), so I did not run it. If the process dies mid-way the inflated max life stays (ledger "before" has the original). |
| `gold N`, `stashgold N`, `xp N` | Work, and persist through Save and Exit and re-entry (checked by parsing the saved file and re-reading memory). Only if the character already has the stat: a fresh character has NO exp, gold, stash gold, stat-point or skill-point entries (they appear on first gain), and a full vector cannot grow (the game owns the allocation). Then: use the save edit (adds the stat), or the experimental insert below. Gold cap and stash cap (level based) are not enforced by a memory write. |
| `xp N` level-up | Writing experience changes the stat only. Whether the server levels up at the next experience gain is UNVERIFIED (no kill; the written value stayed below the next threshold). Use `level` or the save edit. |
| `level N` (memory) | Sets level on both units, experience for the level (from experience.txt via internal/data), max life/mana/stamina by the charstats per-level gains and refills, +5 stat and +1 skill points per level when those stats exist. Consistent after Save and Exit and re-entry (verified 25 to 30). Skills/quests are not touched. Items/skills that check the level work at once. |
| `stats str= dex= vit= ene= points= skills=` | Work (both units); vit/ene also move max life/stamina/mana by the class gain. Persist through re-entry. |
| Movement speed (stat 67, 100 = normal) | 200 made a 9-subtile walk take 0.32-0.40 s instead of 0.81 s (2 runs each). Works as `setstat stat=67 value=200`. Not saved. |
| Magic find (stat 80) | Absent on a character. `setstat stat=80 value=300 insert=1 unit=server` inserts it into a spare slot of the server merged vector (and base if room) keeping the vector sorted. It stayed for 10+ s. Whether drops use it, and whether a stat recalculation drops it, is UNVERIFIED. Lost at re-entry. Fails when the vector is full. |
| Other item-find stats (gold find 79 etc.) | Same insert path, UNVERIFIED. |

What did not work / limits: stats absent from a vector cannot be added except by the insert experiment (free slots vary: seen 0-3);
no experience table or drop rate can be changed in memory by this tool (use the overlay); invincibility against a one-shot of more than
8191 life is not covered; skill levels (hard points) are not written in memory (they live in a skill list, not the stat list; use the save edit).

## What works, save-file cheats (`internal/cheat/d2s`, VERIFIED LIVE on i5)
Pure Go editor for the v96 file (1.10 to 1.14d; D2R 0x61 is rejected). Layout from the Phrozen Keep "Diablo II Character File Format"
and the open-source parsers nokka/d2s and dschu012/d2s (version-96 branches); offsets re-checked against real 1.14d saves:
signature 0xAA55AA55 at 0, version 0x60 at 4, size at 8, checksum at 0xC (`sum = rotl(sum,1) + byte` over the file with the checksum field
zero), name at 0x14, status 0x24 (0x20 expansion), progression 0x25, class 0x28, level 0x2B, difficulty bytes 0xA8..0xAA (bit 7 active, low
bits = act), quests "Woo!" at 0x14F (3 x 96 bytes), waypoints "WS" at 0x279 (3 x 24 bytes: 2 header + bit field), NPC "w4" at 0x2CA,
stats "gf" at 0x2FD (9-bit ids, per-id widths 10/10/10/10/10/8/21x6/7/32/25/25 for ids 0-15, terminator 0x1FF, byte padded), then skills
"if" + 30 bytes, then items "JM" (copied byte for byte). Round trip of Multi, Botty and Testy copies re-encodes identically.
- Tests: `go test ./internal/cheat/...` (synthetic fixtures built by hand; `D2S_REAL=/folder/with/copies go test ./internal/cheat/d2s -run Real -v`
  round-trips real copies; no save file is ever committed).
- Edits: `level` (also exp for the level from experience.txt, +5 stat/+1 skill point per level gained, max life/mana/stamina by class gains, refill),
  `exp gold stashgold statpoints skillpoints str dex vit ene maxlife maxmana maxstamina`, `skills=ID:LVL,...`, `waypoints=MASK` (bit per
  difficulty: 1 normal, 2 nightmare, 4 hell; all 39 bits), `difficulty=D act=A` (act 0-4), `progression=N`; quest words, waypoint bits, unlock and start act: see Progress edits below.
- VERIFIED LIVE: `save Multi level=25 gold=50000 stashgold=100000 waypoints=1` while the game sat at the main menu, then the OOG start:
  the game loaded level 25 with exp 1640359, 75 stat and 15 skill points, 95 life, gold and stash gold (client and server unit identical).
  The whole cycle (leave, edit, start) takes about 25 s. UNVERIFIED in game: the waypoint names per bit (PK table), `progression` and
  `difficulty` unlocking Nightmare/Hell, skill levels, quest bytes (now editable, see Progress edits; also UNVERIFIED).

### Progress edits: quests, waypoints, difficulty, act (agent CHEATQ; offline tested; live results by LIVEQ 2026-10-28 below)
All go through `saveedit` (reason, allowlist, game stopped, backup, checksum, parse-back, ledger; the ledger snapshot now has `questsD`
= non-zero words as `index=0xWORD`). Two edits in one second get backups `-2`, `-3`. Every item below is **UNVERIFIED in game**
until the checklist at the end of this section has been run; the quest bit meanings stay S/U as in docs/saves.md (this tool writes raw
words, it does not make a bit more true).

    d2cheat --reason "..." quest set Multi 1.den_of_evil:done 1.sisters_burial_grounds:started nightmare/2.radament:done w9:0x8001
    d2cheat --reason "..." quest clear Multi 1.den_of_evil        # word := 0
    d2cheat --reason "..." quest reset Multi [normal|nightmare|hell]   # zero all 48 words (default all three)
    d2cheat --reason "..." waypoints Multi all | none | cold_plains+stony_field+4      # prefix nightmare/ or hell/ ; ID = bit 0-38 or name
    d2cheat --reason "..." unlock-difficulty Multi nightmare|hell      # progression byte := max(cur, 5 or 10)
    d2cheat --reason "..." act Multi [nightmare/]3                     # start in act 3 town (see below)
    # same through save params: quest=  questclear=MASK  wp=  unlock=1|2  startat=[D/]A  (plus the older waypoints= difficulty= act= progression=)

- Quest ids and act check come from `internal/saves` (the table is not duplicated: the word index of an id is found by setting one word on a scratch
  copy and asking `saves.ProgressOf`; so a corrected saves table is picked up automatically). `wN` writes raw word N (0-47) when the table is wrong.
- States: `done` = 0x8001 (reward granted + completed-before, what `saves` calls Done), `started` = 0x000c (what Testy's real word 1 reads), `pending` = 0x0006,
  `none` = 0, or any 16-bit number (decimal or 0x..). Which value the game itself writes for a finished quest is unknown (U); try raw values in the live check.
- `waypoints`: `all`/`none` replace the 39 bits, ids add. Names are saves' waypoint names lowercased with `_`. Waypoint order and names are S.
- `unlock-difficulty` only moves the progression byte (U meaning: 5 = Normal beaten, 10 = Nightmare beaten, as in saves.Progress). It does not touch quests.
- `act N` = `startat`: sets the active difficulty byte with act N-1 (V layout, behaviour U), the town waypoint of act N, the "act X done" word of every earlier
  act and the intro word of act N (positions S for acts 1-3, U for 4-5), plus the progression unlock for Nightmare/Hell. For act 4 the saves table has the act 4
  "done" word at 28 and docs/progression/quests.md lists act 5 at 35-40 against saves' 32-39: they disagree, so check act 4/5 on the live character first.
- Tests (hand-built fixtures, no real save): `go test ./internal/cheat/...`; round trip, neighbour difficulties untouched, item section byte-identical,
  checksum valid, refusals leave the file unchanged.

Live results (LIVEQ, 2026-10-28, Multi on i5, ledger reason "V-quests"; the tool backed the save up each time, Multi.d2s restored byte-identical afterwards). Tags: V live unless noted.
- Step 2: `done` (0x8001) shows "You completed this quest in a previous game" in the quest log (q); `started` 0x000c shows Sisters' text "Look for Blood Raven..."; the game rewrote the words on Save and Exit: 0x8001 -> 0x9001 (it ORs 0x1000 once the log was viewed). `saves.Done` (bit 0 or 15) is right.
- Step 3: raw 0x0001, 0x1001 and 0x0003 all show as completed (so bit 0 alone = complete; 0x1001 and 0x9003 on the next save), 0x0010 alone shows Search for Cain as active ("Find and rescue Deckard Cain"). Quest word order of act I confirmed: Tools of the Trade = word 3, Cain = word 4 (the log draws Cain third, Tools fifth).
- Step 4: `waypoints all` gives the nine act-1 rows; `none` gives "No other waypoints activated" with the town row still there (and the game writes the town bit back); `1+3+8` lights exactly Cold Plains, Dark Wood and Catacombs Level 2, so bit order = panel order for bits 0-8. All 39 bits are accepted; tabs for acts 2-5 not opened (U; see WPCHEAT below).
- Step 5: `act Multi 2/3/4/5` loaded Lut Gholein, Kurast Docks, area 103 (Pandemonium Fortress) and area 109 (Harrogath) (area read by navlive state, screenshots of acts 2, 3). `act` needs no expansion edit (Multi already had the bit). The game keeps the 0x8001 act words and the difficulty low bits. `act ... nightmare/1` alone does NOT select Nightmare: the game showed the difficulty screen only because progression was 5 (unlock-difficulty, or the game's own rewrite after word 40 completed), it offered Normal/Nightmare, and the choice (oog `start Multi nightmare`, click 400,323) wrote difficulty bytes 00 80 00 and loaded Rogue Encampment. With progression 0 there is no dialog (V-08).
- The Act V quest-log tab appears only once word 27 (Terror's End) is complete. Quest word positions of acts 4 and 5: words 25 and 37, 40 verified, act 5 is 35-40 (docs/saves.md; saves table fixed). A completed word 40 made the game set progression 5 and the word 0x1401.
- Step 6/7 done; read dumps with `d2cheat qdump FILE` (new, offline).

Live check list (needs the game, run by the main agent on a free instance, character Multi or another throwaway; game at the main menu):
1. `tools/live.sh multi oog leave`; copy `Multi.d2s` aside (the tool also backs up to /mnt/cold/d2lod/backups).
2. `... d2cheat --reason "V-quests" quest set Multi 1.den_of_evil:done 1.sisters_burial_grounds:started`, then `oog "start Multi"`, open the quest log (`q`):
   Den of Evil shown as complete, Sisters started? Record the word the game then writes: Save and Exit and `d2cheat savedump`-style dump of `Quests(0)`
   (compare 0x8001 with what the game itself writes after a real completion; flip saves' `Done` rule if needed). A hand-edited word that the game
   rewrites differently proves the bit meaning.
3. Repeat with `done` replaced by raw `0x0001`, `0x1001`, `0x0003` to see which values the quest log accepts.
4. `waypoints Multi all` then start: open the waypoint panel in town, all 9 (act 1) tabs-worth available? `waypoints Multi none` then confirm only the town works.
5. `act Multi 2`: does the game load in the Lut Gholein town? Then `act Multi 5` (needs the expansion bit, set on Multi). Then `nightmare/1`: is the
   difficulty selected, or does the game ignore it unless unlocked? Same with `unlock-difficulty Multi nightmare` first: does the difficulty dialog offer Nightmare?
6. For each Save and Exit afterwards, dump words/waypoints/difficulty bytes and compare with what was written; the game rewrites what it does not accept.
7. Restore `Multi.d2s` from the aside copy; the ledger holds one line per attempt. Flip tags in docs/saves.md and progression/verification.md (V-01, V-03, V-08).

WPCHEAT (2026-10-29, Navi on i1, ledger reasons "WPCHEAT: ..."): `waypoints Navi all` was NOT broken. V live: the save held all 39 bits in the right
order (block = `02 01` + bit field, bits 0-38 act I..V), the game loaded them, and the Act II rows all worked (clicking Canyon of the Magi (46) in the
panel travelled there; Far Oasis, Halls of the Dead 2 too). What looked like "only Lut Gholein" was the panel drawing only the tabs of reached acts:
with no act-done words it showed tabs I and II, and an Act II town screenshot cannot show rows of acts 3-5. V: the tab of act N appears once the
act-done words of acts 1..N-1 are set (tabs I-IV after `act1_done/act2_done/act3_done` = words 7, 15, 23); tab V additionally needed one or more of
izual (25), hellforge (26), terrors_end (27), act5_intro (32) (terrors_end alone did not show it; which of the others is the trigger is U, all are set).
Fix: every `wp` edit now also marks those words done (`openTabs`, same 0x8001 as `act`), per difficulty edited. This is a quest-state side effect of
`waypoints`, recorded in the ledger quest snapshot. Test: `go test ./internal/cheat/` (TestSaveQuestsWaypointsAct).

## Data/rate cheats (`dataoverlay`)
Feasible and verified for one table. The game started with `-direct -txt` reads loose `Data\global\excel\*.txt` from its own folder instead of the
MPQs. Every instance already has its own hardlinked game folder (`~/Games/d2lod-inst/iN/proton/pfx/drive_c/Program Files (x86)/Diablo II`),
so an overlay stays with that instance; i0's folder (the clone source) is refused. Steps VERIFIED LIVE on i5:
1. `d2cheat --reason ... overlay exppct=20` extracts experience.txt from the MPQs (patch_d2, d2exp, d2data priority), scales the thresholds, writes the overlay.
2. Start the instance with `D2_ARGS="-w -ns -direct -txt"` (by hand as in docs/multi.md "Debug-by-hand"; `tools/game.sh start` does not pass extra args).
3. Result: the in-memory experience table of the new process held 2835 (14175 x 20 %) and no longer 14175 (`d2cheat seq 2835 2835 2835 2835`).
Notes: with `-txt` the game also writes every compiled `.bin` into the Data folder (about 4.5 MB), so undo is `d2cheat overlay clear=1` (deletes the folder),
and without `-txt` a left-over folder with `-direct` would read stale `.bin`s. A restart is needed. Presets: `xprate=N` (monstats Exp, Exp(N), Exp(H) times N;
columns exist, 1608 cells; effect UNVERIFIED, same mechanism), `nodrop=0` (treasureclassex NoDrop; 300 cells; UNVERIFIED effect), `exppct=P`. Muting stays
(`-w -ns` is kept). The overlay is ledgered (kind `dataoverlay`, instance, no character).

## Usage for tests
    tools/live.sh multi d2cheat --actor multi --reason "need a level 25 sorc for skill tests" ...
    # level 25 quickly on a loaded character that already has exp/points (memory, instant)
    d2cheat --reason "..." level 25
    # fresh character (no exp/gold/points yet): save edit, then reload (about 25 s)
    tools/live.sh multi oog leave
    tools/live.sh multi d2cheat --reason "..." save Multi level=25 gold=50000 waypoints=1
    tools/live.sh multi oog "start Multi"
    d2cheat --reason "..." heal | god on --for 5m | gold 100000 | stats str=60 vit=50 | setstat stat=67 value=200
    d2cheat show            # both player copies, no ledger entry
    d2cheat ledger --since 2h
Save-edit parameters are `key=value` after the file/character name. Run `tools/live.sh multi oog leave` before, and restart with `oog "start Multi"` after.

## Bot runner integration (sketch, not wired)
`cheat.Gate{Allow, Env, Actor}`: the runner would add a `--allow-cheats` flag, default false in EVERY mode and never set by a production profile;
`gate.Do(ctx, kind, params, reason)` returns `ErrCheatsDisabled` otherwise, `gate.God(ctx, reason)` runs invincibility in a goroutine and returns `stop()`.
The runner marks its run start and asks `cheat.Cheated(start)` when it writes statistics (death rate, XP/hour) to flag or exclude cheated runs.

## Ideas: cheats as meta rewards (nothing implemented)
- Earned charges: a milestone (boss kill, level, streak) grants one "cheat token" (heal-to-full, +gold, a waypoint, 10 minutes of speed); spending one calls `Gate.Do`
  and the ledger entry becomes the receipt shown on the display.
- Banked comfort: lasting but small (a stash gold stipend per day, magic find +20 for a session) bought with tokens; the same ledger lets the display show active cheats.
- Honest scoreboards: every statistic carries a "cheated" flag from the ledger; the leaderboard shows clean and assisted runs separately.
- Safety valves: allowlist and offline checks stay on; rewards expire at Save and Exit unless they are save edits; save edits always back up first.

## Dead ends and hazards
- No static pointer to the server unit; the scan costs about a second per call (the address is not cached across calls, units move on every game start).
- Server-side attribute writes alone do not reach the client; client-only writes are overwritten or never reach the save: always both.
- `cmd/combatlive` uses `input.FindGamePID` (no instance filter): it attached to i0. Not changed (not mine); other agents should avoid it on multi instances.
- `internal/mem` `TestProcInstance` is flaky in this sandbox (reads the child's environ too early); unrelated to the writer.


## God-mode safety (godsafe)

Problem (V, docs/act2.md "Live results 3"): a killed `act2live god` job left max life 2097151 in the character, and the game's own Save and Exit writes the server unit's max life into the .d2s.

- Before the first write god records the originals in an undo marker `god-pending-<instance>.json` next to the ledger (written before the buffer, removed only after the restore is read back from memory). Restore runs in a defer: normal end, timeout, ctx cancel, panic. `d2cheat` and `act2live` now turn SIGINT and SIGTERM into ctx cancel; act2live also stops god at once on the signal. Tested with mem.Fake (cancel, panic, leftover marker). U: not run against the real game.
- SIGKILL cannot be caught. The marker is the repair: the next `god on` first restores from it (message "restored leftover god state"), or run `d2cheat --reason "..." restore`. Marker for another character is kept and ignored.
- Inflated max life (above 6000, `godSaneMax`, U for real gear) with no marker: god and restore refuse (`ErrInflatedLife`), so the buffer is never taken as the original; force a value with `d2cheat restore --maxlife N`.
- Before any save (Save and Exit, quit, restart) the controlling code calls `Env.GuardSave(ctx, reason)` / `Gate.GuardSave`: it restores from the marker while the game is attached (ledger kind `godrestore`). The bot should call it before its save-and-exit hook; not wired into d2bot here. A killed job should be followed by `d2cheat restore` before the game is closed.
- Not covered: restore only fixes max life and life; mana/stamina are refilled to their own maxima and are not inflated.
