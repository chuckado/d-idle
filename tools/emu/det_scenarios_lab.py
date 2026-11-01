# SPDX-License-Identifier: GPL-3.0-or-later
"""Scripted determinism scenarios on a headless d2d with the lab patch set (card dethash2).

    D2D_BIN=bin/d2d.NAME D2D_DATA=GAMEDIR D2D_SOCK=... D2D_SLOT_DIR=... \
        python3 det_scenarios_lab.py SCENARIO TAG [OUTDIR]

One fresh d2d process per run (seed 3, lockstep, no video): a scripted list of `step` blocks, a snapshot
after each block. Needs the lab verbs `debug spawn`, `debug merc`, `debug pets`, `debug players` and the
read-only `debug monstate` (hp, chill and stun deadlines, mode, position of every live monster).

A snapshot is five read-only replies: `monsters`, `debug monstate`, `debug players` (the player's cell),
`dump merc`, `debug pets`. Two hashes per run, md5 of the lines, first 12 hex:

    mon    the `monsters` replies only (the hash of the first det run, comparable with it)
    state  every snapshot line: monsters, chill/stun, the player's position, the merc, the pets

Scenarios (V = run, see docs/research/det-hashes-2026-11-01b.md):
    passive fight cold static cold2      the five of the first run (clvl 8 / 30, Ice Bolt 39, Static Field 42, Frozen Orb 64)
    hit_plain hit_champion hit_unique hit_boss
                                         Ice Bolt 39, Ice Blast 45, Frost Nova 44, Glacial Spike 55, one after the
                                         other on a freshly spawned target of the class, cast point blank; the log
                                         has the hp, chill and stun of the target per block
    merc                                 hired merc (type 0) and three zombies
    pet                                  druid with a Spirit Wolf and a Raven and two zombies
    follow                               druid with a wolf and a merc, the player walks about, zombies chase
"""

import hashlib
import os
import re
import socket
import subprocess
import sys
import time

SOCK = os.environ.get("D2D_SOCK", "d2d.sock")
OLD = {"passive", "fight", "cold", "static", "cold2"}
OLD_SKILL = {"cold": 39, "static": 42, "cold2": 64}  # Ice Bolt, Static Field, Frozen Orb
HIT_SKILLS = (39, 45, 44, 55)  # Ice Bolt, Ice Blast, Frost Nova, Glacial Spike
HIT_CLASS = {
    "plain": "zombie1 1",
    "champion": "fallen1 1 near champion",
    "unique": "fallen1 1 near unique",
    "boss": "bloodraven 1",
}
CLASS_OF = {"passive": "1", "fight": "1", "merc": "4", "pet": "5", "follow": "5"}


class Ctl:
    """The devctl socket: one command per line, a reply ends at ok / ok ... / err ... / a blank line."""

    def __init__(self, path=SOCK):
        self.sock = socket.socket(socket.AF_UNIX)
        self.sock.settimeout(900)
        self.sock.connect(path)
        self.file = self.sock.makefile("rwb", buffering=0)

    def cmd(self, text):
        self.file.write((text + "\n").encode())
        out = []
        while True:
            line = self.file.readline().decode().rstrip("\n")
            out.append(line)
            if line == "" or line == "ok" or line.startswith("ok ") or line.startswith("err"):
                return out


def start(cls, logfile):
    """Launch d2d (headless, lockstep, seed 3) and wait for the devctl socket."""
    try:
        os.unlink(SOCK)
    except OSError:
        pass
    slot = os.environ.get("D2D_SLOT_DIR", ".")
    env = dict(os.environ, XDG_CONFIG_HOME=slot + "/config", XDG_DATA_HOME=slot + "/data")
    os.makedirs(slot, exist_ok=True)
    argv = ["nice", "-n10", os.environ["D2D_BIN"], "--data", os.environ["D2D_DATA"], "--headless", "--devctl", SOCK,
            "--no-video", "--toggle", "autoloot=off,autoparty=off", "--lockstep", "--seed", "3",
            "--start-screen", "ingame", "--start-class", cls, "--start-name", "Det"]
    proc = subprocess.Popen(argv, stdout=open(logfile, "w"), stderr=subprocess.STDOUT, start_new_session=True, env=env)
    for _ in range(480):
        try:
            ctl = Ctl()
            if ctl.cmd("ping")[-1] == "ok":
                return ctl, proc
        except OSError:
            time.sleep(0.5)
    os.killpg(proc.pid, 15)
    raise SystemExit("d2d did not start")


class Run:
    def __init__(self, ctl):
        self.ctl = ctl
        self.log = []

    def do(self, text):
        reply = self.ctl.cmd(text)
        self.log.append("CMD " + text + " -> " + "|".join(reply))
        return reply

    def snap(self, label):
        """One snapshot: five read-only replies, each line tagged for the compare tool."""
        self.log.append("SNAP " + label)
        for tag, verb in (("M", "monsters"), ("S", "debug monstate"), ("P", "debug players"), ("G", "dump merc"), ("T", "debug pets")):
            reply = self.ctl.cmd(verb)
            if tag == "M":
                self.log.append("MON " + "|".join(reply))  # the first run's hash input, unchanged
            for line in reply:
                if line not in ("ok", ""):
                    self.log.append(tag + "|" + line)

    def rows(self):
        """The live monsters as (id, x, y) from `monsters`, newest id last."""
        out = []
        for line in self.ctl.cmd("monsters"):
            field = line.split("\t")
            if len(field) > 12 and field[6] not in ("DT", "DD"):
                out.append((int(field[12][1:]), float(field[1]), float(field[2])))
        return sorted(out)


