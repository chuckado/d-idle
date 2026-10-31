import os
import stat
import sys
import tempfile
import unittest

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))
import common as dc

FAKE = "fake-token-for-tests"
OTHER = "fake-env-token-for-tests"


class NoKeyring:
    def get(self, account):
        return None


class TokenLoaderTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.dir = os.path.join(self.tmp.name, "d2idle")
        os.mkdir(self.dir, 0o700)
        self.path = os.path.join(self.dir, "discord-bot-token")

    def write(self, content=FAKE + "\n", mode=0o600):
        with open(self.path, "w") as f:
            f.write(content)
        os.chmod(self.path, mode)

    def env(self, **kw):
        e = {dc.BOT_TOKEN_FILE_ENV: self.path}
        e.update(kw)
        return e

    def load(self, env):
        return dc.load_bot_token(env=env, store=NoKeyring())

    def refuses(self, env, *needles):
        with self.assertRaises(dc.DiscordError) as cm:
            self.load(env)
        msg = str(cm.exception)
        self.assertNotIn(FAKE, msg)
        for n in needles:
            self.assertIn(n, msg)

    def test_env_wins_over_file(self):
        self.write()
        self.assertEqual(self.load(self.env(**{dc.BOT_TOKEN_ENV: OTHER})), OTHER)

    def test_file_used_when_env_unset(self):
        self.write()
        self.assertEqual(self.load(self.env()), FAKE)

    def test_default_path_missing_falls_through_to_keyring(self):
        home = os.environ.get("HOME")
        os.environ["HOME"] = self.tmp.name  # default path now points at a missing file
        try:
            class KR:
                def get(self, account):
                    return "fake-keyring-token" if account == "discord-bot-token" else None
            self.assertEqual(dc.load_bot_token(env={}, store=KR()), "fake-keyring-token")
            self.assertIsNone(dc.load_bot_token(env={}, store=NoKeyring()))
        finally:
            os.environ["HOME"] = home

    def test_explicit_missing_file_is_an_error(self):
        self.refuses(self.env(), "does not exist")

    def test_mode_0644_refused_without_token_in_message(self):
        self.write(mode=0o644)
        self.refuses(self.env(), "mode 644", "chmod 600 " + self.path)

    def test_mode_0640_and_0400_refused(self):
        for m in (0o640, 0o400):
            self.write(mode=m)
            self.refuses(self.env(), "exactly 600")

    def test_symlink_refused(self):
        real = os.path.join(self.dir, "real")
        with open(real, "w") as f:
            f.write(FAKE)
        os.chmod(real, 0o600)
        os.symlink(real, self.path)
        self.refuses(self.env(), "symlink")

    def test_group_writable_parent_refused(self):
        self.write()
        os.chmod(self.dir, 0o770)
        self.refuses(self.env(), "writable")

    def test_not_regular_file_refused(self):
        os.mkfifo(self.path, 0o600)
        self.refuses(self.env())

    def test_wrong_owner_refused(self):
        # Needs a second uid (root can chown); otherwise simulate by faking geteuid.
        self.write()
        real = os.geteuid
        os.geteuid = lambda: real() + 1
        try:
            self.refuses(self.env(), "not owned")
        finally:
            os.geteuid = real

    def test_whitespace_and_newlines_stripped(self):
        self.write("  %s \r\n\n" % FAKE)
        self.assertEqual(self.load(self.env()), FAKE)

    def test_empty_refused(self):
        self.write("  \n")
        self.refuses(self.env(), "empty")

    def test_embedded_space_refused(self):
        self.write("fake token for tests\n")
        self.refuses(self.env(), "without spaces")

    def test_loaded_value_is_redacted(self):
        self.write()
        self.load(self.env())
        self.assertNotIn(FAKE, dc.redact("failed with %s" % FAKE))


if __name__ == "__main__":
    unittest.main()
