package physics

import (
	"image/color"
	"math"

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

	//Position vector3.Vector[float32]
	//PreviousPosition vector3.Vector[float32]
	Velocity     *math32.Vector3
	Acceleration *math32.Vector3

	Elasticity     float32
	Radius         float32
	MassMultiplier float32
	Mass           float32
	Color          color.RGBA

	Heat float32
}

func (u *Unit) GetVolume() float32 {
	// Calcola il volume della sfera (4/3 * π * r^3)
	return float32((4.0 / 3.0) * math.Pi * math.Pow(float64(u.Radius), 3))
}

func (u *Unit) GetMass() float32 {
	// Calcola la massa utilizzando il volume e il MassMultiplier
	return u.GetVolume() * u.MassMultiplier
}

func (u *Unit) accelerate(f *math32.Vector3) {
	// log.Print("Before Acceleration: ", u.Acceleration)
	u.Acceleration = u.Acceleration.Add(f.MultiplyScalar(1 / u.Mass))
	// log.Print("After Acceleration: ", u.Acceleration)
}

func (u *Unit) UpdatePosition(dt float32) {

	u.Velocity.Add(u.Acceleration.MultiplyScalar(dt))

	nPos := u.Mesh.Position()

	nPos.Add(u.Velocity.MultiplyScalar(dt)).Add(u.Acceleration.MultiplyScalar(0.5 * dt * dt))

	// u.Mesh.SetPositionVec(&nPos)
	u.Mesh.SetPositionVec(math32.NewVector3(nPos.X, nPos.Y, nPos.Z))

	u.Acceleration = math32.NewVec3()
	if u.Heat > 0.0 {
		u.Heat -= 1
	} else {
		u.Heat = 0.0
	}
}

func (unit *Unit) CheckAndResolveWallCollision(wallBounds *math32.Box3, wallElasticity float32) bool {
	xCorrection, yCorrection, zCorrection := unit.Mesh.Position().X, unit.Mesh.Position().Y, unit.Mesh.Position().Z
	vxCorrection, vyCorrection, vzCorrection := unit.Velocity.X, unit.Velocity.Y, unit.Velocity.Z
	collided := false

	// Correzione asse X
	if unit.Mesh.Position().X-unit.Radius < wallBounds.Min.X {
		overlapX := wallBounds.Min.X - (unit.Mesh.Position().X - unit.Radius)
		xCorrection = unit.Mesh.Position().X + overlapX
		// Applica la restituzione
		vxCorrection = -unit.Velocity.X * wallElasticity
		//nVel.FlipX()
		collided = true
	}
	if unit.Mesh.Position().X+unit.Radius > wallBounds.Max.X {
		overlapX := (unit.Mesh.Position().X + unit.Radius) - wallBounds.Max.X
		xCorrection = unit.Mesh.Position().X - overlapX
		// Applica la restituzione
		vxCorrection = -unit.Velocity.X * wallElasticity
		//nVel.FlipX()
		collided = true
	}

	// Correzione asse Y
	if unit.Mesh.Position().Y-unit.Radius < wallBounds.Min.Y {
		overlapY := wallBounds.Min.Y - (unit.Mesh.Position().Y - unit.Radius)
		yCorrection = unit.Mesh.Position().Y + overlapY
		// Applica la restituzione
		vyCorrection = -unit.Velocity.Y * wallElasticity
		//nVel.FlipY()
		collided = true
	}
	if unit.Mesh.Position().Y+unit.Radius > wallBounds.Max.Y {
		overlapY := (unit.Mesh.Position().Y + unit.Radius) - wallBounds.Max.Y
		yCorrection = unit.Mesh.Position().Y - overlapY
		// Applica la restituzione
		vyCorrection = -unit.Velocity.Y * wallElasticity
		//nVel.FlipY()
		collided = true
	}

	// Correzione asse Z
	if unit.Mesh.Position().Z-unit.Radius < wallBounds.Min.Z {
		overlapZ := wallBounds.Min.Z - (unit.Mesh.Position().Z - unit.Radius)
		zCorrection = unit.Mesh.Position().Z + overlapZ
		// Applica la restituzione
		vzCorrection = -unit.Velocity.Z * wallElasticity
		//nVel.FlipZ()
		collided = true
	}
	if unit.Mesh.Position().Z+unit.Radius > wallBounds.Max.Z {
		overlapZ := (unit.Mesh.Position().Z + unit.Radius) - wallBounds.Max.Z
		zCorrection = unit.Mesh.Position().Z - overlapZ
		// Applica la restituzione
		vzCorrection = -unit.Velocity.Z * wallElasticity
		//nVel.FlipZ()
		collided = true
	}

	if collided {
		//log.Print("\nvel:", unit.Velocity.X(), unit.Velocity.Y(), unit.Velocity.Z())
		//log.Print("\nnVel: ", nVel.X(), nVel.Y(), nVel.Z())
		unit.Mesh.SetPosition(xCorrection, yCorrection, zCorrection)
		unit.Velocity = math32.NewVector3(vxCorrection, vyCorrection, vzCorrection)

		unit.Heat += 2
	}

	return collided

}
