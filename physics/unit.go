package physics

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"sync"

	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/physics/collision"
	"github.com/alexanderi96/go-fluid-simulator/physics/constants"
	"github.com/alexanderi96/go-fluid-simulator/physics/material"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	g3nmat "github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/google/uuid"
)

var (
	mat        = g3nmat.NewStandard(math32.NewColor("white"))
	overlapMat = g3nmat.NewStandard(math32.NewColor("red"))
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

	_position    vector3.Vector[float64]
	_velocity    vector3.Vector[float64]
	Acceleration vector3.Vector[float64]

	Composition  *material.Composition
	_radius      float64
	_mass        float64
	Color        color.RGBA
	canBeAltered bool
	isMerged     bool // Track if this unit has been merged into another

	Heat            float64
	HeatTransferred float64

	// Cache frequently used values
	volume               float64
	surfaceArea          float64
	_elasticity          float64
	emissivity           float64
	thermalConductivity  float64
	specificHeatCapacity float64
}

type PointLightMesh struct {
	*graphic.Mesh
	Light *light.Point
}

func (u *Unit) Unit() *Unit {
	return u
}

func (u *Unit) CanBeAltered() bool {
	return u.canBeAltered && !u.isMerged
}

// Merge combines this unit with another unit, absorbing its mass and properties
func (u *Unit) Merge(other collision.Collidable) {
	// Get the actual volume of the other unit
	var otherVolume float64
	if otherUnit, ok := other.(*Unit); ok {
		otherVolume = otherUnit.volume
	} else {
		// Fallback to sphere volume calculation if not a Unit
		r := other.Radius()
		otherVolume = (4.0 / 3.0) * math.Pi * r * r * r
	}

	// Calculate new mass and total volume
	newMass := u._mass + other.Mass()
	newVolume := u.volume + otherVolume

	// Calculate new radius based on total volume
	newRadius := math.Pow((3.0*newVolume)/(4.0*math.Pi), 1.0/3.0)

	// Calculate new position as mass-weighted average
	newPosition := u._position.Scale(u._mass).Add(other.Position().Scale(other.Mass())).Scale(1.0 / newMass)

	// Calculate resultant velocity based on conservation of momentum
	// p = mv, total momentum = m1v1 + m2v2 = (m1+m2)v_final
	newVelocity := u._velocity.Scale(u._mass).Add(other.Velocity().Scale(other.Mass())).Scale(1.0 / newMass)

	// Calculate combined heat based on mass-weighted average
	totalHeat := u.Heat*u._mass + 0.0
	if otherUnit, ok := other.(*Unit); ok {
		totalHeat += otherUnit.Heat * otherUnit._mass
	}
	combinedHeat := totalHeat / newMass

	// Update properties
	u._mass = newMass
	u._radius = newRadius
	u._position = newPosition
	u._velocity = newVelocity
	u.volume = newVolume  // Update cached volume
	u.Heat = combinedHeat // Set combined heat

	// Update mesh geometry with new radius
	if u.Mesh != nil {
		// Create new sphere geometry with updated radius
		geom := geometry.NewSphere(float64(u._radius), Segments, Segments)
		// Get color from composition
		baseColorRGB := u.Composition.GetColor()
		baseColor := &math32.Color{
			R: float32(baseColorRGB[0]),
			G: float32(baseColorRGB[1]),
			B: float32(baseColorRGB[2]),
		}

		// Create new mesh with updated geometry
		newMat := g3nmat.NewStandard(baseColor)
		u.Mesh.Mesh = graphic.NewMesh(geom, newMat)
		u.Mesh.Mesh.SetVisible(true)

		// Keep the same light but update its position and color
		if u.Mesh.Light != nil {
			u.Mesh.Light.SetColor(baseColor)
			u.Mesh.Add(u.Mesh.Light)
		}
	} else {
		// If no mesh exists, create a new one
		u.NewPointLightMesh()
	}

	// Mark the other unit as merged and clean up its visual representation
	if otherUnit, ok := other.(*Unit); ok {
		if otherUnit.Mesh != nil {
			otherUnit.Mesh.SetVisible(false)
			otherUnit.Mesh.Light = nil
		}
		otherUnit.isMerged = true
		otherUnit.canBeAltered = false
	}
}

