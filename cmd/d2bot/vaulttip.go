package main

import (
	"d2lod-bot/internal/bot"
	"d2lod-bot/internal/data"
	"d2lod-bot/internal/status"
	"d2lod-bot/internal/watch"
)

// vaultTipFor builds the tooltip facts of a vault row from the game tables, the same facts the inventory page's items carry
// (itemTips.fillBase: base type and sell price; itemArt: the icon name). nil db: rows carry none and the tooltip shows the
// stored name, level and the Unidentified line only. The unique and set icon rows are not stored with a record, so those
// qualities draw the base icon (U; the file index would have to be recorded with the item).
func vaultTipFor(db *data.DB) watch.VaultTipFunc {
	if db == nil {
		return nil
	}
	cat := bot.NewDataCatalog(db)
	class := map[string]int{}
	for i, b := range db.Items {
		if b != nil && b.Code != "" {
			class[b.Code] = i
		}
	}
	return func(in watch.VaultTipIn) watch.VaultTip {
		code, quality, identified, ethereal, props := in.Code, in.Quality, in.Identified, in.Ethereal, in.Props
		i, ok := class[code]
		if !ok {
			return watch.VaultTip{}
		}
		b := db.Items[i]
		t := watch.VaultTip{Base: &status.ItemBase{Name: b.Display, MinAC: b.MinAC, MaxAC: b.MaxAC, Block: b.Block, MinDam: b.MinDam, MaxDam: b.MaxDam,
			TwoMinDam: b.TwoMinDam, TwoMaxDam: b.TwoMaxDam, TwoHandedDam: b.TwoHanded, ReqStr: b.ReqStr, ReqDex: b.ReqDex,
			ReqLvl: b.LevelReq, Speed: b.Speed, MaxSockets: b.Sockets}}
		if !b.NoDurability {
			t.Base.Dur = b.Durability
		}
		if m, ok := cat.Item(i); ok {
			t.Kind = m.GoalType
		}
		t.Art = recordArt(cat, i, quality, in.FileIndex, in.Pic)
		if t.Art != "" {
			t.Tint = in.Tint
		}
		if !b.Quest {
			f := db.PriceFactsOf(b)
			f.Identified, f.Ethereal, f.Low = identified, ethereal, quality == 1
			t.Sell = b.SellPrice(f, 0)
		}
		t.Props = db.StatLines(props)
		return t
	}
}

// recordArt is itemArt for a stored record: the same rule (unique and set rows by file index, the picture variant of a
// jewel, ring, amulet or charm), with the quality ids of the vault (5 set, 7 unique).
func recordArt(cat bot.Catalog, class, quality, fileIndex, pic int) string {
	ic, ok := cat.(iconCatalog)
	if !ok {
		return ""
	}
	q := data.InvPlain
	switch quality {
	case 7:
		q = data.InvUnique
	case 5:
		q = data.InvSet
	}
	if pc, ok := cat.(picCatalog); ok && pic > 0 {
		name, _ := pc.InvFilePic(class, q, fileIndex, pic-1)
		return name
	}
	name, _ := ic.InvFile(class, q, fileIndex)
	return name
}
