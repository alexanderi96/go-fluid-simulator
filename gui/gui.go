package gui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/alexanderi96/go-fluid-simulator/physics"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func Draw(s *physics.Simulation) {
	drawSidebar(s)
	drawFluid(s.Fluid)
}

func drawSidebar(s *physics.Simulation) {
	rl.DrawRectangle(s.Config.GameWidth, 0, s.Config.WindowWidth, s.Config.WindowHeight, rl.Gray)

	border := int32(10)
	xStart := s.Config.GameWidth + border
	yStartTop := border
	//yStartBottom := s.Config.WindowHeight - border

	unitNumbers := fmt.Sprintf("Actual Units: %d", len(s.Fluid.Units))
	rl.DrawText(unitNumbers, xStart, yStartTop, 20, rl.Black)
	yStartTop += 20 + 5

}

func drawFluid(f *physics.Fluid) {
	for _, unit := range f.Units {
		unit.Color = getColorFromVelocity(unit.Velocity)
		rl.DrawCircleV(unit.Position, unit.Radius, unit.Color)
	}
}

func getColorFromVelocity(v rl.Vector2) color.RGBA {
	magnitude := math.Sqrt(float64(v.X*v.X + v.Y*v.Y))

	//normalizedMagnitude := magnitude // Usa la magnitudine direttamente per una transizione più rapida
	normalizedMagnitude := math.Min(magnitude/100, 1.0)
	var red, green, blue uint8

	if normalizedMagnitude < 0.5 {
		// Bassa velocità: Blu
		blue = 255
		red = uint8(255 * normalizedMagnitude * 2)
		green = 0
	} else if normalizedMagnitude < 1 {
		// Velocità media: Transizione verso il Rosso
		blue = uint8(255 * (1 - normalizedMagnitude) * 2)
		red = 255
		green = 0
	} else {
		// Alta velocità: Transizione verso il Giallo
		red = 255
		green = uint8(255 * (normalizedMagnitude - 1))
		blue = 0
	}

	return color.RGBA{R: red, G: green, B: blue, A: 255}
}
