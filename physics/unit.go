package physics

import (
	"image/color"
	"math"

	"github.com/EliCDavis/vector/vector3"
	"github.com/google/uuid"
)

type Unit struct {
	Id       uuid.UUID
	Position vector3.Vector[float64]
	//PreviousPosition vector3.Vector[float64]
	Velocity     vector3.Vector[float64]
	Acceleration vector3.Vector[float64]

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

func (u *Unit) accelerate(a vector3.Vector[float64]) {
	u.Acceleration = u.Acceleration.Add(a)
}

func newUnitWithPropertiesAtPosition(position, acceleration, velocity vector3.Vector[float64], radius, massMultiplier, elasticity float64, color color.RGBA) *Unit {
	unit := &Unit{
		Id:       uuid.New(),
		Position: position,
		//PreviousPosition: position,
		Velocity:       velocity,
		Acceleration:   acceleration,
		Radius:         radius,
		MassMultiplier: massMultiplier,
		Elasticity:     elasticity,
		Color:          color,
	}

	unit.Mass = unit.GetMass()

	return unit
}

func (u *Unit) UpdatePosition(dt float64) {
	// x := 2*u.Position.X() - u.PreviousPosition.X() + u.Acceleration.X()*dt*dt
	// y := 2*u.Position.Y() - u.PreviousPosition.Y() + u.Acceleration.Y()*dt*dt
	// z := 2*u.Position.Z() - u.PreviousPosition.Z() + u.Acceleration.Z()*dt*dt
	// newPosition := vector3.New(x, y, z)
	// u.PreviousPosition = u.Position
	// u.Position = newPosition

	nPos := u.Position.Add(u.Velocity.Scale(dt)).Add(u.Acceleration.Scale(0.5 * dt * dt))
	nVelocity := u.Velocity.Add(u.Acceleration.Scale(dt))

	u.Position = nPos
	u.Velocity = nVelocity
	u.Acceleration = vector3.Zero[float64]()
}
