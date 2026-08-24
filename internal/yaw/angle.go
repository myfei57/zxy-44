package yaw

import (
	"fmt"

	"windctl/internal/audit"
	"windctl/internal/cable"
)

// EvaluateTwist derives the twist reading from the buffered angle samples of
// the current unwind generation and raises an alert on threshold exceed. The
// evaluation must ignore samples buffered before the last unwind.
func (y *YawController) EvaluateTwist() (float64, bool, error) {
	samples := y.cable.Buffer().Samples()
	if len(samples) == 0 {
		return 0, false, nil
	}
	peak := 0.0
	for _, sample := range samples {
		if sample.Angle > peak {
			peak = sample.Angle
		}
	}
	tripped := cable.ThresholdExceeded(peak, y.cable.Threshold())
	if tripped {
		if err := y.auditor.Record(y.unitID, audit.KindCable, "twist-evaluation-alert", map[string]string{
			"peak": fmt.Sprintf("%.1f", peak),
		}); err != nil {
			return peak, tripped, err
		}
	}
	return peak, tripped, nil
}
