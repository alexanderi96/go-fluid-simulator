package physics

import (
	"image/color"
	"math"
	"sync"

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
	stefanBoltzmannConstant = 5.67e-8
	specificHeatCapacity    = 4186.0
	ambientTemperature      = 20.0
	heatTransferCoefficient = 0.1
	maxHeatTransferDistance = 10.0
	coolingRate             = 0.5
)

var (
	mat        = material.NewStandard(math32.NewColor("white"))
	overlapMat = material.NewStandard(math32.NewColor("red"))
	// Object pool for vector calculations
	vectorPool = sync.Pool{
		New: func() interface{} {
			return make([]float64, 3)
		},
	}
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
	HeatTransferred float64

	// Cache frequently used values
	volume      float64
	surfaceArea float64
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

	// Use vectorPool for distance calculation
	vec := vectorPool.Get().([]float64)
	defer vectorPool.Put(vec)

	vec[0] = u.Position.X() - other.Position.X()
	vec[1] = u.Position.Y() - other.Position.Y()
	vec[2] = u.Position.Z() - other.Position.Z()

	distanceSquared := vec[0]*vec[0] + vec[1]*vec[1] + vec[2]*vec[2]
	if distanceSquared > maxHeatTransferDistance*maxHeatTransferDistance {
		return
	}

	distance := math.Sqrt(distanceSquared)
	distanceFactor := 1.0 / (1.0 + distance)
	tempDiff := u.Heat - other.Heat
	if tempDiff <= 0 {
		return
	}

	heatTransfer := heatTransferCoefficient * distanceFactor * tempDiff * dt
	massRatio := math.Sqrt(other.Mass / u.Mass)
	heatTransfer *= massRatio
	heatTransfer = math.Min(heatTransfer, u.Heat-ambientTemperature)

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

	// Initialize cached values
	u.volume = (4.0 / 3.0) * math.Pi * math.Pow(u.Radius, 3)
	u.surfaceArea = 4.0 * math.Pi * math.Pow(u.Radius, 2)
	u.Mass = u.volume * u.MassMultiplier
}

func (u *Unit) GetVolume() float64 {
	return u.volume
}

func (u *Unit) GetSurfaceArea() float64 {
	return u.surfaceArea
}

func (u *Unit) GetMass() float64 {
	return u.Mass
}

func (u *Unit) ApplyForce(f vector3.Vector[float64]) {
	invMass := 1.0 / u.Mass
	u.Acceleration = u.Acceleration.Add(f.Scale(invMass))
}

