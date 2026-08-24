package console

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"windctl/internal/blade"
	"windctl/internal/rotor"
	"windctl/internal/safety"
	"windctl/internal/strategy"
	"windctl/internal/turbine"
	"windctl/internal/wind"
)

// ListFarms returns every registered wind farm.
func (s *Server) ListFarms(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.components.Namespace.Farms())
}

// CreateFarm registers a new wind farm.
func (s *Server) CreateFarm(w http.ResponseWriter, r *http.Request) {
	var req FarmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	farm, err := s.components.Namespace.AddFarm(req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, farm)
}

// CreateUnit registers a turbine unit under a farm.
func (s *Server) CreateUnit(w http.ResponseWriter, r *http.Request) {
	var req UnitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	unit, err := s.components.Namespace.AddUnit(req.FarmID, req.Name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	t := turbine.NewTurbine(unit.ID, unit.Name, unit.FarmID)
	if err := s.components.Turbines.Register(t); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusCreated, turbine.Snapshot(t))
}

// GetFarm returns one wind farm by id.
func (s *Server) GetFarm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	farm, ok := s.components.Namespace.LookupFarm(id)
	if !ok {
		writeError(w, http.StatusNotFound, errNotFound(id))
		return
	}
	writeJSON(w, http.StatusOK, farm)
}

// GetFarmUnits returns the units registered under one farm.
func (s *Server) GetFarmUnits(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, ok := s.components.Namespace.LookupFarm(id); !ok {
		writeError(w, http.StatusNotFound, errNotFound(id))
		return
	}
	writeJSON(w, http.StatusOK, s.components.Namespace.UnitsByFarm(id))
}

// ListUnits returns the status of every turbine.
func (s *Server) ListUnits(w http.ResponseWriter, r *http.Request) {
	units := s.components.Turbines.List()
	statuses := make([]turbine.Status, 0, len(units))
	for _, t := range units {
		statuses = append(statuses, turbine.Snapshot(t))
	}
	writeJSON(w, http.StatusOK, statuses)
}

// GetUnit returns the status of one turbine.
func (s *Server) GetUnit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	t, _ := s.components.Turbines.Get(id)
	lastFault, hasFault := t.Faults().Last()
	unit, _ := s.components.Namespace.LookupUnit(id)
	yawCtl := s.components.Yaws[id]
	writeJSON(w, http.StatusOK, map[string]any{
		"status":        turbine.Snapshot(t),
		"recent_faults": t.Faults().Recent(5),
		"last_fault":    lastFault,
		"has_fault":     hasFault,
		"unit":          unit,
		"yaw": map[string]any{
			"azimuth": yawCtl.Azimuth(),
			"target":  yawCtl.Target(),
			"moving":  yawCtl.Moving(),
			"twist_tripped": yawCtl.Cable().Tripped(),
		},
	})
}

// GetSafety returns the safety chain state of a unit.
func (s *Server) GetSafety(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	chain := s.components.Chains[id]
	writeJSON(w, http.StatusOK, map[string]any{
		"unit_id":     id,
		"latched":     chain.Latched(),
		"recovered":   chain.Recovered(),
		"released":    chain.Released(),
		"trip_count":  chain.TripCount(),
		"reset_count": chain.ResetCount(),
		"last_trip":   chain.LastTrip().Format(time.RFC3339),
	})
}

// SetPitch applies a pitch command to a turbine.
func (s *Server) SetPitch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req PitchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	ctl := s.components.Blades[id]
	if err := blade.Execute(ctl, []blade.Step{{Command: blade.CmdTrackWind, Target: req.Angle}}); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, blade.Report(ctl))
}

// ReportFault records a new turbine fault.
func (s *Server) ReportFault(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req FaultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Code == "" {
		req.Code = "FAULT"
	}
	t, _ := s.components.Turbines.Get(id)
	fault := t.RecordFault(req.Code, req.Message)
	writeJSON(w, http.StatusCreated, fault)
}

// Feather feathers the blades of a turbine.
func (s *Server) Feather(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	ctl := s.components.Blades[id]
	if err := blade.Feather(ctl); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, blade.Report(ctl))
}

// LimitPower applies a power limit to a turbine.
func (s *Server) LimitPower(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req LimitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	ctl := s.components.Blades[id]
	if err := blade.LimitPower(ctl, req.MaxPower); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, blade.Report(ctl))
}

// YawFollow rotates the nacelle towards the wind.
func (s *Server) YawFollow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req YawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.components.Yaws[id].Follow(req.WindDir); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, MessageResponse{Message: "yaw follow accepted"})
}

// Unwind runs the cable unwind of a turbine.
func (s *Server) Unwind(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err := s.components.Yaws[id].Unwind(); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, MessageResponse{Message: "unwind accepted"})
}

// Trip raises the safety chain of a turbine.
func (s *Server) Trip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req TripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Reason == "" {
		req.Reason = "manual"
	}
	if err := s.components.Chains[id].Trip(req.Reason); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, MessageResponse{Message: "safety chain tripped"})
}

// Release releases the safety chain of a turbine.
func (s *Server) Release(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err := s.components.Chains[id].Release(); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, MessageResponse{Message: "safety chain released"})
}

// Reset resets a tripped turbine after the chain released.
func (s *Server) Reset(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	t, _ := s.components.Turbines.Get(id)
	if err := safety.DoReset(s.components.Chains[id], t); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, turbine.Snapshot(t))
}

// Stop runs the emergency stop sequence of a turbine.
func (s *Server) Stop(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req TripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	executor := stopExecutor{
		ctl:   s.components.Blades[id],
		brake: s.components.Rotors[id].Brake(),
	}
	if err := safety.RunStop(executor, s.components.Chains[id], req.Reason); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, MessageResponse{Message: "emergency stop accepted"})
}

