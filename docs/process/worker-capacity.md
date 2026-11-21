# Worker capacity

> status: moved to the private ops repo (d2lod-ops, `docs/process/worker-capacity.md`). Not part of the public tree.

Update 2026-11-01 (tuning card): the capacity check now has a bounded form (about 7 s, one verdict line OK / BUSY (cpu) / BUSY (io) / NO) in the private ops tools.
Sustained I/O wait (>= 40 percent over two samples) counts only when a local disk is really busy; wait from tasks stuck on network mounts is reported as phantom and ignored.
Builds should run through the nice and CPU-quota wrapper and the compiler cache described in the private ops tuning doc.

## Builder memory rules

Source: the builder hang of 2026-11-21 (about 45 minutes): the build container (8 GiB memory cap, 6 CPU) hit its limit, the kernel killed 15 of 16 headless Chromium processes while many `cc1plus` compiles ran beside them, the cgroup thrashed (load about 234) and ssh, the NAS API and the forge stopped answering until a reboot.
- Build parallelism is capped at 4: every job under the lane wrapper gets `CMAKE_BUILD_PARALLEL_LEVEL=4`, `MAKEFLAGS=-j4` and a `ninja` shim at the head of PATH that adds `-j 4` unless the caller passed its own `-j` (override: `BLDR_JOBS`). `go test -p` and `GOMAXPROCS` are unchanged.
- Headless Chromium never runs beside a parallel build. `tools/lib/heavy-guard.sh` (`heavy_guard NAME`, exit 3 with a message) refuses to start unless, on the builder, the job runs in the `long` lane, `free -m` shows at least 3000 MB available, and no `cc1plus`, `ninja` or `ctest` process runs. `tools/uishots.sh` uses it; every new browser harness must source it first. Offline test: `sh tools/lib/heavy-guard-test.sh`.
- The `long` lane already pauses its job below 700 MB available and starts only at 2500 MB or more. The `main` lane (merge gates, 45 minutes) has NO memory guard, only the load guard for merges; it is protected by the cap above and by the browser rule (browser harnesses refuse in `main`).
- The container memory limit is to be raised to about 12 GiB: pending with the owner (the NAS container template), not done by agents.
