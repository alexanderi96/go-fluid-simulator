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

	// variables added for the 3d branch

	Camera    rl.Camera

	// Variabili per tenere traccia degli angoli di rotazione (in radianti)
	Angle float32  // Rotazione sull'asse Y
	Pitch float32  // Rotazione sull'asse X
	Distance float32  // Distanza dalla camera al centro del cubo

	// Velocità di rotazione
	MovementSpeed float32

	CubeCenter rl.Vector3
}

func NewSimulation(config *config.Config) (*Simulation, error) {
	config.UpdateWindowSettings()

	sim := &Simulation{
		Fluid:   make([]*Unit, 0, config.UnitNumber),
		Metrics: &metrics.Metrics{},
		Config:  config,
		IsPause: false,

		Camera: rl.Camera{
			Target:   rl.NewVector3(float32(config.GameX)/2, float32(config.GameY)/2, float32(config.GameZ)/2),
			Up:       rl.NewVector3(0, 0, 1),
			Fovy: 45,
			Projection: rl.CameraPerspective,
		},

		Angle: 0,
		Pitch: 0,
		MovementSpeed: 0.01,

		CubeCenter: rl.NewVector3(float32(config.GameX)/2, float32(config.GameY)/2, float32(config.GameZ)/2),

	}

	fovyRadians := sim.Camera.Fovy * (math.Pi / 180)

	sim.Distance = float32((math.Sqrt(3) * math.Max(float64(config.GameX), math.Max(float64(config.GameY), float64(config.GameZ)))) / (2 * math.Tan(float64(fovyRadians) / 2)))
	sim.Camera.Position = rl.NewVector3(float32(config.GameX)/2, float32(sim.Distance), float32(config.GameZ)/2)

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
	}

	if rl.IsKeyPressed(rl.KeySpace) {
		s.IsPause = !s.IsPause
	}

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		s.InitialMousePosition = rl.GetMousePosition()

		for rl.IsMouseButtonDown(rl.MouseLeftButton) && s.InitialMousePosition.X > 0 && s.InitialMousePosition.X < float32(s.Config.ViewportX) &&
			s.InitialMousePosition.Y > 0 && s.InitialMousePosition.Y < float32(s.Config.ViewportY) {
			s.MouseButtonPressed = true

			s.CurrentMousePosition = rl.GetMousePosition()
		}

	}

	if s.MouseButtonPressed && rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		s.MouseButtonPressed = false

		if s.InitialMousePosition.X > 0 && s.InitialMousePosition.X < s.Config.GameX &&
			s.InitialMousePosition.Y > 0 && s.InitialMousePosition.Y < s.Config.GameY {

			delta := rl.Vector2Subtract(s.CurrentMousePosition, s.InitialMousePosition)
			acceleration := rl.Vector3{}
			deltaLength := float32(math.Sqrt(float64(delta.X*delta.X + delta.Y*delta.Y)))
			if deltaLength != 0 {
				acceleration = rl.Vector3{
					X: delta.X * deltaLength,
					Y: delta.Y * deltaLength,
					Z: 0,
				}
			}
			s.Fluid = append(s.Fluid, *newUnitsWithAcceleration(rl.Vector3{X: s.InitialMousePosition.X, Y: s.InitialMousePosition.Y, Z: 0}, s.Config, acceleration)...)

		}
	}

	s.IsInputBeingHandled = false
}

func (s *Simulation) UpdateCameraPosition() error {
	rl.UpdateCamera(&s.Camera, rl.CameraFree)

	return nil
}