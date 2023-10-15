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
	if s.Config.ShowVectors {
		drawVectors(s.Fluid)
	}
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

	fps := fmt.Sprintf("FPS: %d", rl.GetFPS())
	rl.DrawText(fps, xStart, yStartTop, 20, rl.Black)
	yStartTop += 20 + 5

}

func drawFluid(f *physics.Fluid) {
	for _, unit := range f.Units {
		unit.Color = getColorFromVelocity(unit.Velocity)
		rl.DrawCircleV(unit.Position, unit.Radius, unit.Color)
	}
}

func drawVectors(f *physics.Fluid) {
	for _, unit := range f.Units {
		// Calcolo della posizione finale del vettore della velocità
		endVelocity := rl.Vector2Add(unit.Position, rl.Vector2Scale(unit.Velocity, 0.1)) // La scala 1.0 ridimensiona la lunghezza del vettore
		// Disegno del vettore della velocità
		rl.DrawLineEx(unit.Position, endVelocity, 2, rl.Blue) // Il vettore della velocità è blu

		// Calcolo della posizione finale del vettore dell'accelerazione
		endAcceleration := rl.Vector2Add(unit.Position, rl.Vector2Scale(unit.Acceleration, 0.1)) // La scala 1.0 ridimensiona la lunghezza del vettore
		// Disegno del vettore dell'accelerazione
		rl.DrawLineEx(unit.Position, endAcceleration, 2, rl.Orange) // Il vettore dell'accelerazione è arancione
	}
}
func getColorFromVelocity(v rl.Vector2) color.RGBA {
	magnitude := math.Sqrt(float64(v.X*v.X + v.Y*v.Y))
	colorFactor := math.Min(1, math.Pow(magnitude/1000, 0.5))

	// Calcola una scala di colori da blu (freddo, lento) a rosso (caldo, veloce)
	R := uint8(255 * colorFactor)
	G := uint8(0)
	B := uint8(255 * (1 - colorFactor))

	return color.RGBA{
		R: R,
		G: G,
		B: B,
		A: 255,
	}
}
