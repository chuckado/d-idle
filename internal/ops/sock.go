package ops

import (
	"bufio"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Start serves the operator socket in stateDir when Enabled says so, and returns a stop func (never nil). When the
// operator control is off it creates nothing. A failure to listen is returned; the bot goes on without the socket.
func Start(stateDir string, desk func() Desk, getenv func(string) string, logf func(string, ...any)) (func(), error) {
	if on, why := Enabled(getenv); !on {
		if logf != nil && getenv("D2BOT_OPERATOR") != "" {
			logf("operator control not started", "why", why)
		}
		return func() {}, nil
	}
	path := filepath.Join(stateDir, SockName)
	ln, err := Listen(path)
	if err != nil {
		return func() {}, err
	}
	s := New(Config{Desk: desk, Getenv: getenv, Logf: logf})
	go s.Serve(ln)
	if logf != nil {
		logf("operator control on", "socket", SockName)
	}
	return func() { ln.Close(); os.Remove(path) }, nil
}

// Listen binds the socket under a temporary name, sets mode 0600 and renames it into place, so the socket is never
// reachable with a looser mode.
func Listen(path string) (net.Listener, error) {
	tmp := path + ".new"
	os.Remove(tmp)
	ln, err := net.Listen("unix", tmp)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		ln.Close()
		os.Remove(tmp)
		return nil, err
	}
	os.Remove(path) // a stale socket of a previous run
	if err := os.Rename(tmp, path); err != nil {
		ln.Close()
		os.Remove(tmp)
		return nil, err
	}
	return ln, nil
}

// Serve answers one request per connection until the listener closes.
func (s *Server) Serve(ln net.Listener) {
	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		go s.conn(c)
	}
}

func (s *Server) conn(c net.Conn) {
	defer c.Close()
	c.SetDeadline(time.Now().Add(15 * time.Second))
	var rq Request
	dec := json.NewDecoder(bufio.NewReaderSize(c, 8<<10))
	dec.DisallowUnknownFields()
	rsp := no("request", "send one JSON request line")
	if err := dec.Decode(&rq); err == nil {
		rsp = s.Handle(rq)
	}
	_ = json.NewEncoder(c).Encode(rsp)
}

// Call sends one request and reads the answer.
func Call(sock string, rq Request) (Response, error) {
	c, err := net.DialTimeout("unix", sock, 5*time.Second)
	if err != nil {
		return Response{}, errors.New("operator socket not reachable (is D2BOT_OPERATOR=1 set on the container?): " + err.Error())
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(20 * time.Second))
	if err := json.NewEncoder(c).Encode(rq); err != nil {
		return Response{}, err
	}
	var rsp Response
	if err := json.NewDecoder(c).Decode(&rsp); err != nil {
		return Response{}, err
	}
	return rsp, nil
}
