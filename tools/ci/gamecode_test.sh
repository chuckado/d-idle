#!/usr/bin/env bash
# Counter-tests for gamecode_guard.sh: fake listings must fail, the allowed forms must pass.
set -u
cd "$(dirname "$0")/../.."
d=$(mktemp -d); trap 'rm -rf "$d"' EXIT
fail=0
want() { # want 0|1 name file
  tools/ci/gamecode_guard.sh "$3" >/dev/null; rc=$?
  [ "$rc" = "$1" ] || { echo "FAIL: $2: exit $rc, want $1"; fail=1; }
}
printf 'mov eax, [ecx+0x10]\npush ebx\ncall 0x00463dd0\ncmp eax, 0x19\njne 0x401000\nxor eax, eax\nadd esp, 8\nret\n' > "$d/asm10.txt"
want 1 "7 line x86 listing" "$d/asm10.txt"
printf '00401000  55 8b ec 83 ec 10 53 56\n%.0s' 1 > /dev/null
printf '0x00401000: 55 push ebp\n0x00401001: 8b ec mov ebp, esp\n0x00401003: 83 ec 10 sub esp, 0x10\n0x00401006: 53 push ebx\n0x00401007: 56 push esi\n0x00401008: 33 c0 xor eax, eax\n' > "$d/addr.txt"
want 1 "address and bytes listing" "$d/addr.txt"
printf 'undefined4 FUN_00463dd0(int param_1)\n{\n  int iVar1;\n  uVar2 = param_1 + 4;\n  return DAT_007a6a70;\n}\n' > "$d/ghidra.txt"
want 1 "pseudo-C block" "$d/ghidra.txt"
for i in 1 2 3 4 5 6 7 8 9; do printf '0040100%s  55 8b ec 83 ec 10 53 56  U...\n' "$i"; done > "$d/hex.txt"
want 1 "hex dump" "$d/hex.txt"
printf 'Skill\tcharclass\tlvlmana\tminac\n' > "$d/tsv.txt"
want 1 "table header" "$d/tsv.txt"
printf 'The check is `cmp esi, 0x19` then `mov eax,[ecx*4+0x7A27C0]` and `xor eax,eax`.\n| `push ebx` | `pop ebx` |\nFUN_00463dd0 returns DAT_007a6a70; see FUN_00401000.\nSkill\tLevel\n' > "$d/ok.txt"
want 0 "inline quotes and tags" "$d/ok.txt"
printf 'mov eax, 1\nmov ebx, 2\nmov ecx, 3\nmov edx, 4\nmov esi, 5\n' > "$d/five.txt"
want 0 "5 lines is below the limit" "$d/five.txt"
[ $fail = 0 ] && echo "gamecode_test ok"
exit $fail
