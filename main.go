package main

import (
	"os"
	"runtime/pprof"
	"time"

	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/draw"
	"github.com/alexanderi96/go-fluid-simulator/spaceship"

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
	s            *physics.Simulation
	ambientLight = &math32.Color{R: 0.1, G: 0.1, B: 0.1}
	pointLight   = &math32.Color{R: 1.0, G: 1.0, B: 1.0}
	bgColor      = &math32.Color{R: 0.01, G: 0.01, B: 0.01}
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
	s.Cam.SetPosition(0, 2, 8)
	s.Cam.SetFar(1.7e38) // Imposta il valore desiderato, max 1.7e38
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

	draw.SetupHUD(s)
	onResize("", nil)

}

func UpdateCamera(s *physics.Simulation) {
	planeMatrix := s.SpaceShip.Ship.Matrix()
	offset := s.SpaceShip.CameraOffset.Clone()
	cameraPos := s.SpaceShip.Ship.Position()

	rotMatrix := math32.NewMatrix4()
	rotMatrix.ExtractRotation(&planeMatrix)
	offset.ApplyMatrix4(rotMatrix)
	cameraPos.Add(offset)

	s.Cam.SetPositionVec(&cameraPos)

	up := math32.NewVector3(0, 1, 0)
	up.ApplyMatrix4(rotMatrix)

	planePos := s.SpaceShip.Ship.Position()
	s.Cam.LookAt(&planePos, up)
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
	s.App.Gls().ClearColor(bgColor.R, bgColor.G, bgColor.B, 1.0)

	// Create and add lights to the scene
	s.Scene.Add(light.NewAmbient(ambientLight, 1))
	// pointLight := light.NewPoint(pointLight, 1e10)
	// pointLight.SetPosition(float32(s.Config.GameX), float32(s.Config.GameY), float32(s.Config.GameZ))
	// s.Scene.Add(pointLight)

	// Handle mouse input
	s.App.Subscribe(window.OnMouseDown, func(evname string, ev interface{}) {

		if !s.Fly {
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
			} else if mev.Button == window.MouseButtonRight && mev.Mods == window.ModShift {
				// Controlla se sia il pulsante destro del mouse sia il pulsante Ctrl sinistro sono premuti
				units := s.GetUnits()
				s.PositionNewUnitsFibonacci(units)
				s.GiveRotationalVelocity(units)
				s.Fluid = append(s.Fluid, units...)
			} else if mev.Button == window.MouseButtonLeft && mev.Mods == window.ModShift {
				// Controlla se sia il pulsante destro del mouse sia il pulsante Ctrl sinistro sono premuti
				units := s.GetUnits()
				s.PositionNewUnitsCube(units)
				s.GiveRotationalVelocity(units)
				s.Fluid = append(s.Fluid, units...)
			}
		}
	})

	// Handle keyboard input
	s.App.Subscribe(window.OnKeyDown, func(evname string, ev interface{}) {

		kev := ev.(*window.KeyEvent)
		s.SpaceShip.Keys[kev.Key] = true

		if kev.Key == window.KeyF {
			s.Fly = !s.Fly
		}

		if !s.Fly {
			if kev.Key == window.KeyR {
				s.ResetSimulation()
			}
			if kev.Key == window.KeySpace {
				s.IsPause = !s.IsPause
			}
			if kev.Key == window.KeyS {
				s.SaveSimulation("simulation" + time.Now().Format("2006-01-02 15:04:05") + ".json")
			}
			if kev.Key == window.KeyL {

				sim, err := physics.LoadSimulation("simulation.json")
				if err != nil {
					log.Fatal(err)
				}
				s.ResetSimulation()

				s.Fluid = sim.Fluid

				for _, unit := range s.Fluid {
					unit.NewPointLightMesh()
					s.Scene.Add(unit.Mesh)
				}
				s.Config = sim.Config
				s.IsPause = sim.IsPause
				s.WorldBoundray = sim.WorldBoundray
				s.WorldCenter = sim.WorldCenter
			}
		}
	})

	s.App.Subscribe(window.OnKeyUp, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		s.SpaceShip.Keys[kev.Key] = false
	})

	s.AppStartTime = time.Now()
	s.App.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		if !s.IsPause {
			if err := s.Update(); err != nil {
				log.Fatal("Errore durante l'update della simulazione %w", err)
			}

		}

		draw.UpdateHUD(s, deltaTime)

		if s.Fly {
			spaceship.UpdateMovement(s.SpaceShip)
			UpdateCamera(s)
		}
		// gui.Draw(s)
		s.App.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)

		renderer.Render(s.Scene, s.Cam)

		// s.Config.UpdateWindowSettings()
	})

}
