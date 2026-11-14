# Build hosts: moving builds off the worker host (2026-11-14)

> status: study, nothing applied. Request (user, 2026-11-14): stop building on the worker host while the user plays; look at the NAS. Tags: V = measured or read today (command named), S = secondary (a doc or spec), U = unverified or estimate. Read-only on the NAS, the worker host, the Mac mini: no install, no container, no job on the worker host.

## 0. Short answer

1. the NAS already has most of the answer: container `d2-builder` (V: `docker inspect`: `--cpus=6 --cpu-shares=128 --memory=8g`, image `d2-builder:1` 1.52 GB, NAS-OS template `my-d2-builder.xml`, up, visible in the Docker tab, no docker.sock, no ports, user 99:100). It built d2d in 304 s cold at -j6 on 2026-11-01 vs 239 s on the worker host (S: ops `docs/builder-private.md`), i.e. about 1.27x the worker host. It has gcc, clang, cmake, ninja, python3; it has NO Go, no headless browser, no sshd (V: `docker exec d2-builder which ...`).
2. Recommended: grow that container into the general builder ("d2-builder:2": add Go 1.27.3, chromium, sshd on one high port, a Go cache dir) and teach `labrun` a host `nas`. Stage 1 is the Go whole-repo gate only.
3. Memory, not CPU, is the NAS's binding limit (11.6 GiB available, no swap). Keep the 8 GiB cap, one gate at a time.
4. TODAY (user playing): nothing on the worker host; light checks on the head box; Go vet/build and per-package tests on the Mac mini; the merge worker stays stopped.

## 1. Headroom on the NAS (all V, ssh as root with the NAS key, 2026-11-14 about 10:58 EST)

| Item | Value |
|---|---|
| CPU | i5-12600K, 10 cores / 16 threads (`lscpu`), governor performance, max 4.9 GHz, coretemp 42 C idle |
| Load | `uptime` 0.90/0.67/0.67; `vmstat 1 3` id 96-97 percent; `/proc/stat` since boot (12 days): idle 92.2 percent, iowait 0.8 percent |
| Memory | total 31.1 GiB, used 19.7, free 2.4, buff/cache 11.4, MemAvailable 11.4 GiB, swap 0 (`free -m`). ZFS ARC 2.7 GiB of max 3.9 GiB (arcstats), shrinkable, not counted in MemAvailable |
| PSI | NOT available: `/proc/pressure/*` does not exist on this kernel (6.12.85-NAS-OS). A watchdog must poll MemAvailable and load, as the existing wrapper does |
| Containers | 62 running. Top by CPU (`docker stats --no-stream`): Frigate 18 percent (1.9 GiB), chore-tracker-api 17 percent, AMP 14 percent (4.5 GiB), the forge 1.6 percent (0.76 GiB), romm 1 GiB, paperless 0.86 GiB, immich_server 0.66 GiB, tdarr 0.55 GiB; everything else under 0.5 GiB. Sum of container CPU is about 0.6 of 16 threads |
| Plex/Jellyfin | no such container running now (`docker ps`); tdarr is up (2 ffmpeg-matching procs in `ps`, idle at sample time) |
| Docker data-root | `/var/lib/docker` on `/dev/loop2` (docker.img) 150 GB, 90 GB used, 57 GB free (`df -h`); overlay2, cgroup v2, cgroupfs driver. Images 85.5 GB with 82 GB "reclaimable", build cache 21.2 GB (`docker system df`): prune before heavy image builds |
| Fast storage | ZFS pool `cache` (WD SN770 1 TB class): 928 GB, 686 GB used, 242 GB free; `/mnt/user/appdata` shows 213 GB avail. appdata of d2-builder lives at `/mnt/cache/appdata/d2-builder` (on the NVMe pool, not the array) |
| Existing builder data | d2-builder `/data`: ccache 101 MB, src 74 MB, build 514 MB, deps 198 MB |
| Other CI | `forge-runner` container (image `d2-forge-runner:13.2.0-go1.27.3`, `NanoCpus=4000000000`, memory 6 GiB, appdata 7.0 GB): already has Go 1.27.3, git, node, python3 (V: `docker exec forge-runner go version`). It serves forge workflows, not labrun |
| Compose Manager | `/boot/config/plugins/compose.manager/projects/` holds artifex-test, herpstat-exporter, immich, librarygenesis; the d2 containers (d2-builder, didle-updates, forge-runner, d2-board) are dockerMan templates in `/boot/config/plugins/dockerMan/templates-user/` and show in the Docker tab. The d2-builder precedent is the pattern to copy |

