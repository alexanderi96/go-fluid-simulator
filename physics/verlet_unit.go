package physics

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

type Unit struct {
	Id               uuid.UUID
	Position         rl.Vector3
	PreviousPosition rl.Vector3
	Acceleration     rl.Vector3
	Elasticity       float32
	Radius           float32
	MassMultiplier   float32
	Mass             float32
	Color            color.RGBA
}

func (u *Unit) GetVolume() float32 {
	// Calcola il volume della sfera (4/3 * π * r^3)
	return float32((4.0 / 3.0) * math.Pi * math.Pow(float64(u.Radius), 3))
}

func (u *Unit) GetMass() float32 {
	// Calcola la massa utilizzando il volume e il MassMultiplier
	return u.GetVolume() * u.MassMultiplier
}

func (u *Unit) GetVelocity(dt float32) rl.Vector3 {
	return rl.Vector3{
		X: (u.Position.X - u.PreviousPosition.X) / dt,
		Y: (u.Position.Y - u.PreviousPosition.Y) / dt,
		Z: (u.Position.Z - u.PreviousPosition.Z) / dt,
	}
}

func (u *Unit) accelerate(a rl.Vector3) {
	u.Acceleration.X += a.X
	u.Acceleration.Y += a.Y
	u.Acceleration.Z += a.Z
}

func newUnitWithPropertiesAtPosition(position rl.Vector3, acceleration rl.Vector3, radius, massMultiplier, elasticity float32, color color.RGBA) *Unit {
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
