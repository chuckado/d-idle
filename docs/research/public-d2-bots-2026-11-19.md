# Public Diablo II bots and automation: inventory for the design review (2026-11-19)

> status: research input for the design review (coordinator brief 2026-11-19). Docs-only, no code vendored. Nothing here is a decision.

Tags: **V** = read by me from the primary source this run (a shallow `git clone` read in a scratch dir, never executed, or the GitHub API for repo metadata, or a fetched page). **S** = web-search snippet or secondary page, not read in full. **U** = unverified or from memory. Target is classic 1.14d; D2R-only facts are marked **[D2R]**. Dates are "last push" from the GitHub API on 2026-11-19 (V) unless noted; a stale date means abandoned, a recent one only means someone touched the repo.

Licence rule for us (GPL-3.0-or-later): a repo with no licence file is "all rights reserved" by default, so we read it for ideas and data formats only. Even GPL-3.0 and MIT code is not copied: we implement from the ideas. Nothing was downloaded into this repo.

Search method (V): `gh search repos` over about 60 queries (bot, d2bs, kolbot, d2r, koolo, botty, mapassist, clientless, pvpgn, d2gs, nip, autohotkey, autoit, rust, go, rl, llm), direct `gh api repos/...` metadata for about 90 named repos, shallow clones of about 40 of them, plus about 12 web searches. Not covered (see section 6): Blizzhackers forum, d2jsp, MPGH, Reddit, YouTube, Discord, GitLab, Gitee/Chinese-language sites, archived Bnet-era sites. Those hold more projects than I could reach, so the count below is a floor.

## 1. Index table

Columns: name, URL (github.com/ prefix omitted), last push, licence (as declared), versions, classes with full scripts, approach. "none" = no licence file.

### 1a. D2BS / kolbot family (injected DLL + JS, 1.1x LoD)

| # | Name | Repo | Last push | Licence | Versions | Classes | Approach |
|---|---|---|---|---|---|---|---|
| 1 | D2BS (noah- line, SpiderMonkey) | noah-/d2bs | 2021-04 | none | 1.13c/1.13d/1.14d [S for the exact set] | n/a (engine) | DLL injected into Game.exe, JS API (V) |
| 2 | D2BS older/fork copies | ds-hwang/d2bs, sakuracz/D2BS (GPL-2.0, SVN export), 3vlig/d2bs, binaryatrocity/d2bs, bisko/d2bs, theBGuy/d2bs | 2011-2022 | mostly none | 1.1x | n/a | same lineage, history only (V metadata) |
| 3 | d2bsng (V8 rewrite) | ResurrectedTrader/d2bsng | 2026-11-14 | MIT | 1.14d only | n/a (engine) | injected DLL, V8, typed game layer on D2MOO, simulator (V) |
| 4 | D2BS-QJS | ejt1/d2bs-qjs | 2024-03 | none | 1.13c [S] | n/a | QuickJS rewrite of D2BS (V README head) |
| 5 | D2Boot | io-mars/d2boot | 2025-12 | none | 1.13d | n/a | QuickJS port, kolbot scripts (V) |
| 6 | D2Bot# manager + kolbot (original) | kolton/d2bot-with-kolbot | 2021-04 | none | 1.13c/d | all 7 | C# manager + JS libs (V metadata) |
| 7 | kolbot (maintained trunk) | blizzhackers/kolbot | 2026-11-16 | none | 1.13d/1.14d [V: installer targets legacy D2] | all 7 | d2bs + D2Bot# + libs, NIP pickit (V) |
| 8 | Kolbot-SoloPlay | blizzhackers/kolbot-SoloPlay | 2026-11-17 | GPL-3.0 | LoD classic and expansion, English client | 7 classes, 30 named builds (V) | auto-leveling 1 to 99 on kolbot (V) |
| 9 | SoloLeveling (predecessor of 8) | isid0re/SoloLeveling | 2022-08 | GPL-3.0 | LoD | multi-class | original leveling script (V metadata) |
| 10 | D2BotNG | ResurrectedTrader/D2BotNG | 2026-11-15 | none | 1.14d etc. | n/a (manager) | .NET 10 + React web manager, protobuf JSON data (V) |
| 11 | jaenster-kolbot | jaenster/jaenster-kolbot | 2021-02 | none | 1.13 | multi | "fully automatic skills" kolbot rewrite (V) |
| 12 | kolbot-sonic (+ noah fork, mf022 fork) | SetupSonic/d2bot-with-kolbot-sonic | 2019-12 | none | 1.13 | sorc auto-level | kolbot variant (V metadata) |
| 13 | Horde | Adpist/horde | 2021-07 | none | 1.13 | multi | stable kolbot bundle, frozen (V) |
| 14 | KolbotFullAutosmurf | NICKCHEESE/KolbotFullAutosmurf | 2020-09 | none | 1.13, TCP/IP | 8 smurf builds | profile pack for local TCP/IP games (V) |
| 15 | d2-qolbot | whipowill/d2-qolbot | 2023-03 (archived) | none | 1.13d + PlugY | follower | kolbot fork: Follower/Quester, auto stat/skill, LAN (V) |
| 16 | PlugY4Kolbot | theBGuy/PlugY4Kolbot | 2026-11-15 | GPL-3.0 | 1.14d | n/a | PlugY fork hardened for many parallel kolbot games (V) |
| 17 | LimeDrop | blizzhackers/limedrop | 2026-07 | NOASSERTION | n/a | n/a | web UI for item mule/drop queue (V) |
| 18 | discord-d2bs-bot, D2ChannelBot, D2BS_HTML_Log_Viewer | theBGuy/..., fatboi615/..., nitrotech7/... | 2020-2026 | MIT/none/GPL-3.0 | n/a | n/a | side tools for kolbot (V metadata) |
| 19 | vsnip-check, pickits repo, nipcheck/nipper | theBGuy/vsnip-check, blizzhackers/pickits | 2026 | none/NOASSERTION | n/a | n/a | NIP tooling and pickit corpus (V) |
| 20 | UpdatedMapHack | theBGuy/UpdatedMapHack | 2022-04 (archived) | none | 1.13 | n/a | revamped kolton maphack script (V metadata) |