Max safe footprint: 6 CPUs and 8 GiB as already set. Reasoning: 31.1 GiB total, 11.4 available; an 8 GiB hard cap leaves about 3.4 GiB for other services' growth. The existing wrapper stops the container if MemAvailable falls under 3 GiB (S: docs/builder.md; the first build's low point was 7.4 GiB available, S). CPU is not an issue: 6 capped threads of 16 with cpu-shares 128 (weight 1/8 of a default container under contention; shares only act when CPUs are contended, which they almost never are here). Disk: 40 GB of artefacts fits the NVMe pool (242 GB free) but NOT docker.img (57 GB free): keep work dirs in the bind-mounted appdata, never in container layers.

PSI missing means no kernel-side "starved" signal; rely on cgroup limits plus the 15 s MemAvailable watchdog.

## 2. Workloads and what each needs

| W | Workload | Needs | Fit on the NAS (tag) |
|---|---|---|---|
| a1 | Go gate `go vet ./... && go test -short -p 4 ./...` | Go 1.27.3 (the head box `~/.local/sdk/go`, 283 MB: V `du`), repo mirror (rsync), Go build cache | Fits. Measured on the head box (loaded 4 cores, 6 GB): `go build ./...` whole repo cross for darwin, -p 1, 111 s wall, 0 failures from memory (V). `go test -short -count=1 -p 1 ./internal/bot`: 138 s test time, 154 s wall, largest single child RSS 287 MB (V, rusage). So `-p 4` peaks about 1.5-3 GB (U, 4 x test/link processes). The critical path of the gate is the one package internal/bot at about 140 s single-threaded (V), so the worker host's 16 cores buy little beyond -p 4 |
| a2 | `-race` of internal/bot, nav, watch | same, 2-4x memory | the worker host log says about 150 s and GBs (S, brief); U on peak. Run these one at a time with `-p 2`, under the 8 GiB cap. Needs a measured trial (see 6) |
| a3 | tests needing game MPQs (`D2_MPQ_DIR`, about 1.4 GB, the user's own copy) | the data on the builder, private | Put it in `/mnt/cache/appdata/d2-builder/gamedata` (mode 700, owner 99:100, mounted read-only at `/gamedata`). It is not in a public share; whether appdata is exported by Samba is U (check the share's export setting in the UI before staging). Today the data is on the worker host (the head box has no copy: `~/d2d-lab/data` is absent here, V) |
| b1 | d2d fork C++ (cmake + ninja + gcc/clang) | toolchain, ccache, deps | Already V on the NAS: 304 s cold, 15 s warm, test binaries 46 s, sha256 identical to the worker host build (S: builder-private.md). Wrapper `tools/ops.sh run build-fork.sh <branch>` exists |
| b2 | docker image builds (didle:*, d2dbot:devp*, about 430 MB, ubuntu base) | a docker daemon; `context.sh` + `docker build --cpu-shares 2` | The NAS-OS daemon is the only one on the NAS. Builds are fine technically (57 GB free in docker.img, prune the 21 GB cache), but giving a container docker.sock is root on the NAS: not recommended. Option: a restricted ssh key whose forced command is a small script (`imgbuild`, fixed `docker build --cpu-shares 2 --memory 4g`, whitelisted tags, context only from `/mnt/cache/appdata/d2-builder/ctx`). Images then show in the Docker tab under Images. U: not tried. Stage 3 |
| c1 | headless d2d runs, ctest | game data, d2d, 1-1.5 cores per run, about 1 GB | Fits 2 at a time inside 6 CPUs / 8 GiB (U). Needs game data staged (a3) |
| c2 | emulator oracle (`~/d2d-lab/emu` 1.6 GB: python venv with unicorn, pefile, capstone; game.exe from the user's install) | the user's game files, python | Fits (python, CPU only). The tree is on the worker host (not readable from here: U size). Same private staging as a3. Stage 3 |
| d | headless Chrome for `tools/uishots` | chromium in the image (+about 300 MB), fonts | Fits; U (no Chrome in `d2-builder:1`, V). Stage 2 |
| e | friend drills (docker load of the pack, throwaway volumes) | a daemon and volumes | Same as b2 (needs the host daemon); do on the NAS only with the restricted-key script; else the Mac mini (Docker unknown there: U) |
| f | the sandbox `d2play-20261108` | see section 4 | candidate, separate decision |

## 3. Design: how labrun reaches the NAS

What labrun does (V, read `~/.local/bin/labrun`): `HOSTS` is a Python dict (entries: `ssh` alias, `os`, `shell`, `env`, optional `mac` for wake); `run_remote` rsyncs the repo (`--delete --filter=:- .gitignore`) to `<remote $HOME>/labrun/<this host>/<path under $HOME>` and runs `cd there && <shell> '<cmd>'` as a job under `~/labrun/.jobs/<id>`. The hook on the head box blocks heavy local commands and names a `labrun` line. The merge queue gates run `labrun $D2_LABRUN_HOST -- ...` (V: `local.env` has `D2_LABRUN_HOST=worker`; mergeq, plan-agent and api-agent forward it).

Adding `nas` (recommended arrangement):

1. Image `d2-builder:2` (Dockerfile in ops `builder/`): add Go 1.27.3 (copy of the head box SDK or the official tarball, in `/opt/go`), chromium, openssh-server (non-root sshd on port 2222, key-only, `AuthorizedKeysFile /data/ssh/authorized_keys`, host keys in `/data/ssh`), rsync (already there). `GOCACHE=/data/gocache`, `GOMODCACHE=/data/gomod`, `GOFLAGS=-p=4`. NAS-OS template gets one Port config 2222 and the Go paths; limits unchanged except an optional `--memory=8g --cpus=6`.
2. the head box `~/.ssh/config`: `Host nas-build` -> the NAS hostname, port 2222, user `nobody` (99), a dedicated key (not the root NAS key).
3. labrun: add to `HOSTS`:
   `"nas": {"ssh": "nas-build", "os": "linux", "shell": "bash -lc", "env": {"PATH": "/opt/go/bin:$PATH", "D2_MPQ_DIR": "/gamedata"}, "desc": "NAS-OS NAS builder container: capped 6 cpu / 8 GB, low priority, Go gates and fork builds"}`.
   Do not add it to the auto-routing regexes (it must be chosen explicitly: it is small). Source of truth: the brief names `~/code/the head box-box/labrun`, but the file in that checkout does not exist here (V: `ls`); the live copy is `~/.local/bin/labrun` (a Python script). U: which copy Syncthing `claude-bin` syncs; change the synced one only.
4. Work dirs: `/mnt/cache/appdata/d2-builder/labrun/the head box/code/<repo>` (the container HOME is `/data`), on the NVMe pool. The `.jobs` dir sits next to it; jobs survive an ssh drop.
5. Concurrency: 1 gate at a time. The d2-builder runner already has capacity 1 (S); labrun has no queue, so wrap the gate in `flock /data/gate.lock nice -n 10 ...` (a second job waits, bounded). Whole gate at `-p 4` fits 6 CPUs; the race run is a second step in the same lock.
6. Merge queue: `local.env` `D2_LABRUN_HOST=nas` (one line). Timeouts: mergeq should allow 15 min for the first (cold-cache) gate; U for its own limit.
7. Game data staging: rsync once from the worker host (`rsync -a --chmod=D700,F600 worker:~/d2d-lab/data/ nas:/mnt/cache/appdata/d2-builder/gamedata/`, run when the user is not gaming, about 1.4 GB, trivial I/O). The container mounts it `:ro`. No copy in any shared share, git, or image.
8. Safety: the gate wrapper (existing, S) kills the container at MemAvailable < 3 GiB; add `timeout 25m` around a gate; builds under `nice -n 10`; `--pids-limit=2048`, `--blkio-weight=100` already set.

## 4. The sandbox (d2play-20261108) on the NAS

- Today it runs on the worker host (V: `ssh worker ps`: d2d 33 percent, d2bot 25 percent CPU, container `d2play-20261108`, image `d2dbot:devp7`; compose in `~/play-build-20261108` on the worker host, not visible from the head box). Per the brief about 1.5 cores and 1 GB continuous (S); the sandbox hero is the user's real playing hero.
- The "services on the NAS" rule favours moving it; cost on the NAS: +1.5 cores (of 16, about 10 percent) and +1 GB of 11.4 available, continuous. That is acceptable but it is a SERVICE, not a build: it must be an ordinary visible template/compose stack (name `d2play`), not part of d2-builder, and must not share the builder's cap (a gate must never starve the hero).
- Migration (not done): (1) stop the container on the worker host only after the user agrees; (2) copy the volume `d2play-data` (size U: `docker system df -v` on the worker host; the worker host shows 30 GB of local volumes in total) with `docker run --rm -v d2play-data:/v alpine tar c /v | ssh nas docker run -i ... tar x` into `/mnt/cache/appdata/d2play/` (a bind mount on the NVMe pool so it is backed up and visible); (3) `docker save d2dbot:devp7 | ssh nas docker load`; (4) game mount: the private data of 3a; (5) register as a template (copy `my-didle-updates.xml` as the pattern); (6) the watch UI moves from `:8891` on the worker host to a the NAS port: update the Electron viewer address, the friends feed and `tools/dev-update.sh` / `up.sh` / `build3.sh` (D2_LABRUN_HOST for dev-update stays whichever builder host builds the image; image builds still need a daemon: on the NAS via b2); (7) d2board and the status links.
- Risks: the hero's save is the only copy (copy first, verify by hash, keep the worker host's volume until a week of runs); a timing difference (another CPU, V governor performance) could change nothing in d2d replay determinism (S: cross-host determinism tool `d2dfarm det`), U; ZFS/NVMe latency for saves is fine. Recommend as stage 4, after the builder is proven, and ask the user (decision D4).

