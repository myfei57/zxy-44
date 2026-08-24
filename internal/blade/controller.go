package blade

import "sync"

// PitchController tracks the blade pitch state of one turbine and keeps a
// bounded journal of the commands that were applied.
type PitchController struct {
	unitID  string
	mu      sync.Mutex
	angle   float64
	target  float64
	seq     uint64
	maxStep float64
	journal []Command
}

// NewPitchController creates a controller whose blades start at initial.
func NewPitchController(unitID string, initial float64) *PitchController {
	return &PitchController{
		unitID:  unitID,
		angle:   initial,
		target:  initial,
		maxStep: 6,
	}
}

// Current returns the current blade pitch angle.
func (p *PitchController) Current() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.angle
}

// Target returns the latest commanded pitch angle.
func (p *PitchController) Target() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.target
}

// Sequence returns how many pitch transitions were applied.
func (p *PitchController) Sequence() uint64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.seq
}

// MaxStep returns the maximum pitch change per transition.
func (p *PitchController) MaxStep() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.maxStep
}

// SetMaxStep changes the maximum pitch change per transition.
func (p *PitchController) SetMaxStep(step float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.maxStep = step
}

// Journal returns a copy of the applied command journal.
func (p *PitchController) Journal() []Command {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]Command, len(p.journal))
	copy(out, p.journal)
	return out
}

func (p *PitchController) record(cmd Command) {
	const journalLimit = 128
	if len(p.journal) == journalLimit {
		copy(p.journal, p.journal[1:])
		p.journal[len(p.journal)-1] = cmd
		return
	}
	p.journal = append(p.journal, cmd)
}
