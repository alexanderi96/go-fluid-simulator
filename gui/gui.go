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
	if s.Config.ShowQuadtree {
		drawQuadtree(s.Quadtree)
	}
}

func drawQuadtree(quadtree *physics.Quadtree) {
	if quadtree == nil {
		return // Ritorna se il quadtree è nil
	}

	// Converte i valori float32 del rettangolo in int32 per rl.DrawLine
	x1, y1 := int32(quadtree.Bounds.X), int32(quadtree.Bounds.Y)
	x2, y2 := int32(quadtree.Bounds.X+quadtree.Bounds.Width), int32(quadtree.Bounds.Y+quadtree.Bounds.Height)

	// Disegna i bordi del quadtree corrente
	rl.DrawLine(x1, y1, x2, y1, rl.Green) // Linea superiore
	rl.DrawLine(x2, y1, x2, y2, rl.Green) // Linea destra
	rl.DrawLine(x2, y2, x1, y2, rl.Green) // Linea inferiore
	rl.DrawLine(x1, y2, x1, y1, rl.Green) // Linea sinistra

	// Disegna ricorsivamente i bordi dei sotto-quadtrees
	for _, child := range quadtree.Children {
		drawQuadtree(child)
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