# Same seed, same inputs: d2d vs Game.exe frame by frame? (card seedstudy, 2026-11-01)

> status: research, docs and read-only reading only. No game, no instance, nothing built or run. Target classic 1.14d (not D2R; D2R-only facts are flagged).
> Tags: **V** = read or fetched by me this session (file, line or URL named), **D** = decompile note of the d2decomped project (psi29a), read by me at the URL, whose claims rest on its own emulator checks; **S** = secondary (another doc of ours, a search snippet); **I** = my inference; **U** = unverified. "The worker host" = the lab machine. Companion docs: `docs/research/original-game-lockstep.md` (tiers A and B, the time hook), `docs/d2d-determinism.md`, `docs/d2d-fork.md`.

## 0. Answer in ten lines

1. **Game.exe picks a fresh random map seed for a new game** (QueryPerformanceCounter, `FUN_0052c280`), unless a seed is forced (`-seed` switch, or D2MOO `GAME_SetInitSeed`) **or** it is a single-player join whose save has the active-difficulty byte (0xA8..0xAA, bit 0x80) set: then the game takes the save's **map id (d2s 0xAB)** as its seed (`FUN_0056a090`). It writes the game seed back to 0xAB on save. D, matches our live finding that 0xAB is stable across 15 loads on one difficulty and changes on a difficulty change (V live, `docs/saves.md`).
2. **d2d does the same on purpose**: `game_seed()` returns the save's map id when the active-difficulty bit is set and the id is non zero, else the random `--seed`/run seed; `--seed N` forces it (V `apps/d2d/main.cpp` lines 376-380, 1689, 1767). So **one `.d2s` file pins the map seed on both sides with no command line and no memory write**: a file-level setup action (ADR 0014 setup write before the run, snaplib `place` ledger entry).
3. **Layout and monster placement from a seed are the game's own algorithm on both sides**: d2d's DRLG and room population were diffed against game.exe's code run in an emulator (Blood Moor 577/577 populated rooms, Cold Plains 787/787, Stony Field 710/710, Dark Wood 703/703, 20/20 seeds; V `docs/research/upstream-gaps/item3-monster-density.md`; Lut Gholein 20/20 after the MPQ fix, V `docs/research/lutgholein-proof.md`).
4. **The RNG stream after level generation does NOT match.** In game.exe every unit made steps the single game seed (`game+0xd0`) and gets `{step, 666}` as its own seed; monster life, look, drops and chests roll on those. d2d has one fixed-constant `Rng{0x2545f491}` for the world and "steps only for what rooms make" (V `components/game/world.hpp:349`, D `re/monsters.md`). Life and drop rolls therefore differ from the first monster, even with equal layout and equal spawn positions.
5. **Observation parity is good for a layout/spawn/position vector, poor for RNG state.** d2d exposes `monsters`, `dump ...` and `info`/`step` (frame, game_ms, tick), no seeds. Game.exe: client unit tables are readable, server units need a scan (about 1 s), the server game struct has no static pointer (frame at `game+0xa8`, game seed `game+0xd0`, D).
6. **Input parity needs Tier B** (frozen clock, frame barrier). Game.exe has no step today; d2d queues clicks and `cmd` to the next stepped tick.
7. **Verdict.** Static comparison at first load (seed pinned from one `.d2s`, level layout, preset units, spawn list, town NPC positions): **FEASIBLE WITH WORK** (days, no new hook). Frame-by-frame lockstep through play: **BLOCKED** by (a) Game.exe Tier B stepping (not built) and (b) the game-seed stream not modelled in d2d. Idle-town "first N frames" lockstep is the realistic ceiling after Tier B.
8. Top cards: SEEDPIN (live, read-only), SEEDLAYOUT (offline, same `.d2s`), SEEDSPAWN (the same on spawn lists), then Tier B (P1 frame read first), then a d2d-side card to model unit seed stepping.

## 1. How a single-player game seed is chosen and used on Game.exe 1.14d

### 1.1 Where the seed comes from

