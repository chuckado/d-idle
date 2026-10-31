import os
import sys
import tempfile
import unittest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
import common as dc
import notify

URL = "https://discord.com/api/webhooks/" + "123456789/" + "fakeTok-en_1"
URL2 = "https://discord.com/api/webhooks/" + "987654321/" + "otherTok-en_2"


class NoKeyring:
    def get(self, account):
        return None


class WebhookLoaderTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.dir = os.path.join(self.tmp.name, "d2idle")
        os.mkdir(self.dir, 0o700)
        self.path = os.path.join(self.dir, "discord-webhook-progress")
        self.env = {dc.HOME_ENV: self.dir}

    def write(self, content=URL + "\n", mode=0o600):
        with open(self.path, "w") as f:
            f.write(content)
        os.chmod(self.path, mode)

    def load(self, env=None, ch="progress", store=None):
        return dc.load_webhook_url(ch, env=self.env if env is None else env, store=store or NoKeyring())

    def refuses(self, *needles, env=None, ch="progress"):
        with self.assertRaises(dc.DiscordError) as cm:
            self.load(env, ch)
        msg = str(cm.exception)
        self.assertNotIn(URL, msg)
        self.assertNotIn("fakeTok-en_1", msg)
        for n in needles:
            self.assertIn(n, msg)

    def test_env_wins(self):
        self.write(URL + "\n")
        e = dict(self.env, D2IDLE_DISCORD_WEBHOOK_PROGRESS=URL2)
        self.assertEqual(self.load(e), URL2)

    def test_file_used_when_env_unset(self):
        self.write()
        self.assertEqual(self.load(), URL)

    def test_explicit_file_env(self):
        self.write()
        e = {"D2IDLE_DISCORD_WEBHOOK_PROGRESS_FILE": self.path, dc.HOME_ENV: self.tmp.name}
        self.assertEqual(self.load(e), URL)

    def test_explicit_file_missing(self):
        e = {"D2IDLE_DISCORD_WEBHOOK_PROGRESS_FILE": self.path + ".nope"}
        self.refuses("does not exist", env=e)

    def test_mode_0644_refused_names_fix(self):
        self.write(mode=0o644)
        self.refuses("mode 644", "chmod 600")

    def test_symlink_refused(self):
        real = os.path.join(self.dir, "real")
        with open(real, "w") as f:
            f.write(URL)
        os.chmod(real, 0o600)
        os.symlink(real, self.path)
        self.refuses("symlink")

    def test_invalid_shape_refused(self):
        for bad in ["https://evil.example/api/webhooks/1/abc", "http://discord.com/api/webhooks/1/abc",
                    "https://discord.com/api/webhooks/abc/def", "https://discord.com/api/webhooks/1/abc?x=1",
                    "https://discord.com/api/webhooks/1/abc/github", "https://discord.com/api/webhooks/1/"]:
            self.write(bad)
            self.refuses("not a Discord webhook URL")

    def test_whitespace_refused(self):
        self.write(URL + " extra\n")
        self.refuses("without spaces")

    def test_missing_falls_through_to_keyring_then_none(self):
        self.assertIsNone(self.load())
        store = dc.MemoryStore({"discord-webhook-progress": URL2})
        self.assertEqual(self.load(store=store), URL2)

    def test_channel_validation(self):
        for bad in ["../x", "a/b", "UP", "", "-a", "a--b", "a b", "x" * 40, "a_b"]:
            with self.assertRaises(dc.DiscordError, msg=bad):
                self.load(ch=bad)
        self.write()
        os.rename(self.path, os.path.join(self.dir, "discord-webhook-my-feed2"))
        self.assertEqual(self.load(ch="my-feed2"), URL)

    def test_registered_for_redaction(self):
        self.write()
        self.load()
        self.assertNotIn("fakeTok-en_1", dc.redact("oops " + URL))

    def test_home_override_honoured(self):
        self.write()
        # env carries no override: the process-wide one (set by the test modules) is used, not the real home
        old = os.environ.get(dc.HOME_ENV)
        os.environ[dc.HOME_ENV] = self.dir
        try:
            self.assertEqual(dc.load_webhook_url("progress", env={}, store=NoKeyring()), URL)
        finally:
            if old is None:
                del os.environ[dc.HOME_ENV]
            else:
                os.environ[dc.HOME_ENV] = old

    def test_real_default_path_not_read_when_override_set(self):
        empty = os.path.join(self.tmp.name, "empty")
        os.mkdir(empty, 0o700)
        self.assertEqual(dc.data_dir({dc.HOME_ENV: empty}), empty)
        self.assertNotEqual(dc.data_dir({dc.HOME_ENV: empty}), os.path.expanduser(dc.DEFAULT_HOME))
        opened = []
        real_lexists = os.path.lexists
        real_home = os.path.expanduser(dc.DEFAULT_HOME)

        def spy(p):
            opened.append(p)
            return real_lexists(p)

        os.path.lexists = spy
        try:
            self.assertIsNone(dc.load_webhook_url("progress", env={dc.HOME_ENV: empty}, store=NoKeyring()))
        finally:
            os.path.lexists = real_lexists
        self.assertTrue(opened)
        self.assertFalse([p for p in opened if p.startswith(real_home)])

    def test_notify_uses_loader(self):
        self.write()
        self.assertEqual(notify.get_webhook("progress", store=NoKeyring(), env=self.env), URL)


if __name__ == "__main__":
    unittest.main()
