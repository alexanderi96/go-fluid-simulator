package physics

import (
	"image/color"
	"math"

	"github.com/EliCDavis/vector/vector3"
	"github.com/google/uuid"
)

type Unit struct {
	Id               uuid.UUID
	Position         vector3.Vector[float64]
	PreviousPosition vector3.Vector[float64]
	Acceleration     vector3.Vector[float64]

	Elasticity     float64
	Radius         float64
	MassMultiplier float64
	Mass           float64
	Color          color.RGBA
}

func (u *Unit) GetVolume() float64 {
	// Calcola il volume della sfera (4/3 * π * r^3)
	return float64((4.0 / 3.0) * math.Pi * math.Pow(float64(u.Radius), 3))
}

func (u *Unit) GetMass() float64 {
	// Calcola la massa utilizzando il volume e il MassMultiplier
	return u.GetVolume() * u.MassMultiplier
}

func (u *Unit) GetVelocity() vector3.Vector[float64] {
	return vector3.New(
		u.Position.X()-u.PreviousPosition.X(),
		u.Position.Y()-u.PreviousPosition.Y(),
		u.Position.Z()-u.PreviousPosition.Z(),
	)
}

func (u *Unit) accelerate(a vector3.Vector[float64]) {
	u.Acceleration = u.Acceleration.Add(a)
}

func newUnitWithPropertiesAtPosition(position vector3.Vector[float64], acceleration vector3.Vector[float64], radius, massMultiplier, elasticity float64, color color.RGBA) *Unit {
	unit := &Unit{
		Id:               uuid.New(),
		Position:         position,
		PreviousPosition: position,
		Acceleration:     acceleration,
		Radius:           radius,
		MassMultiplier:   massMultiplier,
		Elasticity:       elasticity,
		Color:            color,
	}

	unit.Mass = unit.GetMass()

	return unit
}
