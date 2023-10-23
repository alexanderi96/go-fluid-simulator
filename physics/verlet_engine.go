package physics

import (
	"math"

	"github.com/alexanderi96/go-fluid-simulator/config"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Simulation) UpdateWithVerletIntegration() error {
	// Calcola il numero di step di risoluzione in base al frametime
	resolutionSteps := 1 //int(math.Max(1, math.Min(10, float64(1/s.Metrics.Frametime))))
	//log.Println("Resolution steps:", resolutionSteps)
	fractionalFrametime := s.Metrics.Frametime / float32(resolutionSteps)

	for _, unit := range s.Fluid {
		unit.updatePositionWithVerlet(fractionalFrametime)
	}

	for _, unit := range s.Fluid {
		if s.Config.ApplyGravity {
			unit.accelerate(rl.Vector2{X: 0, Y: s.Config.Gravity})
		}
	}

	for _, unit := range s.Fluid {
		unit.checkWallCollisionVerlet(s.Config, fractionalFrametime)

	}

	for step := 0; step < resolutionSteps; step++ {
		for _, unitA := range s.Fluid {
			if unitA == nil {
				continue
			}

			for _, unitB := range s.Fluid {
				if unitB == nil {
					continue
				}

				if unitA.Id != unitB.Id && areOverlapping(unitA, unitB) {
					handleCollision(unitA, unitB, fractionalFrametime)
				}

				if s.Config.UnitsEmitGravity {
					applyGravitationalAttraction(unitA, unitB, s.Config.UnitGravitationalMultiplier)
				}
			}
		}

	}

	return nil
}

func applyGravitationalAttraction(a, b *Unit, G float32) {
	dx := b.Position.X - a.Position.X
	dy := b.Position.Y - a.Position.Y
	distanceSquared := dx*dx + dy*dy
	distance := float32(math.Sqrt(float64(distanceSquared)))

	// Calcola il raggio totale delle due unità
	totalRadius := a.Radius + b.Radius

	// Determina se le unità sono sovrapposte
	// areOverlapping := distanceSquared < totalRadius*totalRadius

	// Evita la divisione per zero e le forze estremamente forti a distanze molto piccole
	if distance == 0 {
		return
	}

	forceMagnitude := G * (a.Mass() * b.Mass()) / distanceSquared

	// Se le unità sono sovrapposte, inverte la direzione della forza e moltiplica la magnitudine per 10
	if distance < totalRadius+10 {
		//forceMagnitude = -forceMagnitude * 2
		// Ottieni la massa media delle due unità
		averageMass := (a.Mass() + b.Mass()) / 2

		// Modifica la magnitudine della forza in base al reciproco della massa media
		forceMagnitude *= -1 / averageMass // Aggiungi 1 per evitare la divisione per zero
	}
	// if distance < totalRadius {
	// 	forceMagnitude = -forceMagnitude
	// }

	forceX := forceMagnitude * (dx / distance)
	forceY := forceMagnitude * (dy / distance)

	a.Acceleration.X += forceX / a.Mass()
	a.Acceleration.Y += forceY / a.Mass()
	b.Acceleration.X -= forceX / b.Mass()
	b.Acceleration.Y -= forceY / b.Mass()
}

func areOverlapping(a, b *Unit) bool {
	deltaX := b.Position.X - a.Position.X
	deltaY := b.Position.Y - a.Position.Y
	distanceSquared := deltaX*deltaX + deltaY*deltaY
	totalRadius := a.Radius + b.Radius
	return distanceSquared < totalRadius*totalRadius
}

