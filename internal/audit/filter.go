package audit

// Filter returns the entries that match the optional kind and unit filters.
// An empty kind or unit matches everything.
func Filter(entries []Entry, kind, unitID string) []Entry {
	out := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		if kind != "" && entry.Kind != kind {
			continue
		}
		if unitID != "" && entry.UnitID != unitID {
			continue
		}
		out = append(out, entry)
	}
	return out
}
