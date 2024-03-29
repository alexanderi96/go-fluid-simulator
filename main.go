package main

import (
	"os"
	"runtime/pprof"

	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/gui"

	"log"

	"github.com/alexanderi96/go-fluid-simulator/physics"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	s *physics.Simulation
)

func init() {
	config, err := config.ReadConfig("./config.toml")
	if err != nil {
		log.Fatal(err)
	}

	if config.IsResizable {
		rl.SetConfigFlags(rl.FlagWindowResizable)
	}

	rl.InitWindow(config.WindowWidth, config.WindowHeight, "Go Fluid Simulator")
	rl.SetTargetFPS(config.TargetFPS)

	s, err = physics.NewSimulation(config)
	if err != nil {
		log.Fatal(err)
	}
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

	for !rl.WindowShouldClose() {
		// s.Angle += 0.01 // Velocità di rotazione
		// s.Camera.Position = rl.NewVector3(float32(math.Sin(s.Angle)*10.0), 5.0, float32(math.Cos(s.Angle)*10.0))

		if !s.IsInputBeingHandled {
			go s.HandleInput()
		}

		if !s.IsPause {
			if err := s.Update(); err != nil {
				log.Fatal("Errore durante l'update della simulazione %w", err)
			}

		}

		gui.Draw(s)
		s.Config.UpdateWindowSettings()
	}

	rl.CloseWindow()
}
