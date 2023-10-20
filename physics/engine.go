package physics

import (
	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/metrics"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

type Simulation struct {
	Fluid    []*Unit
	Quadtree *Quadtree
	Metrics  *metrics.Metrics
	Config   *config.Config
	IsPause  bool
}

// La funzione che si occupa di creare una nuova simulazione
func NewSimulation(config *config.Config) (*Simulation, error) {

	config.UpdateWindowSettings()
	// Ad esempio, se desideri passare alcuni valori di config a NewFluid:

	bounds := rl.NewRectangle(0, 0, float32(config.GameWidth), float32(config.WindowHeight))
	quadtree := NewQuadtree(0, bounds)

	sim := &Simulation{
		Fluid:    make([]*Unit, 0, config.ParticleNumber),
		Quadtree: quadtree,
		Metrics:  &metrics.Metrics{},
		Config:   config,
		IsPause:  false,
	}

	return sim, nil
}

// La funzione che si occupa di resettare il fluido
func (s *Simulation) Reset() {
	s.Fluid = []*Unit{}
}

func (s *Simulation) NewFluidAtPosition(position rl.Vector2) {
	s.Fluid = append(s.Fluid, *newUnitsAtPosition(position, s.Config.GameWidth, s.Config.WindowHeight, s.Config.ParticleNumber, s.Config.ParticleRadius, s.Config.ParticleMass, s.Config.ParticleInitialSpacing, s.Config.ScaleFactor, s.Config.ParticleElasticity)...)
}

func (s *Simulation) NewFluidWithVelocity(position rl.Vector2) {
	go spawnUnitsWithVelocity(&s.Fluid, position, s.Config.GameWidth, s.Config.WindowHeight, s.Config.ParticleNumber, s.Config.ParticleRadius, s.Config.ParticleMass, s.Config.ParticleInitialSpacing, s.Config.ScaleFactor, s.Config.ParticleElasticity)
}

func (s *Simulation) Update() error {
	s.Metrics.Update()

	if s.Config.UseExperimentalQuadtree {
		return s.UpdateWithQuadtree()
	} else {
		return s.UpdateWithVerletIntegration()
	}

}

func (s *Simulation) findUnitByID(id uuid.UUID) *Unit {
	for _, unit := range s.Fluid {
		if unit.Id == id {
			return unit
		}
	}
	return nil
}
