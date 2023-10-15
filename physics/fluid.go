package physics

import (
	"fmt"
	"image/color"
	"math"

	"github.com/alexanderi96/go-fluid-simulator/config"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

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
				Velocity:   rl.Vector2{X: 0, Y: 100},
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

func (u *Unit) UpdateUnit(dt float32, externalForces rl.Vector2, cfg *config.Config) error {
	// Controlla se il raggio dell'unità è negativo
	if u.Radius < 0 {
		return fmt.Errorf("l'unità %v ha un raggio negativo: %v", u.Id, u.Radius)
	}

	// Aggiorna la velocità dell'unità in base all'accelerazione e le forze esterne
	u.Velocity.X += (u.Acceleration.X + externalForces.X) * dt
	u.Velocity.Y += (u.Acceleration.Y + externalForces.Y) * dt

	// Aggiorna la posizione dell'unità in base alla velocità
	u.Position.X += u.Velocity.X * dt
	u.Position.Y += u.Velocity.Y * dt

	// Controlla e corregge la posizione X
	if u.Position.X-u.Radius < 0 {
		u.Position.X = u.Radius
		u.Velocity.X = -u.Velocity.X // Invertire la velocità X
	} else if u.Position.X+u.Radius > float32(cfg.GameWidth) {
		u.Position.X = float32(cfg.GameWidth) - u.Radius
		u.Velocity.X = -u.Velocity.X // Invertire la velocità X
	}

	// Controlla e corregge la posizione Y
	if u.Position.Y-u.Radius < 0 {
		u.Position.Y = u.Radius
		u.Velocity.Y = -u.Velocity.Y // Invertire la velocità Y
	} else if u.Position.Y+u.Radius > float32(cfg.WindowHeight) {
		u.Position.Y = float32(cfg.WindowHeight) - u.Radius
		u.Velocity.Y = -u.Velocity.Y // Invertire la velocità Y
	}

	// Resetta l'accelerazione per il prossimo frame
	u.Acceleration = rl.Vector2{X: 0, Y: 0}

	return nil
}
