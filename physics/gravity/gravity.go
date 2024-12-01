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
	distance := deltaPos.Length()

	if distance <= 0 {
		return vector3.Zero[float64]()
	}

	magnitude := UniversalGravitationalConstant * a.GetMass() * b.GetMass() / (distance * distance)
	return deltaPos.Normalized().Scale(magnitude)
}