### 1b. D2NT / ETAL / older injected bots (1.09 to 1.13)

| # | Name | Repo | Last push | Licence | Versions | Approach |
|---|---|---|---|---|---|---|
| 21 | D2NT scripts / D2NT-1.13c / BM-Mod ETAL | h3fang/scripts, mf022/D2NT-1.13c, bouletmarc/D2NT-BM-Mod-Etal-Bot, xEquinaux/d2nt (BSD-2) | 2015-2024 | none/BSD-2 | 1.11b to 1.13c | NT script bot on D2NT.dll hook; origin of NIP (V metadata + NipGuide) |
| 22 | D2Hackit (+ modules, beltfiller) | toneomgomg-archived-projects/D2HackIt, inrg/D2HackIt-backup (Unlicense), sumishisan/d2hackit-beltfiller (GPL-3.0) | 2015-2023 | Unlicense/GPL-3.0 | 1.09 to 1.11 [U] | C++ plugin host loaded into the game (V metadata, U on era) |
| 23 | d2au3 (AutoIt framework, mmbot imitation) | TehKiwiz/d2au3 | 2012-04 | GPL-3.0 | 1.1x | AutoIt, memory + input; mmbot "sequence" model (V README) |
| 24 | diablo2-autohotkey-lib | seanfisk/diablo2-autohotkey-lib | 2015-12 | none | LoD | AHK macros (keybind sync, weapon-set skill pairing), not a bot (V) |
| 25 | d2lua | garrett-he/d2lua | (404 for me) | ? | ? | Lua scripting host, inspired by WoW and the "d2jsp bot" (S only) |
| 26 | Awesomo / redvex plugin | sourceforge awesomo | ? | ? | Bnet era | Battle.net bot plugin (S only) |

### 1c. Clientless and server-side (packet level)

| # | Name | Repo | Last push | Licence | Versions | Approach |
|---|---|---|---|---|---|---|
| 27 | OmegaBot | dkuwahara/OmegaBot | 2019-06 | none | 1.09-era | C# clientless Bnet/D2GS bot (V metadata) |
| 28 | D2NG (renamed Diablo2Clientless1.09) | jcageman/Diablo2Clientless1.09, dkuwahara/D2NG | 2026-11-11 | GPL-3.0 | 1.09d | C# .NET 10 clientless: Mephisto/Travincal/Pindle/Cows/CS/Baal bots, NIP pickup, gamble, mule (V) |
| 29 | AutoTathamet | MephisTools/AutoTathamet (+ noah-, rom1504 forks) | 2019-07 | MIT | "1.14" option | Node.js clientless: chat, follow, pickup (V) |
| 30 | clientless (Zig) | jaenster/d2-clientless | 2026-10-26 | none | 1.14d | protocol test harness, explicitly not a bot (V) |
| 31 | d2-dedicated-server | jaenster/d2-dedicated-server | 2026-10-29 | MIT | 1.06b to 1.13c native, 1.14d under Wine | clean-room Zig realm + game server (V metadata) |
| 32 | PvPGN / D2GS | pvpgn/pvpgn-server (GPL-2.0), pvpgn/d2gs109, tesseract2048/d2gs (1.13c), Zakamurite/D2GS, espenmjos/d2gs + pvpgn (MIT, Docker, D2GS 1.14d on Wine), D2GSMonitor | 2011-2026 | mixed | 1.09 to 1.14d | private realms that bots target; not bots (V metadata) |
| 33 | blacha/diablo2 | blacha/diablo2 | 2022-07 | MIT | Classic packets, [D2R] map + memory | TS packet sniffer/huffman/MPQ/bin parsers, map API (V) |

### 1d. D2R-era (2021-2026) [D2R unless noted]

