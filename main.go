package main

import (
	"flag"
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

	testing = flag.Bool("testing", false, "Enable testing mode")
)

type TestData struct {
	Position rl.Vector3
	Duration float32
}

func init() {
	flag.Parse()
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
	physics.InitOctree(config)

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

	if *testing {
		testLoop()
	} else {
		runLoop()
	}

	rl.CloseWindow()
}

func runLoop() {
	for !rl.WindowShouldClose() {

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
}

func testLoop() {
	testsData := make([]*TestData, 0)
	testsNumber := int32(0)
	testDuration := s.Config.TestDuration
	log.Printf("Test duration: %f\nTesting...\n\n", testDuration)

	for !rl.WindowShouldClose() && testsNumber < s.Config.TestIterations {
		s.InitTest()

		duration := float32(0)
		log.Printf("Test %d of %d...", testsNumber+1, s.Config.TestIterations)
		for testDuration > duration {
			if err := s.Update(); err != nil {
				log.Fatal("Errore durante l'update della simulazione %w", err)
			}

			gui.Draw(s)
			s.Config.UpdateWindowSettings()
			duration += float32(rl.GetFrameTime())
		}

		testsData = append(testsData, &TestData{
			Position: s.Fluid[len(s.Fluid)-1].Position,
			Duration: duration,
		})
		s.ResetSimulation()
		testsNumber++
	}

	log.Print("Test results:")
	for test := range testsData {
		log.Printf("Test %d of %d Position: %v Duration: %f", test+1, s.Config.TestIterations, testsData[test].Position, testsData[test].Duration)
	}

}
