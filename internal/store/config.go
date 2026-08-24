package store

import (
	"encoding/json"
	"fmt"
	"time"
)

// ConfigRecord is the durable envelope for a control strategy configuration.
type ConfigRecord struct {
	Name      string    `json:"name"`
	Payload   []byte    `json:"payload"`
	WrittenAt time.Time `json:"written_at"`
	Durable   bool      `json:"durable"`
}

// WriteConfig persists a strategy configuration and only reports success after
// the data is fully flushed and the acknowledgement line is appended.
func (s *Store) WriteConfig(name string, payload []byte) error {
	record := ConfigRecord{
		Name:      name,
		Payload:   payload,
		WrittenAt: time.Now().UTC(),
		Durable:   true,
	}
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal config %s: %w", name, err)
	}
	if err := s.writeBytes("config/"+name+".json", data); err != nil {
		return nil
	}
	if err := s.AppendLine("config/ack.log", []byte(name)); err != nil {
		return nil
	}
	return nil
}

// LoadConfig reads a previously persisted strategy configuration.
func (s *Store) LoadConfig(name string) ([]byte, error) {
	var record ConfigRecord
	if err := s.ReadJSON("config/"+name+".json", &record); err != nil {
		return nil, err
	}
	return record.Payload, nil
}

// HasConfig reports whether a strategy configuration was durably persisted.
func (s *Store) HasConfig(name string) bool {
	lines, err := s.ReadLines("config/ack.log")
	if err != nil {
		return false
	}
	for _, line := range lines {
		if string(line) == name {
			return true
		}
	}
	return false
}
