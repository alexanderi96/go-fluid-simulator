package gui

import (
	"fmt"
	"strconv"

	"github.com/alexanderi96/go-fluid-simulator/physics"
	"github.com/alexanderi96/go-fluid-simulator/utils"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func Draw(s *physics.Simulation) {
	rl.BeginDrawing()
	rl.ClearBackground(rl.LightGray)

	drawSidebar(s)
	drawFluid(s)

	if s.Config.ShowOverlay {
		for _, unit := range s.Fluid {
			drawOverlay(unit)
		}
	}
	rl.EndDrawing()

}

func drawSidebar(s *physics.Simulation) error {
	rl.DrawRectangle(s.Config.ViewportX, 0, s.Config.WindowWidth, s.Config.WindowHeight, rl.RayWhite)

	border := int32(10)
	xStart := s.Config.ViewportX + border
	yStartTop := border

	sliderLength := float32(s.Config.SidebarWidth - 2*border)
	sliderThickness := float32(20)
	//yStartBottom := s.Config.WindowHeight - border

	frametime := fmt.Sprintf("Frametime: %f", s.Metrics.Frametime)
	rl.DrawText(frametime, xStart, yStartTop, 20, rl.Black)
	yStartTop += 20 + 5

	fps := fmt.Sprintf("FPS: %d", s.Metrics.FPS)
	rl.DrawText(fps, xStart, yStartTop, 20, rl.Black)
	yStartTop += 20 + 5

	heapSize := fmt.Sprintf("Heap Size: %d kb", s.Metrics.HeapSize)
	rl.DrawText(heapSize, xStart, yStartTop, 20, rl.Black)
	yStartTop += 20 + 5

	quadtree := fmt.Sprintf("Using qTree: %t", s.Config.UseExperimentalQuadtree)
	rl.DrawText(quadtree, xStart, yStartTop, 20, rl.Black)
	yStartTop += 40 + 5

	selectedUnitNumbers := fmt.Sprintf("Selected Units: %d", s.Config.ParticleNumber)
	rl.DrawText(selectedUnitNumbers, xStart, yStartTop, 20, rl.Black)
	yStartTop += 20 + 5

	unitNumbers := fmt.Sprintf("Spawned Units: %d", len(s.Fluid))
	rl.DrawText(unitNumbers, xStart, yStartTop, 20, rl.Black)
	yStartTop += 20 + 5

	s.Config.ParticleNumber = int32(gui.Slider(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: sliderLength, Height: sliderThickness}, "", "", float32(s.Config.ParticleNumber), 1, 1000))
	yStartTop += 20 + 5

	s.Config.ApplyGravity = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Apply Gravity", s.Config.ApplyGravity)
	yStartTop += 20 + 5

	s.Config.SetRandomColor = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Set Random Color", s.Config.SetRandomColor)
	yStartTop += 20 + 5

	s.Config.ShowSpeedColor = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Show Speed Color", s.Config.ShowSpeedColor)
	yStartTop += 20 + 5

	s.Config.ShowOverlay = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Show Overlay", s.Config.ShowOverlay)
	yStartTop += 20 + 5

	s.Config.SetRandomRadius = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Set Random Radius", s.Config.SetRandomRadius)
	yStartTop += 20 + 5

	if s.Config.SetRandomRadius {
		radMinText := strconv.FormatFloat(float64(s.Config.RadiusMin), 'f', 2, 64)

		radMinInput := rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 200, Height: 30}

		radMinEdited := gui.TextBox(radMinInput, &radMinText, 10, true)
		yStartTop += 30 + 5

		if radMinEdited {

			if radMin, err := utils.CheckTextFloat32(radMinText); err == nil {
				s.Config.RadiusMin = radMin
			}
		}

		radMaxText := strconv.FormatFloat(float64(s.Config.RadiusMax), 'f', 2, 64)

		radMaxInput := rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 200, Height: 30}

		radMaxEdited := gui.TextBox(radMaxInput, &radMaxText, 10, true)
		yStartTop += 30 + 5

		if radMaxEdited {

			if radMax, err := utils.CheckTextFloat32(radMaxText); err == nil {
				s.Config.RadiusMax = radMax
			}
		}

	}

	s.Config.SetRandomElasticity = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Set Random Elasticity", s.Config.SetRandomElasticity)
	yStartTop += 20 + 5

	if s.Config.SetRandomElasticity {
		elasticityMinText := strconv.FormatFloat(float64(s.Config.ElasticityMin), 'f', 2, 64)

		elasticityMinInput := rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 200, Height: 30}

		elasticityMinEdited := gui.TextBox(elasticityMinInput, &elasticityMinText, 10, true)
		yStartTop += 30 + 5

		if elasticityMinEdited {

			if elasticityMin, err := utils.CheckTextFloat32(elasticityMinText); err == nil {
				s.Config.ElasticityMin = elasticityMin
			}
		}

		elasticityMaxText := strconv.FormatFloat(float64(s.Config.ElasticityMax), 'f', 2, 64)

		elasticityMaxInput := rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 200, Height: 30}

		elasticityMaxEdited := gui.TextBox(elasticityMaxInput, &elasticityMaxText, 10, true)
		yStartTop += 30 + 5

		if elasticityMaxEdited {

			if elasticityMax, err := utils.CheckTextFloat32(elasticityMaxText); err == nil {
				s.Config.ElasticityMax = elasticityMax
			}
		}

	}

	s.Config.SetRandomMass = gui.CheckBox(rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 20, Height: 20}, "Set Random Mass", s.Config.SetRandomMass)
	yStartTop += 20 + 5

	if s.Config.SetRandomMass {
		massMinText := strconv.FormatFloat(float64(s.Config.MassMin), 'f', 2, 64)

		MassMinInput := rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 200, Height: 30}

		MassMinEdited := gui.TextBox(MassMinInput, &massMinText, 10, true)
		yStartTop += 30 + 5

		if MassMinEdited {
			if MassMin, err := utils.CheckTextFloat32(massMinText); err == nil {
				s.Config.MassMin = MassMin
			}
		}

		massMaxText := strconv.FormatFloat(float64(s.Config.MassMax), 'f', 2, 64)

		MassMaxInput := rl.Rectangle{X: float32(xStart), Y: float32(yStartTop), Width: 200, Height: 30}

		MassMaxEdited := gui.TextBox(MassMaxInput, &massMaxText, 10, true)
		yStartTop += 30 + 5

		if MassMaxEdited {
			if MassMax, err := utils.CheckTextFloat32(massMaxText); err == nil {
				s.Config.MassMax = MassMax
			}
		}

	}

	return nil

}

