package rotor

import (
	"sync"
	"time"
)

// SpeedRecord is one rotor speed reading.
type SpeedRecord struct {
	UnitID string    `json:"unit_id"`
	Source string    `json:"source"`
	Value  float64   `json:"value"`
	At     time.Time `json:"at"`
	Seq    uint64    `json:"seq"`
}

// SpeedHistory is a bounded ring of speed readings.
type SpeedHistory struct {
	capacity int
	mu       sync.RWMutex
	records  []SpeedRecord
	nextSeq  uint64
}

// NewSpeedHistory creates a history holding at most capacity records.
func NewSpeedHistory(capacity int) *SpeedHistory {
	return &SpeedHistory{capacity: capacity}
}

// Add appends a speed record, evicting the oldest when full.
func (h *SpeedHistory) Add(record SpeedRecord) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.nextSeq++
	record.Seq = h.nextSeq
	if len(h.records) == h.capacity {
		copy(h.records, h.records[1:])
		h.records[len(h.records)-1] = record
		return
	}
	h.records = append(h.records, record)
}

// Latest returns the most recent speed record.
func (h *SpeedHistory) Latest() (SpeedRecord, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.records) == 0 {
		return SpeedRecord{}, false
	}
	return h.records[len(h.records)-1], true
}

// Average returns the mean speed over the retained window.
func (h *SpeedHistory) Average() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.records) == 0 {
		return 0
	}
	var sum float64
	for _, record := range h.records {
		sum += record.Value
	}
	return sum / float64(len(h.records))
}

// Max returns the peak speed in the retained window.
func (h *SpeedHistory) Max() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var peak float64
	for _, record := range h.records {
		if record.Value > peak {
			peak = record.Value
		}
	}
	return peak
}

// Count returns the number of retained records.
func (h *SpeedHistory) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.records)
}
