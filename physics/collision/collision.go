package collision

import (
	"math"
	"sync"

	"github.com/EliCDavis/vector/vector3"
)

var (
	// Object pools for frequently allocated objects
	collisionDataPool = sync.Pool{
		New: func() interface{} { return &CollisionData{} },
	}
	vectorPool = sync.Pool{
		New: func() interface{} { return make([]float64, 3) },
	}
)

// CollisionData represents the data needed for collision resolution
type CollisionData struct {
	UnitA, UnitB     Collidable
	Distance         float64
	TotalRadius      float64
	TotalMass        float64
	Collided         bool
	ImpulseDirection vector3.Vector[float64]
	RelativeVelocity vector3.Vector[float64]
	RelVelNormal     float64
	Elasticity       float64
}

// Collidable defines the interface for objects that can collide
type Collidable interface {
	GetPosition() vector3.Vector[float64]
	GetVelocity() vector3.Vector[float64]
	GetRadius() float64
	GetMass() float64
	GetElasticity() float64
	SetPosition(pos vector3.Vector[float64])
	SetVelocity(vel vector3.Vector[float64])
	AddHeat(heat float64)
	CanBeAltered() bool
	Merge(other Collidable)
}

// GatherCollisionData collects all necessary data for collision resolution
func GatherCollisionData(uA, uB Collidable) *CollisionData {
	// Get collision data from pool
	collData := collisionDataPool.Get().(*CollisionData)
	vec := vectorPool.Get().([]float64)
	defer vectorPool.Put(vec)

	// Calculate distance using components to avoid vector allocation
	posA := uA.GetPosition()
	posB := uB.GetPosition()
	vec[0] = posA.X() - posB.X()
	vec[1] = posA.Y() - posB.Y()
	vec[2] = posA.Z() - posB.Z()

	distanceSquared := vec[0]*vec[0] + vec[1]*vec[1] + vec[2]*vec[2]
	distance := math.Sqrt(distanceSquared)

	totalRadius := uA.GetRadius() + uB.GetRadius()

	collData.UnitA = uA
	collData.UnitB = uB
	collData.Distance = distance
	collData.TotalRadius = totalRadius
	collData.TotalMass = uA.GetMass() + uB.GetMass()
	collData.Collided = distance < totalRadius

	if collData.Collided {
		collData.Elasticity = math.Min(uA.GetElasticity(), uB.GetElasticity())

		// Calculate impulse direction without vector allocation
		invDistance := 1.0 / distance
		collData.ImpulseDirection = vector3.New(
			vec[0]*invDistance,
			vec[1]*invDistance,
			vec[2]*invDistance,
		)

		// Calculate relative velocity components directly
		velA := uA.GetVelocity()
		velB := uB.GetVelocity()
		vec[0] = velA.X() - velB.X()
		vec[1] = velA.Y() - velB.Y()
		vec[2] = velA.Z() - velB.Z()
		collData.RelativeVelocity = vector3.New(vec[0], vec[1], vec[2])

		// Calculate dot product directly
		collData.RelVelNormal = vec[0]*collData.ImpulseDirection.X() +
			vec[1]*collData.ImpulseDirection.Y() +
			vec[2]*collData.ImpulseDirection.Z()
	}

	return collData
}

// ResolveCollision handles the physics of collision between two objects
func ResolveCollision(collData *CollisionData) {
	defer collisionDataPool.Put(collData)

	// Check if units can merge
	if collData.UnitA.CanBeAltered() && collData.UnitB.CanBeAltered() {
		massA := collData.UnitA.GetMass()
		massB := collData.UnitB.GetMass()

		if massA > massB {
			collData.UnitA.Merge(collData.UnitB)
			return
		} else if massB > massA {
			collData.UnitB.Merge(collData.UnitA)
			return
		}
	}

	// If units cannot merge or have equal mass, proceed with normal collision
	// Calculate impulse magnitude
	impulseMag := -(1 + collData.Elasticity) * collData.RelVelNormal
	impulseMag /= (1/collData.UnitA.GetMass() + 1/collData.UnitB.GetMass())

	// Get impulse components
	vec := vectorPool.Get().([]float64)
	defer vectorPool.Put(vec)

	vec[0] = collData.ImpulseDirection.X() * impulseMag
	vec[1] = collData.ImpulseDirection.Y() * impulseMag
	vec[2] = collData.ImpulseDirection.Z() * impulseMag

	// Calculate new velocities directly
	massInvA := 1 / collData.UnitA.GetMass()
	massInvB := 1 / collData.UnitB.GetMass()

	velA := collData.UnitA.GetVelocity()
	velB := collData.UnitB.GetVelocity()

	newVelA := vector3.New(
		velA.X()+vec[0]*massInvA,
		velA.Y()+vec[1]*massInvA,
		velA.Z()+vec[2]*massInvA,
	)

	newVelB := vector3.New(
		velB.X()-vec[0]*massInvB,
		velB.Y()-vec[1]*massInvB,
		velB.Z()-vec[2]*massInvB,
	)

	collData.UnitA.SetVelocity(newVelA)
	collData.UnitB.SetVelocity(newVelB)

	// Resolve position overlap
	overlap := collData.TotalRadius - collData.Distance
	if overlap > 0 {
		moveRatioA := collData.UnitB.GetMass() / collData.TotalMass
		moveRatioB := collData.UnitA.GetMass() / collData.TotalMass

		// Calculate position corrections directly
		posA := collData.UnitA.GetPosition()
		posB := collData.UnitB.GetPosition()

		moveA := overlap * moveRatioA
		moveB := overlap * moveRatioB

		newPosA := vector3.New(
			posA.X()+collData.ImpulseDirection.X()*moveA,
			posA.Y()+collData.ImpulseDirection.Y()*moveA,
			posA.Z()+collData.ImpulseDirection.Z()*moveA,
		)

		newPosB := vector3.New(
			posB.X()-collData.ImpulseDirection.X()*moveB,
			posB.Y()-collData.ImpulseDirection.Y()*moveB,
			posB.Z()-collData.ImpulseDirection.Z()*moveB,
		)

		collData.UnitA.SetPosition(newPosA)
		collData.UnitB.SetPosition(newPosB)

		// Calculate heat transfer based on collision energy
		relativeSpeed := math.Abs(collData.RelVelNormal)
		heatTransfer := relativeSpeed * relativeSpeed * 0.01 // Reduced heat generation factor

		if heatTransfer > 0.01 {
			elasticityLossA := 1.0 - collData.UnitA.GetElasticity()
			elasticityLossB := 1.0 - collData.UnitB.GetElasticity()

			if elasticityLossA > 0.01 {
				collData.UnitA.AddHeat(heatTransfer * elasticityLossA)
			}
			if elasticityLossB > 0.01 {
				collData.UnitB.AddHeat(heatTransfer * elasticityLossB)
			}
		}
	}
}
