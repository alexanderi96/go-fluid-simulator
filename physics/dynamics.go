package physics

import (
	"math"
	"sync"

	"github.com/EliCDavis/vector/vector3"
)

const G = 6.67430e-11 // Costante gravitazionale universale (m^3 kg^-1 s^-2)

func (s *Simulation) UpdateWithOctrees() error {
	if len(s.Fluid) == 0 {
		return nil
	}

	frameTime := s.Config.Frametime

	s.Octree.Clear() // Pulisce il Octree all'inizio di ogni frame

	for _, unit := range s.Fluid {
		s.Octree.Insert(unit)
	}

	// Calcola la forza di gravità in modo concorrente
	var wgGravity sync.WaitGroup
	for _, unit := range s.Fluid {
		wgGravity.Add(1)
		go func(unit *Unit) {
			defer wgGravity.Done()
			force := s.Octree.CalculateGravity(unit, 0.5)
			if s.Config.UnitsEmitGravity {
				unit.accelerate(force)
			}
		}(unit)
	}
	wgGravity.Wait()
	// Controlla le collisioni tra particelle e aggiorna le velocità utilizzando il Octree
	for _, unitA := range s.Fluid {
		if unitA == nil {
			continue
		}

		// Aggiorna la posizione utilizzando il metodo Verlet
		unitA.UpdatePosition(frameTime)

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
					handleCollision(unitA, unitB)
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

func handleCollision(uA, uB *Unit) {

	e := math.Min(uA.Elasticity, uB.Elasticity)

	// define the relative velocity
	vRel := uA.Velocity.Sub(uB.Velocity)

	// calculate the impulseVectorAlongTheNormal
	impulseDirection := uA.Position.Sub(uB.Position).Normalized()
	impulseMag := -(1 + e) * vRel.Dot(impulseDirection) / (1/uA.Mass + 1/uB.Mass)

	jn := impulseDirection.Scale(impulseMag)

	// Apply the impulse to both objects in opposite directions
	uA.Velocity = uA.Velocity.Add(jn.Scale(1 / uA.Mass))
	uB.Velocity = uB.Velocity.Sub(jn.Scale(1 / uB.Mass))

	// move the units along the normals
	totalRadius := uA.Radius + uB.Radius
	overlap := totalRadius - uA.Position.Distance(uB.Position)
	moveDistance := overlap / 2
	normalMove := impulseDirection.Scale(moveDistance)
	uA.Position = uA.Position.Add(normalMove)
	uB.Position = uB.Position.Sub(normalMove)

	// Convert the magnitude of the impulse to heat for each unit
	uA.Heat += 2
	uB.Heat += 2
}
