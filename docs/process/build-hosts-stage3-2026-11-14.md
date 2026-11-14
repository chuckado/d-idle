# Build hosts, stage 3 results (2026-11-14): image builds on the NAS

> status: applied and drilled. Follows `docs/process/build-hosts-2026-11-14.md` (study, decision D3) and `docs/process/build-hosts-stage1-2026-11-14.md`. User decision 2026-11-14: D1 to D7 accepted, no builds on the user's PC any more. Tags: V = measured or read today (command named), U = unverified. Roles only: "the NAS", "the builder" (container `d2-builder`), "the head box".

## 1. Design (V: drilled, section 4)

No docker socket is given to any container. The NAS daemon is reached only through one restricted ssh key whose forced command is a root-owned script, `deploy/nas/imgbuild.sh` (installed on the NAS next to the template helper, not writable by the builder).

```
head box                       builder container (6 cpu / 8 GB)        NAS daemon
tools/imgbuild.sh build N TAG  --ssh (nas-build)-->  context N in /data/ctx/N
                               --ssh (nas-img, forced command)-------------> docker build /mnt/.../d2-builder/ctx/N
tools/imgbuild.sh save TAG ----ssh (nas-img)------------------------------> docker save TAG | gzip -n here
```

The context is assembled on the builder (Go, curl, git there) from a copy of the head box working tree; the NAS builds only a directory that already exists under the fixed context directory (the builder's `/data/ctx`, on the NAS NVMe pool). Nothing is stored in docker.img except the image layers, which show in the Docker tab under Images.

Forced command rules (V: every refusal below was drilled, section 4):
- the key line in the NAS root `authorized_keys` is `command="<imgbuild>",restrict,from="<head box address>" ssh-ed25519 ... d2-imgbuild`: no pty, no forwarding, no agent, no X11, only from the head box;
- verbs: `build NAME TAG [K=V...]`, `save TAG`, `rmi TAG`, `list`; anything else exits 2 with one line;
- NAME is `[a-z0-9][a-z0-9._-]{0,40}` and must be an existing, non-link directory under the fixed context directory with a non-link `Dockerfile`; a `# syntax=` or `# escape=` line (BuildKit would run an arbitrary frontend image) and `--security=insecure` / `--network=host` RUN flags are refused;
- TAG must be `REPO:VERSION` with REPO one of `didle`, `d2dbot`, `imgscratch`; `save` only those; `rmi` only tags whose version starts with `scratch-` (so a real `didle:0.1.2` or `d2dbot:devpN` can never be removed through the key);
- build args only `D2D_SRC DEV APP_VERSION BUILD_JOBS UBUNTU`, values `[A-Za-z0-9._/:@+-]{0,80}`;
- the build is `nice -n 19 ionice -c3 docker build --cpu-shares 2 --memory 4g --memory-swap 4g`, one at a time (flock, waits up to an hour), refused when the NAS has under 3 GiB MemAvailable.

Residual risk (accepted): whoever holds the key (the head box) can build any Dockerfile that the builder container wrote into the context directory; a build step runs unprivileged inside the daemon's sandbox, like any `docker build`. The key cannot run containers, mount the host, save arbitrary images or touch other containers.

## 2. How to run an image build on the NAS

From the repo root on the head box (the ssh aliases `nas-build` and `nas-img` are in the head box `~/.ssh/config`; the private key `id_ed25519_towerimg` stays there, never in the repo):

```
tools/imgbuild.sh build NAME TAG [--arg K=V]... -- CONTEXT-CMD...      # context on the builder, build on the NAS; the context dir is appended to CONTEXT-CMD
tools/imgbuild.sh build didle-0-1-2 didle:0.1.2 -- deploy/d2d-container/context.sh        # the friends image
tools/imgbuild.sh save didle:0.1.2 /path/didle-image.tar.gz              # same bytes as docker save | gzip -n
tools/imgbuild.sh list ; tools/imgbuild.sh rmi imgscratch:scratch-x
tools/friendpack/build.sh ... --image didle:0.1.2 --image-from-nas ...   # the friends pack takes the image from the NAS
tools/dev-update.sh --image d2dbot:devpN --d2d <path ON THE BUILDER>     # the DEV image; the other dev-update modes still act on the host that runs the sandbox (D2_SANDBOX_HOST)
```

`tools/dev-update.sh --image` assembles the context with `tools/dev-update-host.sh --ctx-out DIR` (new flag: stop after the context and the stub patch) and builds it through the same path. The image then lives on the NAS daemon; moving it to the host that RUNS the sandbox is `tools/imgbuild.sh save TAG - | ssh <sandbox host> docker load` (not run, stage 4 decides where the sandbox lives).

## 3. Script changes (docs and scripts only)

| File | Change |
|---|---|
| `deploy/nas/imgbuild.sh` (new) | the forced command |
| `tools/imgbuild.sh` (new) | client: build, save, rmi, list |
| `tools/dev-update.sh`, `tools/dev-update-host.sh` | `--image` goes through the NAS; `--ctx-out`; `D2_SANDBOX_HOST` for the modes that need the sandbox docker |
| `tools/friendpack/build.sh` | `--image-from-nas` |
| ops repo `tools/merge.sh` | gate host default `tower` (was the worker host) |
| ops repo `tools/golive-dry.sh`, `tools/tickstatus.sh` | use `$D2_LABRUN_HOST`, default `tower` (the load read in the status line is now the builder) |
| ops repo `tools/vet-mac.sh` | no default source host: it reads the fork worktree and lab data of the user's PC, so it refuses without an explicit `VETMAC_SRC_HOST` |
| ops repo `tools/tests/test_merge_sh.sh` | expects `tower` (69 of 69 pass) |

Left unchanged on purpose: the live-game tooling and the credit-runner allow list (`labrun orion -- ...` patterns), which name the PC deliberately, and `tools/smoke.sh`, `tools/firsthour.sh`, `tools/floors/extract.sh`, `tools/dev-sandbox.sh`, the emulator wrapper: they need docker of the sandbox host or a live game, so they follow `D2_LABRUN_HOST`/stage 4 (set it for one command if the user allows).

## 4. Drills (V, 2026-11-14, from the head box)

| Drill | Result |
|---|---|
| refusals: `id`, `ls /`, `build ../x ...`, `build x ubuntu:1`, `save ubuntu:latest`, `rmi didle:0.1.2`, `build x didle:a FOO=1` (no such context), port forward, pty | each exits 2 with one line (pty and forwarding refused by `restrict`) |
| small image: Ubuntu base plus one RUN, 78 MB, `imgscratch:scratch-small` | built in 12 s; `save | gzip -n` 30 MB; `rmi` removed it; context removed |
| didle image from the real `context.sh`, tag `didle:scratch-stage3` (see section 5) | see section 5 |
| original NAS root key still works after the edit | yes |

## 5. Real didle image build

DIDLE_RESULT

## 6. Not done and proposals

- Stage 4 (the sandbox `d2play-*` moves to the NAS as an ordinary visible template, its own caps, a copy of the volume verified by hash, the Orion copy kept a week) is NOT started. Proposal: a card after the user's go; the image for it is built through this path and loaded with `save | docker load` on the NAS itself, so no Orion step remains.
- The docker.img has about 52 GB free and 85 GB of reclaimable images (V: `docker system df`); nothing was pruned (not ours). A prune of dangling build cache is the user's call.
- The Backrest/backup side is unchanged: the new key sits in the NAS boot config, the context directory is under the builder data volume, already excluded from the backup except `ssh/`. The ctx directory is not yet in the exclude list (U): it is emptied after every build.

## 7. Rollback

On the NAS (root key): `cp /boot/config/ssh/root/authorized_keys.bak-20261114-stage3 /boot/config/ssh/root/authorized_keys` removes the restricted key; `rm /mnt/cache/appdata/_unraid/imgbuild`; `rmdir /mnt/cache/appdata/d2-builder/ctx` (empty). Remove any leftover scratch image with `docker rmi didle:scratch-stage3 imgscratch:scratch-small`. On the head box: delete the `nas-img` block (backup of the file: `~/.ssh/config.bak-20261114-stage3`) and `id_ed25519_towerimg*`. Script changes: revert the branch; the old behaviour (`D2_LABRUN_HOST=<worker host>` with docker there) still works because `D2_SANDBOX_HOST` defaults to `D2_LABRUN_HOST`, except `--image`, which now always uses the NAS.
