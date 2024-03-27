package physics

import (
	"math"

	"github.com/alexanderi96/go-fluid-simulator/config"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Simulation) UpdateWithVerletIntegration() error {
	// Calcola il numero di step
	resolutionSteps := 3
	fractionalFrametime := s.Metrics.Frametime / float32(resolutionSteps)

	for step := 0; step < resolutionSteps; step++ {
		for _, unit := range s.Fluid {
			unit.updatePositionWithVerlet(fractionalFrametime)
		}
		for _, unit := range s.Fluid {
			if s.Config.ApplyGravity {
				unit.accelerate(rl.Vector3{X: 0, Y: s.Config.Gravity, Z: 0})
			}
		}
		for _, unit := range s.Fluid {
			unit.checkWallCollisionVerlet(s.Config, fractionalFrametime)

		}
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
					applyGravitationalAttraction(unitA, unitB, s.Config.UnitGravitationalMultiplier, s.Config.UnitInitialSpacing)
				}
			}
		}
	}

	return nil
}

func applyGravitationalAttraction(a, b *Unit, G, spacing float32) {
	dx := b.Position.X - a.Position.X
	dy := b.Position.Y - a.Position.Y
	dz := b.Position.Z - a.Position.Z
	distanceSquared := dx*dx + dy*dy + dz*dz
	distance := float32(math.Sqrt(float64(distanceSquared)))

	// Calcola il raggio totale delle due unità
	totalRadius := a.Radius + b.Radius

	// Determina se le unità sono sovrapposte
	// areOverlapping := distanceSquared < totalRadius*totalRadius

	// Evita la divisione per zero e le forze estremamente forti a distanze molto piccole
	if distance <= 0 {
		return
	}

	forceMagnitude := G * (a.Mass() * b.Mass()) / distanceSquared

	// Se le unità sono sovrapposte, inverte la direzione della forza e moltiplica la magnitudine per 10
	if distance < totalRadius+spacing {
		//forceMagnitude = -forceMagnitude * 2
		// Ottieni la massa media delle due unità
		averageMass := (a.Mass() + b.Mass()) / 2

		// Modifica la magnitudine della forza in base al reciproco della massa media
		forceMagnitude *= -1 / averageMass // Aggiungi 1 per evitare la divisione per zero
	}

	forceX := forceMagnitude * (dx / distance)
	forceY := forceMagnitude * (dy / distance)
	forceZ := forceMagnitude * (dz / distance)

	a.Acceleration.X += forceX / a.Mass()
	a.Acceleration.Y += forceY / a.Mass()
	a.Acceleration.Z += forceZ / a.Mass()
	b.Acceleration.X -= forceX / b.Mass()
	b.Acceleration.Y -= forceY / b.Mass()
	b.Acceleration.Z -= forceZ / b.Mass()
}

func areOverlapping(a, b *Unit) bool {
	deltaX := b.Position.X - a.Position.X
	deltaY := b.Position.Y - a.Position.Y
	deltaZ := b.Position.Z - a.Position.Z
	distanceSquared := deltaX*deltaX + deltaY*deltaY + deltaZ*deltaZ
	totalRadius := a.Radius + b.Radius
	return distanceSquared < totalRadius*totalRadius
}

func handleCollision(a, b *Unit, dt float32) {
	dx := b.Position.X - a.Position.X
	dy := b.Position.Y - a.Position.Y
	dz := b.Position.Z - a.Position.Z
	distSq := dx*dx + dy*dy + dz*dz

	// Se la distanza è zero, evita la divisione per zero
	if distSq == 0 {
		return
	}

	dist := float32(math.Sqrt(float64(distSq)))

	// Calcola la normale della collisione
	normalX := dx / dist
	normalY := dy / dist
	normalZ := dz / dist

	// Calcola la sovrapposizione
	overlap := (a.Radius + b.Radius) - dist

	// Calcola la correzione necessaria per risolvere la sovrapposizione
	// la correzione viene divisa tra le due unità in base alla loro massa
	inverseTotalMass := (1 / a.Mass()) + (1 / b.Mass())
	correction := overlap / inverseTotalMass

	a.Position.X -= normalX * correction * (1 / a.Mass())
	a.Position.Y -= normalY * correction * (1 / a.Mass())
	a.Position.Z -= normalZ * correction * (1 / a.Mass())
	b.Position.X += normalX * correction * (1 / b.Mass())
	b.Position.Y += normalY * correction * (1 / b.Mass())
	b.Position.Z += normalZ * correction * (1 / b.Mass())
}

func (u *Unit) updatePositionWithVerlet(dt float32) {
	newPosition := rl.Vector3{}
	newPosition.X = 2*u.Position.X - u.PreviousPosition.X + u.Acceleration.X*dt*dt
	newPosition.Y = 2*u.Position.Y - u.PreviousPosition.Y + u.Acceleration.Y*dt*dt
	newPosition.Z = 2*u.Position.Z - u.PreviousPosition.Z + u.Acceleration.Z*dt*dt

	u.PreviousPosition = u.Position
	u.Position = newPosition
	u.Acceleration = rl.Vector3{X: 0, Y: 0, Z: 0}
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

	// Correzione asse Y
	if u.Position.Z-u.Radius < 0 {
		overlapZ := u.Radius - u.Position.Z
		u.Position.Z += overlapZ
		// Applica la restituzione
		velocity.Z = -velocity.Z * cfg.WallElasticity
		u.PreviousPosition.Z = u.Position.Z - velocity.Z*deltaTime
	} else if u.Position.Z+u.Radius > float32(cfg.GameZ) {
		overlapZ := (u.Position.Z + u.Radius) - float32(cfg.GameZ)
		u.Position.Z -= overlapZ
		// Applica la restituzione
		velocity.Z = -velocity.Z * cfg.WallElasticity
		u.PreviousPosition.Z = u.Position.Z - velocity.Z*deltaTime
	}
}

func distanceBetween(p1, p2 rl.Vector3) float32 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	dz := p2.Z - p1.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}
