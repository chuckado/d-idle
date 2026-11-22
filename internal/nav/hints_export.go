package nav

// RouteHints is the hand-kept outdoor links of Acts 1 to 5 (the act hint tables), for the exits probe that compares
// them with the engine's own link table (tools/probes/exitscmp). Read-only copy.
func RouteHints() [][2]int { return append([][2]int(nil), allHints()...) }