| Fact | Source | Tag |
|---|---|---|
| New game: `FUN_0052c280` seeds the game from `QueryPerformanceCounter`, unless one was forced (`DAT_00731004`, set by `FUN_0052c320`, game +0x84). A forced seed makes the game seed `{map seed, 666}` | `docs/research/re/drlg.md` lines 21-26, `re/monsters.md` lines 100-103, https://raw.githubusercontent.com/psi29a/d2decomped/main/docs/research/re/drlg.md | D (V read) |
| Same logic in D2MOO's decompile of the game create path: `if (nInitSeed == -1) { QueryPerformanceCounter(&seed); SEED_InitLowSeed(&pGameSeed, SEED_GetRandomValue(seed.LowPart)); dwInitSeed = ...} else { InitSeed = 1; dwInitSeed = nInitSeed; SEED_InitLowSeed(&pGameSeed, nInitSeed) }`; `GAME_SetInitSeed(int)` sets the global | https://raw.githubusercontent.com/ThePhrozenKeep/D2MOO/master/source/D2Game/src/GAME/Game.cpp lines 218-220, 501-517 | V (source; version addresses not checked against our exe) |
| Single-player join `FUN_0056a090`: if game type is 3, the seed was not forced, and the save's difficulty byte for this difficulty (d2s +0xA8..) has 0x80, the game takes the save's map id (d2s +0xAB, u32 LE) as its seed. "A character keeps its Act 1 layout game after game on one difficulty" | `re/drlg.md` lines 27-30 | D |
| Save: d2s +0xAB = game +0x7C (`FUN_00568f20`) | `re/drlg.md` line 31 | D |
| Our live data agrees: map id stable across Save and Exit and re-entry on one difficulty (15 loads, with act/quest/waypoint edits), changes whenever the game is entered on a different difficulty than the last saved one (3 of 3); `.map` file seed equals the header map id | `docs/saves.md` lines 23-26 | V live (earlier agents) |
| In a joined (LAN) game the host picks the seed and sends it in `[03]`; the host rewrites the joiner's save header map id with the host game's seed (sent 0x7543c724, returned 0xf221f26c = the host's seed) | `re/net-join.md` lines 310, 387; `docs/d2d-lan-oracle.md` lines 11, 29-30 | D, V live |
| Community: the map seed is the 4-byte field at offset 171 (0xAB) of the d2s; third-party tools read and write it (LoD tool "Diablo2MapIDChanger" rewrites it and fixes the checksum; a D2R tool `d2mapseed` does the same) | https://github.com/nokka/d2s README line 83 ("171 / 4 / Map ID"); https://github.com/Karyoplasma/Diablo2MapIDChanger/releases/tag/1.0.0; https://gh.r-mj.com/feored/d2mapseed | V (tool pages), S for behaviour; D2R tool is D2R-only |

Note the `.d2s` checksum (0x0C, rotate-left-1 plus byte, `d2s.Checksum`) must be recomputed after any edit (V `docs/snaplib.md` line 24).

### 1.2 Conflict to settle first

`docs/orig-spawns.md` line 9 says a fresh game per visit (`oog leave`, `oog start Testy`) gave "a new map seed each time, area layouts differ between visits" (V-orig by that agent), while `docs/saves.md` (V live) and the decompile say the id is reused on one difficulty. Possible reasons (all U): the Testy header's difficulty byte lacked 0x80, the map id was 0 or the level-1 Testy was re-saved at a different difficulty, or "different layouts" meant different monsters. **One read-only check settles it**: read 0xAB of the `.d2s` and the client `MapId` (RVA 0x3A0638, D `re/automap.md`) before and after two starts (card SEEDPIN, section 6).

### 1.3 How the seed is used (the chain)

All of it rolls D2's LCG: `{low, high}`, step `low * 0x6AC690C5 + high` split back, roll below n = `low & (n-1)` for powers of two else `low % n`, fresh seed high word 666 (D `re/drlg.md` lines 9-12; same function in d2d `rules::Rng`, V `components/rules/rules.hpp` lines 154-173; the missile/skill traces in `docs/research/sk-*-trace.md` re-confirm it from game.exe's `FUN_0045c3e0`, V run).

