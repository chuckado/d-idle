#!/bin/sh
# Offline test of tools/lib/heavy-guard.sh: fake `free` and `pgrep` first in PATH.
here=$(cd "$(dirname "$0")" && pwd)
T=$(mktemp -d); trap 'rm -rf "$T"' EXIT
mkdir "$T/bin"
cat > "$T/bin/free" <<'F'
#!/bin/sh
echo "              total        used        free      shared  buff/cache   available"
echo "Mem:           8192        2000        3000         100        3192        ${FAKE_AVAIL:-5000}"
F
cat > "$T/bin/pgrep" <<'F'
#!/bin/sh
[ "$1" = -x ] && [ "$2" = "${FAKE_PROC:-none}" ]
F
chmod +x "$T/bin/free" "$T/bin/pgrep"
PATH=$T/bin:$PATH; export PATH
. "$here/heavy-guard.sh"
fail=0
check() { # name want-rc
	heavy_guard t >/dev/null 2>"$T/err"; rc=$?
	if [ "$rc" = "$2" ]; then echo "ok   $1"; else echo "FAIL $1: rc $rc, want $2: $(cat "$T/err")"; fail=1; fi
}
unset BLDR_LANE; BLDR_LD=$T/nolockdir; export BLDR_LD
check "off builder, enough memory, no builds" 0
FAKE_AVAIL=2999 check "memory 2999 refused" 3
FAKE_AVAIL=3000 check "memory 3000 ok" 0
FAKE_PROC=cc1plus check "cc1plus running refused" 3
FAKE_PROC=ninja check "ninja running refused" 3
FAKE_PROC=ctest check "ctest running refused" 3
BLDR_LANE=main check "main lane refused" 3
BLDR_LANE=long check "long lane ok" 0
BLDR_LANE=long FAKE_AVAIL=100 check "long lane, low memory refused" 3
mkdir "$T/lockdir"; BLDR_LD=$T/lockdir check "builder without lane wrapper refused" 3
HEAVY_GUARD_SKIP=1 BLDR_LANE=main check "skip override" 0
exit $fail
