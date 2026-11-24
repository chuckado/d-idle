//go:build race

package nav

// raceEnabled: the race detector slows code five to ten times, so wall-clock ceilings do not apply under it (docs/process/timing-tests.md).
const raceEnabled = true
