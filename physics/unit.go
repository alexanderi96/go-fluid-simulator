package physics

import (
	"image/color"
	"math"

	"github.com/EliCDavis/vector/vector3"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/google/uuid"
)

const (
	seg = 10
)

var (
	mat = material.NewStandard(math32.NewColor("Green"))
)

type Unit struct {
	Id   uuid.UUID
	Mesh *graphic.Mesh

	Position     vector3.Vector[float64]
	Velocity     vector3.Vector[float64]
	Acceleration vector3.Vector[float64]

	Elasticity     float64
	Radius         float64
	MassMultiplier float64
	Mass           float64
	Color          color.RGBA

	Heat float64
}

func (u *Unit) GetVolume() float64 {
	// Calcola il volume della sfera (4/3 * π * r^3)
	return (4.0 / 3.0) * math.Pi * math.Pow(u.Radius, 3)
}

func (u *Unit) GetMass() float64 {
	// Calcola la massa utilizzando il volume e il MassMultiplier
	return u.GetVolume() * u.MassMultiplier
}

func (u *Unit) accelerate(f vector3.Vector[float64]) {
	u.Acceleration = u.Acceleration.Add(f.Scale(1 / u.Mass))
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
		Heat:           0.0,
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
	u.Mesh.SetPosition(float32(nPos.X()), float32(nPos.Y()), float32(nPos.Z()))
	if u.Heat > 0.0 {
		u.Heat -= 1
	} else {
		u.Heat = 0.0
	}
}

func (unit *Unit) CheckAndResolveWallCollision(wallBounds BoundingBox, wallElasticity float64) bool {
	xCorrection, yCorrection, zCorrection := unit.Position.X(), unit.Position.Y(), unit.Position.Z()
	vxCorrection, vyCorrection, vzCorrection := unit.Velocity.X(), unit.Velocity.Y(), unit.Velocity.Z()
	collided := false

	// Correzione asse X
	if unit.Position.X()-unit.Radius < wallBounds.Min.X() {
		overlapX := wallBounds.Min.X() - (unit.Position.X() - unit.Radius)
		xCorrection = unit.Position.X() + overlapX
		// Applica la restituzione
		vxCorrection = -unit.Velocity.X() * wallElasticity
		//nVel.FlipX()
		collided = true
	}
	if unit.Position.X()+unit.Radius > wallBounds.Max.X() {
		overlapX := (unit.Position.X() + unit.Radius) - wallBounds.Max.X()
		xCorrection = unit.Position.X() - overlapX
		// Applica la restituzione
		vxCorrection = -unit.Velocity.X() * wallElasticity
		//nVel.FlipX()
		collided = true
	}

	// Correzione asse Y
	if unit.Position.Y()-unit.Radius < wallBounds.Min.Y() {
		overlapY := wallBounds.Min.Y() - (unit.Position.Y() - unit.Radius)
		yCorrection = unit.Position.Y() + overlapY
		// Applica la restituzione
		vyCorrection = -unit.Velocity.Y() * wallElasticity
		//nVel.FlipY()
		collided = true
	}
	if unit.Position.Y()+unit.Radius > wallBounds.Max.Y() {
		overlapY := (unit.Position.Y() + unit.Radius) - wallBounds.Max.Y()
		yCorrection = unit.Position.Y() - overlapY
		// Applica la restituzione
		vyCorrection = -unit.Velocity.Y() * wallElasticity
		//nVel.FlipY()
		collided = true
	}

	// Correzione asse Z
	if unit.Position.Z()-unit.Radius < wallBounds.Min.Z() {
		overlapZ := wallBounds.Min.Z() - (unit.Position.Z() - unit.Radius)
		zCorrection = unit.Position.Z() + overlapZ
		// Applica la restituzione
		vzCorrection = -unit.Velocity.Z() * wallElasticity
		//nVel.FlipZ()
		collided = true
	}
	if unit.Position.Z()+unit.Radius > wallBounds.Max.Z() {
		overlapZ := (unit.Position.Z() + unit.Radius) - wallBounds.Max.Z()
		zCorrection = unit.Position.Z() - overlapZ
		// Applica la restituzione
		vzCorrection = -unit.Velocity.Z() * wallElasticity
		//nVel.FlipZ()
		collided = true
	}

	if collided {
		//log.Print("\nvel:", unit.Velocity.X(), unit.Velocity.Y(), unit.Velocity.Z())
		//log.Print("\nnVel: ", nVel.X(), nVel.Y(), nVel.Z())
		unit.Position = vector3.New(xCorrection, yCorrection, zCorrection)
		unit.Velocity = vector3.New(vxCorrection, vyCorrection, vzCorrection)

		unit.Heat += 2
	}

	return collided

}