// Implementazione dell'interfaccia Collidable
func (u *Unit) Position() vector3.Vector[float64] {
	return u._position
}

func (u *Unit) Velocity() vector3.Vector[float64] {
	return u._velocity
}

func (u *Unit) Elasticity() float64 {
	return u._elasticity
}

func (u *Unit) Radius() float64 {
	return u._radius
}

func (u *Unit) Mass() float64 {
	return u._mass
}

func (u *Unit) SetPosition(pos vector3.Vector[float64]) {
	u._position = pos
}

func (u *Unit) SetVelocity(vel vector3.Vector[float64]) {
	u._velocity = vel
}

func (u *Unit) AddHeat(heat float64) {
	u.Heat += heat
}

func (u *Unit) TransferHeatTo(other *Unit, dt float64) {
	if u.Heat <= AmbientTemperature {
		return
	}

	// Use vectorPool for distance calculation
	vec := vectorPool.Get().([]float64)
	defer vectorPool.Put(vec)

	vec[0] = u._position.X() - other._position.X()
	vec[1] = u._position.Y() - other._position.Y()
	vec[2] = u._position.Z() - other._position.Z()

	distanceSquared := vec[0]*vec[0] + vec[1]*vec[1] + vec[2]*vec[2]
	if distanceSquared > MaxHeatTransferDistance*MaxHeatTransferDistance {
		return
	}

	distance := math.Sqrt(distanceSquared)
	distanceFactor := 1.0 / (1.0 + distance)
	tempDiff := u.Heat - other.Heat
	if tempDiff <= 0 {
		return
	}

	// Use material properties for heat transfer
	conductivity := (u.thermalConductivity + other.thermalConductivity) / 2
	heatTransfer := conductivity * distanceFactor * tempDiff * dt
	massRatio := math.Sqrt(other._mass / u._mass)
	heatTransfer *= massRatio
	heatTransfer = math.Min(heatTransfer, u.Heat-AmbientTemperature)

	u.Heat -= heatTransfer
	other.Heat += heatTransfer
}

func (u *Unit) NewPointLightMesh() {
	// Get base color from composition
	baseColorRGB := u.Composition.GetColor()
	baseColor := math32.Color{
		R: float32(baseColorRGB[0]),
		G: float32(baseColorRGB[1]),
		B: float32(baseColorRGB[2]),
	}

	u.Mesh = new(PointLightMesh)
	geom := geometry.NewSphere(float64(u._radius), Segments, Segments)
	mat := g3nmat.NewStandard(&baseColor)
	u.Mesh.Mesh = graphic.NewMesh(geom, mat)
	u.Mesh.Mesh.SetVisible(true)

	light := light.NewPoint(&baseColor, 1.0)
	u.Mesh.Light = light
	u.Mesh.Add(light)

	// Initialize cached values
	u.volume = (4.0 / 3.0) * math.Pi * math.Pow(u._radius, 3)
	u.surfaceArea = 4.0 * math.Pi * math.Pow(u._radius, 2)

	// Get material properties
	density, specificHeat, thermalConductivity, emissivity, elasticity := u.Composition.GetEffectiveProperties()
	// Only calculate mass if it hasn't been set (i.e., not during merge)
	if u._mass == 0 {
		u._mass = u.volume * density
	}
	u.specificHeatCapacity = specificHeat
	u.thermalConductivity = thermalConductivity
	u.emissivity = emissivity
	u._elasticity = elasticity
}

func (u *Unit) GetVolume() float64 {
	return u.volume
}

func (u *Unit) GetSurfaceArea() float64 {
	return u.surfaceArea
}

