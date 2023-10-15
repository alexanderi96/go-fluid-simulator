package physics

import (
	"image/color"
	"math"

	"github.com/alexanderi96/go-fluid-simulator/config"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

var (
	gravity = rl.Vector2{X: 0, Y: 9.81}
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
	Mass         float32
	Radius       float32
	Color        color.RGBA
}

// La funzione che si occupa di creare una nuova simulazione
func NewSimulation(config *config.Config) (*Simulation, error) {
	// Ad esempio, se desideri passare alcuni valori di config a NewFluid:
	fluid := newFluid(config.GameWidth, config.WindowHeight, config.ParticleNumber, config.ParticleRadius, config.ParticleMass)

	return &Simulation{
		Fluid:   fluid,
		Metrics: make(map[string]uint64),
		Config:  config,
		IsPause: false,
	}, nil
}

func newFluid(simulationWidth, simulationHeight, unitNumber int32, unitRadius, unitMass float32) *Fluid {
	units := make([]Unit, 0, unitNumber) // Pre-allocazione
	centerX := float32(simulationWidth) / 2
	centerY := float32(simulationHeight) / 2

	gap := float32(1) // Spazio tra unità

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

			x := centerX + float32(r*math.Cos(float64(angle)))
			y := centerY + float32(r*math.Sin(float64(angle)))
			unit := Unit{
				Id:       uuid.New(),
				Position: rl.Vector2{X: x, Y: y},
				Radius:   unitRadius,
				Mass:     unitMass,
				Color:    color.RGBA{255, 0, 0, 255}, // Red color for illustration
			}

			units = append(units, unit)
		}

		r += float64(unitRadius*2 + gap) // aumenta il raggio per il prossimo cerchio
	}

	return &Fluid{Units: units}
}

// La funzione che si occupa di resettare il fluido
func (s *Simulation) Reset() {
	s.Fluid = newFluid(s.Config.GameWidth, s.Config.WindowHeight, s.Config.ParticleNumber, s.Config.ParticleRadius, s.Config.ParticleMass)
}

// La funzione che si occupa di aggiornale la simulazione ad ogni frame. vengono controllati gli spostamenti, le accelerazioni e le collisioni delle particelle
func (s *Simulation) Update(frametime float32) {
	s.handleCollisions(frametime)
}

func (s *Simulation) handleCollisions(frametime float32) {
	// Applica la gravità a tutte le unità prima di gestire le collisioni
	for i := range s.Fluid.Units {
		unit := &s.Fluid.Units[i]
		unit.Velocity = rl.Vector2Add(unit.Velocity, rl.Vector2Scale(gravity, frametime))
	}

	// Gestione delle collisioni tra unità
	for i := 0; i < len(s.Fluid.Units); i++ {
		for j := i + 1; j < len(s.Fluid.Units); j++ {
			a, b := &s.Fluid.Units[i], &s.Fluid.Units[j]
			if collisionTime, collided := findCollisionTime(a, b, frametime); collided {
				s.IsPause = true
				resolveCollision(a, b, collisionTime)
			}
		}
	}

	// Gestione delle collisioni con i muri con CCD
	for i := range s.Fluid.Units {
		unit := &s.Fluid.Units[i]
		wallCollisionTime, wallNormal := findWallCollisionTime(unit, frametime, s.Config.GameWidth, s.Config.WindowHeight)
		if wallCollisionTime < frametime {
			resolveWallCollision(unit, wallCollisionTime, wallNormal)
		}
	}

	// Aggiorna la posizione delle unità come ultima cosa
	for i := range s.Fluid.Units {
		unit := &s.Fluid.Units[i]
		unit.Position = rl.Vector2Add(unit.Position, rl.Vector2Scale(unit.Velocity, frametime))
	}
}

