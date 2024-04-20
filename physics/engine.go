package physics

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/metrics"
	"github.com/alexanderi96/go-fluid-simulator/utils"
	"github.com/google/uuid"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
)

type BoundingBox struct {
	Min, Max vector3.Vector[float64]
}

type Simulation struct {
	Fluid   []*Unit
	Metrics *metrics.Metrics
	Config  *config.Config
	Octree  *Octree

	IsPause              bool
	InitialMousePosition vector2.Float64
	FinalMousePosition   vector2.Float64
	MouseButtonPressed   bool
	IsInputBeingHandled  bool

	// variables added for the g3n branch
	App   *app.Application
	Scene *core.Node
	Cam   *camera.Camera

	// Velocità di rotazione
	MovementSpeed float64

	WorldBoundray BoundingBox
	WorldCenter   vector3.Vector[float64]

	SpawnDistance        float64
	InitialSpawnPosition vector3.Vector[float64]
	FinalSpawnPosition   vector3.Vector[float64]
}

var (
	front, top, side vector3.Vector[float64]
	fovy             = 60.0
)

func NewSimulation(config *config.Config) (*Simulation, error) {
	// config.UpdateWindowSettings()

	InitOctree(config)

	WorldCenter := vector3.New(0.0, 0.0, 0.0)
	sim := &Simulation{
		Fluid:   make([]*Unit, 0, config.UnitNumber),
		Metrics: &metrics.Metrics{},
		Config:  config,
		IsPause: false,
		WorldBoundray: BoundingBox{
			Min: vector3.New(-config.GameX/2, -config.GameY/2, -config.GameZ/2),
			Max: vector3.New(config.GameX/2, config.GameY/2, config.GameZ/2),
		},
		WorldCenter: WorldCenter,

		App:   app.App(),
		Scene: core.NewNode(),

		SpawnDistance:        0,
		InitialSpawnPosition: WorldCenter,
		FinalSpawnPosition:   WorldCenter,
	}

	sim.Octree = NewOctree(0, sim.WorldBoundray, sim.Scene)

	fovyRadians := fovy * (math.Pi / 180)
	d := (math.Sqrt(3) * math.Max(config.GameX, math.Max(config.GameY, config.GameZ))) / (2 * math.Tan(fovyRadians/2))

	front = vector3.New(0, 0, d)
	top = vector3.New(0, config.GameY, 0)
	side = vector3.New(d, 0, 0)

	sim.ResetCameraPosition(front, fovy)

	return sim, nil
}

func (s *Simulation) ResetCameraPosition(position vector3.Vector[float64], fovy float64) {
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

	return s.FinalSpawnPosition.X() >= s.WorldBoundray.Min.X() && s.FinalSpawnPosition.X() <= s.WorldBoundray.Max.X() &&
		s.FinalSpawnPosition.Y() >= s.WorldBoundray.Min.Y() && s.FinalSpawnPosition.Y() <= s.WorldBoundray.Max.Y() &&
		s.FinalSpawnPosition.Z() >= s.WorldBoundray.Min.Z() && s.FinalSpawnPosition.Z() <= s.WorldBoundray.Max.Z()
}

func (s *Simulation) newUnitWithPropertiesAtPosition(position, acceleration, velocity vector3.Vector[float64], radius, massMultiplier, elasticity float64, color color.RGBA) *Unit {
	unitGeom := geometry.NewSphere(float64(radius), seg, seg)
	mat := material.NewStandard(utils.RgbaToMath32(color))
	unit := &Unit{
		Id:       uuid.New(),
		Mesh:     graphic.NewMesh(unitGeom, mat),
		Position: position,

		Velocity:       velocity,
		Acceleration:   acceleration,
		Radius:         radius,
		MassMultiplier: massMultiplier,
		Elasticity:     elasticity,
		Color:          color,
		Heat:           0.0,
	}

	s.Scene.Add(unit.Mesh)

	unit.Mass = unit.GetMass()

	return unit
}

func (s *Simulation) PositionNewUnitsCube(units []*Unit) {
	positionUnitsCuboidally(units, s.InitialSpawnPosition, s.Config.UnitInitialSpacing*s.Config.UnitRadiusMultiplier)
}

