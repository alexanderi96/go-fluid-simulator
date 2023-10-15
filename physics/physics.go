package physics

import (
	"math"

	"github.com/alexanderi96/go-fluid-simulator/config"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Simulation struct {
	Fluid    *Fluid
	Quadtree *Quadtree
	Metrics  map[string]uint64
	Config   *config.Config
	IsPause  bool
}

// La funzione che si occupa di creare una nuova simulazione
func NewSimulation(config *config.Config) (*Simulation, error) {
	// Ad esempio, se desideri passare alcuni valori di config a NewFluid:
	fluid := newFluid(config.GameWidth, config.WindowHeight, config.ParticleNumber, config.ParticleRadius, config.ParticleMass, config.ParticleInitialSpacing, config.ScaleFactor, config.ParticleElasticity)
	bounds := rl.NewRectangle(0, 0, float32(config.GameWidth), float32(config.WindowHeight))
	quadtree := NewQuadtree(0, bounds)
	return &Simulation{
		Fluid:    fluid,
		Quadtree: quadtree,
		Metrics:  make(map[string]uint64),
		Config:   config,
		IsPause:  false,
	}, nil
}

// La funzione che si occupa di resettare il fluido
func (s *Simulation) Reset() {
	s.Fluid = newFluid(s.Config.GameWidth, s.Config.WindowHeight, s.Config.ParticleNumber, s.Config.ParticleRadius, s.Config.ParticleMass, s.Config.ParticleInitialSpacing, s.Config.ScaleFactor, s.Config.ParticleElasticity)
}

func (s *Simulation) Update(frametime float32) error {
	s.Quadtree.Clear() // Pulisce il quadtree all'inizio di ogni frame
	dt := frametime    // Tempo per l'aggiornamento

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
				deltaX := unitB.Position.X - unitA.Position.X
				deltaY := unitB.Position.Y - unitA.Position.Y
				distanceSquared := deltaX*deltaX + deltaY*deltaY
				totalRadius := unitA.Radius + unitB.Radius
				if distanceSquared < totalRadius*totalRadius {
					calculateCollision(deltaX, deltaY, distanceSquared, totalRadius, unitA, unitB)
				}
			}
		}
	}

	// Aggiorna la posizione delle particelle in base alla loro velocità
	for i := range s.Fluid.Units {
		externalForces := rl.Vector2{}

		// Applica la forza di gravità a tutte le particelle
		if s.Config.ApplyGravity {
			externalForces = rl.Vector2Add(externalForces, rl.Vector2{X: 0, Y: s.Config.Gravity})
		}

		unit := &s.Fluid.Units[i]
		if err := unit.UpdateUnit(dt, externalForces, s.Config); err != nil {
			return err
		}
	}

	return nil

}

func calculateCollision(deltaX, deltaY, distanceSquared, totalRadius float32, unitA, unitB *Unit) {
	// Calcola la risposta alla collisione (questo è molto semplificato)
	normalX := float64(deltaX) / math.Sqrt(float64(distanceSquared))
	normalY := float64(deltaY) / math.Sqrt(float64(distanceSquared))
	relativeVelocityX := unitB.Velocity.X - unitA.Velocity.X
	relativeVelocityY := unitB.Velocity.Y - unitA.Velocity.Y
	dotProduct := float32(normalX)*relativeVelocityX + float32(normalY)*relativeVelocityY
	if dotProduct < 0 {
		coefficientOfRestitution := (unitA.Elasticity + unitB.Elasticity) / 2
		impulse := 2 * dotProduct / (unitA.Mass + unitB.Mass)
		unitA.Velocity.X += impulse * unitB.Mass * float32(normalX) * coefficientOfRestitution
		unitA.Velocity.Y += impulse * unitB.Mass * float32(normalY) * coefficientOfRestitution
		unitB.Velocity.X -= impulse * unitA.Mass * float32(normalX) * coefficientOfRestitution
		unitB.Velocity.Y -= impulse * unitA.Mass * float32(normalY) * coefficientOfRestitution
	}
}
