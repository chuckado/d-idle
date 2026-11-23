package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"d2lod-bot/internal/ops"
)

const opsUsage = `usage: d2bot ops quest status [--json]
       d2bot ops quest <start|retry|cancel|handin> <quest> [--chunk ID] --why "reason"
       d2bot ops quest auto <quest> --on|--off --why "reason"
       d2bot ops quest wait <quest> [--chunk ID] [--timeout 600]
flags: --sock PATH (default $D2BOT_OPERATOR_SOCK, else the sandbox container's state directory)`

// defaultOpsSock is where the container keeps the instance state (docs/operator.md).
const defaultOpsSock = "/data/instances/main/state/" + ops.SockName

// opsMain is `d2bot ops ...`: the client of the operator socket. It returns the process exit code: 0 done, 1 refused
// or not finished, 2 usage or timeout of a wait.
func opsMain(args []string, out io.Writer) (int, error) {
	if len(args) < 2 || args[0] != "quest" {
		return 2, errors.New(opsUsage)
	}
	verb, rest := args[1], args[2:]
	fs := flag.NewFlagSet("ops quest "+verb, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	chunk := fs.String("chunk", "", "chunk id")
	why := fs.String("why", "", "reason (required for every verb that changes something)")
	sock := fs.String("sock", os.Getenv("D2BOT_OPERATOR_SOCK"), "socket path")
	on := fs.Bool("on", false, "auto on")
	off := fs.Bool("off", false, "auto off")
	asJSON := fs.Bool("json", false, "status: raw rows as JSON lines")
	timeout := fs.Int("timeout", 600, "wait: seconds")
	var pos []string
	for len(rest) > 0 { // flags may follow the quest id
		if err := fs.Parse(rest); err != nil {
			return 2, fmt.Errorf("%w\n%s", err, opsUsage)
		}
		if fs.NArg() == 0 {
			break
		}
		pos, rest = append(pos, fs.Arg(0)), fs.Args()[1:]
	}
	if *sock == "" {
		*sock = defaultOpsSock
	}
	quest := ""
	if verb != "status" {
		if len(pos) != 1 {
			return 2, errors.New(opsUsage)
		}
		quest = pos[0]
	} else if len(pos) != 0 {
		return 2, errors.New(opsUsage)
	}
	switch verb {
	case "status":
		r, err := ops.Call(*sock, ops.Request{Verb: "status"})
		if err != nil {
			return 1, err
		}
		if !r.OK || r.Board == nil {
			fmt.Fprintf(out, "ok=false code=%s reason=%q\n", r.Code, r.Reason)
			return 1, nil
		}
		printBoard(out, r.Board, *asJSON)
		return 0, nil
	case "start", "retry", "cancel", "handin", "auto":
		rq := ops.Request{Verb: verb, Quest: quest, Chunk: *chunk, Why: *why}
		if verb == "auto" {
			if *on == *off {
				return 2, errors.New("auto needs exactly one of --on and --off")
			}
			rq.On = on
		}
		r, err := ops.Call(*sock, rq)
		if err != nil {
			return 1, err
		}
		state := ""
		if r.Reply != nil {
			state = r.Reply.State
		}
		fmt.Fprintf(out, "ok=%t code=%s state=%s reason=%q\n", r.OK, r.Code, state, r.Reason)
		if r.OK {
			return 0, nil
		}
		return 1, nil
	case "wait":
		poll := func() (*ops.Board, error) {
			r, err := ops.Call(*sock, ops.Request{Verb: "status"})
			if err != nil {
				return nil, err
			}
			if r.Board == nil {
				return nil, errors.New(r.Reason)
			}
			return r.Board, nil
		}
		res := ops.Wait(poll, quest, *chunk, time.Duration(*timeout)*time.Second, 5*time.Second, time.Sleep, time.Now)
		fmt.Fprintf(out, "RESULT outcome=%s quest=%s chunk=%s state=%s step=%s progress=%d code=%s text=%q\n",
			res.Outcome, res.Quest, res.Chunk, res.State, res.Step, res.Progress, res.Code, res.Text)
		switch res.Outcome {
		case ops.OutFinished:
			return 0, nil
		case ops.OutTimeout:
			return 2, nil
		}
		return 1, nil
	}
	return 2, errors.New(opsUsage)
}

func printBoard(out io.Writer, b *ops.Board, asJSON bool) {
	fmt.Fprintf(out, "board enabled=%t difficulty=%s why=%q\n", b.Enabled, b.Difficulty, b.Why)
	for _, r := range b.Rows {
		if asJSON {
			fmt.Fprintf(out, `{"quest":%q,"chunk":%q,"state":%q,"step":%q,"progress":%d,"code":%q,"text":%q}`+"\n", r.Quest, r.Chunk, r.State, r.Step, r.Progress, r.Code, r.Text)
			continue
		}
		fmt.Fprintf(out, "%s/%s\t%s\tstep=%s\tcode=%s\t%s\n", r.Quest, r.Chunk, r.State, r.Step, r.Code, strings.ReplaceAll(r.Text, "\n", " "))
	}
}
