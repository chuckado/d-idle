# SPDX-License-Identifier: GPL-3.0-or-later
"""Boot helper for Game.exe 1.14d emulator tests.

Delegates to upstream's existing skill_tips.boot() to boot Game.exe and load tables.
"""
import skill_tips


def boot(*args, **kwargs):
    """Boot Game.exe under Unicorn with loaded data tables via skill_tips."""
    return skill_tips.boot()

