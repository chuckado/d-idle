"""Own minimal devctl session for the skeleton-speed probe (docs/probes/skeleton-speed-2026-11-23.md). Not fork code.
Starts a lockstep headless d2d on a private socket and sends control-socket lines. Usage: from skeleton_speed_lib import Session."""
import os, socket, subprocess, sys, tempfile, time

class Session:
    def __init__(self, d2d, extra, name="Sk", cls=1, seed=5):
        data = os.environ["D2_MPQ_DIR"]
        self.home = tempfile.mkdtemp(prefix="d2d_skel_")
        os.makedirs(os.path.join(self.home, ".config/d2d/save"), exist_ok=True)
        self.sock = os.path.join(tempfile.gettempdir(), f"d2d_skel_{os.getpid()}.sock")
        self.proc = subprocess.Popen([d2d, "--headless", "--seed", str(seed), "--data", data, "--devctl", self.sock,
                                      "--start-name", name, "--start-screen", "ingame", "--start-class", str(cls)] + extra,
                                     env=dict(os.environ, HOME=self.home), stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        for _ in range(900):
            if os.path.exists(self.sock):
                break
            assert self.proc.poll() is None, "d2d exited during startup"
            time.sleep(0.1)
        for _ in range(400):
            if "screen=ingame" in self.cmd("state"):
                break
            time.sleep(0.2)

    def cmd(self, line):
        with socket.socket(socket.AF_UNIX) as s:
            s.settimeout(30)
            s.connect(self.sock)
            s.sendall((line + "\n").encode())
            out = ""
            while not out.endswith("\n") or not out.rsplit("\n", 2)[-2].startswith(("ok", "err")):
                chunk = s.recv(65536)
                assert chunk, f"{line!r}: closed after {out!r}"
                out += chunk.decode()
        return out

    def close(self):
        try:
            self.cmd("quit")
        except Exception:
            pass
        try:
            self.proc.wait(15)
        except Exception:
            self.proc.kill()
