package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

// recorder keeps every log record as one line, so a drill can quote the bot's own failure lines, and forwards to the
// stderr handler only what that handler's level lets through.
type recorder struct {
	next slog.Handler
	st   *recStore
}

type recStore struct {
	mu    sync.Mutex
	lines []string
}

func newRecorder(next slog.Handler) *recorder { return &recorder{next: next, st: &recStore{}} }

func (r *recorder) Enabled(context.Context, slog.Level) bool { return true }

func (r *recorder) Handle(ctx context.Context, rec slog.Record) error {
	var b strings.Builder
	b.WriteString(rec.Message)
	rec.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(&b, " %s=%s", a.Key, a.Value.String())
		return true
	})
	r.st.mu.Lock()
	r.st.lines = append(r.st.lines, b.String())
	r.st.mu.Unlock()
	if r.next.Enabled(ctx, rec.Level) {
		return r.next.Handle(ctx, rec)
	}
	return nil
}

func (r *recorder) WithAttrs(a []slog.Attr) slog.Handler {
	return &recorder{next: r.next.WithAttrs(a), st: r.st}
}

func (r *recorder) WithGroup(n string) slog.Handler {
	return &recorder{next: r.next.WithGroup(n), st: r.st}
}

// Mark is the number of lines so far; Since returns the lines after a mark.
func (r *recorder) Mark() int {
	r.st.mu.Lock()
	defer r.st.mu.Unlock()
	return len(r.st.lines)
}

func (r *recorder) Since(mark int) []string {
	r.st.mu.Lock()
	defer r.st.mu.Unlock()
	if mark > len(r.st.lines) {
		mark = len(r.st.lines)
	}
	return append([]string(nil), r.st.lines[mark:]...)
}

// runProbe runs semicolon separated verbs and prints each reply (debug verbs only through the cheat ledger).
func (l *lab) runProbe(ctx context.Context, verbs string) int {
	for _, v := range strings.Split(verbs, ";") {
		v = strings.TrimSpace(v)
		switch {
		case v == "":
		case strings.HasPrefix(v, "debug "):
			out, err := l.Cheat(ctx, "drills:probe", v, "probe: find out what the verb does (docs/probes/drills/README.md)")
			fmt.Printf("> %s\n%s err=%v\n", v, out, err)
		case strings.HasPrefix(v, "step "):
			var n int
			fmt.Sscanf(v, "step %d", &n)
			fmt.Printf("> %s err=%v\n", v, l.Tick(ctx, n))
		default:
			r, err := l.be.Client.Do(ctx, v)
			if r != nil {
				fmt.Printf("> %s\n%s\n%s err=%v\n", v, strings.Join(r.Lines, "\n"), r.Rest, err)
			} else {
				fmt.Printf("> %s err=%v\n", v, err)
			}
		}
	}
	return 0
}
