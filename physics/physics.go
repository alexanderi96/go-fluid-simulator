package physics

import (
	"image/color"
	"math"

	"github.com/alexanderi96/go-fluid-simulator/config"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

type Simulation struct {
	Fluid   *Fluid
	Metrics map[string]uint64
	Config  *config.Config
	IsPause bool
}

type Fluid struct {
	Units []Unit
}

type Unit struct {
	Id           uuid.UUID
	Position     rl.Vector2
	Velocity     rl.Vector2
	Acceleration rl.Vector2
	Elasticity   float32
	Mass         float32
	Radius       float32
	Color        color.RGBA
}

// Calcola il volume di una singola unità
func (u *Unit) Volume() float32 {
	return math.Pi * u.Radius * u.Radius
}

// Calcola la densità del fluido
func (f *Fluid) Density() float32 {
	var totalMass, totalVolume float32
	for _, unit := range f.Units {
		totalMass += unit.Mass
		totalVolume += unit.Volume()
	}
	return totalMass / totalVolume
}

// Aggiunge una unità al fluido
func (f *Fluid) AddUnit(unit Unit) {
	f.Units = append(f.Units, unit)
}

// Rimuove una unità dal fluido dato un indice
func (f *Fluid) RemoveUnit(index int) {
	f.Units = append(f.Units[:index], f.Units[index+1:]...)
}

// La funzione che si occupa di creare una nuova simulazione
func NewSimulation(config *config.Config) (*Simulation, error) {
	// Ad esempio, se desideri passare alcuni valori di config a NewFluid:
	fluid := newFluid(config.GameWidth, config.WindowHeight, config.ParticleNumber, config.ParticleRadius, config.ParticleMass, config.ParticleInitialSpacing, config.ScaleFactor, config.ParticleElasticity)

	return &Simulation{
		Fluid:   fluid,
		Metrics: make(map[string]uint64),
		Config:  config,
		IsPause: false,
	}, nil
}

func newFluid(simulationWidth, simulationHeight, unitNumber int32, unitRadius, unitMass, initialSpacing, scaleFactor, elasticity float32) *Fluid {
	units := make([]Unit, 0, unitNumber) // Pre-allocazione
	centerX := float32(simulationWidth) / 2
	centerY := float32(simulationHeight) / 2

	gap := initialSpacing // Spazio tra unità

	r := float64(unitRadius + gap)
	for len(units) < int(unitNumber) {
		// Calcolare il numero di unità che possono stare in un cerchio di raggio r
		circumference := 2 * math.Pi * r
		numUnits := int(circumference) / int(unitRadius*2+gap)
		if numUnits == 0 {
			numUnits = 1 // Assicura che ci sia almeno 1 unità
		}

		angleIncrement := (math.Pi * 2) / float64(numUnits) // angolo tra ogni particella

		for i := 0; i < numUnits && len(units) < int(unitNumber); i++ {
			angle := float32(i) * float32(angleIncrement)

			x := centerX + scaleFactor*float32(r*math.Cos(float64(angle)))
			y := centerY + scaleFactor*float32(r*math.Sin(float64(angle)))
			unit := Unit{
				Id:         uuid.New(),
				Position:   rl.Vector2{X: x, Y: y},
				Velocity:   rl.Vector2{X: 0, Y: 0},
				Radius:     unitRadius,
				Mass:       unitMass,
				Elasticity: elasticity,
				Color:      color.RGBA{255, 0, 0, 255}, // Red color for illustration
			}

			units = append(units, unit)
		}

		r += float64(unitRadius*2 + gap) // aumenta il raggio per il prossimo cerchio
	}

	return &Fluid{Units: units}
}

// La funzione che si occupa di resettare il fluido
func (s *Simulation) Reset() {
	s.Fluid = newFluid(s.Config.GameWidth, s.Config.WindowHeight, s.Config.ParticleNumber, s.Config.ParticleRadius, s.Config.ParticleMass, s.Config.ParticleInitialSpacing, s.Config.ScaleFactor, s.Config.ParticleElasticity)
}

// La funzione che si occupa di aggiornale la simulazione ad ogni frame. vengono controllati gli spostamenti, le accelerazioni e le collisioni delle particelle
func (s *Simulation) Update(frametime float32) {
	subSteps := 10                      // Numero di sub-steps per l'aggiornamento
	dt := frametime / float32(subSteps) // Tempo per sub-step

	for step := 0; step < subSteps; step++ {
		// Applica la forza di gravità a tutte le particelle
		if s.Config.ApplyGravity {
			for i := range s.Fluid.Units {
				unit := &s.Fluid.Units[i]
				unit.Acceleration = rl.Vector2{X: 0, Y: s.Config.Gravity}
				unit.Velocity.X += unit.Acceleration.X * dt
				unit.Velocity.Y += unit.Acceleration.Y * dt
			}
		}

		// Controlla le collisioni con le pareti e aggiorna la velocità delle particelle
		for i := range s.Fluid.Units {
			unit := &s.Fluid.Units[i]
			if unit.Position.X-unit.Radius < 0 || unit.Position.X+unit.Radius > float32(s.Config.GameWidth) {
				unit.Velocity.X = -unit.Velocity.X * unit.Elasticity
			}
			if unit.Position.Y-unit.Radius < 0 || unit.Position.Y+unit.Radius > float32(s.Config.WindowHeight) {
				unit.Velocity.Y = -unit.Velocity.Y * unit.Elasticity
			}
		}

		// Controlla le collisioni tra particelle e aggiorna le velocità
		for i := range s.Fluid.Units {
			for j := i + 1; j < len(s.Fluid.Units); j++ {
				unitA := &s.Fluid.Units[i]
				unitB := &s.Fluid.Units[j]
				deltaX := unitB.Position.X - unitA.Position.X
				deltaY := unitB.Position.Y - unitA.Position.Y
				distanceSquared := deltaX*deltaX + deltaY*deltaY
				totalRadius := unitA.Radius + unitB.Radius
				if distanceSquared < totalRadius*totalRadius {
					// Calcola la risposta alla collisione (questo è molto semplificato)
					normalX := float64(deltaX) / math.Sqrt(float64(distanceSquared))
					normalY := float64(deltaY) / math.Sqrt(float64(distanceSquared))
					relativeVelocityX := unitB.Velocity.X - unitA.Velocity.X
					relativeVelocityY := unitB.Velocity.Y - unitA.Velocity.Y
					dotProduct := float32(normalX)*relativeVelocityX + float32(normalY)*relativeVelocityY
					if dotProduct < 0 {
						coefficientOfRestitution := (unitA.Elasticity + unitB.Elasticity) / 2
						impulse := 2 * dotProduct / (unitA.Mass + unitB.Mass)
						unitA.Velocity.X += impulse * unitB.Mass * float32(normalX) * coefficientOfRestitution
						unitA.Velocity.Y += impulse * unitB.Mass * float32(normalY) * coefficientOfRestitution
						unitB.Velocity.X -= impulse * unitA.Mass * float32(normalX) * coefficientOfRestitution
						unitB.Velocity.Y -= impulse * unitA.Mass * float32(normalY) * coefficientOfRestitution
					}
				}
			}
		}

		// Aggiorna la posizione delle particelle in base alla loro velocità
		for i := range s.Fluid.Units {
			unit := &s.Fluid.Units[i]
			unit.Position.X += unit.Velocity.X * dt
			unit.Position.Y += unit.Velocity.Y * dt
		}
	}
}
