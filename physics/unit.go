package physics

import (
	"fmt"
	"image/color"
	"math"

	"github.com/EliCDavis/vector/vector3"
	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/physics/collision"
	"github.com/alexanderi96/go-fluid-simulator/physics/material"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	g3nmat "github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/google/uuid"
)

var (
	// Cache dei materiali per riutilizzo
	materialCache = make(map[math32.Color]*g3nmat.Standard)
)

// getMaterial ritorna un materiale dalla cache o ne crea uno nuovo
func getMaterial(color *math32.Color) *g3nmat.Standard {
	if cached, exists := materialCache[*color]; exists {
		return cached
	}
	mat := g3nmat.NewStandard(color)
	materialCache[*color] = mat
	return mat
}

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
	config       *config.Config

	Heat            float64
	HeatTransferred float64

	// Color interpolation
	currentColor math32.Color
	targetColor  math32.Color

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
		// Usa la cache delle geometrie per il nuovo raggio
		var geom *geometry.Geometry
		if cached, exists := geometryCache[u._radius]; exists {
			geom = cached
		} else {
			geom = geometry.NewSphere(float64(u._radius), Segments, Segments)
			geometryCache[u._radius] = geom
		}

		// Ottieni il colore dalla composizione
		baseColorRGB := u.Composition.GetColor()
		baseColor := &math32.Color{
			R: float32(baseColorRGB[0]),
			G: float32(baseColorRGB[1]),
			B: float32(baseColorRGB[2]),
		}

		// Crea un nuovo mesh con la geometria e il materiale dalla cache
		mat := getMaterial(baseColor)
		newMesh := graphic.NewMesh(geom, mat)
		newMesh.SetVisible(true)

		// Mantieni la luce esistente
		if u.Mesh.Light != nil {
			u.Mesh.Light.SetColor(baseColor)
			// Rimuovi la luce dal vecchio mesh e aggiungila al nuovo
			u.Mesh.Remove(u.Mesh.Light)
			newMesh.Add(u.Mesh.Light)
		}

		// Sostituisci il vecchio mesh con quello nuovo
		u.Mesh.Mesh = newMesh
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

// Position returns the current position of the unit
func (u *Unit) Position() vector3.Vector[float64] {
	return u._position
}

// Velocity returns the current velocity of the unit
func (u *Unit) Velocity() vector3.Vector[float64] {
	return u._velocity
}

// Elasticity returns the elasticity coefficient of the unit
func (u *Unit) Elasticity() float64 {
	return u._elasticity
}

// Radius returns the radius of the unit
func (u *Unit) Radius() float64 {
	return u._radius
}

// Mass returns the mass of the unit
func (u *Unit) Mass() float64 {
	return u._mass
}

// SetPosition sets the position of the unit
func (u *Unit) SetPosition(pos vector3.Vector[float64]) {
	u._position = pos
}

// SetVelocity sets the velocity of the unit
func (u *Unit) SetVelocity(vel vector3.Vector[float64]) {
	u._velocity = vel
}

// AddHeat adds the specified amount of heat to the unit
func (u *Unit) AddHeat(heat float64) {
	u.Heat += heat
}

// TransferHeatTo transfers heat from this unit to another unit
func (u *Unit) TransferHeatTo(other *Unit, dt float64) {
	if u.Heat <= AmbientTemperature {
		return
	}

	// Use vectorPool for distance calculation
	vec := VectorPool.Get().([]float64)
	defer VectorPool.Put(vec)

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

	// Trasferimento di calore per conduzione
	conductivity := (u.thermalConductivity + other.thermalConductivity) / 2
	conductiveTransfer := conductivity * distanceFactor * tempDiff * dt
	massRatio := math.Sqrt(other._mass / u._mass)
	conductiveTransfer *= massRatio

	// Trasferimento di calore per irraggiamento
	tempK1 := u.Heat + 273.15
	tempK2 := other.Heat + 273.15
	viewFactor := calculateViewFactor(u, other, distance)

	// Potenza irradiata secondo Stefan-Boltzmann
	radiativePower := u.emissivity * other.emissivity * StefanBoltzmannConstant *
		viewFactor * u.surfaceArea *
		(math.Pow(tempK1, 4) - math.Pow(tempK2, 4))

	// Conversione della potenza in variazione di temperatura
	radiativeTransfer := (radiativePower * dt) / (u._mass * u.specificHeatCapacity)

	// Combina i due tipi di trasferimento di calore
	totalTransfer := conductiveTransfer + radiativeTransfer
	totalTransfer = math.Min(totalTransfer, u.Heat-AmbientTemperature)

	u.Heat -= totalTransfer
	other.Heat += totalTransfer
}

