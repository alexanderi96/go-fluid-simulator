package physics

import (
	"fmt"
	"math"

	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/metrics"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Simulation struct {
	Fluid                []*Unit
	Metrics              *metrics.Metrics
	Config               *config.Config
	IsPause              bool
	InitialMousePosition rl.Vector2
	CurrentMousePosition rl.Vector2
	MouseButtonPressed   bool
	IsInputBeingHandled  bool
}

func NewSimulation(config *config.Config) (*Simulation, error) {
	config.UpdateWindowSettings()

	sim := &Simulation{
		Fluid:   make([]*Unit, 0, config.UnitNumber),
		Metrics: &metrics.Metrics{},
		Config:  config,
		IsPause: false,
	}

	return sim, nil
}

func (s *Simulation) Update() error {
	s.Metrics.Update()

	if s.Config.UseExperimentalQuadtree {
		return fmt.Errorf("quadtree not implemented yet")
	} else {
		return s.UpdateWithVerletIntegration()
	}

}

func (s *Simulation) HandleInput() {
	s.IsInputBeingHandled = true
	if rl.IsKeyPressed(rl.KeyR) {
		s.Fluid = []*Unit{}
	} else if rl.IsKeyPressed(rl.KeySpace) {
		s.IsPause = !s.IsPause
	} else if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		s.InitialMousePosition = rl.GetMousePosition()

		for rl.IsMouseButtonDown(rl.MouseLeftButton) && s.InitialMousePosition.X > 0 && s.InitialMousePosition.X < float32(s.Config.ViewportX) &&
			s.InitialMousePosition.Y > 0 && s.InitialMousePosition.Y < float32(s.Config.ViewportY) {
			s.MouseButtonPressed = true

			s.CurrentMousePosition = rl.GetMousePosition()
		}

	} else if s.MouseButtonPressed && rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		s.MouseButtonPressed = false

		if s.InitialMousePosition.X > 0 && s.InitialMousePosition.X < float32(s.Config.ViewportX) &&
			s.InitialMousePosition.Y > 0 && s.InitialMousePosition.Y < float32(s.Config.ViewportY) {

			delta := rl.Vector2Subtract(s.CurrentMousePosition, s.InitialMousePosition)
			acceleration := rl.Vector2{}
			deltaLength := float32(math.Sqrt(float64(delta.X*delta.X + delta.Y*delta.Y)))
			if deltaLength != 0 {
				acceleration = rl.Vector2{
					X: delta.X * deltaLength,
					Y: delta.Y * deltaLength,
				}

			}
			s.Fluid = append(s.Fluid, *newUnitsWithAcceleration(s.InitialMousePosition, s.Config, acceleration)...)

		}
	} else if rl.IsMouseButtonPressed(rl.MouseRightButton) {
		if s.CurrentMousePosition.X > 0 && s.CurrentMousePosition.X < float32(s.Config.ViewportX) && s.CurrentMousePosition.Y > 0 && s.CurrentMousePosition.Y < float32(s.Config.WindowHeight) {
			go spawnUnitsWithVelocity(&s.Fluid, s.CurrentMousePosition, s.Config)
		}
	}
	s.IsInputBeingHandled = false
}
