import json, os, socket, subprocess, tempfile, threading, unittest

PL = os.path.join(os.path.dirname(os.path.abspath(__file__)), "capture.pl")

class Fake:
    """A d2d control socket that answers a level and logs every verb it receives."""
    def __init__(self, path, level=28):
        self.path, self.level, self.seen = path, level, []
        self.srv = socket.socket(socket.AF_UNIX); self.srv.bind(path); self.srv.listen(8)
        threading.Thread(target=self.loop, daemon=True).start()
    def loop(self):
        while True:
            try:
                c, _ = self.srv.accept()
            except OSError:
                return
            line = c.makefile().readline().strip()
            self.seen.append(line)
            c.sendall(("ok %d 25.00 28.00 50 56\n" % self.level if line == "debug level" else "row 1\nok\n").encode())
            c.close()

@unittest.skipUnless(subprocess.call(["which", "perl"], stdout=subprocess.DEVNULL) == 0, "no perl")
class CapturePl(unittest.TestCase):
    def run_pl(self, level=28, extra=None, wait="3"):
        d = tempfile.mkdtemp(); path = d + "/s"
        f = Fake(path, level)
        env = dict(os.environ)
        if extra: env["CAPTURE_EXTRA"] = extra
        p = subprocess.run(["perl", PL, path, "28", wait, "0"], capture_output=True, text=True, env=env, timeout=60)
        f.srv.close()
        return p, f
    def test_reads_every_verb_once_and_writes_json_lines(self):
        p, f = self.run_pl()
        self.assertEqual(p.returncode, 0, p.stderr)
        rows = [json.loads(l) for l in p.stdout.strip().split("\n")]
        self.assertTrue(all(not r["refused"] for r in rows))
        self.assertIn("debug levelmap", [r["verb"] for r in rows])
    def test_a_write_verb_is_refused_and_never_sent(self):
        p, f = self.run_pl(extra="debug goto 28")
        rows = [json.loads(l) for l in p.stdout.strip().split("\n")]
        bad = [r for r in rows if r["verb"] == "debug goto 28"]
        self.assertEqual(len(bad), 1); self.assertTrue(bad[0]["refused"])
        self.assertNotIn("debug goto 28", f.seen)
    def test_other_area_times_out_with_exit_2(self):
        p, f = self.run_pl(level=5, wait="3")
        self.assertEqual(p.returncode, 2)

if __name__ == "__main__":
    unittest.main()
