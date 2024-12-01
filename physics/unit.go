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
	seg                     = 10
	stefanBoltzmannConstant = 5.67e-8 // Stefan-Boltzmann constant for radiation
	specificHeatCapacity    = 4186.0  // Specific heat capacity (similar to water)
	ambientTemperature      = 20.0    // Ambient temperature in Celsius
	heatTransferCoefficient = 0.1     // Coefficient for heat transfer between units
	maxHeatTransferDistance = 10.0    // Maximum distance for heat transfer between units
	coolingRate             = 0.5     // Increased cooling rate for faster return to ambient temperature
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

	Heat            float64
	HeatTransferred float64 // Track heat transferred in current frame
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

func (u *Unit) TransferHeatTo(other *Unit, dt float64) {
	if u.Heat <= ambientTemperature {
		return
	}

	distance := u.Position.Sub(other.Position).Length()
	if distance > maxHeatTransferDistance {
		return
	}

	distanceFactor := 1.0 / (1.0 + distance*distance)
	tempDiff := u.Heat - other.Heat
	if tempDiff <= 0 {
		return
	}

	heatTransfer := heatTransferCoefficient * distanceFactor * tempDiff * dt
	massRatio := math.Sqrt(other.Mass / u.Mass)
	heatTransfer *= massRatio
	heatTransfer = math.Min(heatTransfer, u.Heat-ambientTemperature)

	u.HeatTransferred += heatTransfer
	other.HeatTransferred += heatTransfer

	u.Heat -= heatTransfer
	other.Heat += heatTransfer
}

func (u *Unit) NewPointLightMesh() {
	normalizedRadius := math32.Clamp(float32(u.Radius/5e3), 0, 1)
	normalizedMass := math32.Clamp(float32(u.Mass/1e5), 0, 1)
	normalizedElasticity := float32(u.Elasticity)

	baseColor := math32.Color{
		R: normalizedRadius*0.8 + normalizedMass*0.2,
		G: normalizedElasticity*0.7 + normalizedRadius*0.3,
		B: (1.0-normalizedMass)*0.6 + (1.0-normalizedElasticity)*0.4,
	}

	u.Mesh = new(PointLightMesh)
	geom := geometry.NewSphere(float64(u.Radius), seg, seg)
	mat := material.NewStandard(&baseColor)
	u.Mesh.Mesh = graphic.NewMesh(geom, mat)
	u.Mesh.Mesh.SetVisible(true)

	light := light.NewPoint(&baseColor, 1.0)
	u.Mesh.Light = light
	u.Mesh.Add(light)
}

func (u *Unit) GetVolume() float64 {
	return (4.0 / 3.0) * math.Pi * math.Pow(u.Radius, 3)
}

func (u *Unit) GetSurfaceArea() float64 {
	return 4.0 * math.Pi * math.Pow(u.Radius, 2)
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
	u.HeatTransferred = 0

	u.Mesh.SetPosition(u.Position.ToFloat32().X(), u.Position.ToFloat32().Y(), u.Position.ToFloat32().Z())

	// Enhanced cooling calculation
	if u.Heat > ambientTemperature {
		// Faster cooling rate
		cooldown := (u.Heat - ambientTemperature) * coolingRate * dt
		u.Heat = math.Max(ambientTemperature, u.Heat-cooldown)

		// Color transition based on temperature
		normalizedTemp := math32.Clamp(float32((u.Heat-ambientTemperature)/50), 0, 1) // Reduced temperature range

		// Get base color
		normalizedRadius := math32.Clamp(float32(u.Radius/5e3), 0, 1)
		normalizedMass := math32.Clamp(float32(u.Mass/1e5), 0, 1)
		normalizedElasticity := float32(u.Elasticity)

		baseColor := math32.Color{
			R: normalizedRadius*0.8 + normalizedMass*0.2,
			G: normalizedElasticity*0.7 + normalizedRadius*0.3,
			B: (1.0-normalizedMass)*0.6 + (1.0-normalizedElasticity)*0.4,
		}

		// Blend between base color and heat color
		heatColor := math32.Color{
			R: 1.0,
			G: math32.Clamp(normalizedTemp*0.6, 0, 0.6), // Reduced green component
			B: math32.Clamp(normalizedTemp*0.3, 0, 0.3), // Reduced blue component
		}

		finalColor := math32.Color{
			R: baseColor.R*(1-normalizedTemp) + heatColor.R*normalizedTemp,
			G: baseColor.G*(1-normalizedTemp) + heatColor.G*normalizedTemp,
			B: baseColor.B*(1-normalizedTemp) + heatColor.B*normalizedTemp,
		}

		mat := u.Mesh.Mesh.GetMaterial(0).(*material.Standard)
		mat.SetColor(&finalColor)

		// Adjust light intensity based on temperature
		intensity := math32.Clamp(float32((u.Heat-ambientTemperature)/50), 0.1, 2.0) // Reduced intensity range
		u.Mesh.Light.SetColor(&finalColor)
		u.Mesh.Light.SetIntensity(intensity)
	} else {
		u.Heat = ambientTemperature
		// Reset to base color
		normalizedRadius := math32.Clamp(float32(u.Radius/5e3), 0, 1)
		normalizedMass := math32.Clamp(float32(u.Mass/1e5), 0, 1)
		normalizedElasticity := float32(u.Elasticity)

		baseColor := &math32.Color{
			R: normalizedRadius*0.8 + normalizedMass*0.2,
			G: normalizedElasticity*0.7 + normalizedRadius*0.3,
			B: (1.0-normalizedMass)*0.6 + (1.0-normalizedElasticity)*0.4,
		}

		mat := u.Mesh.Mesh.GetMaterial(0).(*material.Standard)
		mat.SetColor(baseColor)
		u.Mesh.Light.SetIntensity(0.1)
		u.Mesh.Light.SetColor(baseColor)
	}
}

func (unit *Unit) CheckAndResolveWallCollision(wallBounds BoundingBox, wallElasticity float64) bool {
	xCorrection, yCorrection, zCorrection := unit.Position.X(), unit.Position.Y(), unit.Position.Z()
	vxCorrection, vyCorrection, vzCorrection := unit.Velocity.X(), unit.Velocity.Y(), unit.Velocity.Z()
	collided := false

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

		// Reduced heat generation from wall collisions
		velocityMagnitude := math.Sqrt(math.Pow(unit.Velocity.X(), 2) + math.Pow(unit.Velocity.Y(), 2) + math.Pow(unit.Velocity.Z(), 2))
		heatGenerated := 0.5 * unit.Mass * velocityMagnitude * velocityMagnitude * (1 - wallElasticity)
		unit.Heat += heatGenerated * 0.01 // Reduced heat generation factor
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