func (s *Simulation) GetUnits() []*Unit {
	currentRadius := s.Config.UnitRadius * s.Config.UnitRadiusMultiplier
	currentMassMultiplier := s.Config.UnitMassMultiplier
	currentElasticity := s.Config.UnitElasticity

	unts := make([]*Unit, 0)

	for i := 0; i < int(s.Config.UnitNumber); i++ {
		if s.Config.SetRandomRadius {
			currentRadius = (s.Config.RadiusMin + rand.Float64()*(s.Config.RadiusMax-s.Config.RadiusMin)) * s.Config.UnitRadiusMultiplier
		}
		if s.Config.SetRandomMassMultiplier {
			currentMassMultiplier = s.Config.MassMultiplierMin + rand.Float64()*(s.Config.MassMultiplierMax-s.Config.MassMultiplierMin)
		}
		if s.Config.SetRandomElasticity {
			currentElasticity = s.Config.ElasticityMin + rand.Float64()*(s.Config.ElasticityMax-s.Config.ElasticityMin)
		}

		color := color.RGBA{uint8(255), uint8(255), uint8(255), 255}

		// if s.Config.SetRandomColor {
		// 	color = utils.RandomRaylibColor()
		// }
		static := vector3.Zero[float64]()
		unts = append(unts, s.newUnitWithPropertiesAtPosition(s.FinalSpawnPosition, static, static, currentRadius, currentMassMultiplier, currentElasticity, color))
	}
	return unts
}

func (s *Simulation) PositionNewUnitsFibonacci(units []*Unit) {
	positionUnitsInFibonacciSpiral(units, &s.WorldCenter)
	s.Fluid = append(s.Fluid, units...)
}

func (s *Simulation) ResetSimulation() {
	s.Octree.Clear(s.Scene)
	s.Fluid = []*Unit{}
}

func positionUnitsCuboidally(units []*Unit, FinalspawnPosition vector3.Vector[float64], spacing float64) error {
	if len(units) == 0 {
		return nil
	}

	// Calcoliamo il lato del cubo arrotondando per eccesso
	sideLength := int(math.Ceil(math.Pow(float64(len(units)), 1.0/3.0)))
	unitRadius := units[0].Radius

	// Calcoliamo lo spazio totale richiesto per le unità
	totalWidth := float64(sideLength)*(2*unitRadius+spacing) - spacing
	totalHeight := float64(sideLength)*(2*unitRadius+spacing) - spacing
	totalDepth := float64(sideLength)*(2*unitRadius+spacing) - spacing

	// Calcoliamo la posizione iniziale del cubo
	startX := FinalspawnPosition.X() - totalWidth/2
	startY := FinalspawnPosition.Y() - totalHeight/2
	startZ := FinalspawnPosition.Z() - totalDepth/2

	// Posizioniamo le unità nel cubo
	index := 0
	for x := 0; x < sideLength; x++ {
		for y := 0; y < sideLength; y++ {
			for z := 0; z < sideLength; z++ {
				// Calcoliamo la posizione per questa unità
				unitX := startX + float64(x)*(2*unitRadius+spacing)
				unitY := startY + float64(y)*(2*unitRadius+spacing)
				unitZ := startZ + float64(z)*(2*unitRadius+spacing)

				// Assegniamo la posizione alla unità corrente
				if index < len(units) {
					units[index].Position = vector3.New(unitX, unitY, unitZ)
					index++
				} else {
					break
				}
			}
		}
	}

	return nil
}

func positionUnitsInFibonacciSpiral(units []*Unit, center *vector3.Vector[float64]) {
	phi := math.Phi // Phi è il rapporto aureo (1.618...)
	angle := 0.0
	radiusStep := 0.3 // Passo di incremento del raggio

	for i := 0; i < len(units); i++ {
		// Calcola la posizione della prossima unità sulla spirale di Fibonacci
		radius := math.Sqrt(float64(i)) * radiusStep
		x := center.X() + radius*math.Cos(angle)
		y := center.Y() + radius*math.Sin(angle)
		zMin, zMax := -0.1, 0.1
		z := center.Z() + zMin + rand.Float64()*(zMax-zMin)

		// Assegna la posizione alla unità
		units[i].Position = vector3.New(x, y, z)

		// Aumenta il passo di incremento del raggio
		radiusStep += 0.0005 // Modifica la velocità di aumento a tuo piacimento

		// Aggiorna l'angolo per la prossima unità sulla spirale
		angle += phi * 2 * math.Pi // Incremento dell'angolo utilizzando Phi
	}
}

func (s *Simulation) GiveVelocity(units []*Unit) {
	for _, u := range units {
		u.Velocity = *CalcolaVettoreVelocita(&s.InitialSpawnPosition, &s.FinalSpawnPosition, s.Config.Frametime)
	}
}

func CalcolaVettoreVelocita(p1, p2 *vector3.Vector[float64], dt float64) *vector3.Vector[float64] {
	// Calcola la differenza tra la posizione finale e quella iniziale
	differenzaPosizione := p2.Sub(*p1)

	// Dividi la differenza di posizione per l'intervallo di tempo per ottenere il vettore velocità
	vettoreVelocita := differenzaPosizione.Scale(0.01 / dt)

	return &vettoreVelocita
}
