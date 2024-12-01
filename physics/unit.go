package physics

import (
	"image/color"
	"math"

	"github.com/EliCDavis/vector/vector3"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
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
	Id uuid.UUID

	Mesh *PointLightMesh

	Position     vector3.Vector[float64]
	Velocity     vector3.Vector[float64]
	Acceleration vector3.Vector[float64]

	Elasticity     float64
	Radius         float64
	MassMultiplier float64
	Mass           float64
	Color          color.RGBA
	CanBeAltered   bool

	Heat float64
}

type PointLightMesh struct {
	*graphic.Mesh
	Light *light.Point
}

func (u *Unit) GetUnit() *Unit {
	return u
}

func (u *Unit) GetPosition() vector3.Vector[float64] {
	return u.Position
}

func (u *Unit) GetVelocity() vector3.Vector[float64] {
	return u.Velocity
}

func (u *Unit) GetElasticity() float64 {
	return u.Elasticity
}

func (u *Unit) GetRadius() float64 {
	return u.Radius
}

func (u *Unit) SetPosition(pos vector3.Vector[float64]) {
	u.Position = pos
}

func (u *Unit) SetVelocity(vel vector3.Vector[float64]) {
	u.Velocity = vel
}

func (u *Unit) AddHeat(heat float64) {
	u.Heat += heat
}

func (u *Unit) NewPointLightMesh() {
	// Base color based on size
	minSize := 0e1
	maxSize := 5e3
	normalizedSize := (float64(u.Radius) - minSize) / (maxSize - minSize)
	normalizedSize = float64(math32.Clamp(float32(normalizedSize), 0, 1))

	// Create base color from size
	baseColor := math32.Color{
		R: float32(normalizedSize),
		G: 0.0,
		B: float32(1.0 - normalizedSize),
	}

	// Create mesh
	u.Mesh = new(PointLightMesh)
	geom := geometry.NewSphere(float64(u.Radius), seg, seg)
	mat := material.NewStandard(&baseColor)
	u.Mesh.Mesh = graphic.NewMesh(geom, mat)
	u.Mesh.Mesh.SetVisible(true)

	// Create point light
	light := light.NewPoint(&math32.Color{R: 1, G: 0.7, B: 0.3}, 1.0)
	u.Mesh.Light = light
	u.Mesh.Add(light)
}

func (u *Unit) GetVolume() float64 {
	return (4.0 / 3.0) * math.Pi * math.Pow(u.Radius, 3)
}

func (u *Unit) GetMass() float64 {
	return u.GetVolume() * u.MassMultiplier
}

func (u *Unit) ApplyForce(f vector3.Vector[float64]) {
	u.Acceleration = u.Acceleration.Add(f.Scale(1 / u.Mass))
}

