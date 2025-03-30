package physics

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/metrics"
	"github.com/alexanderi96/go-fluid-simulator/physics/material"
	"github.com/alexanderi96/go-fluid-simulator/spaceship"
	"github.com/alexanderi96/go-fluid-simulator/utils"
	"github.com/google/uuid"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
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

	TimeScale            float64
	IsPause              bool
	Fly                  bool
	InitialMousePosition vector2.Float64 `json:"-"`
	FinalMousePosition   vector2.Float64 `json:"-"`
	MouseButtonPressed   bool            `json:"-"`
	IsInputBeingHandled  bool            `json:"-"`
	AppStartTime         time.Time

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
		TimeScaleLabel    *gui.Label
		NavigationLabel   *gui.Label
		ShipStatusLabel   *gui.Label
		PositionLabel     *gui.Label
		SpeedLabel        *gui.Label
		DirectionLabel    *gui.Label
		OrientationLabel  *gui.Label
		StatusLabel       *gui.Label
		// Add panel fields for responsive UI
		ModePanel     *gui.Panel `json:"-"`
		ControlsPanel *gui.Panel `json:"-"`
		KeysPanel     *gui.Panel `json:"-"`
		ShipPanel     *gui.Panel `json:"-"`
	}

	MovementSpeed float64 `json:"-"`

	WorldBoundray BoundingBox
	WorldCenter   vector3.Vector[float64]

	SpawnDistance        float64                 `json:"-"`
	InitialSpawnPosition vector3.Vector[float64] `json:"-"`
	FinalSpawnPosition   vector3.Vector[float64] `json:"-"`
}

