#!/usr/bin/env python3
"""gridwhole: is `debug grid 5` whole at arrival? (card gridwhole, docs/probes/gridwhole-2026-11-22.md)

Per (level, seed): start a lab d2d, `debug goto L`, read the grid at arrival (A), bring every room up by warping the hero over a
lattice of walkable cells (`debug warp`, SETUP only, cheated mechanics probe), read the grid again (B), and compare A with B
without the cells near any monster of either snapshot (unit stamps). It also asks the engine's own collision query
`debug blocked x y` (cells, 5 subtiles each) for a random sample of A and B cells and counts disagreements.
One JSON line per (level, seed) on stdout. Usage: gridwhole.py D2D DATA LEVEL[,LEVEL..] SEED[,SEED..] [SAMPLES]
Stdlib only; talks to the control socket of a --headless --lab-verbs d2d (docs/control_channel.md of the fork)."""
import json, os, random, socket, subprocess, sys, tempfile, time

def session(d2d, data, seed):
    home = tempfile.mkdtemp(prefix="gw_")
    os.makedirs(home + "/.config/d2d/save", exist_ok=True)
    sock = home + "/d.sock"
    p = subprocess.Popen([d2d, "--headless", "--seed", str(seed), "--data", data, "--devctl", sock, "--start-name", "Gw",
                          "--start-screen", "ingame", "--start-class", "3", "--lab-verbs"],
                         env=dict(os.environ, HOME=home), stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    for _ in range(900):
        if os.path.exists(sock):
            break
        assert p.poll() is None, "d2d exited during startup"
        time.sleep(0.1)

    def cmd(line):
        s = socket.socket(socket.AF_UNIX)
        s.settimeout(40)
        s.connect(sock)
        s.sendall((line + "\n").encode())
        b = b""
        try:
            while True:
                c = s.recv(1 << 20)
                if not c:
                    break
                b += c
                t = b.rstrip().split(b"\n")[-1]
                if t == b"ok" or t.startswith(b"ok ") or t.startswith(b"err"):
                    break
        finally:
            s.close()
        return b.decode()
    for _ in range(300):
        if "screen=ingame" in cmd("state"):
            break
        time.sleep(0.2)
    time.sleep(1.0)
    return p, cmd

def grid(cmd):
    rows = [r for r in cmd("debug grid 5").split("\n") if r and set(r) <= set(".#")]
    return rows

def laid(cmd):
    t = cmd("debug laid").split("\n")
    up = int(t[0].split()[1])
    return up, sum(1 for r in t[1:] if r and r != "ok")

def monsters(cmd):
    out = []
    for line in cmd("monsters").split("\n"):
        f = line.split("\t")
        if len(f) >= 3:
            try:
                out.append((float(f[1]) * 5, float(f[2]) * 5))
            except ValueError:
                pass
    return out

def objects(cmd):
    out = []
    for line in cmd("debug objects").split("\n"):
        f = line.split("\t")
        if len(f) >= 4:
            try:
                out.append((float(f[1]) * 5, float(f[2]) * 5))
            except ValueError:
                pass
    return out

def probe(d2d, data, level, seed, samples):
    p, cmd = session(d2d, data, seed)
    res = {"level": level, "seed": seed}
    try:
        r = cmd("debug goto %d" % level)
        if not r.startswith("ok"):
            res["error"] = "goto: " + r.strip()[:80]
            return res
        time.sleep(1.5)
        lv = cmd("debug level").split()
        w, h = int(lv[4]), int(lv[5])
        a = grid(cmd)
        res.update(cells_w=w, cells_h=h, rows=len(a), cols=len(a[0]) if a else 0,
                   whole=(len(a) == 5 * h and all(len(r) == 5 * w for r in a)))
        mon_a = monsters(cmd) + objects(cmd)
        up0, total = laid(cmd)
        res.update(rooms_up_arrival=up0, rooms_total=total)
        rng = random.Random(level * 1000 + seed)
        # engine's own collision at arrival
        def sample(g, stamps):
            bad = 0
            for _ in range(samples):
                y = rng.randrange(len(g)); x = rng.randrange(len(g[0]))
                r = cmd("debug blocked %.1f %.1f" % (x / 5, y / 5)).split()
                if len(r) < 2 or r[0] != "ok":
                    continue
                if (r[1] == "1") != (g[y][x] == "#") and not any(abs(x - mx) < 15 and abs(y - my) < 15 for mx, my in stamps):
                    bad += 1   # a disagreement away from every unit and object
            return bad
        res["blocked_mismatch_A"] = sample(a, mon_a)
        # bring every room up: warp over a lattice of walkable cells
        step = 5
        pts = []
        for cy in range(0, h, step):
            for cx in range(0, w, step):
                for dy in range(step):
                    for dx in range(step):
                        x, y = 5 * (cx + dx) + 2, 5 * (cy + dy) + 2
                        if y < len(a) and x < len(a[0]) and a[y][x] == "." and a[y - 2][x - 2] == "." and a[y + 2 if y + 2 < len(a) else y][x] == ".":
                            pts.append((x / 5, y / 5))
                            break
                    else:
                        continue
                    break
        warps = 0
        for (x, y) in pts:
            up, total = laid(cmd)
            if up >= total:
                break
            cmd("debug warp %.1f %.1f" % (x, y))
            warps += 1
            time.sleep(0.35)
        time.sleep(1.5)
        up1, total = laid(cmd)
        res.update(rooms_up_final=up1, warps=warps)
        b = grid(cmd)
        mon = monsters(cmd) + objects(cmd) + mon_a
        res["same_size"] = len(b) == len(a) and all(len(x) == len(y) for x, y in zip(a, b))
        if res["same_size"]:
            raw = []
            for y in range(len(a)):
                if a[y] != b[y]:
                    raw += [(x, y) for x in range(len(a[y])) if a[y][x] != b[y][x]]
            near = lambda c: any(abs(c[0] - mx) < 15 and abs(c[1] - my) < 15 for mx, my in mon)
            real = [c for c in raw if not near(c)]
            res.update(diff_raw=len(raw), diff_after_unit_mask=len(real), diff_sample=real[:40], diff_bbox=([min(c[0] for c in real), min(c[1] for c in real), max(c[0] for c in real), max(c[1] for c in real)] if real else None),
                       diff_to_blocked=sum(1 for x, y in real if b[y][x] == "#"), diff_to_open=sum(1 for x, y in real if b[y][x] == "."),
                       walkable_A=sum(r.count(".") for r in a), walkable_B=sum(r.count(".") for r in b))
        res["blocked_mismatch_B"] = sample(b, mon)
        res["samples"] = samples
    finally:
        p.terminate()
        try:
            p.wait(10)
        except Exception:
            p.kill()
    return res

if __name__ == "__main__":
    d2d, data = sys.argv[1], sys.argv[2]
    levels = [int(x) for x in sys.argv[3].split(",")]
    seeds = [int(x) for x in sys.argv[4].split(",")]
    n = int(sys.argv[5]) if len(sys.argv) > 5 else 300
    for lv in levels:
        for sd in seeds:
            try:
                r = probe(d2d, data, lv, sd, n)
            except Exception as e:  # keep going: one level failing is a result
                r = {"level": lv, "seed": sd, "error": repr(e)[:120]}
            print(json.dumps(r), flush=True)
