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
	Velocity         rl.Vector2
	Acceleration     rl.Vector2
	Elasticity       float32
	Radius           float32
	Mass             float32
	Color            color.RGBA
}

func NewUnitWithProperties(cfg *config.Config) *Unit {
	currentRadius := cfg.ParticleRadius
	currentMass := cfg.ParticleMass
	currentElasticity := cfg.ParticleElasticity

	if cfg.SetRandomRadius {
		currentRadius = cfg.RadiusMin + rand.Float32()*(cfg.RadiusMax-cfg.RadiusMin)
	}
	if cfg.SetRandomMass {
		currentMass = cfg.MassMin + rand.Float32()*(cfg.MassMax-cfg.MassMin)
	}
	if cfg.SetRandomElasticity {
		currentElasticity = cfg.ElasticityMin + rand.Float32()*(cfg.ElasticityMax-cfg.ElasticityMin)
	}

	color := color.RGBA{255, 0, 0, 255}

	if cfg.SetRandomColor {
		color = utils.RandomRaylibColor()
	}

	return &Unit{
		Id:         uuid.New(),
		Radius:     currentRadius,
		Mass:       currentMass,
		Elasticity: currentElasticity,
		Color:      color,
	}
}

func isOverlapping(newUnit *Unit, existingUnits []*Unit) bool {
	for _, unit := range existingUnits {
		dx := newUnit.Position.X - unit.Position.X
		dy := newUnit.Position.Y - unit.Position.Y
		distanceSquared := dx*dx + dy*dy
		radiusSum := newUnit.Radius + unit.Radius
		if distanceSquared < radiusSum*radiusSum {
			return true
		}
	}
	return false
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

				if !isOverlapping(newUnit, existingUnits) {
					return newUnit.Position
				}
			}
		}
	}
}

func newUnitsAtPosition(spawnPosition rl.Vector2, cfg *config.Config) *[]*Unit {
	units := make([]*Unit, 0, cfg.ParticleNumber)
	centerX := spawnPosition.X
	centerY := spawnPosition.Y

	for i := 0; i < int(cfg.ParticleNumber); i++ {
		newUnit := *NewUnitWithProperties(cfg)

		if len(units) == 0 {
			newUnit.Position = rl.Vector2{X: centerX, Y: centerY}
			newUnit.PreviousPosition = newUnit.Position
			newUnit.Velocity = rl.Vector2{X: 200, Y: 0}
		} else {
			newPosition := findClosestAvailablePosition(&newUnit, units, newUnit.Radius*2)
			newUnit.Position = newPosition
			newUnit.PreviousPosition = newPosition
			newUnit.Velocity = rl.Vector2{X: 200, Y: 0}
		}

		units = append(units, &newUnit)
	}

	return &units
}

func calculateInitialVelocity(position rl.Vector2, simulationWidth, simulationHeight int32) rl.Vector2 {
	const maxSpeed float32 = 1000.0

	d_left := position.X
	d_right := float32(simulationWidth) - position.X
	d_top := position.Y
	d_bottom := float32(simulationHeight) - position.Y

	var velocityX, velocityY float32

	velocityX = maxSpeed*(d_right/float32(simulationWidth)) - maxSpeed*(d_left/float32(simulationWidth))
	velocityY = maxSpeed*(d_bottom/float32(simulationHeight)) - maxSpeed*(d_top/float32(simulationHeight))

	return rl.Vector2{X: velocityX, Y: velocityY}
}
func spawnUnitsWithVelocity(units *[]*Unit, spawnPosition rl.Vector2, cfg *config.Config) {
	var lastSpawned *Unit = nil

	for i := 0; i < int(cfg.ParticleNumber); i++ {
		unit := *NewUnitWithProperties(cfg)

		for lastSpawned != nil && distanceBetween(lastSpawned.Position, spawnPosition) < 2*unit.Radius {
			time.Sleep(100 * time.Millisecond)
		}

		velocity := calculateInitialVelocity(spawnPosition, cfg.GameWidth, cfg.WindowHeight)

		previousPosition := rl.Vector2{
			X: spawnPosition.X - velocity.X/float32(rl.GetFPS()),
			Y: spawnPosition.Y - velocity.Y/float32(rl.GetFPS()),
		}

		unit.Position = spawnPosition
		unit.PreviousPosition = previousPosition
		unit.Velocity = velocity

		*units = append(*units, &unit)
		lastSpawned = &unit
	}
}

