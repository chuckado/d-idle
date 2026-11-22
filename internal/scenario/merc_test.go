package scenario

import "testing"

func TestMercHired(t *testing.T) {
	str := func(s string) *string { return &s }
	zero := 0
	cases := []struct {
		name string
		a    Assert
		m    *string
		ok   bool
		unob bool
	}{
		{"rogue present", Assert{Kind: AMercHired}, str("12,7:RG"), true, false},
		{"class match", Assert{Kind: AMercHired, Code: "RG"}, str("12,7:RG"), true, false},
		{"class mismatch (desert merc wanted)", Assert{Kind: AMercHired, Code: "0A"}, str("12,7:RG"), false, false},
		{"none, must fail", Assert{Kind: AMercHired}, str("-"), false, false},
		{"none wanted", Assert{Kind: AMercHired, Max: &zero}, str("-"), true, false},
		{"none wanted but present", Assert{Kind: AMercHired, Max: &zero}, str("1,1:0A"), false, false},
		{"unread", Assert{Kind: AMercHired}, nil, false, true},
	}
	for _, c := range cases {
		r := evalOne(c.a, State{Merc: c.m})
		if r.OK != c.ok || r.Unobserved != c.unob {
			t.Errorf("%s: %+v", c.name, r)
		}
	}
}
