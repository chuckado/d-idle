#!/usr/bin/env python3
"""Record a WALKING-HEAVY 25 Hz sequence from a lockstep d2d on the builder (lab instance, throwaway hero, no save).

record2.py OUT_DIR [D2D_BIN] [DATA_DIR]: town idle, a town walk (20 s), Blood Moor walking (72 s, `debug goto 2`, a
fixed-seed click pattern), a fight (spawned fallen, Fire Bolt via `debug cast`), idle in the field. Setup writes
(`debug goto/spawn/points`) are lab setup on a throwaway hero: fine for bytes per frame, never evidence about play.
The first part of the docstring of record.py applies otherwise.

  record.py OUT_DIR [D2D_BIN] [DATA_DIR]

Starts d2d headless (--lockstep --no-save, fresh sorceress in town), then for each scene steps one tick with `step 1 render`
and asks for a `screenshot` PNG, so the files are what a 25 fps feed would carry (the 40 ms world tick). Scenes: town idle,
walking (the view scrolls), inventory and character panels. Frames are never committed. Writes OUT_DIR/f0000.png ...
and OUT_DIR/scenes.txt (first index and name of each scene). Stdlib only; talks to d2d over its devctl unix socket.
"""
import os, socket, subprocess, sys, time

out = sys.argv[1]
d2d = sys.argv[2] if len(sys.argv) > 2 else "/data/build/pf6/apps/d2d/d2d"
data = sys.argv[3] if len(sys.argv) > 3 else "/gamedata"
os.makedirs(out, exist_ok=True)
sock_path = os.path.join(out, "ctl.sock")
if os.path.exists(sock_path):
    os.unlink(sock_path)
p = subprocess.Popen([d2d, "--data", data, "--headless", "--devctl", sock_path, "--no-video", "--lockstep", "--no-save", "--lab-verbs",
                      "--seed", "7", "--start-screen", "ingame", "--start-class", "1", "--start-name", "Vfeed"],
                     stdout=open(os.path.join(out, "d2d.log"), "w"), stderr=subprocess.STDOUT)
try:
    for _ in range(300):
        if os.path.exists(sock_path) or p.poll() is not None:
            break
        time.sleep(0.2)
    if not os.path.exists(sock_path):
        raise SystemExit("no control socket; d2d exit %s; log tail:\n%s" % (p.poll(), open(os.path.join(out, "d2d.log")).read()[-1500:]))
    s = socket.socket(socket.AF_UNIX)
    s.settimeout(180)
    s.connect(sock_path)
    f = s.makefile("rw", newline="\n")

    def do(line):
        f.write(line + "\n")
        f.flush()
        lines = []
        while True:
            r = f.readline()
            if not r:
                raise SystemExit("d2d closed the channel after: " + line)
            r = r.rstrip("\n")
            if r == "ok" or r.startswith("ok ") or r.startswith("err"):
                return r, lines
            lines.append(r)

    for _ in range(120):  # the game loads for a while; ping answers as soon as the channel is up
        r, _ = do("ping")
        r2, _ = do("clock")
        if r.startswith("ok") and r2.startswith("ok"):
            break
        time.sleep(0.5)
    print("clock:", do("clock")[0])
    for _ in range(40):  # wait until the hero is in the world (state answers ok)
        r, ls = do("state")
        if r.startswith("ok"):
            break
        time.sleep(1)
    print("state:", r, ls[:2])
    do("step 75")  # let the town settle (3 s of game time)
    n = 0
    scenes = []

    def frames(name, count, each=1):
        global n
        scenes.append((n, name))
        for _ in range(count):
            do("step %d render" % each)
            r, _ = do("screenshot %s/f%04d.png" % (os.path.abspath(out), n))
            if not r.startswith("ok"):
                raise SystemExit("screenshot: " + r)
            n += 1

    import random
    rng = random.Random(7)

    def walk(name, clicks, each, box):
        # box: x0 y0 x1 y1 of the click area (screen px); the hero runs to each click for `each` frames
        scenes.append((n, name))
        for _ in range(clicks):
            do("click %d %d" % (rng.randint(box[0], box[2]), rng.randint(box[1], box[3])))
            frames_raw(each)

    def frames_raw(count):
        global n
        for _ in range(count):
            do("step 1 render")
            r, _ = do("screenshot %s/f%04d.png" % (os.path.abspath(out), n))
            if not r.startswith("ok"):
                raise SystemExit("screenshot: " + r)
            n += 1

    frames("town idle", 100)
    walk("town walk", 20, 25, (60, 60, 740, 480))
    print("goto Blood Moor:", do("debug goto 2")[0])
    do("step 75")
    frames("field idle", 60)
    walk("field walk", 24, 75, (60, 60, 740, 480))
    do("debug points 36 1")
    print("spawn:", do("debug spawn fallen1 5")[0])
    scenes.append((n, "fight"))
    for i in range(300):
        if i % 12 == 0:
            do("debug cast 36 near")
        do("step 1 render")
        r, _ = do("screenshot %s/f%04d.png" % (os.path.abspath(out), n))
        if not r.startswith("ok"):
            raise SystemExit("screenshot: " + r)
        n += 1
    frames("field idle after", 100)
    open(os.path.join(out, "scenes.txt"), "w").write("".join("%d %s\n" % t for t in scenes))
    print("frames:", n)
finally:
    p.terminate()
    try:
        p.wait(10)
    except Exception:
        p.kill()
    try:
        os.unlink(sock_path)
    except OSError:
        pass
