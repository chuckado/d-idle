#!/usr/bin/env python3
"""capture: READ-ONLY capture of one level of a running d2d through its control socket (card toolstrade13).

Usage: capture.py SOCK OUTDIR [AREA [WAIT_S [GRACE_S]]]
Waits (at most WAIT_S, default 3600) until `debug level` reports AREA (default 28, the Barracks), waits GRACE_S (default 6) for
the neighbours and rooms to settle, then writes one file per read verb into OUTDIR: the replies of `debug level`, `levelmap`,
`levelmonsters`, `levelobjects`, `presets`, `objects`, `warps`, `warptriggers`, `nearby`, `layout`, `grid 5`, `blocked`
samples of the row of the Smith anchor, and a `capture.json` index. Only verbs that READ are sent (ADR 0014: reading is
always allowed); no input, no `debug goto`, no setup verb: the allow-list below is checked before every send, and a verb
not on it raises. Stdlib only. The socket is a unix path: run it where the path exists (inside the container, or on the
host when the socket is mounted). Example, inside the sandbox container:
    python3 capture.py /path/to/d2d.sock /out/barracks-capture-$(date +%s) 28"""
import json, os, socket, sys, time

READ_VERBS = ("debug level", "debug levelmap", "debug levelmonsters", "debug levelobjects", "debug presets", "debug objects",
              "debug warps", "debug warptriggers", "debug nearby", "debug layout", "debug grid 5", "debug blocked", "state", "info")

def allowed(line):
    return any(line == v or line.startswith(v + " ") for v in READ_VERBS)

def ask(sock, line):
    if not allowed(line):
        raise ValueError("not a read verb: " + line)
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
    return b.decode(errors="replace")

def level_of(reply):
    """The level id of a `debug level` reply ("ok 28 25.00 28.00 50 56 ..."), or None."""
    f = reply.split()
    return int(f[1]) if len(f) >= 2 and f[0] == "ok" and f[1].isdigit() else None

def main():
    sock, out = sys.argv[1], sys.argv[2]
    area = int(sys.argv[3]) if len(sys.argv) > 3 else 28
    wait = float(sys.argv[4]) if len(sys.argv) > 4 else 3600
    grace = float(sys.argv[5]) if len(sys.argv) > 5 else 6
    t0 = time.time()
    while time.time() - t0 < wait:
        try:
            if level_of(ask(sock, "debug level")) == area:
                break
        except (OSError, ValueError):
            pass
        time.sleep(2)
    else:
        print("capture: area %d not seen within %ds" % (area, wait)); sys.exit(2)
    time.sleep(grace)
    os.makedirs(out, exist_ok=True)
    index = {"area": area, "at": time.strftime("%Y-%m-%dT%H:%M:%S"), "files": {}}
    for verb in READ_VERBS[:-3] + ("state", "info"):
        name = verb.replace("debug ", "").replace(" ", "_")
        try:
            body = ask(sock, verb)
        except (OSError, ValueError) as e:
            body = "capture error: %s" % e
        with open(os.path.join(out, name + ".txt"), "w") as f:
            f.write(body)
        index["files"][verb] = len(body)
    with open(os.path.join(out, "capture.json"), "w") as f:
        json.dump(index, f, indent=1)
    print("capture: wrote", out)

if __name__ == "__main__":
    main()
