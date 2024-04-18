package physics

import (
	"github.com/g3n/engine/math32"
)

const G = 6.67430e-11 // Costante gravitazionale universale (m^3 kg^-1 s^-2)

func (s *Simulation) UpdateWithOctrees() error {
	if len(s.Fluid) == 0 {
		return nil
	}

	frameTime := s.Config.Frametime

	s.Octree.Clear(s.Scene) // Pulisce il Octree all'inizio di ogni frame

	for _, unit := range s.Fluid {
		s.Octree.Insert(unit, s.Scene)
	}

	// Calcola la forza di gravità in modo concorrente
	if s.Config.UnitsEmitGravity {
		// log.Print("Calculating gravity")
		for _, unit := range s.Fluid {
			unit.accelerate(s.Octree.CalculateGravity(unit, 0.5))
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
		unitA.CheckAndResolveWallCollision(s.WorldBoundray, s.Config.WallElasticity)

		// Trova le particelle vicine utilizzando l'Octree
		nearUnits := []*Unit{}
		s.Octree.Retrieve(&nearUnits, unitA)

		// Gestisci le collisioni tra particelle vicine
		for _, unitB := range nearUnits {
			if unitB != nil && unitA.Id != unitB.Id {
				uAPositon := unitA.Mesh.Position()
				uBPosition := unitB.Mesh.Position()
				distance := uAPositon.DistanceTo(&uBPosition)
				if distance <= unitA.Radius+unitB.Radius {
					handleCollision(unitA, unitB)
				}
			}
		}

		// if unitA.Mesh.Position().X > s.WorldBoundray.Max.X || unitA.Mesh.Position().X < s.WorldBoundray.Min.X || unitA.Mesh.Position().Y > s.WorldBoundray.Max.Y || unitA.Mesh.Position().Y < s.WorldBoundray.Min.Y || unitA.Mesh.Position().Z > s.WorldBoundray.Max.Z || unitA.Mesh.Position().Z < s.WorldBoundray.Min.Z {
		// 	log.Print("Unit out of bounds")
		// 	s.IsPause = true
		// 	return nil
		// }
	}
	return nil
}

func (ot *Octree) CalculateGravity(unit *Unit, theta float32) *math32.Vector3 {
	var force = math32.NewVec3()
	ot.calculateGravityRecursive(unit, theta, force)
	return force
}

func (ot *Octree) calculateGravityRecursive(unit *Unit, theta float32, force *math32.Vector3) {
	if ot.Children[0] == nil {
		// Se siamo in un nodo foglia, calcoliamo la forza tra tutti gli oggetti e la unit.
		for _, obj := range ot.objects {
			if obj != unit {
				objPos := obj.Mesh.Position()
				unitPos := unit.Mesh.Position()
				deltaPos := objPos.Sub(&unitPos)
				distance := deltaPos.Length()
				if distance > 0 {
					// Calcolo della forza gravitazionale.
					magnitude := G * unit.Mass * obj.Mass / (distance * distance)
					direction := deltaPos.Normalize()
					forceToAdd := direction.MultiplyScalar(magnitude)
					force = force.Add(forceToAdd)
				}
			}
		}
	} else {
		// Se non siamo in un nodo foglia, decidiamo se calcolare la forza con il centro di massa o scendere nell'Octree.
		width := ot.Bounds.Max.X - ot.Bounds.Min.X
		unitPos := unit.Mesh.Position()
		distance := unitPos.DistanceTo(&ot.CenterOfMass)
		if (width / distance) < theta {
			// Usiamo il centro di massa per approssimare la forza.
			deltaPos := ot.CenterOfMass.Sub(&unitPos)
			distance := deltaPos.Length()
			if distance > 0 {
				magnitude := G * unit.Mass * ot.TotalMass / (distance * distance)
				direction := deltaPos.Normalize()
				forceToAdd := direction.MultiplyScalar(magnitude)
				force = force.Add(forceToAdd)
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

	e := math32.Min(uA.Elasticity, uB.Elasticity)

	// define the relative velocity
	vRel := uA.Velocity.Sub(uB.Velocity)

	// calculate the impulseVectorAlongTheNormal
	uAPos := uA.Mesh.Position()
	uBPos := uB.Mesh.Position()
	impulseDirection := uAPos.Sub(&uBPos).Normalize()
	impulseMag := float32(-(1 + e)) * vRel.Dot(impulseDirection) / (1/uA.Mass + 1/uB.Mass)

	jn := impulseDirection.MultiplyScalar(impulseMag)

	// Apply the impulse to both objects in opposite directions
	uA.Velocity = uA.Velocity.Add(jn.MultiplyScalar(1 / uA.Mass))
	uB.Velocity = uB.Velocity.Sub(jn.MultiplyScalar(1 / uB.Mass))

	// move the units along the normals
	totalRadius := uA.Radius + uB.Radius
	overlap := totalRadius - uAPos.DistanceTo(&uBPos)
	moveDistance := overlap / 2
	normalMove := impulseDirection.MultiplyScalar(moveDistance)
	uA.Mesh.SetPositionVec(uAPos.Add(normalMove))
	uB.Mesh.SetPositionVec(uBPos.Sub(normalMove))

	// Convert the magnitude of the impulse to heat for each unit
	uA.Heat += 2
	uB.Heat += 2
}
