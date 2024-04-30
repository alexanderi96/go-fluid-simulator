package physics

import (
	"math"
	"sync"

	"github.com/EliCDavis/vector/vector3"
)

const G = 10 //6.67430e-11 // Costante gravitazionale universale (m^3 kg^-1 s^-2)

type CollData struct {
	uA, uB *Unit

	distance    float64
	totalRadius float64
	totalMass   float64
	collided    bool

	// the normal of the collision
	impulseDirection vector3.Vector[float64]

	// define the relative velocity (velocityVector)
	vRel       vector3.Vector[float64]
	rVelNormal float64
	e          float64
}

func (s *Simulation) UpdateWithOctrees() error {
	if len(s.Fluid) == 0 {
		return nil
	}

	frameTime := s.Config.Frametime

	s.Octree.Clear(s.Scene) // Pulisce il Octree all'inizio di ogni frame

	for _, unit := range s.Fluid {
		s.Octree.Insert(unit, s.Scene)
	}

	var wg sync.WaitGroup
	wg.Add(len(s.Fluid)) // Aggiungi il conteggio delle unità

	for _, unit := range s.Fluid {
		go func(u *Unit) {
			defer wg.Done()
			u.accelerate(s.Octree.CalculateGravity(u, 0.9))
		}(unit)
	}

	wg.Wait()

	for _, unitA := range s.Fluid {
		if unitA == nil {
			continue
		}
		// Gestisci le collisioni con le pareti
		unitA.CheckAndResolveWallCollision(s.WorldBoundray, s.Config.WallElasticity)

		// Trova le particelle vicine utilizzando l'Octree
		nearUnits := []*Unit{}
		s.Octree.Retrieve(&nearUnits, unitA)

		// Gestisci le collisioni tra particelle vicine
		for _, unitB := range nearUnits {
			if unitB != nil && unitA.Id != unitB.Id {
				collData := s.gatherCollisionData(unitA, unitB)

				if collData.collided {
					handleCollision(collData)
				}
			}
		}

		unitA.UpdatePosition(frameTime)
	}

	return nil
}

func (ot *Octree) CalculateGravity(unit *Unit, theta float64) vector3.Vector[float64] {
	var force = vector3.Zero[float64]()
	ot.calculateGravityRecursive(unit, theta, &force)
	return force
}

func (ot *Octree) calculateGravityRecursive(unit *Unit, theta float64, force *vector3.Vector[float64]) {
	if !ot.divided {
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

func (s *Simulation) gatherCollisionData(uA, uB *Unit) *CollData {
	collData := &CollData{
		uA:          uA,
		uB:          uB,
		distance:    uA.Position.Distance(uB.Position),
		totalRadius: uA.Radius + uB.Radius,
	}

	collData.collided = collData.distance < collData.totalRadius

	// todo: continua
	if collData.collided {
		collData.e = math.Min(uA.Elasticity, uB.Elasticity)
		collData.impulseDirection = uA.Position.Sub(uB.Position).Normalized()
		collData.vRel = uA.Velocity.Sub(uB.Velocity)
		collData.rVelNormal = collData.vRel.Dot(collData.impulseDirection)
		collData.totalMass = uA.Mass + uB.Mass
	}

	return collData
}

func handleCollision(collData *CollData) {

	impulseMag := -(1 + collData.e) * collData.rVelNormal / (1/collData.uA.Mass + 1/collData.uB.Mass)

	jn := collData.impulseDirection.Scale(impulseMag)

	// Apply the impulse to both objects in opposite directions
	collData.uA.Velocity = collData.uA.Velocity.Add(jn.Scale(1 / collData.uA.Mass))
	collData.uB.Velocity = collData.uB.Velocity.Sub(jn.Scale(1 / collData.uB.Mass))

	// move the units along the normals
	// collData.uA.Mesh.SetMaterial(overlapMat)
	// collData.uB.Mesh.SetMaterial(overlapMat)
	overlap := collData.totalRadius - collData.distance
	massTotal := collData.totalMass
	moveDistanceA := (collData.uB.Mass / massTotal) * overlap
	moveDistanceB := (collData.uA.Mass / massTotal) * overlap

	normalMoveA := collData.impulseDirection.Scale(moveDistanceA)
	normalMoveB := collData.impulseDirection.Scale(moveDistanceB)

	collData.uA.Position = collData.uA.Position.Add(normalMoveA)
	collData.uB.Position = collData.uB.Position.Sub(normalMoveB)

	// Gestisci il calore generato dalla collisione (opzionale)
	heatTransfer := collData.rVelNormal * collData.distance * 0.5
	collData.uA.Heat += heatTransfer * (1.0 - collData.uA.Elasticity)
	collData.uB.Heat += heatTransfer * (1.0 - collData.uB.Elasticity)

}
