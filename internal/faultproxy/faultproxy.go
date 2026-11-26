// Package faultproxy is a line proxy for the d2d control socket that injects faults between the bot and the game: replies that
// come late (a timeout for the bot's client), replies that never come, input that is delayed, and a cancel storm. It is the
// runner-side part of the fault-injection soak (docs/probes/faultsoak-design-2026-11-26.md, P12); it needs nothing of the fork.
// Lab only, never in a measured run: nothing here writes game state, but a faulty channel is not the channel a gate measures.
//
// The d2d control protocol is one command per line and a reply of payload lines ended by a status line, "ok ..." or "err ..."
// (S internal/d2dctl/client.go readReply). One bot connection is served at a time, like the game's own single control client.
package faultproxy

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// Spec is one fault set. Zero fields inject nothing.
type Spec struct {
	// DelayReplyEvery and DelayReplyMs: every Nth command's reply is held for this long before it is forwarded (the bot's
	// client times out when this is above its command timeout, then reads a stale reply).
	DelayReplyEvery int `json:"delay_reply_every,omitempty"`
	DelayReplyMs    int `json:"delay_reply_ms,omitempty"`
	// DropReplyEvery: every Nth command is forwarded to the game but its reply is swallowed (the bot sees only silence).
	DropReplyEvery int `json:"drop_reply_every,omitempty"`
	// DelayInputMs holds every command whose first word is in InputVerbs before it reaches the game.
	DelayInputMs int      `json:"delay_input_ms,omitempty"`
	InputVerbs   []string `json:"input_verbs,omitempty"`
	// CancelEveryMs sends CancelLine to the game on its own every this many ms (a cancel storm: the running action is cancelled
	// again and again). CancelLine is data because the verb is the game's, not ours.
	CancelEveryMs int    `json:"cancel_every_ms,omitempty"`
	CancelLine    string `json:"cancel_line,omitempty"`
}

// Empty reports that the set injects nothing.
func (s Spec) Empty() bool {
	return s.DelayReplyEvery == 0 && s.DropReplyEvery == 0 && s.DelayInputMs == 0 && s.CancelEveryMs == 0
}

// Validate rejects a set that cannot work.
func (s Spec) Validate() error {
	switch {
	case s.DelayReplyEvery < 0 || s.DropReplyEvery < 0 || s.DelayInputMs < 0 || s.CancelEveryMs < 0 || s.DelayReplyMs < 0:
		return fmt.Errorf("fault set: negative number")
	case s.DelayReplyEvery > 0 && s.DelayReplyMs == 0:
		return fmt.Errorf("fault set: delay_reply_every needs delay_reply_ms")
	case s.DelayInputMs > 0 && len(s.InputVerbs) == 0:
		return fmt.Errorf("fault set: delay_input_ms needs input_verbs")
	case s.CancelEveryMs > 0 && strings.TrimSpace(s.CancelLine) == "":
		return fmt.Errorf("fault set: cancel_every_ms needs cancel_line")
	}
	return nil
}

// Event is one injection, stamped so a hit of the invariant monitor can be tied to its fault.
type Event struct {
	At    time.Time `json:"at"`
	Kind  string    `json:"kind"` // delay-reply, drop-reply, delay-input, cancel
	Cmd   string    `json:"cmd,omitempty"`
	Count int       `json:"n"` // commands seen so far
}

// Proxy relays one listener to one upstream socket.
type Proxy struct {
	Spec     Spec
	Listen   string
	Upstream string
	// Now and Sleep are the clock; tests replace them.
	Now   func() time.Time
	Sleep func(time.Duration)

	ln      net.Listener
	mu      sync.Mutex // one exchange with the game at a time (a command and its reply, or a cancel)
	evMu    sync.Mutex
	events  []Event
	n       int
	done    chan struct{}
	conns   map[net.Conn]bool // open bot connections, closed by Stop
	wg      sync.WaitGroup
	stopped bool
}

// Start listens and serves in the background.
func (p *Proxy) Start() error {
	if err := p.Spec.Validate(); err != nil {
		return err
	}
	if p.Now == nil {
		p.Now = time.Now
	}
	if p.Sleep == nil {
		p.Sleep = time.Sleep
	}
	_ = os.Remove(p.Listen)
	ln, err := net.Listen("unix", p.Listen)
	if err != nil {
		return err
	}
	p.ln, p.done, p.conns = ln, make(chan struct{}), map[net.Conn]bool{}
	p.wg.Add(1)
	go p.accept()
	return nil
}

