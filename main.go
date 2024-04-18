package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/alexanderi96/go-fluid-simulator/config"

	"log"

	"github.com/alexanderi96/go-fluid-simulator/physics"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
)

var (
	s *physics.Simulation
)

func init() {
	config, err := config.ReadConfig("./config.toml")
	if err != nil {
		log.Fatal(err)
	}

	// Create application and scene
	s, err = physics.NewSimulation(config)
	if err != nil {
		log.Fatal(err)
	}

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(s.Scene)

	// Create perspective camera
	s.Cam = camera.New(1)
	s.Cam.SetPosition(0, 0, 3)
	s.Scene.Add(s.Cam)

	// Set up orbit control for the camera
	camera.NewOrbitControl(s.Cam)

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	onResize := func(evname string, ev interface{}) {
		// Get framebuffer size and update viewport accordingly
		width, height := s.App.GetSize()
		s.App.Gls().Viewport(0, 0, int32(width), int32(height))
		// Update the camera's aspect ratio
		s.Cam.SetAspect(float32(width) / float32(height))
	}
	s.App.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

}

func main() {
	if s.Config.ShouldBeProfiled {
		f, err := os.Create("cpu.pprof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Set background color to gray
	s.App.Gls().ClearColor(0.5, 0.5, 0.5, 1.0)

	// Create and add lights to the scene
	s.Scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8))
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(0, 0, 0)
	s.Scene.Add(pointLight)

	// Handle right mouse button press to create a new ball
	s.App.Subscribe(window.OnMouseDown, func(evname string, ev interface{}) {
		mev := ev.(*window.MouseEvent)
		if mev.Button == window.MouseButtonRight { // Check if the right mouse button was pressed
			units := s.GetUnits()
			s.PositionNewUnitsCube(units)

			s.Fluid = append(s.Fluid, units...)
		}
	})

	// Create FPS label
	fpsLabel := gui.NewLabel("FPS: 0")
	fpsLabel.SetPosition(10, 10)
	s.Scene.Add(fpsLabel)

	// Create FPS label
	unitLabel := gui.NewLabel("unit: 0")
	unitLabel.SetPosition(10, 20)
	s.Scene.Add(unitLabel)

	s.App.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {

		if !s.IsInputBeingHandled {
			go s.HandleInput()
		}

		if !s.IsPause {
			if err := s.Update(); err != nil {
				log.Fatal("Errore durante l'update della simulazione %w", err)
			}

		}
		// Update FPS label
		deltaTimeSec := float32(deltaTime.Seconds())

		fps := 1.0 / deltaTimeSec
		fpsLabel.SetText("FPS: " + fmt.Sprintf("%.2f", fps))

		unitLabel.SetText("unit: " + fmt.Sprintf("%d", len(s.Fluid)))

		// gui.Draw(s)
		s.App.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)

		renderer.Render(s.Scene, s.Cam)

		// s.Config.UpdateWindowSettings()
	})

}
