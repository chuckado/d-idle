package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

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
	return func(code string, quality int, identified, ethereal bool, props string) watch.VaultTip {
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
		if ic, ok := cat.(iconCatalog); ok {
			t.Art, _ = ic.InvFile(i, data.InvPlain, 0)
		}
		if !b.Quest {
			f := db.PriceFactsOf(b)
			f.Identified, f.Ethereal, f.Low = identified, ethereal, quality == 1
			t.Sell = b.SellPrice(f, 0)
		}
		t.Props = statLines(db, props)
		return t
	}
}

// statLines words d2d's "stat:param:value;..." list (item properties of an identified item) with the game's own itemstatcost
// text. S: descfunc 1 to 5 and 12 follow the standard table's meaning (plus value, percent value, plain value, value*100/128
// percent); every other function is left out rather than worded wrongly (U: the rest of the table is not covered yet).
// Lines come in descpriority order, highest first, as the game lists them.
func statLines(db *data.DB, props string) []string {
	type line struct {
		prio int
		text string
	}
	var out []line
	byID := map[int]*data.StatCost{}
	for _, s := range db.StatCosts {
		byID[s.ID] = s
	}
	for _, p := range strings.Split(props, ";") {
		f := strings.Split(p, ":")
		if len(f) != 3 {
			continue
		}
		id, e1 := strconv.Atoi(f[0])
		val, e2 := strconv.Atoi(f[2])
		s := byID[id]
		if e1 != nil || e2 != nil || s == nil || s.DescFunc == 0 {
			continue
		}
		val = s.FromMemory(val)
		key := s.DescStrPos
		if val < 0 && s.DescStrNeg != "" {
			key = s.DescStrNeg
		}
		str, ok := db.Str(key)
		if !ok || str == "" {
			continue
		}
		var v string
		switch s.DescFunc {
		case 1, 12:
			v = signed(val)
		case 2:
			v = strconv.Itoa(val) + "%"
		case 3:
			v = strconv.Itoa(val)
		case 4:
			v = signed(val) + "%"
		case 5:
			v = strconv.Itoa(val*100/128) + "%"
		default:
			continue
		}
		t := v + " " + str
		if s.DescVal == 2 {
			t = str + " " + v
		}
		out = append(out, line{s.DescPriority, t})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].prio > out[j].prio })
	lines := make([]string, len(out))
	for i, l := range out {
		lines[i] = l.text
	}
	return lines
}

func signed(v int) string {
	if v >= 0 {
		return fmt.Sprintf("+%d", v)
	}
	return strconv.Itoa(v)
}
