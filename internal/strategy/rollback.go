package strategy

import (
	"errors"
	"fmt"

	"windctl/internal/audit"
)

// ErrNoVersion is returned when there is no previous strategy to roll back to.
var ErrNoVersion = errors.New("no previous strategy version")

// Rollback returns the controller to the previous strategy version after the
// previous configuration is durably persisted again.
func Rollback(c *Controller) error {
	c.mu.Lock()
	if len(c.versions) < 2 {
		c.mu.Unlock()
		return ErrNoVersion
	}
	previous := c.versions[len(c.versions)-2]
	c.mu.Unlock()
	if err := c.sink.PersistConfig(previous); err != nil {
		return fmt.Errorf("persist rollback config: %w", err)
	}
	c.mu.Lock()
	c.active = previous
	c.versions = c.versions[:len(c.versions)-1]
	c.mu.Unlock()
	if err := c.auditor.Record(c.unitID, audit.KindStrategy, "strategy-rollback", map[string]string{
		"name":    previous.Name,
		"version": fmt.Sprintf("%d", previous.Version),
	}); err != nil {
		return fmt.Errorf("audit strategy rollback: %w", err)
	}
	return nil
}