## 5. Alternatives, ranked

1. the NAS `d2-builder` container (above): 6 threads of an idle 16-thread box, no impact on the gamer, visible in the UI, precedent working. Rank 1.
2. Mac mini M4 (the Mac mini, V: `ssh mac-mini`: 10 cores, Go 1.26.3 darwin/arm64 at `/opt/homebrew/bin/go`, `labrun hosts`: load 1.6, memfree 6 GB of 24). Good for `go vet`, builds and portable tests. NOT a full replacement: the whole repo cross-builds for darwin/arm64 except two packages (V, `GOOS=darwin GOARCH=arm64 go build -p 1 ./...`, 111 s): `internal/sentinel` (undefined openPanels, crashProcs, helperCmd, Observer, ignoredWin, dialogRe) and `cmd/inputcal` (input.FindGamePID, input.OpenProc). Linux-only files by `_linux.go` suffix/tags exist in launch, d2dreplay, d2dview, setup, cheat. Whether `go test -short` passes on darwin: U (several tests read `/proc`); the repo's go.mod says `go 1.24`, the Mac mini has 1.26.3 (the toolchain 1.27.3 would be downloaded: U). Use for vet and package tests of packages that build there; not for the merge gate until a trial shows the same pass set. Also the only host for Apple builds (V: desktop-electron.md)
3. the worker host in a user-away window, capped: on the worker host `systemd-run --user --scope -p CPUQuota=400% -p CPUWeight=20 -p Nice=19 -- <cmd>` (cgroup v2 user delegation: U) as a cheap interim, only when the user says he is not playing. Interim only.
4. the head box (4 cores, 4 GB, V `labrun hosts`: load 11, memfree 4 GB): per-package `-p 1` tests (V: internal/bot passes in 154 s at 287 MB). Not whole-repo.
5. Cloud/CI (forge runner `forge-runner` on the NAS with Go 1.27.3, or GitHub-hosted): possible for the vet/short-test part as a workflow job; the forge runner is capped at 4 CPUs / 6 GiB and shares the NAS, so it is the same arrangement as rank 1 with less control. Not needed.

