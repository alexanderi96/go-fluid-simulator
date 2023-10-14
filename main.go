package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"sync"

	gui "github.com/gen2brain/raylib-go/raygui"
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
	pos     rl.Vector2
	prevPos rl.Vector2
	// futurePos rl.Vector2
	//velocity   rl.Vector2
	mass   float32
	radius float32
	color  color.RGBA
}

type Quadtree struct {
	boundary                                   Rectangle // Bounding box
	capacity                                   int       // Numero massimo di punti per quad
	points                                     []Unit
	divided                                    bool
	northWest, northEast, southWest, southEast *Quadtree
}

type Rectangle struct {
	x, y, w, h float32
}

const ( // 16ms per frame, corrispondente a 60 FPS
	pixelToMeter float32 = 0.01                // 1 pixel è uguale a 0.01 metri
	gravity      float32 = 9.81 * pixelToMeter // 9.81 m/s^2 convertito in unità di pixel/frame^2

	queueLength = 500 // Length of the data queue

)

var (
	screenWidth  int32 = 1600
	screenHeight int32 = 900
	sidebarWidth int32 = 400
	gameWidth    int32 = screenWidth - sidebarWidth

	isPause bool = false

	desiredFPS   int32 = 999 // Imposta gli FPS desiderati qui
	maxFPS       int32 = 0
	minFPS       int32 = desiredFPS
	avgFPS       int32 = 0
	frameCount   int32 = 0
	fpsSum       int32 = 0
	framesToSkip int32 = 2 // Numero di frame da ignorare all'inizio
	fpsQueue     [queueLength]uint64
	metrics      = make(map[string]uint64)
	memStats     runtime.MemStats

	unitNumber int     = 100
	unitRadius float32 = 10
	unitMass   float32 = 2

	globalQuadtree = NewQuadtree(Rectangle{0, 0, float32(gameWidth), float32(screenHeight)}, 4)
	fluid          = resetField()

	lastSampleTime     float64
	heapUsageQueue     [queueLength]uint64 // Queue for heap usage data
	stackUsageQueue    [queueLength]uint64 // Queue for stack usage data
	numGoroutinesQueue [queueLength]uint64 // Queue for number of Goroutines

	mu sync.Mutex
)

// Function to update a queue with a new data point
func updateQueue(queue *[queueLength]uint64, newData uint64) {
	// Shift all the existing data to the left by 1
	for i := 0; i < queueLength-1; i++ {
		queue[i] = queue[i+1]
	}
	// Insert the new data at the end
	queue[queueLength-1] = newData
}

// Function to draw a graph for a data queue
func drawGraph(data *[queueLength]uint64, x, y, width, height int32, color rl.Color) {
	// Trova il valore massimo per la normalizzazione
	var maxValue uint64 = 0
	for _, value := range data {
		if value > maxValue {
			maxValue = value
		}
	}

	// Se il valore massimo è zero, impostalo a 1 per evitare divisioni per zero
	if maxValue == 0 {
		maxValue = 1
	}

	rl.DrawRectangleLines(x, y, width, height, color)
	prevX := x
	prevY := y + height

	for i, value := range data {
		normalizedValue := float32(value) / float32(maxValue) * float32(height)
		currentX := x + int32(float32(i)*(float32(width)/float32(len(data))))
		currentY := y + height - int32(normalizedValue)
		rl.DrawLine(prevX, prevY, currentX, currentY, color)
		prevX = currentX
		prevY = currentY
	}
}

func NewQuadtree(boundary Rectangle, capacity int) *Quadtree {
	return &Quadtree{
		boundary: boundary,
		capacity: capacity,
		points:   make([]Unit, 0, capacity),
		divided:  false,
	}
}

