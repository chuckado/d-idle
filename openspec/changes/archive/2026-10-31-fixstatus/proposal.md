## Why

Three agents could not update the status board: the public stub `tools/agentstatus/agentstatus.py` was shell code in a `.py` file and raised a Python SyntaxError.

## What Changes

- The stub is a Python file that execs `tools/ops.sh run agentstatus/agentstatus.py`; a regression test covers compile and the missing-ops exit 2.
- No spec delta.

## Impact

Owns tools/agentstatus/** and docs/fixstatus.md. The private ops copy is unchanged and was never at fault.
