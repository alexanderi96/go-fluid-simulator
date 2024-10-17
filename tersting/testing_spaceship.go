package main

import (
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
)

func main() {
	// Create application and scene
	a := app.App()
	scene := core.NewNode()

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

	// Create perspective camera
	cam := camera.New(1)
	cam.SetPosition(0, 2, 8) // Posiziona la camera leggermente più in alto e indietro
	scene.Add(cam)

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	onResize := func(evname string, ev interface{}) {
		width, height := a.GetSize()
		a.Gls().Viewport(0, 0, int32(width), int32(height))
		cam.SetAspect(float32(width) / float32(height))
	}
	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	// Create reference objects (planets/asteroids)
	createSphere := func(radius float32, color string, x, y, z float32) *graphic.Mesh {
		geom := geometry.NewSphere(float64(radius), 32, 32)
		mat := material.NewStandard(math32.NewColor(color))
		sphere := graphic.NewMesh(geom, mat)
		sphere.SetPosition(x, y, z)
		scene.Add(sphere)
		return sphere
	}

	// Create various reference objects
	planet1 := createSphere(1.0, "Red", 10, 0, 10)
	planet2 := createSphere(0.7, "Green", -10, 5, -10)
	planet3 := createSphere(0.5, "Yellow", -5, -3, 15)
	planet4 := createSphere(0.3, "Purple", 15, 2, -8)

	// Add a grid helper for reference
	grid := helper.NewGrid(50, 1, &math32.Color{0.4, 0.4, 0.4})
	scene.Add(grid)

	// Create a blue torus (the "plane") and add it to the scene
	geom := geometry.NewCone(1, 3, 3, 1, true) // Base larga, altezza lunga
	mat := material.NewStandard(math32.NewColor("DarkGreen"))
	plane := graphic.NewMesh(geom, mat)
	scene.Add(plane)

	// Create and add lights to the scene
	scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8))
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 500.0)
	pointLight.SetPosition(planet1.Position().X, planet1.Position().Y, planet1.Position().Z) // Posiziona la luce nel pianeta 1
	scene.Add(pointLight)

	// Create and add an axis helper to the scene
	scene.Add(helper.NewAxes(0.5))

	// Set background color to dark
	a.Gls().ClearColor(0.1, 0.1, 0.2, 1.0)

	// For displaying object data
	infoLabel := gui.NewLabel("Position: (0,0,0)")
	infoLabel.SetPosition(10, 10)
	scene.Add(infoLabel)

	// Variable to control movement and rotation
	var speed float32 = 0.0        // Inizializza la velocità a zero
	var maxSpeed float32 = 0.1     // Velocità massima
	var acceleration float32 = 0.0 // Acceleration control
	var rotationSpeed float32 = 0.01
	var dragCoefficient float32 = 0.01 // Coefficiente di drag

	// Map to keep track of pressed keys
	keys := make(map[window.Key]bool)

	// Subscribe to keyboard events
	a.Subscribe(window.OnKeyDown, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		keys[kev.Key] = true
	})

	a.Subscribe(window.OnKeyUp, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		keys[kev.Key] = false
	})

	// Offset for the camera relative to the plane
	cameraOffset := math32.NewVector3(0, 3, 10) // Fissa la posizione relativa della camera alla navicella

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		// Check which keys are pressed and apply the corresponding transformations
		if keys[window.KeyW] {
			acceleration = 0.01 // Aumenta l'accelerazione
		} else if keys[window.KeyS] {
			acceleration = -0.01 // Diminuisci l'accelerazione
		} else {
			acceleration = 0 // Se nessun tasto è premuto, non accelerare
		}

		// Applica l'accelerazione alla velocità
		speed += acceleration
		if speed > maxSpeed {
			speed = maxSpeed // Limita la velocità massima
		} else if speed < -maxSpeed {
			speed = -maxSpeed // Limita la velocità minima
		}

		// Applica il coefficiente di drag alla velocità
		if acceleration == 0 { // Se non stiamo accelerando o decelerando
			if speed > 0 {
				speed -= dragCoefficient
				if speed < 0 {
					speed = 0 // Non andare sotto zero
				}
			} else if speed < 0 {
				speed += dragCoefficient
				if speed > 0 {
					speed = 0 // Non andare sopra zero
				}
			}
		}

		// Applica la velocità al piano
		plane.TranslateY(speed)

		if keys[window.KeyA] {
			plane.TranslateX(-speed)
		}
		if keys[window.KeyD] {
			plane.TranslateX(speed)
		}
		if keys[window.KeyLeft] {
			plane.RotateZ(rotationSpeed)
		}
		if keys[window.KeyRight] {
			plane.RotateZ(-rotationSpeed)
		}
		if keys[window.KeyUp] {
			plane.RotateX(-rotationSpeed)
		}
		if keys[window.KeyDown] {
			plane.RotateX(rotationSpeed)
		}

		// Update the position of the point light to follow the first planet
		pointLight.SetPosition(planet1.Position().X, planet1.Position().Y, planet1.Position().Z)

		// **Camera Update**
		// Mantieni la posizione della camera fissa rispetto alla navicella
		planePos := plane.Position()
		// La camera segue la navicella, ma mantiene un offset costante
		cam.SetPositionVec(planePos.Clone().Add(cameraOffset))
		cam.LookAt(&planePos, math32.NewVector3(0, 1, 0)) // La camera guarda sempre la navicella

		// Rotate reference planets for visual interest
		planet1.RotateY(0.01)
		planet2.RotateY(-0.005)
		planet3.RotateX(0.007)
		planet4.RotateZ(0.003)

		// Render
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, cam)
	})
}
