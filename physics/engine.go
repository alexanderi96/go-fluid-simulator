package physics

import (
	"encoding/json"
	"image/color"
	"math"
	"math/rand"
	"os"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/metrics"
	"github.com/google/uuid"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/window"
)

type BoundingBox struct {
	Min, Max vector3.Vector[float64]
}

type Simulation struct {
	Fluid   []*Unit
	Metrics *metrics.Metrics `json:"-"`
	Config  *config.Config
	Octree  *Octree `json:"-"`

	IsPause              bool
	InitialMousePosition vector2.Float64 `json:"-"`
	FinalMousePosition   vector2.Float64 `json:"-"`
	MouseButtonPressed   bool            `json:"-"`
	IsInputBeingHandled  bool            `json:"-"`

	// variables added for the g3n branch
	App   *app.Application `json:"-"`
	Scene *core.Node       `json:"-"`
	Cam   *camera.Camera   `json:"-"`

	// Velocità di rotazione
	MovementSpeed float64 `json:"-"`

	WorldBoundray BoundingBox
	WorldCenter   vector3.Vector[float64]

	SpawnDistance        float64                 `json:"-"`
	InitialSpawnPosition vector3.Vector[float64] `json:"-"`
	FinalSpawnPosition   vector3.Vector[float64] `json:"-"`
}

var (
	fovy = 60.0
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

	sim.App.IWindow.(*window.GlfwWindow).SetTitle("Go Fluid Simulator")
	sim.App.IWindow.(*window.GlfwWindow).SetSize(int(config.WindowWidth), int(config.WindowHeight))

	sim.Octree = NewOctree(0, sim.WorldBoundray, sim.Scene)

	return sim, nil
}

func (sim *Simulation) SaveSimulation(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ") // Per un output formattato
	return encoder.Encode(sim)
}

func LoadSimulation(filePath string) (*Simulation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sim Simulation
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&sim)
	if err != nil {
		return nil, err
	}

	return &sim, nil
}

func (s *Simulation) Update() error {
	s.Metrics.Update(s.Config.Frametime)

	return s.UpdateWithOctrees()
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

	unit := &Unit{
		Id:       uuid.New(),
		Position: position,

		Velocity:       velocity,
		Acceleration:   acceleration,
		Radius:         radius,
		MassMultiplier: massMultiplier,
		Elasticity:     elasticity,
		Color:          color,
		Heat:           0.0,
	}

	unit.GenerateMesh()

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
}

func (s *Simulation) ResetSimulation() {
	s.Octree.Clear(s.Scene)

	for _, unit := range s.Fluid {
		s.Scene.Remove(unit.Mesh)
	}
	s.Fluid = []*Unit{}

}

func positionUnitsCuboidally(units []*Unit, finalSpawnPosition vector3.Vector[float64], spacing float64) error {
	if len(units) == 0 {
		return nil
	}

	// Calcoliamo le dimensioni ottimali del cubo
	n := len(units)
	sideLengthX, sideLengthY, sideLengthZ := optimalCuboidDimensions(n)

	unitRadius := units[0].Radius

	// Calcoliamo lo spazio totale richiesto per le unità
	totalWidth := float64(sideLengthX)*(2*unitRadius+spacing) - spacing
	totalHeight := float64(sideLengthY)*(2*unitRadius+spacing) - spacing
	totalDepth := float64(sideLengthZ)*(2*unitRadius+spacing) - spacing

	// Calcoliamo la posizione iniziale del cubo
	startX := finalSpawnPosition.X() - totalWidth/2
	startY := finalSpawnPosition.Y() - totalHeight/2
	startZ := finalSpawnPosition.Z() - totalDepth/2

	// Posizioniamo le unità nel cubo
	index := 0
	for x := 0; x < sideLengthX && index < n; x++ {
		for y := 0; y < sideLengthY && index < n; y++ {
			for z := 0; z < sideLengthZ && index < n; z++ {
				// Calcoliamo la posizione per questa unità
				unitX := startX + float64(x)*(2*unitRadius+spacing)
				unitY := startY + float64(y)*(2*unitRadius+spacing)
				unitZ := startZ + float64(z)*(2*unitRadius+spacing)

				// Assegniamo la posizione alla unità corrente
				units[index].Position = vector3.New(unitX, unitY, unitZ)
				index++
			}
		}
	}

	return nil
}

// Funzione per calcolare le dimensioni ottimali del cubo
func optimalCuboidDimensions(n int) (int, int, int) {
	sideLength := int(math.Ceil(math.Pow(float64(n), 1.0/3.0)))
	for x := sideLength; x > 0; x-- {
		for y := x; y > 0; y-- {
			z := int(math.Ceil(float64(n) / float64(x*y)))
			if x*y*z >= n {
				return x, y, z
			}
		}
	}
	return sideLength, sideLength, sideLength
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
		z := center.Z()

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

func (s *Simulation) GiveRotationalVelocity(units []*Unit) {
	for _, u := range units {
		u.CalcolaVettoreVelocitaRotazione(&s.WorldCenter)
	}
}

func CalcolaVettoreVelocita(p1, p2 *vector3.Vector[float64], dt float64) *vector3.Vector[float64] {
	// Calcola la differenza tra la posizione finale e quella iniziale
	differenzaPosizione := p2.Sub(*p1)

	// Dividi la differenza di posizione per l'intervallo di tempo per ottenere il vettore velocità
	vettoreVelocita := differenzaPosizione.Scale(0.01 / dt)

	return &vettoreVelocita
}

func (u *Unit) CalcolaVettoreVelocitaRotazione(p *vector3.Vector[float64]) {
	// Calcola la distanza dall'origine
	d := math.Sqrt(u.Position.X()*u.Position.X() + u.Position.Y()*u.Position.Y())

	// Calcola la velocità di rotazione proporzionale alla distanza
	k := 0.5 // Costante di proporzionalità (personalizzabile)
	v := k * d

	// Calcola le componenti di velocità lungo gli assi x e y
	v_x := v * u.Position.Y() / d
	v_y := -v * u.Position.X() / d

	// Crea un nuovo vettore velocità con le componenti calcolate
	u.Velocity = vector3.New(v_x, v_y, 0)
}
