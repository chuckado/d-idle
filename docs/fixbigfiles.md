# fixbigfiles: oversized drill logs

Rule: `tools/ci/check.sh` (the CI hygiene job) fails any file over 1 MB outside testdata. DRILLS9 committed five full `-v` run logs (`docs/probes/drills/2026-11-01-drills9-f*.txt`, 1.0 to 1.1 MB each; mostly repeated `combat decision` debug lines).

Fix: each was capped in place to about 218 KB with `d2drill -cap FILE` (`internal/drill.CapLog`): first and last 100 KB, every middle line with DRILL/FAIL/PASS/BLOCKED/ERR/stuck/timeout markers, and one `[capped: ...]` line with the dropped line and byte counts (the full log stayed on the worker host). No other tracked file was over 900 KB.

Guard: cap a pasted log with `go run ./cmd/d2drill -cap FILE` before committing; `tools/ci/check.sh --hygiene-only` now fails any `docs/probes/**` file over 512 KB, before CI does.
