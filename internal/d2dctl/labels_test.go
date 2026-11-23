package d2dctl

import (
	"context"
	"errors"
	"testing"
)

func TestLabelsReadsTheStateBack(t *testing.T) {
	s, c := startStub(t, map[string]string{
		"labels on":     "labels=on\nok\n",
		"labels off":    "labels=off\nok\n",
		"labels toggle": "labels=on\nok\n",
		"labels status": "labels=off\nok\n",
	}, nil)
	for mode, want := range map[string]bool{"on": true, "off": false, "toggle": true, "status": false} {
		if got, err := c.Labels(context.Background(), mode); err != nil || got != want {
			t.Fatalf("%s: %v %v", mode, got, err)
		}
	}
	if len(s.seen) != 4 {
		t.Fatalf("sent %v", s.seen)
	}
	if _, err := c.Labels(context.Background(), "sideways"); err == nil || len(s.seen) != 4 {
		t.Fatalf("a bad mode must send nothing: %v %v", err, s.seen)
	}
}

func TestLabelsRefusedAndMalformed(t *testing.T) {
	_, c := startStub(t, map[string]string{"labels on": "err lab verbs off: labels needs d2d --item-verbs\n", "labels status": "ok\n"}, nil)
	var ce *CmdError
	if _, err := c.Labels(context.Background(), "on"); !errors.As(err, &ce) {
		t.Fatalf("want a CmdError, got %v", err)
	}
	if _, err := c.Labels(context.Background(), "status"); err == nil {
		t.Fatal("a reply without the state is an error")
	}
}
