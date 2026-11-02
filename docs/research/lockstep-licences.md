# Licences and legal footing for a releasable Game.exe 1.14d time/lockstep hook

> status: research, card `licensecheck`. Not legal advice: the author is an AI agent, not a lawyer. Everything below is a reading of public texts for the user and a lawyer to check. Open questions for a lawyer are in section 9.

Evidence tags: V = fetched and read in full or in the quoted part (URL given, fetched 2026-11-02 unless stated), S = snippet or secondary source, U = unverified or from memory. Target: classic 1.14d (Game.exe), not D2R; D2R-only facts are flagged.

Plan being checked: a standalone, releasable hook. Tier A scales the time APIs from a proxy DLL (generic). Tier B steps the game frame by frame and needs 1.14d addresses. Parent docs: `docs/research/original-game-lockstep.md`, `docs/research/lockstep-p0.md` (branch `agent/lockstepp0`). The user's lean: MIT, clean-room, standalone repo, offline-only use, no long-term support.

## 1. What our own rules already say (V, this repo)

- `docs/decisions.md`: "d2d is GPL-3.0: talk to it over its socket as a separate program, never copy its code or notes verbatim into this repo (own summaries with file-path citations)". Also "Licence: GPL (same as d2d, GPL-3.0-or-later) - user, 2026-10-30 (ADR 0010)".
- ADR 0003 (accepted 2026-10-30): "Private forge repo <owner>/d2d-fork (base e36922e); the bot is a separate program talking over the socket; our docs give own summaries with file-path citations; no distribution until the user names a GPL reviewer (none yet)".
- ADR 0010: d2lod-bot is GPL-3.0-or-later. "Third-party code policy: No code copied from projects without a compatible licence. ... D2MOO is MIT (V ...), GPL-compatible; copying would still need its notice kept. Decompiled Blizzard code never enters the repo."
- AGENTS.md: "Patches never go into this repo (GPL, docs/decisions.md)"; "Do not write secrets ... or private hostnames".
- Consequence for this card: the d2d fork patches (private, GPL) are a separate matter. A new MIT hook repo must not contain fork patches, d2d code, or text derived from the fork. Our own `tools/lockstep/hook.c` (about 100 lines, written by us, no d2d code copied) sits in a GPL repo; the copyright holder(s) may release it under MIT, see 9.2.

## 2. Licence verdict table

| Source | Licence (what was read) | Reuse in an MIT repo | Tag |
|---|---|---|---|
| psi29a/d2decomped (d2d) incl. `apps/whisper` | GPL-3.0-or-later | Facts only (addresses, API behaviour, ordinals). No code, no close paraphrase. | V |
| D2MOO (ThePhrozenKeep) | MIT, plus a README sentence about "non-commercial use only" | Code technically allowed with notice, but do not copy (decompiled Blizzard origin, ambiguous non-commercial sentence). Names/addresses as facts. | V |
| libTAS | GPL-3.0 | Ideas only. | V |
| speedhack-rs (Hirtol) | No licence found | Nothing (all rights reserved by default). Design idea only. | V (absence) |
| clock_gettime_override (hi-angel) | No licence found | Nothing. | V (absence) |
| PlugY | GPL-3.0 in the maintained fork; original author's terms not fetched | Ideas only; precedent for distribution. | V fork / U original |
| Cheat Engine speedhack | not fetched | Nothing. | U |
| Blizzard binaries, data, decompiled code | Blizzard copyright, EULA | Never. | V (EULA) |

## 3. d2decomped and the whisper loader (V)

URLs: https://github.com/psi29a/d2decomped, `LICENSE`, `README.md`, `LEGAL.md`, `apps/whisper/` (raw.githubusercontent.com/psi29a/d2decomped/main/...). GitHub API (2026-11-02): licence key `gpl-3.0`, SPDX `GPL-3.0`, last push 2026-11-01.

