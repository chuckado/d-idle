package support

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

// Environment names of the upload endpoint (docs/support-bundle.md); the token comes from the friends pack .env,
// never from git and never from the watch token.
const (
	EnvURL   = "D2BOT_SUPPORT_URL"
	EnvToken = "D2BOT_SUPPORT_TOKEN"
)

var codeRx = regexp.MustCompile(`^[A-Z2-9]{8}$`)

// Sender posts a zip to the receiver (cmd/d2supportrecv).
type Sender struct {
	URL    string
	Token  string
	Client *http.Client // nil: a 60 s client that never follows a redirect (the token would follow it)
}

// Configured reports whether both the URL and the token are set.
func (s Sender) Configured() bool { return s.URL != "" && s.Token != "" }

// ValidURL accepts https, or http on a loopback host (tests and a local receiver).
func ValidURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return errors.New("the support upload address is not a valid URL")
	}
	if u.Scheme == "https" {
		return nil
	}
	if u.Scheme == "http" {
		if h := u.Hostname(); h == "localhost" || (net.ParseIP(h) != nil && net.ParseIP(h).IsLoopback()) {
			return nil
		}
	}
	return errors.New("the support upload address must be https")
}

// Send uploads zip and returns the support code. Errors are words for the sheet, never the token or the URL.
func (s Sender) Send(ctx context.Context, zip []byte) (string, error) {
	if !s.Configured() {
		return "", errors.New("uploading is not set up on this install; save the file instead")
	}
	if err := ValidURL(s.URL); err != nil {
		return "", err
	}
	if len(zip) == 0 || len(zip) > MaxZip {
		return "", errors.New("the bundle is empty or too large")
	}
	c := s.Client
	if c == nil {
		c = &http.Client{Timeout: 60 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL, bytes.NewReader(zip))
	if err != nil {
		return "", errors.New("could not build the upload request")
	}
	req.Header.Set("Content-Type", "application/zip")
	req.Header.Set("Authorization", "Bearer "+s.Token)
	resp, err := c.Do(req)
	if err != nil {
		return "", errors.New("could not reach the upload site; save the file instead")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	switch {
	case resp.StatusCode == http.StatusOK:
		var r struct {
			Code string `json:"code"`
		}
		if json.Unmarshal(body, &r) != nil || !codeRx.MatchString(r.Code) {
			return "", errors.New("the upload site answered without a support code")
		}
		return r.Code, nil
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		return "", errors.New("the upload key was refused (it may have been revoked); save the file instead")
	case resp.StatusCode == http.StatusRequestEntityTooLarge:
		return "", errors.New("the bundle is too large for the upload site")
	case resp.StatusCode == http.StatusTooManyRequests:
		return "", errors.New("too many uploads just now; try again in an hour or save the file instead")
	}
	return "", fmt.Errorf("the upload site answered %d", resp.StatusCode)
}
