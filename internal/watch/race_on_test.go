//go:build race

package watch

// raceEnabled: the race detector slows every call by five to ten times, so timing ceilings do not apply under it.
const raceEnabled = true
