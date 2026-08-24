package wind

// Stats summarises a wind history for the console dashboard.
type Stats struct {
	Mean  float64 `json:"mean"`
	Peak  float64 `json:"peak"`
	Count int     `json:"count"`
}

// ComputeStats derives mean, peak and count from a history.
func ComputeStats(h *History) Stats {
	return Stats{Mean: h.Average(), Peak: h.Max(), Count: h.Count()}
}
