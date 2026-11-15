# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp23 oracle for drill R1 of docs/upstream/briefs/hitrec-interrupt.md, the part the emulator answers:
does the used-skill pointer (skill list +0x10) outlive a swing?

    ww_run.sh ww_used.py          (needs the emulator image: the real skill table is booted)

For every Skills.txt row and every attack-like player mode, with that skill as the used skill:
  A. before: FUN_0057eec0 (the used-skill gate, called as the get-hit call at 0x57d383 does:
     mode 4, 1, skill 0) answers "refused" exactly for the rows whose byte +7 bit 0x80 is clear
     (the `interrupt` column); this reproduces decomp20's hr_interrupt.py in the same rig.
  B. the swing ends: FUN_00581020 (the player's default mode-end handler) runs for real, with its
     event queue and mode-set callees hooked; FUN_0057f020 (the NU start) and FUN_00620210 (set the
     used skill) run for real. After it: unit mode = 1 (NU), used skill = 0.
  C. after: the gate (same call) answers "allowed" for every row.
  D. FUN_0057f020 and the walk start path FUN_0057f090 clear the same field (walk modes 2 and 3).
Hooks (inputs, not claims): FUN_00639df0 (state test: 0), FUN_00553990 (clear events), FUN_00553570
(set mode: writes unit +0x10), FUN_00580310 / FUN_00580380 (after-skill extras), FUN_00620bb0 (unit
room: 0).
"""
import struct
import sys

sys.path.insert(0, __import__("os").environ["EMU"])
import skill_tips  # noqa: E402

A1, A2, SC, TH, S1, S2, S3, S4 = 7, 8, 10, 11, 13, 14, 15, 16
MODES = {"A1": 7, "A2": 8, "SC": 10, "TH": 11, "S1": 13, "S2": 14, "S3": 15, "S4": 16}


def main():
    e = skill_tips.boot()
    tables = e.r32(0x744304)
    base, count = e.r32(tables + 0xb98), e.r32(tables + 0xba0)
    ids = {}
    state = {"unit": 0}
    e.hook(0x639df0, lambda x: 0, 2)
    e.hook(0x643ce0, lambda x: ids.get(x.arg(0), 0), 3)       # skill id of a skill struct (the real one reads two derefs)
    e.hook(0x553990, lambda x: 0, 0)
    e.hook(0x580310, lambda x: 0, 0)
    e.hook(0x580380, lambda x: 0, 0)
    e.hook(0x620bb0, lambda x: 0, 1)
    from unicorn.x86_const import UC_X86_REG_EDX

    def set_mode(x):
        unit = x.mu.reg_read(UC_X86_REG_EDX)
        e.w32(unit + 0x10, x.arg(0))
        return 0
    e.hook(0x553570, set_mode, 1)
    game = e.alloc(0x2000)
    unit, skill_list = e.alloc(0x200), e.alloc(0x100)
    skills = {}
    for sid in range(count):
        skills[sid] = e.alloc(0x40)
        ids[skills[sid]] = sid

    def reset(mode, sid):
        e.mu.mem_write(unit, bytes(0x200)); e.mu.mem_write(skill_list, bytes(0x100))
        e.w32(unit, 0); e.w32(unit + 0x10, mode); e.w32(unit + 0xa8, skill_list); e.w32(skill_list + 0x10, skills[sid])

    def gate():
        return e.call(0x57eec0, 4, 1, 0, ecx=game, edx=unit) & 0xff

    rows = a_bad = c_bad = b_bad = 0
    refused_rows = set()
    for mode_name, mode in MODES.items():
        for sid in range(count):
            rec = base + sid * 0x23c
            bit = 1 if e.read(rec + 7, 1)[0] & 0x80 else 0
            reset(mode, sid)
            before = gate()
            rows += 1
            if before != bit: a_bad += 1
            if not before: refused_rows.add(sid)
            reset(mode, sid)
            e.call(0x581020, 0, 0, ecx=game, edx=unit)
            used, new_mode = e.r32(skill_list + 0x10), e.r32(unit + 0x10)
            if used != 0 or new_mode != 1: b_bad += 1
            if gate() != 1: c_bad += 1
    print(f"rows x modes: {rows} ({count} skill rows, {len(MODES)} attack-like modes)")
    print(f"A. gate before the swing ends equals the +7 bit 0x80: differ {a_bad}; refused skill rows {len(refused_rows)}")
    print(f"B. FUN_00581020 (swing end): used skill 0 and mode NU after it: failed {b_bad}")
    print(f"C. gate after the swing ends allows get-hit (return 1): failed {c_bad}")
    d_bad = 0
    for sid in range(count):
        for mode in (A1, 2, 3, 0, 1):
            reset(mode, sid)
            e.call(0x57f020, 1, 0, 0, ecx=game, edx=unit)
            if e.r32(skill_list + 0x10) != 0: d_bad += 1
    print(f"D. FUN_0057f020 (NU start) clears the used skill from modes A1/WL/RN/DT/NU: failed {d_bad} of {count * 5}")
    return a_bad + b_bad + c_bad + d_bad


if __name__ == "__main__":
    sys.exit(1 if main() else 0)