// Cache delle geometrie per riutilizzo
var geometryCache = make(map[float64]*geometry.Geometry)

// temperatureToColor calcola il colore basato sulla temperatura in Kelvin usando
// un'approssimazione accurata della legge di Planck per la radiazione del corpo nero
func temperatureToColor(tempK float64) math32.Color {
	// Implementazione basata sulla legge di Planck
	// Riferimento: https://en.wikipedia.org/wiki/Planckian_locus

	// Limiti di temperatura per il calcolo del colore
	t := math.Max(1000, math.Min(40000, tempK))

	// Calcolo delle coordinate di cromaticità CIE 1960
	// Formule di approssimazione valide per 1000K <= T <= 40000K
	if t <= 4000 {
		// Per temperature basse (1000K-4000K)
		x := -0.2661239e9/(t*t*t) - 0.2343589e6/(t*t) + 0.8776956e3/t + 0.179910
		y := -1.1063814*x*x*x - 1.34811020*x*x + 2.18555832*x - 0.20219683

		// Conversione da xy a RGB
		return xyChromaticityToRGB(float32(x), float32(y))
	} else {
		// Per temperature alte (4000K-40000K)
		x := -3.0258469e9/(t*t*t) + 2.1070379e6/(t*t) + 0.2226347e3/t + 0.240390
		y := -1.1063814*x*x*x - 1.34811020*x*x + 2.18555832*x - 0.20219683

		// Conversione da xy a RGB
		return xyChromaticityToRGB(float32(x), float32(y))
	}
}

// xyChromaticityToRGB converte le coordinate di cromaticità CIE 1960 in colori RGB
func xyChromaticityToRGB(x, y float32) math32.Color {
	// Matrice di conversione da XYZ a RGB sRGB
	const (
		r11 = 3.2404542
		r12 = -1.5371385
		r13 = -0.4985314
		r21 = -0.9692660
		r22 = 1.8760108
		r23 = 0.0415560
		r31 = 0.0556434
		r32 = -0.2040259
		r33 = 1.0572252
	)

	// Calcolo dei valori XYZ
	X := x / y
	Y := 1.0 // Luminanza normalizzata
	Z := (1 - x - y) / y

	// Conversione da XYZ a RGB lineare
	r := float32(r11*float64(X) + r12*float64(Y) + r13*float64(Z))
	g := float32(r21*float64(X) + r22*float64(Y) + r23*float64(Z))
	b := float32(r31*float64(X) + r32*float64(Y) + r33*float64(Z))

	// Correzione gamma e normalizzazione
	r = math32.Clamp(gammaCorrect(r), 0, 1)
	g = math32.Clamp(gammaCorrect(g), 0, 1)
	b = math32.Clamp(gammaCorrect(b), 0, 1)

	return math32.Color{R: r, G: g, B: b}
}

// gammaCorrect applica la correzione gamma per sRGB
func gammaCorrect(linear float32) float32 {
	if linear <= 0.0031308 {
		return 12.92 * linear
	}
	return 1.055*math32.Pow(linear, 1/2.4) - 0.055
}

