# Orion as a second worker for survival waves and Proving Ground

> status: current (2026-11-20). The user allows Orion for heavy d.idle work until 9 am EST on 2026-11-20 only; after that stop and clean up (section 6). Orion is also the user's desktop and runs the sandbox container `d2play-20261108` (hero, ports 8891 and 8892): never stop, restart or modify that container, its image, overlays, ports or data, never take screenshots of the desktop, never search from `/` or `/mnt`.

Tags: V verified in this run (how named), S secondary, U unverified.

## 1. What the builder harness needs, and where Orion gets each part

| input | builder | Orion (this setup) |
|---|---|---|
| repo at origin/main | synced worktree | `~/d2lod-orion-work/repo` (`git archive origin/main`, no `.git`, `.synced-from` holds the hash) |
| bot, prep, report, scenario binaries | `go build` per cmd | `~/d2lod-orion-work/bin/{d2bot,prep.abmerc,abreport,d2scenario}`; Go 1.27.3 at `~/.local/bin` (V: build took seconds) |
| lab d2d with `--lab-verbs` and the dump `hero`/`weapons` keys | `/data/d2d-lab/survhold/bin/d2d.lab` | `~/d2d-lab/orionwave/bin/d2d.lab`, a copy of `~/d2d-lab/fork/bin/d2d.survtune-base-fabae935-linuxbuild` (V: strings show `lab-verbs` and `"hero"`; it is the binary of the Nov 16 tuning waves on Orion). Whether it is the same commit as the builder's `d2d.lab`: U (the builder was too congested to read its hash; compare before treating A/B numbers across hosts as one batch) |
| game data (MPQs, the user's own licence, never copied elsewhere) | `/data/d2d-lab/data` | `~/d2d-lab/data` (9 MPQs, 1.4 GB, already there; `D2D_DATA` default of `abrun.sh`) |
| slots | `tools/d2dslot.sh`, pool in `~/d2d-lab/slots` | same script copied to `bin/`; slots 0..3 free of admission, ports 9100+n (bot) and 9200+n (view), clear of 8891/8892 (V: `d2dslot.sh status`) |
| Xvfb, Wine | not used | not used: the lab d2d is a headless native binary (V: smoke ran with no display) |

No credentials, CD keys or the user's backup zip are involved; nothing was read from them.

## 2. One-time setup (done 2026-11-20; repeat only to refresh)

```
# from the d2lod-bot worktree on herdr (exact tree of origin/main, or of your own branch: replace origin/main)
git fetch origin && H=$(git rev-parse --short origin/main)
ssh orion.lan.birbhub.com 'mkdir -p ~/d2lod-orion-work/repo'
git archive origin/main | ssh orion.lan.birbhub.com "tar -x -C ~/d2lod-orion-work/repo && echo $H > ~/d2lod-orion-work/repo/.synced-from"
ssh orion.lan.birbhub.com 'cd ~/d2lod-orion-work/repo && export PATH=$PATH:$HOME/.local/bin
  for p in "d2bot ./cmd/d2bot" "prep.abmerc ./tools/ab/prep" "abreport ./tools/survival/abreport" "d2scenario ./cmd/d2scenario"; do
    set -- $p; nice -n 19 go build -o ~/d2lod-orion-work/bin/$1 $2 || exit 1; done
  B=~/d2d-lab/orionwave/bin; mkdir -p $B
  cp tools/d2dslot.sh tools/agentrun.sh $B/
  cp ~/d2lod-orion-work/bin/d2bot $B/d2bot.survab; cp ~/d2lod-orion-work/bin/prep.abmerc ~/d2lod-orion-work/bin/abreport $B/
  cp ~/d2d-lab/fork/bin/d2d.survtune-base-fabae935-linuxbuild $B/d2d.lab; chmod +x $B/*'
```
Work dirs are under `~/d2d-lab/<DIR>` because `wave.sh` and `abrun.sh` build that path from the `DIR` argument; use one `DIR` per agent (`orionwave` is the smoke dir). Nothing is written to the sandbox volume, `/data/didle` or any `devp` dir.

## 3. Start a survival wave on Orion

`wave.sh` must find `bin/{d2bot.survab,prep.abmerc,d2d.lab,d2dslot.sh,agentrun.sh}` in `~/d2d-lab/DIR/bin`. The agent's own branch decides which bot it tests: replace the sync source in section 2 by the agent's branch (`git archive agent/<branch>`) and use its own `DIR` (for example `survhold4o`), then:
```
ssh orion.lan.birbhub.com 'cd ~/d2lod-orion-work/repo && nice -n 19 tools/agentrun.sh <agent> --ttl <SECS> -- tools/survival/wave.sh <DIR> <HOURS> <SPEED> <CLASS> legacy:107 margin:107 ... > ~/d2lod-orion-work/out/<name>.log 2>&1'
```
`HOURS` virtual hours, `SPEED` 8 (the smoke value; the wall limit is `HOURS*60/SPEED*1.6+10` minutes per run), `CLASS` 2 Paladin or 4 Sorceress, pairs `arm:seed`. All pairs run at once (`AB_MAX` = number of pairs): on Orion keep that at 3 or fewer while the user is on the desktop. Results: `~/d2d-lab/<DIR>/ab/<arm>/<seed>.log`, last line `writes after start: N` (must be 0); then
```
ssh orion.lan.birbhub.com '~/d2d-lab/<DIR>/bin/abreport ~/d2d-lab/<DIR>/ab'      # AB_CLASS=4 for the Sorceress
```
Fetch logs with `rsync -a orion.lan.birbhub.com:d2d-lab/<DIR>/ab/ <local>/ab/`.

## 4. Start a Proving Ground scenario on Orion

```
ssh orion.lan.birbhub.com 'cd ~/d2lod-orion-work/repo && export PG_DIR=~/d2lod-orion-work/pg PG_D2D_BIN=~/d2d-lab/orionwave/bin/d2d.lab PG_DATA=~/d2d-lab/data \
  && nice -n 19 tools/agentrun.sh <agent> --ttl 1800 -- tools/provingground/run.sh -id <scenario> -class Paladin -seed 11 -parallel 1 > ~/d2lod-orion-work/out/<name>.log 2>&1'
```
`-tier fast` runs the merge-gate subset, `-dry-run` only lists. `run.sh` builds `d2scenario` and `d2bot` into `PG_DIR/bin` from the synced tree and writes `provingground-out/<stamp>` inside the repo copy. The three `PG_*` variables are required: the defaults point at `/data/...` paths of the builder.

## 5. What was verified (this run, Orion)

- V smoke wave: `wave.sh orionwave 0.17 8 2 legacy:107` (Paladin, 0.17 virtual hours = 10.2 virtual minutes) ran 33 s wall, exit 0, `writes after start: 0`, hero line `life 55/55 def 10 dmg 3-6 ar 65`, identical to the seed 107 hero of the Nov 16 Orion wave (`survtune2`). The bot log ends with `virtual run length reached`.
- V repeat run: a second identical run also ended exit 0 with writes 0, but its mid-run status differs from the first (position and state at the cut-off). So two runs of one seed on one host are not bit-identical; U whether Orion versus the builder differs more. Virtual time (lockstep clock) makes the run length independent of CPU speed (V: both runs stopped at the same virtual length, 612 s); the bot's decisions under `--run-for` wall limits and the 8x clock are U across hosts. Do not read a single paired difference from mixed hosts: run both arms of a pair on the same host.
- `abreport` printed `no usable logs` on the smoke (needs both arms and enough virtual time), as expected for one run.
- V Proving Ground: `run.sh -id heal.town -class Paladin -seed 11 -parallel 1` on Orion: GREEN, 1.0 virtual minute; `-dry-run -tier fast` lists the fast tier.
- V isolation: the sandbox container `d2play-20261108` was `Up` and healthy before and after; no process of the smoke is left (`agentreap.sh orionworker` found nothing, `agentreap.sh --list` shows only the sandbox updater of another owner). Orion load at start 4.4 on 16 cores, disk 202 GB free on the NVMe.

## 6. Cleanup and the window

- Every process goes through `tools/agentrun.sh <agent> --ttl ...`; before you report: `ssh orion.lan.birbhub.com 'cd ~/d2lod-orion-work/repo && tools/agentreap.sh <agent> && tools/agentreap.sh --list'` (the list must hold none of yours). Never `pkill` or `killall`, never kill a pid you did not start.
- Capacity: `tools/capacity.sh` hangs; use `labrun orion -- top -bn2 -d1` (idle plus nice at least 25 percent), keep to 3 concurrent runs in total across agents, `nice -n 19`.
- After 9 am EST on 2026-11-20 start nothing new. Optional removal of this setup: `rm -rf ~/d2lod-orion-work ~/d2d-lab/orionwave` (both are ours only; the bulky per-run `data-*` dirs of finished waves can go first).
- Do not use Orion's `~/d2d-lab/survab*`, `survtune*` or `fork/` for new runs: they are old work dirs of earlier waves.
