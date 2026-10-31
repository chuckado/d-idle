#!/bin/sh
# Offline test of the agentstatus.py stub. Run: sh tools/agentstatus/stub-test.sh
here=$(cd "$(dirname "$0")" && pwd)
fail() { echo "FAIL: $1"; exit 1; }
python3 -m py_compile "$here/agentstatus.py" || fail "stub does not compile"
rm -rf "$here/__pycache__"
err=$(D2_OPS_DIR=/nonexistent python3 "$here/agentstatus.py" set --help 2>&1 >/dev/null); rc=$?
[ "$rc" = 2 ] || fail "exit $rc, want 2"
[ "$(printf '%s\n' "$err" | wc -l)" = 1 ] || fail "want one line, got: $err"
case $err in "ops repo not present"*) ;; *) fail "message: $err" ;; esac
echo "ALL OK"
