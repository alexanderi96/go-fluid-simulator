package main

import (
	"fmt"
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Fluid struct {
	viscosity      float32
	surfaceTension float32
	pressure       float32
	temperature    float32
	units          []Unit
}

type Unit struct {
	pos     Vector2D
	prevPos Vector2D
	//velocity Vector2D
	mass   float32
	radius float32
	color  color.RGBA
}

type Vector2D struct {
	x float32
	y float32
}

const ( // 16ms per frame, corrispondente a 60 FPS
	pixelToMeter float32 = 0.01                // 1 pixel è uguale a 0.01 metri
	gravity      float32 = 9.81 * pixelToMeter //* frameTime * frameTime // 9.81 m/s^2 convertito in unità di pixel/frame^2
)

var (
	screenWidth  int32 = 1280
	screenHeight int32 = 720
	sidebarWidth int32 = 200
	gameWidth    int32 = screenWidth - sidebarWidth

	desiredFPS int32 = 999 // Imposta gli FPS desiderati qui
	frameTime        = 1.0 / float32(desiredFPS)

	unitNumber int     = 500
	unitRadius float32 = 6
	unitMass   float32 = 1
)

func (u *Unit) SetColorFromVelocity() {
	velocity := u.pos.Sub(u.prevPos)
	magnitude := math.Sqrt(float64(velocity.x*velocity.x + velocity.y*velocity.y))

	// Utilizza il quadrato della magnitudine per un cambiamento più lento
	normalizedMagnitude := magnitude * magnitude * 1 // Rivedi il fattore di normalizzazione

	// Applica un fattore di smorzamento
	smoothingFactor := 0.9
	normalizedMagnitude = smoothingFactor*normalizedMagnitude + (1-smoothingFactor)*normalizedMagnitude

	red := uint8(math.Min(255, 255*normalizedMagnitude))
	blue := uint8(math.Min(255, 255*(1-normalizedMagnitude)))

	u.color = color.RGBA{R: red, G: 0, B: blue, A: 255}
}

// Calcola il volume di una singola unità
func (u *Unit) Volume() float32 {
	return math.Pi * u.radius * u.radius
}

// Calcola la densità del fluido
func (f *Fluid) Density() float32 {
	var totalMass, totalVolume float32
	for _, unit := range f.units {
		totalMass += unit.mass
		totalVolume += unit.Volume()
	}
	return totalMass / totalVolume
}

// Aggiunge una unità al fluido
func (f *Fluid) AddUnit(unit Unit) {
	f.units = append(f.units, unit)
}

// Rimuove una unità dal fluido dato un indice
func (f *Fluid) RemoveUnit(index int) {
	f.units = append(f.units[:index], f.units[index+1:]...)
}

func (v Vector2D) Add(other Vector2D) Vector2D {
	return Vector2D{v.x + other.x, v.y + other.y}
}

func (v Vector2D) Sub(other Vector2D) Vector2D {
	return Vector2D{v.x - other.x, v.y - other.y}
}

func (v Vector2D) Mul(scalar float32) Vector2D {
	return Vector2D{v.x * scalar, v.y * scalar}
}

func acceleration(unit Unit, otherUnits []Unit) Vector2D {
	accel := Vector2D{0, gravity} // Starting with gravity
	// Add other forces here
	return accel
}

func resetField(screenWidth, screenHeight, totalUnits int) Fluid {
	var units []Unit
	centerX := float32(screenWidth) / 2
	centerY := float32(screenHeight) / 2

	gap := float32(1) // Spazio tra unità

	// Calcolo del raggio totale in base al numero totale di unità
	maxRadius := math.Sqrt(float64(totalUnits)/math.Pi) * float64(unitRadius*2+gap)

	for r := float64(unitRadius + gap); r < maxRadius; r += float64(unitRadius*2 + gap) {
		// Calcolare il numero di unità che possono stare in un cerchio di raggio r
		circumference := 2 * math.Pi * r
		numUnits := int(circumference) / int(unitRadius*2+gap)

		for i := 0; i < numUnits; i++ {
			angle := float32(i) * (math.Pi * 2 / float32(numUnits))

			x := centerX + float32(r*math.Cos(float64(angle)))
			y := centerY + float32(r*math.Sin(float64(angle)))
			unit := Unit{
				pos:     Vector2D{x, y},
				prevPos: Vector2D{x, y},
				mass:    unitMass,
				radius:  unitRadius,
			}
			units = append(units, unit)
		}
	}

	return Fluid{units: units}
}

// Function to handle collisions between two units
func handleCollision(unit1 *Unit, unit2 *Unit) {
	dx := unit1.pos.x - unit2.pos.x
	dy := unit1.pos.y - unit2.pos.y
	distance := math.Sqrt(float64(dx*dx + dy*dy))
	minDistance := unit1.radius + unit2.radius

	if distance < float64(minDistance) {
		overlap := float64(minDistance) - distance

		// Compute normal vectors
		nx := dx / float32(distance)
		ny := dy / float32(distance)

		// Move unit1 and unit2 so that they no longer overlap
		unit1.pos.x += float32(overlap) / 2 * nx
		unit1.pos.y += float32(overlap) / 2 * ny
		unit2.pos.x -= float32(overlap) / 2 * nx
		unit2.pos.y -= float32(overlap) / 2 * ny
	}
}

// Funzione per gestire tutte le collisioni
func handleAllCollisions(fluid *Fluid) {
	for i := 0; i < len(fluid.units); i++ {
		for j := i + 1; j < len(fluid.units); j++ {
			handleCollision(&fluid.units[i], &fluid.units[j])
		}
	}
}

// Funzione per applicare l'integrazione di Verlet
func VerletIntegration(unit *Unit) {
	//currentPos := unit.pos
	accel := acceleration(*unit, nil) // Passiamo nil perché non stiamo considerando altre forze

	// Verlet integration
	tempPos := unit.pos
	unit.pos = unit.pos.Add(unit.pos.Sub(unit.prevPos)).Add(accel.Mul(10 * frameTime))
	unit.prevPos = tempPos
}

// Funzione per gestire la fisica e il movimento delle unità
// Funzione per gestire la fisica e il movimento delle unità
func updateUnits(fluid *Fluid) {
	for i, unit := range fluid.units {
		// Applica l'integrazione di Verlet
		VerletIntegration(&unit)

		// Gestire la collisione con i bordi
		if unit.pos.x <= 0 || unit.pos.x >= float32(gameWidth) {
			unit.prevPos.x = unit.pos.x // Aggiorna la posizione precedente
			unit.pos.x = float32(math.Max(0, math.Min(float64(unit.pos.x), float64(gameWidth))))
		}
		if unit.pos.y <= 0 || unit.pos.y >= float32(screenHeight) {
			unit.prevPos.y = unit.pos.y // Aggiorna la posizione precedente
			unit.pos.y = float32(math.Max(0, math.Min(float64(unit.pos.y), float64(screenHeight))))
		}

		unit.SetColorFromVelocity() // Imposta il colore in base alla velocità

		rl.DrawCircleV(rl.NewVector2(unit.pos.x, unit.pos.y), unit.radius, unit.color)

		fluid.units[i] = unit // Aggiorna l'unità nel fluido

	}
}

func drawSidebar() {
	// Disegna il rettangolo della barra laterale
	rl.DrawRectangle(gameWidth, 0, sidebarWidth, screenHeight, rl.Gray)

	// Disegna il testo per gli FPS
	fpsText := fmt.Sprintf("FPS: %d", rl.GetFPS())
	rl.DrawText(fpsText, screenWidth-190, 10, 20, rl.Black)

	// Qui potresti aggiungere altri controlli o informazioni
}

func main() {
	rl.InitWindow(screenWidth, screenHeight, "Raylib Go Fluid Simulation")
	rl.SetTargetFPS(desiredFPS)

	fluid := resetField(int(gameWidth), int(screenHeight), unitNumber)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		if rl.IsKeyPressed(rl.KeyR) {
			fluid = resetField(int(gameWidth), int(screenHeight), unitNumber) // Resetta il campo
		}

		updateUnits(&fluid)
		handleAllCollisions(&fluid)

		drawSidebar() // Disegna la barra laterale

		rl.EndDrawing()
	}

	rl.CloseWindow()
}
