package senses

// Objective is the place the hero is walking to (the goal of the walk in progress). It is the one thing the map shows
// beyond the walked ground: the hero has it (a quest object or boss read from the engine, a waypoint, an exit), the
// level outline around it stays fogged (docs/mapfog.md "Walked ground and the objective").
type Objective struct {
	Area  int
	X, Y  int
	Label string // short, optional ("Countess"); empty when the bot has no name for it
}

// SetObjective stores the hero's current objective; ok false clears it. Called from the nav goroutine at about 2 Hz.
func (p *Publisher) SetObjective(o Objective, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.obj, p.hasObj = o, ok
}

// ObjectiveOf returns the objective when it belongs to the area the hero is in now. It is not masked by the fog: the
// marker is the point of having it.
func (p *Publisher) ObjectiveOf() (Objective, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.have || !p.hasObj || p.obj.Area != p.areaNo {
		return Objective{}, false
	}
	return p.obj, true
}

// NewGame forgets the walked ground of every area: a new game lays its levels out again (the same area id is a
// different maze), so the blocks the hero walked in the last game must not reveal this one. The trail ring and the
// event history are time-keyed and stay. The walked ground lives in memory only: it is not kept over a bot restart (the
// map fogs again, which is the honest picture of a hero that has not been there in this run).
func (p *Publisher) NewGame() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.seen, p.visits, p.hasObj = nil, nil, false
	p.have = false // the next Observe rebuilds the map from the empty fog
	p.snap.Neighbours = nil
}
