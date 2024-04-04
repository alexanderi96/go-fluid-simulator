package physics

import (
	"math"

	"github.com/alexanderi96/go-fluid-simulator/config"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

func (s *Simulation) UpdateWithOctrees() error {
	s.Octree.Clear() // Pulisce il Octree all'inizio di ogni frame

	s.ClusterMasses = make(map[uuid.UUID]float32)

	// Costruisci l'Octree
	for _, unit := range s.Fluid {
		s.Octree.Insert(unit)
	}

	// Controlla le collisioni tra particelle e aggiorna le velocità utilizzando il Octree
	for _, unitA := range s.Fluid {
		if unitA == nil {
			continue
		}
		unitA.checkWallCollisionVerlet(s.WorldBoundray, s.Config.WallElasticity, s.Metrics.Frametime)

		nearestUnit := Unit{}
		nearestValidDistance := float32(math.Inf(1))

		nearUnits := []*Unit{}

		s.Octree.Retrieve(&nearUnits, unitA)
		for _, unitB := range nearUnits {
			if unitB != nil || unitA.Id != unitB.Id {
				surfaceDistance := getSurfaceDistance(unitA, unitB)
				if surfaceDistance < 0 {
					handleCollision(unitA, unitB, surfaceDistance, s.Metrics.Frametime)
				}
			}
		}

		if nearestValidDistance <= s.Config.ClusterThreshold {
			// in questo caso consideriamo le 2 unit appartenenti allo stesso cluster
			updateClusters(unitA, &nearestUnit, s.ClusterMasses)

		} else {
			unitA.OldCluster = unitA.Cluster
			unitA.Cluster = nil
		}

		unitA.update(s.Config.ApplyGravity, s.Config.Gravity, s.Metrics.Frametime)
	}

	return nil

}

func (s *Simulation) UpdateWithVerletIntegration() error {
	// Calcola il numero di step
	resolutionSteps := int(s.Config.ResolutionSteps)
	fractionalFrametime := s.Metrics.Frametime / float32(resolutionSteps)

	s.ClusterMasses = make(map[uuid.UUID]float32)

	for step := 0; step < resolutionSteps; step++ {

		for _, unitA := range s.Fluid {
			if unitA == nil {
				continue
			}

			unitA.checkWallCollisionVerlet(s.WorldBoundray, s.Config.WallElasticity, fractionalFrametime)

			if unitA.Cluster != nil {
				s.ClusterMasses[unitA.Cluster.Id] += unitA.Mass / float32(resolutionSteps)
			}

			nearestUnit := Unit{}
			nearestValidDistance := float32(math.Inf(-1))

			for _, unitB := range s.Fluid {
				if unitB == nil || unitA.Id == unitB.Id {
					continue
				}

				surfaceDistance := getSurfaceDistance(unitA, unitB)

				if s.Config.UnitsEmitGravity {
					applyGravitationalAttraction(unitA, unitB, s.Config)
				}

				if (nearestUnit == Unit{} || (nearestUnit != *unitB && surfaceDistance < nearestValidDistance)) {
					nearestUnit = *unitB
					nearestValidDistance = surfaceDistance
				}

				if surfaceDistance < 0 {
					handleCollision(unitA, unitB, surfaceDistance, fractionalFrametime)
				}
			}

			// TODO: find nearest unit in order to determin if unitA is in cluster
			if nearestValidDistance <= s.Config.ClusterThreshold {
				// in questo caso consideriamo le 2 unit appartenenti allo stesso cluster
				updateClusters(unitA, &nearestUnit, s.ClusterMasses)

			} else {
				unitA.OldCluster = unitA.Cluster
				unitA.Cluster = nil
			}

			unitA.update(s.Config.ApplyGravity, s.Config.Gravity, fractionalFrametime)
		}
	}

	return nil
}

func getDistance(unitA, unitB *Unit) float32 {
	return rl.Vector3Distance(unitA.Position, unitB.Position)
}

func getSurfaceDistance(unitA, unitB *Unit) float32 {
	return getDistance(unitA, unitB) - (unitA.Radius + unitB.Radius)
}

func applyGravitationalAttraction(a, b *Unit, config *config.Config) {
	dx := b.Position.X - a.Position.X
	dy := b.Position.Y - a.Position.Y
	dz := b.Position.Z - a.Position.Z
	distanceSquared := dx*dx + dy*dy + dz*dz
	distance := float32(math.Sqrt(float64(distanceSquared)))

	// Calcola il raggio totale delle due unità
	totalRadius := a.Radius + b.Radius

	// Evita la divisione per zero e le forze estremamente forti a distanze molto piccole
	if distance+totalRadius <= 0 {
		return
	}

	forceMagnitude := config.UnitGravitationalMultiplier * (a.Mass * b.Mass) / distanceSquared

	forceX := forceMagnitude * (dx / distance)
	forceY := forceMagnitude * (dy / distance)
	forceZ := forceMagnitude * (dz / distance)

	a.Acceleration.X += forceX / a.Mass
	a.Acceleration.Y += forceY / a.Mass
	a.Acceleration.Z += forceZ / a.Mass
	b.Acceleration.X -= forceX / b.Mass
	b.Acceleration.Y -= forceY / b.Mass
	b.Acceleration.Z -= forceZ / b.Mass
}

func areOverlapping(a, b *Unit) bool {
	return getSurfaceDistance(a, b) <= 0
}

func handleCollision(a, b *Unit, surfaceDistance, dt float32) {
	dx := b.Position.X - a.Position.X
	dy := b.Position.Y - a.Position.Y
	dz := b.Position.Z - a.Position.Z
	distanceSquared := dx*dx + dy*dy + dz*dz
	distance := float32(math.Sqrt(float64(distanceSquared)))

	if distanceSquared == 0 {
		return
	}

	normalX := dx / distance
	normalY := dy / distance
	normalZ := dz / distance

	overlap := -surfaceDistance // Sovrapposizione positiva
	inverseMassA := 1 / a.Mass
	inverseMassB := 1 / b.Mass
	inverseTotalMass := inverseMassA + inverseMassB
	correction := overlap / inverseTotalMass

	// Applica la correzione alla posizione corrente
	a.Position.X -= normalX * correction * inverseMassA
	a.Position.Y -= normalY * correction * inverseMassA
	a.Position.Z -= normalZ * correction * inverseMassA
	b.Position.X += normalX * correction * inverseMassB
	b.Position.Y += normalY * correction * inverseMassB
	b.Position.Z += normalZ * correction * inverseMassB
}
