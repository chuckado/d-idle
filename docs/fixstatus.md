# fixstatus: the agentstatus stub did not run

Cause: `tools/agentstatus/agentstatus.py` was a POSIX shell stub (`exec "$(dirname "$0")/../ops.sh" run ...`) saved under a `.py` name. Run as `python3 tools/agentstatus/agentstatus.py`, as AGENTS.md tells agents to, Python parsed the shell line and raised `SyntaxError: Missing parentheses in call to 'exec'`. The private ops copy was fine; only the public stub was wrong.

Fix (public): the stub is now a real Python file that `os.execv`s `tools/ops.sh run agentstatus/agentstatus.py` with the same arguments. It works with `python3 FILE` and as an executable. A missing ops checkout still gives the one-line message from `ops.sh` and exit 2.

Verify: `sh tools/agentstatus/stub-test.sh` (compiles the stub; with `D2_OPS_DIR` pointing nowhere it must exit 2 with one line). With the ops checkout present, `python3 tools/agentstatus/agentstatus.py set NAME --state working` updates the board.

Status: VERIFIED LIVE on the worker host's board (one `set` accepted); the stub test is offline.
