package physics

import (
	"math"

	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/metrics"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Simulation struct {
	Fluid    *Fluid
	Quadtree *Quadtree
	Metrics  *metrics.Metrics
	Config   *config.Config
	IsPause  bool
}

// La funzione che si occupa di creare una nuova simulazione
func NewSimulation(config *config.Config) (*Simulation, error) {

	config.UpdateWindowSettings()
	// Ad esempio, se desideri passare alcuni valori di config a NewFluid:
	fluid := newFluid(config.GameWidth, config.WindowHeight, config.ParticleNumber, config.ParticleRadius, config.ParticleMass, config.ParticleInitialSpacing, config.ScaleFactor, config.ParticleElasticity)

	bounds := rl.NewRectangle(0, 0, float32(config.GameWidth), float32(config.WindowHeight))
	quadtree := NewQuadtree(0, bounds)

	sim := &Simulation{
		Fluid:    fluid,
		Quadtree: quadtree,
		Metrics:  &metrics.Metrics{},
		Config:   config,
		IsPause:  false,
	}

	return sim, nil
}

// La funzione che si occupa di resettare il fluido
func (s *Simulation) Reset() {
	s.Fluid = newFluid(s.Config.GameWidth, s.Config.WindowHeight, s.Config.ParticleNumber, s.Config.ParticleRadius, s.Config.ParticleMass, s.Config.ParticleInitialSpacing, s.Config.ScaleFactor, s.Config.ParticleElasticity)
}

func (s *Simulation) NewFluidAtPosition(position rl.Vector2) {
	s.Fluid.Units = append(s.Fluid.Units, *newUnitsAtPosition(position, s.Config.GameWidth, s.Config.WindowHeight, s.Config.ParticleNumber, s.Config.ParticleRadius, s.Config.ParticleMass, s.Config.ParticleInitialSpacing, s.Config.ScaleFactor, s.Config.ParticleElasticity)...)
}

func (s *Simulation) Update() error {
	s.Metrics.Update()
	currentFrameTime := s.Metrics.Frametime
	s.Quadtree.Clear() // Pulisce il quadtree all'inizio di ogni frame

	// Costruisci il quadtree
	for i := range s.Fluid.Units {
		s.Quadtree.Insert(&s.Fluid.Units[i])
	}

	// Controlla le collisioni tra particelle e aggiorna le velocità utilizzando il quadtree
	for i := range s.Fluid.Units {
		unitA := &s.Fluid.Units[i]
		nearUnits := []*Unit{}
		s.Quadtree.Retrieve(&nearUnits, unitA)
		for _, unitB := range nearUnits {
			if unitA.Id != unitB.Id {

				if collisionTime, collided := findCollisionTime(unitA, unitB, currentFrameTime); collided {
					calculateCollision(
						collisionTime,
						unitA,
						unitB,
					)
					currentFrameTime -= collisionTime
				}
			}
		}
	}

	// Aggiorna la posizione delle particelle in base alla loro velocità
	for i := range s.Fluid.Units {
		unit := &s.Fluid.Units[i]

		unit.ApplyExternalForce(currentFrameTime, rl.Vector2{X: 0, Y: s.Config.Gravity})

		if err := unit.Update(currentFrameTime, s.Config); err != nil {
			return err
		}

		checkWallCollision(unit, s.Config)
	}
	return nil

}

