package physics

import (
	"github.com/EliCDavis/vector/vector3"
)

const G = 6.67430e1 // Costante gravitazionale universale (m^3 kg^-1 s^-2)

func (s *Simulation) UpdateWithOctrees() error {
	frameTime := s.Metrics.Frametime

	s.Octree.Clear() // Pulisce il Octree all'inizio di ogni frame

	// Costruisci l'Octree
	for _, unit := range s.Fluid {
		s.Octree.Insert(unit)
	}

	for _, unit := range s.Fluid {
		if s.Config.UnitsEmitGravity {
			unit.accelerate(s.Octree.CalculateGravity(unit, 0.5))
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
		unitA.UpdatePosition(frameTime)

		// Gestisci le collisioni con le pareti
		if unitA.CheckAndResolveWallCollision(s.WorldBoundray, s.Config.WallElasticity) {
			//s.IsPause = true
		}

		// Trova le particelle vicine utilizzando l'Octree
		nearUnits := []*Unit{}
		s.Octree.Retrieve(&nearUnits, unitA)

		// Gestisci le collisioni tra particelle vicine
		// for _, unitB := range nearUnits {
		// 	if unitB != nil && unitA.Id != unitB.Id {
		// 		distance := unitA.Position.Distance(unitB.Position)
		// 		if distance <= unitA.Radius+unitB.Radius {
		// 			handleCollision(unitA, unitB)
		// 		}
		// 	}
		// }
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

func handleCollision(uA, uB *Unit) {
	// Calculate the unit normal and unit tangent vectors
	normal := uB.Position.Sub(uA.Position).Normalized()
	tangent := vector3.New(-normal.Y(), normal.X(), normal.Z())

	// Project the velocities onto the normal and tangent vectors
	v1n := normal.Dot(uA.Velocity)
	v1t := tangent.Dot(uA.Velocity)
	v2n := normal.Dot(uB.Velocity)
	v2t := tangent.Dot(uB.Velocity)

	// Calculate new normal velocities using the one-dimensional elastic collision equations
	v1nPrime := (v1n*(uA.Mass-uB.Mass) + 2*uB.Mass*v2n) / (uA.Mass + uB.Mass)
	v2nPrime := (v2n*(uB.Mass-uA.Mass) + 2*uA.Mass*v1n) / (uA.Mass + uB.Mass)

	// Convert scalar normal and tangent velocities into vectors
	v1nPrimeVec := normal.Scale(v1nPrime)
	v1tVec := tangent.Scale(v1t)
	v2nPrimeVec := normal.Scale(v2nPrime)
	v2tVec := tangent.Scale(v2t)

	// Update the velocities of the particles by adding the normal and tangent components
	uA.Velocity = v1nPrimeVec.Add(v1tVec)
	uB.Velocity = v2nPrimeVec.Add(v2tVec)
}