// SwitchSpeedSource switches the active rotor speed sensor.
func (s *Server) SwitchSpeedSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err := s.components.Rotors[id].Switch(); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, MessageResponse{Message: "speed source switched"})
}

// SetOverspeedLimit tunes the overspeed threshold of a unit.
func (s *Server) SetOverspeedLimit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req OverspeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	detector := s.components.Detectors[id]
	detector.SetLimit(req.Limit)
	writeJSON(w, http.StatusOK, map[string]any{"unit_id": id, "limit": detector.Limit()})
}

// SetTwistThreshold tunes the cable twist threshold of a unit.
func (s *Server) SetTwistThreshold(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req ThresholdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	monitor := s.components.Cables[id]
	monitor.SetThreshold(req.Threshold)
	writeJSON(w, http.StatusOK, map[string]any{"unit_id": id, "threshold": monitor.Threshold()})
}

// GetStrategy returns the active strategy and its version history.
func (s *Server) GetStrategy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	sc := s.components.Strategies[id]
	writeJSON(w, http.StatusOK, map[string]any{
		"active":   sc.Active(),
		"versions": sc.Versions(),
		"durable":  sc.Sink().HasConfig(sc.Active().Name),
	})
}

// SwitchStrategy activates a new strategy after its config is durable.
func (s *Server) SwitchStrategy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	var req StrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	cfg := strategy.Config{Name: req.Name, Mode: req.Mode, MaxPower: req.MaxPower}
	if err := strategy.SwitchActive(s.components.Strategies[id], cfg); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, s.components.Strategies[id].Active())
}

// RollbackStrategy returns to the previous strategy version.
func (s *Server) RollbackStrategy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if err := strategy.Rollback(s.components.Strategies[id]); err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusOK, s.components.Strategies[id].Active())
}

// GetTwist returns the cable twist state of a unit.
func (s *Server) GetTwist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	monitor := s.components.Cables[id]
	writeJSON(w, http.StatusOK, TwistResponse{
		UnitID:      monitor.UnitID(),
		Accumulated: monitor.Accumulated(),
		Baseline:    monitor.Baseline(),
		Twist:       monitor.Twist(),
		Threshold:   monitor.Threshold(),
		Tripped:     monitor.Tripped(),
		BufferCount: monitor.Buffer().Count(),
		Generation:  monitor.Buffer().Generation(),
		Samples:     monitor.Buffer().Samples(),
	})
}

// GetFeedback returns the blade pitch feedback of a unit.
func (s *Server) GetFeedback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, blade.Report(s.components.Blades[id]))
}

// GetRotor returns the rotor speed state of a unit.
func (s *Server) GetRotor(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	rot := s.components.Rotors[id]
	writeJSON(w, http.StatusOK, map[string]any{
		"unit_id":     id,
		"source":      rot.SourceName(),
		"speed":       rot.Speed(),
		"brake":       rot.Brake().Applied(),
		"feathered":   rot.Brake().Feathered(),
		"brake_seq":   rot.Brake().Sequence(),
		"history":     rotorStats(rot),
	})
}

func rotorStats(rot *rotor.Rotor) map[string]any {
	latest, ok := rot.SpeedHistory().Latest()
	return map[string]any{
		"latest":  latest.Value,
		"source":  latest.Source,
		"has":     ok,
		"average": rot.SpeedHistory().Average(),
		"peak":    rot.SpeedHistory().Max(),
		"count":   rot.SpeedHistory().Count(),
	}
}

// ListAudit returns recent audit events.
func (s *Server) ListAudit(w http.ResponseWriter, r *http.Request) {
	req := AuditRequest{Kind: r.URL.Query().Get("kind"), Limit: 50}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			req.Limit = parsed
		}
	}
	entries, err := s.components.Auditor.Recent(req.Kind, req.Limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

// GetAcks returns the durable acknowledgements of a kind.
func (s *Server) GetAcks(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")
	if kind == "" {
		kind = "safety"
	}
	acks, err := s.components.Store.ReadAcks(kind)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, acks)
}

// PruneAudit removes the audit log after it is archived by an operator.
func (s *Server) PruneAudit(w http.ResponseWriter, r *http.Request) {
	if err := s.components.Store.Remove("audit/control.log"); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, MessageResponse{Message: "audit log pruned"})
}

// GetQuota returns the remaining sampling budget of a unit.
func (s *Server) GetQuota(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"unit_id":   id,
		"remaining": s.components.Quota.Remaining(id),
	})
}

// GetState returns the persisted plant state snapshot.
func (s *Server) GetState(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.components.Store.Recover()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

// WindStats returns the wind statistics of a unit.
func (s *Server) WindStats(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := s.unit(id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	stats := wind.ComputeStats(s.components.Winds[id].History())
	writeJSON(w, http.StatusOK, map[string]any{"unit_id": id, "stats": stats})
}

func errNotFound(id string) error {
	return &notFoundError{id: id}
}

type notFoundError struct {
	id string
}

func (e *notFoundError) Error() string {
	return "resource " + e.id + " not found"
}

// stopExecutor adapts the blade and brake components to the emergency stop
// sequence. Feathering marks the brake interlock before the blades move.
type stopExecutor struct {
	ctl   *blade.PitchController
	brake *rotor.Brake
}

func (s stopExecutor) Feather() error {
	s.brake.SetFeathered(true)
	return blade.Feather(s.ctl)
}

func (s stopExecutor) Brake() error {
	return s.brake.Engage()
}