| # | Name | Repo | Last push | Licence | Approach |
|---|---|---|---|---|---|
| 34 | Botty (original aeon0/botty) | johannes-do/botty (archived copy; upstream gone, 404) | 2022-07 | MIT | Python + OpenCV pixel bot, 720p, English (V) |
| 35 | Botty forks | ggodev, seekerlee, navy91, itsmeCin, nalanminmoon (zh-CN, 2560x1080) | 2021-2026 | MIT/none | forks of 34 (V metadata) |
| 36 | Koolo (original hectorgimenez, gone) | dulingzhi/koolo (2023-11), xg6636/koolo (2025-01) | 2023-2025 | MIT | Go, reads D2R memory, sends input; uses LoD 1.13c files for map gen (V) |
| 37 | Koolo "DiobyteVersion" (fork of kwader2k/koolo, also gone) | vietdungdev/Koolo-DiobyteVersion | 2026-02 | (see section 3) | Go, mouse humanizer, leveling builds, obfuscated build (V) |
| 38 | Safe-d2r-bot | Botty-d2r/Safe-d2r-bot | 2024-12 | MIT | Koolo-derived, "safety" marketing (V) |
| 39 | autod2r | huangyiyang89/autod2r | 2021-10 | MIT | Python D2R bot, README is two words (V) |
| 40 | Diablo2Bot (Rust CV bot) | kasparses/Diablo2Bot | 2025-02 | GPL-3.0 | Rust, screenshots only, no memory; 1.14b LoD under Linux/Wine capable (V) |
| 41 | DIABLO_2_BOT (C++/OpenCV minimap) | KAfus1231/DIABLO_2_BOT | 2026-01 | MIT per README badge, none on API | minimap navigation, experimental (V README head) |
| 42 | MapAssist (OneXDeveloper) | OneXDeveloper/MapAssist (gone, 404); forks xsyww, noxiaz (GPL-3.0) | 2021-2022 | GPL-3.0 | C# memory-read overlay with map-server data, not a bot (V metadata; the original README not read) |
| 43 | d2r-mapview | joffreybesos/d2r-mapview (+ camy011 fork) | 2022-12 (archived) | AGPL-3.0 | AutoHotkey overlay reading memory + map server; README records a ban wave (V) |
| 44 | D2RMH | tw1st40/D2RMH | 2022-05 | MIT | C++ read-only map hack using D2 1.13c data (V) |
| 45 | D2R-Bots (collection) | goblindevelopment/D2R-Bots | 2024-01 | none | rar binaries of third-party bots, "memory based and undetected" claim (V; not opened) |
| 46 | d2rmab "cracked source" | d2rbots/diablo-2-resurrected-bots | 2023-02 | Unlicense label | AutoIt dump presented as leaked paid-bot source (V header); treat as stolen code, never read further |
| 47 | SigmaDrift | ck0i/SigmaDrift | 2026-10 | MIT | mouse movement humanizer used by 37 (V) |
| 48 | Terror-zone/loot loggers | Timmynat0r/koolo_item_logger and similar | 2025 | none | side tools (V metadata) |

### 1e. Other game engines / research agents (teach us something)

| # | Name | Repo | Last push | Licence | Approach |
|---|---|---|---|---|---|
| 49 | DevilutionX-AI | rouming/DevilutionX-AI | 2026-11-05 | NOASSERTION | Diablo 1: RL model + algorithmic supervisor, shared-memory state, headless (V) |
| 50 | DAPI | NiteKat/DAPI | 2026-02 | none | Diablo 1 bot API, protobuf server/client, "human information limits" (V) |
| 51 | D2MOO | ThePhrozenKeep/D2MOO | 2026-11-15 | MIT | clean reimplementation of D2 1.14d logic (V metadata; d2bsng builds on it) |
| 52 | diablo_edit | daidodo/diablo_edit | 2024-09 | MIT | save editor, not a bot (V metadata) |
| 53 | (no entry) | Diablo II LLM or RL agents | - | - | none found; web search (S) only surfaced the Diablo 1 projects above |

Count: 53 rows, about 75 distinct repos once forks are folded in (about 46 distinct projects). Appendix A lists non-D2 ARPG bots (U).

## 2. The D2BS / kolbot family in depth

### 2.1 D2BS (core)
- **How it reads and acts (V, d2bsng README + repo layout):** a DLL injected into Game.exe exposes a JS API (`getUnit`, `getPath`, `clickMap`, `move` ...) that calls into the client's own functions and reads its structures. Scripts run in game threads with a per-frame tick. It drives the real client; there is no separate memory-read process.
- **Honest limits (V):** every kolbot fork README carries the banner "D2BS is not safe from detection" (jaenster-kolbot, kolbot-sonic) and "meant for offline or private servers" (kolbot, Horde). The family targets 1.13c/d with 1.14d support in the trunk; D2R is unsupported (V: SoloPlay FAQ in the search snippet and kolbot README say legacy D2 only).
- **Manager (V):** D2Bot# (C#) starts one Game.exe per profile, handles login, CD-key lists, game names, restart on crash, mule logging, and schedules. D2BotNG is a 2026 rewrite with a web UI and protobuf-JSON data files (profiles, frameworks with version bundles, patches, key lists, character snapshots in SQLite).
- **Notable engineering (V):** `docs/lod_timing.md` in d2bsng documents a 1.14d single-player clock rollover bug at 2^31 ms and a two-byte fix; `docs/belt_pickup_race.md`, `lod_read_pacing.md` and `game_thread_safety.md` document races a native in-process bot hits. These are bug notes worth reading before we assume memory reads are consistent across frames.

### 2.2 d2bsng and its simulator (most relevant to our d2d backend)
- **URL/licence/activity (V):** ResurrectedTrader/d2bsng, MIT, pushed 2026-11-14, README says it is "vibe coded with Claude".
- **Architecture (V):** `contract` (version-agnostic game interface: Unit, Room, Level, Party, Control) with a `lod114d` backend (memory reads, function calls, hooks, offsets) and a V8 frontend; frontend and backend never reference each other. Game logic types come from D2MOO. Navigation (A*, level-exit finder) lives in the game-agnostic runtime and is tested against a fake game layer (`js_tests` with doctest).
- **Simulator (V, README + `docs/host_hooks.md`):** `d2bs-sim.exe` maps Game.exe into its own console process, runs the game, scripts and clock deterministically and faster than real time, runs kolbot by default, and can submit runs to a public leaderboard. `host_hooks.md` states about 157x to 175x real time on a fresh sorceress (3 x 600 virtual seconds) and lists the link-time hooks a host must implement (Sleep, clock, `RunFrame`, `OnNativeCall`, `ObserveNativeResult` that hashes every binding result for determinism). What the leaderboard scores is not documented in the clone and the site is JS-rendered, so I could not read it (U).
- **Why it matters:** it is the closest public precedent to "bot runs against a decomposed game engine, headless, faster than real time" and its hook list is a ready checklist for a deterministic harness. It runs the real game binary, not a reimplementation.
- **Reusable ideas (not code):** deterministic host contract; per-binding result hashing for replay diffs; speed scaling as a config knob; version-agnostic game interface with a fake for unit tests; per-script CPU/heap diagnostics; PlugY multi-tab stash abstraction (`StashTab`).
- **Incompatible:** MIT is permissive, so code could legally be reused, but it is C++/V8 and 32-bit injected; we implement from the design.

