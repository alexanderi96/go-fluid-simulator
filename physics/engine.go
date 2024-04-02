package physics

import (
	"fmt"
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
	Fluid                []*Unit
	ClusterMasses        map[uuid.UUID]float32
	Metrics              *metrics.Metrics
	Config               *config.Config
	IsPause              bool
	InitialMousePosition rl.Vector2
	CurrentMousePosition rl.Vector2
	MouseButtonPressed   bool
	IsInputBeingHandled  bool

	// variables added for the 3d branch

	Camera rl.Camera

	// Velocità di rotazione
	MovementSpeed float32

	CubeCenter rl.Vector3

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

		CubeCenter: cubeCenter,

		ControlMode:   UnitSpawnMode,
		SpawnDistance: 0,
		SpawnPosition: cubeCenter,
	}

	sim.ResetCameraPosition()

	return sim, nil
}

func (s *Simulation) ResetCameraPosition() {
	fovy := float32(60)
	fovyRadians := fovy * (math.Pi / 180)
	d := float32((math.Sqrt(3) * math.Max(float64(s.Config.GameX), math.Max(float64(s.Config.GameY), float64(s.Config.GameZ)))) / (2 * math.Tan(float64(fovyRadians)/2)))

	s.Camera = rl.Camera{
		Position:   rl.NewVector3(float32(s.Config.GameX)/2, float32(d), float32(s.Config.GameZ)/2),
		Target:     rl.NewVector3(float32(s.Config.GameX)/2, float32(s.Config.GameY)/2, float32(s.Config.GameZ)/2),
		Up:         rl.NewVector3(0, 0, 1),
		Fovy:       fovy,
		Projection: rl.CameraPerspective,
	}

	s.SpawnDistance = rl.Vector3Distance(s.CubeCenter, s.Camera.Position)
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
		s.ResetCameraPosition()
		s.Fluid = []*Unit{}
	} else if rl.IsKeyPressed(rl.KeySpace) {
		s.IsPause = !s.IsPause
	} else if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		s.InitialMousePosition = rl.GetMousePosition()
		s.MouseButtonPressed = true

	} else if s.MouseButtonPressed && rl.IsMouseButtonDown(rl.MouseLeftButton) {
		// Continua a registrare la posizione corrente del mouse mentre il pulsante è premuto
		s.CurrentMousePosition = rl.GetMousePosition()

	} else if s.MouseButtonPressed && rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		// Calcola il vettore di spostamento
		displacementVector := rl.Vector3{
			X: s.CurrentMousePosition.X - s.InitialMousePosition.X,
			Y: s.CurrentMousePosition.Y - s.InitialMousePosition.Y,
			Z: 0, // Assumendo un movimento del mouse in 2D
		}

		// Qui puoi calcolare il vettore di accelerazione in base al vettore di spostamento
		// e ad altre logiche specifiche del tuo gioco, come la velocità del movimento del mouse
		// e il tempo trascorso tra il press e il release del pulsante.
		accelerationVector := CalculateAccelerationVector(displacementVector)

		// Usa il vettore di accelerazione per spawnare le unità
		s.Fluid = append(s.Fluid, *newUnitsWithAcceleration(s.SpawnPosition, s.Config, accelerationVector)...)

		s.MouseButtonPressed = false

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
		s.CalculateSpawnPosition()
	}

	s.IsInputBeingHandled = false
}

// Funzione ausiliaria per calcolare il vettore di accelerazione
func CalculateAccelerationVector(displacement rl.Vector3) rl.Vector3 {
	// Implementa qui la logica per calcolare l'accelerazione in base allo spostamento
	// e ad altri fattori come il tempo trascorso e la velocità del movimento del mouse.
	// Questo è solo un esempio e potrebbe richiedere un'implementazione specifica.
	acceleration := rl.Vector3{
		X: displacement.X / 2, // Esempio di calcolo
		Y: displacement.Y / 2, // Esempio di calcolo
		Z: 0,
	}
	return acceleration
}

// func (s *Simulation) HandleInput() {
// 	s.IsInputBeingHandled = true

// 	if rl.IsKeyPressed(rl.KeyR) {
// 		s.ResetCameraPosition()
// 		s.Fluid = []*Unit{}
// 	} else if rl.IsKeyPressed(rl.KeySpace) {
// 		s.IsPause = !s.IsPause
// 	} else if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
// 		s.InitialMousePosition = rl.GetMousePosition()

// 		for rl.IsMouseButtonDown(rl.MouseLeftButton) && s.InitialMousePosition.X > 0 && s.InitialMousePosition.X < float32(s.Config.ViewportX) &&
// 			s.InitialMousePosition.Y > 0 && s.InitialMousePosition.Y < float32(s.Config.ViewportY) {
// 			s.MouseButtonPressed = true

// 			s.CurrentMousePosition = rl.GetMousePosition()
// 		}

// 	} else if s.MouseButtonPressed && rl.IsMouseButtonReleased(rl.MouseLeftButton) {
// 		s.MouseButtonPressed = false

// 		s.Fluid = append(s.Fluid, *newUnitsWithAcceleration(s.SpawnPosition, s.Config, rl.Vector3{})...)

// 	} else if rl.IsKeyPressed(rl.KeyM) {
// 		// Cambio modalità con il tasto M (esempio)
// 		if s.ControlMode == CameraMovementMode {
// 			s.ControlMode = UnitSpawnMode
// 		} else {
// 			s.ControlMode = CameraMovementMode
// 		}
// 	}

// 	switch s.ControlMode {
// 	case CameraMovementMode:
// 		s.UpdateCameraPosition()

// 	case UnitSpawnMode:
// 		s.CalculateSpawnPosition()
// 	}

// 	s.IsInputBeingHandled = false
// }

func (s *Simulation) UpdateCameraPosition() error {
	rl.UpdateCamera(&s.Camera, rl.CameraFree)

	return nil
}

func (s *Simulation) CalculateSpawnPosition() {
	mouseRay := rl.GetMouseRay(rl.GetMousePosition(), s.Camera)

	// Calcola la distanza basata sulla rotazione della rotella del mouse
	s.SpawnDistance += rl.GetMouseWheelMove() // Adatta questa formula secondo le tue necessità

	// Calcola la posizione del segnalino di anteprima lungo il raggio
	s.SpawnPosition = rl.Vector3Add(mouseRay.Position, rl.Vector3Scale(mouseRay.Direction, s.SpawnDistance))

}

func (s *Simulation) IsSpawnInRange() bool {
	return s.SpawnPosition.X >= 0 && s.SpawnPosition.X <= float32(s.Config.GameX) &&
		s.SpawnPosition.Y >= 0 && s.SpawnPosition.Y <= float32(s.Config.GameY) && s.SpawnPosition.Z >= 0 && s.SpawnPosition.Z <= float32(s.Config.GameZ)
}
