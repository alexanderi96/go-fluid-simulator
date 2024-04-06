package physics

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/metrics"
	"github.com/alexanderi96/go-fluid-simulator/utils"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	CameraMovementMode ControlMode = iota
	UnitSpawnMode
)

type ControlMode int

type Simulation struct {
	Fluid   []*Unit
	Metrics *metrics.Metrics
	Config  *config.Config

	Octree *Octree

	IsPause              bool
	InitialMousePosition rl.Vector2
	CurrentMousePosition rl.Vector2
	MouseButtonPressed   bool
	IsInputBeingHandled  bool

	// variables added for the 3d branch
	Camera     rl.Camera
	CameraMode rl.CameraMode

	// Velocità di rotazione
	MovementSpeed float32

	WorldBoundray rl.BoundingBox
	WorldCenter   rl.Vector3

	ControlMode   ControlMode
	SpawnDistance float32
	SpawnPosition rl.Vector3
}

func NewSimulation(config *config.Config) (*Simulation, error) {
	config.UpdateWindowSettings()
	WorldCenter := rl.NewVector3(float32(config.GameX)/2, float32(config.GameY)/2, float32(config.GameZ)/2)

	sim := &Simulation{
		Fluid:   make([]*Unit, 0, config.UnitNumber),
		Metrics: &metrics.Metrics{},
		Config:  config,
		IsPause: false,

		WorldBoundray: rl.NewBoundingBox(rl.NewVector3(-float32(config.GameX)/2, 0, -float32(config.GameZ)/2), rl.NewVector3(float32(config.GameX)/2, float32(config.GameY), float32(config.GameZ)/2)),
		WorldCenter:   WorldCenter,

		ControlMode:   UnitSpawnMode,
		SpawnDistance: 0,
		SpawnPosition: WorldCenter,
	}

	sim.Octree = NewOctree(1, sim.WorldBoundray)

	sim.ResetCameraPosition()

	return sim, nil
}

func (s *Simulation) ResetCameraPosition() {
	fovy := float32(60)
	fovyRadians := fovy * (math.Pi / 180)
	d := float32((math.Sqrt(3) * math.Max(float64(s.Config.GameX), math.Max(float64(s.Config.GameY), float64(s.Config.GameZ)))) / (2 * math.Tan(float64(fovyRadians)/2)))

	s.Camera = rl.Camera{
		Position:   rl.NewVector3(0, float32(d), float32(d)),
		Target:     s.WorldCenter,
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       fovy,
		Projection: rl.CameraPerspective,
	}

	s.SpawnDistance = rl.Vector3Distance(s.WorldCenter, s.Camera.Position)
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
		s.ResetSimulation()
	} else if rl.IsKeyPressed(rl.KeyC) {
		s.ResetCameraPosition()
	} else if rl.IsKeyPressed(rl.KeyOne) {
		s.CameraMode = rl.CameraFree
	} else if rl.IsKeyPressed(rl.KeyTwo) {
		s.CameraMode = rl.CameraOrbital
	} else if rl.IsKeyPressed(rl.KeyThree) {
		s.CameraMode = rl.CameraFirstPerson
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
			s.SpawnNewUnits()
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

	rl.UpdateCamera(&s.Camera, s.CameraMode)

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

func (s *Simulation) SpawnNewUnits() {
	currentRadius := s.Config.UnitRadius * s.Config.UnitRadiusMultiplier
	currentMassMultiplier := s.Config.UnitMassMultiplier
	currentElasticity := s.Config.UnitElasticity

	unts := make([]*Unit, 0)

	for i := 0; i < int(s.Config.UnitNumber); i++ {
		if s.Config.SetRandomRadius {
			currentRadius = (s.Config.RadiusMin + rand.Float32()*(s.Config.RadiusMax-s.Config.RadiusMin)) * s.Config.UnitRadiusMultiplier
		}
		if s.Config.SetRandomMassMultiplier {
			currentMassMultiplier = s.Config.MassMultiplierMin + rand.Float32()*(s.Config.MassMultiplierMax-s.Config.MassMultiplierMin)
		}
		if s.Config.SetRandomElasticity {
			currentElasticity = s.Config.ElasticityMin + rand.Float32()*(s.Config.ElasticityMax-s.Config.ElasticityMin)
		}

		color := color.RGBA{255, 0, 0, 255}

		if s.Config.SetRandomColor {
			color = utils.RandomRaylibColor()
		}

		unts = append(unts, newUnitWithPropertiesAtPosition(rl.Vector3{}, rl.Vector3{}, currentRadius, currentMassMultiplier, currentElasticity, color))
	}

	positionSpheres(unts, s.SpawnPosition)
	s.Fluid = append(s.Fluid, unts...)
}

func (s *Simulation) InitTest() {
	s.Fluid = append(s.Fluid, newUnitWithPropertiesAtPosition(s.WorldCenter, rl.Vector3{}, 1, 1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255}))
}

func (s *Simulation) ResetSimulation() {
	s.Octree.Clear()
	s.Fluid = []*Unit{}
}

func positionSpheres(units []*Unit, cubeCenter rl.Vector3) {
	numSpheres := len(units)
	cubeSideLength := int(math.Ceil(math.Pow(float64(numSpheres), 1.0/3.0)))

	// Calcola il lato di ogni cubo in base al numero di sfere
	cubeSide := float32(cubeSideLength)

	// Calcola il passo tra ogni sfera
	step := 2 * cubeSide / float32(cubeSideLength-1)

	// Posiziona le sfere all'interno del cubo
	index := 0
	for x := 0; x < cubeSideLength; x++ {
		for y := 0; y < cubeSideLength; y++ {
			for z := 0; z < cubeSideLength; z++ {
				if index < numSpheres {
					// Calcola la posizione della sfera rispetto al centro del cubo
					posX := float32(x)*step - cubeSide + cubeCenter.X
					posY := float32(y)*step - cubeSide + cubeCenter.Y
					posZ := float32(z)*step - cubeSide + cubeCenter.Z

					// Assegna la posizione alla sfera
					units[index].Position = rl.Vector3{X: posX, Y: posY, Z: posZ}
					units[index].PreviousPosition = units[index].Position
					index++
				}
			}
		}
	}
}