func NewSimulation(config *config.Config) (*Simulation, error) {
	InitOctree(config)

	WorldCenter := vector3.New(0.0, 0.0, 0.0)
	sim := &Simulation{
		Fluid:     make([]*Unit, 0, config.UnitNumber),
		Metrics:   &metrics.Metrics{},
		Config:    config,
		TimeScale: 1.0,
		IsPause:   false,
		Fly:       false,
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

	if config.CentralMass > 0 {
		sim.Fluid = append(sim.Fluid, sim.newUnitWithPropertiesAtPosition(WorldCenter, static, static, 0.01, config.CentralMass, 0, false))
	}

	// if sim.SpaceShip != nil {
	// 	sim.SpaceShip.SetupShip()
	// 	sim.Scene.Add(sim.SpaceShip.Ship)
	// }

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
	encoder.SetIndent("", "    ")
	return encoder.Encode(sim)
}

func LoadSimulation(filePath string) (*Simulation, error) {
	// Load JSON simulation file
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
	deltaTime := float64(time.Since(s.AppStartTime).Seconds()) - s.Metrics.SimDuration
	s.Metrics.Update(deltaTime)
	return s.UpdateWithOctrees()
}

func (s *Simulation) GetDeltaTime() float64 {
	return (float64(time.Since(s.AppStartTime).Seconds()) - s.Metrics.SimDuration) * s.TimeScale
}

func (s *Simulation) newUnitWithPropertiesAtPosition(position, acceleration, velocity vector3.Vector[float64], radius float64, density float64, elasticity float64, canBeAltered bool) *Unit {
	// Create a composition with Iron for the central mass
	comp := material.NewComposition(map[*material.Material]float64{
		&material.Iron: 1.0,
	})

	unit := &Unit{
		Id:           uuid.New(),
		_position:    position,
		_velocity:    velocity,
		Acceleration: acceleration,
		_radius:      radius,
		Composition:  comp,
		Heat:         0.0,
		canBeAltered: canBeAltered,
		config:       s.Config,
	}

	unit.NewPointLightMesh()
	s.Scene.Add(unit.Mesh)

	return unit
}

func (s *Simulation) GetUnits() []*Unit {
	currentRadius := s.Config.UnitRadius

	// Array of available materials
	materials := []*material.Material{&material.Iron, &material.Copper, &material.Ice}

	unts := make([]*Unit, 0)

	for i := 0; i < int(s.Config.UnitNumber); i++ {
		if s.Config.SetRandomRadius {
			currentRadius = s.Config.RadiusMin + rand.Float64()*(s.Config.RadiusMax-s.Config.RadiusMin)
		}

		// Select a random material
		randomMaterial := materials[rand.Intn(len(materials))]

		// Create composition with the random material
		comp := material.NewComposition(map[*material.Material]float64{
			randomMaterial: 1.0,
		})

		// Create unit with the random material composition
		unit := &Unit{
			Id:           uuid.New(),
			_position:    s.FinalSpawnPosition,
			_velocity:    static,
			Acceleration: static,
			_radius:      currentRadius,
			Composition:  comp,
			Heat:         0.0,
			canBeAltered: true,
			config:       s.Config,
		}

		unit.NewPointLightMesh()
		s.Scene.Add(unit.Mesh)
		unts = append(unts, unit)
	}
	return unts
}

func (s *Simulation) PositionNewUnitsCube(units []*Unit) {
	positionUnitsCuboidally(units, s.InitialSpawnPosition, s.Config.UnitInitialSpacing)
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

// RemoveUnit removes a unit from the simulation and cleans up its resources
func (s *Simulation) RemoveUnit(unit *Unit) {
	// Remove from scene
	if unit.Mesh != nil {
		s.Scene.Remove(unit.Mesh)
		unit.Mesh = nil
	}

	// Remove from Fluid array
	for i, u := range s.Fluid {
		if u.Id == unit.Id {
			// Remove by swapping with last element and truncating
			lastIdx := len(s.Fluid) - 1
			s.Fluid[i] = s.Fluid[lastIdx]
			s.Fluid = s.Fluid[:lastIdx]
			break
		}
	}
}

func (s *Simulation) GiveRotationalVelocity(units []*Unit) {
	for _, u := range units {
		u.CalcolaVettoreVelocitaRotazione(&s.WorldCenter)
	}
}

func (u *Unit) CalcolaVettoreVelocitaRotazione(p *vector3.Vector[float64]) {
	pos := u.Position()
	d := math.Sqrt(pos.X()*pos.X() + pos.Y()*pos.Y())
	k := 0.2
	v := k * d

	v_x := v * pos.Y() / d
	v_y := -v * pos.X() / d

	u.SetVelocity(vector3.New(v_x, v_y, 0))
}

func positionUnitsCuboidally(units []*Unit, finalSpawnPosition vector3.Vector[float64], spacing float64) error {
	if len(units) == 0 {
		return nil
	}

	n := len(units)
	sideLengthX, sideLengthY, sideLengthZ := optimalCuboidDimensions(n)

	unitRadius := units[0].Radius()

	totalWidth := float64(sideLengthX)*(2*unitRadius+spacing) - spacing
	totalHeight := float64(sideLengthY)*(2*unitRadius+spacing) - spacing
	totalDepth := float64(sideLengthZ)*(2*unitRadius+spacing) - spacing

	startX := finalSpawnPosition.X() - totalWidth/2
	startY := finalSpawnPosition.Y() - totalHeight/2
	startZ := finalSpawnPosition.Z() - totalDepth/2

	index := 0
	for x := 0; x < sideLengthX && index < n; x++ {
		for y := 0; y < sideLengthY && index < n; y++ {
			for z := 0; z < sideLengthZ && index < n; z++ {
				unitX := startX + float64(x)*(2*unitRadius+spacing)
				unitY := startY + float64(y)*(2*unitRadius+spacing)
				unitZ := startZ + float64(z)*(2*unitRadius+spacing)

				units[index].SetPosition(vector3.New(unitX, unitY, unitZ))
				index++
			}
		}
	}

	return nil
}

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
	phi := math.Phi
	angle := 0.0
	radiusStep := 0.3

	for i := 0; i < len(units); i++ {
		radius := math.Sqrt(float64(i)) * radiusStep
		x := center.X() + radius*math.Cos(angle)
		y := center.Y() + radius*math.Sin(angle)
		z := center.Z()

		units[i].SetPosition(vector3.New(x, y, z))
		radiusStep += 0.0005
		angle += phi * 2 * math.Pi
	}
}
