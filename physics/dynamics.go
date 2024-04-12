package physics

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const G = 6.67430e-11 // Costante gravitazionale universale (m^3 kg^-1 s^-2)

func (s *Simulation) UpdateWithOctrees() error {
	frameTime := s.Metrics.Frametime

	s.Octree.Clear() // Pulisce il Octree all'inizio di ogni frame

	// Costruisci l'Octree
	for _, unit := range s.Fluid {
		s.Octree.Insert(unit)
	}

	for _, unit := range s.Fluid {
		if s.Config.UnitsEmitGravity {
			unit.accelerate(s.Octree.CalculateGravity(unit, 5))
		}

		// Calcola la forza di gravità solo se abilitato
		if s.Config.ApplyGravity {
			// Aggiorna l'accelerazione con la forza di gravità
			unit.accelerate(rl.Vector3{X: 0, Y: s.Config.Gravity, Z: 0})
		}
	}

	// Controlla le collisioni tra particelle e aggiorna le velocità utilizzando il Octree
	for _, unitA := range s.Fluid {
		if unitA == nil {
			continue
		}

		updatePositionWithVerlet(unitA, frameTime)

		// Aggiorna la posizione utilizzando il metodo Verlet

		// Gestisci le collisioni con le pareti
		checkWallCollisionVerlet(unitA, s.WorldBoundray, s.Config.WallElasticity, frameTime)

		// Trova le particelle vicine utilizzando l'Octree
		nearUnits := []*Unit{}
		s.Octree.Retrieve(&nearUnits, unitA)

		// Gestisci le collisioni tra particelle vicine
		for _, unitB := range nearUnits {
			if unitB != nil && unitA.Id != unitB.Id {
				distance := getDistance(unitA, unitB)
				if distance <= unitA.Radius+unitB.Radius {
					handleCollision(unitA, unitB, frameTime)
				}
			}
		}
	}

	return nil
}

func (ot *Octree) CalculateGravity(unit *Unit, theta float32) rl.Vector3 {
	var force rl.Vector3
	ot.calculateGravityRecursive(unit, theta, &force)
	return force
}

func (ot *Octree) calculateGravityRecursive(unit *Unit, theta float32, force *rl.Vector3) {
	if ot.Children[0] == nil {
		// Se siamo in un nodo foglia, calcoliamo la forza tra tutti gli oggetti e la unit.
		for _, obj := range ot.objects {
			if obj != unit {
				deltaPos := rl.Vector3Subtract(obj.Position, unit.Position)
				distance := rl.Vector3Length(deltaPos)
				if distance > 0 {
					// Calcolo della forza gravitazionale.
					magnitude := G * unit.Mass * obj.Mass / (distance * distance)
					direction := rl.Vector3Normalize(deltaPos)
					forceToAdd := rl.Vector3Scale(direction, magnitude)
					*force = rl.Vector3Add(*force, forceToAdd)
				}
			}
		}
	} else {
		// Se non siamo in un nodo foglia, decidiamo se calcolare la forza con il centro di massa o scendere nell'Octree.
		width := ot.Bounds.Max.X - ot.Bounds.Min.X
		distance := rl.Vector3Distance(unit.Position, ot.CenterOfMass)
		if (width / distance) < theta {
			// Usiamo il centro di massa per approssimare la forza.
			deltaPos := rl.Vector3Subtract(ot.CenterOfMass, unit.Position)
			distance := rl.Vector3Length(deltaPos)
			if distance > 0 {
				magnitude := G * unit.Mass * ot.TotalMass / (distance * distance)
				direction := rl.Vector3Normalize(deltaPos)
				forceToAdd := rl.Vector3Scale(direction, magnitude)
				*force = rl.Vector3Add(*force, forceToAdd)
			}
		} else {
			// Altrimenti, calcoliamo la forza ricorsivamente sui figli dell'Octree.
			for _, child := range ot.Children {
				if child != nil {
					child.calculateGravityRecursive(unit, theta, force)
				}
			}
		}
	}
}

func getDistance(unitA, unitB *Unit) float32 {
	return rl.Vector3Distance(unitA.Position, unitB.Position)
}

func handleCollision(a, b *Unit, dt float32) {
	// Calcola la distanza tra le unità
	distance := rl.Vector3Distance(a.Position, b.Position)
	minSeparation := 2 * a.Radius

	// Se le unità non sono sufficientemente vicine, non c'è collisione
	if distance > minSeparation {
		return
	}

	// Calcola la direzione della collisione normalizzata
	collisionDir := rl.Vector3Normalize(rl.Vector3Subtract(a.Position, b.Position))

	// Calcola la distanza di sovrapposizione (MTD)
	mtd := rl.Vector3Scale(collisionDir, minSeparation-distance)

	// Aggiorna le posizioni per risolvere la collisione
	a.Position = rl.Vector3Add(a.Position, rl.Vector3Scale(mtd, 0.5))
	b.Position = rl.Vector3Subtract(b.Position, rl.Vector3Scale(mtd, 0.5))

	// Calcola la velocità relativa
	relativeVelocity := rl.Vector3Subtract(a.GetVelocity(dt), b.GetVelocity(dt))

	// Calcola la velocità di impatto
	impactSpeed := rl.Vector3DotProduct(relativeVelocity, collisionDir)

	// Se le unità si stanno allontanando, non c'è collisione da risolvere
	if impactSpeed > 0 {
		return
	}

	// Calcola la massa combinata
	totalMass := a.Mass + b.Mass

	// Calcola la magnitudine dell'impulso
	impulseMagnitude := (1 + a.Elasticity) * rl.Vector3DotProduct(relativeVelocity, collisionDir) / totalMass

	// Calcola e applica l'impulso alle velocità delle unità
	impulse := rl.Vector3Scale(collisionDir, impulseMagnitude)

	a.Acceleration = rl.Vector3Add(a.Acceleration, rl.Vector3Scale(impulse, 1.0/a.Mass))
	b.Acceleration = rl.Vector3Add(b.Acceleration, rl.Vector3Scale(impulse, -1.0/b.Mass))
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
