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
	Angle     float64
	Direction rl.Vector3
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

	if rl.IsMouseButtonPressed(rl.MouseRightButton) {
		println("Camera Position:", s.Camera.Position.X, s.Camera.Position.Y, s.Camera.Position.Z)
		println("Camera Target:", s.Camera.Target.X, s.Camera.Target.Y, s.Camera.Target.Z)

		// Calcolo del delta del mouse
		mouseDelta := rl.GetMouseDelta()

		// Aggiornamento dell'angolo di rotazione basato sul movimento orizzontale del mouse
		s.Angle += float64(mouseDelta.X) * 0.01

		// Calcolo della nuova direzione della camera
		cameraDirection := rl.Vector3{
			X: float32(math.Sin(s.Angle)), // Calcola la componente X
			Y: 0,                          // Mantiene la camera all'altezza corrente
			Z: float32(math.Cos(s.Angle)), // Calcola la componente Z
		}

		// Aggiornamento del target della camera basato sulla nuova direzione
		s.Camera.Target = rl.Vector3Add(s.Camera.Position, cameraDirection)

		// Assicurati di aggiornare la camera in ogni frame
		rl.UpdateCamera(&s.Camera, rl.CameraMode(rl.CameraPerspective))
	}

	s.Direction = rl.Vector3{
		X: float32(math.Sin(s.Angle)),
		Z: float32(-math.Cos(s.Angle)),
	}
	// Controllo della posizione della camera con WASD
	if rl.IsKeyDown(rl.KeyW) {
		s.Camera.Position = rl.Vector3Add(s.Camera.Position, rl.Vector3Scale(s.Direction, 0.05))
		s.Camera.Target = rl.Vector3Add(s.Camera.Target, rl.Vector3Scale(s.Direction, 0.05))
	}
	if rl.IsKeyDown(rl.KeyS) {
		s.Camera.Position = rl.Vector3Add(s.Camera.Position, rl.Vector3Scale(s.Direction, -0.05))
		s.Camera.Target = rl.Vector3Add(s.Camera.Target, rl.Vector3Scale(s.Direction, -0.05))
	}
	// Aggiunta di movimento laterale (strafe)
	right := rl.Vector3Normalize(rl.Vector3CrossProduct(s.Direction, s.Camera.Up))
	if rl.IsKeyDown(rl.KeyA) {
		s.Camera.Position = rl.Vector3Add(s.Camera.Position, rl.Vector3Scale(right, -0.05))
		s.Camera.Target = rl.Vector3Add(s.Camera.Target, rl.Vector3Scale(right, -0.05))
	}
	if rl.IsKeyDown(rl.KeyD) {
		s.Camera.Position = rl.Vector3Add(s.Camera.Position, rl.Vector3Scale(right, 0.05))
		s.Camera.Target = rl.Vector3Add(s.Camera.Target, rl.Vector3Scale(right, 0.05))
	}
	s.IsInputBeingHandled = false
}