### 2.3 kolbot libraries
- **Layout (V):** `libs/core/*` (Attack, Pather, Pickit, Town, Storage, Cubing, Runewords, Precast, NTItemParser, Config, Loader), `libs/scripts/*` (about 85 run scripts (directory listing)), `libs/config/<Class>.js` (7 class config files plus Templates), `pickit/*.nip`.
- **Scripts (V, directory listing):** Andariel, Mephisto, Diablo, Baal, Duriel, Countess, Pindleskin, Nihlathak, Tombs, Cows, Travincal, Summoner, Radament, Izual, Tristram, KurastTemples, Eldritch, ShopBot, Gamble, OrgTorch (uber torch), KillDclone, plus helpers (Rusher, Rushee, MFHelper, Wakka, ControlBot, AutoBaal, AutoChaos). Scripts compose into a per-profile run list.
- **Config model (V, `libs/config/Sorceress.js`):** one big JS file per class, edited by hand, values below. This is "profile = a JS object" with comments as documentation.
  - Safety: `Config.LifeChicken = 30`, `ManaChicken = 0`, `MercChicken = 0`, `TownHP/TownMP = 0`, `UseHP = 75`, `MercWatch` (instant merc revive), `Dodge/DodgeRange/DodgeHP`.
  - Potions: `BeltColumn = ["hp","hp","mp","rv"]`, `MinColumn = [3,3,3,0]`, `HPBuffer`, `MPBuffer`.
  - Town: `StashGold = 100000`, `RepairPercent = 40`, `CainID.*`, `FieldID.*` (id in the field with packets), `Gamble` with `GambleGoldStart/Stop` and item list, `Inventory[4][10]` lock grid.
  - Combat: `AttackSkill[0..6]` = preattack, boss primary timed/untimed, other primary timed/untimed, secondary if immune; `SkipImmune`; `MFSwitchPercent`; `TeleSwitch`; `PickRange = 40`.
  - Crafting/runewords: `Cubing`, `MakeRunewords`, `Config.Runewords.push([Runeword.X, "base", Roll.Eth])`.
- **Auto stat/skill (V):** three layers. `Config.AutoSkill.Build = [[skillId, count, satisfy], ...]` (count = target hard points; `satisfy` = stop here until the character level allows the next goal; prerequisites are learned automatically). `Config.AutoStat.Build = [[stat, target], ...]` with `Save`, `BlockChance` (for shield builds) and `UseBulk`. `AutoBuild` (older) uses a per-class template file with level tiers. Example builds shipped as text (Blizzard sorc, Fire sorc). These tables are exactly the "skill order table" format we asked about: a priority list of (skill, target level) with a gating boolean.

### 2.4 NIP pickit grammar (V, blizzhackers/pickits NipGuide.md and kolton.nip)
- A rule is `{properties} # {stats} # {maxquantity}`; properties before the first `#`, stats after.
- Property keys: `[type]`, `[name]`, `[class]` (normal/exceptional/elite), `[quality]` (grey .. unique/set/crafted), `[flag]` (ethereal, identified, runeword), `[level]` (ilvl), `[prefix]`, `[suffix]`. Stat keys: any alias from `NTItemAlias.dbl` (e.g. `[fcr]`, `[frw]`, `[fireresist]`, `[itemaddclassskills]`, `[sockets]`, `[defense]`) or a numeric stat id, and `[description]` for matching tooltip text.
- Operators `== != > >= < <= && || ( ) + - * /` and `//` comments. Sums are allowed (`[strength]+[dexterity] >= 30`).
- Semantics that matter: a rule is a predicate over an item unit; unidentified items are matched on what is visible, and a rule that needs hidden stats makes the bot keep the item unidentified or identify then re-test (V: Koolo README states the same two-stage behaviour). Files are loaded per character from the class config.
- Extensions seen in the wild (V): Botty adds `[allres]`, `[idname]` for named uniques, raw poison damage min/max instead of computed, and does not support `[maxquantity]`. SoloPlay adds a **third `#` section `[tier] == N`** (numeric item rank, or `tierscore(item, N)` computed by a function) for auto-equip upgrades, in `finalGear` string lists inside build files. Koolo reuses NIP files unchanged and refuses to start on a parse error.
- Licence angle: the grammar and alias names are a data format and community convention; implementing a parser from the public guide is clean. The shipped `.nip` corpora (kolton.nip etc.) carry no licence; we read them for coverage ideas only.