func (qt *Quadtree) Insert(unit Unit) bool {
	// Se l'unità non è nel quadtree, ritorna
	if !qt.boundary.contains(unit) {
		return false
	}

	// Se c'è spazio in questo quadtree e non è diviso, inserisci l'unità
	if len(qt.points) < qt.capacity && !qt.divided {
		qt.points = append(qt.points, unit)
		return true
	}

	// Altrimenti, suddividi e poi inserisci l'unità nei sottonodi
	if !qt.divided {
		qt.Subdivide()
	}

	if qt.northWest.Insert(unit) {
		return true
	}
	if qt.northEast.Insert(unit) {
		return true
	}
	if qt.southWest.Insert(unit) {
		return true
	}
	if qt.southEast.Insert(unit) {
		return true
	}

	return false
}

func (qt *Quadtree) Subdivide() {
	x, y, w, h := qt.boundary.x, qt.boundary.y, qt.boundary.w, qt.boundary.h
	halfW, halfH := w/2, h/2

	qt.northWest = NewQuadtree(Rectangle{x, y, halfW, halfH}, qt.capacity)
	qt.northEast = NewQuadtree(Rectangle{x + halfW, y, halfW, halfH}, qt.capacity)
	qt.southWest = NewQuadtree(Rectangle{x, y + halfH, halfW, halfH}, qt.capacity)
	qt.southEast = NewQuadtree(Rectangle{x + halfW, y + halfH, halfW, halfH}, qt.capacity)

	qt.divided = true
}
func (qt *Quadtree) Query(unit Unit) []Unit {
	var found []Unit
	// Se l'unità non è nel quadtree, ritorna un array vuoto
	if !qt.boundary.contains(unit) {
		return found
	}

	// Controlla le collisioni con le unità in questo quadtree
	for _, point := range qt.points {
		if unit != point { // Evita di collidere con se stessa
			found = append(found, point)
		}
	}

	// Se questo quadtree è diviso, cerca anche nei sottonodi
	if qt.divided {
		found = append(found, qt.northWest.Query(unit)...)
		found = append(found, qt.northEast.Query(unit)...)
		found = append(found, qt.southWest.Query(unit)...)
		found = append(found, qt.southEast.Query(unit)...)
	}

	return found
}
func (r Rectangle) contains(unit Unit) bool {
	x, y := unit.pos.X, unit.pos.Y
	return x >= r.x && x <= r.x+r.w && y >= r.y && y <= r.y+r.h
}

func (qt *Quadtree) Clear() {
	qt.points = qt.points[:0] // Svuota la slice dei punti
	qt.divided = false        // Imposta divided a false
	// Clear eventuali sottonodi
	if qt.northWest != nil {
		qt.northWest.Clear()
		qt.northWest = nil
	}
	if qt.northEast != nil {
		qt.northEast.Clear()
		qt.northEast = nil
	}
	if qt.southWest != nil {
		qt.southWest.Clear()
		qt.southWest = nil
	}
	if qt.southEast != nil {
		qt.southEast.Clear()
		qt.southEast = nil
	}
}

