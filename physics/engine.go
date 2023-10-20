package physics

import (
	"fmt"

	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/metrics"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Simulation struct {
	Fluid   []*Unit
	Metrics *metrics.Metrics
	Config  *config.Config
	IsPause bool
}

func NewSimulation(config *config.Config) (*Simulation, error) {
	config.UpdateWindowSettings()

	sim := &Simulation{
		Fluid:   make([]*Unit, 0, config.ParticleNumber),
		Metrics: &metrics.Metrics{},
		Config:  config,
		IsPause: false,
	}

	return sim, nil
}

func (s *Simulation) Reset() {
	s.Fluid = []*Unit{}
}

func (s *Simulation) NewFluidAtPosition(position rl.Vector2) {
	s.Fluid = append(s.Fluid, *newUnitsAtPosition(position, s.Config)...)
}

func (s *Simulation) NewFluidWithVelocity(position rl.Vector2) {
	go spawnUnitsWithVelocity(&s.Fluid, position, s.Config)
}

func (s *Simulation) Update() error {
	s.Metrics.Update()

	if s.Config.UseExperimentalQuadtree {
		return fmt.Errorf("quadtree not implemented yet")
	} else {
		return s.UpdateWithVerletIntegration()
	}

}
