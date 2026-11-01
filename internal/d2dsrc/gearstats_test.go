package d2dsrc

import "testing"

func TestGearStats(t *testing.T) {
	for _, c := range []struct {
		name, js string
		want     map[int]int
	}{
		{"older fork, no gear key", `{"base":{}}`, nil},
		{"empty gear, staff", `{"gear":[],"weapons":[{"code":"sst","slot":4,"mindam":1,"maxdam":5}]}`,
			map[int]int{19: 0, 25: 0, 21: 1, 22: 5}},
		{"sums plus weapon", `{"gear":[{"stat":19,"param":0,"sum":40,"n":2},{"stat":21,"param":0,"sum":2,"n":1},{"stat":22,"param":0,"sum":3,"n":1},{"stat":25,"param":0,"sum":60,"n":1},{"stat":107,"param":36,"sum":2,"n":1}],"weapons":[{"code":"lsd","slot":4,"mindam":3,"maxdam":10}]}`,
			map[int]int{19: 40, 25: 60, 21: 5, 22: 13}},
		{"left hand only: no 21/22", `{"gear":[{"stat":25,"param":0,"sum":10,"n":1}],"weapons":[{"code":"x","slot":5,"mindam":1,"maxdam":2}]}`,
			map[int]int{19: 0, 25: 10}},
	} {
		o, err := parseObj([]byte(c.js))
		if err != nil {
			t.Fatal(err)
		}
		got := gearStats(o)
		if len(got) != len(c.want) {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
			continue
		}
		for k, v := range c.want {
			if got[k] != v {
				t.Errorf("%s: stat %d = %d, want %d", c.name, k, got[k], v)
			}
		}
	}
}
