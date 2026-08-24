package wind

import "sync"

// History is a bounded ring of wind samples used for rolling averages.
type History struct {
	capacity int
	mu       sync.RWMutex
	samples  []Sample
	nextSeq  uint64
}

// NewHistory creates a history holding at most capacity samples.
func NewHistory(capacity int) *History {
	return &History{capacity: capacity}
}

// Add appends a sample, evicting the oldest one when the ring is full.
func (h *History) Add(sample Sample) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.nextSeq++
	sample.Seq = h.nextSeq
	if len(h.samples) == h.capacity {
		copy(h.samples, h.samples[1:])
		h.samples[len(h.samples)-1] = sample
		return
	}
	h.samples = append(h.samples, sample)
}

// Latest returns the most recent sample.
func (h *History) Latest() (Sample, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.samples) == 0 {
		return Sample{}, false
	}
	return h.samples[len(h.samples)-1], true
}

// Average returns the mean speed over the retained window.
func (h *History) Average() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.samples) == 0 {
		return 0
	}
	var sum float64
	for _, sample := range h.samples {
		sum += sample.Speed
	}
	return sum / float64(len(h.samples))
}

// Max returns the peak speed in the retained window.
func (h *History) Max() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var peak float64
	for _, sample := range h.samples {
		if sample.Speed > peak {
			peak = sample.Speed
		}
	}
	return peak
}

// Count returns the number of retained samples.
func (h *History) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.samples)
}
