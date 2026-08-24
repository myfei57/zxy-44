package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"windctl/internal/audit"
	"windctl/internal/blade"
	"windctl/internal/cable"
	"windctl/internal/console"
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

// sinkAdapter bridges the file store to the strategy config sink.
type sinkAdapter struct {
	store *store.Store
}

func (a sinkAdapter) PersistConfig(cfg strategy.Config) error {
	return a.store.WriteConfig(cfg.Name, cfg.Payload)
}

func (a sinkAdapter) LoadConfig(name string) ([]byte, error) {
	return a.store.LoadConfig(name)
}

func (a sinkAdapter) HasConfig(name string) bool {
	return a.store.HasConfig(name)
}

type unitControllers struct {
	turbine  *turbine.Turbine
	blade    *blade.PitchController
	cable    *cable.Monitor
	yaw      *yaw.YawController
	rotor    *rotor.Rotor
	chain    *safety.Chain
	strategy *strategy.Controller
	powerLoop   *strategy.Loop
	protectLoop *strategy.Loop
	evaluator   *strategy.Evaluator
	wind     *wind.Sampler
	rotorMon *rotor.Monitor
	safety   *safety.Monitor
}

func main() {
	addr := flag.String("addr", ":8080", "console listen address")
	dataDir := flag.String("data", "./data", "control data directory")
	flag.Parse()

	st, err := store.New(*dataDir)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	auditor := audit.New(st)
	quotaManager, err := quota.New(quota.DefaultPolicy(), st)
	if err != nil {
		log.Fatalf("init quota: %v", err)
	}

	names := ns.NewRegistry()
	farm, err := names.AddFarm("East Wind Farm")
	if err != nil {
		log.Fatalf("create farm: %v", err)
	}
	turbines := turbine.NewRegistry()
	units := make([]*unitControllers, 0, 6)
	blades := make(map[string]*blade.PitchController)
	yaws := make(map[string]*yaw.YawController)
	cables := make(map[string]*cable.Monitor)
	chains := make(map[string]*safety.Chain)
	rotors := make(map[string]*rotor.Rotor)
	detectors := make(map[string]*rotor.OverspeedDetector)
	strategies := make(map[string]*strategy.Controller)
	winds := make(map[string]*wind.Sampler)

	unitNames := []string{"WTG-01", "WTG-02", "WTG-03", "WTG-04", "WTG-05", "WTG-06"}
	for _, name := range unitNames {
		unit, err := names.AddUnit(farm.ID, name)
		if err != nil {
			log.Fatalf("create unit %s: %v", name, err)
		}
		t := turbine.NewTurbine(unit.ID, unit.Name, unit.FarmID)
		if err := turbines.Register(t); err != nil {
			log.Fatalf("register unit: %v", err)
		}
		ctl := blade.NewPitchController(unit.ID, 0)
		ctl.SetMaxStep(8)
		cableMon := cable.NewMonitor(unit.ID, 150, st, auditor)
		yawCtl := yaw.NewYawController(unit.ID, 0, cableMon, auditor, st)
		primary := rotor.NewSpeedSource("primary", func() (float64, error) {
			return 9 + rand.Float64()*2, nil
		})
		backup := rotor.NewSpeedSource("backup", func() (float64, error) {
			return 9.5 + rand.Float64()*2, nil
		})
		rot := rotor.NewRotor(unit.ID, primary, backup)
		chain, err := safety.Restore(unit.ID, st, auditor)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Printf("restore chain %s: %v (using defaults)", unit.ID, err)
			}
			chain = safety.NewChain(unit.ID, st, auditor)
			if err := chain.PersistState(); err != nil {
				log.Printf("persist chain %s: %v", unit.ID, err)
			}
		}
		sc := strategy.NewController(unit.ID, sinkAdapter{store: st}, rot, ctl, auditor)
		windSampler := wind.NewSampler(unit.ID, func() (float64, error) {
			return 7 + rand.Float64()*4, nil
		}, quotaManager)

		detector := rotor.NewOverspeedDetector(rot, 18)
		rotorMon := rotor.NewMonitor(rot, detector, chain)
		evaluator := strategy.NewEvaluator(rot)
		powerLoop := strategy.NewLoop("power", sc, func() ([]blade.Step, error) {
			return strategy.Decide(sc, windSampler.Average(), false)
		})
		protectLoop := strategy.NewLoop("protection", sc, func() ([]blade.Step, error) {
			limit := strategy.LimitForMode(sc.Active().Mode)
			return strategy.Decide(sc, windSampler.Average(), evaluator.Check(limit))
		})
		safetyMon := safety.NewMonitor(chain, safety.Conditions{
			SpeedOK:      func() bool { return rot.Speed() < rotorMon.Detector().Limit() },
			VibrationOK: func() bool { return true },
		})

		uc := &unitControllers{
			turbine:  t,
			blade:    ctl,
			cable:    cableMon,
			yaw:      yawCtl,
			rotor:    rot,
			chain:    chain,
			strategy: sc,
			powerLoop:   powerLoop,
			protectLoop: protectLoop,
			evaluator:   evaluator,
			wind:     windSampler,
			rotorMon: rotorMon,
			safety:   safetyMon,
		}
		units = append(units, uc)
		blades[unit.ID] = ctl
		yaws[unit.ID] = yawCtl
		cables[unit.ID] = cableMon
		chains[unit.ID] = chain
		rotors[unit.ID] = rot
		detectors[unit.ID] = detector
		strategies[unit.ID] = sc
		winds[unit.ID] = windSampler
	}

	if err := auditor.Record("", audit.KindSystem, "control-start", map[string]string{
		"units": fmt.Sprintf("%d", len(units)),
		"farm":  farm.Name,
	}); err != nil {
		log.Printf("record startup audit: %v", err)
	}

	components := console.Components{
		Store:      st,
		Auditor:    auditor,
		Namespace:  names,
		Turbines:   turbines,
		Quota:      quotaManager,
		Blades:     blades,
		Yaws:       yaws,
		Cables:     cables,
		Chains:     chains,
		Rotors:     rotors,
		Detectors:  detectors,
		Strategies: strategies,
		Winds:      winds,
	}
	server := console.NewServer(components)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr:              *addr,
		Handler:           server.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("windctl console listening on %s (data: %s)", *addr, st.Root())
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("console server: %v", err)
		}
	}()

	runLoops(ctx, units, st, quotaManager)

	<-ctx.Done()
	log.Printf("shutting down windctl")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	for _, uc := range units {
		if err := uc.chain.PersistState(); err != nil {
			log.Printf("persist chain %s on shutdown: %v", uc.turbine.ID, err)
		}
	}
	if err := quotaManager.Persist(); err != nil {
		log.Printf("persist quota on shutdown: %v", err)
	}
	log.Printf("windctl stopped")
}

