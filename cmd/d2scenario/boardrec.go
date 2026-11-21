package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"d2lod-bot/internal/scenario"
)

// parseReply reads a questboard reply body (watch.quests/1 verbs: ok, already, code, reason, and from card bsm6 changed).
func parseReply(index int, path string, status int, body []byte) scenario.APIReply {
	var r struct {
		OK      bool   `json:"ok"`
		Already bool   `json:"already"`
		Code    string `json:"code"`
		Reason  string `json:"reason"`
		Changed []struct {
			Quest, Chunk string
		} `json:"changed"`
	}
	_ = json.Unmarshal(body, &r)
	out := scenario.APIReply{Index: index, Path: path, Status: status, OK: r.OK, Already: r.Already, Code: r.Code, Reason: r.Reason}
	for _, c := range r.Changed {
		out.Changed = append(out.Changed, c.Quest+"/"+c.Chunk)
	}
	return out
}

// parseBoard reads GET /api/quests into chunk rows. The reason is the view's reason code: a string or an object with
// `code` (none before card bsm6, then it stays empty).
func parseBoard(body []byte) []scenario.BoardChunk {
	var v struct {
		Quests []struct {
			ID     string `json:"id"`
			Chunks []struct {
				ID     string          `json:"id"`
				State  string          `json:"state"`
				Reason json.RawMessage `json:"reason"`
			} `json:"chunks"`
		} `json:"quests"`
	}
	if json.Unmarshal(body, &v) != nil {
		return nil
	}
	var out []scenario.BoardChunk
	for _, q := range v.Quests {
		for _, c := range q.Chunks {
			bc := scenario.BoardChunk{Quest: q.ID, Chunk: c.ID, State: c.State}
			var s string
			var o struct{ Code string }
			switch {
			case json.Unmarshal(c.Reason, &s) == nil:
				bc.Reason = s
			case json.Unmarshal(c.Reason, &o) == nil:
				bc.Reason = o.Code
			}
			out = append(out, bc)
		}
	}
	return out
}

// boardRow is one line of the samples file: the wall time and the chunk rows.
type boardRow struct {
	At     time.Time             `json:"at"`
	Chunks []scenario.BoardChunk `json:"chunks"`
}

// sampleBoard reads GET /api/quests every `every` until ctx ends and appends one JSON line per sample to out.
func sampleBoard(ctx context.Context, view, token string, every time.Duration, out io.Writer) {
	client := &http.Client{Timeout: 10 * time.Second}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(every):
		}
		req, _ := http.NewRequestWithContext(ctx, "GET", view+"/api/quests", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		resp.Body.Close()
		if chunks := parseBoard(body); len(chunks) > 0 {
			b, _ := json.Marshal(boardRow{At: time.Now(), Chunks: chunks})
			out.Write(append(b, '\n'))
		}
	}
}

// boardFiles are the side files of a run, next to the bot log.
func boardFiles(trace string) (samples, replies string) {
	base := strings.TrimSuffix(trace, ".bot.log")
	return base + ".board.jsonl", base + ".replies.json"
}

// addBoardFacts loads the side files into the state: the board samples (virtual minutes by the linear map of the bot log
// span, as the log lines) and the api replies.
func addBoardFacts(st *scenario.State, s scenario.Scenario, trace string) {
	samples, replies := boardFiles(trace)
	if b, err := os.ReadFile(replies); err == nil {
		_ = json.Unmarshal(b, &st.APIReplies)
	}
	b, err := os.ReadFile(samples)
	if err != nil {
		return
	}
	var rows []boardRow
	for _, l := range strings.Split(string(b), "\n") {
		var r boardRow
		if l != "" && json.Unmarshal([]byte(l), &r) == nil {
			rows = append(rows, r)
		}
	}
	log, _ := os.ReadFile(trace)
	first, last, ok := scenario.LogSpan(log)
	if len(rows) == 0 || !ok {
		return
	}
	at := make([]time.Time, len(rows))
	for i, r := range rows {
		at[i] = r.At
	}
	ts := scenario.MapSampleTimes(at, first, last, s.Budget.VirtualMin)
	for i, r := range rows {
		st.Board = append(st.Board, scenario.BoardSample{T: ts[i], Chunks: r.Chunks})
	}
	st.Has = append(st.Has, "board")
}