| Stage | Rule | Tag |
|---|---|---|
| Game creation `FUN_00530930` | after seeding, the game seed is stepped four times: monster regions (`FUN_00547d20`), object seed (`FUN_00546c60`), sunitproxy (`FUN_00536070`, vendor stock and hire lists), quests (`FUN_00545d80`). D2MOO's create path calls the same four in the same order (`MONSTERREGION_InitializeForGame`, `OBJRGN_AllocObjectControl`, `SUNITPROXY_InitializeNpcControl`, `QUESTS_QuestInit`, Game.cpp 520-527) | D, V read |
| Act seed | `{init, 666}` from the map seed, stepped once, low word kept at act +0x470 | D |
| Level seed (+0x1C4) | `{act+0x470 + level id, 666}`; "every level's contents depend only on the map seed and its id" | D |
| Room seed | room2 seed stepped once after its tiles, low word is room1's `{s, 666}` (+0x6C); population rolls on it | D `re/monsters.md` 153-157 |
| After creation | "Nothing else steps [the game seed] inline; after that only room population (`FUN_0054ec90`) rolls it", and **every unit made** (`FUN_00555230` -> `FUN_00552df0`: step `game+0xd0`, unit seed `{low,666}`) | D `re/monsters.md` 122-166 |
| Drops, life, look | monster life and look roll on the unit seed (`FUN_00573cb0`); a drop starts from the seed "as the monster's life left it"; nothing on the death path draws on it | D `re/drops.md` 106-130 |
| Per-missile / per-shard | missile seed set from the spot and age (`{cx + age, 666}` for Blizzard shards, `FUN_00650e40`) | V run (game.exe in the emulator, `docs/research/sk-ground-trace.md`) |

Consequence (D, stated by the project itself): "which puts the stream's absolute state out of reach without modelling every unit (d2d steps only for what rooms make)". With a forced seed and identical inputs and order, game.exe itself is reproducible (D `docs/sme/d2decomped-findings-2.md` line 249, a harvest of the same notes).

### 1.4 Client side RNG

