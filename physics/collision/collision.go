package collision

import (
	"math"

	"github.com/EliCDavis/vector/vector3"
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
}

// GatherCollisionData collects all necessary data for collision resolution
func GatherCollisionData(uA, uB Collidable) *CollisionData {
	distance := uA.GetPosition().Distance(uB.GetPosition())
	totalRadius := uA.GetRadius() + uB.GetRadius()

	collData := &CollisionData{
		UnitA:       uA,
		UnitB:       uB,
		Distance:    distance,
		TotalRadius: totalRadius,
		TotalMass:   uA.GetMass() + uB.GetMass(),
	}

	collData.Collided = distance < totalRadius

	if collData.Collided {
		collData.Elasticity = math.Min(uA.GetElasticity(), uB.GetElasticity())
		collData.ImpulseDirection = uA.GetPosition().Sub(uB.GetPosition()).Normalized()
		collData.RelativeVelocity = uA.GetVelocity().Sub(uB.GetVelocity())
		collData.RelVelNormal = collData.RelativeVelocity.Dot(collData.ImpulseDirection)
	}

	return collData
}

// ResolveCollision handles the physics of collision between two objects
func ResolveCollision(collData *CollisionData) {
	// Calculate impulse magnitude
	impulseMag := -(1 + collData.Elasticity) * collData.RelVelNormal
	impulseMag /= (1/collData.UnitA.GetMass() + 1/collData.UnitB.GetMass())

	// Calculate impulse vector
	impulse := collData.ImpulseDirection.Scale(impulseMag)

	// Apply velocities
	newVelA := collData.UnitA.GetVelocity().Add(impulse.Scale(1 / collData.UnitA.GetMass()))
	newVelB := collData.UnitB.GetVelocity().Sub(impulse.Scale(1 / collData.UnitB.GetMass()))

	collData.UnitA.SetVelocity(newVelA)
	collData.UnitB.SetVelocity(newVelB)

	// Resolve position overlap
	overlap := collData.TotalRadius - collData.Distance
	moveDistanceA := (collData.UnitB.GetMass() / collData.TotalMass) * overlap
	moveDistanceB := (collData.UnitA.GetMass() / collData.TotalMass) * overlap

	normalMoveA := collData.ImpulseDirection.Scale(moveDistanceA)
	normalMoveB := collData.ImpulseDirection.Scale(moveDistanceB)

	newPosA := collData.UnitA.GetPosition().Add(normalMoveA)
	newPosB := collData.UnitB.GetPosition().Sub(normalMoveB)

	collData.UnitA.SetPosition(newPosA)
	collData.UnitB.SetPosition(newPosB)

	// Handle heat transfer - reduced intensity and scaled by relative velocity
	relativeSpeed := math.Abs(collData.RelVelNormal)
	heatTransfer := relativeSpeed * 0.05 // Reduced heat generation factor
	collData.UnitA.AddHeat(heatTransfer * (1.0 - collData.UnitA.GetElasticity()))
	collData.UnitB.AddHeat(heatTransfer * (1.0 - collData.UnitB.GetElasticity()))
}