func findCollisionTime(unitA, unitB *Unit, frametime float32) (t float32, collided bool) {
	// Controlla se le unità sono già sovrapposte
	deltaX := unitB.Position.X - unitA.Position.X
	deltaY := unitB.Position.Y - unitA.Position.Y
	distanceSquared := deltaX*deltaX + deltaY*deltaY
	totalRadius := unitA.Radius + unitB.Radius
	if distanceSquared < totalRadius*totalRadius {
		return 0, true // Le unità sono già sovrapposte, restituisci un tempo di collisione di 0
	}

	// Scala le velocità delle particelle per il frametime
	scaledVelocityA := rl.Vector2Scale(unitA.Velocity, frametime)
	scaledVelocityB := rl.Vector2Scale(unitB.Velocity, frametime)

	// Calcola le differenze nelle velocità
	deltaVX := scaledVelocityB.X - scaledVelocityA.X
	deltaVY := scaledVelocityB.Y - scaledVelocityA.Y

	// Risolvi l'equazione quadratica per trovare il tempo di collisione t
	a := deltaVX*deltaVX + deltaVY*deltaVY
	b := 2 * (deltaX*deltaVX + deltaY*deltaVY)
	c := deltaX*deltaX + deltaY*deltaY - (unitA.Radius+unitB.Radius)*(unitA.Radius+unitB.Radius)

	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return 0, false // Nessuna collisione
	}

	sqrtDiscriminant := math.Sqrt(float64(discriminant))
	t1 := (-b - float32(sqrtDiscriminant)) / (2 * a)
	t2 := (-b + float32(sqrtDiscriminant)) / (2 * a)

	// Scegli il tempo di collisione più piccolo che sia all'interno dell'intervallo [0, 1]
	if 0 <= t1 && t1 <= 1 {
		t, collided = t1*frametime, true
	} else if 0 <= t2 && t2 <= 1 {
		t, collided = t2*frametime, true
	} else {
		t, collided = 0, false // Nessuna collisione in questo frame
	}

	return
}

func calculateCollision(collisionTime float32, unitA, unitB *Unit) {
	// Calcola la posizione delle particelle al momento della collisione
	posXA := unitA.Position.X + unitA.Velocity.X*collisionTime
	posYA := unitA.Position.Y + unitA.Velocity.Y*collisionTime
	posXB := unitB.Position.X + unitB.Velocity.X*collisionTime
	posYB := unitB.Position.Y + unitB.Velocity.Y*collisionTime

	// Calcola la differenza di posizione tra le particelle al momento della collisione
	deltaX := posXB - posXA
	deltaY := posYB - posYA

	// Calcola la distanza al quadrato e la normale della collisione
	distanceSquared := deltaX*deltaX + deltaY*deltaY
	normalX := float64(deltaX) / math.Sqrt(float64(distanceSquared))
	normalY := float64(deltaY) / math.Sqrt(float64(distanceSquared))

	// Calcola la velocità relativa al momento della collisione
	relativeVelocityX := unitB.Velocity.X - unitA.Velocity.X
	relativeVelocityY := unitB.Velocity.Y - unitA.Velocity.Y
	dotProduct := float32(normalX)*relativeVelocityX + float32(normalY)*relativeVelocityY

	if dotProduct < 0 {
		coefficientOfRestitution := (unitA.Elasticity + unitB.Elasticity) / 2
		impulse := 2 * dotProduct / (unitA.Mass + unitB.Mass)

		// Aggiorna solo le velocità delle particelle
		unitA.Velocity.X += impulse * unitB.Mass * float32(normalX) * coefficientOfRestitution
		unitA.Velocity.Y += impulse * unitB.Mass * float32(normalY) * coefficientOfRestitution
		unitB.Velocity.X -= impulse * unitA.Mass * float32(normalX) * coefficientOfRestitution
		unitB.Velocity.Y -= impulse * unitA.Mass * float32(normalY) * coefficientOfRestitution
	}
}

func checkWallCollision(u *Unit, cfg *config.Config) {
	// Controlla e corregge la posizione X
	if u.Position.X-u.Radius < 0 {
		u.Position.X = u.Radius
		u.Velocity.X = -u.Velocity.X // Invertire la velocità X
	} else if u.Position.X+u.Radius > float32(cfg.GameWidth) {
		u.Position.X = float32(cfg.GameWidth) - u.Radius
		u.Velocity.X = -u.Velocity.X // Invertire la velocità X
	}

	// Controlla e corregge la posizione Y
	if u.Position.Y-u.Radius < 0 {
		u.Position.Y = u.Radius
		u.Velocity.Y = -u.Velocity.Y * cfg.WallElasticity // Invertire la velocità Y
	} else if u.Position.Y+u.Radius > float32(cfg.WindowHeight) {
		u.Position.Y = float32(cfg.WindowHeight) - u.Radius
		u.Velocity.Y = -u.Velocity.Y * cfg.WallElasticity // Invertire la velocità Y
	}
}
