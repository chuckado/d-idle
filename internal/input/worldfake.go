package input

import "fmt"

// FakeWorld is a WorldVerbs for tests: it records what was asked and can refuse any call. Fail keys are the verb names
// ("buy", "sell", "waypoint", "portal", "pickup"); a *fakeRefusal is a rule refusal, any other error a fallback reason.
type FakeWorld struct {
	Calls []string
	Fail  map[string]error
}

type fakeRefusal struct{ Code string }

func (e fakeRefusal) RefusalCode() string { return e.Code }
func (e fakeRefusal) Error() string       { return "refused: " + e.Code }
func (fakeRefusal) RuleRefusal() bool     { return true }

// Refusal builds a rule refusal for FakeWorld.Fail.
func Refusal(code string) error { return fakeRefusal{Code: code} }

func (f *FakeWorld) do(verb, call string) error {
	f.Calls = append(f.Calls, call)
	return f.Fail[verb]
}

func (f *FakeWorld) BuyAt(tab, x, y int) error {
	return f.do("buy", fmt.Sprintf("buy %d %d,%d", tab, x, y))
}
func (f *FakeWorld) SellItem(id uint32) error { return f.do("sell", fmt.Sprintf("sell %d", id)) }
func (f *FakeWorld) TakeWaypoint(level int) error {
	return f.do("waypoint", fmt.Sprintf("waypoint %d", level))
}
func (f *FakeWorld) EnterPortal(which int) error {
	return f.do("portal", fmt.Sprintf("portal %d", which))
}
func (f *FakeWorld) PickupItem(id uint32) error { return f.do("pickup", fmt.Sprintf("pickup %d", id)) }

var _ WorldVerbs = (*FakeWorld)(nil)