func (u *Unit) SetColorFromVelocity() {
	velocity := rl.Vector2Subtract(u.pos, u.prevPos)
	magnitude := math.Sqrt(float64(velocity.X*velocity.X + velocity.Y*velocity.Y))

	normalizedMagnitude := magnitude // Usa la magnitudine direttamente per una transizione più rapida

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

	u.color = color.RGBA{R: red, G: green, B: blue, A: 255}
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

func acceleration(unit Unit, otherUnits []Unit) rl.Vector2 {
	accel := rl.Vector2{X: 0, Y: gravity} // Starting with gravity
	// Add other forces here
	return accel
}

func resetField() Fluid {
	units := make([]Unit, 0, unitNumber) // Pre-allocazione
	centerX := float32(gameWidth) / 2
	centerY := float32(screenHeight) / 2

	gap := float32(1) // Spazio tra unità

	// Calcolo del raggio totale in base al numero totale di unità
	maxRadius := math.Sqrt(float64(unitNumber)/math.Pi) * float64(unitRadius*2+gap)

	// Creazione di una funzione di controllo delle collisioni
	isColliding := func(unit Unit) bool {
		for _, existingUnit := range units {
			if rl.CheckCollisionCircles(unit.pos, unit.radius, existingUnit.pos, existingUnit.radius) {
				return true
			}
		}
		return false
	}

	for len(units) < unitNumber {
		for r := float64(unitRadius + gap); r < maxRadius && len(units) < unitNumber; r += float64(unitRadius*2 + gap) {
			// Calcolare il numero di unità che possono stare in un cerchio di raggio r
			circumference := 2 * math.Pi * r
			numUnits := int(circumference) / int(unitRadius*2+gap)

			for i := 0; i < numUnits && len(units) < unitNumber; i++ {
				angle := float32(i) * (math.Pi * 2 / float32(numUnits))

				x := centerX + float32(r*math.Cos(float64(angle)))
				y := centerY + float32(r*math.Sin(float64(angle)))
				unit := Unit{
					pos:     rl.Vector2{X: x, Y: y},
					prevPos: rl.Vector2{X: x, Y: y},
					mass:    unitMass,
					radius:  unitRadius,
				}

				// Controllo delle collisioni prima di aggiungere la particella
				if !isColliding(unit) {
					units = append(units, unit)
				}
			}
		}
	}

	return Fluid{units: units}
}

// Function to handle collisions between two units
func handleCollision(unit1 *Unit, unit2 *Unit) {
	// Utilizza CheckCollisionCircles di Raylib per il controllo delle collisioni
	if rl.CheckCollisionCircles(unit1.pos, unit1.radius, unit2.pos, unit2.radius) {
		// Calcola la distanza e il vettore normale
		distance := rl.Vector2Distance(unit1.pos, unit2.pos)
		overlap := float32(unit1.radius+unit2.radius) - distance
		norm := rl.Vector2Normalize(rl.Vector2Subtract(unit1.pos, unit2.pos))

		// Sposta le unità in modo che non si sovrappongano più
		adjustment := rl.Vector2Scale(norm, overlap/2)
		unit1.pos = rl.Vector2Add(unit1.pos, adjustment)
		unit2.pos = rl.Vector2Subtract(unit2.pos, adjustment)
	}
}

func (qt *Quadtree) Draw() {
	// Disegna il rettangolo del nodo corrente
	rl.DrawRectangleLines(int32(qt.boundary.x), int32(qt.boundary.y), int32(qt.boundary.w), int32(qt.boundary.h), rl.Black)

	// Chiama Draw sui nodi figli, se esistono
	if qt.divided {
		qt.northWest.Draw()
		qt.northEast.Draw()
		qt.southWest.Draw()
		qt.southEast.Draw()
	}
}

func handleAllCollisionsWithQuadtree(fluid *Fluid) {
	// Pulisci e popola il Quadtree
	globalQuadtree.Clear()
	for _, unit := range fluid.units {
		globalQuadtree.Insert(unit)
	}

	// Numero totale di unità
	numUnits := len(fluid.units)

	// Controllo delle collisioni tra tutte le unità
	for i := 0; i < numUnits; i++ {
		nearbyUnits := globalQuadtree.Query(fluid.units[i])
		for j := 0; j < len(nearbyUnits); j++ {
			// Assicurati che non si stia controllando una unità con se stessa
			if &fluid.units[i] != &nearbyUnits[j] {
				handleCollision(&fluid.units[i], &nearbyUnits[j])
			}
		}
	}

	// Aggiorna tutte le unità
	for i := 0; i < numUnits; i++ {
		updateUnit(&fluid.units[i])
	}
}

// Funzione per applicare l'integrazione di Verlet
func VerletIntegration(unit *Unit) {
	//currentPos := unit.pos
	accel := acceleration(*unit, nil) // Passiamo nil perché non stiamo considerando altre forze

	deltaTime := float32(rl.GetFrameTime())
	//deltaTime = float32(0.5 * math.Pow(float64(deltaTime), 2))

	// Verlet integration
	tempPos := unit.pos
	unit.pos = rl.Vector2Add(rl.Vector2Add(unit.pos, rl.Vector2Subtract(unit.pos, unit.prevPos)), rl.Vector2Scale(accel, deltaTime))
	unit.prevPos = tempPos
}

// Funzione per gestire la fisica e il movimento delle unità
func updateUnit(unit *Unit) {

	// Applica l'integrazione di Verlet
	VerletIntegration(unit)

	// Gestire la collisione con i bordi
	unit.handleBorderCollision()
	unit.updateMassAndRadius()

	unit.SetColorFromVelocity() // Imposta il colore in base alla velocità
}

func (unit *Unit) handleBorderCollision() {
	if unit.pos.X <= 0 || unit.pos.X >= float32(gameWidth) {
		unit.prevPos.X = unit.pos.X // Aggiorna la posizione precedente
		unit.pos.X = float32(math.Max(0, math.Min(float64(unit.pos.X), float64(gameWidth))))
	}
	if unit.pos.Y <= 0 || unit.pos.Y >= float32(screenHeight) {
		unit.prevPos.Y = unit.pos.Y // Aggiorna la posizione precedente
		unit.pos.Y = float32(math.Max(0, math.Min(float64(unit.pos.Y), float64(screenHeight))))
	}
}

// Funzione separata per aggiornare massa e raggio
func (unit *Unit) updateMassAndRadius() {
	if unit.mass != unitMass {
		unit.mass = unitMass
	}

	if unit.radius != unitRadius {
		unit.radius = unitRadius
	}
}

func findEmptyPosition(fluid *Fluid, radius float32) rl.Vector2 {
	for {
		// Genera una posizione casuale
		x := float32(rand.Intn(int(gameWidth)))
		y := float32(rand.Intn(int(screenHeight)))
		newPos := rl.Vector2{X: x, Y: y}

		// Controlla se la posizione è vuota
		isEmpty := true
		for _, unit := range fluid.units {
			distance := rl.Vector2Distance(newPos, unit.pos)
			if distance < unit.radius+radius {
				isEmpty = false
				break
			}
		}

		if isEmpty {
			return newPos
		}
	}
}

func (fluid *Fluid) AddRandomUnit() {
	fluid.AddUnit(Unit{
		pos:    findEmptyPosition(fluid, unitRadius),
		mass:   unitMass,
		radius: unitRadius,
	})
}

func (fluid *Fluid) RemoveRandomUnit() {
	fluid.RemoveUnit(rand.Intn(len(fluid.units)))
}

// Function to gather all the performance metrics
func gatherMetrics() map[string]uint64 {

	// Memory Usage
	runtime.ReadMemStats(&memStats)
	metrics["heapUsage"] = memStats.HeapAlloc   // Heap memory in use
	metrics["stackUsage"] = memStats.StackInuse // Stack memory in use

	// Number of Goroutines
	metrics["numGoroutines"] = uint64(runtime.NumGoroutine())

	currentFPS := rl.GetFPS()

	if framesToSkip > 0 {
		framesToSkip--
	} else {
		updateQueue(&fpsQueue, uint64(currentFPS))
		updateQueue(&heapUsageQueue, uint64(metrics["heapUsage"]))
		updateQueue(&stackUsageQueue, uint64(metrics["stackUsage"]))
		updateQueue(&numGoroutinesQueue, uint64(metrics["numGoroutines"]))

		frameCount++
		fpsSum += currentFPS

		if currentFPS > maxFPS {
			maxFPS = currentFPS
		}
		if currentFPS < minFPS {
			minFPS = currentFPS
		}
		avgFPS = fpsSum / frameCount
	}

	return metrics
}

func drawFPSInfo(xStart *int32, yStart *int32) {
	fpsText := fmt.Sprintf("FPS: %d Avg FPS: %d\nMax FPS: %d Min FPS: %d", rl.GetFPS(), avgFPS, maxFPS, minFPS)
	rl.DrawText(fpsText, *xStart, *yStart-50, 20, rl.Black)
	*yStart -= 50 + 5
}

func drawGraphInfo(xStart *int32, yStart *int32, queue *[queueLength]uint64, color rl.Color) {
	drawGraph(queue, *xStart, *yStart-100, sidebarWidth-20, 100, color)
	*yStart -= 100 + 5
}

func drawSlider(xStart *int32, yStart *int32, value *int, min, max int) {
	*value = int(gui.Slider(rl.Rectangle{X: float32(*xStart), Y: float32(*yStart), Width: float32(sidebarWidth - 20), Height: 20}, "", "", float32(*value), float32(min), float32(max)))
	*yStart += 20 + 5
}

func drawFloatSlider(xStart *int32, yStart *int32, value *float32, min, max int) {
	*value = float32(gui.Slider(rl.Rectangle{X: float32(*xStart), Y: float32(*yStart), Width: float32(sidebarWidth - 20), Height: 20}, "", "", *value, float32(min), float32(max)))
	*yStart += 20 + 5
}

func drawSidebar() {
	rl.DrawRectangle(gameWidth, 0, sidebarWidth, screenHeight, rl.Gray)

	xStart := gameWidth + 10
	yStartTop := int32(10)
	yStartBottom := screenHeight - 10

	// Parte superiore della sidebar
	drawSlider(&xStart, &yStartTop, &unitNumber, 1, 5000)
	drawFloatSlider(&xStart, &yStartTop, &unitRadius, 1, 10)

	unitNumbers := fmt.Sprintf("Desired Units: %d Actual Units: %d", unitNumber, len(fluid.units))
	rl.DrawText(unitNumbers, xStart, yStartTop, 20, rl.Black)
	yStartTop += 20 + 5

	// Potresti continuare qui con altri slider o elementi nella parte superiore

	// Parte inferiore della sidebar
	drawFPSInfo(&xStart, &yStartBottom)
	drawGraphInfo(&xStart, &yStartBottom, &fpsQueue, rl.Black)

	heapUsageText := fmt.Sprintf("Heap Usage: %.2f MB", float64(metrics["heapUsage"])/(1024.0*1024.0))
	rl.DrawText(heapUsageText, xStart, yStartBottom-20, 20, rl.Black)
	yStartBottom -= 20 + 5
	drawGraphInfo(&xStart, &yStartBottom, &heapUsageQueue, rl.Red)

	stackUsageText := fmt.Sprintf("Stack Usage: %.2f MB", float64(metrics["stackUsage"])/(1024.0*1024.0))
	rl.DrawText(stackUsageText, xStart, yStartBottom-20, 20, rl.Black)
	yStartBottom -= 20 + 5
	drawGraphInfo(&xStart, &yStartBottom, &stackUsageQueue, rl.Green)

	numGoroutinesText := fmt.Sprintf("Active Goroutines: %d", metrics["numGoroutines"])
	rl.DrawText(numGoroutinesText, xStart, yStartBottom-20, 20, rl.Black)
	yStartBottom -= 20 + 5
	drawGraphInfo(&xStart, &yStartBottom, &numGoroutinesQueue, rl.Blue)
	// Potresti continuare qui con altri grafici o elementi nella parte inferiore
}

func (fluid *Fluid) Draw() {
	for _, unit := range fluid.units {
		rl.DrawCircleV(unit.pos, unit.radius, unit.color)
	}
}

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(screenWidth, screenHeight, "Raylib Go Fluid Simulation")
	rl.SetTargetFPS(desiredFPS)

	fluid = resetField()

	metrics = gatherMetrics()

	for !rl.WindowShouldClose() {
		currentSampleTime := rl.GetTime()
		if currentSampleTime-lastSampleTime >= 0.5 {
			// Aggiorna le code qui
			metrics = gatherMetrics()
			lastSampleTime = currentSampleTime
		}

		// Gestione del ridimensionamento della finestra
		screenWidth, screenHeight = int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight())
		gameWidth = screenWidth - sidebarWidth

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		if rl.IsKeyPressed(rl.KeyR) {
			fluid = resetField() // Resetta il campo
		} else if rl.IsKeyPressed(rl.KeySpace) {
			isPause = !isPause
		}

		if !isPause {
			handleAllCollisionsWithQuadtree(&fluid)

			if unitNumber > len(fluid.units) {
				fluid.AddRandomUnit()
			} else if unitNumber < len(fluid.units) {
				fluid.RemoveRandomUnit()
			}
		}

		fluid.Draw()
		globalQuadtree.Draw()

		drawSidebar() // Disegna la barra laterale

		rl.EndDrawing()
	}

	rl.CloseWindow()
}
