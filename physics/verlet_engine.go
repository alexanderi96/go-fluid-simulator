package physics

import (
	"math"
	"image/color"

	"github.com/alexanderi96/go-fluid-simulator/config"
	// "github.com/alexanderi96/go-fluid-simulator/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
	

)

func (s *Simulation) UpdateWithVerletIntegration() error {
	// Calcola il numero di step
resolutionSteps := 3
fractionalFrametime := s.Metrics.Frametime / float32(resolutionSteps)

s.ClusterMasses = make(map[uuid.UUID]float32)

for step := 0; step < resolutionSteps; step++ {

	for _, unitA := range s.Fluid {
		if unitA == nil {
			continue
		}

		unitA.checkWallCollisionVerlet(s.Config, fractionalFrametime)

		if unitA.Cluster != nil {
			s.ClusterMasses[unitA.Cluster.Id] += unitA.Mass/float32(resolutionSteps)
		}

		nearestUnit := Unit{}
		nearestValidDistance := float32(math.Inf(-1))

		for _, unitB := range s.Fluid {
			if unitB == nil || unitA.Id == unitB.Id  {
				continue
			}

			surfaceDistance := getSurfaceDistance(unitA, unitB)

			if surfaceDistance < 0 {
				handleCollision(unitA, unitB, surfaceDistance, fractionalFrametime)
			}

			if s.Config.UnitsEmitGravity {
				applyGravitationalAttraction(unitA, unitB, s.Config)
			}
			
			if (nearestUnit == Unit{} || (nearestUnit != *unitB && surfaceDistance < nearestValidDistance)) {
				nearestUnit = *unitB
				nearestValidDistance = surfaceDistance
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

func getEuclidianDistance(unitA, unitB *Unit) float32 {
	dx := unitB.Position.X - unitA.Position.X
	dy := unitB.Position.Y - unitA.Position.Y
	dz := unitB.Position.Z - unitA.Position.Z
	distanceSquared := dx*dx + dy*dy + dz*dz
	return float32(math.Sqrt(float64(distanceSquared)))
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
	if distance + totalRadius <= 0 {
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

	a.Position.X -= normalX * correction * inverseMassA
	a.Position.Y -= normalY * correction * inverseMassA
	a.Position.Z -= normalZ * correction * inverseMassA
	b.Position.X += normalX * correction * inverseMassB
	b.Position.Y += normalY * correction * inverseMassB
	b.Position.Z += normalZ * correction * inverseMassB
}

func handleCollisionAcceleration(a, b *Unit, surfaceDistance, dt float32) {
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

    // Calcola la forza di correzione basata sulla sovrapposizione
    // Assumiamo una costante k per la forza elastica (può essere adattata per il tuo caso specifico)
    k := float32(100.0) // Costante elastica
    forceMagnitude := k * overlap

    // Calcola l'accelerazione basata sulla forza di correzione
    // e la massa delle unità
    accelerationX := forceMagnitude * normalX
    accelerationY := forceMagnitude * normalY
    accelerationZ := forceMagnitude * normalZ

    // Applica l'accelerazione alle unità
    // L'accelerazione viene aggiunta perché assumiamo che la forza di correzione agisca in aggiunta alle forze esistenti
    a.Acceleration.X -= accelerationX / a.Mass
    a.Acceleration.Y -= accelerationY / a.Mass
    a.Acceleration.Z -= accelerationZ / a.Mass
    b.Acceleration.X += accelerationX / b.Mass
    b.Acceleration.Y += accelerationY / b.Mass
    b.Acceleration.Z += accelerationZ / b.Mass
}



func (u *Unit) updatePositionWithVerlet(dt float32) {
	newPosition := rl.Vector3{}
	newPosition.X = 2*u.Position.X - u.PreviousPosition.X + u.Acceleration.X*dt*dt
	newPosition.Y = 2*u.Position.Y - u.PreviousPosition.Y + u.Acceleration.Y*dt*dt
	newPosition.Z = 2*u.Position.Z - u.PreviousPosition.Z + u.Acceleration.Z*dt*dt

	u.PreviousPosition = u.Position
	u.Position = newPosition
	u.Acceleration = rl.Vector3{X: 0, Y: 0, Z: 0}
}

func (u *Unit) checkWallCollisionVerlet(cfg *config.Config, deltaTime float32) {
	// Calcola la velocità
	velocity := u.Velocity(deltaTime)

	// Correzione asse X
	if u.Position.X-u.Radius < 0 {
		overlapX := u.Radius - u.Position.X
		u.Position.X += overlapX
		// Applica la restituzione
		velocity.X = -velocity.X * cfg.WallElasticity
		u.PreviousPosition.X = u.Position.X - velocity.X*deltaTime
	} else if u.Position.X+u.Radius > float32(cfg.GameX) {
		overlapX := (u.Position.X + u.Radius) - float32(cfg.GameX)
		u.Position.X -= overlapX
		// Applica la restituzione
		velocity.X = -velocity.X * cfg.WallElasticity
		u.PreviousPosition.X = u.Position.X - velocity.X*deltaTime
	}

	// Correzione asse Y
	if u.Position.Y-u.Radius < 0 {
		overlapY := u.Radius - u.Position.Y
		u.Position.Y += overlapY
		// Applica la restituzione
		velocity.Y = -velocity.Y * cfg.WallElasticity
		u.PreviousPosition.Y = u.Position.Y - velocity.Y*deltaTime
	} else if u.Position.Y+u.Radius > float32(cfg.GameY) {
		overlapY := (u.Position.Y + u.Radius) - float32(cfg.GameY)
		u.Position.Y -= overlapY
		// Applica la restituzione
		velocity.Y = -velocity.Y * cfg.WallElasticity
		u.PreviousPosition.Y = u.Position.Y - velocity.Y*deltaTime
	}

	// Correzione asse Y
	if u.Position.Z-u.Radius < 0 {
		overlapZ := u.Radius - u.Position.Z
		u.Position.Z += overlapZ
		// Applica la restituzione
		velocity.Z = -velocity.Z * cfg.WallElasticity
		u.PreviousPosition.Z = u.Position.Z - velocity.Z*deltaTime
	} else if u.Position.Z+u.Radius > float32(cfg.GameZ) {
		overlapZ := (u.Position.Z + u.Radius) - float32(cfg.GameZ)
		u.Position.Z -= overlapZ
		// Applica la restituzione
		velocity.Z = -velocity.Z * cfg.WallElasticity
		u.PreviousPosition.Z = u.Position.Z - velocity.Z*deltaTime
	}
}

func distanceBetween(p1, p2 rl.Vector3) float32 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	dz := p2.Z - p1.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

func(u *Unit) update(applyGravity bool, gravity, dt float32) {

	if applyGravity {
		u.accelerate(rl.Vector3{X: 0, Y: -gravity, Z: 0})
	}
	u.updatePositionWithVerlet(dt)


	if u.TransitionTimer < u.TransitionDuration && u.Cluster != nil {
		u.TransitionTimer += dt
		if u.TransitionTimer > u.TransitionDuration {
			u.TransitionTimer = u.TransitionDuration
		}
	} else if u.Cluster == nil && u.TransitionTimer > 0 {
		u.TransitionTimer -= dt
		if u.TransitionTimer < 0 {
			u.TransitionTimer = 0
			u.OldCluster = &Cluster{}
		}
	}
}

func newCluster(unitA, unitB *Unit) *Cluster {
	return &Cluster{
		Id:             uuid.New(),
		Color:          blendedColorBasedOnMasses(unitA, unitB),
	}
}

func blendedColorBasedOnMasses(unitA, unitB *Unit) color.RGBA {
	// Assicurati che t sia compreso tra 0 e 1
	t := unitA.Mass / unitB.Mass

	unitAColor := unitA.BlendedColor()
	unitBColor := unitB.BlendedColor()
    if t <= 0 {
        return unitAColor
    } else if t > 1 {
		return unitBColor
    }

    // Calcola i componenti del nuovo colore interpolando linearmente tra color1 e clusterColor
    r := float32(unitAColor.R) + (float32(unitBColor.R) - float32(unitAColor.R)) * t
    g := float32(unitAColor.G) + (float32(unitBColor.G) - float32(unitAColor.G)) * t
    b := float32(unitAColor.B) + (float32(unitBColor.B) - float32(unitAColor.B)) * t
    a := float32(unitAColor.A) + (float32(unitBColor.A) - float32(unitAColor.A)) * t

	return rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a))
}

func updateClusters(unitA, unitB *Unit, clusterMasses map[uuid.UUID]float32) {
	if unitA.Cluster != nil && unitB.Cluster != nil && unitA.Cluster.Id != unitB.Cluster.Id {
		if clusterMasses[unitA.Cluster.Id] < clusterMasses[unitB.Cluster.Id] {
			unitA.TransitionTimer = 0
			unitA.OldCluster = unitA.Cluster
			unitA.Cluster = unitB.Cluster
		} else {
			unitB.TransitionTimer = 0
			unitB.OldCluster = unitB.Cluster
			unitB.Cluster = unitA.Cluster
		}
	} else if unitA.Cluster == nil && unitB.Cluster != nil {
		unitA.TransitionTimer = 0
		unitA.Cluster = unitB.Cluster
	 } else if unitB.Cluster == nil && unitA.Cluster != nil {
		unitB.TransitionTimer = 0
		unitB.Cluster = unitA.Cluster
	} else if unitA.Cluster == nil && unitB.Cluster == nil {
		c := *newCluster(unitA, unitB)
		unitA.Cluster = &c
		unitB.Cluster = &c
	}
}

func(s *Simulation) GetClusterMass(clusterId uuid.UUID) float32 {
	mass := float32(0)
	for _, unit := range s.Fluid {
		if unit.Cluster == nil {
			continue
		}
		if unit.Cluster.Id != clusterId {
			continue
		}
		mass += unit.Mass
	}
	return mass
}
