package physics

import (
	"math"
	"sync"

	"github.com/EliCDavis/vector/vector3"
)

const G = 6.67430e-11 // Costante gravitazionale universale (m^3 kg^-1 s^-2)

// Definisci il coefficiente di damping
const dampingCoefficient = 1.1 // Sostituisci con il valore desiderato

type CollData struct {
	uA, uB *Unit

	distance        float64
	surfaceDistance float64
	collided        bool
	dampingDistance float64

	// the normal of the collision
	impulseDirection vector3.Vector[float64]

	// define the relative velocity (velocityVector)
	vRel        vector3.Vector[float64]
	rVelNormal  float64
	relativeVel float64
	e           float64
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
				collData := s.gatherCollisionData(unitA, unitB)

				if collData.collided {
					handleCollision(collData)
				} else if collData.distance <= collData.dampingDistance {
					handleDamping(collData)
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

func handleDamping(collData *CollData) {
	targetDistance := collData.dampingDistance / 2

	// Calcola la forza di damping basata sulla velocità relativa
	dampingForceMag := -dampingCoefficient * collData.rVelNormal

	// Calcola l'accelerazione di damping e applicala alle accelerazioni delle unità
	dampingAcceleration := collData.impulseDirection.Scale(dampingForceMag)
	collData.uA.Acceleration = collData.uA.Acceleration.Add(dampingAcceleration.Scale(1 / collData.uA.Mass))
	collData.uB.Acceleration = collData.uB.Acceleration.Sub(dampingAcceleration.Scale(1 / collData.uB.Mass))

	correctionForceMag := (targetDistance - collData.distance) * dampingCoefficient

	// Calcola l'accelerazione di correzione e applicala alle accelerazioni delle unità
	correctionAcceleration := collData.impulseDirection.Scale(correctionForceMag)
	collData.uA.Acceleration = collData.uA.Acceleration.Add(correctionAcceleration.Scale(1 / collData.uA.Mass))
	collData.uB.Acceleration = collData.uB.Acceleration.Sub(correctionAcceleration.Scale(1 / collData.uB.Mass))
}

func (s *Simulation) gatherCollisionData(uA, uB *Unit) (collData *CollData) {
	collData = &CollData{
		uA:               uA,
		uB:               uB,
		e:                0,
		impulseDirection: vector3.Zero[float64](),
		vRel:             vector3.Zero[float64](),
		rVelNormal:       0,
		collided:         false,
		relativeVel:      0,
		distance:         0,
		surfaceDistance:  0,
		dampingDistance:  0,
	}
	collData.distance = uA.Position.Distance(uB.Position)
	collData.surfaceDistance = collData.distance - uA.Radius + uB.Radius

	collData.collided = collData.distance <= collData.uA.Radius+collData.uB.Radius

	if collData.distance > collData.surfaceDistance*2 {
		return
	}

	collData.vRel = uA.Velocity.Sub(uB.Velocity)
	collData.impulseDirection = uA.Position.Sub(uB.Position).Normalized()
	collData.rVelNormal = collData.vRel.Dot(collData.impulseDirection)
	collData.dampingDistance = (collData.uA.Radius + collData.uB.Radius) * s.Config.UnitRadiusMultiplier * 2

	if collData.collided {
		collData.e = math.Min(uA.Elasticity, uB.Elasticity)
	}

	return
}

func handleCollision(collData *CollData) {

	impulseMag := -(1 + collData.e) * collData.rVelNormal / (1/collData.uA.Mass + 1/collData.uB.Mass)

	jn := collData.impulseDirection.Scale(impulseMag)

	// Apply the impulse to both objects in opposite directions
	collData.uA.Velocity = collData.uA.Velocity.Add(jn.Scale(1 / collData.uA.Mass))
	collData.uB.Velocity = collData.uB.Velocity.Sub(jn.Scale(1 / collData.uB.Mass))

	// move the units along the normals
	totalRadius := collData.uA.Radius + collData.uB.Radius
	overlap := totalRadius - collData.distance
	if overlap > 0 {
		collData.uA.Mesh.SetMaterial(overlapMat)
		collData.uB.Mesh.SetMaterial(overlapMat)

		moveDistance := overlap / 2
		normalMove := collData.impulseDirection.Scale(moveDistance)
		collData.uA.Position = collData.uA.Position.Add(normalMove)
		collData.uB.Position = collData.uB.Position.Sub(normalMove)

	}

	// Gestisci il calore generato dalla collisione (opzionale)
	heatTransfer := collData.rVelNormal * collData.distance * 0.5
	collData.uA.Heat += heatTransfer * (1.0 - collData.uA.Elasticity)
	collData.uB.Heat += heatTransfer * (1.0 - collData.uB.Elasticity)
}
