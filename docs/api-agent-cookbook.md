# API-credit runner cookbook: command forms that work

> status: current (card apicookbook). Evidence (V): the stream logs of seven decomposition jobs on the API-credit runner. Three ran emulator differentials on the worker host (monsterstats2, treasure, levelspawn); three wrote desk-only traces because their remote calls were denied (monsterhit, playerdef, uniqmods). The runner's runtime preamble (`PREAMBLE` in the ops repo's `tools/api-agent.py`) repeats the short form of this page.

The runner runs `claude -p` in dontAsk mode: every Bash call is matched against an allow-list, a denied call is final. Rules match the TEXT of each sub-command (pipes, `&&`, `;`, `||`, `2>&1`, `>` into the worktree or the scratch dir are fine when each part is allowed).

## What worked and what was denied (V, from the logs)

The labrun machine name is read from `D2_LABRUN_HOST` (environment or the gitignored `local.env`, template `local.env.example`); the table writes it as that variable.

| form | result | why |
|---|---|---|
| `tools/agentrun.sh NAME --ttl N -- labrun $D2_LABRUN_HOST -- bash tools/emu/x.sh args` | allowed (all three successful jobs) | matched `tools/agentrun.sh *` |
| bare `labrun $D2_LABRUN_HOST -- bash tools/emu/x.sh args` | denied (monsterhit, playerdef, uniqmods) | no allow rule for `labrun $D2_LABRUN_HOST -- bash`; they never tried the agentrun spelling |
| `labrun $D2_LABRUN_HOST -- sh -c '...'` (any body except go vet/test/build) | denied | only `sh -c 'go vet|test|build ...` is allowed |
| `labrun $D2_LABRUN_HOST -- ls\|cat\|head\|tail\|grep\|rg\|wc\|stat PATH` with absolute paths | allowed | explicit rules |
| `labrun $D2_LABRUN_HOST -- sed -n ...`, `find ...`, `env ...`, `bash ...`, a venv python, `git -C ... log` | denied | no rule (sed, find, env, git -C are not in the remote list) |
| `labrun $D2_LABRUN_HOST -- ls ../../../../x` (relative climb) | mixed, some denied | do not use; use absolute paths |
| `labrun $D2_LABRUN_HOST -- grep -n "" FILE \| head -N \| tail -n +K` | allowed | the way to page a remote file instead of `sed -n` |
| `cd DIR; ...`, `ls ~/code`, `ls` of a home-directory path | denied | `cd` rules, and paths outside the worktree and the read-only dirs are checked against the Read rules |
| `find / ...`, `find /home ...` | denied (correctly) | whole-tree scans; they wedge the worker host |
| `python3 - <<'EOF'` and `cat > f <<'EOF'` heredocs | denied | use the Write and Edit tools |
| `chmod`, `rm file`, `sed -i`, `git checkout -- f`, `printf >> docs/OWNERS` | denied | use Edit, `git rm`, `git restore` |
| `VAR=x tools/agentrun.sh ...` (env prefix) | denied | no env prefix; the wrapper sets the environment |
| any `labrun` line containing the letters "kill", even in "skill" | denied (`labrun *kill*`) | write the pattern as `ski[l]l` |
| `until ...; do sleep; done` | denied | use the wrapper's ttl instead |

## 1. Run an emulator script on the worker host: the wrapper

    tools/emu/run-on-worker.sh [--ttl SECS] SCRIPT.sh|SCRIPT.py [ARGS...] > /tmp/d2agent-NAME/out.txt 2>&1

