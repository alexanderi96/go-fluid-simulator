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
	fileSelector *draw.FileSelect
	ambientLight = &math32.Color{R: 0.3, G: 0.3, B: 0.4} // Adjusted for medium galaxy intensity with slight blue tint
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
	s.Scene.Add(helper.NewAxes(10))

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	onResize := func(evname string, ev interface{}) {
		// Get framebuffer size and update viewport accordingly
		width, height := s.App.GetSize()
		s.App.Gls().Viewport(0, 0, int32(width), int32(height))
		// Update the camera's aspect ratio
		s.Cam.SetAspect(float32(width) / float32(height))
	}
	s.App.Subscribe(window.OnWindowSize, onResize)

	// Create file selector
	fileSelector, err = draw.NewFileSelect(400, 300, func(path string) error {
		s.ResetSimulation()
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		return s.LoadFSSScene(file)
	})
	if err != nil {
		log.Fatal(err)
	}
	s.Scene.Add(fileSelector)
	fileSelector.SetPath("assets/scenes")

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
	s.App.Gls().ClearColor(bgColor.R, bgColor.G, bgColor.B, 0.1)

	// Create and add ambient light to simulate galaxy lighting
	s.Scene.Add(light.NewAmbient(ambientLight, 1.0))

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
		if s.SpaceShip != nil {
			s.SpaceShip.Keys[kev.Key] = true
		}

		if kev.Key == window.KeyF {
			s.Fly = !s.Fly
			if s.Fly {
				// Create spaceship if not present
				if s.SpaceShip == nil {
					s.SpaceShip = &spaceship.SpaceShip{
						Speed:           0,
						MaxSpeed:        1,
						MaxEngineThrust: 0.5,
						Thrust:          0,
						RotationSpeed:   0.02,
						BreakingPower:   0.1,
						Mass:            10,
						Keys:            make(map[window.Key]bool),
						CameraOffset:    math32.NewVector3(0, 2, -8),
					}
					s.SpaceShip.LoadShip()
					s.Scene.Add(s.SpaceShip.Ship)
				}
			} else {
				// Remove spaceship when exiting flight mode
				if s.SpaceShip != nil {
					s.Scene.Remove(s.SpaceShip.Ship)
					s.SpaceShip = nil
				}
			}
		}

		if !s.Fly {
			if kev.Key == window.KeyR {
				s.ResetSimulation()
			}
			if kev.Key == window.KeySpace {
				s.IsPause = !s.IsPause
			}
			if kev.Key == window.KeyEqual && kev.Mods == window.ModShift { // + key
				s.TimeScale *= 10
			}
			if kev.Key == window.KeyMinus { // - key
				s.TimeScale /= 10
				if s.TimeScale < 1 {
					s.TimeScale = 1
				}
			}
			if kev.Key == window.KeyS {
				s.SaveSimulation("simulation" + time.Now().Format("2006-01-02 15:04:05") + ".json")
			}
			if kev.Key == window.KeyL {
				fileSelector.Show(true)
			}
		}
	})

	s.App.Subscribe(window.OnKeyUp, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		if s.SpaceShip != nil {
			s.SpaceShip.Keys[kev.Key] = false
		}

	})

	s.AppStartTime = time.Now()
	s.App.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		if !s.IsPause {
			if err := s.Update(); err != nil {
				log.Fatal("Errore durante l'update della simulazione %w", err)
			}

		}

		draw.UpdateHUD(s, deltaTime)

		if s.Fly && !s.IsPause && s.SpaceShip != nil {
			spaceship.UpdateMovement(s.SpaceShip)
			UpdateCamera(s)
		}
		// gui.Draw(s)
		s.App.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)

		renderer.Render(s.Scene, s.Cam)

		// s.Config.UpdateWindowSettings()
	})

}