func (u *Unit) GetVelocityWithVerlet() rl.Vector2 {
	return rl.Vector2{
		X: u.Position.X - u.PreviousPosition.X,
		Y: u.Position.Y - u.PreviousPosition.Y,
	}
}

func distanceBetween(p1, p2 rl.Vector2) float32 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}

func (u *Unit) accelerate(a rl.Vector2) {
	u.Acceleration.X += a.X
	u.Acceleration.Y += a.Y
}

func areOverlapping(a, b *Unit) bool {
	deltaX := b.Position.X - a.Position.X
	deltaY := b.Position.Y - a.Position.Y
	distanceSquared := deltaX*deltaX + deltaY*deltaY
	totalRadius := a.Radius + b.Radius
	return distanceSquared < totalRadius*totalRadius
}
func calculateCollisionWithVerlet(unitA, unitB *Unit) {

	deltaX := unitB.Position.X - unitA.Position.X
	deltaY := unitB.Position.Y - unitA.Position.Y

	distance := float32(math.Sqrt(float64(deltaX*deltaX + deltaY*deltaY)))
	overlap := unitA.Radius + unitB.Radius - distance

	if overlap <= 0 {
		return
	}

	normalX := deltaX / distance
	normalY := deltaY / distance

	correctionX := (overlap / 2) * normalX
	correctionY := (overlap / 2) * normalY

	unitA.Position.X -= correctionX
	unitA.Position.Y -= correctionY
	unitB.Position.X += correctionX
	unitB.Position.Y += correctionY

	relativeVelocityX := unitB.Velocity.X - unitA.Velocity.X
	relativeVelocityY := unitB.Velocity.Y - unitA.Velocity.Y
	impulse := -(1.0 + (unitA.Elasticity+unitB.Elasticity)/2) * (relativeVelocityX*normalX + relativeVelocityY*normalY) / (1/unitA.Mass + 1/unitB.Mass)

	impulseX := impulse * normalX
	impulseY := impulse * normalY

	unitA.Velocity.X -= impulseX / unitA.Mass
	unitA.Velocity.Y -= impulseY / unitA.Mass
	unitB.Velocity.X += impulseX / unitB.Mass
	unitB.Velocity.Y += impulseY / unitB.Mass
}

func (u *Unit) updatePositionWithVerlet(dt float32) {
	newPosition := rl.Vector2{}
	newPosition.X = 2*u.Position.X - u.PreviousPosition.X + u.Acceleration.X*dt*dt
	newPosition.Y = 2*u.Position.Y - u.PreviousPosition.Y + u.Acceleration.Y*dt*dt

	u.PreviousPosition = u.Position
	u.Position = newPosition
	u.Acceleration = rl.Vector2{X: 0, Y: 0}
}

func (u *Unit) checkWallCollisionVerlet(cfg *config.Config, deltaTime float32) {

	oldPosition := u.Position

	u.Position.X = 2*u.Position.X - oldPosition.X + u.Acceleration.X*deltaTime*deltaTime
	u.Position.Y = 2*u.Position.Y - oldPosition.Y + u.Acceleration.Y*deltaTime*deltaTime

	if u.Position.X-u.Radius < 0 {
		u.Position.X = u.Radius

		oldPosition.X = u.Position.X + (u.Position.X-oldPosition.X)*cfg.WallElasticity
	} else if u.Position.X+u.Radius > float32(cfg.GameWidth) {
		u.Position.X = float32(cfg.GameWidth) - u.Radius

		oldPosition.X = u.Position.X + (u.Position.X-oldPosition.X)*cfg.WallElasticity
	}

	if u.Position.Y-u.Radius < 0 {
		u.Position.Y = u.Radius

		oldPosition.Y = u.Position.Y + (u.Position.Y-oldPosition.Y)*cfg.WallElasticity
	} else if u.Position.Y+u.Radius > float32(cfg.WindowHeight) {
		u.Position.Y = float32(cfg.WindowHeight) - u.Radius

		oldPosition.Y = u.Position.Y + (u.Position.Y-oldPosition.Y)*cfg.WallElasticity
	}

	//u.PreviousPosition = oldPosition
}