- `LICENSE` is the GNU GPL version 3 text. README section "License": "D2Decomp is licensed under the GNU General Public License, version 3 or later (GPL-3.0-or-later); the full text is in LICENSE." ADR 0010 recorded the same reading and that source files carry `SPDX-License-Identifier: GPL-3.0-or-later`.
- `LEGAL.md` (additional notice, not a licence change): "**Our own code only.** Every line of the engine is written for this project and licensed under the GPL-3.0-or-later." "**No Blizzard files.** The repository holds no game data ... no game executables or libraries, no copied code, and no dumps of the original program's bytes." It relies on EU Directive 2009/24/EC Articles 1(2), 5(3), 6 and says "It is not legal advice." It lists "Two pieces of licensed artwork" (Freepik icon, AvQest font) that are NOT GPL. Under "Honest limits": "Scope is debated", "Other countries differ", and "Blizzard's online services are a separate question. Their terms forbid unauthorized connections to Battle.net."
- `apps/whisper/` contains `README.md`, `build.sh`, `ijl11.def`, `whisper.cpp`, `whisper.py`. `whisper.cpp` has no per-file copyright or SPDX header (grep for spdx, copyright, licen: no hits), so it falls under the repository's GPL-3.0-or-later. No separate licence file in the folder.
- Design (V, `whisper.cpp` header comment and README): "loaded as ijl11.dll (game.exe imports ordinals 2, 3, 5 from it) and forwards those to the real one, renamed ijl11_real.dll"; listens only when `whisper.txt` sits beside game.exe; runs work on the game's thread via a `WM_APP` message; verbs include `peek ADDR N`, `click`, `key`, `send` (raw packet), `screenshot`. It has NO time hook and NO lockstep or control channel for time; the lockstep we use is our private fork of the d2d engine (a different program, not Game.exe).
- What is a fact and so free to use (U, legal view): that Game.exe 1.14d imports `IJL11.DLL` ordinals 2, 3 and 5; that a proxy DLL with the same exports loaded from the game folder runs first. What is expression: the `whisper.cpp` forwarder code (naked `jmp *_real_ordinal_N` thunks), the verb set and protocol, the hardcoded address tables as arranged in the source.
- Verdict: GPL. Nothing is copied. The parent doc already says "code is GPL, so we would write our own". Our P0 `hook.c` is a clock IAT patch and a stub DLL written independently (V, `lockstep-p0.md`); the forwarder for ordinals 2, 3, 5 is NOT yet written. Write it from the PE export-table format and a `.def` file derived from the DLL's own export table read on our own copy (see section 6), not from `whisper.cpp`.
- Taint note (V, parent doc): an earlier agent read `apps/whisper` and summarised it in `docs/research/d2d-whisper-lan-oracle.md`. Whoever writes the forwarder must not read that source or that summary section; give them the facts list only (ordinals, filename rename).

## 4. D2MOO (V)

URLs: https://github.com/ThePhrozenKeep/D2MOO (`LICENSE`, `README.md`, GitHub API, 2026-11-02).

- `LICENSE`: "MIT License / Copyright (c) 2020-2025 The Phrozen Keep community / Permission is hereby granted, free of charge, to any person obtaining a copy of this software ... to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software ..." subject to keeping the notice. GitHub API: `mit`.
- README "Legal": "The source code in this repository is intended for non-commercial use only. However it uses a permissive license (MIT) so that any modder may use this. Credits to the team are appreciated, and derivative work must respect the license." Plus Blizzard trademark notices and "D2MOO and any of its' maintainers are in no way associated with or endorsed by Blizzard Entertainment(R)." The "non-commercial" sentence sits beside the MIT text and is not a clause of it; whether it narrows the grant is unclear (9.3). The README also says: "The game assets, main executable code and anything related to hacking or piracy is NOT endorsed by the authors".
- The frame gate: `source/D2Game/src/GAME/Game.cpp`, `int32_t __stdcall GAME_UpdateGamesProgress(int32_t bLimitCatchUp)` at line 2184 of master (V, file fetched, 3430 lines). It calls `QueryPerformanceCounter`, takes `timeGetTime() & 0x7FFFFFFF`, keeps a last-frame-time global, and computes elapsed ms; a `#ifdef NO_BUG_FIX` block keeps original behaviour. The 1.14d address `0x0052FC20` in our parent doc is separate (from the earlier read, U here).
- Caveat that applies to every "decompiled" project: the code is a reconstruction of Blizzard's compiled program. MIT from the D2MOO authors cannot give rights they may not hold over Blizzard's expression (U, legal). That is a reason not to copy lines even though the licence text allows it; behaviour and names as facts are the safe use.
- Verdict: reuse as a reading source. If any line were ever copied, MIT notice and the copyright line must travel with it, but the clean-room rule below forbids copying.

