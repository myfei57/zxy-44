package store

import (
	"encoding/json"
	"fmt"
)

// AppendAudit serialises one audit event and appends it to the append-only
// control audit log. The entry is any JSON-marshalable value.
func (s *Store) AppendAudit(entry any) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal audit entry: %w", err)
	}
	return s.AppendLine("audit/control.log", data)
}

// ReadAuditLines returns every raw audit line in chronological order.
func (s *Store) ReadAuditLines() ([][]byte, error) {
	return s.ReadLines("audit/control.log")
}
