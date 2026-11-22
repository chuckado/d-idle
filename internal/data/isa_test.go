package data

import "testing"

// isaDB is a type table with a diamond, a cycle, a dangling Equiv and an empty Equiv, shaped like itemtypes.txt.
func isaDB() *DB {
	rows := []*ItemType{
		{Code: "weap"}, {Code: "swor", Equiv: [2]string{"weap", ""}}, {Code: "mele", Equiv: [2]string{"weap", ""}},
		{Code: "axe", Equiv: [2]string{"mele", "swor"}}, {Code: "pott"}, {Code: "hpot", Equiv: [2]string{"pott", "gone"}},
		{Code: "cyca", Equiv: [2]string{"cycb", ""}}, {Code: "cycb", Equiv: [2]string{"cyca", "weap"}},
	}
	db := &DB{typeByCode: map[string]*ItemType{}}
	for _, r := range rows {
		db.typeByCode[r.Code] = r
	}
	return db
}

func TestIsAMatchesWalkOverWholeTable(t *testing.T) {
	db := isaDB()
	codes := []string{"", "nope"}
	for c := range db.typeByCode {
		codes = append(codes, c)
	}
	codes = append(codes, "gone")
	for _, built := range []bool{false, true} {
		if built {
			db.buildTypeAnc()
		}
		for _, a := range codes {
			for _, b := range codes {
				if got, want := db.IsA(a, b), db.isAWalk(a, b); got != want && !(a == b && a == "") {
					t.Errorf("built=%v IsA(%q,%q)=%v, walk says %v", built, a, b, got, want)
				}
			}
		}
	}
}

func BenchmarkIsA(b *testing.B) {
	db := isaDB()
	db.buildTypeAnc()
	b.Run("table", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = db.IsA("axe", "weap") || db.IsA("axe", "pott")
		}
	})
	b.Run("walk", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = db.isAWalk("axe", "weap") || db.isAWalk("axe", "pott")
		}
	})
}
