package physics

import (
	"math"
	"sync"

	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/physics/collision"
)

func (s *Simulation) UpdateWithOctrees() error {
	if len(s.Fluid) == 0 {
		return nil
	}

	// Clean up merged units
	s.cleanupMergedUnits()

	// Update octree
	s.updateOctree()

	// Create WaitGroup for parallel operations
	var wg sync.WaitGroup

	// Apply gravitational forces in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.applyGravitationalForces()
	}()

	// Handle collisions in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.handleCollisions()
	}()

	// Wait for parallel operations to complete
	wg.Wait()

	// Handle heat transfer in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.handleHeatTransfer()
	}()

	// Wait for heat transfer to complete
	wg.Wait()

	// Update positions (must be sequential after forces are applied)
	s.updatePositions()

	return nil
}

// cleanupMergedUnits removes units that have been merged from the simulation
func (s *Simulation) cleanupMergedUnits() {
	// Create a new slice with the same capacity
	newFluid := make([]*Unit, 0, len(s.Fluid))

	// Only keep units that haven't been merged
	for _, unit := range s.Fluid {
		if unit != nil && !unit.isMerged {
			newFluid = append(newFluid, unit)
		} else if unit != nil && unit.isMerged {
			// Clean up merged unit's resources
			if unit.Mesh != nil {
				s.Scene.Remove(unit.Mesh)
				unit.Mesh = nil
			}
		}
	}

	// Update the Fluid slice
	s.Fluid = newFluid
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
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 8) // Limit concurrent goroutines

	for _, unitA := range s.Fluid {
		if unitA == nil || !unitA.CanBeAltered() {
			continue
		}

		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(unit *Unit) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			unit.CheckAndResolveWallCollision(s.WorldBoundray, s.Config.WallElasticity)

			// Get slice from pool
			nearUnitsPtr := NearUnitsPool.Get().(*[]*Unit)
			nearUnits := *nearUnitsPtr
			nearUnits = nearUnits[:0] // Reset slice but keep capacity

			s.Octree.Retrieve(&nearUnits, unit)

			for _, unitB := range nearUnits {
				if !isValidCollisionPair(unit, unitB) {
					continue
				}

				collData := collision.GatherCollisionData(unit, unitB)
				if collData.Collided {
					collision.ResolveCollision(collData)
				}
			}

			// Return slice to pool
			*nearUnitsPtr = nearUnits
			NearUnitsPool.Put(nearUnitsPtr)
		}(unitA)
	}

	wg.Wait()
}

func (s *Simulation) handleHeatTransfer() {
	deltaTime := s.GetDeltaTime()
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 8) // Limit concurrent goroutines

	for _, unitA := range s.Fluid {
		if unitA == nil || !unitA.CanBeAltered() {
			continue
		}

		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(unit *Unit) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			// Get slice from pool
			nearUnitsPtr := NearUnitsPool.Get().(*[]*Unit)
			nearUnits := *nearUnitsPtr
			nearUnits = nearUnits[:0] // Reset slice but keep capacity

			s.Octree.Retrieve(&nearUnits, unit)

			for _, unitB := range nearUnits {
				if !isValidHeatTransferPair(unit, unitB) {
					continue
				}

				unit.TransferHeatTo(unitB, deltaTime)
			}

			// Return slice to pool
			*nearUnitsPtr = nearUnits
			NearUnitsPool.Put(nearUnitsPtr)
		}(unitA)
	}

	wg.Wait()
}

func (s *Simulation) updatePositions() {
	deltaTime := s.GetDeltaTime()
	for _, unit := range s.Fluid {
		if unit != nil && unit.CanBeAltered() {
			unit.UpdatePosition(deltaTime)
		}
	}
}

func isValidCollisionPair(unitA, unitB *Unit) bool {
	return unitB != nil &&
		unitA.Id != unitB.Id &&
		unitB.CanBeAltered()
}

func isValidHeatTransferPair(unitA, unitB *Unit) bool {
	return unitB != nil &&
		unitA.Id != unitB.Id &&
		unitB.CanBeAltered() &&
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
	deltaPos := g.Position().Sub(ot.CenterOfMass)
	distanceSquared := deltaPos.X()*deltaPos.X() + deltaPos.Y()*deltaPos.Y() + deltaPos.Z()*deltaPos.Z()

	// Avoid sqrt by comparing squares
	if (width * width) < (theta * theta * distanceSquared) {
		ot.approximateGravityWithCenterOfMass(g, force)
		return
	}

	ot.calculateChildrenGravity(g, theta, force)
}

func (ot *Octree) calculateLeafNodeGravity(g Gravitable, force *vector3.Vector[float64]) {
	gMass := g.Mass()    // Cache mass value
	gPos := g.Position() // Cache position

	// Get temporary vector from pool
	vec := VectorPool.Get().([]float64)
	defer VectorPool.Put(vec)

	for _, obj := range ot.objects {
		if obj != g.Unit() {
			// Calculate distance components directly
			vec[0] = obj.Position().X() - gPos.X()
			vec[1] = obj.Position().Y() - gPos.Y()
			vec[2] = obj.Position().Z() - gPos.Z()

			distanceSquared := vec[0]*vec[0] + vec[1]*vec[1] + vec[2]*vec[2]

			if distanceSquared > 0 {
				// Avoid sqrt when possible using distanceSquared
				invDistCubed := 1.0 / (distanceSquared * math.Sqrt(distanceSquared))

				// Calculate force magnitude
				forceMagnitude := G * gMass * obj.Mass()

				// Calculate force components directly
				fx := vec[0] * forceMagnitude * invDistCubed
				fy := vec[1] * forceMagnitude * invDistCubed
				fz := vec[2] * forceMagnitude * invDistCubed

				// Update force vector
				*force = force.Add(vector3.New(fx, fy, fz))
			}
		}
	}
}

func (ot *Octree) approximateGravityWithCenterOfMass(g Gravitable, force *vector3.Vector[float64]) {
	// Get temporary vector from pool
	vec := VectorPool.Get().([]float64)
	defer VectorPool.Put(vec)

	// Calculate position difference directly
	pos := g.Position()
	vec[0] = ot.CenterOfMass.X() - pos.X()
	vec[1] = ot.CenterOfMass.Y() - pos.Y()
	vec[2] = ot.CenterOfMass.Z() - pos.Z()

	distanceSquared := vec[0]*vec[0] + vec[1]*vec[1] + vec[2]*vec[2]

	if distanceSquared > 0 {
		// Avoid sqrt when possible using distanceSquared
		invDistCubed := 1.0 / (distanceSquared * math.Sqrt(distanceSquared))

		// Calculate force magnitude
		forceMagnitude := G * g.Mass() * ot.TotalMass

		// Calculate force components directly
		fx := vec[0] * forceMagnitude * invDistCubed
		fy := vec[1] * forceMagnitude * invDistCubed
		fz := vec[2] * forceMagnitude * invDistCubed

		// Update force vector
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