func (u *Unit) GetMass() float64 {
	return u._mass
}

func (u *Unit) ApplyForce(f vector3.Vector[float64]) {
	invMass := 1.0 / u._mass
	u.Acceleration = u.Acceleration.Add(f.Scale(invMass))
}

func (u *Unit) UpdatePosition(dt float64) {
	log.Print("unit mass:", u._mass, " position: ", u._position)
	// Update position and velocity using Verlet integration
	halfDtSq := 0.5 * dt * dt
	newPos := u._position.Add(u._velocity.Scale(dt)).Add(u.Acceleration.Scale(halfDtSq))
	newVel := u._velocity.Add(u.Acceleration.Scale(dt))

	u._position = newPos
	u._velocity = newVel
	u.Acceleration = vector3.Zero[float64]()

	u.Mesh.SetPosition(float32(u._position.X()), float32(u._position.Y()), float32(u._position.Z()))

	if u.Heat > AmbientTemperature {
		cooldown := (u.Heat - AmbientTemperature) * CoolingRate * dt
		u.Heat = math.Max(AmbientTemperature, u.Heat-cooldown)

		normalizedTemp := math32.Clamp(float32((u.Heat-AmbientTemperature)/50), 0, 1)
		baseColorRGB := u.Composition.GetColor()
		baseColor := math32.Color{
			R: float32(baseColorRGB[0]),
			G: float32(baseColorRGB[1]),
			B: float32(baseColorRGB[2]),
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

		mat := u.Mesh.Mesh.GetMaterial(0).(*g3nmat.Standard)
		mat.SetColor(&finalColor)

		intensity := math32.Clamp(float32((u.Heat-AmbientTemperature)/50), 0.1, 2.0)
		u.Mesh.Light.SetColor(&finalColor)
		u.Mesh.Light.SetIntensity(intensity)
	} else {
		u.Heat = AmbientTemperature
		normalizedRadius := math32.Clamp(float32(u._radius/5e3), 0, 1)
		normalizedMass := math32.Clamp(float32(u._mass/1e5), 0, 1)
		normalizedElasticity := float32(u._elasticity)

		baseColor := &math32.Color{
			R: normalizedRadius*0.8 + normalizedMass*0.2,
			G: normalizedElasticity*0.7 + normalizedRadius*0.3,
			B: (1.0-normalizedMass)*0.6 + (1.0-normalizedElasticity)*0.4,
		}

		mat := u.Mesh.Mesh.GetMaterial(0).(*g3nmat.Standard)
		mat.SetColor(baseColor)

		// Calculate light intensity based on mass and radius
		// Using both mass and radius ensures larger, more massive objects emit more light
		baseIntensity := math32.Clamp(float32(u._mass/1e5)*float32(u._radius/5e3), 0.1, 2.0)
		u.Mesh.Light.SetIntensity(baseIntensity)
		u.Mesh.Light.SetColor(baseColor)
	}
}

func (unit *Unit) CheckAndResolveWallCollision(wallBounds BoundingBox, wallElasticity float64) bool {
	var collided bool
	vec := vectorPool.Get().([]float64)
	defer vectorPool.Put(vec)

	pos := unit.Position()
	vel := unit.Velocity()
	rad := unit.Radius()

	vec[0] = pos.X()
	vec[1] = pos.Y()
	vec[2] = pos.Z()

	vx, vy, vz := vel.X(), vel.Y(), vel.Z()

	// X-axis collision
	if vec[0]-rad < wallBounds.Min.X() {
		vec[0] = wallBounds.Min.X() + rad
		vx = -vx * wallElasticity
		collided = true
	} else if vec[0]+rad > wallBounds.Max.X() {
		vec[0] = wallBounds.Max.X() - rad
		vx = -vx * wallElasticity
		collided = true
	}

	// Y-axis collision
	if vec[1]-rad < wallBounds.Min.Y() {
		vec[1] = wallBounds.Min.Y() + rad
		vy = -vy * wallElasticity
		collided = true
	} else if vec[1]+rad > wallBounds.Max.Y() {
		vec[1] = wallBounds.Max.Y() - rad
		vy = -vy * wallElasticity
		collided = true
	}

	// Z-axis collision
	if vec[2]-rad < wallBounds.Min.Z() {
		vec[2] = wallBounds.Min.Z() + rad
		vz = -vz * wallElasticity
		collided = true
	} else if vec[2]+rad > wallBounds.Max.Z() {
		vec[2] = wallBounds.Max.Z() - rad
		vz = -vz * wallElasticity
		collided = true
	}

	if collided {
		unit.SetPosition(vector3.New(vec[0], vec[1], vec[2]))
		unit.SetVelocity(vector3.New(vx, vy, vz))

		speed := math.Sqrt(vx*vx + vy*vy + vz*vz)
		// Reduced heat generation from collisions and scaled by mass
		heatGenerated := 0.5 * unit.Mass() * speed * speed * (1 - wallElasticity) * 0.001
		// Scale heat increase based on current temperature to avoid spikes
		heatIncrease := heatGenerated * (1.0 - (unit.Heat-AmbientTemperature)/100.0)
		if heatIncrease > 0 {
			unit.Heat += heatIncrease
		}
	}

	return collided
}

// orbit calculates and sets the velocity needed for orbit around another unit
func (u *Unit) orbit(target *Unit) error {
	// Get current position vectors
	pos1 := u.Position()
	pos2 := target.Position()

	// Calculate direction vector from target to unit
	dir := pos1.Sub(pos2)

	// Check if units are at the same position
	if dir.X() == 0 && dir.Y() == 0 && dir.Z() == 0 {
		return fmt.Errorf("units cannot be at the same position for orbit")
	}

	// Calculate actual distance between bodies
	currentDistance := math.Sqrt(dir.X()*dir.X() + dir.Y()*dir.Y() + dir.Z()*dir.Z())

	// Calculate orbital velocity for circular orbit
	// v = sqrt(GM/(2r)) where:
	// G = gravitational constant
	// M = mass of the central body
	// r = orbital radius
	v := math.Sqrt((constants.G * target.Mass()) / (2 * currentDistance))

	// For a circular orbit in the XZ plane, the velocity should be perpendicular to the radius vector
	// and parallel to the XZ plane. Since the radius vector points from the Sun to Earth,
	// we want the velocity vector to be perpendicular to it in the XZ plane.

	// The velocity vector should be perpendicular to the radius vector in the XZ plane
	// If radius vector is (x,0,z), then velocity should be (-z,0,x) normalized
	velDir := vector3.New(-dir.Z(), 0, dir.X())
	velMag := math.Sqrt(velDir.X()*velDir.X() + velDir.Z()*velDir.Z())
	velDir = velDir.Scale(1.0 / velMag)

	// Set the orbital velocity
	u.SetVelocity(velDir.Scale(v))

	return nil
}

func (u *Unit) GiveMassAndCenterOfMassForBounds(bounds BoundingBox) (vector3.Vector[float64], float64) {
	// Simplified calculation using bounding box intersection
	pos := u.Position()
	rad := u.Radius()

	minX := math.Max(bounds.Min.X(), pos.X()-rad)
	maxX := math.Min(bounds.Max.X(), pos.X()+rad)
	minY := math.Max(bounds.Min.Y(), pos.Y()-rad)
	maxY := math.Min(bounds.Max.Y(), pos.Y()+rad)
	minZ := math.Max(bounds.Min.Z(), pos.Z()-rad)
	maxZ := math.Min(bounds.Max.Z(), pos.Z()+rad)

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
	mass := u.Mass() * massRatio

	centerX := (minX + maxX) / 2
	centerY := (minY + maxY) / 2
	centerZ := (minZ + maxZ) / 2

	return vector3.New(centerX, centerY, centerZ), mass
}