## 6. Cost and benefit

- Gate time (estimate, U until a trial): the worker host about 2-3 min (S, brief). the NAS: the floor is the internal/bot test at about 140 s on a loaded the head box (V); the 12600K P-core at 4.9 GHz is likely at least as fast per thread as the worker host's cores, but the C++ ratio measured 1.27x slower (V from the first-build record, ccache-less, nice 10, cpu-shares 128 competing with the game's host load at that time). Expect 3-5 min cold-cache first gate (go build cache empty: add about 1-2 min), 2.5-4 min warm. A Go build cache on the NVMe pool makes later gates fast.
- Trial to take (cheap, no game): once the image exists, run the gate three times and `-race` once under `docker stats` + `free -m` sampling every 5 s; record wall time and peak memory in this doc. Read the d2-builder watchdog log.
- Effect on the NAS: CPU 7 percent busy on average, a gate adds up to 6 threads for minutes; low shares keep Frigate/tdarr/Immich first. Memory: up to 8 GiB hard cap, the rest unchanged; the watchdog stops it before the NAS runs short. Network/IO: rsync of about 100 MB repo mirror, ZFS on NVMe. Power/heat/noise: S estimate for 6 loaded threads of a 125 W-class part: about +50 to +80 W for a few minutes per gate; idle coretemp 42 C (V); fans are U (a NAS in the house; the C++ build on 2026-11-01 produced no report of noise: U).
- Benefit: the worker host's 16 cores stay free for the user's game; gates become independent of the gaming PC being on/awake; nice parallel capacity for builds (the worker host remains for GPU and live drills).