## 5. libTAS, speedhack-rs, other speed hacks, PlugY

### libTAS (V)
- https://github.com/clementgallet/libTAS, `COPYING` fetched: GNU GPL version 3 (API: `gpl-3.0`). GPL section 5(c) as read: "You must license the entire work, as a whole, under this License to anyone who comes into possession of a copy." Any code of ours that incorporates libTAS code, or is a "work based on" it, must be GPL.
- Use: ideas only (game starts paused, frame advance, time hooking is the generic technique). Techniques and concepts are not copyrightable subject matter in themselves (U, general law; LEGAL.md of d2d makes the same point for the EU, V). Read its code only if the reader is not the writer.
- It is a Linux LD_PRELOAD tool; parent doc already rejected it for Game.exe under Wine.

### speedhack-rs (V, absence)
- https://github.com/Hirtol/speedhack-rs: README says "A simple utility to accomplish the same speedhack as can be found in CheatEngine"; ships `version.dll` (or `hid.dll`) beside the exe. The repo root has `.github`, `.gitignore`, `Cargo.toml`, `README.md`, `rustfmt.toml`, `speedhack_proxy`, `speedhack_rs`; no LICENSE/COPYING file (four names probed, 404), no `license` field in the three Cargo.toml files, GitHub API licence null. No licence means all rights reserved by default (U, general law): we may read it, may not copy it. We may borrow the publicly described design (proxy DLL named after a system DLL, hook the time functions, config file, "wait_with_hook" delay for games that crash on instant hooks).

### clock_gettime_override (V, absence)
- https://github.com/hi-angel/clock_gettime_override: no licence file (four names probed), API licence null. Same verdict. Linux preload, not applicable to Game.exe anyway.

### PlugY (precedent) 
- Maintained fork https://github.com/ChaosMarc/PlugY: `LICENSE` is the GNU GPL version 3 (V, API says `GPL-3.0`; the fork's last push 2022-07-14). README title: "PlugY, The Survival Kit - Plug-in for Diablo II Lord of Destruction (1.09b - 1.13d + 1.14d)". Repo folders: `PlugY`, `PlugYInstall`, `PlugYInstaller`, `PlugYRun` (a loader `.cpp`/`.vcproj`), `Commons`. Features list includes "Disable access to Battle.net" and "Always regenerate maps in SinglePlayer" (V, README): the plug-in itself is positioned as offline and TCP/IP only. The original author's own licence statement was not fetched (the author's site and a mod-hosting page returned nothing usable), so whether the original was GPL from the start is U.
- Precedent value: a source-available D2 LoD plug-in that supports 1.14d, is loaded by a launcher, and is published as source plus an installer with no Blizzard files in the repository (V: the folder list above contains none; what the release zip contains was not inspected, U). That a D2 loader/hook can be published openly is shown; that Blizzard tolerates it is NOT documented anywhere I fetched (U). Its being GPL is a reminder that "same kind of tool" projects exist under copyleft; it does not constrain us unless we copy.
- Other D2 tools (BH, d2hackmap, D2SE, D2 Loader, Median XL): licences not fetched (U), not used.

## 6. Blizzard terms (V unless marked)

### 6.1 Blizzard End User License Agreement (LAST REVISED March 21, 2024)
URL: https://www.blizzard.com/en-us/legal/fba4d00f-c7e4-4883-b8b9-1b4500a402ea/blizzard-end-user-license-agreement (fetched, 188 lines of text). It covers "the Battle.net desktop app software ... each of the Games" as the "Platform". Whether it governs a retail 1.14d copy installed from a disc is U: the page says games from other sources may have "a separate End User License Agreement". The original 2000/2001 Diablo II LoD EULA text was not fetched; the only quoted form I found is a court opinion's wording (6.4, S/V).

