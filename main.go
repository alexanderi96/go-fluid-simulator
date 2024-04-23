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
	"github.com/g3n/engine/util/helper"
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

	// Create and add an axis helper to the scene
	s.Scene.Add(helper.NewAxes(10 * float32(s.Config.UnitRadiusMultiplier)))

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
	s.App.Gls().ClearColor(0, 0, 0, 1.0)

	// Create and add lights to the scene
	s.Scene.Add(light.NewAmbient(&math32.Color{R: 1.0, G: 1.0, B: 1.0}, 0.1))
	pointLight := light.NewPoint(&math32.Color{R: 1, G: 1, B: 1}, 1000.0)
	pointLight.SetPosition(float32(s.Config.GameX), float32(s.Config.GameY), float32(s.Config.GameZ))
	s.Scene.Add(pointLight)

	// Handle mouse input
	s.App.Subscribe(window.OnMouseDown, func(evname string, ev interface{}) {

		mev := ev.(*window.MouseEvent)
		if mev.Button == window.MouseButtonLeft && mev.Mods == window.ModControl {
			// Controlla se sia il pulsante destro del mouse sia il pulsante Ctrl sinistro sono premuti
			units := s.GetUnits()
			s.PositionNewUnitsCube(units)

			s.Fluid = append(s.Fluid, units...)
		} else if mev.Button == window.MouseButtonRight && mev.Mods == window.ModControl {
			// Controlla se sia il pulsante destro del mouse sia il pulsante Ctrl sinistro sono premuti
			units := s.GetUnits()
			s.PositionNewUnitsFibonacci(units)

			s.Fluid = append(s.Fluid, units...)
		}
	})

	// Handle keyboard input
	s.App.Subscribe(window.OnKeyDown, func(evname string, ev interface{}) {

		kev := ev.(*window.KeyEvent)
		if kev.Key == window.KeyR {
			s.ResetSimulation()
		}
		if kev.Key == window.KeySpace {
			s.IsPause = !s.IsPause
		}
		if kev.Key == window.KeyS {
			s.SaveSimulation("simulation.json")
		}
		if kev.Key == window.KeyL {

			sim, err := physics.LoadSimulation("simulation.json")
			if err != nil {
				log.Fatal(err)
			}
			s.ResetSimulation()

			s.Fluid = sim.Fluid

			for _, unit := range s.Fluid {
				unit.GenerateMesh()
				s.Scene.Add(unit.Mesh)
			}
			s.Config = sim.Config
			s.IsPause = sim.IsPause
			s.WorldBoundray = sim.WorldBoundray
			s.WorldCenter = sim.WorldCenter
		}
	})

	// Create FPS label
	fpsLabel := gui.NewLabel("FPS: 0")
	fpsLabel.SetPosition(10, 10)
	s.Scene.Add(fpsLabel)

	// Create FPS label
	ftLabel := gui.NewLabel("FrameTime: 0")
	ftLabel.SetPosition(100, 10)
	s.Scene.Add(ftLabel)

	// Create FPS label
	unitLabel := gui.NewLabel("unit: 0")
	unitLabel.SetPosition(10, 25)
	s.Scene.Add(unitLabel)

	// Create FPS label
	simDurationLabel := gui.NewLabel("Simulation duration: 0")
	simDurationLabel.SetPosition(10, 40)
	s.Scene.Add(simDurationLabel)

	// Create FPS label
	realDurationLabel := gui.NewLabel("Real duration: 0")
	realDurationLabel.SetPosition(10, 55)
	s.Scene.Add(realDurationLabel)

	appStartTime := time.Now()
	s.App.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		if !s.IsPause {
			if err := s.Update(); err != nil {
				log.Fatal("Errore durante l'update della simulazione %w", err)
			}

		}

		fps := 1.0 / float64(deltaTime.Seconds())
		fpsLabel.SetText("FPS: " + fmt.Sprintf("%.2f", fps))

		ftLabel.SetText("FrameTime: " + fmt.Sprintf("%.2f", s.Config.Frametime))

		unitLabel.SetText("unit: " + fmt.Sprintf("%d", len(s.Fluid)))

		simDurationLabel.SetText("Simulation duration: " + fmt.Sprintf("%.2f", s.Metrics.SimDuration))

		realDurationLabel.SetText("Real duration: " + fmt.Sprintf("%.2f", -time.Until(appStartTime).Seconds()))

		// gui.Draw(s)
		s.App.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)

		renderer.Render(s.Scene, s.Cam)

		// s.Config.UpdateWindowSettings()
	})

}
