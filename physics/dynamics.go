package physics

import (
	"github.com/EliCDavis/vector/vector3"
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
			unit.accelerate(vector3.New[float64](0, s.Config.Gravity, 0))
		}
	}

	// Controlla le collisioni tra particelle e aggiorna le velocità utilizzando il Octree
	for _, unitA := range s.Fluid {
		if unitA == nil {
			continue
		}

		// Aggiorna la posizione utilizzando il metodo Verlet
		updatePositionWithVerlet(unitA, frameTime)

		// Gestisci le collisioni con le pareti
		unitA.CheckAndResolveWallCollision(s.WorldBoundray, s.Config.WallElasticity)

		// Trova le particelle vicine utilizzando l'Octree
		nearUnits := []*Unit{}
		s.Octree.Retrieve(&nearUnits, unitA)

		// Gestisci le collisioni tra particelle vicine
		for _, unitB := range nearUnits {
			if unitB != nil && unitA.Id != unitB.Id {
				distance := unitA.Position.Distance(unitB.Position)
				if distance <= unitA.Radius+unitB.Radius {
					handleCollision(unitA, unitB, distance)
				}
			}
		}
	}
	return nil
}

func (ot *Octree) CalculateGravity(unit *Unit, theta float64) vector3.Vector[float64] {
	var force = vector3.Zero[float64]()
	ot.calculateGravityRecursive(unit, theta, &force)
	return force
}

