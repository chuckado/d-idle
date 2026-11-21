package main

import "testing"

func TestRestartBudget(t *testing.T) {
	log := []byte(`{"msg":"pace","inst":"x","wall":1,"virtual":600000000000,"ratio":2}` + "\n" + `{"msg":"pace","inst":"x","wall":2,"virtual":1200000000000}` + "\n")
	if m := lastVirtualMin(log); m != 20 {
		t.Fatalf("last pace line is 20 virtual minutes, got %v", m)
	}
	if lastVirtualMin([]byte("nothing")) != 0 {
		t.Fatal("no pace line is 0")
	}
	if h := restHours(30, 20); h != "0.16666666666666666" {
		t.Fatalf("10 minutes left: %s", h)
	}
	if h := restHours(30, 40); h != "0.016666666666666666" {
		t.Fatalf("a spent budget still gets one minute: %s", h)
	}
}
