package main

import "testing"

func TestFarmOffNeed(t *testing.T) {
	if _, ok := farmOffNeed(""); ok {
		t.Fatal("no row while the loop is on")
	}
	n, ok := farmOffNeed("3 fresh games in 30 minutes")
	if !ok || n.Sev != "info" || n.Kind != "farm_reset_off" || n.Actions == nil {
		t.Fatalf("%+v", n)
	}
}
