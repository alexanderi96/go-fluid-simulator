package physics

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func (s *Simulation) UpdateWithVerletIntegration() error {
	for _, unitA := range s.Fluid {
		if unitA == nil {
			continue
		}

		for _, unitB := range s.Fluid {
			if unitB == nil {
				continue
			}

			if unitA.Id != unitB.Id && areOverlapping(unitA, unitB) {
				calculateCollisionWithVerlet(unitA, unitB)
			}
		}
	}

	for _, unit := range s.Fluid {
		if s.Config.ApplyGravity {
			unit.accelerate(rl.Vector2{X: 0, Y: s.Config.Gravity})
		}
		unit.updatePositionWithVerlet(s.Metrics.Frametime)
		unit.checkWallCollisionVerlet(s.Config, s.Metrics.Frametime)
	}

	return nil

}
