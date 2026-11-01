# Linux builder (build host for fork binaries)

> status: current. Specifics (container name, template, limits, host) are in the private ops repo, `docs/builder-private.md`.

## What it is

A resource-limited container on the NAS that builds the private d2d fork on Debian 13 (x86_64), so builds stop competing with the game on the worker host. It carries the d2d build toolchain (gcc, clang, clang-tidy, cmake, ninja, ccache, ruff, SDL3, OpenAL, Qt6) and a Forgejo Actions runner for the user's repositories (labels `linux-build`, `linux-x86_64`, host executor, one job at a time).

## Limits (why the NAS is safe)

Hard limits are set on the container before any build: about 6 CPUs worth, a very low CPU share against normal containers, 8 GB RAM with no swap, a process cap, low IO weight, and builds run under `nice`. The wrapper checks the NAS every 15 s and stops the container if available memory falls under 3 GB (V: coded in the wrapper; what the first build did to the NAS is in the card result).

## How to use

Private wrapper, run from the d2lod-bot checkout: `tools/ops.sh run build-fork.sh <fork-branch> [--tests]`.

- It takes the branch tip from the fork checkout on the worker host (`git archive`, so no git hash in the version stamp), stages it into the builder, builds `d2d` (and with `--tests` the test binaries), and waits with a bounded timeout.
- Artifacts are copied to the worker host next to the other fork binaries with a new suffix (`d2d.<branch>-<hash>-linuxbuild`, tests in `lbtests-<branch>-<hash>/`); an existing file is never overwritten. Sizes and hash prefixes are printed.
- The same host also serves Forgejo workflows with `runs-on: linux-build`.

## Behaviour check (V, card linuxbuild)

Test binaries built here were run on the worker host against the same stored game-side data as the proof docs; the counts are in the card result (they have an owner: the proof docs).

## Stop or remove it

In the NAS web UI, Docker tab: stop the container (nothing else depends on it) or remove it; its data folder under appdata holds ccache, staged sources and the runner registration and can be deleted afterwards. The runner then shows offline in Forgejo and can be deleted there.

## What it is NOT for

No game, no live runs, no soak, no drills; it does not hold game data. Binaries it builds run on the worker host.

## Shared-file lines for the merger

docs/BACKLOG.md, add: `- LINUXBUILD: NAS build container + Forgejo runner (linux-build) and build-fork.sh wrapper for fork binaries (docs/builder.md)`

docs/INDEX.md, add: `- docs/builder.md: the Linux build host for fork binaries (limits, wrapper use, stop/remove, not for the game)`

CHANGELOG.md (unreleased, docs), add: `- docs: Linux builder host for fork and PR-candidate builds, off the game host (docs/builder.md).`
