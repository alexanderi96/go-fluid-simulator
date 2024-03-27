package gui

import (
	"github.com/alexanderi96/go-fluid-simulator/physics"
	"github.com/alexanderi96/go-fluid-simulator/utils"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Draw(s *physics.Simulation) {
	rl.BeginDrawing()
	rl.ClearBackground(rl.RayWhite)

	rl.BeginMode3D(s.Camera)
	drawGameArea(s)

	// Disegno di un pavimento a griglia come punto di riferimento
	rl.DrawGrid(20, 1.0)

	// Disegno di alcuni oggetti come punti di riferimento
	rl.DrawCube(rl.NewVector3(-3.0, 1.5, 0), 1, 3, 1, rl.Blue)
	rl.DrawCube(rl.NewVector3(3.0, 1.5, 0), 1, 3, 1, rl.Red)

	drawFluid(s)

	// if s.Config.ShowOverlay {
	// 	for _, unit := range s.Fluid {
	// 		drawOverlay(unit)
	// 	}
	// }

	if s.MouseButtonPressed && s.InitialMousePosition.X > 0 && s.InitialMousePosition.X < float32(s.Config.WindowWidth-s.Config.SidebarWidth) {
		rl.DrawLineEx(s.InitialMousePosition, s.CurrentMousePosition, 5, rl.Black)
	}

	rl.EndMode3D()
	drawSidebar(s)

	rl.DrawFPS(10, 10)
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

		// if s.Config.ShowVectors {
		// 	drawVectors(unit, s.Metrics.Frametime)
		// }

		rl.DrawSphere(rl.NewVector3(unit.Position.X, unit.Position.Y, unit.Position.Z), unit.Radius, color)
	}
}

// func drawOverlay(u *physics.Unit) {
// 	mouseX := float32(rl.GetMouseX())
// 	mouseY := float32(rl.GetMouseY())

// 	if rl.CheckCollisionPointCircle(rl.NewVector2(mouseX, mouseY), u.Position, u.Radius) {

// 		overlayText := fmt.Sprintf(
// 			"ID: %s\nRadius: %.2f\nMassMultiplier: %.2f\nElasticity: %.2f",
// 			u.Id,
// 			u.Radius,
// 			u.MassMultiplier,
// 			u.Elasticity,
// 		)
// 		x := int32(u.Position.X + u.Radius + 10)
// 		y := int32(u.Position.Y - u.Radius - 10)

// 		textWidth := rl.MeasureText(overlayText, 20)
// 		textHeight := 35 * 3

// 		rl.DrawRectangle(x-5, y-5, textWidth+10, int32(textHeight+10), rl.Color{255, 255, 255, 128})

// 		rl.DrawText(overlayText, x, y, 20, rl.Black)
// 	}
// }

// func drawVectors(u *physics.Unit, dt float32) {

// 	endVelocity := rl.Vector2Add(u.Position, rl.Vector2Scale(u.Velocity(dt), 0.1))

// 	rl.DrawLineEx(u.Position, endVelocity, 5, rl.Blue)

// 	endAcceleration := rl.Vector2Add(u.Position, rl.Vector2Scale(u.Acceleration, 0.1))

// 	rl.DrawLineEx(u.Position, endAcceleration, 5, rl.Orange)
// }

func drawGameArea(s *physics.Simulation) {
	// Vertici del cubo
	v1 := rl.NewVector3(0, 0, 0)
	v2 := rl.NewVector3(float32(s.Config.GameX), 0, 0)
	v3 := rl.NewVector3(float32(s.Config.GameX), float32(s.Config.GameY), 0)
	v4 := rl.NewVector3(0, float32(s.Config.GameY), 0)
	v5 := rl.NewVector3(0, 0, float32(s.Config.GameZ))
	v6 := rl.NewVector3(float32(s.Config.GameX), 0, float32(s.Config.GameZ))
	v7 := rl.NewVector3(float32(s.Config.GameX), float32(s.Config.GameY), float32(s.Config.GameZ))
	v8 := rl.NewVector3(0, float32(s.Config.GameY), float32(s.Config.GameZ))

	// Disegna le linee del cubo
	// Base inferiore
	rl.DrawLine3D(v1, v2, rl.Black)
	rl.DrawLine3D(v2, v3, rl.Black)
	rl.DrawLine3D(v3, v4, rl.Black)
	rl.DrawLine3D(v4, v1, rl.Black)
	// Colonne
	rl.DrawLine3D(v1, v5, rl.Black)
	rl.DrawLine3D(v2, v6, rl.Black)
	rl.DrawLine3D(v3, v7, rl.Black)
	rl.DrawLine3D(v4, v8, rl.Black)
	// Base superiore
	rl.DrawLine3D(v5, v6, rl.Black)
	rl.DrawLine3D(v6, v7, rl.Black)
	rl.DrawLine3D(v7, v8, rl.Black)
	rl.DrawLine3D(v8, v5, rl.Black)
}
