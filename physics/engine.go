package physics

import (
	"encoding/json"
	"image/color"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/metrics"
	"github.com/alexanderi96/go-fluid-simulator/spaceship"
	"github.com/alexanderi96/go-fluid-simulator/utils"
	"github.com/google/uuid"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/window"
)

var (
	static = vector3.Zero[float64]()
)

type BoundingBox struct {
	Min, Max vector3.Vector[float64]
	Wf       *graphic.Lines
}

func (b *BoundingBox) DrawBounds(scene *core.Node) {
	b.Wf = utils.GetBoundsLine(b.Min, b.Max)
	b.Wf.SetVisible(true)
	scene.Add(b.Wf)
}

func (b *BoundingBox) RemoveBounds(scene *core.Node) {
	scene.Remove(b.Wf)
	b.Wf = nil
}

type Simulation struct {
	Fluid   []*Unit
	Metrics *metrics.Metrics `json:"-"`
	Config  *config.Config
	Octree  *Octree `json:"-"`

	IsPause              bool
	Fly                  bool
	InitialMousePosition vector2.Float64 `json:"-"`
	FinalMousePosition   vector2.Float64 `json:"-"`
	MouseButtonPressed   bool            `json:"-"`
	IsInputBeingHandled  bool            `json:"-"`
	AppStartTime         time.Time

	// variables added for the g3n branch
	App   *app.Application `json:"-"`
	Scene *core.Node       `json:"-"`
	Cam   *camera.Camera   `json:"-"`

	SpaceShip *spaceship.SpaceShip
	Hud       struct {
		FpsLabel          *gui.Label
		FtLabel           *gui.Label
		UnitLabel         *gui.Label
		SimDurationLabel  *gui.Label
		RealDurationLabel *gui.Label

		PositionLabel    *gui.Label
		SpeedLabel       *gui.Label
		DirectionLabel   *gui.Label
		OrientationLabel *gui.Label
		StatusLabel      *gui.Label
	}

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
		Fly:     true,
		WorldBoundray: BoundingBox{
			Min: vector3.New(-config.GameX/2, -config.GameY/2, -config.GameZ/2),
			Max: vector3.New(config.GameX/2, config.GameY/2, config.GameZ/2),
		},
		WorldCenter: WorldCenter,

		App:   app.App(),
		Scene: core.NewNode(),

		SpaceShip: &spaceship.SpaceShip{
			Speed:           0.0,
			MaxSpeed:        1e1,
			MaxEngineThrust: 1e5,
			Thrust:          0.0,
			RotationSpeed:   0.01,
			BreakingPower:   1e10,
			Keys:            make(map[window.Key]bool),
			CameraOffset:    math32.NewVector3(0, 5, -10),
		},

		SpawnDistance:        0,
		InitialSpawnPosition: WorldCenter,
		FinalSpawnPosition:   WorldCenter,
	}

	sim.App.IWindow.(*window.GlfwWindow).SetTitle("Go Fluid Simulator")
	sim.App.IWindow.(*window.GlfwWindow).SetSize(int(config.WindowWidth), int(config.WindowHeight))

	sim.Octree = NewOctree(0, sim.WorldBoundray, sim.Scene, config.ShowOctree)

	if config.CentralMass > 0 {
		sim.Fluid = append(sim.Fluid, sim.newUnitWithPropertiesAtPosition(WorldCenter, static, static, 0.01, config.CentralMass, 0, false, color.RGBA{uint8(255), uint8(1), uint8(1), 255}))
	}

	if config.GenerateWorld {
		sim.generatePlanetarySystem(0.0000001)
	}

	sim.SpaceShip.SetupShip()

	sim.Scene.Add(sim.SpaceShip.Ship)
	// planeAxes := helper.NewAxes(2.0)
	// sim.Scene.Add(planeAxes)

	// Create Skybox
	if sim.Config.ShowSkybox {
		skybox, err := graphic.NewSkybox(graphic.SkyboxData{
			"./assets/img/space/dark-s_", "jpg",
			[6]string{"px", "nx", "py", "ny", "pz", "nz"}})
		if err != nil {
			panic(err)
		}
		sim.Scene.Add(skybox)
	}

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

