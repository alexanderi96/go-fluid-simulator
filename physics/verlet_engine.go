package physics

import (
	"math"

	"github.com/alexanderi96/go-fluid-simulator/config"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Simulation) UpdateWithOctrees() error {
	s.Octree.Clear() // Pulisce il Octree all'inizio di ogni frame

	// Costruisci l'Octree
	for _, unit := range s.Fluid {
		s.Octree.Insert(unit)
	}

	// Controlla le collisioni tra particelle e aggiorna le velocità utilizzando il Octree
	for _, unitA := range s.Fluid {
		if unitA == nil {
			continue
		}

		updatePositionWithVerlet(unitA, s.Metrics.Frametime)

		if s.Config.ApplyGravity {
			unitA.accelerate(rl.Vector3{X: 0, Y: s.Config.Gravity, Z: 0})
		}

		checkWallCollisionVerlet(unitA, s.WorldBoundray, s.Config.WallElasticity, s.Metrics.Frametime)

		nearUnits := []*Unit{}

		s.Octree.Retrieve(&nearUnits, unitA)
		for _, unitB := range nearUnits {
			if unitB != nil && unitA.Id != unitB.Id {
				distance := getDistance(unitA, unitB)
				if distance < unitA.Radius+unitB.Radius {
					handleCollision(unitA, unitB, distance)
				}
			}
		}
	}

	return nil
}

func (s *Simulation) UpdateWithVerletIntegration() error {
	// Calcola il numero di step
	resolutionSteps := int(s.Config.ResolutionSteps)
	fractionalFrametime := s.Metrics.Frametime / float32(resolutionSteps)

	for step := 0; step < resolutionSteps; step++ {

		for _, unitA := range s.Fluid {
			if unitA == nil {
				continue
			}

			if s.Config.ApplyGravity {
				unitA.accelerate(rl.Vector3{X: 0, Y: s.Config.Gravity, Z: 0})
			}

			checkWallCollisionVerlet(unitA, s.WorldBoundray, s.Config.WallElasticity, fractionalFrametime)

			for _, unitB := range s.Fluid {
				if unitB == nil || unitA.Id == unitB.Id {
					continue
				}

				distance := getDistance(unitA, unitB)

				if s.Config.UnitsEmitGravity {
					applyGravitationalAttraction(unitA, unitB, s.Config)
				}

				if distance < unitA.Radius+unitB.Radius {
					handleCollision(unitA, unitB, distance)
				}
			}
			updatePositionWithVerlet(unitA, fractionalFrametime)

		}
	}

	return nil
}

func getDistance(unitA, unitB *Unit) float32 {
	return rl.Vector3Distance(unitA.Position, unitB.Position)
}

// func getSurfaceDistance(unitA, unitB *Unit) float32 {
// 	return getDistance(unitA, unitB) - (unitA.Radius + unitB.Radius)
// }

func applyGravitationalAttraction(a, b *Unit, config *config.Config) {
	dx := b.Position.X - a.Position.X
	dy := b.Position.Y - a.Position.Y
	dz := b.Position.Z - a.Position.Z
	distanceSquared := dx*dx + dy*dy + dz*dz
	distance := float32(math.Sqrt(float64(distanceSquared)))

	// Calcola il raggio totale delle due unità
	totalRadius := a.Radius + b.Radius

	// Evita la divisione per zero e le forze estremamente forti a distanze molto piccole
	if distance+totalRadius <= 0 {
		return
	}

	forceMagnitude := config.UnitGravitationalMultiplier * (a.Mass * b.Mass) / distanceSquared

	forceX := forceMagnitude * (dx / distance)
	forceY := forceMagnitude * (dy / distance)
	forceZ := forceMagnitude * (dz / distance)

	a.Acceleration.X += forceX / a.Mass
	a.Acceleration.Y += forceY / a.Mass
	a.Acceleration.Z += forceZ / a.Mass
	b.Acceleration.X -= forceX / b.Mass
	b.Acceleration.Y -= forceY / b.Mass
	b.Acceleration.Z -= forceZ / b.Mass
}

