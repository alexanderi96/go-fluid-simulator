package physics

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Simulation) UpdateWithQuadtree() error {
	s.Quadtree.Clear() // Pulisce il quadtree all'inizio di ogni frame

	// Costruisci il quadtree
	for i := range s.Fluid {
		s.Quadtree.Insert(s.Fluid[i])
	}

	// Controlla le collisioni tra particelle e aggiorna le velocità utilizzando il quadtree
	for i := range s.Fluid {
		unitA := s.Fluid[i]
		localFrametime := s.Metrics.Frametime

		//log.Println(unitA.Velocity)

		//log.Panic(unitA.Velocity)
		nearUnits := []*Unit{}
		s.Quadtree.Retrieve(&nearUnits, unitA)
		for _, unitB := range nearUnits {
			if unitA.Id != unitB.Id {

				if collisionTime, collided := findCollisionTime(unitA, unitB, localFrametime); collided {
					calculateCollisionWithQuadtree(
						collisionTime,
						unitA,
						unitB,
					)
					localFrametime -= collisionTime
				}
			}
		}
	}

	// Aggiorna le unità nell'ordine ottenuto
	for _, unit := range s.Fluid {
		if unit != nil {
			unit.accelerate(rl.Vector2{X: 0, Y: s.Config.Gravity})
			unit.checkWallCollision(s.Config)
			unit.updateVelocity(s.Metrics.Frametime)
			unit.updatePosition(s.Metrics.Frametime)
		}
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
		separation := (unitA.Radius + unitB.Radius) - float32(math.Sqrt(float64(distanceSquared)))
		unitA.Position.X -= separation / 2
		unitA.Position.Y -= separation / 2
		unitB.Position.X += separation / 2
		unitB.Position.Y += separation / 2
		return 0, true
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

func calculateCollisionWithQuadtree(collisionTime float32, unitA, unitB *Unit) {
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

	// Calcola la distanza di sovrapposizione
	overlap := (unitA.Radius + unitB.Radius) - float32(math.Sqrt(float64(deltaX*deltaX+deltaY*deltaY)))

	if overlap > 0 {
		// Calcola la direzione di spostamento per separare le particelle
		moveX := float32(normalX) * overlap / 2
		moveY := float32(normalY) * overlap / 2

		// Sposta le particelle fuori dalla sovrapposizione
		unitA.Position.X -= moveX
		unitA.Position.Y -= moveY
		unitB.Position.X += moveX
		unitB.Position.Y += moveY
	}
}
