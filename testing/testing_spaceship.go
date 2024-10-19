package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light" // per file .obj
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
)

type GameState struct {
	app           *app.Application
	scene         *core.Node
	cam           *camera.Camera
	plane         *graphic.Mesh
	planets       []*graphic.Mesh
	pointLight    *light.Point
	speed         float32
	maxSpeed      float32
	acceleration  float32
	rotationSpeed float32
	dragCoeff     float32
	airBrake      float32
	keys          map[window.Key]bool
	cameraOffset  *math32.Vector3

	hud struct {
		positionLabel    *gui.Label
		speedLabel       *gui.Label
		directionLabel   *gui.Label
		orientationLabel *gui.Label
		statusLabel      *gui.Label
	}
}

func initGameState(a *app.Application) *GameState {
	return &GameState{
		app:           a,
		scene:         core.NewNode(),
		speed:         0.0,
		maxSpeed:      0.1,
		acceleration:  0.0,
		rotationSpeed: 0.005,
		dragCoeff:     0.00001,
		airBrake:      0.001,
		keys:          make(map[window.Key]bool),
		cameraOffset:  math32.NewVector3(0, 5, -10),
	}
}

func setupHUD(gs *GameState) {
	// Posizione
	gs.hud.positionLabel = gui.NewLabel("")
	gs.hud.positionLabel.SetPosition(10, 10)
	gs.scene.Add(gs.hud.positionLabel)

	// Velocità
	gs.hud.speedLabel = gui.NewLabel("")
	gs.hud.speedLabel.SetPosition(10, 30)
	gs.scene.Add(gs.hud.speedLabel)

	// Direzione di movimento
	gs.hud.directionLabel = gui.NewLabel("")
	gs.hud.directionLabel.SetPosition(10, 50)
	gs.scene.Add(gs.hud.directionLabel)

	// Orientamento
	gs.hud.orientationLabel = gui.NewLabel("")
	gs.hud.orientationLabel.SetPosition(10, 70)
	gs.scene.Add(gs.hud.orientationLabel)

	// Status
	gs.hud.statusLabel = gui.NewLabel("")
	gs.hud.statusLabel.SetPosition(10, 90)
	gs.scene.Add(gs.hud.statusLabel)

	// Aggiungi una bussola grafica semplice
	width, _ := gs.app.GetSize() // Usa GetSize invece di GetWidth
	compassSize := float32(100)
	compass := gui.NewPanel(compassSize, compassSize)
	compass.SetPosition(float32(width)-compassSize-10, 10)

	// Aggiungi indicatori N/S/E/W
	addCompassLabel := func(text string, x, y float32) {
		label := gui.NewLabel(text)
		label.SetPosition(x, y)
		compass.Add(label)
	}

	addCompassLabel("N", compassSize/2-5, 0)
	addCompassLabel("S", compassSize/2-5, compassSize-20)
	addCompassLabel("E", compassSize-20, compassSize/2-10)
	addCompassLabel("W", 0, compassSize/2-10)

	gs.scene.Add(compass)
}

func updateHUD(gs *GameState) {
	// Aggiorna posizione
	pos := gs.plane.Position()
	gs.hud.positionLabel.SetText(fmt.Sprintf("Position: X: %.1f Y: %.1f Z: %.1f", pos.X, pos.Y, pos.Z))

	// Aggiorna velocità
	gs.hud.speedLabel.SetText(fmt.Sprintf("Speed: %.1f units/s", gs.speed))

	// Calcola e aggiorna la direzione
	forward := math32.NewVector3(0, 0, 1)
	matrix := gs.plane.Matrix()
	forward.ApplyMatrix4(&matrix)
	forward.Normalize()

	// Determina le direzioni cardinali
	directions := []string{}

	if forward.Z > 0.3 {
		directions = append(directions, "North")
	}
	if forward.Z < -0.3 {
		directions = append(directions, "South")
	}
	if forward.X > 0.3 {
		directions = append(directions, "East")
	}
	if forward.X < -0.3 {
		directions = append(directions, "West")
	}
	if forward.Y > 0.3 {
		directions = append(directions, "Up")
	}
	if forward.Y < -0.3 {
		directions = append(directions, "Down")
	}

	directionText := strings.Join(directions, "-")
	if directionText == "" {
		directionText = "Neutral"
	}
	gs.hud.directionLabel.SetText(fmt.Sprintf("Direction: %s", directionText))

	// Calcola e aggiorna l'orientamento in gradi
	rot := gs.plane.Rotation()
	gs.hud.orientationLabel.SetText(fmt.Sprintf("Orientation - Pitch: %.1f° Roll: %.1f° Yaw: %.1f°",
		math32.RadToDeg(rot.X),
		math32.RadToDeg(rot.Z),
		math32.RadToDeg(rot.Y)))

	// Aggiornamento status
	var status []string

	if math.Abs(float64(gs.speed)) < 0.001 {
		status = append(status, "HOVERING")
	} else if gs.speed > 0 {
		status = append(status, "FORWARD")
	} else {
		status = append(status, "REVERSE")
	}

	if math.Abs(float64(rot.Z)) > 0.1 {
		if rot.Z > 0 {
			status = append(status, "ROLLING RIGHT")
		} else {
			status = append(status, "ROLLING LEFT")
		}
	}

	if math.Abs(float64(rot.X)) > 0.1 {
		if rot.X > 0 {
			status = append(status, "PITCHING UP")
		} else {
			status = append(status, "PITCHING DOWN")
		}
	}

	statusText := strings.Join(status, " | ")
	gs.hud.statusLabel.SetText(fmt.Sprintf("Status: %s", statusText))
}

