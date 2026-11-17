#!/usr/bin/env python3
"""Record the same town frame from a lockstep d2d at several gamma values (brightness probe, docs/probes/viewer-brightness-2026-11-17.md).

  gamma_frames.py OUT_DIR [D2D_BIN] [DATA_DIR] [GAMMA...]

Same seed, same hero, same 75 ticks, so the frames differ only by the d2d.cfg `gamma` (default 155 when the key is absent;
D2D_USER_DIR points at a scratch dir with a one-line d2d.cfg). Writes OUT_DIR/g<gamma>_<scene>.png. Stdlib only; builder only.
"""
import os, socket, subprocess, sys, tempfile, time

out = sys.argv[1]
d2d = sys.argv[2] if len(sys.argv) > 2 else "/data/build/d2d.lab3-labverbs-f75816c6-d2d/apps/d2d/d2d"
data = sys.argv[3] if len(sys.argv) > 3 else "/gamedata"
gammas = [int(g) for g in sys.argv[4:]] or [155, 100]
os.makedirs(out, exist_ok=True)

def run(gamma):
    ud = tempfile.mkdtemp(prefix="gf")
    open(os.path.join(ud, "d2d.cfg"), "w").write("gamma = %d\n" % gamma)
    sock_path = os.path.join(ud, "ctl.sock")
    env = dict(os.environ, D2D_USER_DIR=ud)
    p = subprocess.Popen([d2d, "--data", data, "--headless", "--devctl", sock_path, "--no-video", "--lockstep", "--no-save",
                          "--seed", "7", "--start-screen", "ingame", "--start-class", "1", "--start-name", "Vgam"],
                         stdout=open(os.path.join(ud, "d2d.log"), "w"), stderr=subprocess.STDOUT, env=env)
    try:
        for _ in range(300):
            if os.path.exists(sock_path) or p.poll() is not None:
                break
            time.sleep(0.2)
        s = socket.socket(socket.AF_UNIX); s.settimeout(180); s.connect(sock_path)
        f = s.makefile("rw", newline="\n")
        def do(line):
            f.write(line + "\n"); f.flush()
            while True:
                r = f.readline()
                if not r:
                    raise SystemExit("closed after " + line)
                r = r.rstrip("\n")
                if r == "ok" or r.startswith("ok ") or r.startswith("err"):
                    return r
        for _ in range(120):
            if do("ping").startswith("ok") and do("clock").startswith("ok"):
                break
            time.sleep(0.5)
        for _ in range(40):
            if do("state").startswith("ok"):
                break
            time.sleep(1)
        do("step 75")
        for name, cmd in [("town", None), ("walk", "click 420 120")]:
            if cmd:
                do(cmd); do("step 60")
            do("step 1 render")
            r = do("screenshot %s/g%d_%s.png" % (os.path.abspath(out), gamma, name))
            print(gamma, name, r)
    finally:
        p.terminate()
        try: p.wait(10)
        except Exception: p.kill()

for g in gammas:
    run(g)
