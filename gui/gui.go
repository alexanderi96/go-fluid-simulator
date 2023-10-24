package gui

import (
	"fmt"

	"github.com/alexanderi96/go-fluid-simulator/physics"
	"github.com/alexanderi96/go-fluid-simulator/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Draw(s *physics.Simulation) {
	rl.BeginDrawing()
	rl.ClearBackground(rl.White)

	drawSidebar(s)
	drawFluid(s)

	if s.Config.ShowOverlay {
		for _, unit := range s.Fluid {
			drawOverlay(unit)
		}
	}

	if s.MouseButtonPressed && s.InitialMousePosition.X > 0 && s.InitialMousePosition.X < float32(s.Config.WindowWidth-s.Config.SidebarWidth) {
		rl.DrawLineEx(s.InitialMousePosition, s.CurrentMousePosition, 5, rl.Black)
	}

	rl.EndDrawing()

}

func drawFluid(s *physics.Simulation) {
	for _, unit := range s.Fluid {

		color := unit.Color
		if s.Config.ShowSpeedColor {
			if s.Config.UseExperimentalQuadtree {
				//color = utils.GetColorFromVelocity(unit.Velocity)
			} else {
				color = utils.GetColorFromVelocity(unit.Velocity(s.Metrics.Frametime))
			}
		}

		if s.Config.ShowVectors {
			drawVectors(unit, s.Metrics.Frametime)
		}

		rl.DrawCircleV(unit.Position, unit.Radius, color)
	}
}

func drawOverlay(u *physics.Unit) {
	mouseX := float32(rl.GetMouseX())
	mouseY := float32(rl.GetMouseY())

	if rl.CheckCollisionPointCircle(rl.NewVector2(mouseX, mouseY), u.Position, u.Radius) {

		overlayText := fmt.Sprintf(
			"ID: %s\nRadius: %.2f\nMassMultiplier: %.2f\nElasticity: %.2f",
			u.Id,
			u.Radius,
			u.MassMultiplier,
			u.Elasticity,
		)
		x := int32(u.Position.X + u.Radius + 10)
		y := int32(u.Position.Y - u.Radius - 10)

		textWidth := rl.MeasureText(overlayText, 20)
		textHeight := 35 * 3

		rl.DrawRectangle(x-5, y-5, textWidth+10, int32(textHeight+10), rl.Color{255, 255, 255, 128})

		rl.DrawText(overlayText, x, y, 20, rl.Black)
	}
}

func drawVectors(u *physics.Unit, dt float32) {

	endVelocity := rl.Vector2Add(u.Position, rl.Vector2Scale(u.Velocity(dt), 0.1))

	rl.DrawLineEx(u.Position, endVelocity, 5, rl.Blue)

	endAcceleration := rl.Vector2Add(u.Position, rl.Vector2Scale(u.Acceleration, 0.1))

	rl.DrawLineEx(u.Position, endAcceleration, 5, rl.Orange)
}
