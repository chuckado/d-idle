## Why

Fork binaries and PR-candidate builds ran on the worker host, which is also the game host: a 16-way build competes with the live soak. The user asked (2026-11-01) for a Linux x86_64 build host.

## What Changes

- A resource-limited build container on the NAS (details in the private ops repo), with the d2d build toolchain and a Forgejo Actions runner (labels `linux-build`, `linux-x86_64`).
- A private wrapper `build-fork.sh` that builds a fork branch there and copies the binaries back under a new name.
- docs/builder.md: what it is, its limits, how to use and stop it, what it is not for.

## Impact

Docs and a change card only in this repo; the container, template and wrapper live in the private ops repo and on the NAS. No spec, policy, ADR or glossary change.
