# How to run d.idle

> status: current as of 2026-11-01 (card readmedocs). Every example says whether the author ran it. "Ran in this doc's worktree" means a real run on a Linux machine with Go and NO game install. Everything that needs a game says "not run by the author". Older user guides for the Wine path ([quickstart.md](quickstart.md), [troubleshooting.md](troubleshooting.md)) are marked stale in their own banners; this page is the current entry point. What the build can do: [capabilities.md](capabilities.md).

## What this will NOT do

It never connects to Battle.net, realms or chat, it plays offline single player only, and it contains no game files and no
d2d code. You bring your own legitimate Diablo II: Lord of Destruction 1.14d install. Nothing here is a way to get the game.

## Two kinds of reader

- **You only want to look around** (build, tests, replay a recorded soak, see the dashboard on simulated data): everything in
  sections 1 to 4 works with Go alone.
- **You want the bot to play** against a game: sections 5 to 8. The verified path needs a d2d binary built with the
  maintainer's private lab patches (virtual clock, state dump), which is not published. That part is the maintainer's private lab
  setup; the public starting point is the upstream project, [d2decomped](https://github.com/psi29a/d2decomped). Whether the
  bot works against an unmodified upstream build is not verified ([capabilities.md](capabilities.md), known gaps). The older Game.exe path under
  Wine is frozen ([adr/0013-retire-game-exe-gates.md](adr/0013-retire-game-exe-gates.md)).

## 1. Requirements

| need | detail |
|---|---|
| Go | 1.27 or newer; the standard library only, nothing is downloaded (`go.mod` says `go 1.24`, which is only the language floor) |
| For playing: your own Diablo II LoD 1.14d install | the three archives `d2data.mpq`, `d2exp.mpq` and `Patch_D2.mpq` are read for tables; d2d reads the MPQ directory too (`--d2d-data` or `D2D_DATA`). Never copied into this repo ([../THIRD_PARTY.md](../THIRD_PARTY.md)) |
| For playing: a d2d binary | see the two kinds of reader above; build notes for GCC 16 in section 9 |
| Wine or Proton | only for the frozen Game.exe backend ([quickstart.md](quickstart.md)); not needed for d2d |
| Linux x86_64 | what the author used; other systems are not verified |

## 2. Build

Ran in this doc's worktree (these two binaries, plus `d2dash`, built clean):

```sh
go version
go build ./...                           # everything; on a small machine build only what you need:
go build -o bin/ ./cmd/d2bot ./cmd/d2dsoak
```

The programs in `cmd/` (each has `-h`; the list changes, `ls cmd` is the truth):

| program | what it is |
|---|---|
| `d2bot` | the bot and its dashboard (`--backend game` or `--backend d2d`) |
| `d2dsoak` | runs and judges long soaks on d2d; `compare` and `extract` read minute logs offline |
| `d2drill` | the live drill suite (needs d2d and a game install) |
| `d2dview` | tap viewer for a d2d session: state, frames and a sanitized public page |
| `d2dash` | the dashboard against simulated bots, no game |
| `d2tables` | extracts the game tables from your MPQ files into a local cache |
| `d2doctor` | pre-flight checks of the machine and install |
| `d2cheat` | logged test cheats, throwaway lab characters only ([cheats.md](cheats.md)) |
| `d2fleet`, `d2soak`, `d2watch`, `d2input` and others | fleet supervisor, Game.exe-era soak and viewer, input helper, lab and probe tools |
| `changecheck`, `leakcheck` | repository hygiene checks used by CI |

## 3. Offline dry run (no game)

Ran in this doc's worktree. The repository contains the minute log of a real soak; `d2dsoak compare` turns it into the
verdict table. This exercises the judging code, not the bot.

```sh
gunzip -c docs/probes/soak-logs/postfix2-minutes.jsonl.gz > bin/postfix2.log
bin/d2dsoak compare -hours 21.25 -baselines tools/d2dsoak/baselines-early.json bin/postfix2.log
```

First lines of the real output, trimmed:

```
metric                     d2d mean        sd      n |     baseline (min .. max)          |      gap       tol  verdict
xp_per_hour                    5437         0      1 |         9172 (6600 .. 11743)       |   -40.7%      1834  INCONCLUSIVE [too few seeds]
...
INVARIANTS: VIOLATED(...: BELT,NOKILL)
OVERALL: INCONCLUSIVE (exit 2)
```

Exit 2 is the verdict, not a crash: one seed cannot pass or fail, and the run violates two Gate 1 criteria
([probes/postfix-soak2.md](probes/postfix-soak2.md)). Exit codes are in [d2d-soak.md](d2d-soak.md). Also ran:
`go test -count=1 ./internal/d2dsoakstat/` (passes), and the simulated dashboard:

```sh
bin/d2dash -addr 127.0.0.1:8420 &      # then open http://127.0.0.1:8420/ or: curl -s 127.0.0.1:8420/api/v1/status
```

The status reply came back with a simulated Sorceress named Testy fighting in Blood Moor. Stop it with `kill %1` (it is your own
background job). To run the bot's simulated runner (`d2bot --sim`) you first need the extracted tables (section 4): without
them it stops at once, which the author ran: `d2bot: gear: tables not extracted: run go run ./cmd/d2tables --data <MPQ dir> first`.

## 4. Extract the game tables (needs your game files)

Not run by the author (no install on the machine used for this page). Details and the cache layout: [data.md](data.md).

```sh
go run ./cmd/d2tables --data <game-dir>          # dir holding d2data.mpq, d2exp.mpq, Patch_D2.mpq
# or set D2_DATA_DIR=<game-dir> and let d2bot extract on first start
```

The cache (about 1 MB, a trimmed copy of a few tables) is written under your own user cache directory (`--out` or `D2_TABLES_DIR`
to change it); it is never committed. A cache built before a column was added is no longer used silently: with the install present it is rebuilt,
and without one the load fails naming the missing column, for example `has no column skills.auralencalc` ([fixbuff.md](fixbuff.md), offline test only).

## 5. Run the bot on the d2d backend

Not run by the author: needs a game install and a d2d binary. Flags are the real ones from `d2bot -h`; every one has an
environment twin. `<d2d-binary>`, `<game-dir>` and `<character>` are yours.

Attach to a d2d that is already listening (the bot never stops it):

```sh
bin/d2bot --backend d2d --attach --d2d-attach --d2d-sock <socket> --goals 'farm:Blood Moor' --addr 127.0.0.1:8420
```

Managed: the bot starts d2d, drives it, then saves and quits it on a stop. This is the recipe the soaks used (throwaway
character, lockstep clock, fixed seed; [d2d-baselines-early.md](d2d-baselines-early.md) section 4):

```sh
D2D_RUN=on bin/d2bot --backend d2d --managed --d2d-bin <d2d-binary> --d2d-data <game-dir> \
  --d2d-class 4 --char <character> --seed 5 --toggle autoloot=off,autoparty=off \
  --d2d-args "--lockstep" --d2d-linkwait 60s --goals campaign --autostart --run-for 120m
```

`--d2d-class 4` is a new Sorceress (d2d's list order); `-1` loads a saved `--char`. `--no-save` makes a throwaway character that is
never written. Add `--supervise` to restart the bot after a crash. More: [d2d-bot.md](d2d-bot.md), knob defaults in [knobs.md](knobs.md).

Goals are `;`-separated (`campaign:normal_act1`, `reach:30`, `farm:Blood Moor`, `idle`; see [quickstart.md](quickstart.md) section 5, which was checked
against the code). Only early Normal Act 1 has ever run unattended ([capabilities.md](capabilities.md)).

## 6. A short drill, and a soak

Not run by the author (both need d2d, a game install and the private lab build). The drill harness takes a group name and its
own d2d paths; its flags are in `d2drill -h` (some defaults point at the maintainer's lab: always pass your own):

```sh
go run ./cmd/d2drill -h
go run ./cmd/d2drill -group D1 -bin <d2d-binary> -gamedir <game-dir> -cache <tables-cache>
```

Results format and the drill list: [probes/drills/README.md](probes/drills/README.md), [primitives.md](primitives.md). A soak is the
managed command of section 5 with `--run-for`, judged afterwards by `d2dsoak extract` and `d2dsoak compare` as in section 3
(`d2dsoak run` can run several seeds, see [d2d-soak.md](d2d-soak.md)). The maintainer's lab wraps both in slot, lock and
process-registry scripts that need the private lab setup; the public tree has no replacement for those.

## 7. Watching it

- Dashboard: `http://127.0.0.1:8420/` by default (`--addr`; a non-loopback address needs `--token`). Status API: `/api/v1/status`,
  health: `/api/v1/bot/health` ([api.md](api.md)).
- `d2dview` sits between the bot and d2d and serves a state and frame page and a sanitized public feed (`/api/pub`);
  usage in [d2d-view.md](d2d-view.md). Not run by the author.

## 8. Logs and stopping cleanly

- The bot logs to its terminal; with `--supervise`, soak runs and the d2dsoak tools write `bot.log`, `d2d.log` and a minute-line log
  per seed (the `minute` lines are what `d2dsoak` reads). Per-instance state (goals, deaths) lives under `--data`.
- Stop with Ctrl-C or SIGTERM and wait: the bot finishes its step, goes to town, saves and quits the game (about a minute at most).
  `--run-for` does the same on a timer. A kill -9 while a game runs leaves the game behind, and the unsaved progress is lost.
  Do not close the game window yourself. (Written from [quickstart.md](quickstart.md) section 7, not re-run.)

## 9. Troubleshooting (only what the docs record)

1. **`gear: tables not extracted`**: run the extraction of section 4 first. Ran: this is what `d2bot --sim` prints without the cache.
2. **Stale tables cache** (it once left armor recast timing silently wrong): a load without the install now fails naming the missing column;
   rerun `d2tables --force` with your install ([fixbuff.md](fixbuff.md), proven by a unit test, not re-run here).
3. **d2d does not build on GCC 16**: configure with `-DD2D_WERROR=OFF` (two false-positive warnings in d2d's own code stop the default
   build); recorded in [d2d-spike.md](d2d-spike.md) and [d2d-fork.md](d2d-fork.md).
4. **A managed start finds a d2d already answering on the socket**: it refuses; a crashed earlier run left it. Use `--supervise`
   (it sets `--clean-start`) or stop the old d2d yourself ([d2d-bot.md](d2d-bot.md)). `--attach` alone is refused: add `--d2d-attach`.
5. **Nothing happens after start**: goals are queued but the bot waits for Start; pass `--autostart` or press Start in the dashboard
   ([troubleshooting.md](troubleshooting.md) section B). The character stays in the cave mouth of the Den of Evil: a known open bug
   ([capabilities.md](capabilities.md)).

Open problems seen in soaks (poverty loops, early deaths) are not troubleshooting items; they are findings in
[probes/postfix-soak2-analysis.md](probes/postfix-soak2-analysis.md).