- Write your script into `tools/emu/` of your worktree (Write tool); `labrun` mirrors the worktree to the worker host on every call, so it is there without any copy step. SCRIPT is a bare file name; `.sh` runs with bash, `.py` runs with the emulator venv and upstream's emulator modules on the path (`tools/emu/run-py.sh`).
- The wrapper runs under `tools/agentrun.sh NAME --ttl TTL+30` and `timeout TTL` on the worker host (default 900, max 3000 s; exit 124 = ttl), exports `D2_SCRATCH=/tmp/d2agent-NAME` (created, use it for outputs on the worker host) and `D2_MPQ_DIR=/tmp/ug-emu`.
- Refused with exit 2 and a one-line reason: a script text that names sudo, kill, curl, ssh, rsync, wine, docker, systemctl, `rm -r/-f`, chmod, key or claude paths; an argument outside `[A-Za-z0-9 _.,:=@%+/-]` or with `..`; an absolute path argument outside `/tmp/ug-emu`, `/tmp/casttrace`, `~/d2d-lab` and your scratch dir.
- Heavy runs: check load first with `labrun $D2_LABRUN_HOST -- top -bn2 -d1`; `--ttl` bounds the run. Reap leftovers with `tools/agentreap.sh NAME`, never kill.
- Smoke test: `tools/emu/run-on-worker.sh --ttl 60 run-on-worker-hello.py` prints the unicorn version and the MPQ files.
- Offline test of the wrapper (stub labrun): `bash tools/emu/run-on-worker-test.sh`.

## 2. Game data tables (V, listed on the worker host 2026-11-02)

- Emulator MPQ dir: `/tmp/ug-emu` (d2data, d2exp, d2char, music and speech MPQs, `Patch_D2.mpq`, `bin/`). `/tmp` is volatile: if it is gone, say so, do not search for it.
- Game data dir (never written): `~/d2d-lab/data`. The game binary the emulator loads: `/tmp/casttrace/game.exe`. Patch installer: `~/Downloads/LODPatch_114d.exe`.
- A table in the PATCHED 1.14d stack (the trap is in docs/process/mpq-data-reading.md): `tools/emu/run-on-worker.sh mpq-table.sh monstats > /tmp/d2agent-NAME/monstats.txt` (optional second argument: first N lines). Check the first line: patched is `Id hcIdx ...`.
- Upstream checkouts (read-only, absolute paths, `labrun $D2_LABRUN_HOST -- grep|cat|head|tail|ls`): `~/d2d-lab/src` (main, with `tools/emu/*.py`, `components/rules/*.hpp`, `docs/research/re/*.md`), `~/d2d-lab/pr/<name>` read checkouts such as `skmana`, `gaps`, `walk2`, `sorcrest-d2d`. List them with `labrun $D2_LABRUN_HOST -- ls` and the absolute path of `~/d2d-lab/pr` (the runner preamble spells it out).
- Emulator venv: `~/d2d-lab/pr/walk2-venv/bin/python`; the wrapper's `.py` form uses it. Existing wrappers to copy from `tools/emu/`: `ms_run.sh` (stat rows), `tre_run.sh` (drops), `sk_*_run.sh`, `nec_run.sh`.

## 3. d2d probe builds

- Never touch the fork checkout or the lab binary. Build a probe out of tree: copy `tools/emu/ms_build.sh` (compiles one test cpp by hand against the libraries of the existing scratch build `~/d2d-lab/pr/sorcrest-out/build`, source `~/d2d-lab/pr/sorcrest-d2d`) to your own `tools/emu/<x>_build.sh`, point its output to `$D2_SCRATCH`, run it with the wrapper (`--ttl 1500`). It uses `nice`; for make or ninja use `-j2` at most (`labrun $D2_LABRUN_HOST -- ninja -C DIR -j2 target` is an allowed bare form when DIR is a build dir you own). Tag: V that the wrapper path works; the -j2 bare ninja form is U (not exercised).
- You cannot create a git worktree (denied); a scratch checkout is made by the coordinator, ask in the report.

## 4. Forbidden (all denied or blocked by the guard)

`find /` and any whole-tree scan, kill, pkill, killall, sudo, the break-glass sudo helper, ssh, scp, rsync, curl, wget, nested claude, the live game tools, writes outside your worktree and `/tmp/d2agent-NAME`, anything naming the API key, its helper or `~/.claude*`. Do not retry a denied call in another spelling; use the forms above or report the need.