License Limitations ("You agree that you will not, in whole or in part or under any circumstances, do the following"):
- "Derivative Works: Copy or reproduce ..., translate, reverse engineer, derive source code from, modify, disassemble, decompile, or create derivative works based on or related to the Platform."
- "Cheating: Create, use, offer, promote, advertise, make available and/or distribute the following or assist therein: cheats; i.e. methods not expressly authorized by Blizzard ... influencing and/or facilitating gameplay ... and thereby granting you and/or any other user an advantage over other players not using such methods; bots; i.e. any code and/or software, not expressly authorized by Blizzard, that allows the automated control of a Game ... e.g. the automated control of a character in a Game; hacks; i.e. accessing or modifying the software of the Platform in any manner not expressly authorized by Blizzard; and/or any code and/or software, not expressly authorized by Blizzard, that can be used in connection with the Platform and/or any component or feature thereof which changes and/or facilitates the gameplay or other functionality".
- "Data Mining: Use any unauthorized process or software that intercepts, collects, reads, or 'mines' information generated or stored by the Platform".
- "Unauthorized Connections: ... any connection using third-party programs or tools not expressly authorized by Blizzard." "Matchmaking: ... intercept, emulate or redirect the communication protocols used by Blizzard in any way".
- Reserved rights: "The right to create derivative works, and as part of this Agreement, you agree that you will not create any work based on the Platform, except as expressly set forth in this Agreement or otherwise by Blizzard in certain contest rules, Blizzard's Fan Policies, or addenda".
- Warden-like monitoring is documented only as a consent clause: "WHILE RUNNING, THE PLATFORM (INCLUDING A GAME) MAY MONITOR YOUR COMPUTER ... MEMORY FOR UNAUTHORIZED THIRD PARTY PROGRAMS ... IN THE EVENT THAT THE PLATFORM DETECTS AN UNAUTHORIZED THIRD PARTY PROGRAM, (a) THE PLATFORM MAY COMMUNICATE INFORMATION BACK TO BLIZZARD, INCLUDING ... YOUR ACCOUNT NAME ... AND/OR (b) BLIZZARD MAY EXERCISE ANY OR ALL OF ITS RIGHTS UNDER THIS AGREEMENT". The words "Warden" and "anti-cheat" do not appear in this EULA (grep, V).
- Reading for us: the EULA's "bots" and "modify" limits on their face cover the d2lod-bot itself and any in-process hook, online or offline; the clause is not limited to Battle.net play. Our whole project sits outside what Blizzard "expressly authorized". That is the main legal-terms risk and it already applies today; a public hook makes it more visible, not different in kind (U, judgment).

### 6.2 Blizzard Legal FAQ
URL: https://www.blizzard.com/en-us/legal/28d5ebbf-c245-4408-8ba9-043dd5f056bf/ (undated page, fetched; old, still hosted, text refers to Diablo III and Hearthstone forums so it is at least 2010s).
- "Can I make add-ons or expansions for Blizzard Entertainment games? ... We do not allow add-ons or expansions for Diablo(R)."
- "Can I do a total conversion of your games? Yes. ... no problems with total conversions so long as they are for personal use and do not infringe our End User License Agreement included in our games, nor the rights of any other parties including copyrights, trademarks or other rights."
- "Can I create and/or distribute hack and cheats for your games? No. Blizzard Entertainment does not support or condone the use or distribution of cheats and/or hacks for use with Blizzard Entertainment games under any circumstance."
- "Can I put your patches and demos on my web site for download? ... so long as you do not alter the patches or demos in any way"; and on fan-site use of Blizzard material: licence for "home, noncommercial and personal use only", revocable "at any time, for any reason".
- On network play: "you may not host or provide matchmaking services ... or emulate or redirect the communication protocols ... through protocol emulation, tunneling, modifying or adding components to the game(s), use of a utility program".
- Reading: the FAQ's friendly line (total conversions) is for personal-use data conversions; the Diablo line "no add-ons" and the cheats line point against a public release. A time-scaling hook that helps only the offline owner is not obviously a "cheat" in the EULA's own definition (advantage "over other players") but is "software that can be used in connection with the Platform ... which changes ... gameplay" (U, a lawyer should read it).