func (ot *Octree) calculateGravityRecursive(unit *Unit, theta float64, force *vector3.Vector[float64]) {
	if ot.Children[0] == nil {
		// Se siamo in un nodo foglia, calcoliamo la forza tra tutti gli oggetti e la unit.
		for _, obj := range ot.objects {
			if obj != unit {
				deltaPos := obj.Position.Sub(unit.Position)
				distance := deltaPos.Length()
				if distance > 0 {
					// Calcolo della forza gravitazionale.
					magnitude := G * unit.Mass * obj.Mass / (distance * distance)
					direction := deltaPos.Normalized()
					forceToAdd := direction.Scale(magnitude)
					*force = force.Add(forceToAdd)
				}
			}
		}
	} else {
		// Se non siamo in un nodo foglia, decidiamo se calcolare la forza con il centro di massa o scendere nell'Octree.
		width := ot.Bounds.Max.X() - ot.Bounds.Min.X()
		distance := unit.Position.Distance(ot.CenterOfMass)
		if (width / distance) < theta {
			// Usiamo il centro di massa per approssimare la forza.
			deltaPos := ot.CenterOfMass.Sub(unit.Position)
			distance := deltaPos.Length()
			if distance > 0 {
				magnitude := G * unit.Mass * ot.TotalMass / (distance * distance)
				direction := deltaPos.Normalized()
				forceToAdd := direction.Scale(magnitude)
				*force = force.Add(forceToAdd)
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

func handleCollision(a, b *Unit, distance float64) {

	// Calcola la direzione della collisione normalizzata
	collisionDir := a.Position.Sub(b.Position).Normalized()

	// Calcola la distanza di sovrapposizione (MTD)
	mtd := collisionDir.Scale(a.Radius + b.Radius - distance)

	// Aggiorna le posizioni per risolvere la collisione
	a.Position.Add(mtd.Scale(0.5))
	b.Position.Sub(mtd.Scale(0.5))

	// Calcola la velocità relativa
	relativeVelocity := a.GetVelocity().Sub(b.GetVelocity())

	// Calcola la velocità di impatto
	impactSpeed := relativeVelocity.Dot(collisionDir)

	// Se le unità si stanno allontanando, non c'è collisione da risolvere
	if impactSpeed > 0 {
		return
	}

	// Calcola la massa combinata
	totalMass := a.Mass + b.Mass

	// Calcola la magnitudine dell'impulso
	impulseMagnitude := (1 + a.Elasticity) * relativeVelocity.Dot(collisionDir) / totalMass

	// Calcola e applica l'impulso alle velocità delle unità
	impulse := collisionDir.Scale(impulseMagnitude)

	a.Acceleration.Add(impulse.Scale(1.0 / a.Mass))
	b.Acceleration.Add(impulse.Scale(-1.0 / b.Mass))
}

func updatePositionWithVerlet(u *Unit, dt float64) {
	x := 2*u.Position.X() - u.PreviousPosition.X() + u.Acceleration.X()*dt*dt
	y := 2*u.Position.Y() - u.PreviousPosition.Y() + u.Acceleration.Y()*dt*dt
	z := 2*u.Position.Z() - u.PreviousPosition.Z() + u.Acceleration.Z()*dt*dt
	newPosition := vector3.New(x, y, z)
	u.PreviousPosition = u.Position
	u.Position = newPosition
	u.Acceleration = vector3.Zero[float64]()
}

func (unit *Unit) CheckAndResolveWallCollision(wallBounds BoundingBox, wallElasticity float64) {
	velocity := unit.GetVelocity()

	// Correzione asse X
	if unit.Position.X()-unit.Radius < wallBounds.Min.X() {
		overlapX := wallBounds.Min.X() - (unit.Position.X() - unit.Radius)
		unit.Position.SetX(wallBounds.Min.X() + overlapX)
		// Applica la restituzione
		velocity.SetX(-velocity.X() * wallElasticity)
		unit.PreviousPosition = unit.Position.Add(velocity) // Aggiorna correttamente la posizione precedente

	}
	if unit.Position.X()+unit.Radius > wallBounds.Max.X() {
		overlapX := (unit.Position.X() + unit.Radius) - wallBounds.Max.X()
		unit.Position.SetX(wallBounds.Max.X() - overlapX)
		// Applica la restituzione
		velocity.SetX(-velocity.X() * wallElasticity)
		unit.PreviousPosition = unit.Position.Add(velocity) // Aggiorna correttamente la posizione precedente

	}

	// Correzione asse Y
	if unit.Position.Y()-unit.Radius < wallBounds.Min.Y() {
		overlapY := wallBounds.Min.Y() - (unit.Position.Y() - unit.Radius)
		unit.Position.SetY(wallBounds.Min.Y() + overlapY)
		// Applica la restituzione
		velocity.SetY(-velocity.Y() * wallElasticity)
		unit.PreviousPosition = unit.Position.Add(velocity) // Aggiorna correttamente la posizione precedente

	}
	if unit.Position.Y()+unit.Radius > wallBounds.Max.Y() {
		overlapY := (unit.Position.Y() + unit.Radius) - wallBounds.Max.Y()
		unit.Position.SetY(wallBounds.Max.Y() - overlapY)
		// Applica la restituzione
		velocity.SetY(-velocity.Y() * wallElasticity)
		unit.PreviousPosition = unit.Position.Add(velocity) // Aggiorna correttamente la posizione precedente

	}

	// Correzione asse Z
	if unit.Position.Z()-unit.Radius < wallBounds.Min.Z() {
		overlapZ := wallBounds.Min.Z() - (unit.Position.Z() - unit.Radius)
		unit.Position.SetZ(wallBounds.Min.Z() + overlapZ)
		// Applica la restituzione
		velocity.SetZ(-velocity.Z() * wallElasticity)
		unit.PreviousPosition = unit.Position.Add(velocity) // Aggiorna correttamente la posizione precedente

	}
	if unit.Position.Z()+unit.Radius > wallBounds.Max.Z() {
		overlapZ := (unit.Position.Z() + unit.Radius) - wallBounds.Max.Z()
		unit.Position.SetZ(wallBounds.Max.Z() - overlapZ)
		// Applica la restituzione
		velocity.SetZ(-velocity.Z() * wallElasticity)
		unit.PreviousPosition = unit.Position.Add(velocity) // Aggiorna correttamente la posizione precedente

	}

}
