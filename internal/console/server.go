package console

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"

	"windctl/internal/audit"
	"windctl/internal/blade"
	"windctl/internal/cable"
	"windctl/internal/ns"
	"windctl/internal/quota"
	"windctl/internal/rotor"
	"windctl/internal/safety"
	"windctl/internal/store"
	"windctl/internal/strategy"
	"windctl/internal/turbine"
	"windctl/internal/wind"
	"windctl/internal/yaw"
)

// Components bundles every control component the console exposes.
type Components struct {
	Store      *store.Store
	Auditor    *audit.Auditor
	Namespace  *ns.Registry
	Turbines   *turbine.Registry
	Quota      *quota.Quota
	Blades     map[string]*blade.PitchController
	Yaws       map[string]*yaw.YawController
	Cables     map[string]*cable.Monitor
	Chains     map[string]*safety.Chain
	Rotors     map[string]*rotor.Rotor
	Detectors  map[string]*rotor.OverspeedDetector
	Strategies map[string]*strategy.Controller
	Winds      map[string]*wind.Sampler
}

// Server is the HTTP control console.
type Server struct {
	router     http.Handler
	components Components
	mu         sync.RWMutex
}

// NewServer builds the console with all routes wired.
func NewServer(components Components) *Server {
	s := &Server{components: components}
	router := chi.NewRouter()
	router.Use(RequestLog, Recover)
	router.Get("/healthz", s.HealthHandler)
	router.Get("/api/farms", s.ListFarms)
	router.Get("/api/farms/{id}", s.GetFarm)
	router.Get("/api/farms/{id}/units", s.GetFarmUnits)
	router.Post("/api/farms", s.CreateFarm)
	router.Post("/api/units", s.CreateUnit)
	router.Get("/api/units", s.ListUnits)
	router.Get("/api/units/{id}", s.GetUnit)
	router.Get("/api/units/{id}/safety", s.GetSafety)
	router.Post("/api/units/{id}/pitch", s.SetPitch)
	router.Post("/api/units/{id}/feather", s.Feather)
	router.Post("/api/units/{id}/limit", s.LimitPower)
	router.Post("/api/units/{id}/fault", s.ReportFault)
	router.Post("/api/units/{id}/yaw", s.YawFollow)
	router.Post("/api/units/{id}/unwind", s.Unwind)
	router.Post("/api/units/{id}/trip", s.Trip)
	router.Post("/api/units/{id}/release", s.Release)
	router.Post("/api/units/{id}/reset", s.Reset)
	router.Post("/api/units/{id}/stop", s.Stop)
	router.Post("/api/units/{id}/speed-source/switch", s.SwitchSpeedSource)
	router.Post("/api/units/{id}/overspeed-limit", s.SetOverspeedLimit)
	router.Post("/api/units/{id}/twist-threshold", s.SetTwistThreshold)
	router.Get("/api/units/{id}/strategy", s.GetStrategy)
	router.Post("/api/units/{id}/strategy/switch", s.SwitchStrategy)
	router.Post("/api/units/{id}/strategy/rollback", s.RollbackStrategy)
	router.Get("/api/units/{id}/twist", s.GetTwist)
	router.Get("/api/units/{id}/feedback", s.GetFeedback)
	router.Get("/api/units/{id}/rotor", s.GetRotor)
	router.Get("/api/units/{id}/wind", s.WindStats)
	router.Get("/api/audit", s.ListAudit)
	router.Post("/api/audit/prune", s.PruneAudit)
	router.Get("/api/acks", s.GetAcks)
	router.Get("/api/quota/{id}", s.GetQuota)
	router.Get("/api/state", s.GetState)
	s.router = router
	return s
}

// Router returns the HTTP handler of the console.
func (s *Server) Router() http.Handler {
	return s.router
}

// ListenAndServe starts the console on addr.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) unit(id string) error {
	if _, ok := s.components.Turbines.Get(id); !ok {
		return fmt.Errorf("turbine %s not found", id)
	}
	return nil
}