### 6.3 Mods, Warden and enforcement: what is documented
- Blizzard statement on D2R mods (S, Kotaku, 2021-05-27, https://kotaku.com/blizzard-sends-cease-and-desists-to-diablo-ii-resurrec-1846977935, fetched via a summariser): "a big part of Diablo II's longevity is the modding community"; "some mods are atypical and pose security threats to our games"; "Classic Diablo II and its mods will continue to exist". Cease-and-desist letters went to the authors of D2ROffline and D2RModding (two D2R-era tools; D2R-ONLY facts, not 1.14d). No statement found that names 1.14d tools or PlugY.
- Warden: Blizzard documents no Warden behaviour for Diablo II in any page I fetched. Secondary only (S, search snippets from a fan wiki and EFF articles): Warden has been part of Diablo II since patch 1.11 and compares running programs against known cheat tools; Blizzard said in 2005 it collects no personally identifiable information. Whether Warden runs in offline single player is U (commonly said to be Battle.net-only, not verified).
- Enforcement precedent for classic D2 (S, https://gamedeveloper.com/game-platforms/blizzard-suspends-350k-i-starcraft-diablo-ii-i-users-for-third-party-hacks-, fetched, 2008-11-12): Blizzard "expanded our efforts to remove cheaters from StarCraft and Diablo II"; "Cheating ruins the game experience for legitimate players, and we will not tolerate it."; the affected Diablo II CD keys were restricted from Battle.net for 30 days; the article does not say how detection worked or mention Warden.
- No Blizzard text I fetched addresses "single-player mods" for classic 1.14d.

### 6.4 Court precedent on the EULA (V for the opinion text via summariser, S otherwise)
Davidson & Associates v. Jung, 422 F.3d 630 (8th Cir. Sept 1, 2005), https://law.resource.org/pub/us/case/reporter/F3/422/422.F3d.630.04-3654.html: EULA wording quoted "you may not, in whole or in part, copy, photocopy, reproduce, translate, reverse engineer, derive source code, modify"; the court held the EULA and terms of use enforceable, copyright law did not pre-empt the contract claims ("By signing the TOUs and EULAs, Appellants expressly relinquished their rights to reverse engineer"), and the 17 U.S.C. 1201(f) interoperability exception did not apply because the bnetd emulator let games reach servers without valid CD keys. That was a Battle.net emulator (hosting and protocol), not an offline loader, but it shows a US court enforced the anti-reverse-engineering clause. Other jurisdictions differ (d2d's LEGAL.md cites EU Article 8: contract terms cannot remove Article 5(3) and 6 rights, V as their claim, not checked by us).

## 7. Answers

### 7.1 What may be reused under an MIT release
Allowed (facts and own work):
- API behaviour of Windows (documented), the existence of the five clock APIs, import-table patching as a technique, proxy-DLL loading order (U, general; P0 tested it V).
- Facts about Game.exe 1.14d that we re-verify ourselves on our own copy: the imports (`IJL11.DLL` ordinals 2, 3, 5), function addresses and their behaviour (`GAME_UpdateGamesProgress` at `0x0052FC20`, which reads `timeGetTime` and `QueryPerformanceCounter`), struct offsets. Addresses are facts about a program, but a large table copied in the same arrangement from one source may count as copying its expression, and in the EU a database right may apply (9.4). So: record each address with an independent verification (procedure below), publish only what we use.
- Our own code and docs (we hold the copyright; see 9.2): `tools/lockstep/hook.c`.
Not allowed: any code or close paraphrase from d2d/whisper, libTAS, PlugY; anything from speedhack-rs or hi-angel's override; D2MOO lines (technically licensed, but excluded by policy); Cheat Engine code.

### 7.2 Clean-room procedure (3 lines)
1. Two roles: reader (may read GPL/unlicensed sources, d2d notes, D2MOO) writes only a facts sheet (name, address or ordinal, behaviour in one sentence, how a third party can re-check on their own Game.exe) and never code; writer sees only the facts sheet and writes the code, and never opens the sources; the writer is a different agent session from anyone who read `apps/whisper` or `docs/research/d2d-whisper-lan-oracle.md`.
2. Every address, ordinal, offset and constant enters the hook repo only with a `PROVENANCE.md` row showing an independent check on our own legally owned Game.exe (disassembler or live read, hash of the exe), not just "d2d says so".
3. A release check (script) greps the tree for Blizzard bytes (exe/DLL hashes, MPQ headers, tables), for GPL identifier strings and for code fragments from the reader's source list; the user or a lawyer signs the provenance review before the first tag.

### 7.3 PROVENANCE.md format (proposal)
One table per category (`Addresses`, `Imports and exports`, `Constants`, `Behaviour claims`), header block first:

```
# PROVENANCE
Target: Game.exe 1.14d, SHA-256 <hash of the author's own copy>, checked on <date>
Method key: D = static disassembly of own copy; L = live read on own copy (offline); P = public doc (Microsoft); F = fact sheet from reader (see Reader log)
Rule: a row with method F only is NOT enough to ship; it needs D or L or P.

| ID | Item | Value | Method | How to re-check (steps a stranger can follow) | Checked by / date | Other sources consulted (name, licence, what was taken: fact only) | Notes |
| A001 | frame gate function entry | 0x0052FC20 | D+L | open Game.exe at VA, see calls to timeGetTime, QueryPerformanceCounter, compare of elapsed with a target ms | writer-1 / 2026-.. | D2MOO Game.cpp (MIT): function name only | |
| I001 | IJL11.DLL imports | ordinals 2, 3, 5 | D | dump import table, filter IJL11.DLL | ... | d2decomped whisper (GPL): ordinals only, no code | |

## Reader log
date | reader | source (URL, commit, licence) | what the facts sheet took | writer never saw the source: yes/no
```
Plus a `THIRD_PARTY.md` listing anything shipped from others (ideally nothing) and `NOTICE` text from section 7.5.

### 7.4 What the release must NOT contain
- Blizzard binaries or fragments: Game.exe, D2*.dll, IJL11.DLL (the real one stays in the user's folder; the proxy must load the user's renamed copy), Storm.dll, MPQs, tables (`.txt`/`.bin` data), strings, art, sound.
- Decompiled or disassembled code, pseudo-C or listings of Game.exe, byte dumps of its code, signature patterns copied wholesale from others (pattern bytes are Blizzard code bytes; keep to the minimum needed and generate them ourselves).
- d2d, whisper, libTAS, PlugY, speedhack-rs, Cheat Engine code, or D2MOO source; d2d fork patches.
- Blizzard trademarks as product names or logos ("Diablo" in the repo name suggests endorsement: prefer a neutral name such as "d2 time hook" with a "not affiliated" line; D2MOO and d2d use such lines, V).
- Any Battle.net client, key, or protocol code; anything that enables online play; secrets and private hostnames (leakcheck).
- A pre-patched Game.exe. Distribute only our DLL and instructions.

### 7.5 Risks and the offline-only notice
Risks (U unless cited):
- EULA breach by the user (bots, modify, reverse engineer, 6.1) and by us if we reverse-engineered; account or CD-key action if used on Battle.net (precedent 6.3: key restrictions in 2008).
- Warden detection if a hooked client connects to Battle.net: undocumented for us; assume detection and a ban.
- Copyright and contract claims over address tables and any Blizzard-derived bytes (6.4).
- Trademark use in the name.
- Reputational: a "bot" tool; a scaled-time tool could be abused for online cheating, which our notice and design should discourage (for example the hook refuses to load unless the game is in single player or an explicit offline flag file exists, if that can be done cheaply; U, a design option, not verified).

Suggested notice (README top, `NOTICE`, and printed in the hook log at load):

> This project is an independent, unofficial tool. It is not affiliated with, endorsed by, or supported by Blizzard Entertainment, Inc. Diablo and Diablo II are trademarks of Blizzard Entertainment. This repository contains no Blizzard code, data, or game files; you must own a legal copy of Diablo II: Lord of Destruction 1.14d and run it from your own installation. It is for OFFLINE, single-player research and testing only. Do not use it on Battle.net or with any realm, and do not use it to gain an advantage over other players. Blizzard's terms (the End User License Agreement) restrict modifying the game and using bots or similar software; whether your use is permitted is your own responsibility. Using this tool may get an account or CD key banned and may violate your agreement with Blizzard. It is provided "as is", without warranty, and without support or promise of maintenance. Not legal advice.

## 8. Licence recommendation

MIT is workable for a hook that contains only our own code, with the clean-room rules above. Reasons: (a) we copy nothing under copyleft, so nothing forces GPL; (b) a 100-line hook has little to protect, and MIT matches D2MOO (a community that may reuse it); (c) it does not bind users who combine it with other tools (GPL would, even though our bot talks to it over a socket or file only); (d) the standalone repo keeps the GPL d2lod-bot and d2d fork untouched (ADR 0003/0010 stay true). Alternatives: Unlicense or 0BSD (no attribution, even simpler), or keep GPL-3.0-or-later for consistency with ADR 0010 (safest if any GPL-derived idea slips in; costs reuse). A permissive licence does nothing to Blizzard's rights; it only grants ours.

What would force GPL (or force not releasing):
- Copying, translating, or closely paraphrasing any d2d/whisper, libTAS or PlugY code, or building the hook as a modified version of them (GPL section 5(c): the entire work under GPL).
- Statically linking or embedding a GPL library, or shipping GPL binaries in the same package as one work.
- Re-using GPL-licensed files from this repo (d2lod-bot) that include any d2d-derived text or code. Our own authored files are fine only if the holder(s) agree (9.2).
- A "derived from D2MOO" claim: D2MOO is MIT, so it would not force GPL, but would need its notice and the "non-commercial" sentence resolved (9.3).
Things that force "do not release": any Blizzard bytes, decompiled code, or table dumps in the tree.

## 9. Open questions for a lawyer
1. Does the EULA that governs a retail or Battle.net-installed Diablo II 1.14d copy (the original LoD EULA, not the 2024 platform text) ban an offline, single-player, time-scaling hook and its public release, and does its reverse-engineering clause apply to the author's study of the exe for interoperability (Davidson v. Jung went against the engineers; EU rules differ)? Which jurisdiction applies to the releaser and to users?
2. Who holds copyright in code and docs written by AI agents for the user, and may the user relicense our own `hook.c` from GPL (in this repo) to MIT? Are all contributors identifiable and in agreement?
3. D2MOO: does "intended for non-commercial use only" next to MIT limit the grant? Are its decompiled-code origins a defect we inherit if we ever copied anything? (We plan not to.)
4. Are address tables, function names and short signature byte patterns protectable (copyright, database right, trade-secret theory) when taken from d2d/D2MOO versus independently derived? How much pattern data is safe to publish?
5. Does publishing a hook that other people can combine with a bot expose the releaser to contributory liability, or to DMCA 1201 anti-trafficking arguments (Warden is an access/anti-cheat control; a hook that is not designed to defeat it is the intended position, U)?
6. Is the clean-room reader/writer split enough evidence, given that earlier agent sessions already read `apps/whisper` and D2MOO summaries?
7. Wording of the offline-only notice and the trademark disclaimer; whether to name the target version and the game at all.
8. The original PlugY author's licence terms and whether Blizzard has ever objected to PlugY (not found in public sources I fetched).

## 10. Sources fetched (2026-11-02)
- https://github.com/psi29a/d2decomped (API, LICENSE, README.md, LEGAL.md, apps/whisper/README.md, apps/whisper/whisper.cpp, folder listings)
- https://github.com/ThePhrozenKeep/D2MOO (API, LICENSE, README.md, source/D2Game/src/GAME/Game.cpp)
- https://github.com/clementgallet/libTAS (API, COPYING)
- https://github.com/Hirtol/speedhack-rs (API, README, Cargo.toml files, root listing)
- https://github.com/hi-angel/clock_gettime_override (API, root listing)
- https://github.com/ChaosMarc/PlugY (API, LICENSE, README, listing)
- https://www.blizzard.com/en-us/legal/fba4d00f-c7e4-4883-b8b9-1b4500a402ea/blizzard-end-user-license-agreement
- https://www.blizzard.com/en-us/legal/28d5ebbf-c245-4408-8ba9-043dd5f056bf/ (Legal FAQ)
- https://kotaku.com/blizzard-sends-cease-and-desists-to-diablo-ii-resurrec-1846977935 and https://gamedeveloper.com/game-platforms/blizzard-suspends-350k-i-starcraft-diablo-ii-i-users-for-third-party-hacks- (summariser fetch)
- https://law.resource.org/pub/us/case/reporter/F3/422/422.F3d.630.04-3654.html (summariser fetch)