func (u *Unit) UpdatePosition(dt float64) {
	u.Position = u.Position.Add(u.Velocity.Scale(dt))
	u.Velocity = u.Velocity.Add(u.Acceleration.Scale(dt))

	u.Acceleration = vector3.Zero[float64]()

	// Update mesh position
	u.Mesh.SetPosition(u.Position.ToFloat32().X(), u.Position.ToFloat32().Y(), u.Position.ToFloat32().Z())

	// Update light based on heat
	if u.Heat > 0.0 {
		// Calculate light intensity based on heat and mass
		intensity := math.Min(5.0, u.Heat*u.Mass*0.001)

		// Calculate color based on heat (blackbody radiation approximation)
		// As heat increases: red -> orange -> yellow -> white
		heatColor := &math32.Color{R: 1, G: 0, B: 0}
		if u.Heat > 10 {
			heatColor.G = float32(math.Min(1.0, (u.Heat-10)/20))
		}
		if u.Heat > 30 {
			heatColor.B = float32(math.Min(1.0, (u.Heat-30)/20))
		}

		// Update light properties
		u.Mesh.Light.SetColor(heatColor)
		u.Mesh.Light.SetLinearDecay(1.0)
		u.Mesh.Light.SetQuadraticDecay(1.0)
		u.Mesh.Light.SetIntensity(float32(intensity))

		// Update mesh material color
		mat := u.Mesh.Mesh.GetMaterial(0).(*material.Standard)
		mat.SetColor(heatColor)

		// Decrease heat over time
		u.Heat -= 1
	} else {
		u.Heat = 0.0
		// Minimum light intensity when cold
		u.Mesh.Light.SetIntensity(0.1)

		// Reset to base color when cold
		minSize := 0e1
		maxSize := 5e3
		normalizedSize := (float64(u.Radius) - minSize) / (maxSize - minSize)
		normalizedSize = float64(math32.Clamp(float32(normalizedSize), 0, 1))
		baseColor := &math32.Color{
			R: float32(normalizedSize),
			G: 0.0,
			B: float32(1.0 - normalizedSize),
		}
		mat := u.Mesh.Mesh.GetMaterial(0).(*material.Standard)
		mat.SetColor(baseColor)
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
		vxCorrection = -unit.Velocity.X() * wallElasticity
		collided = true
	}
	if unit.Position.X()+unit.Radius > wallBounds.Max.X() {
		overlapX := (unit.Position.X() + unit.Radius) - wallBounds.Max.X()
		xCorrection = unit.Position.X() - overlapX
		vxCorrection = -unit.Velocity.X() * wallElasticity
		collided = true
	}

	// Correzione asse Y
	if unit.Position.Y()-unit.Radius < wallBounds.Min.Y() {
		overlapY := wallBounds.Min.Y() - (unit.Position.Y() - unit.Radius)
		yCorrection = unit.Position.Y() + overlapY
		vyCorrection = -unit.Velocity.Y() * wallElasticity
		collided = true
	}
	if unit.Position.Y()+unit.Radius > wallBounds.Max.Y() {
		overlapY := (unit.Position.Y() + unit.Radius) - wallBounds.Max.Y()
		yCorrection = unit.Position.Y() - overlapY
		vyCorrection = -unit.Velocity.Y() * wallElasticity
		collided = true
	}

	// Correzione asse Z
	if unit.Position.Z()-unit.Radius < wallBounds.Min.Z() {
		overlapZ := wallBounds.Min.Z() - (unit.Position.Z() - unit.Radius)
		zCorrection = unit.Position.Z() + overlapZ
		vzCorrection = -unit.Velocity.Z() * wallElasticity
		collided = true
	}
	if unit.Position.Z()+unit.Radius > wallBounds.Max.Z() {
		overlapZ := (unit.Position.Z() + unit.Radius) - wallBounds.Max.Z()
		zCorrection = unit.Position.Z() - overlapZ
		vzCorrection = -unit.Velocity.Z() * wallElasticity
		collided = true
	}

	if collided {
		unit.Position = vector3.New(xCorrection, yCorrection, zCorrection)
		unit.Velocity = vector3.New(vxCorrection, vyCorrection, vzCorrection)
		unit.Heat += 2
	}

	return collided
}

func (u *Unit) GiveMassAndCenterOfMassForBounds(bounds BoundingBox) (vector3.Vector[float64], float64) {
	points := make([]vector3.Vector[float64], 0)
	for phi := 0.0; phi < 2*math.Pi; phi += math.Pi / 10 {
		for theta := 0.0; theta < math.Pi; theta += math.Pi / 10 {
			x := u.Radius * math.Sin(theta) * math.Cos(phi)
			y := u.Radius * math.Sin(theta) * math.Sin(phi)
			z := u.Radius * math.Cos(theta)
			point := vector3.New(x, y, z).Add(u.Position)
			points = append(points, point)
		}
	}

	var totalMass float64
	centerOfMass := vector3.Zero[float64]()
	for _, point := range points {
		if point.X() >= bounds.Min.X() && point.X() <= bounds.Max.X() &&
			point.Y() >= bounds.Min.Y() && point.Y() <= bounds.Max.Y() &&
			point.Z() >= bounds.Min.Z() && point.Z() <= bounds.Max.Z() {
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
