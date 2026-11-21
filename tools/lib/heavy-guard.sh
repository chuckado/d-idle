#!/bin/sh
# Heavy-job guard for headless-browser harnesses (tools/uishots.sh, the phone UI scroll and probe harnesses).
# Source it and call `heavy_guard NAME`; it returns 0 when the job may start, else prints why on stderr and
# returns 3 (callers do `heavy_guard NAME || exit $?`).
#
# Why: on 2026-11-21 the builder container hit its memory limit while parallel C++ builds ran beside headless
# Chromium; the host hung until a reboot (docs/process/worker-capacity.md, "Builder memory rules").
#
# Refuses unless ALL of these hold:
#   1. on the builder (BLDR_LANE set by the lane wrapper, or its lock directory exists) the job runs in the
#      `long` lane, so it holds that lane itself and never sits in the merge-gate lane. Elsewhere this check is skipped.
#   2. `free -m` shows at least HEAVY_GUARD_MIN_MB (default 3000) MB available.
#   3. no cc1plus, ninja or ctest process is running (HEAVY_GUARD_PROCS overrides the list, space separated).
# HEAVY_GUARD_SKIP=1 skips everything (drills only; never set it in a script).

heavy_guard() {
	_hg_name=${1:-job}
	[ "${HEAVY_GUARD_SKIP:-0}" = 1 ] && return 0
	if [ -n "${BLDR_LANE:-}" ] || [ -d "${BLDR_LD:-/data/lock.d}" ]; then
		if [ "${BLDR_LANE:-}" != long ]; then
			echo "$_hg_name: refused: on the builder this runs in the long lane only (labrun builder --lane long -- ...), lane here is '${BLDR_LANE:-none}'" >&2
			return 3
		fi
	fi
	_hg_min=${HEAVY_GUARD_MIN_MB:-3000}
	_hg_avail=$(free -m 2>/dev/null | awk '/^Mem:/ {print $7}')
	case $_hg_avail in
	'' | *[!0-9]*) echo "$_hg_name: refused: cannot read available memory from free -m" >&2; return 3 ;;
	esac
	if [ "$_hg_avail" -lt "$_hg_min" ]; then
		echo "$_hg_name: refused: only $_hg_avail MB available, need $_hg_min MB (a headless browser beside a build ran the builder out of memory)" >&2
		return 3
	fi
	for _hg_p in ${HEAVY_GUARD_PROCS:-cc1plus ninja ctest}; do
		if pgrep -x "$_hg_p" >/dev/null 2>&1; then
			echo "$_hg_name: refused: a $_hg_p process is running; wait for the build or test to end" >&2
			return 3
		fi
	done
	return 0
}
