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
	simulation *physics.Simulation
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

	simulation, err = physics.NewSimulation(config)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	if simulation.Config.ShouldBeProfiled {
		f, err := os.Create("cpu.pprof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	for !rl.WindowShouldClose() {
		simulation.Config.UpdateWindowSettings()

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		if rl.IsKeyPressed(rl.KeyR) {
			simulation.Reset() // Resetta il campo
		} else if rl.IsKeyPressed(rl.KeySpace) {
			simulation.IsPause = !simulation.IsPause
		} else if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			mousePosition := rl.GetMousePosition()
			simulation.NewFluidAtPosition(mousePosition)
		} else if rl.IsMouseButtonPressed(rl.MouseRightButton) {
			mousePosition := rl.GetMousePosition()
			simulation.NewFluidWithVelocity(mousePosition)
		}

		if !simulation.IsPause {
			if err := simulation.Update(); err != nil {
				log.Fatal("Errore durante l'update della simulazione %w", err)
			}

		}

		gui.Draw(simulation)

		rl.EndDrawing()
	}

	rl.CloseWindow()
}
