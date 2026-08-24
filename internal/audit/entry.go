package audit

import (
	"time"

	"github.com/google/uuid"
)

// Event kinds recorded by the control components.
const (
	KindPitch    = "pitch"
	KindYaw      = "yaw"
	KindSafety   = "safety"
	KindStrategy = "strategy"
	KindCable    = "cable"
	KindReset    = "reset"
	KindUnwind   = "unwind"
	KindBrake    = "brake"
	KindQuota    = "quota"
	KindSystem   = "system"
)

// Entry is one immutable audit event produced by a control component.
type Entry struct {
	ID      string            `json:"id"`
	UnitID  string            `json:"unit_id"`
	Kind    string            `json:"kind"`
	Message string            `json:"message"`
	At      time.Time         `json:"at"`
	Meta    map[string]string `json:"meta,omitempty"`
}

// NewEntry creates an audit event with a fresh unique identifier.
func NewEntry(unitID, kind, message string, meta map[string]string) Entry {
	return Entry{
		ID:      uuid.NewString(),
		UnitID:  unitID,
		Kind:    kind,
		Message: message,
		At:      time.Now().UTC(),
		Meta:    meta,
	}
}
