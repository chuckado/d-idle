#!/usr/bin/env bash
# Keeps decompiled or disassembled game code and game data tables out of the repo (docs/process/gamecode-guard.md).
# Usage: tools/ci/gamecode_guard.sh [FILE...]    no args: every tracked and untracked-not-ignored text file.
# Exit 0 clean, 1 findings. Short inline backtick quotes and FUN_xxxxxxxx address tags are fine: only blocks and tables fail.
set -u
cd "$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"

AWKPROG='
function hit(l, m) { print prev ":" l ": " m; found = 1 }
function flush() {
  if (asm >= 6) hit(asm_at, "x86 listing run of " asm " lines")
  if (pc >= 4) hit(pc_at, "Ghidra pseudo-C block of " pc " marker lines")
  if (hx >= 8) hit(hx_at, "hex dump run of " hx " lines")
  asm = pc = hx = gap = 0
}
BEGIN {
  n = split("skill charclass lvlmana minac maxac monstats level treasure class", w, " ")
  for (i = 1; i <= n; i++) names[w[i]] = 1
  names["treasure class"] = 1
}
FNR == 1 { if (prev != "") flush(); prev = FILENAME }
{
  low = tolower($0)
  # an instruction at the start of the line (after address, bytes, bullet or quote) with a register or 0x operand
  t = low
  sub(/^[ \t>*+`-]*/, "", t)
  sub(/^(0x[0-9a-f]+|[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]+):?[ \t]+/, "", t)
  while (t ~ /^[0-9a-f][0-9a-f][ \t]+/) sub(/^[0-9a-f][0-9a-f][ \t]+/, "", t)
  if (length($0) < 120 && t ~ /^(mov|movzx|movsx|push|pop|call|jmp|cmp|lea|test|xor|add|sub|j[a-z][a-z]?)[ \t]+(.*[^a-z0-9_])?(e?[abcd]x|e?[sd]i|e?[bs]p|[abcd][lh]|0x[0-9a-f]+)([^a-z0-9_]|$)/) {
    if (asm == 0) asm_at = FNR
    asm++
  } else { if (asm >= 6) hit(asm_at, "x86 listing run of " asm " lines"); asm = 0 }
  if (low ~ /undefined[1-8]|[ui]var[0-9]+|param_[0-9]+|local_[0-9a-f]+/ || (length($0) < 80 && index($0, "|") == 0 && low ~ /(lab|dat)_[0-9a-f]+/)) {
    if (pc == 0) pc_at = FNR
    pc++; gap = 0
  } else if (pc > 0 && ++gap > 2) { if (pc >= 4) hit(pc_at, "Ghidra pseudo-C block of " pc " marker lines"); pc = 0 }
  if (low ~ /^[ \t]*[0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][ \t][ \t][0-9a-f][0-9a-f] [0-9a-f][0-9a-f] [0-9a-f][0-9a-f] [0-9a-f][0-9a-f] /) {
    if (hx == 0) hx_at = FNR
    hx++
  } else { if (hx >= 8) hit(hx_at, "hex dump run of " hx " lines"); hx = 0 }
  if (index($0, "\t") > 0) {
    c = split(low, f, "\t"); delete seen; s = 0
    for (i = 1; i <= c; i++) if ((f[i] in names) && !(f[i] in seen)) { seen[f[i]] = 1; s++ }
    if (s >= 4) hit(FNR, "game data table header (" s " known columns)")
  }
}
END { if (prev != "") flush(); exit found ? 1 : 0 }
'

# Fixtures, binaries and the guard itself are not scanned.
text_files() {
  git ls-files -co --exclude-standard -z | while IFS= read -r -d '' f; do
    [ -f "$f" ] || continue
    case "$f" in
      */testdata/*|testdata/*|tools/ci/gamecode_*|*.png|*.gz|*.snap|*.d2mi|*.ico|*.jpg) continue ;;
    esac
    grep -Iq . "$f" 2>/dev/null && printf '%s\0' "$f"
  done
}

big_files() {
  local f sz
  while IFS= read -r -d '' f; do
    case "$f" in docs/*) ;; *) continue ;; esac
    [ -f "$f" ] || continue
    sz=$(stat -c %s "$f")
    [ "$sz" -le 262144 ] && continue
    awk -v f="$f" '$1 == f { ok = 1 } END { exit ok ? 0 : 1 }' tools/ci/gamecode_allow.txt \
      || echo "$f:1: $sz bytes, over 256 KB and not in tools/ci/gamecode_allow.txt"
  done < <(git ls-files -co --exclude-standard -z)
}

rc=0
if [ $# -gt 0 ]; then
  out=$(awk "$AWKPROG" "$@") || rc=1
else
  out=$(text_files | xargs -0 -r awk "$AWKPROG") || rc=1
  big=$(big_files)
  [ -n "$big" ] && { out="$out"$'\n'"$big"; rc=1; }
fi
if [ $rc != 0 ]; then
  echo "$out" | sed '/^$/d; s/^/gamecode: /'
  echo "FAIL: game code or data in the repo (docs/process/gamecode-guard.md)"
else
  echo "gamecode guard ok"
fi
exit $rc
