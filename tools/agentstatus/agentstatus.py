#!/usr/bin/env python3
"""Stub: the status board CLI is in the private ops repo (tools/ops.sh run)."""
import os
import sys

ops = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "ops.sh")
os.execv(ops, [ops, "run", "agentstatus/agentstatus.py"] + sys.argv[1:])