### 2.5 Kolbot-SoloPlay (V: clone read)
- **What it is:** an auto-leveling layer on kolbot; the user supplies a profile name encoding class and build and presses start. Claims Hell within about 24 h (48 h for Barbarian), plays through to the end, hires a class-specific merc, auto-equips gear and charms by scoring, makes CTA and Spirit shield, kills Diablo Clone in expansion. 
- **Classes and builds (V README):** Amazon (Javazon, Witchyzon), Sorceress (Cold, Blizzballer, Meteorb, Blova, Lightning), Necromancer (Poison, Bone, Summon, Griefmancer), Paladin (Hammerdin, Smiter, Auradin, Zealer, Torchadin, Classicauradin, Hammershock, Sancdreamer), Barbarian (Whirlwind, Immortalwhirl, Frenzy, Uberconc, Singer), Druid (Wind, Elemental, Plaguewolf, Wolf), Assassin (Trapsin, Whirlsin). Some are expansion only.
- **Build definition file (V, `BuildFiles/paladin/paladin.HammerdinBuild.js` and `...LevelingBuild.js`):** one JS module per build with: `wantedskills`, `usefulskills`, `precastSkills`, `wantedMerc`, `skills` (ordered `[skill, hardPoints]` list), `stats` (ordered `[stat, target]` with `"all"` as the sink; separate classic vs expansion tables), `charms` (named charm slots with a max count and a predicate on stats), `finalGear` (NIP strings with `[tier]`), `AutoBuildTemplate[levelTier].Update()` (overrides attack skills, `BeltColumn`, `SkipImmune` per tier), `respec()` (condition to respec, e.g. Enigma owned in expansion, clvl 75 and Diablo done in classic) and `active()`. The Hammerdin file cites a 2010s forum build guide as its source; the leveling file reaches Blessed Hammer 20 before pumping Concentration/Vigor, then Blessed Aim and Holy Shield last.
- **Run order (V, `Tools/SoloIndex.js`):** a fixed ordered list with `preReq`, `skipIf`, `shouldRun` predicates and a `goldScripts` fallback list (`bishibosh, tristram, treehead, countess, lowerkurast`). Order: Act 1 corpsefire, mausoleum, den, bishibosh, bloodraven, tristram, treehead, countess, smith, pits, jail, boneash, andariel, a1chests, cows; Act 2 cube, radament, creepingfeature, beetleburst, amulet, summoner, fireeye, maggotlair, tombs, ancienttunnels, staff, duriel; Act 3 lamessen, templeruns, lowerkurast, eye, heart, brain, travincal, mephisto; Act 4 izual, hellforge, river, hephasto, diablo; Act 5 shenk, savebarby, anya, pindle, nith, ancients, baal, a5chests. This is the community's quest-complete leveling skeleton, per-run gated by act access and quest state.
- **Other modules (V):** `DynamicTiers.js` (item scoring with weights per stat, separate weights for the merc), `CharmEquip.js`, `Mercenary.js` (merc choice per build), `TownChicken` worker, `SoloEvents`, `Quest.js`, `PatherOverrides`, `StorageOverrides`, per-class attack overrides, `Tools/Tracker.js`, `Overlay.js`, `CharData.js` (persisted per-character JSON).
- **Licence:** GPL-3.0, same as ours, but we still do not copy; the README and build files are the data we use as reference.

### 2.6 Other kolbot descendants (V unless noted)
- **jaenster-kolbot:** "fully automatic skills": enumerate known skills, estimate damage on the target and neighbours, pick the best each cast; single configuration file; faster weapon switching. Idea: damage-estimating skill choice instead of per-class attack tables. No licence.
- **KolbotFullAutosmurf:** a pack of eight Level-1 character saves plus profiles for TCP/IP LAN play, multiple builds (Blizzard, Fire, Hammer, Nova, Java, Trap, Summon, Barb). Idea: fully local multi-bot smurf pipeline; also the "level 1 char reset files" approach (replace the save, rerun).
- **d2-qolbot:** adds Follower/Quester behaviour, gallows-humour chat lines, waypoint capture, quest item pickup, auto-allocation templates, custom inventory sizes; "never replaces Unique/Set/Runeword gear" (auto-equip policy). Archived.
- **PlugY4Kolbot:** PlugY (infinite and shared stash, disables Bnet, enables ladder runewords and uber quest offline) hardened for many parallel games. Relevance: infinite stash removes the "stash full" problem for idle bots; we should check our assumption on stash size.

## 3. D2NT, ETAL and the older bots