func setupCamera(gs *GameState) {
	gs.cam = camera.New(1)
	gs.cam.SetPosition(0, 2, 8)
	gs.scene.Add(gs.cam)

	cameraAxes := helper.NewAxes(1.0)
	gs.cam.Add(cameraAxes)
}

func setupScene(gs *GameState) {
	gui.Manager().Set(gs.scene)

	// Add grid helper
	grid := helper.NewGrid(50, 1, &math32.Color{0.4, 0.4, 0.4})
	gs.scene.Add(grid)

	// Set background color
	gs.app.Gls().ClearColor(0.1, 0.1, 0.2, 1.0)
}

func createSphere(radius float32, color string, x, y, z float32, scene *core.Node) *graphic.Mesh {
	geom := geometry.NewSphere(float64(radius), 32, 32)
	mat := material.NewStandard(math32.NewColor(color))
	sphere := graphic.NewMesh(geom, mat)
	sphere.SetPosition(x, y, z)
	scene.Add(sphere)
	return sphere
}

func setupPlanets(gs *GameState) {
	gs.planets = make([]*graphic.Mesh, 4)
	gs.planets[0] = createSphere(1.0, "Red", 10, 0, 10, gs.scene)
	gs.planets[1] = createSphere(0.7, "Green", -10, 5, -10, gs.scene)
	gs.planets[2] = createSphere(0.5, "Yellow", -5, -3, 15, gs.scene)
	gs.planets[3] = createSphere(0.3, "Purple", 15, 2, -8, gs.scene)
}

func setupPlane(gs *GameState) {
	// Creiamo il corpo principale della nave
	bodyGeom := geometry.NewBox(1, 0.3, 2) // Corpo allungato e sottile
	bodyMat := material.NewStandard(math32.NewColor("Gray"))
	bodyMesh := graphic.NewMesh(bodyGeom, bodyMat)

	// Parte frontale appuntita (cockpit a forma di prisma)
	noseGeom := geometry.NewCone(0.2, 0.5, 4, 1, true) // Cono a base quadrata
	noseMat := material.NewStandard(math32.NewColor("Black"))
	noseMesh := graphic.NewMesh(noseGeom, noseMat)
	noseMesh.SetRotationX(math32.Pi / 2) // Ruotiamo il cono per farlo puntare in avanti
	noseMesh.SetPosition(0, 0.1, 1.2)    // Posizioniamo la parte frontale sulla nave

	// Creiamo le ali larghe, puntate verso il fronte e il basso
	wingGeom := geometry.NewBox(1.5, 0.05, 0.6) // Ali più larghe e sottili
	wingMat := material.NewStandard(math32.NewColor("Gray"))
	leftWingMesh := graphic.NewMesh(wingGeom, wingMat)
	rightWingMesh := graphic.NewMesh(wingGeom, wingMat)

	// Impostiamo l'inclinazione delle ali
	leftWingMesh.SetRotationZ(math32.DegToRad(20))   // Inclinazione verso il basso
	leftWingMesh.SetRotationY(math32.DegToRad(15))   // Inclinazione verso il fronte
	rightWingMesh.SetRotationZ(-math32.DegToRad(20)) // Inclinazione verso il basso
	rightWingMesh.SetRotationY(-math32.DegToRad(15)) // Inclinazione verso il fronte

	leftWingMesh.SetPosition(-0.9, -0.2, -0.3) // Posiziona l'ala sinistra
	rightWingMesh.SetPosition(0.9, -0.2, -0.3) // Posiziona l'ala destra

	// Aggiungiamo dei bracci anteriori (cannoni)
	armGeom := geometry.NewBox(0.05, 0.05, 0.6) // Bracci allungati
	armMat := material.NewStandard(math32.NewColor("Gray"))
	leftArmMesh := graphic.NewMesh(armGeom, armMat)
	rightArmMesh := graphic.NewMesh(armGeom, armMat)

	leftArmMesh.SetPosition(-0.4, 0, 0.8) // Posiziona il braccio sinistro verso la parte anteriore
	rightArmMesh.SetPosition(0.4, 0, 0.8) // Posiziona il braccio destro

	// Aggiungi tutte le componenti al corpo principale della nave
	bodyMesh.Add(noseMesh)      // Aggiungi il "cockpit" anteriore
	bodyMesh.Add(leftWingMesh)  // Aggiungi l'ala sinistra
	bodyMesh.Add(rightWingMesh) // Aggiungi l'ala destra
	bodyMesh.Add(leftArmMesh)   // Aggiungi il braccio sinistro
	bodyMesh.Add(rightArmMesh)  // Aggiungi il braccio destro

	// Aggiungi la nave alla scena
	gs.plane = bodyMesh
	gs.scene.Add(gs.plane)

	// Aggiungiamo degli assi di riferimento per visualizzare l'orientamento
	planeAxes := helper.NewAxes(2.0)
	gs.plane.Add(planeAxes)
}

