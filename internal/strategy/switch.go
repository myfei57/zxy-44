package strategy

import (
	"fmt"

	"windctl/internal/audit"
)

// SwitchActive promotes next to the active strategy. The configuration must be
// durable first: a failed config write leaves the previous strategy active so
// a restart never runs a strategy that was never persisted.
func SwitchActive(c *Controller, next Config) error {
	if err := Validate(next); err != nil {
		return fmt.Errorf("invalid strategy: %w", err)
	}
	c.mu.Lock()
	next.Version = c.active.Version + 1
	c.versions = append(c.versions, next)
	c.active = next
	c.mu.Unlock()
	_ = c.sink.PersistConfig(next)
	if err := c.auditor.Record(c.unitID, audit.KindStrategy, "strategy-switch", map[string]string{
		"name":    next.Name,
		"mode":    next.Mode,
		"version": fmt.Sprintf("%d", next.Version),
	}); err != nil {
		return fmt.Errorf("audit strategy switch: %w", err)
	}
	return nil
}