func findWallCollisionTime(unit *Unit, frametime float32, width, height int32) (float32, rl.Vector2) {
	times := []float32{
		(unit.Radius - unit.Position.X) / unit.Velocity.X,
		(float32(width) - unit.Radius - unit.Position.X) / unit.Velocity.X,
		(unit.Radius - unit.Position.Y) / unit.Velocity.Y,
		(float32(height) - unit.Radius - unit.Position.Y) / unit.Velocity.Y,
	}
	normals := []rl.Vector2{
		{X: 1, Y: 0},
		{X: -1, Y: 0},
		{X: 0, Y: 1},
		{X: 0, Y: -1},
	}

	// Trova il primo tempo di impatto e la normale corrispondente
	earliestTime := frametime
	var earliestNormal rl.Vector2
	for i, t := range times {
		if 0 <= t && t <= frametime && t < earliestTime {
			earliestTime = t
			earliestNormal = normals[i]
		}
	}

	return earliestTime, earliestNormal
}

func resolveWallCollision(unit *Unit, collisionTime float32, normal rl.Vector2) {
	// Aggiorna la posizione al tempo di collisione
	unit.Position = rl.Vector2Add(unit.Position, rl.Vector2Scale(unit.Velocity, collisionTime))

	// Calcola la velocità relativa lungo la normale della collisione
	velocityAlongNormal := rl.Vector2DotProduct(unit.Velocity, normal)

	// Se la particella si sta allontanando dal muro, non c'è bisogno di risolvere la collisione
	if velocityAlongNormal > 0 {
		return
	}

	// Inverte la componente della velocità lungo la normale
	impulse := rl.Vector2Scale(normal, -2*velocityAlongNormal)
	unit.Velocity = rl.Vector2Add(unit.Velocity, impulse)
}

func findCollisionTime(a, b *Unit, frametime float32) (float32, bool) {
	// Calcola i vettori relativi
	relativePosition := rl.Vector2Subtract(b.Position, a.Position)
	relativeVelocity := rl.Vector2Subtract(b.Velocity, a.Velocity)

	// Calcola i coefficienti dell'equazione quadratica
	aCoeff := relativeVelocity.X*relativeVelocity.X + relativeVelocity.Y*relativeVelocity.Y
	bCoeff := 2 * (relativeVelocity.X*relativePosition.X + relativeVelocity.Y*relativePosition.Y)
	cCoeff := relativePosition.X*relativePosition.X + relativePosition.Y*relativePosition.Y - (a.Radius+b.Radius)*(a.Radius+b.Radius)

	// Calcola il discriminante
	discriminant := bCoeff*bCoeff - 4*aCoeff*cCoeff
	if discriminant < 0 {
		// Nessuna collisione
		return 0, false
	}

	// Trova le radici dell'equazione quadratica
	t1 := (-bCoeff - float32(math.Sqrt(float64(discriminant)))) / (2 * aCoeff)
	t2 := (-bCoeff + float32(math.Sqrt(float64(discriminant)))) / (2 * aCoeff)

	// Scegli la radice che rappresenta il tempo di collisione più presto nell'intervallo [0, frametime]
	if 0 <= t1 && t1 <= frametime {
		return t1, true
	} else if 0 <= t2 && t2 <= frametime {
		return t2, true
	}

	return 0, false
}

func resolveCollision(a, b *Unit, collisionTime float32) {
	// Aggiorna la posizione delle particelle al tempo di collisione
	a.Position = rl.Vector2Add(a.Position, rl.Vector2Scale(a.Velocity, collisionTime))
	b.Position = rl.Vector2Add(b.Position, rl.Vector2Scale(b.Velocity, collisionTime))

	// Calcola la normale della collisione
	collisionNormal := rl.Vector2Normalize(rl.Vector2Subtract(b.Position, a.Position))

	// Calcola la velocità relativa lungo la normale della collisione
	relativeVelocity := rl.Vector2Subtract(b.Velocity, a.Velocity)
	velocityAlongNormal := rl.Vector2DotProduct(relativeVelocity, collisionNormal)

	// Se le particelle si stanno allontanando l'una dall'altra, non c'è bisogno di risolvere la collisione
	if velocityAlongNormal > 0 {
		return
	}

	// Calcola l'impulso di risposta
	j := -2 * velocityAlongNormal / (1/a.Mass + 1/b.Mass)

	// Applica l'impulso alle particelle
	impulse := rl.Vector2Scale(collisionNormal, j)
	a.Velocity = rl.Vector2Add(a.Velocity, rl.Vector2Scale(impulse, 1/a.Mass))
	b.Velocity = rl.Vector2Subtract(b.Velocity, rl.Vector2Scale(impulse, 1/b.Mass))
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
