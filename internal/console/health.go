package console

import (
	"encoding/json"
	"net/http"
	"time"
)

// healthResponse is the liveness payload of the control console.
type healthResponse struct {
	Status         string `json:"status"`
	Time           string `json:"time"`
	DataDir        string `json:"data_dir"`
	StatePath      string `json:"state_path"`
	StatePersisted bool   `json:"state_persisted"`
	Units          int    `json:"units"`
	Farms          int    `json:"farms"`
	NsUnits        int    `json:"ns_units"`
}

// HealthHandler reports whether the console and its store are usable.
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:         "ok",
		Time:           time.Now().UTC().Format(time.RFC3339),
		DataDir:        s.components.Store.Root(),
		StatePath:      s.components.Store.Path("state/plant.json"),
		StatePersisted: s.components.Store.Exists("state/plant.json"),
		Units:          s.components.Turbines.Count(),
		Farms:          s.components.Namespace.FarmCount(),
		NsUnits:        s.components.Namespace.UnitCount(),
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, ErrorResponse{Error: err.Error()})
}
