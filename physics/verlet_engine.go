package physics

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Simulation) UpdateWithVerletIntegration() error {

	// // Aggiorna la velocità di ciascuna particella in base alla sua forza e alla sua massa
	// for _, unit := range s.Fluid {
	// 	unit.updateVelocity(s.Metrics.Frametime)
	// }

	// Aggiorna la posizione di ciascuna particella in base alla sua velocità
	for _, unit := range s.Fluid {
		if s.Config.ApplyGravity {
			unit.accelerate(rl.Vector2{X: 0, Y: s.Config.Gravity})
		}
		unit.updatePositionWithVerlet(s.Metrics.Frametime)
		unit.checkWallCollisionVerlet(s.Config, s.Metrics.Frametime)

	}

	// Verifica se ciascuna particella sta collidendo con un'altra particella
	for _, unitA := range s.Fluid {
		if unitA == nil {
			continue
		}

		for _, unitB := range s.Fluid {
			if unitB == nil {
				continue
			}

			if unitA.Id != unitB.Id && areOverlapping(unitA, unitB) {
				// Risolvi la collisione tra le due particelle
				calculateCollisionWithVerlet(unitA, unitB)
			}
		}
	}

	return nil

}
func calculateCollisionWithVerlet(unitA, unitB *Unit) {
	// Differenza di posizione tra le particelle
	deltaX := unitB.Position.X - unitA.Position.X
	deltaY := unitB.Position.Y - unitA.Position.Y

	// Distanza e sovrapposizione
	distance := float32(math.Sqrt(float64(deltaX*deltaX + deltaY*deltaY)))
	overlap := unitA.Radius + unitB.Radius - distance

	if overlap <= 0 {
		return
	}

	// Normale della collisione
	normalX := deltaX / distance
	normalY := deltaY / distance

	// Calcolo della correzione basata sulla sovrapposizione
	correctionX := (overlap / 2) * normalX
	correctionY := (overlap / 2) * normalY

	// Applicazione della correzione alle posizioni delle particelle
	unitA.Position.X -= correctionX
	unitA.Position.Y -= correctionY
	unitB.Position.X += correctionX
	unitB.Position.Y += correctionY

	// Calcolo della velocità post-collisione basata sull'impulso
	relativeVelocityX := unitB.Velocity.X - unitA.Velocity.X
	relativeVelocityY := unitB.Velocity.Y - unitA.Velocity.Y
	impulse := -(1.0 + (unitA.Elasticity+unitB.Elasticity)/2) * (relativeVelocityX*normalX + relativeVelocityY*normalY) / (1/unitA.Mass + 1/unitB.Mass)

	impulseX := impulse * normalX
	impulseY := impulse * normalY

	unitA.Velocity.X -= impulseX / unitA.Mass
	unitA.Velocity.Y -= impulseY / unitA.Mass
	unitB.Velocity.X += impulseX / unitB.Mass
	unitB.Velocity.Y += impulseY / unitB.Mass
}
