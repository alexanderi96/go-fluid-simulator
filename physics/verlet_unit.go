package physics

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/utils"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

type Unit struct {
	Id               uuid.UUID
	Position         rl.Vector3
	PreviousPosition rl.Vector3
	Acceleration     rl.Vector3
	Velocity         rl.Vector3
	Elasticity       float32
	Radius           float32
	MassMultiplier   float32
	Mass             float32
	Volume           float32
	Color            color.RGBA
	OldColor         color.RGBA

	Cluster            *Cluster
	OldCluster         *Cluster
	TransitionTimer    float32 // Timer per la transizione del colore
	TransitionDuration float32 // Durata della transizione
}

type Cluster struct {
	Id    uuid.UUID
	Color rl.Color
}

func (u *Unit) GetVolume() float32 {
	// Calcola il volume della sfera (4/3 * π * r^3)
	return float32((4.0 / 3.0) * math.Pi * math.Pow(float64(u.Radius), 3))
}

func (u *Unit) GetMass() float32 {
	// Calcola la massa utilizzando il volume e il MassMultiplier
	return u.Volume * u.MassMultiplier
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

func (u *Unit) update(applyGravity bool, gravity, dt float32) {

	if applyGravity {
		u.accelerate(rl.Vector3{X: 0, Y: -gravity, Z: 0})
	}
	u.updatePositionWithVerlet(dt)

	if u.TransitionTimer < u.TransitionDuration && u.Cluster != nil {
		u.TransitionTimer += dt
		if u.TransitionTimer > u.TransitionDuration {
			u.TransitionTimer = u.TransitionDuration
		}
	} else if u.Cluster == nil && u.TransitionTimer > 0 {
		u.TransitionTimer -= dt
		if u.TransitionTimer < 0 {
			u.TransitionTimer = 0
			u.OldCluster = &Cluster{}
		}
	}
}

func (u *Unit) updatePositionWithVerlet(dt float32) {
	newPosition := rl.Vector3{}
	newPosition.X = 2*u.Position.X - u.PreviousPosition.X + u.Acceleration.X*dt*dt
	newPosition.Y = 2*u.Position.Y - u.PreviousPosition.Y + u.Acceleration.Y*dt*dt
	newPosition.Z = 2*u.Position.Z - u.PreviousPosition.Z + u.Acceleration.Z*dt*dt

	u.PreviousPosition = u.Position
	u.Position = newPosition
	u.Acceleration = rl.Vector3{X: 0, Y: 0, Z: 0}
}

func (u *Unit) checkWallCollisionVerlet(boundary rl.BoundingBox, wallElasticity, deltaTime float32) {
	// Calcola la velocità
	velocity := u.GetVelocity(deltaTime)
	elasticity := float32(math.Min(float64(u.Elasticity), float64(wallElasticity)))

	// Funzione di correzione per ogni asse
	correctAxis := func(pos, prevPos *float32, vel *float32, min, max float32) {
		if *pos-u.Radius < min {
			overlap := u.Radius - *pos
			*pos += overlap
			*vel = -*vel * elasticity
			*prevPos = *pos - *vel*deltaTime
		} else if *pos+u.Radius > max {
			overlap := (*pos + u.Radius) - max
			*pos -= overlap
			*vel = -*vel * elasticity
			*prevPos = *pos - *vel*deltaTime
		}
	}

	// Correzione per ogni asse
	correctAxis(&u.Position.X, &u.PreviousPosition.X, &velocity.X, boundary.Min.X, boundary.Max.X)
	correctAxis(&u.Position.Y, &u.PreviousPosition.Y, &velocity.Y, boundary.Min.Y, boundary.Max.Y)
	correctAxis(&u.Position.Z, &u.PreviousPosition.Z, &velocity.Z, boundary.Min.Z, boundary.Max.Z)
}

func newUnitWithPropertiesAndAcceleration(cfg *config.Config, acceleration rl.Vector3) *Unit {
	currentRadius := cfg.UnitRadius * cfg.UnitRadiusMultiplier
	currentMassMultiplier := cfg.UnitMassMultiplier
	currentElasticity := cfg.UnitElasticity
	currentTransitionDuration := cfg.UnitTransitionDuration

	if cfg.SetRandomRadius {
		currentRadius = (cfg.RadiusMin + rand.Float32()*(cfg.RadiusMax-cfg.RadiusMin)) * cfg.UnitRadiusMultiplier
	}
	if cfg.SetRandomMassMultiplier {
		currentMassMultiplier = cfg.MassMultiplierMin + rand.Float32()*(cfg.MassMultiplierMax-cfg.MassMultiplierMin)
	}
	if cfg.SetRandomElasticity {
		currentElasticity = cfg.ElasticityMin + rand.Float32()*(cfg.ElasticityMax-cfg.ElasticityMin)
	}

	color := color.RGBA{255, 0, 0, 255}

	if cfg.SetRandomColor {
		color = utils.RandomRaylibColor()
	}

	unit := &Unit{
		Id:                 uuid.New(),
		Acceleration:       acceleration,
		Radius:             currentRadius,
		MassMultiplier:     currentMassMultiplier,
		Elasticity:         currentElasticity,
		Color:              color,
		TransitionTimer:    0,
		TransitionDuration: float32(currentTransitionDuration),
	}

	unit.Volume = unit.GetVolume()
	unit.Mass = unit.GetMass()

	return unit
}

func newUnitWithPropertiesAtPosition(position rl.Vector3, acceleration rl.Vector3, radius, massMultiplier, elasticity float32, color color.RGBA) *Unit {
	return &Unit{
		Id:                 uuid.New(),
		Position:           position,
		PreviousPosition:   position,
		Acceleration:       acceleration,
		Radius:             radius,
		MassMultiplier:     massMultiplier,
		Elasticity:         elasticity,
		Color:              color,
		TransitionTimer:    0,
		TransitionDuration: 0,
	}
}

func findClosestAvailablePosition(newUnit *Unit, existingUnits []*Unit, step float32) rl.Vector3 {
	for radiusMultiplier := 1; ; radiusMultiplier++ {
		for _, unit := range existingUnits {
			for angle := 0.0; angle <= 2*math.Pi; angle += 0.1 {
				for zMultiplier := -radiusMultiplier; zMultiplier <= radiusMultiplier; zMultiplier++ {
					dx := step * float32(math.Cos(angle)) * float32(radiusMultiplier)
					dy := step * float32(math.Sin(angle)) * float32(radiusMultiplier)
					dz := step * float32(zMultiplier)

					candidateX := unit.Position.X + dx
					candidateY := unit.Position.Y + dy
					candidateZ := unit.Position.Z + dz

					newUnit.Position = rl.Vector3{X: candidateX, Y: candidateY, Z: candidateZ}

					overlap := false
					for _, existingUnit := range existingUnits {
						if areOverlapping(newUnit, existingUnit) {
							overlap = true
							break
						}
					}
					if !overlap {
						return newUnit.Position
					}
				}
			}
		}
	}
}

func newUnitsWithAcceleration(spawnPosition, acceleration rl.Vector3, cfg *config.Config) *[]*Unit {
	units := make([]*Unit, 0, cfg.UnitNumber)
	centerX := spawnPosition.X
	centerY := spawnPosition.Y
	centerZ := spawnPosition.Z

	for i := 0; i < int(cfg.UnitNumber); i++ {
		newUnit := *newUnitWithPropertiesAndAcceleration(cfg, acceleration)

		if len(units) == 0 {
			newUnit.Position = rl.Vector3{X: centerX, Y: centerY, Z: centerZ}
			newUnit.PreviousPosition = newUnit.Position
		} else {
			newPosition := findClosestAvailablePosition(&newUnit, units, newUnit.Radius*2)
			newUnit.Position = newPosition
			newUnit.PreviousPosition = newPosition
		}

		units = append(units, &newUnit)
	}

	return &units
}

func (u *Unit) BlendedColor() rl.Color {
	// Assicurati che t sia compreso tra 0 e 1
	t := u.TransitionTimer / u.TransitionDuration

	if t <= 0 {
		return u.Color
	} else if t > 1 {
		return u.Cluster.Color
	}

	color1, color2 := u.Color, u.OldColor
	if u.Cluster != nil {
		color2 = u.Cluster.Color
	} else if u.OldCluster != nil {
		color2 = u.OldCluster.Color
	}

	// Calcola i componenti del nuovo colore interpolando linearmente tra color1 e clusterColor
	r := float32(color1.R) + (float32(color2.R)-float32(color1.R))*t
	g := float32(color1.G) + (float32(color2.G)-float32(color1.G))*t
	b := float32(color1.B) + (float32(color2.B)-float32(color1.B))*t
	a := float32(color1.A) + (float32(color2.A)-float32(color1.A))*t

	u.OldColor = rl.NewColor(uint8(r), uint8(g), uint8(b), uint8(a))
	// Crea e ritorna il nuovo colore
	return u.OldColor
}