- Game.exe runs client and server in one process; the client's main loop calls the server frame and flush (D `re/network.md` lines 3-7). The client keeps its own copies of units (V `docs/cheats.md` 91-95: two UnitAny copies of the player). Visual rolls (rain) use the player unit's seed in game.exe (D `re/deviations.md` Approximations #7: "player unit's seed"; d2d uses a client seed of its own) but on the client's copy, so I expect no feedback into server state (I; U that no client roll ever reaches a server unit: the code is not traced for it).
- Client-only effects that can still differ between two runs: animation timing off wall clock, frame skipping, UI sounds, light/shadow quality by FPS (D `re/deviations.md` Approximations #4). None of them is a server input (I).
- A risk that is NOT client RNG: each server frame gets a "speed factor" (game +0x1DBC, 0x400 = 1.0) computed from real time since the last frame, clamped 10..0x800 (D `re/network.md` lines 15-18). If any unit logic reads it, wall time leaks into state. D2MOO's note says the CPU ratio is forced to 1024 (V, `original-game-lockstep.md` section 2.1). Which is true on 1.14d is U; a Tier B hook must freeze it to 1024 or the test fails by design.

## 2. Can the seed be fixed, and what the rules allow

| Route | Works? | Policy (ADR 0014, `docs/adr/0014-memory-writes-allowed-no-crutch.md` on branch agent/memrules, not yet on main when I read it) |
|---|---|---|
| **Save map id** (d2s 0xAB + active-difficulty bit set + checksum): copy one `.d2s` set (snaplib `place`, or a hex edit through a logged tool) | V live that the id is honoured across loads (`docs/saves.md`); that it pins the layout is D, not yet proven by us (`docs/snaplib.md` line 43 asks for exactly this test). Both games load the same file (V `docs/research/character-factory.md` 4.7: d2d reads our `.d2s` from its XDG save dir) | A file copy before the run is a declared SETUP action; a copy of an honest character stays `copy-of-honest`; a write AFTER the run-start marker makes the run CHEATED (V `docs/snaplib.md` line 28). Editing 0xAB itself is a setup write: the character is then "setup-edited", not honest, for gate purposes. For a comparison experiment that counts nowhere (research run) this is fine. |
| `-seed N` on the Game.exe command line | D: the switch exists in the 58-switch table (DEBUG/SEED, struct offset 538) and `FUN_0052c280`/`FUN_0052c320` read it; d2d keeps "game.exe's own switches" but ignores `-seed` until traced (D `re/switches.md` lines 146, 186-188, 233-235). Under Wine untested; whether retail 1.14d honours debug switches without `-cheats` is U. Launching games is outside agent rights (AGENTS.md); only the OOG/launch tooling may start instances | The command line is not a memory write: allowed as setup, declared. |
| Memory write at game creation (set `nInitSeed` global `DAT_00731004` or the game +0x7C before level 1) | untested (U); needs a hook or a pre-game write | INSTRUMENTATION/SETUP under ADR 0014 needs the user's go per hook set; no bot decision depends on it. Not needed if the save route works. |
| d2d `--seed N` | V (V `docs/d2d-fork.md`, `docs/d2d-spike.md` line 56: layout identical across sessions) | free |

Recommendation (I): use the save route. It is the only one that needs no new code, no launch flag and no memory write, and it is the only one that makes both sides read the same state file.

## 3. How d2d chooses and uses the same

- Choice: `--seed N` forces; otherwise a loaded save with an active difficulty bit and map id != 0 supplies the seed; else `std::random_device` (V `apps/d2d/main.cpp` 376-380, 1689-1692, 1767-1768; D `re/deviations.md` Improvements #2: a saved map id of 0 is reused as seed by game.exe, d2d treats 0 as "none"). Joining a host replaces it by the host's (V main.cpp 1375-1376).
- Generator: `components/drlg/*` is a re-implementation of game.exe's DRLG (not the game's binary) and has been differentially checked against game.exe's own functions in an emulator: rooms, tiles, units, room seeds, monster entries per level (V `docs/research/upstream-gaps/item3-monster-density.md`, `docs/research/lutgholein-proof.md`; D `re/monsters.md` 115-117 for the regions on 125 levels, 5 seeds). Known fixed gap: Lut Gholein's two level-seed steps (a `drlg-dump` tool MPQ issue, not the generator, V lutproof).
- Runtime RNG: one world `Rng rng{0x2545f491u}` (also stock, talk, gambles, merc offers), a Town one `0x7f4a7c15u`, sound `0x5eed5u` (V `world.hpp` line 349; S `docs/d2d-fork.md` section 2). So d2d draws are determined by seed and by the order of draws, not by game.exe's unit-seed rule; d2d notes "the roll is on the unit id, not the unit's seed" in places (V `world.hpp` ~595). Same `--seed` plus same inputs is reproducible on d2d (V `det.py`, `docs/d2d-fork.md` section 5; full bot loop reproducible for about 3000 ticks and longer after fixes, V `docs/d2d-determinism.md`).
- Lockstep: our private fork's `--lockstep` and `step N [render]` reply `frame=<f> game_ms=<ms> tick=<t>`; the `dump` verb has a `time` section; `info` gives the process frame counter, which differs by start-up idle frames and is masked in the journal (V `docs/d2d-fork.md` 61-90, `docs/d2d-determinism.md` row 5). `tick` and `game_ms` are the comparable counters. **The RNG state is not exposed** (no verb in the public channel nor in the fork usage I read, V `docs/control_channel.md`; U for the fork's newer verbs). Upstream's public tree has no step verb (V, `original-game-lockstep.md` section 1).

## 4. Observation parity: a minimal comparable state vector

| Field | Game.exe 1.14d (external read) | d2d | Comparable? |
|---|---|---|---|
| frame counter | server `game+0xA8` (D); no static pointer, found by scan (cheats doc). Client counters: U | `tick`, `game_ms` from `step`, `dump time` | needs a Game.exe pointer (probe P1 of `original-game-lockstep.md`) |
| map seed | `Act.dwMapSeed` +0x0C and `MapId` RVA 0x3A0638 (V D2Structs.h:359-363 in `docs/d2-1.14d-memory.md` lines 71, 236; D `re/automap.md`); d2s 0xAB after save | the seed printed at start / `debug tomb`; `--seed` | yes, equal u32 |
| level, area id, room list | `pRoom1`/`pRoom2`/`pLevel` (V `docs/d2-1.14d-memory.md` 231-257) and the dump tools (`d2mapdump`, whisper `town`) | `debug level`, `drlg-dump DATA SEED LEVEL tiles/units/seeds` | yes (offline oracle exists) |
| player position, life, mana, stats, mode | client unit via `PlayerUnit` 0x3A6A70, stats vector (V memory doc 2.2-2.4); server copy by scan | `state`, `dump player,stats` | yes in subtiles; client lags the server by one packet round (`original-game-lockstep.md` 2.2), so compare after the barrier |
| monster list: class, x, y, life/max, mode, flags | client unit hash tables (V 2.5); only rooms activated near the player (V `docs/orig-spawns.md`: lower bound); whisper `units` also lists client-known units | `monsters` (all monsters of the current outdoor level, the fork keeps every room loaded) | classes and positions at spawn: yes after filtering to the rooms Game.exe has activated; life: only if the RNG stream matches (no) |
| unit ids | the game's id counter | `#<unit>` | U that they agree; do not key on them: key on (class, x, y) |
| items on ground / inventory | client item units (V 2.7) | `ground`, `dump items` | drops: no (seed stream) |
| unit seed `+0x20`, game seed `+0xD0` | server units only (D); needs the scan; U that they are stable to read | not exposed | the single most informative field, one-sided |

**Minimal vector V1** (per sample): `{frame, map_seed, level_id, player{x,y,life,mana,mode}, monsters sorted by (class,x,y){class,x,y,mode}, count}`. Hash with the journal scheme (`D2DCTL_JOURNAL`, `d2ddet compare`, V `docs/d2d-determinism.md` section 1) minus life/drops for cross-side runs. Same-side runs (P3a, P3b) can use the full vector including life and unit seeds where readable.

## 5. Input parity

- d2d: `click`/`cmd` are queued and applied at the next stepped tick; click-while-stepping is not possible by construction (V `docs/d2d-fork.md` section 5). The bot's `Pacer` releases input at frame boundaries (V `docs/d2d-determinism.md` 3.1).
- Game.exe: the bot clicks through the window (mouse messages, `internal/input`), which the client turns into a packet applied on the server's next frame through the in-process loopback queue; the arrival frame depends on wall time and thread order (V `original-game-lockstep.md` 2.3). Alternatives: whisper's `send HEX` raw C->S packet verb (D `docs/research/d2d-whisper-lan-oracle.md` line 13), which goes into the same queue as the client's own packets and skips the mouse path. It is GPL and not loaded in our Wine runs (U).
- Same-frame injection on both sides needs: frozen clock and barrier on Game.exe (Tier B), `step` on d2d, and the input expressed as game commands (a move to subtile X,Y, a skill id at X,Y) not as screen clicks, because d2d's camera and click mapping differ (I). A smaller first target: no inputs at all (idle town, frame 0 to N), which only needs Tier B's frozen clock.

## 6. Determinism test plan P3 (extends P3 of `original-game-lockstep.md`)

Success criteria are equality of the state vector at checkpoints every 250 ticks (same cadence as the d2d journal) and the tick of the first difference.

| Step | What | Needs | Pass |
|---|---|---|---|
| **P3-0** SEEDPIN (read only, tier none) | One throwaway honest character, two cold starts of Game.exe on the same `.d2s` (difficulty bit set). Read 0xAB of the file before, `MapId`/`Act.dwMapSeed` in game, 0xAB after. Then copy the file to a second prefix (snaplib) and start there. | a live agent with an instance (d2-live); no hook | in-game seed == file id in all starts; or the conflict of 1.2 is explained |
| **P3-1** layout parity | Same `.d2s` to d2d (`--no-save`, loaded from its slot); compare Act 1 rooms of Rogue Encampment and Blood Moor: `drlg-dump` (offline) vs `d2mapdump`/whisper-style dump from the live game | tier none | room list and per-room seeds equal; tiles equal for the rooms Game.exe has up (use the activation order of `lutgholein-trace.md` point 2) |
| **P3-2** spawn parity | Walk the same route in Blood Moor in both; compare monster (class, x, y) for the rooms Game.exe activated | tier none, bot walk (non-lockstep on Game.exe) | sets equal; life/mode differences expected |
| **P3a** same-side repeat, Game.exe | Same `.d2s`, same scripted inputs, twice; first 5000 frames (idle town, then one walk) | **Tier B** (inputs at frame boundaries; with Tier A only the seed/layout check) | equal V1 including unit seeds; failure with equal seeds = race, expected without Tier B |
| **P3b** same-side repeat, d2d | already done for 12000 ticks (V `docs/d2d-determinism.md`) | none | stays as the control |
| **P3c** cross-side | Equal seed from the file, equal command list, frames 0..N | Tier B + a d2d card (section 8) | first divergence tick reported; see section 7 |

Tier A (speed-up) suffices for: P3-0, P3-1, P3-2, seed and layout checks, and for making P3-2 faster. **Tier B is needed for** P3a and P3c (frame equality needs input at frames and a frozen clock/speed factor).

Hard blockers for cross-side frame equality (named): B1 no Game.exe stepper (Tier B, `original-game-lockstep.md`, estimate 20 to 40 agent-hours, high risk); B2 d2d's world RNG is a single fixed stream, not game.exe's per-unit seeds, and the game-seed stream is advanced by every unit made (camp NPCs, merc, items, missiles) which d2d does not model (D); B3 no Game.exe pointer to the server game struct (scan only), so seeds and the frame counter cannot be read cheaply; B4 the wall-time speed factor (1.3, U).

## 7. Where the first divergence is expected (in order, I unless tagged)

1. **Not at level generation.** Layout, room seeds, preset units and populated monster entries should match (V emulator diffs; the P3-1 check confirms on the live game). Exceptions known: Act 2 town (fixed), tiles on shared room edges depend on activation order (V lutgholein-trace), and `drlg-dump` brings all rooms up while the game brings up the near ones.
2. **First draw on a unit seed.** Monster life roll and look (`FUN_00573cb0`) at creation: Game.exe's seed is a step of `game+0xD0` after four creation steps plus every unit made so far (camp NPCs and preset objects first); d2d draws from its own stream. Expect HP and champion/rare modifiers to differ as soon as monsters are up, **frame 0 of the first outdoor level**, even though positions match (V item3: counts equal; positions are placed by room seeds, not the game seed, D).
3. **Vendor stock, gambles, hire lists** (sunitproxy seed, then rerolled every 4 minutes in game.exe, D `docs/sme/d2decomped-findings-2.md` section 10 item 4): d2d's `rng` draws differ from the first shop open.
4. **First attack**: hit/block/crit rolls on the attacker's seed (D `re/combat.md` 157-160): different outcomes once any unit seed differs.
5. **Wall-time effects**: speed factor, waypoint 10 s window (V `waypoint-trace.md`; never fires in single player), teleport landing stamp, client smoothing. Equal only under Tier B's frozen clock.
6. **Client-only**: rain, ambience, light quality: invisible to the vector V1.
7. **Frame offset**: d2d's town idle loop is fixed 40 ms ticks from the first step; Game.exe's frame counter starts at game creation and loading frames run in wall time (I). The comparison must align on "frames since the game was ready" (a handshake: first frame after the player is placed), not on absolute frame number.

Practical reading: the "time to first divergence" will be about 0 frames for life/drops and unbounded (equal) for positions of static things, so the method must compare *classes of state separately*, not one hash, otherwise it reports tick 0 forever.

## 8. Ranked cards

| # | Card | Owns | Needs | Effort | Why |
|---|---|---|---|---|---|
| 1 | SEEDPIN: read-only live probe P3-0, resolve section 1.2, one throwaway honest character, no cheats | d2-live, one instance | instance, ADR 0014 setup declaration | 1 to 2 h | decides whether the save pin works; cheap |
| 2 | SEEDLAYOUT: same `.d2s` into d2d and Game.exe, compare room lists / seeds / tiles (P3-1) and a `d2diffseed` script | docs + tool | card 1; offline oracle `tools/emu` + `drlg-dump` | 4 to 8 h | the layout half of the verdict, no hook |
| 3 | SEEDSPAWN: monster (class, x, y) lists on a fixed walk (P3-2) | d2-live | cards 1, 2 | 4 to 8 h | proves "same seed, same world" for Normal Blood Moor |
| 4 | P1 + P0 of the lockstep doc (frame counter read, Wine clock test) with the server `game+0xA8` pointer found by scan | docs/tool | none | 3 to 6 h | gives B3 and the clock model |
| 5 | Tier B stepper (frozen clock, barrier, speed factor = 0x400, control socket) | separate card, user's go (hook) | cards 1, 4; ADR 0014 hook go | 20 to 40 h | B1, B4, only way to P3a/P3c |
| 6 | d2d unit-seed fidelity: make unit creation step the game seed like `FUN_00555230`, order of creation as game.exe (camp NPCs, merc), per-unit seeds for life, look, drops | fork card; upstream PR only with the user's go | emulator oracle (hooks exist: drops 39980/39980 for the drop path, D `re/drops.md`) | large (days) | B2; without it P3c diverges at frame 0 of the first monster |
| 7 | Cross-side harness (journal + `d2ddet compare` extended to V1 with per-class mask) | tool | cards 3, 5 | 6 to 10 h | reports first divergence per state class |
| 8 | Optional: `-seed` on a Wine launch (D says the switch exists) | launch tooling | user's go to launch with a flag | 1 h | second route; only if the save route fails |

## 9. Open questions

- Does the save pin really reproduce the Game.exe layout on the worker host (card 1)? Conflict with `docs/orig-spawns.md` unresolved.
- Does the map id with the `.map` file matter (the 24-byte file repeats the seed)? Unknown whether a mismatch between 0xAB and `.map` is tolerated (U).
- Is the speed factor used in state (U)? Does any client roll reach a server unit (U)?
- d2d unit ids vs game ids (U). Whether the fork's newer verbs expose any seed (U).
- Whether retail 1.14d honours `-seed` without `-cheats`/debug flags and under Wine (U).
- ADR 0014's text was read on branch agent/memrules; the final wording on main may differ.