// calculateFinalColor combina il colore base del materiale con il colore di emissione termica
func calculateFinalColor(baseColor math32.Color, tempK float64, emissivity float64) math32.Color {
	if tempK <= 273.15+AmbientTemperature {
		return baseColor
	}

	// Calcola il colore termico
	thermalColor := temperatureToColor(tempK)

	// Usa l'emissività per bilanciare il colore base con il colore termico
	e := float32(emissivity)
	return math32.Color{
		R: baseColor.R*(1-e) + thermalColor.R*e,
		G: baseColor.G*(1-e) + thermalColor.G*e,
		B: baseColor.B*(1-e) + thermalColor.B*e,
	}
}

// interpolateColor interpola gradualmente tra il colore corrente e il colore target
func interpolateColor(current, target math32.Color, factor float32) math32.Color {
	return math32.Color{
		R: current.R + (target.R-current.R)*factor,
		G: current.G + (target.G-current.G)*factor,
		B: current.B + (target.B-current.B)*factor,
	}
}

// NewPointLightMesh creates a new mesh with a point light for the unit
func (u *Unit) NewPointLightMesh() {
	// Inizializza i valori cached
	u.volume = (4.0 / 3.0) * math.Pi * math.Pow(u._radius, 3)
	u.surfaceArea = 4.0 * math.Pi * math.Pow(u._radius, 2)

	// Ottieni le proprietà del materiale
	density, specificHeat, thermalConductivity, emissivity, elasticity := u.Composition.GetEffectiveProperties()
	if u._mass == 0 {
		u._mass = u.volume * density
	}
	u.specificHeatCapacity = specificHeat
	u.thermalConductivity = thermalConductivity
	u.emissivity = emissivity
	u._elasticity = elasticity

	// Ottieni il colore base dalla composizione
	baseColorRGB := u.Composition.GetColor()
	baseColor := math32.Color{
		R: float32(baseColorRGB[0]),
		G: float32(baseColorRGB[1]),
		B: float32(baseColorRGB[2]),
	}

	// Crea o riutilizza la geometria dalla cache
	var geom *geometry.Geometry
	if cached, exists := geometryCache[u._radius]; exists {
		geom = cached
	} else {
		geom = geometry.NewSphere(float64(u._radius), Segments, Segments)
		geometryCache[u._radius] = geom
	}

	// Calcola la temperatura iniziale basata sulla massa e capacità termica
	// T = Q/(m*c) dove:
	// Q = energia termica iniziale (J)
	// m = massa (kg)
	// c = capacità termica specifica (J/kg·K)
	u.Heat = InitialThermalEnergy / (u._mass * u.specificHeatCapacity)

	// Inizializza i colori
	u.currentColor = baseColor
	u.targetColor = baseColor

	// Crea il mesh con la geometria e il materiale dalla cache
	u.Mesh = new(PointLightMesh)
	mat := getMaterial(&baseColor)
	u.Mesh.Mesh = graphic.NewMesh(geom, mat)
	u.Mesh.Mesh.SetVisible(true)

	// Crea e aggiungi la luce
	light := light.NewPoint(&baseColor, 1.0)
	u.Mesh.Light = light
	u.Mesh.Add(light)
}

// GetVolume returns the volume of the unit
func (u *Unit) GetVolume() float64 {
	return u.volume
}

// GetSurfaceArea returns the surface area of the unit
func (u *Unit) GetSurfaceArea() float64 {
	return u.surfaceArea
}

// GetMass returns the mass of the unit
func (u *Unit) GetMass() float64 {
	return u._mass
}

// calculateViewFactor calcola il fattore di vista tra due unità
// Il fattore di vista è un valore tra 0 e 1 che rappresenta la frazione di energia radiante
// che viene scambiata tra le due unità, considerando la loro geometria e posizione relativa
func calculateViewFactor(unit1, unit2 *Unit, distance float64) float64 {
	// Per sfere, il fattore di vista dipende dai raggi e dalla distanza
	r1 := unit1.Radius()
	r2 := unit2.Radius()

	// Formula semplificata: area1 * area2 / (π * distance^2)
	// Normalizzata per evitare valori troppo grandi a distanze piccole
	return (r1 * r2) / (distance*distance + r1*r2)
}

