package questspec

import "fmt"

// World is a scripted fake world: atomic facts plus the hero's area.
type World struct {
	Area  string
	Facts map[string]bool
	// Effect overrides what executing a step does, to model a world that
	// disagrees with the chain. Default: set the first alternative of every
	// post, clear the step's consumes.
	Effect map[string]func(*World)
}

func NewWorld(c *Chain, area string) *World {
	w := &World{Area: area, Facts: map[string]bool{}, Effect: map[string]func(*World){}}
	for _, cl := range c.InitialClauses() {
		w.Facts[cl[0].String()] = true
	}
	return w
}

// Holds evaluates a clause against the world; in_area reads the position.
func (w *World) Holds(cl Clause) bool {
	for _, p := range cl {
		if p.Name == "in_area" {
			if w.Area == p.Args[0] {
				return true
			}
		} else if w.Facts[p.String()] {
			return true
		}
	}
	return false
}

// Run executes the chain from whatever the world shows. After `restartAfter`
// executed steps (negative: never) the runner is thrown away and rebuilt, the
// hero is back in `home`, as after a bot restart; the next step must again come
// from facts alone. It returns the executed step ids in order.
func Run(c *Chain, w *World, restartAfter int, home string) ([]string, error) {
	var done []string
	for guard := 0; guard < len(c.Steps)*3+3; guard++ {
		if restartAfter >= 0 && len(done) == restartAfter {
			restartAfter = -1
			w.Area = home // state of the old runner is gone; only w.Facts survives
		}
		i := Derive(c, w.Holds)
		if i < 0 {
			return done, nil
		}
		s := c.Steps[i]
		for _, pre := range s.PreClauses() {
			if len(pre) == 1 && pre[0].Name == "in_area" {
				w.Area = pre[0].Args[0] // the executor travels there
			} else if !w.Holds(pre) {
				return done, fmt.Errorf("step %s: pre %s does not hold", s.ID, pre)
			}
		}
		if s.Target.Kind == "area" && s.Act == "travel" {
			w.Area = s.Target.Name
		}
		if f := w.Effect[s.ID]; f != nil {
			f(w)
		} else {
			for _, cl := range s.PostClauses() {
				w.Facts[cl[0].String()] = true
			}
			for _, cl := range s.ConsumedPreds() {
				for _, p := range cl {
					delete(w.Facts, p.String())
				}
			}
		}
		for _, cl := range s.PostClauses() {
			if !w.Holds(cl) {
				return done, fmt.Errorf("step %s: post %s not observed after executing", s.ID, cl)
			}
		}
		done = append(done, s.ID)
	}
	return done, fmt.Errorf("no progress: ran %d steps without finishing", len(done))
}