// func areOverlapping(a, b *Unit) bool {
// 	return getSurfaceDistance(a, b) <= 0
// }

func handleCollision(a, b *Unit, distance float32) {
	dx := b.Position.X - a.Position.X
	dy := b.Position.Y - a.Position.Y
	dz := b.Position.Z - a.Position.Z

	normalX := dx / distance
	normalY := dy / distance
	normalZ := dz / distance

	overlap := (a.Radius + b.Radius) - distance
	totalMass := a.Mass + b.Mass
	correction := overlap / totalMass

	// Applica la correzione alle posizioni delle unità
	a.Position.X -= normalX * correction * a.Mass
	a.Position.Y -= normalY * correction * a.Mass
	a.Position.Z -= normalZ * correction * a.Mass
	b.Position.X += normalX * correction * b.Mass
	b.Position.Y += normalY * correction * b.Mass
	b.Position.Z += normalZ * correction * b.Mass
}

func updatePositionWithVerlet(u *Unit, dt float32) {
	newPosition := rl.Vector3{}
	newPosition.X = 2*u.Position.X - u.PreviousPosition.X + u.Acceleration.X*dt*dt
	newPosition.Y = 2*u.Position.Y - u.PreviousPosition.Y + u.Acceleration.Y*dt*dt
	newPosition.Z = 2*u.Position.Z - u.PreviousPosition.Z + u.Acceleration.Z*dt*dt

	u.PreviousPosition = u.Position
	u.Position = newPosition
	u.Acceleration = rl.Vector3{X: 0, Y: 0, Z: 0}
}

func checkWallCollisionVerlet(u *Unit, boundrais rl.BoundingBox, wallElasticity, deltaTime float32) {
	// Calcola la velocità
	velocity := u.GetVelocity(deltaTime)

	// Correzione asse X
	if u.Position.X-u.Radius < boundrais.Min.X {
		overlapX := u.Radius - u.Position.X + boundrais.Min.X
		u.Position.X += overlapX
		// Applica la restituzione
		velocity.X = -velocity.X * wallElasticity
		u.PreviousPosition.X = u.Position.X - velocity.X*deltaTime
	} else if u.Position.X+u.Radius > boundrais.Max.X {
		overlapX := u.Position.X + u.Radius - boundrais.Max.X
		u.Position.X -= overlapX
		// Applica la restituzione
		velocity.X = -velocity.X * wallElasticity
		u.PreviousPosition.X = u.Position.X - velocity.X*deltaTime
	}

	// Correzione asse Y
	if u.Position.Y-u.Radius < boundrais.Min.Y {
		overlapY := u.Radius - u.Position.Y + boundrais.Min.Y
		u.Position.Y += overlapY
		// Applica la restituzione
		velocity.Y = -velocity.Y * wallElasticity
		u.PreviousPosition.Y = u.Position.Y - velocity.Y*deltaTime
	} else if u.Position.Y+u.Radius > boundrais.Max.Y {
		overlapY := (u.Position.Y + u.Radius) - boundrais.Max.Y
		u.Position.Y -= overlapY
		// Applica la restituzione
		velocity.Y = -velocity.Y * wallElasticity
		u.PreviousPosition.Y = u.Position.Y - velocity.Y*deltaTime
	}

	// Correzione asse Y
	if u.Position.Z-u.Radius < boundrais.Min.Z {
		overlapZ := u.Radius - u.Position.Z + boundrais.Min.Z
		u.Position.Z += overlapZ
		// Applica la restituzione
		velocity.Z = -velocity.Z * wallElasticity
		u.PreviousPosition.Z = u.Position.Z - velocity.Z*deltaTime
	} else if u.Position.Z+u.Radius > boundrais.Max.Z {
		overlapZ := (u.Position.Z + u.Radius) - boundrais.Max.Z
		u.Position.Z -= overlapZ
		// Applica la restituzione
		velocity.Z = -velocity.Z * wallElasticity
		u.PreviousPosition.Z = u.Position.Z - velocity.Z*deltaTime
	}
}
