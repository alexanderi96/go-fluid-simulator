package main

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/alexanderi96/go-fluid-simulator/config"
	"github.com/alexanderi96/go-fluid-simulator/gui"
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
		s.Config.UpdateWindowSettings()

		if !s.IsInputBeingHandled {
			go s.HandleInput()
		}

		if !s.IsPause {
			if err := s.Update(); err != nil {
				log.Fatal("Errore durante l'update della simulazione %w", err)
			}

		}

		gui.Draw(s)
	}

	rl.CloseWindow()
}
