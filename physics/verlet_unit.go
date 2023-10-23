package physics

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/utils"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/google/uuid"
)

type Unit struct {
	Id               uuid.UUID
	Position         rl.Vector2
	PreviousPosition rl.Vector2
	Acceleration     rl.Vector2
	Elasticity       float32
	Radius           float32
	MassMultiplier   float32
	Color            color.RGBA
}

func (u *Unit) Volume() float32 {
	// Calcola il volume come area del cerchio (π * r^2)
	return math.Pi * u.Radius * u.Radius
}

func (u *Unit) Mass() float32 {
	// Calcola la massa utilizzando il volume e il MassMultiplier
	return u.Volume() * u.MassMultiplier
}

func (u *Unit) Velocity(dt float32) rl.Vector2 {
	return rl.Vector2{
		X: (u.Position.X - u.PreviousPosition.X) / dt,
		Y: (u.Position.Y - u.PreviousPosition.Y) / dt,
	}
}

func (u *Unit) accelerate(a rl.Vector2) {
	u.Acceleration.X += a.X
	u.Acceleration.Y += a.Y
}

func newUnitWithPropertiesAndAcceleration(cfg *config.Config, acceleration rl.Vector2) *Unit {
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

func findClosestAvailablePosition(newUnit *Unit, existingUnits []*Unit, step float32) rl.Vector2 {
	for radiusMultiplier := 1; ; radiusMultiplier++ {
		for _, unit := range existingUnits {
			for angle := 0.0; angle <= 2*math.Pi; angle += 0.1 {
				dx := step * float32(math.Cos(angle)) * float32(radiusMultiplier)
				dy := step * float32(math.Sin(angle)) * float32(radiusMultiplier)

				candidateX := unit.Position.X + dx
				candidateY := unit.Position.Y + dy

				newUnit.Position = rl.Vector2{X: candidateX, Y: candidateY}

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

func newUnitsWithAcceleration(spawnPosition rl.Vector2, cfg *config.Config, acceleration rl.Vector2) *[]*Unit {
	units := make([]*Unit, 0, cfg.UnitNumber)
	centerX := spawnPosition.X
	centerY := spawnPosition.Y

	for i := 0; i < int(cfg.UnitNumber); i++ {
		newUnit := *newUnitWithPropertiesAndAcceleration(cfg, acceleration)

		if len(units) == 0 {
			newUnit.Position = rl.Vector2{X: centerX, Y: centerY}
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

func calculateInitialVelocity(position rl.Vector2, gameX, GameY int32) rl.Vector2 {
	const maxSpeed float32 = 800.0

	d_left := position.X
	d_right := float32(gameX) - position.X
	d_top := position.Y
	d_bottom := float32(GameY) - position.Y

	var velocityX, velocityY float32

	velocityX = maxSpeed*(d_right/float32(gameX)) - maxSpeed*(d_left/float32(gameX))
	velocityY = maxSpeed*(d_bottom/float32(GameY)) - maxSpeed*(d_top/float32(GameY))

	return rl.Vector2{X: velocityX, Y: velocityY}
}
func spawnUnitsWithVelocity(units *[]*Unit, spawnPosition rl.Vector2, cfg *config.Config) {
	var lastSpawned *Unit = nil

	for i := 0; i < int(cfg.UnitNumber); i++ {
		unit := *newUnitWithPropertiesAndAcceleration(cfg, rl.Vector2{X: 0, Y: 0})

		for lastSpawned != nil && distanceBetween(lastSpawned.Position, spawnPosition) < 2*unit.Radius {
			time.Sleep(100 * time.Millisecond)
		}

		velocity := calculateInitialVelocity(spawnPosition, cfg.GameX, cfg.GameY)

		previousPosition := rl.Vector2{
			X: spawnPosition.X - velocity.X/float32(rl.GetFPS()),
			Y: spawnPosition.Y - velocity.Y/float32(rl.GetFPS()),
		}

		unit.Position = spawnPosition
		unit.PreviousPosition = previousPosition

		*units = append(*units, &unit)
		lastSpawned = &unit
	}
}
