import unittest, capture

class Capture(unittest.TestCase):
    def test_only_read_verbs_pass(self):
        for v in ("debug level", "debug levelmap", "debug blocked 1.5 2.5", "debug grid 5", "state"):
            self.assertTrue(capture.allowed(v), v)
        for v in ("debug goto 28", "debug warp 3 4", "debug kill", "debug statpts 5", "cmd move 1 2", "debug levelxyz"):
            self.assertFalse(capture.allowed(v), v)
        with self.assertRaises(ValueError):
            capture.ask("/nonexistent", "debug goto 28")
    def test_level_of(self):
        self.assertEqual(capture.level_of("ok 28 25.00 28.00 50 56 2982 904"), 28)
        self.assertIsNone(capture.level_of("err no game"))

if __name__ == "__main__":
    unittest.main()
