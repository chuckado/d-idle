package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"d2lod-bot/internal/scenario"
)

// ledgerSpool is the per-output-directory copy of every line the runner shipped: written on the host that ran the
// scenario, so a lost pull cannot lose the record (design H7), and the source for `ledger add` when shipping failed.
const ledgerSpool = "ledger.jsonl"

// shipLedger records one finished run: always into <out>/ledger.jsonl, and through $PG_LEDGER_CMD when set (a command
// that reads the JSON line on stdin and appends it to the ledger on the board host; its address is host config, never
// in the repo). Failures are reported on stderr and never change the run's verdict.
func shipLedger(res scenario.Result, tier, out string, started time.Time) {
	job := os.Getenv("PG_JOB")
	if job == "" {
		job = filepath.Base(out)
	}
	rec, err := scenario.NewLedgerRecord(res, scenario.LedgerCtx{Job: job, Event: os.Getenv("PG_EVENT"), Branch: os.Getenv("PG_BRANCH"),
		Card: os.Getenv("PG_CARD"), Lane: os.Getenv("PG_LANE"), Tier: tier, Started: started})
	if err != nil {
		fmt.Fprintf(os.Stderr, "d2scenario: not recorded in the ledger: %v\n", err)
		return
	}
	line, err := rec.LedgerLine()
	if err != nil {
		fmt.Fprintf(os.Stderr, "d2scenario: not recorded in the ledger: %v\n", err)
		return
	}
	if err := appendLine(filepath.Join(out, ledgerSpool), line); err != nil {
		fmt.Fprintf(os.Stderr, "d2scenario: ledger spool: %v\n", err)
	}
	if cmd := os.Getenv("PG_LEDGER_CMD"); cmd != "" {
		c := exec.Command("sh", "-c", cmd)
		c.Stdin = bytes.NewReader(append(line, '\n'))
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "d2scenario: ledger ingest command failed (%v); the line is in %s\n", err, filepath.Join(out, ledgerSpool))
		}
	}
}

// appendLine appends one line with a single O_APPEND write (atomic for lines below the pipe buffer on a local file).
func appendLine(path string, line []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(append([]byte(nil), line...), '\n'))
	return err
}

// cmdLedger: `ledger status-doc -in L [-branch B] [-out F]` renders the generated status page; `ledger check -in L`
// validates every line. There is no hand-written append: lines come from the runner (shipLedger) only.
func cmdLedger(args []string, w io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ledger status-doc|check -in ledger.jsonl [-branch B] [-out file]")
	}
	sub := args[0]
	fs := newFlags("ledger " + sub)
	in := fs.String("in", "", "ledger file (JSON lines)")
	branch := fs.String("branch", "main", "status-doc: branch to show (empty: all)")
	out := fs.String("out", "", "status-doc: write here instead of stdout")
	if _, err := parseArgs(fs, args[1:]); err != nil {
		return err
	}
	if *in == "" {
		return fmt.Errorf("ledger %s: -in is required", sub)
	}
	f, err := os.Open(*in)
	if err != nil {
		return err
	}
	defer f.Close()
	recs, err := scenario.ReadLedger(f)
	if err != nil {
		return err
	}
	switch sub {
	case "check":
		for i, r := range recs {
			if err := r.Validate(); err != nil {
				return fmt.Errorf("line %d: %v", i+1, err)
			}
		}
		fmt.Fprintf(w, "ledger ok: %d lines\n", len(recs))
	case "status-doc":
		doc := scenario.StatusDoc(recs, strings.TrimSpace(*branch))
		if *out == "" {
			_, err = io.WriteString(w, doc)
			return err
		}
		return os.WriteFile(*out, []byte(doc), 0o644)
	default:
		return fmt.Errorf("ledger: unknown subcommand %q", sub)
	}
	return nil
}