def setup(run, level, power=True):
    if power:
        for stat, val in ((7, 900), (6, 900), (9, 300), (8, 300)):  # life, max life, mana, max mana
            run.do("debug stat %d %d" % (stat, val * 256))
    run.do("debug stat 12 %d" % level)
    run.do("debug goto 2")  # Blood Moor
    time.sleep(3)
    run.do("step 10")


def old_scenario(run, mode):
    steps = 500 if mode == "passive" else 250
    setup(run, 8 if mode == "passive" else 30, power=mode in OLD_SKILL)
    if mode != "passive":
        run.do("debug warp 36.0 36.0")
    if mode in OLD_SKILL:
        skill = OLD_SKILL[mode]
        run.do("debug points %d 20" % skill)
        run.do("step 3")
        run.do("debug skill left %d" % skill)
    run.do("click 400 500")
    run.do("step 5")
    for block in range(8):
        if mode in OLD_SKILL:
            run.do("debug stat 8 %d" % (300 * 256))
            run.do("debug stat 6 %d" % (900 * 256))
            rows = run.ctl.cmd("monsters")
            first = [line.split("\t") for line in rows if "\t" in line]
            if first:
                unit = [tok for tok in first[0] if tok.startswith("#")]
                uid = int(unit[0][1:]) if unit else -1
                for _ in range(2):
                    run.do("cmd skill %d %s %s %d 1" % (OLD_SKILL[mode], first[0][1], first[0][2], uid))
                    run.do("step 12")
        run.do("step %d" % steps)
        run.snap("block %d" % block)


def hit_scenario(run, cls):
    setup(run, 30)
    run.do("debug kill")
    run.do("step 5")
    for skill in HIT_SKILLS:
        run.do("debug points %d 1" % skill)
    run.do("step 3")
    for skill in HIT_SKILLS:
        run.do("debug kill")
        run.do("step 3")
        run.do("debug stat 8 %d" % (300 * 256))
        run.do("debug warp 44.2 24.2")
        run.do("step 2")
        run.do("debug spawn " + HIT_CLASS[cls])
        run.do("step 1")
        rows = run.rows()
        if not rows:
            run.log.append("NOTARGET skill %d" % skill)
            continue
        uid, tx, ty = rows[-1]
        run.do("debug warp %.2f %.2f" % (tx, ty + 1.6))  # point blank: the target 1.6 cells north of the player
        run.do("debug skill left %d" % skill)
        run.do("step 2")
        run.snap("%s skill %d before" % (cls, skill))
        run.do("cmd skill %d %.2f %.2f %d 1" % (skill, tx, ty, uid))
        for block in range(8):
            run.do("step 6")
            run.snap("%s skill %d block %d" % (cls, skill, block))


def party_scenario(run, mode):
    setup(run, 30)
    run.do("debug kill")
    run.do("step 5")
    if mode in ("pet", "follow"):
        for skill in (227, 221):  # Summon Spirit Wolf, Raven
            run.do("debug points %d 5" % skill)
        run.do("step 3")
        run.do("cmd skill 227 44.2 22.2 0 1")
        run.do("step 30")
        if mode == "pet":
            run.do("cmd skill 221 44.2 22.2 0 1")
            run.do("step 30")
    if mode in ("merc", "follow"):
        run.do("debug merc hire 0")
        run.do("step 20")
    run.do("debug spawn zombie1 %d" % (3 if mode == "merc" else 2))
    run.do("click 400 300")
    run.snap("start")
    goals = [(38, 30), (52, 22), (40, 18), (46, 30)]
    for block in range(8):
        if mode == "follow":
            run.do("cmd move %d %d" % goals[block % len(goals)])
        run.do("step 100")
        run.snap("block %d" % block)


def hashes(log):
    def norm(line):  # wall-clock driven counters are not part of the state
        return re.sub(r"(frame|music|music_old|sfx|quake)=\S*", "", line)

    mon = [line for line in log if line.startswith("MON ")]
    full = [norm(line) for line in log if line[:2] in ("M|", "S|", "P|", "G|", "T|", "SN", "NO")]
    return (hashlib.md5("\n".join(mon).encode()).hexdigest()[:12], hashlib.md5("\n".join(full).encode()).hexdigest()[:12], len(mon))


def main(argv):
    mode, tag = argv[1], argv[2]
    outdir = argv[3] if len(argv) > 3 else "out"
    os.makedirs(outdir, exist_ok=True)
    if mode in OLD_SKILL or mode.startswith("hit_"):
        cls = "4"
    else:
        cls = CLASS_OF[mode]
    ctl, proc = start(cls, os.path.join(outdir, "d2d_%s_%s.log" % (mode, tag)))
    run = Run(ctl)
    try:
        time.sleep(2)
        if mode in OLD:
            old_scenario(run, mode)
        elif mode.startswith("hit_"):
            hit_scenario(run, mode[4:])
        else:
            party_scenario(run, mode)
    finally:
        time.sleep(1)
        for sig in (15, 9):
            try:
                os.killpg(proc.pid, sig)
                proc.wait(30)
                break
            except (OSError, subprocess.TimeoutExpired):
                pass
    mon, state, count = hashes(run.log)
    with open(os.path.join(outdir, "det8_%s_%s.txt" % (mode, tag)), "w") as out:
        out.write("\n".join(run.log) + "\n")
    print("%s %s mon=%s state=%s monlines=%d" % (mode, tag, mon, state, count))


if __name__ == "__main__":
    main(sys.argv)
