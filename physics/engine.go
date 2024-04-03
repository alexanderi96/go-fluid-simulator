package physics

import (
	"math"

	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/metrics"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

const (
	CameraMovementMode ControlMode = iota
	UnitSpawnMode
)

type ControlMode int

type Simulation struct {
	Fluid         []*Unit
	ClusterMasses map[uuid.UUID]float32
	Metrics       *metrics.Metrics
	Config        *config.Config

	Octree *Octree

	IsPause              bool
	InitialMousePosition rl.Vector2
	CurrentMousePosition rl.Vector2
	MouseButtonPressed   bool
	IsInputBeingHandled  bool

	// variables added for the 3d branch
	Camera rl.Camera

	// Velocità di rotazione
	MovementSpeed float32

	WorldBoundray rl.BoundingBox
	CubeCenter    rl.Vector3

	ControlMode   ControlMode
	SpawnDistance float32
	SpawnPosition rl.Vector3
}

func NewSimulation(config *config.Config) (*Simulation, error) {
	config.UpdateWindowSettings()
	cubeCenter := rl.NewVector3(float32(config.GameX)/2, float32(config.GameY)/2, float32(config.GameZ)/2)

	sim := &Simulation{
		Fluid:   make([]*Unit, 0, config.UnitNumber),
		Metrics: &metrics.Metrics{},
		Config:  config,
		IsPause: false,

		WorldBoundray: rl.NewBoundingBox(rl.NewVector3(0, 0, 0), rl.NewVector3(float32(config.GameX), float32(config.GameY), float32(config.GameZ))),
		CubeCenter:    cubeCenter,

		ControlMode:   UnitSpawnMode,
		SpawnDistance: 0,
		SpawnPosition: cubeCenter,
	}

	sim.Octree = NewOctree(0, sim.WorldBoundray)

	sim.ResetCameraPosition()

	return sim, nil
}

func (s *Simulation) ResetCameraPosition() {
	fovy := float32(60)
	fovyRadians := fovy * (math.Pi / 180)
	d := float32((math.Sqrt(3) * math.Max(float64(s.Config.GameX), math.Max(float64(s.Config.GameY), float64(s.Config.GameZ)))) / (2 * math.Tan(float64(fovyRadians)/2)))

	s.Camera = rl.Camera{
		Position:   rl.NewVector3(s.Config.GameX/2, float32(d), float32(d)),
		Target:     s.CubeCenter,
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       fovy,
		Projection: rl.CameraPerspective,
	}

	s.SpawnDistance = rl.Vector3Distance(s.CubeCenter, s.Camera.Position)
}

func (s *Simulation) Update() error {
	s.Metrics.Update()

	if s.Config.UseExperimentalOctree {
		return s.UpdateWithOctrees()
	} else {
		return s.UpdateWithVerletIntegration()
	}

}

func (s *Simulation) HandleInput() {
	s.IsInputBeingHandled = true

	if rl.IsKeyPressed(rl.KeyR) {
		s.ResetCameraPosition()
		s.Octree.Clear()
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

		if s.IsSpawnInRange() {
			s.Fluid = append(s.Fluid, *newUnitsWithAcceleration(s.SpawnPosition, rl.Vector3{}, s.Config)...)
		}

	} else if rl.IsKeyPressed(rl.KeyM) {
		// Cambio modalità con il tasto M (esempio)
		if s.ControlMode == CameraMovementMode {
			s.ControlMode = UnitSpawnMode
		} else {
			s.ControlMode = CameraMovementMode
		}
	}

	switch s.ControlMode {
	case CameraMovementMode:
		s.UpdateCameraPosition()

	case UnitSpawnMode:
		s.UpdateSpawnPosition()
	}

	s.IsInputBeingHandled = false
}

func (s *Simulation) UpdateCameraPosition() error {
	rl.UpdateCamera(&s.Camera, rl.CameraFree)

	return nil
}

func (s *Simulation) UpdateSpawnPosition() {
	mouseRay := rl.GetMouseRay(rl.GetMousePosition(), s.Camera)

	// Calcola la distanza basata sulla rotazione della rotella del mouse
	s.SpawnDistance += rl.GetMouseWheelMove() // Adatta questa formula secondo le tue necessità

	// Calcola la posizione del segnalino di anteprima lungo il raggio
	s.SpawnPosition = rl.Vector3Add(mouseRay.Position, rl.Vector3Scale(mouseRay.Direction, s.SpawnDistance))

}

func (s *Simulation) IsSpawnInRange() bool {
	return s.SpawnPosition.X >= s.WorldBoundray.Min.X && s.SpawnPosition.X <= s.WorldBoundray.Max.X &&
		s.SpawnPosition.Y >= s.WorldBoundray.Min.Y && s.SpawnPosition.Y <= s.WorldBoundray.Max.Y && s.SpawnPosition.Z >= s.WorldBoundray.Min.Z && s.SpawnPosition.Z <= s.WorldBoundray.Max.Z
}
