package console

import "windctl/internal/cable"

// dto.go holds the request and response payloads of the control console.

// FarmRequest registers a new wind farm.
type FarmRequest struct {
	Name string `json:"name"`
}

// UnitRequest registers a new turbine unit under a farm.
type UnitRequest struct {
	FarmID string `json:"farm_id"`
	Name   string `json:"name"`
}

// PitchRequest commands a blade pitch angle.
type PitchRequest struct {
	Angle float64 `json:"angle"`
}

// LimitRequest commands a power limit.
type LimitRequest struct {
	MaxPower float64 `json:"max_power"`
}

// YawRequest commands a yaw follow with a wind direction.
type YawRequest struct {
	WindDir float64 `json:"wind_dir"`
}

// TripRequest raises the safety chain.
type TripRequest struct {
	Reason string `json:"reason"`
}

// StrategyRequest switches the active strategy.
type StrategyRequest struct {
	Name     string  `json:"name"`
	Mode     string  `json:"mode"`
	MaxPower float64 `json:"max_power"`
}

// AuditRequest filters the audit log.
type AuditRequest struct {
	Kind  string `json:"kind"`
	Limit int    `json:"limit"`
}

// FaultRequest reports a new turbine fault.
type FaultRequest struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OverspeedRequest tunes the overspeed limit of a unit.
type OverspeedRequest struct {
	Limit float64 `json:"limit"`
}

// ThresholdRequest tunes the cable twist threshold of a unit.
type ThresholdRequest struct {
	Threshold float64 `json:"threshold"`
}

// ErrorResponse carries a single error message.
type ErrorResponse struct {
	Error string `json:"error"`
}

// MessageResponse carries a short operation result.
type MessageResponse struct {
	Message string `json:"message"`
}

// TwistResponse reports the cable twist state of a unit.
type TwistResponse struct {
	UnitID       string  `json:"unit_id"`
	Accumulated  float64 `json:"accumulated"`
	Baseline     float64 `json:"baseline"`
	Twist        float64 `json:"twist"`
	Threshold    float64 `json:"threshold"`
	Tripped      bool    `json:"tripped"`
	BufferCount  int     `json:"buffer_count"`
	Generation   uint64  `json:"generation"`
	Samples      []cable.Sample `json:"samples"`
}