func (s *Simulation) newUnitWithPropertiesAtPosition(position, acceleration, velocity vector3.Vector[float64], radius, massMultiplier, elasticity float64, canBeAltered bool, color color.RGBA) *Unit {

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

		CanBeAltered: canBeAltered,
	}

	unit.NewPointLightMesh()

	s.Scene.Add(unit.Mesh)
	// s.Scene.Add(unit.Mesh.Light)

	unit.Mass = unit.GetMass()

	return unit
}

func (sim *Simulation) generatePlanetarySystem(sf float64) {
	// Dimensione del mondo
	worldSize := 1e5

	// Fattori di scala adattati al mondo
	maxSystemRadius := worldSize * 0.4
	starRadius := worldSize * 0.05

	// Massa della stella
	const SOLAR_MASS = 1e5

	// Creiamo la stella centrale
	star := sim.newUnitWithPropertiesAtPosition(
		vector3.New(0.0, 0.0, 0.0),
		vector3.New(0.0, 0.0, 0.0),
		vector3.New(0.0, 0.0, 0.0),
		starRadius,
		SOLAR_MASS,
		0.0,
		true,
		color.RGBA{255, 225, 0, 255},
	)
	sim.Fluid = append(sim.Fluid, star)

	// Numero di pianeti (3-6)
	numPlanets := rand.Intn(4) + 3

	// Parametri per la distribuzione dei pianeti
	minDistance := starRadius * 2.5
	maxDistance := maxSystemRadius

	planetSizeFactors := []float64{
		0.15, // Mercurio
		0.25, // Venere
		0.3,  // Terra
		0.2,  // Marte
		0.45, // Giove
		0.4,  // Saturno
		0.35, // Urano
		0.35, // Nettuno
	}

	for i := 0; i < numPlanets; i++ {
		// Distanza con progressione più compatta
		ratio := math.Pow(1.4, float64(i))
		distance := minDistance * ratio
		if distance > maxDistance {
			distance = maxDistance
		}

		// Calcolo dell'accelerazione gravitazionale iniziale
		// a = GM/r² * direzione
		acceleration := G * SOLAR_MASS / (distance * distance) * 15.0 // Fattore di scala per accelerazione più forte

		// Raggio del pianeta
		planetRadius := starRadius * planetSizeFactors[i%len(planetSizeFactors)]

		// Angolo casuale per posizione iniziale
		angle := rand.Float64() * 2 * math.Pi

		// Posizione iniziale su un'orbita circolare
		position := vector3.New(
			distance*math.Cos(angle),
			0.0,
			distance*math.Sin(angle),
		)

		// Vettore accelerazione perpendicolare alla posizione
		accelerationVector := vector3.New(
			-acceleration*math.Sin(angle),
			0.0,
			acceleration*math.Cos(angle),
		)

		// Massa del pianeta
		planetMass := 1.0

		// Creazione del pianeta con velocità iniziale zero e accelerazione perpendicolare
		planet := sim.newUnitWithPropertiesAtPosition(
			position,
			vector3.New(0.0, 0.0, 0.0), // Velocità iniziale zero
			accelerationVector,         // Accelerazione perpendicolare
			planetRadius,
			planetMass,
			0.2,
			true,
			generatePlanetColor(i),
		)

		sim.Fluid = append(sim.Fluid, planet)
	}
}

func generatePlanetColor(index int) color.RGBA {
	planetColors := []color.RGBA{
		{170, 150, 140, 255}, // Grigio-marrone (tipo Mercurio)
		{255, 198, 73, 255},  // Giallo-crema (tipo Venere)
		{100, 149, 237, 255}, // Blu (tipo Terra)
		{193, 68, 14, 255},   // Rosso (tipo Marte)
		{176, 127, 53, 255},  // Marrone chiaro (tipo Giove)
		{238, 232, 205, 255}, // Beige (tipo Saturno)
		{173, 216, 230, 255}, // Azzurro chiaro (tipo Urano)
		{0, 0, 128, 255},     // Blu scuro (tipo Nettuno)
	}
	return planetColors[index%len(planetColors)]
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
		unts = append(unts, s.newUnitWithPropertiesAtPosition(s.FinalSpawnPosition, static, static, currentRadius, currentMassMultiplier, currentElasticity, true, color))
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
