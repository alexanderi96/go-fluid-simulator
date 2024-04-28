package physics

import (
	"image/color"
	"math"

	"github.com/EliCDavis/vector/vector3"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/google/uuid"
)

const (
	seg = 10
)

var (
	mat        = material.NewStandard(math32.NewColor("white"))
	overlapMat = material.NewStandard(math32.NewColor("red"))
)

type Unit struct {
	Id   uuid.UUID
	Mesh *graphic.Mesh `json:"-"`

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

func (u *Unit) GenerateMesh() {
	u.Mesh = graphic.NewMesh(geometry.NewSphere(float64(u.Radius), seg, seg), mat)
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

func (u *Unit) UpdatePosition(dt float64) {
	u.Position = u.Position.Add(u.Velocity.Scale(dt))
	u.Velocity = u.Velocity.Add(u.Acceleration.Scale(dt))

	u.Acceleration = vector3.Zero[float64]()

	u.Mesh.SetPosition(u.Position.ToFloat32().X(), u.Position.ToFloat32().Y(), u.Position.ToFloat32().Z())

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

// Calcola la massa parziale e il centro di massa di una porzione di sfera all'interno di una BoundingBox
func (u *Unit) GiveMassAndCenterOfMassForBounds(bounds BoundingBox) (vector3.Vector[float64], float64) {
	// Discretizza la sfera in punti (questo è un esempio molto semplice e non ottimizzato)
	points := make([]vector3.Vector[float64], 0)
	for phi := 0.0; phi < 2*math.Pi; phi += math.Pi / 10 { // Incremento arbitrario
		for theta := 0.0; theta < math.Pi; theta += math.Pi / 10 { // Incremento arbitrario
			x := u.Radius * math.Sin(theta) * math.Cos(phi)
			y := u.Radius * math.Sin(theta) * math.Sin(phi)
			z := u.Radius * math.Cos(theta)
			point := vector3.New(x, y, z).Add(u.Position)
			points = append(points, point)
		}
	}

	// Calcola la massa parziale e il centro di massa
	var totalMass float64
	centerOfMass := vector3.Zero[float64]()
	for _, point := range points {
		if point.X() >= bounds.Min.X() && point.X() <= bounds.Max.X() &&
			point.Y() >= bounds.Min.Y() && point.Y() <= bounds.Max.Y() &&
			point.Z() >= bounds.Min.Z() && point.Z() <= bounds.Max.Z() {
			// Assumi che ogni punto abbia una massa uguale (massa totale / numero di punti)
			pointMass := u.Mass / float64(len(points))
			totalMass += pointMass
			centerOfMass = centerOfMass.Add(point.Scale(pointMass))
		}
	}

	if totalMass > 0 {
		centerOfMass = centerOfMass.Scale(1 / totalMass)
	}

	return centerOfMass, totalMass
}
