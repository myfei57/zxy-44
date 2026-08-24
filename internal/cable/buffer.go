package cable

import (
	"sync"
	"time"
)

// Sample is one buffered twist angle with the unwind generation it belongs to.
type Sample struct {
	Angle      float64   `json:"angle"`
	At         time.Time `json:"at"`
	Generation uint64    `json:"generation"`
}

// Buffer keeps the recent twist samples of one turbine. Every unwind starts a
// new generation and drops the samples of the previous generation so stale
// angles are never replayed.
type Buffer struct {
	mu         sync.Mutex
	capacity   int
	samples    []Sample
	generation uint64
}

// NewBuffer creates a buffer holding at most capacity samples.
func NewBuffer(capacity int) *Buffer {
	return &Buffer{capacity: capacity}
}

// Add appends one angle sample to the buffer.
func (b *Buffer) Add(angle float64, at time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.samples = append(b.samples, Sample{
		Angle:      angle,
		At:         at,
		Generation: b.generation,
	})
	if len(b.samples) > b.capacity {
		copy(b.samples, b.samples[len(b.samples)-b.capacity:])
		b.samples = b.samples[:b.capacity]
	}
}

// DropBeforeUnwind starts a new unwind generation and removes every sample
// recorded before it.
func (b *Buffer) DropBeforeUnwind() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.generation++
}

// Samples returns a copy of every retained sample.
func (b *Buffer) Samples() []Sample {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]Sample, len(b.samples))
	copy(out, b.samples)
	return out
}

// AfterGeneration returns only the samples of the current unwind generation.
func (b *Buffer) AfterGeneration() []Sample {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []Sample
	for _, sample := range b.samples {
		if sample.Generation == b.generation {
			out = append(out, sample)
		}
	}
	return out
}

// Generation returns the current unwind generation number.
func (b *Buffer) Generation() uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.generation
}

// Count returns the number of retained samples.
func (b *Buffer) Count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.samples)
}