// Stop closes the listener, waits for the handlers and returns the injections.
func (p *Proxy) Stop() []Event {
	if p.ln != nil && !p.stopped {
		p.stopped = true
		close(p.done)
		_ = p.ln.Close()
		p.evMu.Lock()
		for c := range p.conns {
			_ = c.Close()
		}
		p.evMu.Unlock()
		p.wg.Wait()
		_ = os.Remove(p.Listen)
	}
	return p.Events()
}

// Events returns a copy of the injections so far.
func (p *Proxy) Events() []Event {
	p.evMu.Lock()
	defer p.evMu.Unlock()
	return append([]Event(nil), p.events...)
}

func (p *Proxy) note(kind, cmd string, n int) {
	p.evMu.Lock()
	p.events = append(p.events, Event{At: p.Now(), Kind: kind, Cmd: cmd, Count: n})
	p.evMu.Unlock()
}

func (p *Proxy) accept() {
	defer p.wg.Done()
	for {
		c, err := p.ln.Accept()
		if err != nil {
			return
		}
		p.evMu.Lock()
		p.conns[c] = true
		p.evMu.Unlock()
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.serve(c)
		}()
	}
}

func (p *Proxy) serve(client net.Conn) {
	defer client.Close()
	up, err := net.Dial("unix", p.Upstream)
	if err != nil {
		fmt.Fprintf(client, "err proxy: upstream: %v\n", err)
		return
	}
	defer up.Close()
	urd := bufio.NewReader(up)
	stop := make(chan struct{})
	defer close(stop)
	if p.Spec.CancelEveryMs > 0 {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			p.storm(up, urd, stop)
		}()
	}
	crd := bufio.NewReader(client)
	for {
		line, err := crd.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		p.mu.Lock()
		p.n++
		n := p.n
		if p.Spec.DelayInputMs > 0 && hasVerb(line, p.Spec.InputVerbs) {
			p.note("delay-input", line, n)
			p.Sleep(time.Duration(p.Spec.DelayInputMs) * time.Millisecond)
		}
		reply, err := exchange(up, urd, line)
		if err != nil {
			p.mu.Unlock()
			return
		}
		switch {
		case p.Spec.DropReplyEvery > 0 && n%p.Spec.DropReplyEvery == 0:
			p.note("drop-reply", line, n)
			p.mu.Unlock()
			continue
		case p.Spec.DelayReplyEvery > 0 && n%p.Spec.DelayReplyEvery == 0:
			p.note("delay-reply", line, n)
			p.Sleep(time.Duration(p.Spec.DelayReplyMs) * time.Millisecond)
		}
		p.mu.Unlock()
		if _, err := client.Write([]byte(reply)); err != nil {
			return
		}
	}
}

// storm sends the cancel line between commands; its reply is swallowed.
func (p *Proxy) storm(up net.Conn, urd *bufio.Reader, stop <-chan struct{}) {
	t := time.NewTicker(time.Duration(p.Spec.CancelEveryMs) * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-p.done:
			return
		case <-t.C:
			p.mu.Lock()
			p.note("cancel", p.Spec.CancelLine, p.n)
			_, err := exchange(up, urd, p.Spec.CancelLine)
			p.mu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

// exchange sends one command and reads its whole reply up to and including the status line.
func exchange(up net.Conn, urd *bufio.Reader, line string) (string, error) {
	if _, err := up.Write([]byte(line + "\n")); err != nil {
		return "", err
	}
	var sb strings.Builder
	for {
		l, err := urd.ReadString('\n')
		if err != nil {
			return "", err
		}
		sb.WriteString(l)
		t := strings.TrimRight(l, "\r\n")
		if t == "ok" || t == "err" || strings.HasPrefix(t, "ok ") || strings.HasPrefix(t, "err ") {
			return sb.String(), nil
		}
	}
}

func hasVerb(line string, verbs []string) bool {
	f := strings.Fields(line)
	if len(f) == 0 {
		return false
	}
	for _, v := range verbs {
		if f[0] == v {
			return true
		}
	}
	return false
}