// ApplyForce applies a force to the unit
func (u *Unit) ApplyForce(f vector3.Vector[float64]) {
	invMass := 1.0 / u._mass
	u.Acceleration = u.Acceleration.Add(f.Scale(invMass))
}

// UpdatePosition updates the position and visual representation of the unit
func (u *Unit) UpdatePosition(dt float64) {
	// Update position and velocity using Verlet integration
	halfDtSq := 0.5 * dt * dt
	newPos := u._position.Add(u._velocity.Scale(dt)).Add(u.Acceleration.Scale(halfDtSq))
	newVel := u._velocity.Add(u.Acceleration.Scale(dt))

	// Aggiorna la posizione e velocità
	u._position = newPos
	u._velocity = newVel
	u.Acceleration = vector3.Zero[float64]()

	// Aggiorna sempre la posizione del mesh
	u.Mesh.SetPosition(float32(u._position.X()), float32(u._position.Y()), float32(u._position.Z()))

	// Calcolo del raffreddamento secondo la legge di Stefan-Boltzmann
	tempK := u.Heat + 273.15 // Conversione in Kelvin
	ambientK := AmbientTemperature + 273.15

	// Potenza irradiata secondo Stefan-Boltzmann
	power := u.emissivity * StefanBoltzmannConstant * u.surfaceArea *
		(math.Pow(tempK, 4) - math.Pow(ambientK, 4))

	// Conversione della potenza in variazione di temperatura
	cooldown := (power * dt) / (u._mass * u.specificHeatCapacity)
	newHeat := math.Max(AmbientTemperature, u.Heat-cooldown)
	u.Heat = newHeat

	// Calcola il colore target basato sulle proprietà fisiche
	baseColorRGB := u.Composition.GetColor()
	baseColor := math32.Color{
		R: float32(baseColorRGB[0]),
		G: float32(baseColorRGB[1]),
		B: float32(baseColorRGB[2]),
	}

	// Calcola il colore finale considerando temperatura ed emissività
	u.targetColor = calculateFinalColor(baseColor, tempK, u.emissivity)

	// Interpola gradualmente verso il colore target
	const smoothingFactor = 0.1 // 10% di interpolazione per frame
	u.currentColor = interpolateColor(u.currentColor, u.targetColor, smoothingFactor)

	// Aggiorna il materiale usando la cache
	mat := getMaterial(&u.currentColor)
	u.Mesh.Mesh.SetMaterial(mat)

	// Calcola l'intensità della luce basata sulla legge di Stefan-Boltzmann
	// L'intensità è proporzionale a T^4
	tempRatio := tempK / (AmbientTemperature + 273.15)
	intensity := math32.Clamp(float32(u.emissivity*math.Pow(tempRatio, 4)), 0.1, 5.0)

	// Applica il colore e l'intensità alla luce
	u.Mesh.Light.SetColor(&u.currentColor)
	u.Mesh.Light.SetIntensity(intensity)
}

// CheckAndResolveWallCollision checks for and resolves collisions with the world boundaries
func (unit *Unit) CheckAndResolveWallCollision(wallBounds BoundingBox, wallElasticity float64) bool {
	var collided bool
	vec := VectorPool.Get().([]float64)
	defer VectorPool.Put(vec)

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

// Orbit calculates and sets the velocity needed for orbit around another unit
func (u *Unit) Orbit(target *Unit) error {
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
	v := math.Sqrt((G * target.Mass()) / (2 * currentDistance))

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

// GetConfig returns the unit's configuration
func (u *Unit) GetConfig() *config.Config {
	return u.config
}

// GiveMassAndCenterOfMassForBounds calculates the mass and center of mass for a given bounding box
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
