package physics

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/metrics"
	"github.com/google/uuid"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/math32"
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
	Octree  *Octree

	IsPause              bool
	InitialMousePosition math32.Vector2
	FinalMousePosition   math32.Vector2
	MouseButtonPressed   bool
	IsInputBeingHandled  bool

	// variables added for the g3n branch
	App   *app.Application
	Scene *core.Node
	Cam   *camera.Camera

	// Velocità di rotazione
	MovementSpeed float32

	WorldBoundray *math32.Box3
	WorldCenter   *math32.Vector3

	SpawnDistance        float32
	InitialSpawnPosition *math32.Vector3
	FinalSpawnPosition   *math32.Vector3
}

var (
	front, top, side math32.Vector3
	fovy             = float32(60)
)

func NewSimulation(config *config.Config) (*Simulation, error) {
	// config.UpdateWindowSettings()

	InitOctree(config)

	WorldCenter := math32.NewVector3(0.0, 0.0, 0.0)
	sim := &Simulation{
		Fluid:   make([]*Unit, 0, config.UnitNumber),
		Metrics: &metrics.Metrics{},
		Config:  config,
		IsPause: false,
		WorldBoundray: &math32.Box3{
			Min: *math32.NewVector3(-config.GameX/2, -config.GameY/2, -config.GameZ/2),
			Max: *math32.NewVector3(config.GameX/2, config.GameY/2, config.GameZ/2),
		},
		WorldCenter: WorldCenter,

		App:   app.App(),
		Scene: core.NewNode(),

		SpawnDistance:        0,
		InitialSpawnPosition: WorldCenter,
		FinalSpawnPosition:   WorldCenter,
	}

	sim.Octree = NewOctree(0, sim.WorldBoundray, sim.Scene)

	fovyRadians := fovy * (math32.Pi / 180)
	d := float32((math32.Sqrt(3) * math32.Max(config.GameX, math32.Max(config.GameY, config.GameZ))) / (2 * math32.Tan(fovyRadians/2)))

	front = *math32.NewVector3(0, 0, d)
	top = *math32.NewVector3(0, config.GameY, 0)
	side = *math32.NewVector3(d, 0, 0)

	sim.ResetCameraPosition(front, fovy)

	return sim, nil
}

func (s *Simulation) ResetCameraPosition(position math32.Vector3, fovy float32) {
	// rlWc := utils.ToRlVector3(s.WorldCenter)
	// s.Camera = rl.Camera{
	// 	Position:   utils.ToRlVector3(position),
	// 	Target:     rlWc,
	// 	Up:         rl.NewVector3(0, 1, 0),
	// 	Fovy:       float32(fovy),
	// 	Projection: rl.CameraPerspective,
	// }

	s.SpawnDistance = 100
}

func (s *Simulation) Update() error {
	s.Metrics.Update()

	return s.UpdateWithOctrees()
}

func (s *Simulation) HandleInput() {
	s.IsInputBeingHandled = true

	// if rl.IsKeyPressed(rl.KeyR) {
	// 	s.ResetSimulation()
	// } else if rl.IsKeyPressed(rl.KeyOne) {
	// 	s.ResetCameraPosition(front, fovy)
	// } else if rl.IsKeyPressed(rl.KeyTwo) {
	// 	s.ResetCameraPosition(top, fovy)
	// } else if rl.IsKeyPressed(rl.KeyThree) {
	// 	s.ResetCameraPosition(side, fovy)
	// } else if rl.IsKeyPressed(rl.KeySpace) {
	// 	s.IsPause = !s.IsPause
	// } else if s.MouseButtonPressed && rl.IsMouseButtonReleased(rl.MouseLeftButton) {
	// 	s.MouseButtonPressed = false

	// 	if s.IsSpawnInRange() {
	// 		units := s.GetUnits()
	// 		s.PositionNewUnitsCube(units)

	// 		if s.InitialMousePosition != s.FinalMousePosition {
	// 			s.GiveVelocity(units)
	// 		}

	// 		s.Fluid = append(s.Fluid, units...)
	// 	}

	// } else if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
	// 	s.InitialMousePosition = rl.GetMousePosition()
	// 	//s.UpdateInitialSpawnPosition()
	// 	// updateSpawnPosition(&s.InitialSpawnPosition, &s.SpawnDistance, &s.Camera)

	// 	for rl.IsMouseButtonDown(rl.MouseLeftButton) && s.InitialMousePosition.X > 0 && s.InitialMousePosition.X < float32(s.Config.ViewportX) &&
	// 		s.InitialMousePosition.Y > 0 && s.InitialMousePosition.Y < float32(s.Config.ViewportY) {
	// 		s.MouseButtonPressed = true

	// 		s.FinalMousePosition = rl.GetMousePosition()
	// 		// updateSpawnPosition(&s.FinalSpawnPosition, &s.SpawnDistance, &s.Camera)
	// 		//s.UpdateFinalSpawnPosition()
	// 	}

	// } else if rl.IsMouseButtonReleased(rl.MouseRightButton) {

	// 	if s.IsSpawnInRange() {
	// 		s.PositionNewUnitsFibonacci(s.GetUnits())
	// 	}

	// } else if rl.IsKeyPressed(rl.KeyM) {
	// 	// Cambio modalità con il tasto M (esempio)
	// 	if s.ControlMode == CameraMovementMode {
	// 		s.ControlMode = UnitSpawnMode
	// 	} else {
	// 		s.ControlMode = CameraMovementMode
	// 	}
	// }

	// switch s.ControlMode {
	// case CameraMovementMode:
	// 	s.UpdateCameraPosition()

	// case UnitSpawnMode:
	// 	// updateSpawnPosition(&s.FinalSpawnPosition, &s.SpawnDistance, &s.Camera)

	// 	//		s.UpdateFinalSpawnPosition()

	// default:
	// }

	s.IsInputBeingHandled = false
}

