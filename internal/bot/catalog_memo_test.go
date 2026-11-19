package bot

import (
	"fmt"
	"testing"

	"d2lod-bot/internal/data"
)

func memoCatalog() (dataCatalog, *data.DB) {
	db := &data.DB{}
	for i := 0; i < 600; i++ {
		typ := []string{"weap", "hpot", "mpot", "rpot", "tors", "ring", "rune", ""}[i%8]
		db.Items = append(db.Items, &data.ItemBase{Name: fmt.Sprint("item", i), Code: fmt.Sprintf("c%03d", i), Type: typ, Width: 1 + i%2, Height: 1 + i%3, Cost: i * 7})
	}
	return NewDataCatalog(db).(dataCatalog), db
}

func TestCatalogItemMemoEqualsUncached(t *testing.T) {
	c, db := memoCatalog()
	for pass := 0; pass < 2; pass++ { // the first pass fills the memo, the second reads it
		for class := range db.Items {
			got, ok := c.Item(class)
			want := c.item(class)
			if !ok || got != want {
				t.Fatalf("pass %d class %d: %+v vs %+v", pass, class, got, want)
			}
		}
	}
	if _, ok := c.Item(-1); ok {
		t.Error("class -1")
	}
	if _, ok := c.Item(len(db.Items)); ok {
		t.Error("class past the table")
	}
}

func BenchmarkCatalogItem(b *testing.B) {
	c, _ := memoCatalog()
	b.Run("memo", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			c.Item(i % 600)
		}
	})
	b.Run("uncached", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			c.item(i % 600)
		}
	})
}
