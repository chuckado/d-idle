## Why

The chill-length candidate for #155 (`fix/static-field-chill`, card chillpatch) needs the independent verifier of docs/process/upstream-pr-standard.md section 5 before it may join #155.

## What Changes

- Re-derive game.exe's cold chill rule from the disassembly without reading the author's docs first; own generator and harness (`tools/emu/verify_chill.py`, `tools/emu/verify_chill_mode.inc`) against the real d2d builds before (#155's head) and after (the candidate); classify every residual difference.
- Sweep d2d for other cold-chill paths; CI-class checks (pr_guard per commit, the four upstream test binaries, presubmit --quick with ruff, gcc shadow sweep, clang-tidy names / includes, PR size, fast-forward and merge checks).
- Verdict appended to `docs/research/pr-static-chill-proof.md`, the row in `docs/research/pr-bundle-queue.md`, one `docs/BACKLOG.md` line. Nothing pushed, posted or changed on the author's branch.
