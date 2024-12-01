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

		nearUnits := []*Unit{}
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

func (ot *Octree) CalculateGravity(unit Gravitable, theta float64) vector3.Vector[float64] {
	var force = vector3.Zero[float64]()
	ot.calculateGravityRecursive(unit, theta, &force)
	return force
}

func (ot *Octree) calculateGravityRecursive(g Gravitable, theta float64, force *vector3.Vector[float64]) {
	if !ot.divided {
		ot.calculateLeafNodeGravity(g, force)
		return
	}

	width := ot.Bounds.Max.X() - ot.Bounds.Min.X()
	distance := g.GetPosition().Distance(ot.CenterOfMass)

	if (width / distance) < theta {
		ot.approximateGravityWithCenterOfMass(g, force)
		return
	}

	ot.calculateChildrenGravity(g, theta, force)
}

func (ot *Octree) calculateLeafNodeGravity(g Gravitable, force *vector3.Vector[float64]) {
	for _, obj := range ot.objects {
		if obj != g.GetUnit() {
			forceToAdd := gravity.CalculateForce(g, obj)
			*force = force.Add(forceToAdd)
		}
	}
}

func (ot *Octree) approximateGravityWithCenterOfMass(g Gravitable, force *vector3.Vector[float64]) {
	deltaPos := ot.CenterOfMass.Sub(g.GetPosition())
	distance := deltaPos.Length()

	if distance > 0 {
		magnitude := gravity.UniversalGravitationalConstant * g.GetMass() * ot.TotalMass / (distance * distance)
		direction := deltaPos.Normalized()
		*force = force.Add(direction.Scale(magnitude))
	}
}

func (ot *Octree) calculateChildrenGravity(g Gravitable, theta float64, force *vector3.Vector[float64]) {
	for _, child := range ot.Children {
		if child != nil {
			child.calculateGravityRecursive(g, theta, force)
		}
	}
}
