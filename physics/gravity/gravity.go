package gravity

import (
	"github.com/EliCDavis/vector/vector3"
)

// UniversalGravitationalConstant is G in m^3 kg^-1 s^-2
const UniversalGravitationalConstant = 6.67430e-1

// Gravitable defines objects affected by gravity
type Gravitable interface {
	GetPosition() vector3.Vector[float64]
	GetMass() float64
}

// Calculator handles gravity calculations
type Calculator struct {
	Theta float64
}

// NewCalculator creates a new gravity calculator
func NewCalculator(theta float64) *Calculator {
	return &Calculator{
		Theta: theta,
	}
}

// CalculateForce calculates gravitational force between two objects
func CalculateForce(a, b Gravitable) vector3.Vector[float64] {
	deltaPos := b.GetPosition().Sub(a.GetPosition())

	// Calculate squared distance directly to avoid sqrt
	distanceSquared := deltaPos.X()*deltaPos.X() + deltaPos.Y()*deltaPos.Y() + deltaPos.Z()*deltaPos.Z()

	if distanceSquared <= 0 {
		return vector3.Zero[float64]()
	}

	// Pre-calculate mass product and constant
	massProduct := a.GetMass() * b.GetMass()
	forceMagnitude := UniversalGravitationalConstant * massProduct / distanceSquared

	// Avoid normalization by dividing by distance directly
	invDistance := 1.0 / distanceSquared
	return vector3.New(
		deltaPos.X()*forceMagnitude*invDistance,
		deltaPos.Y()*forceMagnitude*invDistance,
		deltaPos.Z()*forceMagnitude*invDistance,
	)
}
