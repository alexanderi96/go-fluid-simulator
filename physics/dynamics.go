package physics

import (
	"sync"

	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/physics/collision"
	"github.com/alexanderi96/go-fluid-simulator/physics/gravity"
)

func (s *Simulation) UpdateWithOctrees() error {
	if len(s.Fluid) == 0 {
		return nil
	}

	s.updateOctree()
	s.applyGravitationalForces()
	s.handleCollisions()
	s.handleHeatTransfer()
	s.updatePositions()

	return nil
}

func (s *Simulation) updateOctree() {
	s.Octree.Clear(s.Scene)
	for _, unit := range s.Fluid {
		s.Octree.Insert(unit, s.Scene)
	}
}

func (s *Simulation) applyGravitationalForces() {
	var wg sync.WaitGroup
	wg.Add(len(s.Fluid))

	for _, unit := range s.Fluid {
		go func(u *Unit) {
			defer wg.Done()
			u.ApplyForce(s.Octree.CalculateGravity(u, 0.5))
		}(unit)
	}

	wg.Wait()
}

func (s *Simulation) handleCollisions() {
	for _, unitA := range s.Fluid {
		if unitA == nil || !unitA.CanBeAltered {
			continue
		}

		unitA.CheckAndResolveWallCollision(s.WorldBoundray, s.Config.WallElasticity)

		nearUnits := make([]*Unit, 0, 8) // Pre-allocate with typical capacity
		s.Octree.Retrieve(&nearUnits, unitA)

		for _, unitB := range nearUnits {
			if !isValidCollisionPair(unitA, unitB) {
				continue
			}

			collData := collision.GatherCollisionData(unitA, unitB)
			if collData.Collided {
				collision.ResolveCollision(collData)
			}
		}
	}
}

func (s *Simulation) handleHeatTransfer() {
	for _, unitA := range s.Fluid {
		if unitA == nil || !unitA.CanBeAltered {
			continue
		}

		nearUnits := make([]*Unit, 0, 8) // Pre-allocate with typical capacity
		s.Octree.Retrieve(&nearUnits, unitA)

		for _, unitB := range nearUnits {
			if !isValidHeatTransferPair(unitA, unitB) {
				continue
			}

			unitA.TransferHeatTo(unitB, s.Config.Frametime)
		}
	}
}

func (s *Simulation) updatePositions() {
	for _, unit := range s.Fluid {
		if unit != nil && unit.CanBeAltered {
			unit.UpdatePosition(s.Config.Frametime)
		}
	}
}

func isValidCollisionPair(unitA, unitB *Unit) bool {
	return unitB != nil &&
		unitA.Id != unitB.Id &&
		unitB.CanBeAltered
}

func isValidHeatTransferPair(unitA, unitB *Unit) bool {
	return unitB != nil &&
		unitA.Id != unitB.Id &&
		unitB.CanBeAltered &&
		unitA.Heat > unitB.Heat
}

func (ot *Octree) CalculateGravity(unit Gravitable, theta float64) vector3.Vector[float64] {
	force := vector3.Zero[float64]()
	ot.calculateGravityRecursive(unit, theta, &force)
	return force
}

func (ot *Octree) calculateGravityRecursive(g Gravitable, theta float64, force *vector3.Vector[float64]) {
	if !ot.divided {
		ot.calculateLeafNodeGravity(g, force)
		return
	}

	// Calculate width and distance squared directly
	width := ot.Bounds.Max.X() - ot.Bounds.Min.X()
	deltaPos := g.GetPosition().Sub(ot.CenterOfMass)
	distanceSquared := deltaPos.X()*deltaPos.X() + deltaPos.Y()*deltaPos.Y() + deltaPos.Z()*deltaPos.Z()

	// Avoid sqrt by comparing squares
	if (width * width) < (theta * theta * distanceSquared) {
		ot.approximateGravityWithCenterOfMass(g, force)
		return
	}

	ot.calculateChildrenGravity(g, theta, force)
}

func (ot *Octree) calculateLeafNodeGravity(g Gravitable, force *vector3.Vector[float64]) {
	gMass := g.GetMass()    // Cache mass value
	gPos := g.GetPosition() // Cache position

	for _, obj := range ot.objects {
		if obj != g.GetUnit() {
			deltaPos := obj.GetPosition().Sub(gPos)
			distanceSquared := deltaPos.X()*deltaPos.X() + deltaPos.Y()*deltaPos.Y() + deltaPos.Z()*deltaPos.Z()

			if distanceSquared > 0 {
				// Pre-calculate common factors
				forceMagnitude := gravity.UniversalGravitationalConstant * gMass * obj.GetMass() / distanceSquared
				invDistance := 1.0 / distanceSquared

				// Calculate force components directly
				fx := deltaPos.X() * forceMagnitude * invDistance
				fy := deltaPos.Y() * forceMagnitude * invDistance
				fz := deltaPos.Z() * forceMagnitude * invDistance

				*force = force.Add(vector3.New(fx, fy, fz))
			}
		}
	}
}

func (ot *Octree) approximateGravityWithCenterOfMass(g Gravitable, force *vector3.Vector[float64]) {
	deltaPos := ot.CenterOfMass.Sub(g.GetPosition())
	distanceSquared := deltaPos.X()*deltaPos.X() + deltaPos.Y()*deltaPos.Y() + deltaPos.Z()*deltaPos.Z()

	if distanceSquared > 0 {
		// Pre-calculate force magnitude
		forceMagnitude := gravity.UniversalGravitationalConstant * g.GetMass() * ot.TotalMass / distanceSquared
		invDistance := 1.0 / distanceSquared

		// Calculate force components directly
		fx := deltaPos.X() * forceMagnitude * invDistance
		fy := deltaPos.Y() * forceMagnitude * invDistance
		fz := deltaPos.Z() * forceMagnitude * invDistance

		*force = force.Add(vector3.New(fx, fy, fz))
	}
}

func (ot *Octree) calculateChildrenGravity(g Gravitable, theta float64, force *vector3.Vector[float64]) {
	for _, child := range ot.Children {
		if child != nil {
			child.calculateGravityRecursive(g, theta, force)
		}
	}
}