func runLoops(ctx context.Context, units []*unitControllers, st *store.Store, q *quota.Quota) {
	windTicker := time.NewTicker(500 * time.Millisecond)
	rotorTicker := time.NewTicker(300 * time.Millisecond)
	strategyTicker := time.NewTicker(200 * time.Millisecond)
	safetyTicker := time.NewTicker(500 * time.Millisecond)
	yawTicker := time.NewTicker(1 * time.Second)
	cableTicker := time.NewTicker(1 * time.Second)
	stateTicker := time.NewTicker(10 * time.Second)
	quotaTicker := time.NewTicker(60 * time.Second)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-windTicker.C:
				for _, uc := range units {
					if _, err := uc.wind.Sample(); err != nil {
						log.Printf("wind sample %s: %v", uc.turbine.ID, err)
					}
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-rotorTicker.C:
				for _, uc := range units {
					speed, err := uc.rotorMon.Tick()
					if err != nil {
						log.Printf("rotor tick %s: %v", uc.turbine.ID, err)
						continue
					}
					uc.turbine.SetSpeed(speed)
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-strategyTicker.C:
				for _, uc := range units {
					if err := uc.powerLoop.Tick(); err != nil {
						log.Printf("power loop %s: %v", uc.powerLoop.Name(), err)
					}
					if err := uc.protectLoop.Tick(); err != nil {
						log.Printf("protection loop %s: %v", uc.protectLoop.Name(), err)
					}
					uc.turbine.SetPitch(uc.blade.Current())
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-safetyTicker.C:
				for _, uc := range units {
					if err := uc.safety.Tick(); err != nil {
						log.Printf("safety tick %s: %v", uc.turbine.ID, err)
					}
					if uc.chain.Latched() {
						uc.turbine.Enter(turbine.StateTripped)
					} else if uc.turbine.State() == turbine.StateIdle {
						uc.turbine.Enter(turbine.StateRunning)
					}
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-yawTicker.C:
				for _, uc := range units {
					windDir := uc.wind.Average() * 10
					if err := uc.yaw.Follow(windDir); err != nil {
						if errors.Is(err, yaw.ErrCableTwist) {
							log.Printf("yaw %s blocked by cable twist", uc.turbine.ID)
						} else {
							log.Printf("yaw follow %s: %v", uc.turbine.ID, err)
						}
					}
					uc.turbine.SetAzimuth(uc.yaw.Azimuth())
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-cableTicker.C:
				for _, uc := range units {
					if _, err := uc.cable.Evaluate(); err != nil {
						log.Printf("cable evaluate %s: %v", uc.turbine.ID, err)
					}
					if _, _, err := uc.yaw.EvaluateTwist(); err != nil {
						log.Printf("twist evaluate %s: %v", uc.turbine.ID, err)
					}
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-stateTicker.C:
				snapshot := store.StateSnapshot{
					UnitStates:       make(map[string]string),
					ActiveStrategies: make(map[string]string),
					SpeedSources:     make(map[string]string),
					SavedAt:          time.Now().UTC(),
				}
				for _, uc := range units {
					snapshot.UnitStates[uc.turbine.ID] = uc.turbine.State().String()
					snapshot.ActiveStrategies[uc.turbine.ID] = uc.strategy.Active().Name
					snapshot.SpeedSources[uc.turbine.ID] = uc.rotor.SourceName()
				}
				if err := st.SaveState(snapshot); err != nil {
					log.Printf("save state: %v", err)
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-quotaTicker.C:
				if err := q.ResetCycle(); err != nil {
					log.Printf("quota reset: %v", err)
				}
			}
		}
	}()
}
