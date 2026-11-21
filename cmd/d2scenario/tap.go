package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"d2lod-bot/internal/scenario"
)

// tapper plays the part of the person on the quest board of the watch screen: it taps Start on the offered chunk of one
// quest and Hand in when the board waits for it, exactly the two taps of a player (V docs/probes/cainl3/poll.py, the
// harness of the unattended Search for Cain run). It reads and writes only the board's own routes (watch.QuestsRoute);
// it never touches game state.
type tapper struct {
	view, answer, token, quest string
	chunks                     map[string]bool // empty: every chunk
	client                     *http.Client
	log                        io.Writer
	Taps                       []string // "start scroll", "handin read" in order
	lastWhy                    string   // the last "not startable" line written, so a long wait logs each change once
}

func newTapper(view, answer, token, quest string, chunks []string, log io.Writer) *tapper {
	t := &tapper{view: view, answer: answer, token: token, quest: quest, chunks: map[string]bool{}, client: &http.Client{Timeout: 10 * time.Second}, log: log}
	for _, c := range chunks {
		t.chunks[c] = true
	}
	return t
}

type boardView struct {
	Running *struct {
		Quest, Chunk, State string
	} `json:"running"`
	Awaiting *struct {
		Quest, Chunk, State string
	} `json:"awaiting"`
	Blocked *struct {
		Quest, Chunk, Why string
	} `json:"blocked"`
	HandIn any `json:"handIn"`
	Offers []struct {
		Quest, Chunk string
		Startable    bool
	} `json:"offers"`
	History []struct{ Quest, Chunk string } `json:"history"`
	Quests  []struct {
		ID     string
		State  string
		Chunks []struct{ ID, State, Readiness, Why string }
	} `json:"quests"`
}

func (t *tapper) allowed(chunk string) bool { return len(t.chunks) == 0 || t.chunks[chunk] }

func (t *tapper) get(ctx context.Context) (boardView, error) {
	var v boardView
	req, _ := http.NewRequestWithContext(ctx, "GET", t.view+"/api/quests", nil)
	req.Header.Set("Authorization", "Bearer "+t.token)
	resp, err := t.client.Do(req)
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return v, fmt.Errorf("board: HTTP %d", resp.StatusCode)
	}
	return v, json.NewDecoder(resp.Body).Decode(&v)
}

func (t *tapper) post(ctx context.Context, path string) string {
	req, _ := http.NewRequestWithContext(ctx, "POST", t.answer+path, bytes.NewReader([]byte("{}")))
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return "error " + err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
	return fmt.Sprintf("HTTP %d %s", resp.StatusCode, strings.TrimSpace(string(b)))
}

// Step does one poll: it returns the tap it made ("" for none) and whether the board is blocked on the quest (a parked chunk
// is the bot's verdict, the tapper does not retry it).
func (t *tapper) Step(ctx context.Context) (tap string, blocked bool, err error) {
	v, err := t.get(ctx)
	if err != nil {
		return "", false, err
	}
	if v.Blocked != nil && v.Blocked.Quest == t.quest {
		return "", true, nil
	}
	switch {
	case v.Awaiting != nil && v.Awaiting.Quest == t.quest && v.HandIn == nil:
		if !t.allowed(v.Awaiting.Chunk) {
			return "", false, nil
		}
		tap = "handin " + v.Awaiting.Chunk
		fmt.Fprintf(t.log, "%s tap %s: %s\n", time.Now().Format("15:04:05"), tap, t.post(ctx, "/api/quests/"+t.quest+"/handin"))
	case v.Running == nil && v.Awaiting == nil:
		for _, o := range v.Offers {
			if o.Quest == t.quest && o.Startable && t.allowed(o.Chunk) {
				tap = "start " + o.Chunk
				fmt.Fprintf(t.log, "%s tap %s: %s\n", time.Now().Format("15:04:05"), tap, t.post(ctx, "/api/quests/"+t.quest+"/start"))
				break
			}
		}
	}
	if tap != "" {
		t.Taps = append(t.Taps, tap)
	} else if v.Running == nil && v.Awaiting == nil {
		t.logWhyNot(v)
	}
	return tap, false, nil
}

// logWhyNot writes the board's own reason when the wanted chunk is not startable (state, readiness word, why), once per
// change: a quest that is never offered must say why in the taps log (card andariel1; before it the log stayed silent).
func (t *tapper) logWhyNot(v boardView) {
	for _, q := range v.Quests {
		if q.ID != t.quest {
			continue
		}
		for _, c := range q.Chunks {
			if !t.allowed(c.ID) {
				continue
			}
			line := fmt.Sprintf("not startable %s/%s: state %s readiness %q why %q", q.ID, c.ID, c.State, c.Readiness, c.Why)
			if line != t.lastWhy {
				t.lastWhy = line
				fmt.Fprintf(t.log, "%s %s\n", time.Now().Format("15:04:05"), line)
			}
		}
	}
}

// Run polls every interval until ctx ends or the board parks the quest.
func (t *tapper) Run(ctx context.Context, interval time.Duration) {
	for ctx.Err() == nil {
		if _, blocked, err := t.Step(ctx); blocked {
			fmt.Fprintf(t.log, "%s board parked the quest\n", time.Now().Format("15:04:05"))
			return
		} else if err != nil && ctx.Err() == nil {
			fmt.Fprintf(t.log, "%s board read failed: %v\n", time.Now().Format("15:04:05"), err)
		}
		select {
		case <-ctx.Done():
		case <-time.After(interval):
		}
	}
}

// newWatchToken writes a random token file (mode 0600) the bot's --watch-token-file reads, and returns the token.
func newWatchToken(path string) (string, error) {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		return "", err
	}
	defer f.Close()
	b := make([]byte, 24)
	if _, err := io.ReadFull(f, b); err != nil {
		return "", err
	}
	tok := fmt.Sprintf("%x", b)
	return tok, os.WriteFile(path, []byte(tok), 0o600)
}

// runAPICalls sends the scenario's trigger.api calls to the bot's answer listener, once each in order, and writes the
// reply to the taps log. The token is the scenario-local one of newWatchToken.
func runAPICalls(ctx context.Context, base, token string, calls []scenario.APICall, log io.Writer) {
	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()
	for _, c := range calls {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(start.Add(time.Duration(c.After) * time.Second))):
		}
		method := c.Method
		if method == "" {
			method = "POST"
		}
		b, _ := json.Marshal(c.Body)
		req, _ := http.NewRequestWithContext(ctx, method, base+c.Path, bytes.NewReader(b))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(log, "%s api %s %s: error %v\n", time.Now().Format("15:04:05"), method, c.Path, err)
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 300))
		resp.Body.Close()
		fmt.Fprintf(log, "%s api %s %s: HTTP %d %s\n", time.Now().Format("15:04:05"), method, c.Path, resp.StatusCode, strings.TrimSpace(string(body)))
	}
}