func handleCollision(a, b *Unit, dt float32) {
	// Calcola la normale della collisione
	dx := b.Position.X - a.Position.X
	dy := b.Position.Y - a.Position.Y
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	// Se la distanza è zero, evita la divisione per zero
	if dist == 0 {
		return
	}

	normalX := dx / dist
	normalY := dy / dist

	// Calcola la sovrapposizione
	overlap := (a.Radius + b.Radius) - dist

	// Calcola la correzione necessaria per risolvere la sovrapposizione
	// la correzione viene divisa tra le due unità in base alla loro massa
	inverseTotalMass := (1 / a.Mass()) + (1 / b.Mass())
	correctionX := overlap * (normalX / inverseTotalMass)
	correctionY := overlap * (normalY / inverseTotalMass)

	// Applica la correzione alle posizioni delle unità
	a.Position.X -= correctionX * (1 / a.Mass())
	a.Position.Y -= correctionY * (1 / a.Mass())
	b.Position.X += correctionX * (1 / b.Mass())
	b.Position.Y += correctionY * (1 / b.Mass())

	// Calcola la velocità relativa
	relativeVelocityX := (b.Position.X - b.PreviousPosition.X) - (a.Position.X - a.PreviousPosition.X)
	relativeVelocityY := (b.Position.Y - b.PreviousPosition.Y) - (a.Position.Y - a.PreviousPosition.Y)
	velocityAlongNormal := relativeVelocityX*normalX + relativeVelocityY*normalY

	// Se le unità si stanno allontanando l'una dall'altra, non gestire la collisione
	if velocityAlongNormal > 0 {
		return
	}

	// Calcola il fattore di restituzione (elasticità) come la media delle elasticità delle due unità
	e := (a.Elasticity + b.Elasticity) / 2

	// Calcola e applica l'impulso
	j := -(1 + e) * velocityAlongNormal
	j /= inverseTotalMass

	impulseX := j * normalX // Rimuovi * dt
	impulseY := j * normalY // Rimuovi * dt

	a.Position.X -= impulseX * (1 / a.Mass())
	a.Position.Y -= impulseY * (1 / a.Mass())
	b.Position.X += impulseX * (1 / b.Mass())
	b.Position.Y += impulseY * (1 / b.Mass())

}

func (u *Unit) updatePositionWithVerlet(dt float32) {
	newPosition := rl.Vector2{}
	newPosition.X = 2*u.Position.X - u.PreviousPosition.X + u.Acceleration.X*dt*dt
	newPosition.Y = 2*u.Position.Y - u.PreviousPosition.Y + u.Acceleration.Y*dt*dt

	u.PreviousPosition = u.Position
	u.Position = newPosition
	u.Acceleration = rl.Vector2{X: 0, Y: 0}
}

func (u *Unit) checkWallCollisionVerlet(cfg *config.Config, deltaTime float32) {
	// Calcola la velocità
	velocity := u.Velocity(deltaTime)

	// Correzione asse X
	if u.Position.X-u.Radius < 0 {
		overlapX := u.Radius - u.Position.X
		u.Position.X += overlapX
		// Applica la restituzione
		velocity.X = -velocity.X * cfg.WallElasticity
		u.PreviousPosition.X = u.Position.X - velocity.X*deltaTime
	} else if u.Position.X+u.Radius > float32(cfg.GameX) {
		overlapX := (u.Position.X + u.Radius) - float32(cfg.GameX)
		u.Position.X -= overlapX
		// Applica la restituzione
		velocity.X = -velocity.X * cfg.WallElasticity
		u.PreviousPosition.X = u.Position.X - velocity.X*deltaTime
	}

	// Correzione asse Y
	if u.Position.Y-u.Radius < 0 {
		overlapY := u.Radius - u.Position.Y
		u.Position.Y += overlapY
		// Applica la restituzione
		velocity.Y = -velocity.Y * cfg.WallElasticity
		u.PreviousPosition.Y = u.Position.Y - velocity.Y*deltaTime
	} else if u.Position.Y+u.Radius > float32(cfg.GameY) {
		overlapY := (u.Position.Y + u.Radius) - float32(cfg.GameY)
		u.Position.Y -= overlapY
		// Applica la restituzione
		velocity.Y = -velocity.Y * cfg.WallElasticity
		u.PreviousPosition.Y = u.Position.Y - velocity.Y*deltaTime
	}
}

func distanceBetween(p1, p2 rl.Vector2) float32 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}