## 7. What else loads the worker host that is not ours (V: one `ps -eo pcpu,rss,etimes,comm --sort=-pcpu` at about 11:00 EST, plus `labrun hosts`: load 14.1 on 16 cores)

| Process | CPU | Notes |
|---|---|---|
| `Main` | 322 percent, 5.1 GB RSS, running 11 min | the user's game (name only; the user said he is playing) |
| `python3` | 98 percent, 0.74 GB, running 15 h | ComfyUI on :8188 (the user's own; the brief says so) |
| `kwin_wayland`, `Discord` | 10 percent each | desktop |
| `d2d` + `d2bot` | 33 + 25 percent | OUR sandbox hero (container d2play-20261108) |
| `d2bot.test` | 13 percent, running 47 s | OUR test run in progress at that moment (some agent or gate): confirm the merge worker and agents are all stopped |

The user can decide on ComfyUI (a pure 1-core load, 15 h): stopping it frees one core; it is not ours.

## 8. Staged plan and decisions

Decisions needed (recommendation first):

- D1. Make `d2-builder` the general builder (add Go, chromium, sshd, labrun host `nas`), not a second container. Recommend yes: one visible container, existing limits and watchdog.
- D2. Game data on the NAS: stage the MPQs privately in the builder's appdata (mode 700, read-only mount) so table tests and the emulator can run there. Recommend yes, with the Samba export of appdata checked first; no if appdata is shared. Alternative: those tests stay on the worker host, run only when the user is away.
- D3. Docker image builds on the NAS via a restricted-key forced-command script (not docker.sock). Recommend yes, stage 3, after prune of the 21 GB cache.
- D4. Move the sandbox (d2play) to the NAS. Recommend later (stage 4), after 1-2 weeks of builder use; it is a service under its own template.
- D5. Memory cap: keep 8 GiB (or lower to 6 GiB if the race trial shows under 5 GiB peak). Recommend keep.
- D6. Interim: allow the worker host capped windows only when the user says he is away. Recommend yes, explicit only.
- D7. Mac mini for vet: allow it as an extra labrun target for vet/package tests (not the merge gate). Recommend yes.

Stage 1 (smallest useful: Go gates on the NAS in the existing capped, visible container). Steps, to run only after the user's go; nothing here was applied:

1. Add to ops `builder/Dockerfile` (as `d2-builder:2`):
   ```
   FROM debian:13   # existing base
   ...existing apt list... + openssh-server rsync chromium fonts-liberation
   ARG GO=1.27.3
   RUN curl -fsSL https://go.dev/dl/go${GO}.linux-amd64.tar.gz | tar -C /opt -xz   # or COPY the head box SDK
   ENV PATH=/opt/go/bin:$PATH GOCACHE=/data/gocache GOMODCACHE=/data/gomod GOTOOLCHAIN=local
   COPY sshd_config /etc/ssh/sshd_config.d/10-d2.conf   # Port 2222, PasswordAuthentication no, HostKey /data/ssh/host_ed25519, AuthorizedKeysFile /data/ssh/authorized_keys, PidFile /tmp/sshd.pid
   ```
   The entrypoint additionally starts `sshd -D -e` (non-root works with a high port and `UsePAM no`). Build on the NAS with the existing `builder/build.sh` (low priority).
2. Template draft (the existing XML with these changes; apply with `bash /mnt/user/appdata/_nas-ops/tpl-apply.sh d2-builder`):
   ```xml
   <Repository>d2-builder:2</Repository>
   <ExtraParams>--restart=unless-stopped --user 99:100 --cpus=6 --cpu-shares=128 --memory=8g --memory-swap=8g --pids-limit=2048 --blkio-weight=100</ExtraParams>
   <Config Name="SSH" Target="2222" Default="2222" Mode="tcp" Description="labrun ssh (key only)" Type="Port" Display="always" Required="true" Mask="false">2222</Config>
   <Config Name="Game data (private)" Target="/gamedata" Default="/mnt/cache/appdata/d2-builder/gamedata" Mode="ro" Description="private MPQs, mode 700" Type="Path" Display="always" Required="false" Mask="false">/mnt/cache/appdata/d2-builder/gamedata</Config>
   ```
   (keep the existing Data and TZ entries). Compose Manager is NOT used: the container is a dockerMan template like the other d2 services, so it is in the Docker tab.
3. `mkdir -p /mnt/cache/appdata/d2-builder/ssh` on the NAS; put the new the head box public key into `authorized_keys` (a dedicated key, `ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519_towerbuild`), mode 600, owner 99:100.
4. the head box: `~/.ssh/config` entry `nas-build` (port 2222, user nobody); add the `nas` entry to `labrun` (section 3); `labrun hosts` should show it up.
5. Trial (no game, user may be playing): `labrun nas -- sh -c 'go vet ./... && go test -short -count=1 -p 4 ./...'` from a worktree, three runs; sample `docker stats` and `free -m` every 5 s from the head box; record in this doc.
6. If green: set `D2_LABRUN_HOST=nas` in the coordinator's `local.env`, restart the merge worker; keep the worker host as a documented fallback by one line.

Stage 2: chromium shots (`tools/uishots`) and `-race` runs; the MPQ staging (D2). Stage 3: fork builds via the container (already works), image builds via the forced-command key, friend drills, emulator oracles, hero runs. Stage 4: the sandbox.

## 9. TODAY (the user is playing)

- the worker host: nothing from us. The merge worker is stopped (V, per the brief). The `d2bot.test` seen at 11:00 EST (13 percent CPU) is ours: whoever owns it should finish or be stopped by its owner; the sandbox hero (d2d + d2bot, about 0.6 cores now) stays.
- the head box: docs, per-package `-p 1` tests (V: 154 s for internal/bot), `gofmt`, contracts and backlog checks, leakcheck.
- Mac mini (the Mac mini): `go vet` and `go build` for packages that build on darwin (V list above); `labrun mac-mini -- sh -c 'go vet ./internal/... '` on an agreed set. Not the merge gate.
- the NAS: nothing new until the user's go; the existing d2-builder can build the d2d fork today with `tools/ops.sh run build-fork.sh <branch>` (V, 304 s cold).
- Whole-repo Go gates are deferred or run after the user stops playing.

## 10. Not done (U)

- No measured Go gate on the NAS (no Go there); peak memory of `-p 4` and `-race` are estimates.
- Whether appdata is exported by Samba; whether the labrun copy in a the head box-box checkout is synced; the worker host's `d2play-data` size; Docker on the Mac mini; `go test -short` on darwin.
