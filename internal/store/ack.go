package store

import (
	"encoding/json"
	"fmt"
	"time"
)

// AckRecord is a durable acknowledgement for a critical control action such as
// an unwind or a safety reset.
type AckRecord struct {
	Kind string    `json:"kind"`
	Key  string    `json:"key"`
	At   time.Time `json:"at"`
}

// WriteAck durably records that a critical action completed.
func (s *Store) WriteAck(kind, key string) error {
	record := AckRecord{Kind: kind, Key: key, At: time.Now().UTC()}
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal ack: %w", err)
	}
	return s.AppendLine("acks/"+kind+".log", data)
}

// ReadAcks returns every acknowledgement of a given kind.
func (s *Store) ReadAcks(kind string) ([]AckRecord, error) {
	lines, err := s.ReadLines("acks/" + kind + ".log")
	if err != nil {
		return nil, err
	}
	out := make([]AckRecord, 0, len(lines))
	for _, line := range lines {
		var record AckRecord
		if err := json.Unmarshal(line, &record); err != nil {
			return nil, fmt.Errorf("parse ack line: %w", err)
		}
		out = append(out, record)
	}
	return out, nil
}
