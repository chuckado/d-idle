package boss

import "testing"

func TestOptsCanTeleportFnWinsAndIsLive(t *testing.T) {
	if (Opts{}).canTeleport() {
		t.Fatal("zero Opts must not teleport")
	}
	if !(Opts{CanTeleport: true}).canTeleport() {
		t.Fatal("static flag ignored")
	}
	on := false
	o := Opts{CanTeleport: true, CanTeleportFn: func() bool { return on }}
	if o.canTeleport() {
		t.Fatal("the function must win over the flag")
	}
	on = true // learned after the runner was built
	if !o.canTeleport() {
		t.Fatal("function result not read at use time")
	}
}
