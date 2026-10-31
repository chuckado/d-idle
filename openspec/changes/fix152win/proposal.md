## Why

Upstream PR #152 (vendor prices) failed its Windows (MSVC) job on head 70e54f57: C4456 in the upstream test file tests/test_rules.cpp, promoted to an error. Our price test block declared a local `unknown` that the identify loop further down the same function reuses; the loop is upstream code, our declaration is the one that makes it shadow.

## What Changes

- one rename in the PR branch (`unknown` to `unid_magic` in the item_price block), no behaviour change; pushed as a fast-forward to the PR branch
- a dated `posted:` line in docs/research/pr-price-evidence-pack.md
