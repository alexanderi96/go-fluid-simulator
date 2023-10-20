package physics

import (
	"image/color"
	"math"
	"time"

	"github.com/alexanderi96/go-fluid-simulator/config"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

type Unit struct {
	Id               uuid.UUID
	Position         rl.Vector2
	PreviousPosition rl.Vector2
	Velocity         rl.Vector2
	Acceleration     rl.Vector2
	Elasticity       float32
	Radius           float32
	Mass             float32
	Color            color.RGBA
}

func newUnitsAtPosition(spawnPosition rl.Vector2, simulationWidth, simulationHeight, unitNumber int32, unitRadius, unitMass, initialSpacing, scaleFactor, elasticity float32) *[]*Unit {
	units := make([]*Unit, 0, unitNumber) // Pre-allocazione
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
				Id:               uuid.New(),
				Position:         rl.Vector2{X: x, Y: y},
				PreviousPosition: rl.Vector2{X: x, Y: y},
				Velocity:         rl.Vector2{X: 200, Y: 0},
				Radius:           unitRadius,
				Mass:             unitMass,
				Elasticity:       elasticity,
				Color:            color.RGBA{255, 0, 0, 255}, // Red color for illustration
			}

			units = append(units, &unit)
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

func calculateInitialVelocity(position rl.Vector2, simulationWidth, simulationHeight int32) rl.Vector2 {
	const maxSpeed float32 = 1000.0 // Definisci la velocità massima come desideri

	// Calcola le distanze dai muri
	d_left := position.X
	d_right := float32(simulationWidth) - position.X
	d_top := position.Y
	d_bottom := float32(simulationHeight) - position.Y

	// Determina la velocità basata sulla distanza dai muri
	var velocityX, velocityY float32

	// Calcola le componenti delle velocità in base alle distanze dai bordi
	// E sommale per ottenere la velocità complessiva su quell'asse
	velocityX = maxSpeed*(d_right/float32(simulationWidth)) - maxSpeed*(d_left/float32(simulationWidth))
	velocityY = maxSpeed*(d_bottom/float32(simulationHeight)) - maxSpeed*(d_top/float32(simulationHeight))

	return rl.Vector2{X: velocityX, Y: velocityY}
}
func spawnUnitsWithVelocity(units *[]*Unit, spawnPosition rl.Vector2, simulationWidth, simulationHeight, unitNumber int32, unitRadius, unitMass, initialSpacing, scaleFactor, elasticity float32) {
	var lastSpawned *Unit = nil

	for i := 0; i < int(unitNumber); i++ {
		// Se c'è una unità spawnata precedentemente, aspetta che si sia allontanata abbastanza
		for lastSpawned != nil && distanceBetween(lastSpawned.Position, spawnPosition) < 2*unitRadius {
			time.Sleep(100 * time.Millisecond) // attendi brevemente prima di controllare di nuovo
		}

		velocity := calculateInitialVelocity(spawnPosition, simulationWidth, simulationHeight)

		previousPosition := rl.Vector2{
			X: spawnPosition.X - velocity.X/float32(rl.GetFPS()),
			Y: spawnPosition.Y - velocity.Y/float32(rl.GetFPS()),
		}
		unit := Unit{
			Id:               uuid.New(),
			Position:         spawnPosition, // Utilizza direttamente la posizione di spawn
			PreviousPosition: previousPosition,
			Velocity:         velocity, // Utilizza la velocità calcolata
			Radius:           unitRadius,
			Mass:             unitMass,
			Elasticity:       elasticity,
			Color:            color.RGBA{255, 0, 0, 255},
		}
		*units = append(*units, &unit)
		lastSpawned = &unit
	}
}

// Calcola il volume di una singola unità
func (u *Unit) volume() float32 {
	return math.Pi * u.Radius * u.Radius
}

func (u *Unit) accelerate(a rl.Vector2) {
	u.Acceleration.X += a.X
	u.Acceleration.Y += a.Y
}

func (u *Unit) updateVelocity(dt float32) {
	u.Velocity.X += u.Acceleration.X * dt
	u.Velocity.Y += u.Acceleration.Y * dt
}

func (u *Unit) updatePosition(dt float32) {
	u.Position.X += u.Velocity.X * dt
	u.Position.Y += u.Velocity.Y * dt

}

func (u *Unit) GetVelocityWithVerlet() rl.Vector2 {
	return rl.Vector2{
		X: u.Position.X - u.PreviousPosition.X,
		Y: u.Position.Y - u.PreviousPosition.Y,
	}
}

func (u *Unit) updatePositionWithVerlet(dt float32) {
	newPosition := rl.Vector2{}
	newPosition.X = 2*u.Position.X - u.PreviousPosition.X + u.Acceleration.X*dt*dt
	newPosition.Y = 2*u.Position.Y - u.PreviousPosition.Y + u.Acceleration.Y*dt*dt

	u.PreviousPosition = u.Position
	u.Position = newPosition
	u.Acceleration = rl.Vector2{X: 0, Y: 0}
}

func (u *Unit) checkWallCollision(cfg *config.Config) {
	// Controlla e corregge la posizione X
	if u.Position.X-u.Radius < 0 {
		u.Position.X = u.Radius
		u.Velocity.X = -u.Velocity.X * cfg.WallElasticity // Invertire la velocità X
	} else if u.Position.X+u.Radius > float32(cfg.GameWidth) {
		u.Position.X = float32(cfg.GameWidth) - u.Radius
		u.Velocity.X = -u.Velocity.X * cfg.WallElasticity // Invertire la velocità X
	}

	// Controlla e corregge la posizione Y
	if u.Position.Y-u.Radius < 0 {
		u.Position.Y = u.Radius
		u.Velocity.Y = -u.Velocity.Y * cfg.WallElasticity // Invertire la velocità Y
	} else if u.Position.Y+u.Radius > float32(cfg.WindowHeight) {
		u.Position.Y = float32(cfg.WindowHeight) - u.Radius
		u.Velocity.Y = -u.Velocity.Y * cfg.WallElasticity // Invertire la velocità Y
	}
}

func (u *Unit) checkWallCollisionVerlet(cfg *config.Config, deltaTime float32) {
	// Store old position for Verlet integration
	oldPosition := u.Position

	// Update position using Verlet integration
	// new_position = 2 * current_position - old_position + acceleration * dt^2
	u.Position.X = 2*u.Position.X - oldPosition.X + u.Acceleration.X*deltaTime*deltaTime
	u.Position.Y = 2*u.Position.Y - oldPosition.Y + u.Acceleration.Y*deltaTime*deltaTime

	// Check and correct X position
	if u.Position.X-u.Radius < 0 {
		u.Position.X = u.Radius
		// Compute reflected position for Verlet
		oldPosition.X = u.Position.X + (u.Position.X-oldPosition.X)*cfg.WallElasticity
	} else if u.Position.X+u.Radius > float32(cfg.GameWidth) {
		u.Position.X = float32(cfg.GameWidth) - u.Radius
		// Compute reflected position for Verlet
		oldPosition.X = u.Position.X + (u.Position.X-oldPosition.X)*cfg.WallElasticity
	}

	// Check and correct Y position
	if u.Position.Y-u.Radius < 0 {
		u.Position.Y = u.Radius
		// Compute reflected position for Verlet
		oldPosition.Y = u.Position.Y + (u.Position.Y-oldPosition.Y)*cfg.WallElasticity
	} else if u.Position.Y+u.Radius > float32(cfg.WindowHeight) {
		u.Position.Y = float32(cfg.WindowHeight) - u.Radius
		// Compute reflected position for Verlet
		oldPosition.Y = u.Position.Y + (u.Position.Y-oldPosition.Y)*cfg.WallElasticity
	}

	// Update old position after collision response
	u.PreviousPosition = oldPosition
}

func distanceBetween(p1, p2 rl.Vector2) float32 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}

func areOverlapping(a, b *Unit) bool {
	deltaX := b.Position.X - a.Position.X
	deltaY := b.Position.Y - a.Position.Y
	distanceSquared := deltaX*deltaX + deltaY*deltaY
	totalRadius := a.Radius + b.Radius
	return distanceSquared < totalRadius*totalRadius
}
