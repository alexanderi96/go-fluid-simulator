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
	Elasticity       float32
	Radius           float32
	MassMultiplier   float32
	Color            color.RGBA
}

func (u *Unit) Volume() float32 {
	// Calcola il volume della sfera (4/3 * π * r^3)
	return float32((4.0 / 3.0) * math.Pi * math.Pow(float64(u.Radius), 3))
}

func (u *Unit) Mass() float32 {
	// Calcola la massa utilizzando il volume e il MassMultiplier
	return u.Volume() * u.MassMultiplier
}

func (u *Unit) Velocity(dt float32) rl.Vector3 {
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

func newUnitWithPropertiesAndAcceleration(cfg *config.Config, acceleration rl.Vector3) *Unit {
	currentRadius := cfg.UnitRadius
	currentMassMultiplier := cfg.UnitMassMultiplier
	currentElasticity := cfg.UnitElasticity

	if cfg.SetRandomRadius {
		currentRadius = cfg.RadiusMin + rand.Float32()*(cfg.RadiusMax-cfg.RadiusMin)
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

	return &Unit{
		Id:             uuid.New(),
		Acceleration:   acceleration,
		Radius:         currentRadius,
		MassMultiplier: currentMassMultiplier,
		Elasticity:     currentElasticity,
		Color:          color,
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

func newUnitsWithAcceleration(spawnPosition rl.Vector3, cfg *config.Config, acceleration rl.Vector3) *[]*Unit {
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

// func calculateInitialVelocity(position rl.Vector3, gameX, gameY, gameZ int32) rl.Vector3 {
// 	const maxSpeed float32 = 800.0

// 	d_left := position.X
// 	d_right := float32(gameX) - position.X
// 	d_front := position.Y // In uno spazio 3D, Y potrebbe rappresentare la profondità/front-back
// 	d_back := float32(gameY) - position.Y
// 	d_bottom := position.Z // Assumendo che Z rappresenti l'asse verticale
// 	d_top := float32(gameZ) - position.Z

// 	var velocityX, velocityY, velocityZ float32

// 	velocityX = maxSpeed*(d_right/float32(gameX)) - maxSpeed*(d_left/float32(gameX))
// 	velocityY = maxSpeed*(d_back/float32(gameY)) - maxSpeed*(d_front/float32(gameY))
// 	velocityZ = maxSpeed*(d_top/float32(gameZ)) - maxSpeed*(d_bottom/float32(gameZ))

// 	return rl.Vector3{X: velocityX, Y: velocityY, Z: velocityZ}
// }
// func spawnUnitsWithVelocity(units *[]*Unit, spawnPosition rl.Vector3, cfg *config.Config) {
// 	var lastSpawned *Unit = nil

// 	for i := 0; i < int(cfg.UnitNumber); i++ {
// 		unit := *newUnitWithPropertiesAndAcceleration(cfg, rl.Vector3{X: 0, Y: 0, Z: 0})

// 		for lastSpawned != nil && distanceBetween(lastSpawned.Position, spawnPosition) < 2*unit.Radius {
// 			time.Sleep(100 * time.Millisecond)
// 		}

// 		velocity := calculateInitialVelocity(spawnPosition, int32(cfg.GameX), int32(cfg.GameY), int32(cfg.GameZ))

// 		previousPosition := rl.Vector3{
// 			X: spawnPosition.X - velocity.X/float32(rl.GetFPS()),
// 			Y: spawnPosition.Y - velocity.Y/float32(rl.GetFPS()),
// 			Z: spawnPosition.Z - velocity.Z/float32(rl.GetFPS()),
// 		}

// 		unit.Position = spawnPosition
// 		unit.PreviousPosition = previousPosition

// 		*units = append(*units, &unit)
// 		lastSpawned = &unit
// 	}
// }