func (s *Simulation) UpdateCameraPosition() error {

	// rl.UpdateCamera(&s.Camera, s.CameraMode)

	return nil
}

// func updateSpawnPosition(position *math32.Vector3, spawnDistance *float64, camera *rl.Camera3D) {
// 	mouseRay := rl.GetMouseRay(rl.GetMousePosition(), *camera)

// 	// Calcola la distanza basata sulla rotazione della rotella del mouse
// 	*spawnDistance += float64(rl.GetMouseWheelMove()) // Adatta questa formula secondo le tue necessità

//		// Calcola la posizione del segnalino di anteprima lungo il raggio
//		*position = utils.ToVector3FromRlVector3(rl.Vector3Add(mouseRay.Position, rl.Vector3Scale(mouseRay.Direction, float32(*spawnDistance))))
//	}
func (s *Simulation) IsSpawnInRange() bool {

	return s.FinalSpawnPosition.X >= s.WorldBoundray.Min.X && s.FinalSpawnPosition.X <= s.WorldBoundray.Max.X &&
		s.FinalSpawnPosition.Y >= s.WorldBoundray.Min.Y && s.FinalSpawnPosition.Y <= s.WorldBoundray.Max.Y &&
		s.FinalSpawnPosition.Z >= s.WorldBoundray.Min.Z && s.FinalSpawnPosition.Z <= s.WorldBoundray.Max.Z
}

func (s *Simulation) newUnitWithPropertiesAtPosition(position, acceleration, velocity *math32.Vector3, radius, massMultiplier, elasticity float32, color color.RGBA) *Unit {
	unitGeom := geometry.NewSphere(float64(radius), seg, seg)

	unit := &Unit{
		Id:   uuid.New(),
		Mesh: graphic.NewMesh(unitGeom, mat),
		//Position: position,
		//PreviousPosition: position,
		Velocity:       velocity,
		Acceleration:   acceleration,
		Radius:         radius,
		MassMultiplier: massMultiplier,
		Elasticity:     elasticity,
		Color:          color,
		Heat:           0.0,
	}

	unit.Mesh.SetPositionVec(position)

	s.Scene.Add(unit.Mesh)

	unit.Mass = unit.GetMass()

	return unit
}

func (s *Simulation) PositionNewUnitsCube(units []*Unit) {
	positionUnitsCuboidally(units, s.InitialSpawnPosition, s.Config.UnitInitialSpacing)
}

func (s *Simulation) GetUnits() []*Unit {
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

		color := color.RGBA{uint8(255), uint8(255), uint8(255), 255}

		// if s.Config.SetRandomColor {
		// 	color = utils.RandomRaylibColor()
		// }
		static := *math32.NewVector3(0.0, 0.0, 0.0)
		unts = append(unts, s.newUnitWithPropertiesAtPosition(s.FinalSpawnPosition, &static, &static, currentRadius, currentMassMultiplier, currentElasticity, color))
	}
	return unts
}

func (s *Simulation) PositionNewUnitsFibonacci(units []*Unit) {
	positionUnitsInFibonacciSpiral(units, s.WorldCenter)
	s.Fluid = append(s.Fluid, units...)
}

func (s *Simulation) ResetSimulation() {
	s.Octree.Clear(s.Scene)
	s.Fluid = []*Unit{}
}

func positionUnitsCuboidally(units []*Unit, FinalspawnPosition *math32.Vector3, spacing float32) error {
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
	startX := FinalspawnPosition.X - totalWidth/2
	startY := FinalspawnPosition.Y - totalHeight/2
	startZ := FinalspawnPosition.Z - totalDepth/2

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
					units[index].Mesh.SetPosition(unitX, unitY, unitZ)
					index++
				} else {
					break
				}
			}
		}
	}

	return nil
}

func positionUnitsInFibonacciSpiral(units []*Unit, center *math32.Vector3) {
	phi := math.Phi // Phi è il rapporto aureo (1.618...)
	angle := 0.0
	radiusStep := 0.3 // Passo di incremento del raggio

	for i := 0; i < len(units); i++ {
		// Calcola la posizione della prossima unità sulla spirale di Fibonacci
		radius := math.Sqrt(float64(i)) * radiusStep
		x := center.X + float32(radius*math.Cos(float64(angle)))
		y := center.Y + float32(radius*math.Sin(float64(angle)))
		zMin, zMax := float32(-0.1), float32(0.1)
		z := center.Z + zMin + rand.Float32()*(zMax-zMin)

		// Assegna la posizione alla unità
		units[i].Mesh.SetPosition(x, y, z)

		// Aumenta il passo di incremento del raggio
		radiusStep += 0.0005 // Modifica la velocità di aumento a tuo piacimento

		// Aggiorna l'angolo per la prossima unità sulla spirale
		angle += phi * 2 * math.Pi // Incremento dell'angolo utilizzando Phi
	}
}

func (s *Simulation) GiveVelocity(units []*Unit) {
	for _, u := range units {
		u.Velocity = CalcolaVettoreVelocita(s.InitialSpawnPosition, s.FinalSpawnPosition, s.Config.Frametime)
	}
}

func CalcolaVettoreVelocita(p1, p2 *math32.Vector3, dt float32) *math32.Vector3 {
	// Calcola la differenza tra la posizione finale e quella iniziale
	differenzaPosizione := p2.Sub(p1)

	// Dividi la differenza di posizione per l'intervallo di tempo per ottenere il vettore velocità
	vettoreVelocita := differenzaPosizione.AddScalar(0.01 / dt)

	return vettoreVelocita
}
