package main

import (
	"fmt"
	"net/http"

	"d2lod-bot/internal/data"
	"d2lod-bot/internal/skilltree"
	"d2lod-bot/internal/watch"
)

// fakeSkills serves the Skills page (/watch/skills) a Sorceress with three small trees, so the size of the board can be seen at
// every window size. Fake tables, no game data.
type fakeSkills struct{}

func (fakeSkills) View() skilltree.View {
	var sk []*data.Skill
	id := 1
	for page := 1; page <= 3; page++ {
		for row := 1; row <= 6; row++ {
			for col := 1; col <= 3; col++ {
				if (row+col)%2 == 0 && row != 1 {
					continue
				}
				s := &data.Skill{ID: id, Key: fmt.Sprintf("Skill %d", id), Name: fmt.Sprintf("Skill %d", id), Class: "sor", Page: page, Row: row, Column: col, ReqLevel: row * 6, MaxLevel: 20}
				sk = append(sk, s)
				id++
			}
		}
	}
	pts := map[int]skilltree.Points{1: {Base: 4}, 2: {Base: 1}, 3: {Base: 3, Bonus: 1}}
	return skilltree.Build(skilltree.Tables{Skills: sk, Text: func(string) string { return "" }},
		skilltree.Hero{ClassCode: "sor", ClassName: "Sorceress", Level: 18, Unspent: 2, Points: pts}, nil)
}

func mountSkills(m *http.ServeMux) {
	api := watch.NewSkillsAPI()
	api.Bind(fakeSkills{})
	h := api.Handler()
	for _, p := range watch.SkillsRoutes {
		m.Handle(p, h)
	}
}
