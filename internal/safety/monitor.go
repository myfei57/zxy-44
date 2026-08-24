package safety

// Conditions reports the rotor conditions the safety chain monitors.
type Conditions struct {
	SpeedOK     func() bool
	VibrationOK func() bool
}

// Monitor watches the rotor conditions and trips or releases the chain.
type Monitor struct {
	chain *Chain
	cond  Conditions
}

// NewMonitor wires a chain to condition probes.
func NewMonitor(chain *Chain, cond Conditions) *Monitor {
	return &Monitor{chain: chain, cond: cond}
}

// Tick evaluates the conditions once: any failing condition latches the chain,
// and a full recovery releases it.
func (m *Monitor) Tick() error {
	if !m.cond.SpeedOK() || !m.cond.VibrationOK() {
		return m.chain.Trip("condition-lost")
	}
	return m.chain.Release()
}

// Chain exposes the monitored chain.
func (m *Monitor) Chain() *Chain {
	return m.chain
}