func (u *Unit) UpdatePosition(dt float64) {
	// Update position and velocity using Verlet integration
	halfDtSq := 0.5 * dt * dt
	newPos := u.Position.Add(u.Velocity.Scale(dt)).Add(u.Acceleration.Scale(halfDtSq))
	newVel := u.Velocity.Add(u.Acceleration.Scale(dt))

	u.Position = newPos
	u.Velocity = newVel
	u.Acceleration = vector3.Zero[float64]()

	u.Mesh.SetPosition(float32(u.Position.X()), float32(u.Position.Y()), float32(u.Position.Z()))

	if u.Heat > ambientTemperature {
		cooldown := (u.Heat - ambientTemperature) * coolingRate * dt
		u.Heat = math.Max(ambientTemperature, u.Heat-cooldown)

		normalizedTemp := math32.Clamp(float32((u.Heat-ambientTemperature)/50), 0, 1)
		normalizedRadius := math32.Clamp(float32(u.Radius/5e3), 0, 1)
		normalizedMass := math32.Clamp(float32(u.Mass/1e5), 0, 1)
		normalizedElasticity := float32(u.Elasticity)

		baseColor := math32.Color{
			R: normalizedRadius*0.8 + normalizedMass*0.2,
			G: normalizedElasticity*0.7 + normalizedRadius*0.3,
			B: (1.0-normalizedMass)*0.6 + (1.0-normalizedElasticity)*0.4,
		}

		heatColor := math32.Color{
			R: 1.0,
			G: math32.Clamp(normalizedTemp*0.6, 0, 0.6),
			B: math32.Clamp(normalizedTemp*0.3, 0, 0.3),
		}

		finalColor := math32.Color{
			R: baseColor.R*(1-normalizedTemp) + heatColor.R*normalizedTemp,
			G: baseColor.G*(1-normalizedTemp) + heatColor.G*normalizedTemp,
			B: baseColor.B*(1-normalizedTemp) + heatColor.B*normalizedTemp,
		}

		mat := u.Mesh.Mesh.GetMaterial(0).(*material.Standard)
		mat.SetColor(&finalColor)

		intensity := math32.Clamp(float32((u.Heat-ambientTemperature)/50), 0.1, 2.0)
		u.Mesh.Light.SetColor(&finalColor)
		u.Mesh.Light.SetIntensity(intensity)
	} else {
		u.Heat = ambientTemperature
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
	var collided bool
	vec := vectorPool.Get().([]float64)
	defer vectorPool.Put(vec)

	vec[0] = unit.Position.X()
	vec[1] = unit.Position.Y()
	vec[2] = unit.Position.Z()

	vx, vy, vz := unit.Velocity.X(), unit.Velocity.Y(), unit.Velocity.Z()

	// X-axis collision
	if vec[0]-unit.Radius < wallBounds.Min.X() {
		vec[0] = wallBounds.Min.X() + unit.Radius
		vx = -vx * wallElasticity
		collided = true
	} else if vec[0]+unit.Radius > wallBounds.Max.X() {
		vec[0] = wallBounds.Max.X() - unit.Radius
		vx = -vx * wallElasticity
		collided = true
	}

	// Y-axis collision
	if vec[1]-unit.Radius < wallBounds.Min.Y() {
		vec[1] = wallBounds.Min.Y() + unit.Radius
		vy = -vy * wallElasticity
		collided = true
	} else if vec[1]+unit.Radius > wallBounds.Max.Y() {
		vec[1] = wallBounds.Max.Y() - unit.Radius
		vy = -vy * wallElasticity
		collided = true
	}

	// Z-axis collision
	if vec[2]-unit.Radius < wallBounds.Min.Z() {
		vec[2] = wallBounds.Min.Z() + unit.Radius
		vz = -vz * wallElasticity
		collided = true
	} else if vec[2]+unit.Radius > wallBounds.Max.Z() {
		vec[2] = wallBounds.Max.Z() - unit.Radius
		vz = -vz * wallElasticity
		collided = true
	}

	if collided {
		unit.Position = vector3.New(vec[0], vec[1], vec[2])
		unit.Velocity = vector3.New(vx, vy, vz)

		speed := math.Sqrt(vx*vx + vy*vy + vz*vz)
		heatGenerated := 0.5 * unit.Mass * speed * speed * (1 - wallElasticity) * 0.01
		unit.Heat += heatGenerated
	}

	return collided
}

func (u *Unit) GiveMassAndCenterOfMassForBounds(bounds BoundingBox) (vector3.Vector[float64], float64) {
	// Simplified calculation using bounding box intersection
	minX := math.Max(bounds.Min.X(), u.Position.X()-u.Radius)
	maxX := math.Min(bounds.Max.X(), u.Position.X()+u.Radius)
	minY := math.Max(bounds.Min.Y(), u.Position.Y()-u.Radius)
	maxY := math.Min(bounds.Max.Y(), u.Position.Y()+u.Radius)
	minZ := math.Max(bounds.Min.Z(), u.Position.Z()-u.Radius)
	maxZ := math.Min(bounds.Max.Z(), u.Position.Z()+u.Radius)

	// Calculate volume of intersection
	dx := maxX - minX
	dy := maxY - minY
	dz := maxZ - minZ

	if dx <= 0 || dy <= 0 || dz <= 0 {
		return vector3.Zero[float64](), 0
	}

	volume := dx * dy * dz
	totalVolume := u.GetVolume()

	if volume > totalVolume {
		volume = totalVolume
	}

	massRatio := volume / totalVolume
	mass := u.Mass * massRatio

	centerX := (minX + maxX) / 2
	centerY := (minY + maxY) / 2
	centerZ := (minZ + maxZ) / 2

	return vector3.New(centerX, centerY, centerZ), mass
}
