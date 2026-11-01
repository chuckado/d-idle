# Worker capacity

> status: moved to the private ops repo (d2lod-ops, `docs/process/worker-capacity.md`). Not part of the public tree.

Update 2026-11-01 (tuning card): the capacity check now has a bounded form (about 7 s, one verdict line OK / BUSY (cpu) / BUSY (io) / NO) in the private ops tools.
Sustained I/O wait (>= 40 percent over two samples) counts only when a local disk is really busy; wait from tasks stuck on network mounts is reported as phantom and ignored.
Builds should run through the nice and CPU-quota wrapper and the compiler cache described in the private ops tuning doc.