func drawFluid(s *physics.Simulation) {
	for _, unit := range s.Fluid {

		color := unit.Color
		if s.Config.ShowSpeedColor {
			if s.Config.UseExperimentalQuadtree {
				//color = utils.GetColorFromVelocity(unit.Velocity)
			} else {
				color = utils.GetColorFromVelocity(unit.GetVelocityWithVerlet())
			}
		}

		if s.Config.ShowVectors {
			drawVectors(unit)
		}

		rl.DrawCircleV(unit.Position, unit.Radius, color)
	}
}

func drawOverlay(u *physics.Unit) {
	mouseX := float32(rl.GetMouseX())
	mouseY := float32(rl.GetMouseY())

	if rl.CheckCollisionPointCircle(rl.NewVector2(mouseX, mouseY), u.Position, u.Radius) {

		overlayText := fmt.Sprintf(
			"ID: %s\nRadius: %.2f\nMass: %.2f\nElasticity: %.2f",
			u.Id,
			u.Radius,
			u.Mass,
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

func drawVectors(u *physics.Unit) {

	endVelocity := rl.Vector2Add(u.Position, rl.Vector2Scale(u.GetVelocityWithVerlet(), 0.1))

	rl.DrawLineEx(u.Position, endVelocity, 2, rl.Blue)

	endAcceleration := rl.Vector2Add(u.Position, rl.Vector2Scale(u.Acceleration, 0.1))

	rl.DrawLineEx(u.Position, endAcceleration, 2, rl.Orange)
}
