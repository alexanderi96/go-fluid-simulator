package main

import (
	"fmt"
	"math"
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
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(1, 0, 2)
	scene.Add(pointLight)

	// Create and add an axis helper to the scene
	scene.Add(helper.NewAxes(0.5))

	// Set background color to dark
	a.Gls().ClearColor(0.1, 0.1, 0.2, 1.0)

	// Constants for movement and rotation
	const (
		moveSpeed       = 0.05
		rotationSpeed   = 0.05
		speedMultiplier = 0.5
		drag            = 0.98 // Resistance factor
		cameraLag       = 0.1  // How quickly the camera follows (0-1)
		cameraDistance  = 8.0  // Distance the camera maintains behind the ship
		cameraHeight    = 2.0  // Height of the camera above the ship
	)

	// Current rotation angles
	rotX := float32(0)
	rotY := float32(0)
	rotZ := float32(0)

	// Variables for movement
	var velocity math32.Vector3
	var isMoving bool

	// Direction vector (initially pointing forward)
	direction := math32.NewVector3(0, 0, 1)

	// For displaying object data
	infoLabel := gui.NewLabel("Position: (0,0,0)\nVelocity: (0,0,0)\nRotation: (0,0,0)\nDirection: (0,0,0)")
	infoLabel.SetPosition(10, 10)
	scene.Add(infoLabel)

	// Function to update camera position
	updateCamera := func() {
		// Get current plane position
		planePos := plane.Position()

		// Calculate desired camera position (behind and slightly above the plane)
		desiredPos := math32.NewVector3(
			planePos.X-direction.X*float32(cameraDistance),
			planePos.Y+float32(cameraHeight),
			planePos.Z-direction.Z*float32(cameraDistance),
		)

		// Get current camera position
		currentPos := cam.Position()

		// Interpolate between current and desired position
		currentPos.X += (desiredPos.X - currentPos.X) * float32(cameraLag)
		currentPos.Y += (desiredPos.Y - currentPos.Y) * float32(cameraLag)
		currentPos.Z += (desiredPos.Z - currentPos.Z) * float32(cameraLag)

		// Update camera position
		cam.SetPosition(currentPos.X, currentPos.Y, currentPos.Z)

		// Make camera look at plane
		cam.LookAt(&planePos, math32.NewVector3(0, 1, 0))
	}

	// Function to calculate direction vector based on rotation angles
	updateDirection := func() {
		// Convert rotation angles to radians
		rx := float64(rotX)
		ry := float64(rotY)
		// rz := float64(rotZ)

		// Calculate direction using rotation matrices
		x := float32(math.Cos(ry) * math.Sin(rx))
		y := float32(math.Sin(ry))
		z := float32(math.Cos(ry) * math.Cos(rx))

		direction.Set(x, y, z)
		direction.Normalize()
	}

	// Handle keyboard input
	a.Subscribe(window.OnKeyDown, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)

		updateDirection()

		switch kev.Key {
		case window.KeyW: // Forward
			velocity.X += direction.X * moveSpeed * speedMultiplier
			velocity.Y += direction.Y * moveSpeed * speedMultiplier
			velocity.Z += direction.Z * moveSpeed * speedMultiplier
			isMoving = true
		case window.KeyS: // Backward
			velocity.X -= direction.X * moveSpeed * speedMultiplier
			velocity.Y -= direction.Y * moveSpeed * speedMultiplier
			velocity.Z -= direction.Z * moveSpeed * speedMultiplier
			isMoving = true
		case window.KeyA: // Strafe left
			left := math32.NewVector3(0, 1, 0).Cross(direction)
			velocity.X += left.X * moveSpeed * speedMultiplier
			velocity.Z += left.Z * moveSpeed * speedMultiplier
			isMoving = true
		case window.KeyD: // Strafe right
			right := direction.Cross(math32.NewVector3(0, 1, 0))
			velocity.X += right.X * moveSpeed * speedMultiplier
			velocity.Z += right.Z * moveSpeed * speedMultiplier
			isMoving = true
		case window.KeyUp: // Pitch up
			rotX += rotationSpeed
			plane.SetRotationX(rotX)
		case window.KeyDown: // Pitch down
			rotX -= rotationSpeed
			plane.SetRotationX(rotX)
		case window.KeyLeft: // Yaw left
			rotY += rotationSpeed
			plane.SetRotationY(rotY)
		case window.KeyRight: // Yaw right
			rotY -= rotationSpeed
			plane.SetRotationY(rotY)
		case window.KeyQ: // Roll left
			rotZ += rotationSpeed
			plane.SetRotationZ(rotZ)
		case window.KeyE: // Roll right
			rotZ -= rotationSpeed
			plane.SetRotationZ(rotZ)
		}
	})

	// Stop movement when key is released
	a.Subscribe(window.OnKeyUp, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		switch kev.Key {
		case window.KeyW, window.KeyS, window.KeyA, window.KeyD:
			isMoving = false
		}
	})

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		// Apply drag to velocity when not actively moving
		if !isMoving {
			velocity.MultiplyScalar(drag)
		}

		// Update position based on velocity
		pos := plane.Position()
		pos.Add(&velocity)
		plane.SetPosition(pos.X, pos.Y, pos.Z)

		// Update camera position
		updateCamera()

		// Rotate reference planets for visual interest
		planet1.RotateY(0.01)
		planet2.RotateY(-0.005)
		planet3.RotateX(0.007)
		planet4.RotateZ(0.003)

		// Update info label
		infoLabel.SetText(fmt.Sprintf(
			"Position: (%.2f, %.2f, %.2f)\nVelocity: (%.2f, %.2f, %.2f)\nRotation: (%.2f, %.2f, %.2f)\nDirection: (%.2f, %.2f, %.2f)",
			pos.X, pos.Y, pos.Z,
			velocity.X, velocity.Y, velocity.Z,
			rotX, rotY, rotZ,
			direction.X, direction.Y, direction.Z,
		))

		// Render
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, cam)
	})
}
