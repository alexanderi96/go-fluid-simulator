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
	History      []UnitSnapshot // Coda di snapshot	Elasticity   float32
}

type UnitSnapshot struct {
	Snapshot  Unit
	Timestamp float32
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

func newUnitsAtPosition(spawnPosition rl.Vector2, simulationWidth, simulationHeight, unitNumber int32, unitRadius, unitMass, initialSpacing, scaleFactor, elasticity float32) *[]Unit {
	units := make([]Unit, 0, unitNumber) // Pre-allocazione
	centerX := spawnPosition.X
	centerY := spawnPosition.Y

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

			// Verifica ed adatta la posizione se è fuori dai confini
			x = clamp(x, unitRadius, float32(simulationWidth)-unitRadius)
			y = clamp(y, unitRadius, float32(simulationHeight)-unitRadius)

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

	return &units
}

func clamp(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
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

func (u *Unit) ApplyExternalForce(dt float32, externalForces rl.Vector2) {
	// Aggiorna la velocità dell'unità in base all'accelerazione e le forze esterne
	u.Velocity.X += (u.Acceleration.X + externalForces.X) * dt
	u.Velocity.Y += (u.Acceleration.Y + externalForces.Y) * dt
}

func (u *Unit) TakeSnapshot(dt float32) {
	// Crea un nuovo oggetto UnitSnapshot con le informazioni correnti
	snapshotInfo := UnitSnapshot{
		Snapshot:  *u, // Crea una copia dell'oggetto Unit corrente
		Timestamp: dt, // Assumendo che dt restituisca il tempo corrente
	}

	// Aggiungi il nuovo oggetto UnitSnapshot alla coda
	u.History = append(u.History, snapshotInfo)

	maxHistoryLength := 10 // Sostituisci con il numero desiderato di snapshot da memorizzare
	if len(u.History) > maxHistoryLength {
		u.History = u.History[1:]
	}
}

func (u *Unit) Update(dt float32, cfg *config.Config) error {
	// Controlla se il raggio dell'unità è negativo
	if u.Radius < 0 {
		return fmt.Errorf("l'unità %v ha un raggio negativo: %v", u.Id, u.Radius)
	}

	// Aggiorna la posizione dell'unità in base alla velocità
	u.Position.X += u.Velocity.X * dt
	u.Position.Y += u.Velocity.Y * dt

	// Resetta l'accelerazione per il prossimo frame
	u.Acceleration = rl.Vector2{X: 0, Y: 0}

	u.TakeSnapshot(dt)

	return nil
}