func setupLighting(gs *GameState) {
	gs.scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8))
	gs.pointLight = light.NewPoint(&math32.Color{1, 1, 1}, 500.0)
	gs.pointLight.SetPosition(gs.planets[0].Position().X, gs.planets[0].Position().Y, gs.planets[0].Position().Z)
	gs.scene.Add(gs.pointLight)
}

func setupCallbacks(gs *GameState) {
	// Window resize callback
	onResize := func(evname string, ev interface{}) {
		width, height := gs.app.GetSize()
		gs.app.Gls().Viewport(0, 0, int32(width), int32(height))
		gs.cam.SetAspect(float32(width) / float32(height))
	}
	gs.app.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	// Keyboard callbacks
	gs.app.Subscribe(window.OnKeyDown, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		gs.keys[kev.Key] = true
	})

	gs.app.Subscribe(window.OnKeyUp, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		gs.keys[kev.Key] = false
	})
}

func updateMovement(gs *GameState) {
	// Update acceleration based on input
	if gs.keys[window.KeyW] {
		gs.acceleration = 0.01
	} else if gs.keys[window.KeyS] {
		gs.acceleration = -0.01
	} else {
		gs.acceleration = 0
	}

	// Update speed
	gs.speed += gs.acceleration
	gs.speed = math32.Clamp(gs.speed, -gs.maxSpeed, gs.maxSpeed)

	// Apply drag
	if gs.acceleration == 0 {
		if gs.speed > 0 {
			gs.speed = math32.Max(0, gs.speed-gs.dragCoeff)
		} else if gs.speed < 0 {
			gs.speed = math32.Min(0, gs.speed+gs.dragCoeff)
		}
	}

	// Apply movement
	gs.plane.TranslateZ(gs.speed)
	if gs.keys[window.KeySpace] {
		if gs.speed > 0 {
			gs.speed = math32.Max(0, gs.speed-gs.airBrake)
		} else if gs.speed < 0 {
			gs.speed = math32.Min(0, gs.speed+gs.airBrake)
		}
	}
	if gs.keys[window.KeyE] {
		gs.plane.RotateY(-gs.rotationSpeed)
	}
	if gs.keys[window.KeyQ] {
		gs.plane.RotateY(gs.rotationSpeed)
	}
	if gs.keys[window.KeyA] {
		gs.plane.RotateZ(-gs.rotationSpeed)
	}
	if gs.keys[window.KeyD] {
		gs.plane.RotateZ(gs.rotationSpeed)
	}
	if gs.keys[window.KeyM] {
		gs.plane.RotateX(-gs.rotationSpeed)
	}
	if gs.keys[window.KeyK] {
		gs.plane.RotateX(gs.rotationSpeed)
	}
}

func updateCamera(gs *GameState) {
	// Crea una matrice di trasformazione per la camera basata sulla trasformazione dell'aereo
	planeMatrix := gs.plane.Matrix()

	// Crea il vettore di offset della camera nel sistema di coordinate locale
	offset := gs.cameraOffset.Clone()

	// Calcola la posizione della camera
	cameraPos := gs.plane.Position()

	// Crea una matrice di rotazione basata solo sulla rotazione dell'aereo
	rotMatrix := math32.NewMatrix4()
	rotMatrix.ExtractRotation(&planeMatrix)

	// Applica la rotazione all'offset
	offset.ApplyMatrix4(rotMatrix)

	// Aggiungi l'offset alla posizione dell'aereo
	cameraPos.Add(offset)

	// Imposta la posizione della camera
	gs.cam.SetPositionVec(&cameraPos)

	// Calcola il vettore "up" ruotato
	up := math32.NewVector3(0, 1, 0)
	up.ApplyMatrix4(rotMatrix)

	// Fai puntare la camera verso l'aereo
	planePos := gs.plane.Position()
	gs.cam.LookAt(&planePos, up)
}

func updatePlanets(gs *GameState) {
	gs.planets[0].RotateY(0.01)
	gs.planets[1].RotateY(-0.005)
	gs.planets[2].RotateX(0.007)
	gs.planets[3].RotateZ(0.003)

	// Update light position
	gs.pointLight.SetPosition(gs.planets[0].Position().X, gs.planets[0].Position().Y, gs.planets[0].Position().Z)
}

func main() {
	a := app.App()
	gs := initGameState(a)

	setupScene(gs)
	setupCamera(gs)
	setupPlanets(gs)
	setupPlane(gs)
	setupLighting(gs)
	setupCallbacks(gs)
	setupHUD(gs) // Aggiungi questa riga

	// Main game loop
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		updateMovement(gs)
		updateCamera(gs)
		updatePlanets(gs)

		// Update HUD
		updateHUD(gs)

		// Render
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(gs.scene, gs.cam)
	})
}