- **D2NT / ETAL / NTBot (V metadata; U history):** `h3fang/scripts` contains `NTBot.ntj`, `NTLeader`, `NTLeecher`, `NTMap`, `NTHelper`. The NT bots predate D2BS, used a hook DLL (`D2NT.dll`) per game version, and are where the "NIP" pickit syntax originates (the NipGuide refers to D2NT pickit errors and an `NTItemAlias.ntl`). xEquinaux/d2nt is a 2024 BSD-2 service fork for 1.13 single player only; its README says 1.14d would need another DLL (V).
- **D2Hackit (V metadata, U for era):** a C++ plugin host that loaded modules into the game; archived by its author in 2015 with an Unlicense. Spawned beltfillers, chat tools and the first scripted bots. For us: history only.
- **d2au3 (V):** AutoIt re-creation of the "mmbot" sequence system; open issue that AutoIt bots read memory and send input without a DLL.
- **Awesomo / OmegaBot / BNLS tools (S/V):** Bnet-era clientless bots speak BNCS/MCP/D2GS. OmegaBot (C#, 2011-2019) became D2NG (see 4).

## 4. Clientless and server-side

- **D2NG (V):** jcageman's continuation, GPL-3.0, .NET 10, pushed 2026-11-11, targets **1.09d** only. README (V): bots for Mephisto, Travincal, Pindle, Cows, Chaos Sanctuary, Baal; pathing via an external map API (`d2mapapi`); NIP-based keep rules; gambling; mule; three config files (`bot.json`, `mule.json`, `pickitconfig-*.json`). Chicken config keys visible in the README: `lifeChickenPercent 0.2`, `useHealthPotionPercent 0.9`, `useRejuvenationPercent 0.3`, `useManaPotionPercent 0.3`, `minPotionIntervalMs`. Value for us: a worked example of every bot loop being rebuilt from packets with a map service; it never needs the client at all.
- **AutoTathamet (V):** Node.js, MIT, 2019; chat, follow, pick up items; claims a `version: '1.14'` option; roadmap items (map, inventory) were never finished.
- **jaenster/d2-clientless (V):** a protocol test harness (Zig, no licence file) that logs in, joins and prints the packet stream; its README says "not a bot". It is the test client for jaenster/d2-dedicated-server (MIT), a clean-room Zig realm and game server for seven legacy versions (1.06b to 1.13c on the real DLLs, 1.14d via Wine, plus a wine-free native build). This is the nearest thing in public to a re-implemented server; relevant to our d2d backend as prior art and as a possible differential target (U on completeness).
- **PvPGN/D2GS (V metadata):** pvpgn-server is the standard private-realm stack (GPL-2.0, 2026-08). D2GS builds exist for 1.09d, 1.11b, 1.13c and one repo packages "D2GS 1.14d on Wine" in Docker (espenmjos). Bots targeting private realms use these; they are infrastructure, not bots.
- **blacha/diablo2 (V):** packet sniffer, Huffman, MPQ and .bin parsers, and a Docker map-generation API (`map` package) used by several D2R bots (Koolo asks for LoD 1.13c files for exactly this reason, V). MIT, last push 2022-07.

## 5. D2R-era bots [D2R]

Applicability to us is low (different binary and anti-cheat), but they show the current state of design and their feature lists are a good coverage checklist.

### 5.1 Botty (V: clone of the archived copy)
- Python + OpenCV pixel bot, MIT, last commit 2022-07-01; upstream repo (aeon0/botty) is 404 for me; the archived copy is by johannes-do, 570 stars. Reads the screen only, 720p, English only; requires hero positioned in Act 3/4/5 at start.
- **Structure (V):** `bot.py` state machine; `main.py` starts three threads: death monitor, health monitor, bot. `char/` classes (barbarian, necro, paladin, sorceress, trapsin) inherit `IChar`; `run/` scripts (pindle, trav, nihlathak, arcane, diablo, shenk/eld); `pather.py` follows prerecorded template "nodes" created with a node-creator tool (no map data); `template_finder.py` image matching; `health_manager.py` with belt row/column config and `chicken` percent; `shopper.py` for vendor refreshes.
- **BNIP (V):** NIP with extensions (section 2.4). Config in `params.ini`: belt rows and columns per potion type, chicken threshold, screenshots on every stuck/chicken/timeout.
- Idea for us: the strict three-thread supervision (bot, health, death); screenshots on every anomaly; the dev doc note that most modules run standalone for testing.

### 5.2 Koolo (V: three clones)
- Original repos (hectorgimenez, kwader2k) are 404; copies: dulingzhi (2023-11), xg6636 (2025-01), vietdungdev "DiobyteVersion" (2026-02). Licence MIT on the first two (V). The 404 of the upstreams is consistent with a takedown, but I have no source for the reason (U).
- **Reads memory and injects input (V README):** Go, reads D2R process memory (through the `d2go` package, also gone) and sends clicks and key events; resolves map data through a map helper that needs LoD 1.13c data (the Blacha map generator). Windowed 1280x720.
- **Features (V README):** Blizzard and Nova Sorceress, Hammerdin, FoH, Berserk Barb, Javazon, Winddruid in various stages; about 30 runs (Countess, Andariel, Ancient Tunnels, Summoner, Mephisto, Council, Pindle, Nihlathak, Tristram, Baal, Diablo, Cows, TZ and more); Discord/Telegram integration; companion (leader/follower) mode; NIP pickit; auto potion incl. merc; chicken; inventory slot locking; merc revive; CTA and class buffs; auto repair; skip immune; auto gambling; cubing; auto-leveling sorc and paladin ("WIP"); multi-window.
- **Config (V, `config/template/config.yaml` of the DiobyteVersion):** `health: healingPotionAt 75, mercHealingPotionAt 80, chickenAt 30, townChickenAt 0, mercChickenAt 10, beltColumns [healing, healing, mana, rejuvenation], healingPotionCount`. These are almost line-for-line kolbot's numbers (UseHP 75, LifeChicken 30). Packet casting options exist but ship disabled "for safety".
- **Chicken extras (V `internal/chicken`):** configurable chicken on curses (Amplify Damage, Decrepify, Lower Resist, Blood Mana) and on nearby monster auras (Fanaticism, Might, Conviction, Holy Fire) within 25 units. Good checklist item for a survival policy.
- **Leveling data (V):** `config/template/builds_leveling/{amazon,assassin,druid,necromancer,paladin,sorceress}.json` list NIP files per build; `internal/run/leveling_act1..5.go` encode the quest route; `sequences_leveling/default_sequence.json`.
- **Anti-detection (V fork README):** the fork adds a "SigmaDrift" lognormal mouse-trajectory humanizer (replacing teleporting the cursor), and builds with `garble` obfuscation. These matter on Bnet only; offline they are irrelevant to us.
- **Skills from keybindings (V):** the bot reads the in-game skill key bindings to cast, instead of keeping its own table; you must set the left skill (Glacial Spike, Blessed Hammer) yourself.

### 5.3 Rust and vision-only bots (V)
- **kasparses/Diablo2Bot (GPL-3.0, 2025-02):** Rust computer-vision bot with a stated rule: screenshots only, no memory. Per-character TOML profile (`character_class`, `primary_attack_skill`, faster-cast-rate per weapon set, left skill per weapon set, `zone_to_farm`, `min_gold_to_pickup`) and a TOML loot filter per item base name with booleans per quality (`grey, common, magic, rare, set, unique, rune`) and inventory footprint (`row_size`, `col_size`). It decodes DC6/DCC/DT1 sprites and MPQs in-repo to read the minimap and navigate. Linux/Wine instructions use `xdotool` and `wmctrl`. Requires 1.14b install. Cleanest public precedent for "build the bot as vision + Wine + xdotool". Last activity 2025-02, a single developer.
- **KAfus1231/DIABLO_2_BOT:** C++/OpenCV minimap navigation, experimental; README in Russian; Status "Experimental" (V).
- **DiaFlow (V):** AutoHotkey, pixel/image only, targets Median XL, last push 2020-12; scripts for rune/gem/unique pickup, auto teleport, chat info.

### 5.4 Map overlays (not bots, but the data layer) [D2R]
- **MapAssist (U details):** original repos gone (404). Forks (GPL-3.0) survive; per web snippets it reads D2R memory and renders map, monsters and immunities as an overlay, with an upstream map-server component. Not verified by reading.
- **d2r-mapview (V):** AutoHotkey, AGPL-3.0, archived 2022-12. README's first lines say Blizzard issued a ban wave targeting this tool for online use on 14 Oct 2022. Reads memory, renders a map overlay, with a separate map server.
- **D2RMH (V):** C++, MIT, 2022-05, "only reads process memory ... no injects, hooks or writes ... not guaranteed ban-free", uses D2 1.13c data files for map generation. Shows that the map-data approach (run the old-engine map generator offline) is how all D2R tools get walkable maps.
- **blacha/diablo2 `map` (V):** Docker map-generation API using the LoD DLLs under Wine; used by Koolo, D2NG-style bots. Relevant to us because the same trick (use the real engine's level generator through Wine) has already been shipped publicly for 1.13c.

### 5.5 Paid / underground D2R bots (S)
- Web searches (S) show community claims of widespread botting on D2R, infrequent ban waves, and credential-theft risk with paid bot services. I found no primary data. `goblindevelopment/D2R-Bots` is a binary dump (not opened); `d2rbots/diablo-2-resurrected-bots` carries a header pointing to a "cracked bot source" site and a spoofed "OneDrive" product string: leaked paid code and malware-shaped. Do not read or reuse.

## 6. Coverage gaps (what I could not reach)

- Blizzhackers forum (discourse), d2jsp, MPGH, Reddit r/diablo2 and r/D2R, Discord-linked repos, YouTube descriptions, archive.org, GitLab, Gitee (only seen as search hits: gitee.com/yohaa0/kolbot mirror), Chinese-language scene (nalanminmoon's zh-CN botty fork hints at one). The brief's named items I could not locate at all: "ClassicD2Bot", "d2bs map" as a separate project, "Diablo2.js", "MedianXL kolbot" (only DiaFlow found), "D2 Bot #" beyond D2Bot#/D2BotNG, "BNLS tools" (only snippet-level), "Kolbot-SoloPlay-D2R" (does not exist; SoloPlay FAQ says D2 only), "d2r-bot" (a 3-file JS stub, sin-tz/d2r-bot, unreadable README), "d2loader".
- GitHub topic pages were not enumerated; rate limits cut some search passes.
- Anything behind login (d2jsp, MPGH, private Discords).

## 7. What the community knows (and bots encode)

All claims below come from the bot files read in sections 2 to 5 (V) unless tagged otherwise; community guides (Maxroll, Almar, PureDiablo, diabloii.net) were only seen as search snippets (S) and were not read in full.

### 7.1 Leveling and run structure
- **Quest-complete skeleton (V, SoloIndex):** Act 1 Den, Bishibosh, Blood Raven, Tristram (cairn stones, rescue Cain), Countess, Smith, Andariel; Act 2 cube, Radament, Horadric staff parts, Summoner, Duriel; Act 3 Lam Esen's tome, temple runs, Khalim parts, Travincal, Mephisto; Act 4 Izual, Hellforge, Diablo; Act 5 Shenk (socket quest), rescue Barbarians, Anya (potion), Pindle, Nihlathak, Ancients, Baal. Quest details in this bullet are inferred from the script names (U); the order and the predicates on quest state and act access are V. The same ordering shows in Koolo's `leveling_act1..5.go`.
- **Gold fallback (V):** when broke, SoloPlay reruns cheap farms (Bishibosh, Tristram, Treehead, Countess, Lower Kurast).
- **Common farm list for later levels (V, kolbot + Koolo scripts):** Pindleskin (Nihlathak's portal), Mephisto/Travincal, Baal, Chaos, Countess for runes, Cows, Tombs, Ancient Tunnels, Pit/Arachnid for early MF/XP, Diablo clone and uber torch scripts as endgame goals.
- **Classic "Tristram runs 1 to 15, Tomb of Tal Rasha 15 to 24" fast-route style (S: Almar's Guides snippet)** and "Travincal 60 to 97, Chaos at the end (S: Maxroll general-leveling snippet; the page targets D2R)".

### 7.2 Build and stat tables (V)
- Hammerdin: Blessed Hammer maxed first, then Concentration and Vigor level by level, Blessed Aim and Holy Shield last; stats in expansion: vitality/dexterity/strength interleaved to fixed targets (examples in the file: vit 60/dex 30/str 27 up to vit 279/dex 147/str 103 and then all vitality), classic table dexterity 51, strength 80, rest vitality. The build's `wantedMerc` is a Holy Freeze mercenary (which act and tier: U).
- Charm policy (V): small charms with all-res 5 plus life 20 or MF 7 or FHR 5 (max counts 3/2/3), grand charms with +1 skill tab and life 40 or better.
- Respec trigger (V): expansion respecs once Enigma is owned; classic respecs at clvl 75 with Diablo done. Hence leveling builds are deliberately different from the final build.
- Auto-equip policy (V, d2-qolbot, SoloPlay): never replace Unique/Set/Runeword gear automatically; score gear with per-stat weights (SoloPlay `DynamicTiers`: IAS 3.5, min/max damage 3, lifeleech 8, crushing blow and open wounds 3, elemental damage 2, AR 0.1 for merc scoring).

### 7.3 Survival and town
- Chicken numbers shared by kolbot and Koolo (V): leave game at 30 percent life, drink healing at 75 percent, belt of 4 columns (hp, hp, mp, rv), min column 3. Dodging, merc chicken at 10 percent (Koolo), `TownHP` off by default.
- Debuff and aura chickens (V, Koolo): Amplify Damage, Decrepify, Lower Resist, Blood Mana; Fanaticism, Might, Conviction, Holy Fire within 25 units.
- SoloPlay sets `SkipImmune = ["magic"]` for Hammerdin in Hell before Act 5 access (V): the build's weakness (magic immunes) is encoded as a temporary skip rule.
- Hardcore and Diablo-clone safety lists: **not found in the bot sources**. Web guides (S, snippets only) say: resistances first, carry rejuvenation potions, do not face-check doors, walking vs running affects block/defense. I did not find a Diablo-clone-walk guide. Flag for the reviewer; the d2-sme agent may be the better source.

### 7.4 Town chores order (V, kolbot `Town.doChores`)
Heal, identify, clear inventory, fill TP tome, buy potions (and ID tome if field ID), repair, gamble, stash; with merc revive and key purchase handled by tasks keyed by NPC per act. Koolo mirrors this order. Item flow: pick up unidentified if the rule already matches, else identify and re-test, else sell (V, Koolo README).

### 7.5 Item filter philosophy (V, NipGuide)
- Write two or three strong stats plus a pool of "other good stats" rather than many exact matches; avoid catch-all lines that grab junk; split fast-cash and big-profit pickits; keep unid when the rule can match before identification. The NipGuide has a "pickit philosophy" section (fast cash, big profit, editing) which is the best human-written summary.

## 8. Quick comparison to d.idle (for the reviewer; these are my reading, not a decision)

| Topic | Public precedent | d.idle position |
|---|---|---|
| Reading the game | injected DLL calling engine functions (D2BS family), memory read + input (Koolo, MapAssist), pixels only (Botty, Diablo2Bot), packets only (D2NG) | memory read + a d2d decomp backend (per the brief) |
| Headless/faster than real time | d2bsng simulator on the real binary (V) | d2d reimplementation would allow it without the real binary |
| Pickit | NIP everywhere, with 3-section extension for auto-equip tiers | not read here; the SoloPlay `[tier]` idea is the relevant delta |
| Builds | JS module per build with skill list, stat list, gear NIP, respec rule | plan table (per AGENTS.md "builds are data") |
| Leveling | ordered run list with predicates (SoloIndex) | n/a |
| Safety | two-number chicken (30 and 75) plus aura/curse chickens | n/a |
| Licensing | most kolbot code has no licence | read-only for ideas |

## 9. Open questions for the reviewer

1. Is a deterministic, faster-than-real-time run mode (as in d2bsng's simulator) a goal for the d2d backend? If yes, its hook list in `docs/host_hooks.md` (V) is the checklist; confirm whether the leaderboard scoring is worth reading (needs the docs site rendered; I could not).
2. Should we adopt the NIP grammar as our item-filter format (large community corpus, vsnip-check tooling, two-stage unid matching), plus the SoloPlay `[tier]` third section, or keep our own? Licence of shipped `.nip` files is none, so corpora are reference only.
3. The kolbot skill-order format (`[skill, targetHardPoints, satisfy]`) and SoloPlay's `[stat, target]` list with an `"all"` sink are the de facto build-definition shapes. Do we want the `satisfy` gate and prerequisite auto-learning?
4. Do we need the "chicken on curse/aura" family (Koolo) in our survival policy? It is cheap to model and absent from the kolbot baseline.
5. PlugY-style infinite stash: kolbot users run with it. Do we want to assume it or model the vanilla 10 by 10 stash?
6. The search was GitHub-heavy. Forums (Blizzhackers, d2jsp, MPGH), Reddit and YouTube may hold more bots, especially Median XL and Chinese-language tools. Worth a second pass with a headless browser, or is the family (D2BS, D2NT, clientless, vision, D2R memory) already enough?
7. Hardcore and Diablo-clone safety lists were not found in the bots. Ask the d2-sme agent, or fetch Maxroll and diabloii.net archives, before claiming a "community safety list".
8. Take-down pattern: Koolo, MapAssist and original Botty repos are gone from GitHub (V 404). I have no source for why (U). Assume copies can disappear and snapshot anything we depend on only if its licence allows it.

## Appendix A: other ARPG bots (U, from memory, not fetched)

- **Path of Exile 1:** ExileBuddy (closed-source, memory-based plugin bot with a "profile/routine" model, paid) and similar "Oldschool bot" projects; lessons are the plugin/routine architecture and the account-ban economics. U.
- **Path of Exile 2:** community automation limited to overlays and trade helpers; I have not verified any bot repo. U.
- **Grim Dawn:** loot-filter and build-helper tools exist (item-filter mods, stasher tools such as the one seen in a snippet for NTQV5, S); I know of no mature bot. U.
- **Diablo 1:** DAPI and DevilutionX-AI above are V; they are the only RL/agent projects I verified.
- **Idle games inspired by D2:** none verified.
