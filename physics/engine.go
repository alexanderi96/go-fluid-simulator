package physics

import (
	"math"

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

var (
	front, top, side rl.Vector3
	fovy             = float32(60)
)

func NewSimulation(config *config.Config) (*Simulation, error) {
	config.UpdateWindowSettings()

	InitOctree(config)

	WorldCenter := rl.NewVector3(0, 0, 0)

	sim := &Simulation{
		Fluid:   make([]*Unit, 0, config.UnitNumber),
		Metrics: &metrics.Metrics{},
		Config:  config,
		IsPause: false,

		WorldBoundray: rl.NewBoundingBox(rl.NewVector3(-float32(config.GameX)/2, -float32(config.GameY)/2, -float32(config.GameZ)/2), rl.NewVector3(float32(config.GameX)/2, float32(config.GameY)/2, float32(config.GameZ)/2)),
		WorldCenter:   WorldCenter,

		ControlMode:   UnitSpawnMode,
		SpawnDistance: 0,
		SpawnPosition: WorldCenter,
	}

	sim.Octree = NewOctree(0, sim.WorldBoundray)

	fovyRadians := fovy * (math.Pi / 180)
	d := float32((math.Sqrt(3) * math.Max(float64(config.GameX), math.Max(float64(config.GameY), float64(config.GameZ)))) / (2 * math.Tan(float64(fovyRadians)/2)))

	front = rl.NewVector3(0, 0, float32(d))
	top = rl.NewVector3(0, config.GameY, 0)
	side = rl.NewVector3(float32(d), 0, 0)

	sim.ResetCameraPosition(front, fovy)

	return sim, nil
}

func (s *Simulation) ResetCameraPosition(position rl.Vector3, fovy float32) {
	s.Camera = rl.Camera{
		Position:   position,
		Target:     s.WorldCenter,
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       fovy,
		Projection: rl.CameraPerspective,
	}

	s.SpawnDistance = rl.Vector3Distance(s.WorldCenter, s.Camera.Position)
}

func (s *Simulation) Update() error {
	s.Metrics.Update()

	return s.UpdateWithOctrees()
}

func (s *Simulation) HandleInput() {
	s.IsInputBeingHandled = true

	if rl.IsKeyPressed(rl.KeyR) {
		s.ResetSimulation()
	} else if rl.IsKeyPressed(rl.KeyOne) {
		s.ResetCameraPosition(front, fovy)
	} else if rl.IsKeyPressed(rl.KeyTwo) {
		s.ResetCameraPosition(top, fovy)
	} else if rl.IsKeyPressed(rl.KeyThree) {
		s.ResetCameraPosition(side, fovy)
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
		s.SpawnPosition.Y >= s.WorldBoundray.Min.Y && s.SpawnPosition.Y <= s.WorldBoundray.Max.Y &&
		s.SpawnPosition.Z >= s.WorldBoundray.Min.Z && s.SpawnPosition.Z <= s.WorldBoundray.Max.Z
}

func (s *Simulation) SpawnNewUnits() {
	currentRadius := s.Config.UnitRadius * s.Config.UnitRadiusMultiplier
	currentMassMultiplier := s.Config.UnitMassMultiplier
	currentElasticity := s.Config.UnitElasticity

	unts := make([]*Unit, 0)

	for i := 0; i < int(s.Config.UnitNumber); i++ {
		// if s.Config.SetRandomRadius {
		// 	currentRadius = (s.Config.RadiusMin + rand.Float32()*(s.Config.RadiusMax-s.Config.RadiusMin)) * s.Config.UnitRadiusMultiplier
		// }
		// if s.Config.SetRandomMassMultiplier {
		// 	currentMassMultiplier = s.Config.MassMultiplierMin + rand.Float32()*(s.Config.MassMultiplierMax-s.Config.MassMultiplierMin)
		// }
		// if s.Config.SetRandomElasticity {
		// 	currentElasticity = s.Config.ElasticityMin + rand.Float32()*(s.Config.ElasticityMax-s.Config.ElasticityMin)
		// }

		color := rl.RayWhite

		if s.Config.SetRandomColor {
			color = utils.RandomRaylibColor()
		}

		unts = append(unts, newUnitWithPropertiesAtPosition(rl.Vector3{}, rl.Vector3{X: 0, Y: 0, Z: 0}, currentRadius, currentMassMultiplier, currentElasticity, color))
	}

	// positionSpheres(unts, s.SpawnPosition)

	positionUnitsCuboidally(unts, s.SpawnPosition, s.Config.UnitInitialSpacing)
	s.Fluid = append(s.Fluid, unts...)
}

func (s *Simulation) ResetSimulation() {
	s.Octree.Clear()
	s.Fluid = []*Unit{}
}

func positionUnitsCuboidally(units []*Unit, spawnPosition rl.Vector3, spacing float32) error {
	if len(units) == 0 {
		return nil
	}

	// Calcoliamo il lato del cubo arrotondando per eccesso
	sideLength := int(math.Ceil(math.Pow(float64(len(units)), 1.0/3.0)))
	unitRadius := units[0].Radius

	// Calcoliamo lo spazio totale richiesto per le unità
	totalWidth := float32(sideLength)*(2*unitRadius+spacing) - spacing
	totalHeight := float32(sideLength)*(2*unitRadius+spacing) - spacing
	totalDepth := float32(sideLength)*(2*unitRadius+spacing) - spacing

	// Calcoliamo la posizione iniziale del cubo
	startX := spawnPosition.X - totalWidth/2
	startY := spawnPosition.Y - totalHeight/2
	startZ := spawnPosition.Z - totalDepth/2

	// Posizioniamo le unità nel cubo
	index := 0
	for x := 0; x < sideLength; x++ {
		for y := 0; y < sideLength; y++ {
			for z := 0; z < sideLength; z++ {
				// Calcoliamo la posizione per questa unità
				unitX := startX + float32(x)*(2*unitRadius+spacing)
				unitY := startY + float32(y)*(2*unitRadius+spacing)
				unitZ := startZ + float32(z)*(2*unitRadius+spacing)

				// Assegniamo la posizione alla unità corrente
				if index < len(units) {
					units[index].Position = rl.Vector3{X: unitX, Y: unitY, Z: unitZ}
					units[index].PreviousPosition = units[index].Position
					index++
				} else {
					break
				}
			}
		}
	}

	return nil
}

// func positionUnitsInFibonacciSpiral(units []*Unit, center rl.Vector3) {
// 	phi := float32(math.Phi) // Phi è il rapporto aureo (1.618...)
// 	angle := float32(0)
// 	radiusStep := float32(0.3) // Passo di incremento del raggio

// 	for i := 0; i < len(units); i++ {
// 		// Calcola la posizione della prossima unità sulla spirale di Fibonacci
// 		radius := float32(math.Sqrt(float64(i))) * radiusStep
// 		x := center.X + radius*float32(math.Cos(float64(angle)))
// 		y := center.Y + radius*float32(math.Sin(float64(angle)))
// 		z := center.Z

// 		// Assegna la posizione alla unità
// 		units[i].Position = rl.NewVector3(x, y, z)
// 		units[i].PreviousPosition = units[i].Position

// 		// Aumenta il passo di incremento del raggio
// 		radiusStep += 0.0005 // Modifica la velocità di aumento a tuo piacimento

// 		// Aggiorna l'angolo per la prossima unità sulla spirale
// 		angle += phi * 2 * float32(math.Pi) // Incremento dell'angolo utilizzando Phi
// 	}
// }
